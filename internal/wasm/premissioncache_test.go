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
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestPermissionCache(t *testing.T) {
	// Flag to control whether the permission check should fail.
	failPermissionCheck := false

	reg := registry.New()
	// Set up a fake registry for OCI images.
	tos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	t.Run("Cached permission should be updated", func(t *testing.T) {
		failPermissionCheck = false
		cache := newPermissionCache(1*time.Millisecond, logging.DefaultLogger(egv1a1.LogLevelInfo))
		ctx := context.Background()
		defer ctx.Done()
		cache.Start(ctx)
		image, _ := url.Parse(ociURLWithTag)
		secret := []byte("")
		first := time.Now()
		entry := permissionCacheEntry{
			image: image,
			fetcherOption: &ImageFetcherOption{
				PullSecret: secret,
				Insecure:   true,
			},
			lastCheck: first,
			allowed:   true,
		}
		cache.Put(&entry)
		time.Sleep(3 * time.Millisecond)
		allow, err := cache.Allow(image, secret)
		require.NoError(t, err)
		require.True(t, allow)
		require.True(t, entry.lastCheck.After(first))
	})

	t.Run("Cached permission failed after recheck", func(t *testing.T) {
		failPermissionCheck = true
		cache := newPermissionCache(1*time.Millisecond, logging.DefaultLogger(egv1a1.LogLevelInfo))
		ctx := context.Background()
		defer ctx.Done()
		cache.Start(ctx)
		image, _ := url.Parse(ociURLWithTag)
		secret := []byte("")
		first := time.Now()
		entry := permissionCacheEntry{
			image: image,
			fetcherOption: &ImageFetcherOption{
				PullSecret: secret,
				Insecure:   true,
			},
			lastCheck: first,
			allowed:   true,
		}
		cache.Put(&entry)
		time.Sleep(3 * time.Millisecond)
		allow, err := cache.Allow(image, secret)
		require.NoError(t, err)
		require.False(t, allow)
		require.True(t, entry.lastCheck.After(first))
	})
}
