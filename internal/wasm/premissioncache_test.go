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
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestPermissionCache(t *testing.T) {
	lock := sync.Mutex{}
	// Flag to control whether the permission check should fail.
	failPermissionCheck := false

	// Set up a fake registry for OCI images.
	reg := registry.New()
	tos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		if failPermissionCheck {
			http.Error(w, "permission denied", http.StatusUnauthorized)
			return
		}
		reg.ServeHTTP(w, r)
	}))
	defer tos.Close()
	ou, err := url.Parse(tos.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = setupOCIRegistry(t, ou.Host)
	ociURLWithTag := fmt.Sprintf("oci://%s/test/valid/docker:v0.1.0", ou.Host)
	ociURLWithLatestTag := fmt.Sprintf("oci://%s/test/valid/docker:latest", ou.Host)
	image, _ := url.Parse(ociURLWithTag)
	latestImage, _ := url.Parse(ociURLWithLatestTag)
	secret := []byte("")

	t.Run("Cached permission should be updated", func(t *testing.T) {
		lock.Lock()
		failPermissionCheck = false
		lock.Unlock()

		ctx := context.Background()
		defer ctx.Done()
		cache, entry := setupTestPermissionCache(
			permissionCacheOptions{
				checkInterval:    1 * time.Nanosecond,
				permissionExpiry: 10 * time.Nanosecond,
			},
			image,
			latestImage,
			secret)
		cache.Start(ctx)

		lastAccessTime := entry.lastAccess
		lastCheckTime := entry.lastCheck

		time.Sleep(10 * time.Millisecond)
		allowed, err := cache.IsAllowed(context.Background(), image, secret, true)
		require.True(
			t,
			allowed,
			"permission should be rechecked and allowed after permission expired")
		require.NoError(
			t,
			err,
			"permission should be rechecked and allowed after permission expired")

		entry, ok := cache.getForTest(entry.key())
		require.True(t, ok, "cache entry should exist")
		require.True(t, entry.lastAccess.After(lastAccessTime), "last access time should be updated")
		require.True(t, entry.lastCheck.After(lastCheckTime), "last check time should be updated")
	})

	t.Run("Cached permission failed after recheck", func(t *testing.T) {
		lock.Lock()
		failPermissionCheck = true
		lock.Unlock()

		ctx := context.Background()
		defer ctx.Done()
		cache, entry := setupTestPermissionCache(
			permissionCacheOptions{
				checkInterval:    1 * time.Nanosecond,
				permissionExpiry: 10 * time.Nanosecond,
			},
			image,
			latestImage,
			secret)
		cache.Start(ctx)

		lastAccessTime := entry.lastAccess
		lastCheckTime := entry.lastCheck

		time.Sleep(10 * time.Millisecond)
		allowed, err := cache.IsAllowed(context.Background(), image, secret, true)
		require.False(t, isRetriableError(err), "permission check error should not be retriable")
		require.False(
			t,
			allowed,
			"permission should be rechecked and denied after permission expired and secret is invalid")
		require.Error(
			t,
			err,
			"permission should be rechecked and denied after permission expired and secret is invalid")

		entry, ok := cache.getForTest(entry.key())
		require.True(t, ok, "cache entry should exist")
		require.True(t, entry.lastAccess.After(lastAccessTime), "last access time should be updated")
		require.True(t, entry.lastCheck.After(lastCheckTime), "last check time should be updated")
	})

	t.Run("Cached permission should be removed after expiry", func(t *testing.T) {
		lock.Lock()
		failPermissionCheck = false
		lock.Unlock()

		ctx := context.Background()
		defer ctx.Done()
		cache, entry := setupTestPermissionCache(
			permissionCacheOptions{
				checkInterval: 1 * time.Nanosecond,
				cacheExpiry:   10 * time.Nanosecond,
			},
			image,
			latestImage,
			secret)
		cache.Start(ctx)

		lastAccessTime := entry.lastAccess
		lastCheckTime := entry.lastCheck

		time.Sleep(10 * time.Millisecond)
		key := entry.key()
		entry, ok := cache.getForTest(key)
		require.False(t, ok, "cache entry should be removed after expiry")
		allowed, err := cache.IsAllowed(context.Background(), image, secret, true)
		require.True(t,
			allowed,
			"permission should be rechecked and allowed after cache removed")
		require.NoError(t,
			err,
			"permission should be rechecked and allowed after cache removed")
		entry, ok = cache.getForTest(key)
		require.True(t, ok, "expired entry should be added after recheck")
		require.True(t, entry.lastAccess.After(lastAccessTime), "last access time should be updated")
		require.True(t, entry.lastCheck.After(lastCheckTime), "last check time should be updated")
	})

	t.Run("Non-exist permission should be checked and cached after first access for allowed permission", func(t *testing.T) {
		lock.Lock()
		failPermissionCheck = false
		lock.Unlock()

		ctx := context.Background()
		defer ctx.Done()
		cache, entry := setupTestPermissionCache(
			permissionCacheOptions{
				checkInterval: 1 * time.Nanosecond,
			},
			image,
			latestImage,
			secret)
		key := entry.key()
		// remove the cache entry
		cache.deleteForTest(key)
		cache.Start(ctx)

		_, ok := cache.getForTest(key)
		require.False(t, ok, "cache entry should not exist before access")

		now := time.Now()
		allowed, err := cache.IsAllowed(context.Background(), image, secret, true)
		require.True(t,
			allowed,
			"non-exist permission should be checked and allowed at first access")
		require.NoError(t,
			err,
			"non-exist permission should be checked and allowed at first access")

		entry, ok = cache.getForTest(key)
		require.True(t, ok, "non-exist permission should be added to the cache after first access ")
		require.True(t, entry.lastAccess.After(now), "last access time should be updated after first access")
		require.True(t, entry.lastCheck.After(now), "last check time should be updated after first access")
	})

	t.Run("Non-exist permission should be checked and cached after first access for denied permission", func(t *testing.T) {
		lock.Lock()
		failPermissionCheck = true
		lock.Unlock()

		ctx := context.Background()
		defer ctx.Done()
		cache, entry := setupTestPermissionCache(
			permissionCacheOptions{
				checkInterval: 1 * time.Nanosecond,
			},
			image,
			latestImage,
			secret)
		key := entry.key()
		// remove the cache entry
		cache.deleteForTest(key)
		cache.Start(ctx)

		_, ok := cache.getForTest(key)
		require.False(t, ok, "cache entry should not exist before access")

		now := time.Now()
		allowed, err := cache.IsAllowed(context.Background(), image, secret, true)
		require.False(t,
			allowed,
			"non-exist permission should be checked and denied at first access if secret is invalid")
		require.Error(t,
			err,
			"non-exist permission should be checked and denied at first access if secret is invalid")

		entry, ok = cache.getForTest(key)
		require.True(t, ok, "non-exist permission should be added to the cache after first access ")
		require.True(t, entry.lastAccess.After(now), "last access time should be updated after first access")
		require.True(t, entry.lastCheck.After(now), "last check time should be updated after first access")
	})
}

// setupTestPermissionCache sets up a permission cache for testing.
func setupTestPermissionCache(options permissionCacheOptions, image, latestImage *url.URL, secret []byte) (*permissionCache, permissionCacheEntry) {
	// Setup the permission cache.
	cache := newPermissionCache(
		options,
		logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))

	entry := &permissionCacheEntry{
		image: image,
		fetcherOption: &ImageFetcherOption{
			PullSecret: secret,
			Insecure:   true,
		},
		lastCheck: time.Now(),
	}
	cache.Put(entry)

	// Add one more entry for the latest image to test the cache can handle multiple entries correctly.
	cache.Put(&permissionCacheEntry{
		image: latestImage,
		fetcherOption: &ImageFetcherOption{
			PullSecret: secret,
			Insecure:   true,
		},
		lastCheck: time.Now(),
	})

	return cache, *entry
}
