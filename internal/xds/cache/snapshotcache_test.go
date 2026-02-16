// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cache

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

func newTestSnapshotCache(t *testing.T) *snapshotCache {
	t.Helper()
	logger := logging.DefaultLogger(os.Stderr, egv1a1.LogLevelInfo)
	cache := NewSnapshotCache(false, logger)
	return cache.(*snapshotCache)
}

// TestOnStreamResponseConcurrentAccess verifies that OnStreamResponse and
// OnStreamOpen can safely run concurrently without a data race on streamIDNodeInfo.
func TestOnStreamResponseConcurrentAccess(t *testing.T) {
	sc := newTestSnapshotCache(t)

	err := sc.OnStreamOpen(context.Background(), 1, "")
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		streamID := int64(i + 100)

		go func() {
			defer wg.Done()
			sc.OnStreamResponse(context.Background(), 1, nil, nil)
		}()

		go func(id int64) {
			defer wg.Done()
			_ = sc.OnStreamOpen(context.Background(), id, "")
		}(streamID)
	}
	wg.Wait()
}

// TestOnStreamDeltaResponseConcurrentAccess verifies the same for delta streams.
func TestOnStreamDeltaResponseConcurrentAccess(t *testing.T) {
	sc := newTestSnapshotCache(t)

	err := sc.OnDeltaStreamOpen(context.Background(), 1, "")
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		streamID := int64(i + 100)

		go func() {
			defer wg.Done()
			sc.OnStreamDeltaResponse(1, nil, nil)
		}()

		go func(id int64) {
			defer wg.Done()
			_ = sc.OnDeltaStreamOpen(context.Background(), id, "")
		}(streamID)
	}
	wg.Wait()
}
