// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
)

var (
	testListener = &listenerv3.Listener{
		Name: "test-listener",
	}
	testSecret = &tlsv3.Secret{
		Name: "test-secret",
	}
)

func TestDeepCopy(t *testing.T) {
	testCases := []struct {
		name string
		in   *ResourceVersionTable
		out  *ResourceVersionTable
	}{
		{
			name: "nil",
			in:   nil,
			out:  nil,
		},
		{
			name: "listener",
			in: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{testListener},
				},
			},
			out: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{testListener},
				},
			},
		},
		{
			name: "kitchen-sink",
			in: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{testListener},
					resourcev3.SecretType:   []types.Resource{testSecret},
				},
			},
			out: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{testListener},
					resourcev3.SecretType:   []types.Resource{testSecret},
				},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.out == nil {
				require.Nil(t, tc.in.DeepCopy())
			} else {
				diff := cmp.Diff(tc.out, tc.in.DeepCopy(), protocmp.Transform())
				require.Empty(t, diff)
			}
		})
	}
}

func TestAddOrReplaceXdsResource(t *testing.T) {
	testListener := &listenerv3.Listener{
		Name: "test-listener",
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address: "exampleservice.examplenamespace.svc.cluster.local",
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: 5000,
					},
					Protocol: corev3.SocketAddress_TCP,
				},
			},
		},
	}
	updatedListener := &listenerv3.Listener{
		Name: "test-listener",
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address: "newsvc.newns.svc.cluster.local",
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: 8000,
					},
					Protocol: corev3.SocketAddress_TCP,
				},
			},
		},
	}
	testCluster := &clusterv3.Cluster{
		Name: "test-cluster",
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: "test-cluster",
			Endpoints: []*endpointv3.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpointv3.LbEndpoint{
						{
							HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
								Endpoint: &endpointv3.Endpoint{
									Address: &corev3.Address{
										Address: &corev3.Address_SocketAddress{
											SocketAddress: &corev3.SocketAddress{
												Address: "exampleservice.examplenamespace.svc.cluster.local",
												PortSpecifier: &corev3.SocketAddress_PortValue{
													PortValue: 5000,
												},
												Protocol: corev3.SocketAddress_TCP,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	updatedCluster := &clusterv3.Cluster{
		Name: "test-cluster",
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: "test-cluster",
			Endpoints: []*endpointv3.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpointv3.LbEndpoint{
						{
							HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
								Endpoint: &endpointv3.Endpoint{
									Address: &corev3.Address{
										Address: &corev3.Address_SocketAddress{
											SocketAddress: &corev3.SocketAddress{
												Address: "modified.example.svc.cluster.local",
												PortSpecifier: &corev3.SocketAddress_PortValue{
													PortValue: 5000,
												},
												Protocol: corev3.SocketAddress_TCP,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	testCases := []struct {
		name       string
		tableIn    *ResourceVersionTable
		typeIn     resourcev3.Type
		resourceIn types.Resource
		funcIn     func(existing types.Resource, new types.Resource) bool
		tableOut   *ResourceVersionTable
	}{
		{
			name: "inject-new-cluster",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ClusterType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.ClusterType,
			resourceIn: testCluster,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldCluster := existing.(*clusterv3.Cluster)
				newCluster := new.(*clusterv3.Cluster)
				if newCluster == nil || oldCluster == nil {
					return false
				}
				if oldCluster.Name == newCluster.Name {
					return true
				}
				return false
			},
			tableOut: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ClusterType: []types.Resource{testCluster},
				},
			},
		},
		{
			name: "replace-cluster",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ClusterType: []types.Resource{testCluster},
				},
			},
			typeIn:     resourcev3.ClusterType,
			resourceIn: updatedCluster,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldCluster := existing.(*clusterv3.Cluster)
				newCluster := new.(*clusterv3.Cluster)
				if newCluster == nil || oldCluster == nil {
					return false
				}
				if oldCluster.Name == newCluster.Name {
					return true
				}
				return false
			},
			tableOut: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ClusterType: []types.Resource{updatedCluster},
				},
			},
		},
		{
			name: "inject-new-listener",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{testListener},
				},
			},
			typeIn:     resourcev3.ListenerType,
			resourceIn: testListener,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldListener := existing.(*listenerv3.Listener)
				newListener := new.(*listenerv3.Listener)
				if newListener == nil || oldListener == nil {
					return false
				}
				if oldListener.Name == newListener.Name {
					return true
				}
				return false
			},
			tableOut: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{testListener},
				},
			},
		},
		{
			name: "replace-listener",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{testListener},
				},
			},
			typeIn:     resourcev3.ListenerType,
			resourceIn: updatedListener,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldListener := existing.(*listenerv3.Listener)
				newListener := new.(*listenerv3.Listener)
				if newListener == nil || oldListener == nil {
					return false
				}
				if oldListener.Name == newListener.Name {
					return true
				}
				return false
			},
			tableOut: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{updatedListener},
				},
			},
		},
		{
			name: "inject-nil-resourcetype",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{},
			},
			typeIn:     resourcev3.ClusterType,
			resourceIn: testCluster,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldCluster := existing.(*clusterv3.Cluster)
				newCluster := new.(*clusterv3.Cluster)
				if newCluster == nil || oldCluster == nil {
					return false
				}
				if oldCluster.Name == newCluster.Name {
					return true
				}
				return false
			},
			tableOut: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ClusterType: []types.Resource{testCluster},
				},
			},
		},
		{
			name:       "inject-nil resources",
			tableIn:    &ResourceVersionTable{},
			typeIn:     resourcev3.ClusterType,
			resourceIn: testCluster,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldCluster := existing.(*clusterv3.Cluster)
				newCluster := new.(*clusterv3.Cluster)
				if newCluster == nil || oldCluster == nil {
					return false
				}
				if oldCluster.Name == newCluster.Name {
					return true
				}
				return false
			},
			tableOut: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ClusterType: []types.Resource{testCluster},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.tableIn.AddOrReplaceXdsResource(tc.typeIn, tc.resourceIn, tc.funcIn)
			diff := cmp.Diff(tc.tableOut, tc.tableIn.DeepCopy(), protocmp.Transform())
			require.Empty(t, diff)
		})
	}
}
