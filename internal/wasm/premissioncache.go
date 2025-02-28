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
	"net/url"
	"sync"
	"time"

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
}

// validate validates the permission cache options.
func (o *permissionCacheOptions) validate() {
	if o.checkInterval == 0 {
		o.checkInterval = 5 * time.Minute
	}
	if o.permissionExpiry == 0 {
		o.permissionExpiry = 1 * time.Hour
	}
	if o.cacheExpiry == 0 {
		o.cacheExpiry = 24 * time.Hour
	}
}

// permissionCache is a cache for permission check for private OCI images.
// After a new permission is put into the cache, it will be checked periodically by a background goroutine.
// It is used to avoid blocking the translator due to the permission check.
type permissionCache struct {
	sync.Mutex
	permissionCacheOptions

	cache  map[string]*permissionCacheEntry // key: sha256(imageURL + pullSecret), value: permissionCacheEntry
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
	// Whether the permission is allowed.
	allowed bool
	// The last time the cache entry is accessed.
	lastAccess time.Time
}

// key generates a key for a permission cache entry.
// The key is a sha256 hash of the image URL and the pull secret.
func (e *permissionCacheEntry) key() string {
	return permissionCacheKey(e.image, e.fetcherOption.PullSecret)
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

func permissionCacheKey(image *url.URL, pullSecret []byte) string {
	b := make([]byte, len(image.String())+len(pullSecret))
	copy(b, image.String())
	copy(b[len(image.String()):], pullSecret)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// newPermissionCache creates a new permission cache with a given TTL.
func newPermissionCache(options permissionCacheOptions, logger logging.Logger) *permissionCache {
	options.validate()
	return &permissionCache{
		cache:                  make(map[string]*permissionCacheEntry),
		permissionCacheOptions: options,
		logger:                 logger,
	}
}

// checkAndUpdatePermission checks the permission of the image against the pull secret and updates the cache entry.
func (p *permissionCache) checkAndUpdatePermission(ctx context.Context, e *permissionCacheEntry) {
	fetcher := NewImageFetcher(ctx, *e.fetcherOption, p.logger)
	if _, _, err := fetcher.PrepareFetch(e.image.Host + e.image.Path); err != nil {
		// TDOO: check if the error is due to permission issue.
		p.logger.Error(err, "failed to check permission for image", "image", e.image.String())
		e.allowed = false
	} else {
		e.allowed = true
	}
	e.lastCheck = time.Now()
}

// start starts a background goroutine to periodically check the permission for the cached permission entries.
func (p *permissionCache) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(p.checkInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				func() {
					p.Lock()
					defer p.Unlock()
					for _, e := range p.cache {
						if e.isCacheExpired(p.cacheExpiry) {
							p.logger.Info("removing permission cache entry", "image", e.image.String())
							delete(p.cache, e.key())
							continue
						}
						if e.isPermissionExpired(p.permissionExpiry) {
							p.logger.Info("rechecking permission for image", "image", e.image.String())
							p.checkAndUpdatePermission(ctx, e)
						}
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()
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
func (p *permissionCache) IsAllowed(ctx context.Context, image *url.URL, insecure bool, pullSecret []byte) bool {
	p.Lock()
	defer p.Unlock()
	key := permissionCacheKey(image, pullSecret)
	if e, ok := p.cache[key]; ok {
		e.lastAccess = time.Now()
		return e.allowed
	}

	e := &permissionCacheEntry{
		image: image,
		fetcherOption: &ImageFetcherOption{
			Insecure:   insecure,
			PullSecret: pullSecret,
		},
	}
	p.checkAndUpdatePermission(ctx, e)
	e.lastAccess = time.Now()
	p.cache[key] = e
	return e.allowed
}

// get_test is a test helper to get a permission cache entry from the cache.
func (p *permissionCache) get_test(key string) (permissionCacheEntry, bool) {
	p.Lock()
	defer p.Unlock()
	entry, ok := p.cache[key]
	if !ok {
		return permissionCacheEntry{}, false
	}
	return *entry, true
}

// delete_test is a test helper to delete a permission cache entry from the cache.
func (p *permissionCache) delete_test(key string) {
	p.Lock()
	defer p.Unlock()
	delete(p.cache, key)
}
