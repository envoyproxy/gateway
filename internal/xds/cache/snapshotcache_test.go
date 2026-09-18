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

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoytypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
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

// TestUpdateEndpointResources verifies that the endpoint fast path patch bumps
// only the EDS resource version, replaces only the given CLAs, keeps everything
// else version-stable, and never mutates the previous snapshot.
func TestUpdateEndpointResources(t *testing.T) {
	const irKey = "test-cluster"

	sc := newTestSnapshotCache(t)

	cla := func(name, address string) *endpointv3.ClusterLoadAssignment {
		return &endpointv3.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints: []*endpointv3.LocalityLbEndpoints{{
				LbEndpoints: []*endpointv3.LbEndpoint{{
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Address: &corev3.Address{
								Address: &corev3.Address_SocketAddress{
									SocketAddress: &corev3.SocketAddress{
										Address:       address,
										PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: 8080},
									},
								},
							},
						},
					},
				}},
			}},
		}
	}

	resources := types.XdsResources{
		resourcev3.ClusterType: {&clusterv3.Cluster{Name: "cluster-a"}, &clusterv3.Cluster{Name: "cluster-b"}},
		resourcev3.EndpointType: {
			cla("cluster-a", "1.1.1.1"),
			cla("cluster-b", "2.2.2.2"),
		},
	}
	err := sc.GenerateNewSnapshot(irKey, resources, context.Background())
	require.NoError(t, err)

	oldSnapshot := sc.lastSnapshot[irKey]
	oldEDSVersion := oldSnapshot.GetVersion(resourcev3.EndpointType)
	oldCDSVersion := oldSnapshot.GetVersion(resourcev3.ClusterType)

	// Patching an unknown IR key is a no-op.
	err = sc.UpdateEndpointResources("unknown", []envoytypes.Resource{cla("cluster-a", "3.3.3.3")})
	require.ErrorIs(t, err, ErrNoSnapshot)

	require.NoError(t, sc.UpdateEndpointResources(irKey, []envoytypes.Resource{cla("cluster-a", "3.3.3.3")}))

	newSnapshot := sc.lastSnapshot[irKey]
	require.NotSame(t, oldSnapshot, newSnapshot)

	// Only the EDS version is bumped.
	require.NotEqual(t, oldEDSVersion, newSnapshot.GetVersion(resourcev3.EndpointType))
	require.Equal(t, oldCDSVersion, newSnapshot.GetVersion(resourcev3.ClusterType))

	// The patched CLA is replaced, the other one is kept.
	eds := newSnapshot.GetResources(resourcev3.EndpointType)
	require.Len(t, eds, 2)
	gotA := eds["cluster-a"].(*endpointv3.ClusterLoadAssignment)
	require.Equal(t, "3.3.3.3", gotA.GetEndpoints()[0].GetLbEndpoints()[0].GetEndpoint().GetAddress().GetSocketAddress().GetAddress())
	gotB := eds["cluster-b"].(*endpointv3.ClusterLoadAssignment)
	require.Equal(t, "2.2.2.2", gotB.GetEndpoints()[0].GetLbEndpoints()[0].GetEndpoint().GetAddress().GetSocketAddress().GetAddress())

	// The previous snapshot is untouched.
	oldA := oldSnapshot.GetResources(resourcev3.EndpointType)["cluster-a"].(*endpointv3.ClusterLoadAssignment)
	require.Equal(t, "1.1.1.1", oldA.GetEndpoints()[0].GetLbEndpoints()[0].GetEndpoint().GetAddress().GetSocketAddress().GetAddress())

	// The patched snapshot carries no delta version map, so go-control-plane
	// rebuilds it on demand and it reflects the patched endpoints.
	require.Nil(t, newSnapshot.VersionMap)
	require.NoError(t, newSnapshot.ConstructVersionMap())
	require.NoError(t, oldSnapshot.ConstructVersionMap())
	oldEDSVM := oldSnapshot.GetVersionMap(resourcev3.EndpointType)
	newEDSVM := newSnapshot.GetVersionMap(resourcev3.EndpointType)
	require.NotEqual(t, oldEDSVM["cluster-a"], newEDSVM["cluster-a"])
	require.Equal(t, oldEDSVM["cluster-b"], newEDSVM["cluster-b"])
}
