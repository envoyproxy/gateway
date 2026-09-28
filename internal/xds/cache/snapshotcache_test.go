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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	statusv3 "google.golang.org/genproto/googleapis/rpc/status"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/xds/types"
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

// One translation can publish up to three snapshots under one trace. No two may share a
// version: a SotW client that acks the first only after the last is published would
// otherwise see matching versions and never receive the last. The resources of the last
// one must also be readable back for the next translation to order against.
func TestGenerateNewSnapshotSequence(t *testing.T) {
	logger := logging.DefaultLogger(os.Stderr, egv1a1.LogLevelInfo)
	c := NewSnapshotCache(false, logger).(*snapshotCache)

	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{1},
		TraceFlags: trace.FlagsSampled,
	}))

	first := types.XdsResources{
		resourcev3.ExtensionConfigType: {&corev3.TypedExtensionConfig{Name: "b"}, &corev3.TypedExtensionConfig{Name: "a"}},
	}
	second := types.XdsResources{
		resourcev3.ExtensionConfigType: {&corev3.TypedExtensionConfig{Name: "a"}},
	}
	third := types.XdsResources{
		resourcev3.ExtensionConfigType: {&corev3.TypedExtensionConfig{Name: "c"}},
	}

	require.NoError(t, c.GenerateNewSnapshot("key", first, ctx))
	v1 := c.lastSnapshot["key"].GetVersion(resourcev3.ExtensionConfigType)
	require.NoError(t, c.GenerateNewSnapshot("key", second, ctx))
	v2 := c.lastSnapshot["key"].GetVersion(resourcev3.ExtensionConfigType)
	require.NoError(t, c.GenerateNewSnapshot("key", third, ctx))
	v3 := c.lastSnapshot["key"].GetVersion(resourcev3.ExtensionConfigType)

	// The first keeps the bare trace id; the others are suffixed, and all three differ.
	assert.Equal(t, trace.TraceID{1}.String(), v1)
	assert.NotEqual(t, v1, v2)
	assert.NotEqual(t, v2, v3)
	assert.NotEqual(t, v1, v3)

	last := c.LastResources("key")
	require.Len(t, last[resourcev3.ExtensionConfigType], 1)
	assert.Equal(t, "c", last[resourcev3.ExtensionConfigType][0].(*corev3.TypedExtensionConfig).Name)
	assert.Nil(t, c.LastResources("unknown"))
}
