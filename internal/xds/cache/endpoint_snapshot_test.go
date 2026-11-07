// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cache

import (
	"fmt"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestReplaceEndpointsPreservesOtherResourcesAndHashes(t *testing.T) {
	first := &endpointv3.ClusterLoadAssignment{ClusterName: "one"}
	second := &endpointv3.ClusterLoadAssignment{ClusterName: "two"}
	route := &routev3.RouteConfiguration{Name: "route"}
	original, err := PrepareSnapshot(types.XdsResources{
		resourcev3.EndpointType: {first, second}, resourcev3.RouteType: {route},
	})
	require.NoError(t, err)
	newAssignment := &endpointv3.ClusterLoadAssignment{ClusterName: "one", Endpoints: []*endpointv3.LocalityLbEndpoints{{LoadBalancingWeight: wrapperspb.UInt32(2)}}}
	updated, err := ReplaceEndpoints(original, map[string]*endpointv3.ClusterLoadAssignment{"one": newAssignment})
	require.NoError(t, err)
	require.NotSame(t, original, updated)
	require.Same(t, first, original.Resources[cachetypes.Endpoint].Items["one"].Resource)
	require.Same(t, newAssignment, updated.Resources[cachetypes.Endpoint].Items["one"].Resource)
	require.Same(t, second, updated.Resources[cachetypes.Endpoint].Items["two"].Resource)
	require.Same(t, route, updated.Resources[cachetypes.Route].Items["route"].Resource)
	require.Equal(t, original.Resources[cachetypes.Route].Version, updated.Resources[cachetypes.Route].Version)
	require.Equal(t, original.VersionMap[resourcev3.RouteType], updated.VersionMap[resourcev3.RouteType])
	require.NotEqual(t, original.Resources[cachetypes.Endpoint].Version, updated.Resources[cachetypes.Endpoint].Version)
	require.NotEqual(t, original.VersionMap[resourcev3.EndpointType]["one"], updated.VersionMap[resourcev3.EndpointType]["one"])
	require.Equal(t, original.VersionMap[resourcev3.EndpointType]["two"], updated.VersionMap[resourcev3.EndpointType]["two"])
	unchanged, err := ReplaceEndpoints(updated, map[string]*endpointv3.ClusterLoadAssignment{"one": newAssignment})
	require.NoError(t, err)
	require.Same(t, updated, unchanged)
}

func BenchmarkEndpointSnapshot(b *testing.B) {
	for _, count := range []int{100, 1000} {
		resources := make(types.XdsResources)
		for i := range count {
			assignment := &endpointv3.ClusterLoadAssignment{
				ClusterName: fmt.Sprintf("cluster-%d", i),
				Endpoints:   []*endpointv3.LocalityLbEndpoints{{LoadBalancingWeight: wrapperspb.UInt32(1)}},
			}
			for range 100 {
				assignment.Endpoints[0].LbEndpoints = append(assignment.Endpoints[0].LbEndpoints, &endpointv3.LbEndpoint{
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{Endpoint: &endpointv3.Endpoint{
						Address: &corev3.Address{Address: &corev3.Address_SocketAddress{SocketAddress: &corev3.SocketAddress{
							Address: "10.0.0.1", PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: 8080},
						}}},
					}},
				})
			}
			resources[resourcev3.EndpointType] = append(resources[resourcev3.EndpointType], assignment)
			resources[resourcev3.RouteType] = append(resources[resourcev3.RouteType], &routev3.RouteConfiguration{Name: fmt.Sprintf("route-%d", i)})
		}
		original, err := PrepareSnapshot(resources)
		require.NoError(b, err)
		replacement := &endpointv3.ClusterLoadAssignment{ClusterName: "cluster-0", Endpoints: []*endpointv3.LocalityLbEndpoints{{
			LoadBalancingWeight: wrapperspb.UInt32(2), LbEndpoints: resources[resourcev3.EndpointType][0].(*endpointv3.ClusterLoadAssignment).Endpoints[0].LbEndpoints,
		}}}
		b.Run(fmt.Sprintf("fast/%d_clusters", count), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, err := ReplaceEndpoints(original, map[string]*endpointv3.ClusterLoadAssignment{"cluster-0": replacement})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("full_snapshot/%d_clusters", count), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if _, err := PrepareSnapshot(resources); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
