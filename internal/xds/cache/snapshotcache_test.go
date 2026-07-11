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

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	statusv3 "google.golang.org/genproto/googleapis/rpc/status"

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

// TestOnStreamRequestNACK verifies that a NACK from Envoy is handled
// gracefully on both the SotW and delta request paths.
func TestOnStreamRequestNACK(t *testing.T) {
	const (
		nodeID  = "test-node"
		cluster = "test-cluster"
	)

	newPrimedCache := func(t *testing.T) *snapshotCache {
		t.Helper()
		sc := newTestSnapshotCache(t)
		// A non-nil last snapshot for the node's cluster is required, otherwise the
		// request handlers return early before reaching the NACK handling.
		snap, err := cachev3.NewSnapshot("1", nil)
		require.NoError(t, err)
		sc.lastSnapshot[cluster] = snap
		return sc
	}

	node := &corev3.Node{Id: nodeID, Cluster: cluster}
	errorDetail := &statusv3.Status{Code: 13, Message: "invalid access log format"}

	t.Run("sotw", func(t *testing.T) {
		sc := newPrimedCache(t)
		require.NoError(t, sc.OnStreamOpen(context.Background(), 1, ""))
		err := sc.OnStreamRequest(1, &discoveryv3.DiscoveryRequest{
			Node:        node,
			TypeUrl:     resourcev3.ListenerType,
			ErrorDetail: errorDetail,
		})
		require.NoError(t, err)
	})

	t.Run("delta", func(t *testing.T) {
		sc := newPrimedCache(t)
		require.NoError(t, sc.OnDeltaStreamOpen(context.Background(), 1, ""))
		err := sc.OnStreamDeltaRequest(1, &discoveryv3.DeltaDiscoveryRequest{
			Node:        node,
			TypeUrl:     resourcev3.ListenerType,
			ErrorDetail: errorDetail,
		})
		require.NoError(t, err)
	})
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
