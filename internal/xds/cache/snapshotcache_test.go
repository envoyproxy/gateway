// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cache

import (
	"io"
	"testing"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

func newTestCache(t *testing.T) *snapshotCache {
	t.Helper()
	logger := logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo)
	return NewSnapshotCache(true, logger).(*snapshotCache)
}

func TestPruneStaleStreamsRemovesInactiveEntries(t *testing.T) {
	cache := newTestCache(t)
	now := time.Now()

	cache.streamIDNodeInfo[1] = &corev3.Node{Id: "node-1", Cluster: "cluster-1"}
	cache.streamDuration[1] = now.Add(-2 * staleStreamThreshold)
	cache.streamLastActivity[1] = now.Add(-2 * staleStreamThreshold)
	cache.nodeFrequency["node-1"] = 1

	cache.pruneStaleStreams(now)

	_, ok := cache.streamDuration[1]
	require.False(t, ok, "stale stream should be pruned")
	_, ok = cache.streamLastActivity[1]
	require.False(t, ok, "last activity should be pruned")
	_, ok = cache.nodeFrequency["node-1"]
	require.False(t, ok, "node frequency should be cleared when last stream pruned")
}

func TestPruneStaleStreamsKeepsActiveEntries(t *testing.T) {
	cache := newTestCache(t)
	now := time.Now()

	cache.streamIDNodeInfo[1] = &corev3.Node{Id: "node-1", Cluster: "cluster-1"}
	cache.streamDuration[1] = now.Add(-2 * staleStreamThreshold)
	cache.streamLastActivity[1] = now // recent activity
	cache.nodeFrequency["node-1"] = 1

	cache.pruneStaleStreams(now)

	_, ok := cache.streamDuration[1]
	require.True(t, ok, "active stream should not be pruned")
	_, ok = cache.nodeFrequency["node-1"]
	require.True(t, ok, "node frequency should remain for active stream")
}

func TestPruneStaleDeltaStreamsRemovesInactiveEntries(t *testing.T) {
	cache := newTestCache(t)
	now := time.Now()

	cache.streamIDNodeInfo[2] = &corev3.Node{Id: "node-2", Cluster: "cluster-1"}
	cache.deltaStreamDuration[2] = now.Add(-2 * staleStreamThreshold)
	cache.deltaLastActivity[2] = now.Add(-2 * staleStreamThreshold)
	cache.nodeFrequency["node-2"] = 1

	cache.pruneStaleStreams(now)

	_, ok := cache.deltaStreamDuration[2]
	require.False(t, ok, "stale delta stream should be pruned")
	_, ok = cache.nodeFrequency["node-2"]
	require.False(t, ok, "node frequency should be cleared for delta stream")
}

func TestPruneStaleDeltaStreamsKeepsActiveEntries(t *testing.T) {
	cache := newTestCache(t)
	now := time.Now()

	cache.streamIDNodeInfo[2] = &corev3.Node{Id: "node-2", Cluster: "cluster-1"}
	cache.deltaStreamDuration[2] = now.Add(-2 * staleStreamThreshold)
	cache.deltaLastActivity[2] = now // recent activity
	cache.nodeFrequency["node-2"] = 1

	cache.pruneStaleStreams(now)

	_, ok := cache.deltaStreamDuration[2]
	require.True(t, ok, "active delta stream should not be pruned")
	_, ok = cache.nodeFrequency["node-2"]
	require.True(t, ok, "node frequency should remain for active delta stream")
}
