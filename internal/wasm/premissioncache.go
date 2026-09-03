// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wasm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/avast/retry-go/v5"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"

	"github.com/envoyproxy/gateway/internal/logging"
)

type permissionCacheOptions struct {
	// checkInterval is the interval to recheck the permission for the cached permission entries.
	checkInterval time.Duration

	// permissionExpiry is the expiry time for permission cache entry.
	// The permission cache entry will be updated by rechecking the OCI image permission against the pull secret.
	permissionExpiry time.Duration

	// cacheExpiry is the expiry time for the permission cache.
	// The permission cache will be removed if it is not accessed for the specified expiry time.
	// This is used to purge the cache.
	cacheExpiry time.Duration

	// permissionCheckTimeout is the maximum time for one background permission recheck.
	permissionCheckTimeout time.Duration
}

// sanitize validates and sets the default values for the permission cache options.
func (o *permissionCacheOptions) sanitize() {
	if o.checkInterval == 0 {
		o.checkInterval = 5 * time.Minute
	}
	if o.permissionExpiry == 0 {
		o.permissionExpiry = 1 * time.Hour
	}
	if o.cacheExpiry == 0 {
		o.cacheExpiry = 24 * time.Hour
	}
	if o.permissionCheckTimeout == 0 {
		o.permissionCheckTimeout = DefaultPermissionCheckTimeout
	}
}

// permissionCache is a cache for permission check for private OCI images.
// After a new permission is put into the cache, it will be checked periodically by a background goroutine.
// It is used to avoid blocking the translator due to the permission check.
type permissionCache struct {
	sync.Mutex
	permissionCacheOptions

	cache  map[string]*permissionCacheEntry
	logger logging.Logger
}

// permissionCacheEntry is an entry in the permission cache.
type permissionCacheEntry struct {
	// The oci image URL.
	image *url.URL
	// fetcherOption contains the pull secret for the image.
	fetcherOption *ImageFetcherOption
	// The last time the pull secret is checked against the image.
	lastCheck time.Time
	// The error returned by the OCI registry when checking the permission.
	// If error is not nil, the permission is not allowed.
	// If it's a permission error, it's represented by a transport.Error with 401 or 403 HTTP status code.
	// But it's not necessarily a permission error, it could be other errors like network error, non-exist image, etc.
	// In this case, the permission is also not allowed.
	checkError error
	// The last time the cache entry is accessed.
	lastAccess time.Time
}

func (e *permissionCacheEntry) key() string {
	return permissionCacheKey(e.image, e.fetcherOption.PullSecret, e.fetcherOption.CACert)
}

func (e *permissionCacheEntry) copy() *permissionCacheEntry {
	return &permissionCacheEntry{
		image:         e.image,
		fetcherOption: e.fetcherOption,
		lastCheck:     e.lastCheck,
		checkError:    e.checkError,
		lastAccess:    e.lastAccess,
	}
}

// isPermissionExpired returns true if the permission check is older
// than the specified expiry duration. If this is true, the entry
// should be rechecked.
func (e *permissionCacheEntry) isPermissionExpired(expiry time.Duration) bool {
	return time.Now().After(e.lastCheck.Add(expiry))
}

// isCacheExpired returns true if the cache entry has not been accessed
// for the specified expiry duration. If this is true, the entry
// should be removed.
func (e *permissionCacheEntry) isCacheExpired(expiry time.Duration) bool {
	return time.Now().After(e.lastAccess.Add(expiry))
}

// permissionCacheKey generates a key for a permission cache entry.
// The key is the hex encoded SHA-256 of the image URL, the pull secret, and the
// CA certificate. The pull secret and CA certificate are hex encoded before
// hashing so that they can't contain the separator, making the split between the
// components unambiguous.
func permissionCacheKey(image *url.URL, pullSecret, caCert []byte) string {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s|%x|%x", image.String(), pullSecret, caCert))
	return hex.EncodeToString(sum[:])
}

// newPermissionCache creates a new permission cache with a given TTL.
func newPermissionCache(options permissionCacheOptions, logger logging.Logger) *permissionCache {
	options.sanitize()
	return &permissionCache{
		cache:                  make(map[string]*permissionCacheEntry),
		permissionCacheOptions: options,
		logger:                 logger,
	}
}

// checkPermission checks the permission of the image against the pull secret.
func (p *permissionCache) checkPermission(ctx context.Context, image *url.URL, fetcherOption ImageFetcherOption) (time.Time, error) {
	fetcher, err := NewImageFetcher(ctx, fetcherOption, p.logger)
	if err != nil {
		// Return a more descriptive error as messages downstream do not indicate the source very well.
		return time.Now(), fmt.Errorf("failed to create image fetcher: %w", err)
	}
	_, _, err = fetcher.PrepareFetch(image.Host + image.Path)
	return time.Now(), err
}

