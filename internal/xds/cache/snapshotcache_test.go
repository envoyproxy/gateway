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
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
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

// TestOnStreamRequestNACK verifies that a NACK from Envoy is handled gracefully
// on both the SotW and delta request paths, and that the installed onNACK handler
// observes the rejection and the subsequent ACK-driven clear.
func TestOnStreamRequestNACK(t *testing.T) {
	const (
		nodeID  = "test-node"
		cluster = "test-cluster"
	)

	newPrimedCache := func(t *testing.T) (*snapshotCache, *[]NACKEvent) {
		t.Helper()
		sc := newTestSnapshotCache(t)
		// A non-nil last snapshot for the node's cluster is required, otherwise the
		// request handlers return early before reaching the NACK handling. It carries a
		// Listener so GetVersion(ListenerType) returns the snapshot version ("1"), which
		// the clear path matches against the proxy's reported version_info.
		snap, err := cachev3.NewSnapshot("1", map[resourcev3.Type][]types.Resource{
			resourcev3.ListenerType: {&listenerv3.Listener{Name: "test-listener"}},
		})
		require.NoError(t, err)
		sc.lastSnapshot[cluster] = snap

		var events []NACKEvent
		sc.SetNACKHandler(func(e NACKEvent) {
			events = append(events, e)
		})
		return sc, &events
	}

	node := &corev3.Node{Id: nodeID, Cluster: cluster}
	errorDetail := &statusv3.Status{Code: 13, Message: "invalid access log format"}

	t.Run("sotw", func(t *testing.T) {
		sc, events := newPrimedCache(t)
		require.NoError(t, sc.OnStreamOpen(context.Background(), 1, ""))

		// A NACK should surface the rejection to the handler.
		err := sc.OnStreamRequest(1, &discoveryv3.DiscoveryRequest{
			Node:          node,
			TypeUrl:       resourcev3.ListenerType,
			ResponseNonce: "1",
			ErrorDetail:   errorDetail,
		})
		require.NoError(t, err)
		require.Len(t, *events, 1)
		require.Equal(t, NACKEvent{
			IRKey:   cluster,
			NodeID:  nodeID,
			TypeURL: resourcev3.ListenerType,
			Code:    13,
			Message: "invalid access log format",
			Version: "1",
		}, (*events)[0])

		// A clean ACK of a *stale* version (not the latest pushed) must NOT clear: the
		// proxy is still catching up and may yet reject the newest config.
		err = sc.OnStreamRequest(1, &discoveryv3.DiscoveryRequest{
			Node:          node,
			TypeUrl:       resourcev3.ListenerType,
			ResponseNonce: "2",
			VersionInfo:   "0",
		})
		require.NoError(t, err)
		require.Len(t, *events, 1)

		// A clean ACK of the latest pushed version (carrying a nonce, no error) clears it.
		err = sc.OnStreamRequest(1, &discoveryv3.DiscoveryRequest{
			Node:          node,
			TypeUrl:       resourcev3.ListenerType,
			ResponseNonce: "2",
			VersionInfo:   "1",
		})
		require.NoError(t, err)
		require.Len(t, *events, 2)
		require.Equal(t, NACKEvent{IRKey: cluster, NodeID: nodeID, TypeURL: resourcev3.ListenerType, Code: 0, Version: "1"}, (*events)[1])
	})

	t.Run("delta", func(t *testing.T) {
		sc, events := newPrimedCache(t)
		require.NoError(t, sc.OnDeltaStreamOpen(context.Background(), 1, ""))

		err := sc.OnStreamDeltaRequest(1, &discoveryv3.DeltaDiscoveryRequest{
			Node:          node,
			TypeUrl:       resourcev3.ListenerType,
			ResponseNonce: "1",
			ErrorDetail:   errorDetail,
		})
		require.NoError(t, err)
		require.Len(t, *events, 1)
		require.Equal(t, NACKEvent{
			IRKey:   cluster,
			NodeID:  nodeID,
			TypeURL: resourcev3.ListenerType,
			Code:    13,
			Message: "invalid access log format",
			Version: "1",
		}, (*events)[0])

		// Delta requests carry no version_info, so the clear is scoped by node+type+nonce
		// only (no version gate).
		err = sc.OnStreamDeltaRequest(1, &discoveryv3.DeltaDiscoveryRequest{
			Node:          node,
			TypeUrl:       resourcev3.ListenerType,
			ResponseNonce: "2",
		})
		require.NoError(t, err)
		require.Len(t, *events, 2)
		require.Equal(t, NACKEvent{IRKey: cluster, NodeID: nodeID, TypeURL: resourcev3.ListenerType, Code: 0}, (*events)[1])
	})

	// The first request on a stream has no response nonce: it is not an ACK of any
	// pushed config, so it must not emit a spurious clear that could wipe a legitimate
	// NACK recorded for this irKey by another proxy.
	t.Run("no clear without nonce", func(t *testing.T) {
		sc, events := newPrimedCache(t)
		require.NoError(t, sc.OnStreamOpen(context.Background(), 1, ""))

		err := sc.OnStreamRequest(1, &discoveryv3.DiscoveryRequest{
			Node:    node,
			TypeUrl: resourcev3.ListenerType,
		})
		require.NoError(t, err)
		require.Empty(t, *events)
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
