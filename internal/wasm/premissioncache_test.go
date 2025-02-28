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
	image, _ := url.Parse(ociURLWithTag)
	secret := []byte("")

	t.Run("Cached permission should be updated", func(t *testing.T) {
		lock.Lock()
		failPermissionCheck = false
		lock.Unlock()

		ctx := context.Background()
		defer ctx.Done()
		cache, entry := setupTestPermissionCache(
			permissionCacheOptions{
				checkInterval:    10 * time.Nanosecond,
				permissionExpiry: 10 * time.Nanosecond,
			},
			image,
			secret)
		cache.Start(ctx)

		lastAccessTime := entry.lastAccess
		lastCheckTime := entry.lastCheck

		time.Sleep(1 * time.Millisecond)
		require.True(
			t,
			cache.IsAllowed(context.Background(), image, true, secret),
			"permission should be rechecked and allowed after permission expired")

		entry,ok := cache.get_test(entry.key())
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
				checkInterval:    10 * time.Nanosecond,
				permissionExpiry: 10 * time.Nanosecond,
			},
			image,
			secret)
		cache.Start(ctx)

		lastAccessTime := entry.lastAccess
		lastCheckTime := entry.lastCheck

		time.Sleep(1 * time.Millisecond)
		require.False(
			t,
			cache.IsAllowed(context.Background(), image, true, secret),
			"permission should be rechecked and denied after permission expired and secret is invalid")

		entry,ok := cache.get_test(entry.key())
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
				checkInterval: 10 * time.Nanosecond,
				cacheExpiry:   10 * time.Nanosecond,
			},
			image,
			secret)
		cache.Start(ctx)

		lastAccessTime := entry.lastAccess
		lastCheckTime := entry.lastCheck

		time.Sleep(1 * time.Millisecond)
		key := entry.key()
		entry,ok := cache.get_test(key)
		require.False(t, ok, "cache entry should be removed after expiry")
		require.True(t,
			cache.IsAllowed(context.Background(), image, true, secret),
			"permission should be rechecked and allowed after cache removed")
		entry,ok= cache.get_test(key)
		require.True(t, ok, "expired entry should be added after recheck")
		require.True(t, entry.lastAccess.After(lastAccessTime), "last access time should be updated")
		require.True(t, entry.lastCheck.After(lastCheckTime), "last check time should be updated")
	})

	t.Run("Non-exist permission should be checked and cached after first access", func(t *testing.T) {
		lock.Lock()
		failPermissionCheck = false
		lock.Unlock()

		ctx := context.Background()
		defer ctx.Done()
		cache, entry := setupTestPermissionCache(
			permissionCacheOptions{
				checkInterval: 10 * time.Nanosecond,
				cacheExpiry:   10 * time.Nanosecond,
			},
			image,
			secret)
		key := entry.key()
		// remove the cache entry
		cache.delete_test(key)
		cache.Start(ctx)

		_,ok := cache.get_test(key)
		require.False(t, ok, "cache entry should not exist before access")

		now := time.Now()
		require.True(t,
			cache.IsAllowed(context.Background(), image, true, secret),
			"non-exist permission should be checked and allowed at first access")

		entry,ok =cache.get_test(key)
		require.True(t, ok, "non-exist permission should be added to the cache after first access ")
		require.True(t, entry.lastAccess.After(now), "last access time should be updated after first access")
		require.True(t, entry.lastCheck.After(now), "last check time should be updated after first access")
	})
}

// setupTestPermissionCache sets up a permission cache for testing.
func setupTestPermissionCache(options permissionCacheOptions, image *url.URL, secret []byte) (*permissionCache, permissionCacheEntry) {
	// Setup the permission cache.
	cache := newPermissionCache(
		options,
		logging.DefaultLogger(egv1a1.LogLevelInfo))

	now := time.Now()
	entry := &permissionCacheEntry{
		image: image,
		fetcherOption: &ImageFetcherOption{
			PullSecret: secret,
			Insecure:   true,
		},
		allowed:   true,
		lastCheck: now,
	}
	cache.Put(entry)
	return cache, *entry
}
