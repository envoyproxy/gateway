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
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
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

	t.Run("Cached permission should be updated after expiry", func(t *testing.T) {
		require.Eventually(t, func() bool {
			lock.Lock()
			failPermissionCheck = false
			lock.Unlock()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
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

			// Wait for the cache to expire
			time.Sleep(10 * time.Millisecond)
			allowed, err := cache.IsAllowed(context.Background(), image, secret, true, nil)
			if err != nil {
				t.Logf("permission rechecked should not return error: %v", err)
				return false
			}
			if !allowed {
				t.Log("ppermission should be rechecked and allowed after permission expired")
				return false
			}
			entry, ok := cache.getForTest(entry.key())
			if !ok {
				t.Log("cache entry should exist")
				return false
			}
			// Verify the cached image permission is rechecked
			if !entry.lastAccess.After(lastAccessTime) {
				t.Log("last access time should be updated")
				return false
			}
			if !entry.lastCheck.After(lastCheckTime) {
				t.Log("last check time should be updated")
				return false
			}
			return true
		}, time.Second*5, time.Millisecond*20)
	})

	t.Run("Cached permission failed after recheck", func(t *testing.T) {
		require.Eventually(t, func() bool {
			lock.Lock()
			failPermissionCheck = true
			lock.Unlock()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
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

			// Wait for the cache to expire
			time.Sleep(10 * time.Millisecond)
			allowed, err := cache.IsAllowed(context.Background(), image, secret, true, nil)
			if err == nil {
				t.Log("permission rechecked should return error if failed permission check")
				return false
			}
			if isRetriableError(err) {
				t.Logf("permission check error should not be retriable: %v", err)
				return false
			}
			if allowed {
				t.Log("permission should be rechecked and denied after permission expired and secret is invalid")
				return false
			}
			entry, ok := cache.getForTest(entry.key())
			if !ok {
				t.Log("cache entry should exist")
				return false
			}
			if !entry.lastAccess.After(lastAccessTime) {
				t.Log("last access time should be updated")
				return false
			}
			if !entry.lastCheck.After(lastCheckTime) {
				t.Log("last check time should be updated")
				return false
			}
			return true
		}, time.Second*5, time.Millisecond*20)
	})

	t.Run("Cached permission should be removed after expiry", func(t *testing.T) {
		require.Eventually(t, func() bool {
			lock.Lock()
			failPermissionCheck = false
			lock.Unlock()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
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

			// Wait for the cache to expire
			time.Sleep(10 * time.Millisecond)
			key := entry.key()
			entry, ok := cache.getForTest(key)
			if ok {
				t.Log("cache entry should be removed after expiry")
				return false
			}
			allowed, err := cache.IsAllowed(context.Background(), image, secret, true, nil)
			if err != nil {
				t.Logf("permission rechecked should not return error: %v", err)
				return false
			}
			if !allowed {
				t.Log("permission should be rechecked and allowed after cache removed")
				return false
			}
			entry, ok = cache.getForTest(key)
			if !ok {
				t.Log("expired entry should be added after recheck")
				return false
			}
			if !entry.lastAccess.After(lastAccessTime) {
				t.Log("last access time should be updated")
				return false
			}
			if !entry.lastCheck.After(lastCheckTime) {
				t.Log("last check time should be updated")
				return false
			}
			return true
		}, time.Second*5, time.Millisecond*20)
	})

	t.Run("Non-exist permission should be checked and cached after first access for allowed permission", func(t *testing.T) {
		require.Eventually(t, func() bool {
			lock.Lock()
			failPermissionCheck = false
			lock.Unlock()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
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
			if ok {
				t.Log("cache entry should not exist before access")
				return false
			}

			now := time.Now()
			allowed, err := cache.IsAllowed(context.Background(), image, secret, true, nil)
			if err != nil {
				t.Logf("permission check should not return error: %v", err)
				return false
			}
			if !allowed {
				t.Log("non-exist permission should be checked and allowed at first access")
				return false
			}

			entry, ok = cache.getForTest(key)
			if !ok {
				t.Log("non-exist permission should be added to the cache after first access")
				return false
			}
			if !entry.lastAccess.After(now) {
				t.Log("last access time should be updated after first access")
				return false
			}
			if !entry.lastCheck.After(now) {
				t.Log("last check time should be updated after first access")
				return false
			}
			return true
		}, time.Second*5, time.Millisecond*20)
	})

	t.Run("Non-exist permission should be checked and cached after first access for denied permission", func(t *testing.T) {
		require.Eventually(t, func() bool {
			lock.Lock()
			failPermissionCheck = true
			lock.Unlock()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
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
			if ok {
				t.Log("cache entry should not exist before access")
				return false
			}

			now := time.Now()
			allowed, err := cache.IsAllowed(context.Background(), image, secret, true, nil)
			if err == nil {
				t.Logf("non-exist permission should be checked and denied at first access if secret is invalid")
				return false
			}
			if allowed {
				t.Log("non-exist permission should be checked and denied at first access if secret is invalid")
				return false
			}

			entry, ok = cache.getForTest(key)
			if !ok {
				t.Log("non-exist permission should be added to the cache after first access")
				return false
			}
			if !entry.lastAccess.After(now) {
				t.Log("last access time should be updated after first access")
				return false
			}
			if !entry.lastCheck.After(now) {
				t.Log("last check time should be updated after first access")
				return false
			}
			return true
		}, time.Second*5, time.Millisecond*20)
	})
}