// start starts a background goroutine to periodically check the permission for the cached permission entries.
func (p *permissionCache) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(p.checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				entries := p.pruneExpiredAndGetEntriesToRecheck()
				for _, e := range entries {
					const retryAttempts = 3
					const retryDelay = 1 * time.Second
					lastCheck := e.lastCheck
					checkCtx, cancel := context.WithTimeout(ctx, p.permissionCheckTimeout)
					p.logger.Info("rechecking permission for image", "image", e.image.String())
					err := retry.New(
						retry.Attempts(retryAttempts),
						retry.DelayType(retry.BackOffDelay),
						retry.Delay(retryDelay),
						retry.Context(checkCtx),
					).Do(func() error {
						lastCheck, err := p.checkPermission(checkCtx, e.image, *e.fetcherOption)
						e.checkError = err
						e.lastCheck = lastCheck
						if err != nil && isRetriableError(err) {
							p.logger.Error(
								err,
								"failed to check permission for image, will retry again",
								"image",
								e.image.String())
							return err
						}
						return nil
					})
					cancel()
					if err != nil {
						p.logger.Error(
							err,
							fmt.Sprintf("failed to recheck permission for image after %d attempts", retryAttempts),
							"image",
							e.image.String())
					}
					p.updateRecheckedPermission(e, lastCheck)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *permissionCache) pruneExpiredAndGetEntriesToRecheck() []*permissionCacheEntry {
	p.Lock()
	defer p.Unlock()

	var entries []*permissionCacheEntry
	for key, e := range p.cache {
		if e.isCacheExpired(p.cacheExpiry) {
			p.logger.Info("removing permission cache entry", "image", e.image.String())
			delete(p.cache, key)
			continue
		}
		if e.isPermissionExpired(p.permissionExpiry) {
			entries = append(entries, e.copy())
		}
	}
	return entries
}

func (p *permissionCache) updateRecheckedPermission(e *permissionCacheEntry, lastCheck time.Time) {
	p.Lock()
	defer p.Unlock()

	cached, ok := p.cache[e.key()]
	if !ok {
		return
	}
	// If the last check time is different, it means the permission has been updated by other goroutine,
	// so we should not update the cache with the rechecked result to avoid overwriting the newer permission with the older one.
	if !cached.lastCheck.Equal(lastCheck) {
		return
	}
	cached.checkError = e.checkError
	cached.lastCheck = e.lastCheck
}

func (p *permissionCache) upsertPermission(e *permissionCacheEntry) {
	p.Lock()
	defer p.Unlock()

	key := e.key()
	if cached, ok := p.cache[key]; ok {
		if e.lastCheck.After(cached.lastCheck) {
			cached.checkError = e.checkError
			cached.lastCheck = e.lastCheck
		}
		if e.lastAccess.After(cached.lastAccess) {
			cached.lastAccess = e.lastAccess
		}
		return
	}
	p.cache[key] = e
}

func (p *permissionCache) lookupPermission(key string) (error, bool) {
	p.Lock()
	defer p.Unlock()

	e, ok := p.cache[key]
	if ok {
		e.lastAccess = time.Now()
		return e.checkError, true
	}
	return nil, false
}

// isRetriableError checks if the error is retriable.
// If the error is a permission error, it's not retriable. For example, 401 and 403 HTTP status code.
func isRetriableError(err error) bool {
	var terr *transport.Error
	if errors.As(err, &terr) {
		if terr.StatusCode == http.StatusUnauthorized || terr.StatusCode == http.StatusForbidden {
			return false
		}
	}
	return true
}

// put puts a new permission cache entry into the cache.
func (p *permissionCache) Put(e *permissionCacheEntry) {
	p.Lock()
	defer p.Unlock()
	e.lastAccess = time.Now()
	e.lastCheck = time.Now()
	p.cache[e.key()] = e
}

// IsAllowed checks if the given image is allowed to be accessed with the provided pull secret.
// If the permission is not found in the cache, this method will block until the permission is checked and cached.
// This blocking won't be too long as it's only for the first time permission check and won't retry. Subsequent
// permission checks will be done in a background goroutine by the permission cache.
//
// If any error occurs, the permission is considered not allowed.
// The error can be a permission error or other errors like network error, non-exist image, etc.
func (p *permissionCache) IsAllowed(ctx context.Context, image *url.URL, pullSecret []byte, insecure bool, caCert []byte) (bool, error) {
	key := permissionCacheKey(image, pullSecret, caCert)
	if err, ok := p.lookupPermission(key); ok {
		return err == nil, err
	}

	fetcherOption := ImageFetcherOption{
		Insecure:   insecure,
		PullSecret: pullSecret,
		CACert:     caCert,
	}
	// Do not retry if the permission check fails because we don't want to block the translator for too long.
	// The permission check will be retried in the background goroutine by the permission cache.
	lastCheck, err := p.checkPermission(ctx, image, fetcherOption)
	if err != nil {
		p.logger.Error(err, "failed to check permission for image", "image", image.String())
	}

	p.upsertPermission(&permissionCacheEntry{
		image:         image,
		fetcherOption: &fetcherOption,
		lastCheck:     lastCheck,
		checkError:    err,
		lastAccess:    time.Now(),
	})
	return err == nil, err
}

// getForTest is a test helper to get a permission cache entry from the cache.
func (p *permissionCache) getForTest(key string) (permissionCacheEntry, bool) {
	p.Lock()
	defer p.Unlock()
	entry, ok := p.cache[key]
	if !ok {
		return permissionCacheEntry{}, false
	}
	return *entry, true
}

// deleteForTest is a test helper to delete a permission cache entry from the cache.
func (p *permissionCache) deleteForTest(key string) {
	p.Lock()
	defer p.Unlock()
	delete(p.cache, key)
}
