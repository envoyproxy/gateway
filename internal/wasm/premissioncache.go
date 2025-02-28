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
	"net/url"
	"sync"
	"time"

	"github.com/envoyproxy/gateway/internal/logging"
)

// permissionCache is a cache for permission check for private OCI images.
// After a new permission is put into the cache, it will be checked periodically by a background goroutine.
// It is used to avoid blocking the translator due to the permission check.
// TODO (zhaohuabing): the cache entry is not deleted, which is not a serious issue since the cache is not expected to
// grow large. However, it is better to add a mechanism to purge the cache.
type permissionCache struct {
	sync.Mutex

	cache map[string]*permissionCacheEntry // key: sha256(imageURL + pullSecret), value: permissionCacheEntry
	// checkInterval is the interval to recheck the permission for the cached permission entries.
	checkInterval time.Duration
	// logger
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
}

// key generates a key for a permission cache entry.
// The key is a sha256 hash of the image URL and the pull secret.
func (e *permissionCacheEntry) key() string {
	return permissionCacheKey(e.image, e.fetcherOption.PullSecret)
}

func permissionCacheKey(image *url.URL, pullSecret []byte) string {
	b := make([]byte, len(image.String())+len(pullSecret))
	copy(b, image.String())
	copy(b[len(image.String()):], pullSecret)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:])
}

// newPermissionCache creates a new permission cache with a given TTL.
func newPermissionCache(interval time.Duration, logger logging.Logger) *permissionCache {
	return &permissionCache{
		cache:         make(map[string]*permissionCacheEntry),
		checkInterval: interval,
		logger:        logger,
	}
}

func (p *permissionCache) checkPermission(ctx context.Context, e *permissionCacheEntry) {
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
						if time.Now().After(e.lastCheck.Add(p.checkInterval)) {
							p.checkPermission(ctx, e)
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
	p.cache[e.key()] = e
}

// Allow checks if the given image is allowed to be accessed with the provided pull secret.
func (p *permissionCache) Allow(image *url.URL, pullSecret []byte) (bool, error) {
	p.Lock()
	defer p.Unlock()
	key := permissionCacheKey(image, pullSecret)
	if e, ok := p.cache[key]; ok {
		return e.allowed, nil
	}
	return false, errors.New("permission cache entry not found")
}