func TestPermissionCacheDoesNotHoldLockDuringMissCheck(t *testing.T) {
	blockingImage, started, cleanup := setupBlockingRegistry(t)
	defer cleanup()

	cache := newPermissionCache(
		permissionCacheOptions{},
		logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))

	cachedImage, err := url.Parse("oci://example.com/test/cached:v1")
	require.NoError(t, err)
	cache.Put(&permissionCacheEntry{
		image: cachedImage,
		fetcherOption: &ImageFetcherOption{
			PullSecret: nil,
		},
		checkError: nil,
	})
	cache.Lock()
	cache.cache[permissionCacheKey(cachedImage, nil, nil)].lastCheck = time.Now().Add(time.Hour)
	cache.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	go func() {
		_, _ = cache.IsAllowed(ctx, blockingImage, nil, true, nil)
	}()

	require.Eventually(t, func() bool {
		return started.Load()
	}, time.Second, 10*time.Millisecond)
	requireCachedPermissionReturns(t, cache, cachedImage)
}

func TestPermissionCacheDoesNotHoldLockDuringBackgroundCheck(t *testing.T) {
	blockingImage, started, cleanup := setupBlockingRegistry(t)
	defer cleanup()

	cache := newPermissionCache(
		permissionCacheOptions{
			checkInterval:          10 * time.Millisecond,
			permissionExpiry:       time.Nanosecond,
			permissionCheckTimeout: 200 * time.Millisecond,
		},
		logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))
	cache.Put(&permissionCacheEntry{
		image: blockingImage,
		fetcherOption: &ImageFetcherOption{
			Insecure: true,
		},
		checkError: nil,
	})

	cachedImage, err := url.Parse("oci://example.com/test/cached:v1")
	require.NoError(t, err)
	cache.Put(&permissionCacheEntry{
		image: cachedImage,
		fetcherOption: &ImageFetcherOption{
			PullSecret: nil,
		},
		checkError: nil,
	})
	cache.Lock()
	cache.cache[permissionCacheKey(cachedImage, nil, nil)].lastCheck = time.Now().Add(time.Hour)
	cache.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cache.Start(ctx)

	require.Eventually(t, func() bool {
		return started.Load()
	}, time.Second, 10*time.Millisecond)
	requireCachedPermissionReturns(t, cache, cachedImage)
}

func TestPermissionCacheUpdateRecheckedPermission(t *testing.T) {
	image, err := url.Parse("oci://example.com/test/cached:v1")
	require.NoError(t, err)

	cache := newPermissionCache(
		permissionCacheOptions{},
		logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))

	originalLastCheck := time.Now().Add(-time.Hour)
	originalLastAccess := time.Now().Add(-30 * time.Minute)
	entry := &permissionCacheEntry{
		image: image,
		fetcherOption: &ImageFetcherOption{
			PullSecret: nil,
		},
		lastCheck:  originalLastCheck,
		lastAccess: originalLastAccess,
		checkError: nil,
	}
	cache.cache[entry.key()] = entry

	recheckErr := errors.New("recheck failed")
	recheckedEntry := entry.copy()
	recheckedEntry.lastCheck = time.Now()
	recheckedEntry.lastAccess = originalLastAccess.Add(-time.Minute)
	recheckedEntry.checkError = recheckErr

	cache.updateRecheckedPermission(recheckedEntry, originalLastCheck)

	got, ok := cache.getForTest(entry.key())
	require.True(t, ok)
	require.ErrorIs(t, got.checkError, recheckErr)
	require.True(t, got.lastCheck.Equal(recheckedEntry.lastCheck))
	require.True(t, got.lastAccess.Equal(originalLastAccess))
}

func TestPermissionCacheSkipsStaleRecheckResult(t *testing.T) {
	image, err := url.Parse("oci://example.com/test/cached:v1")
	require.NoError(t, err)

	cache := newPermissionCache(
		permissionCacheOptions{},
		logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo))

	originalLastCheck := time.Now().Add(-time.Hour)
	liveLastCheck := time.Now()
	liveLastAccess := time.Now().Add(-30 * time.Minute)
	liveErr := errors.New("live result")
	entry := &permissionCacheEntry{
		image: image,
		fetcherOption: &ImageFetcherOption{
			PullSecret: nil,
		},
		lastCheck:  liveLastCheck,
		lastAccess: liveLastAccess,
		checkError: liveErr,
	}
	cache.cache[entry.key()] = entry

	staleErr := errors.New("stale result")
	staleEntry := entry.copy()
	staleEntry.lastCheck = liveLastCheck.Add(time.Minute)
	staleEntry.lastAccess = liveLastAccess.Add(-time.Minute)
	staleEntry.checkError = staleErr

	cache.updateRecheckedPermission(staleEntry, originalLastCheck)

	got, ok := cache.getForTest(entry.key())
	require.True(t, ok)
	require.ErrorIs(t, got.checkError, liveErr)
	require.True(t, got.lastCheck.Equal(liveLastCheck))
	require.True(t, got.lastAccess.Equal(liveLastAccess))
}

func setupBlockingRegistry(t *testing.T) (*url.URL, *atomic.Bool, func()) {
	t.Helper()

	started := &atomic.Bool{}
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		started.Store(true)
		<-r.Context().Done()
	}))
	u, err := url.Parse(server.URL)
	require.NoError(t, err)
	image, err := url.Parse(fmt.Sprintf("oci://%s/test/block:v1", u.Host))
	require.NoError(t, err)

	return image, started, server.Close
}

func requireCachedPermissionReturns(t *testing.T, cache *permissionCache, image *url.URL) {
	t.Helper()

	done := make(chan error, 1)
	go func() {
		allowed, err := cache.IsAllowed(context.Background(), image, nil, false, nil)
		if err != nil {
			done <- err
			return
		}
		if !allowed {
			done <- fmt.Errorf("permission should be allowed")
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("cached permission lookup blocked while another permission check was in flight")
	}
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
