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
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
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
	testEndpoint := &endpointv3.ClusterLoadAssignment{
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
	}
	updatedEndpoint := &endpointv3.ClusterLoadAssignment{
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
	}
	testRouteConfig := &routev3.RouteConfiguration{
		Name: "test-route-config",
		VirtualHosts: []*routev3.VirtualHost{
			{
				Name:    "test-virtual-host",
				Domains: []string{"test.example.com"},
				Routes: []*routev3.Route{
					{
						Match: &routev3.RouteMatch{
							PathSpecifier: &routev3.RouteMatch_Prefix{
								Prefix: "/",
							},
						},
						Action: &routev3.Route_Route{
							Route: &routev3.RouteAction{
								ClusterSpecifier: &routev3.RouteAction_Cluster{
									Cluster: "test-cluster",
								},
							},
						},
					},
				},
			},
		},
	}
	testSecret := &tlsv3.Secret{
		Name: "example-secret",
		Type: &tlsv3.Secret_TlsCertificate{
			TlsCertificate: &tlsv3.TlsCertificate{
				CertificateChain: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: []byte("-----BEGIN CERTIFICATE-----\n... Your certificate data ... \n-----END CERTIFICATE-----"),
					},
				},
				PrivateKey: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: []byte("-----BEGIN PRIVATE KEY-----\n... Your private key data ... \n-----END PRIVATE KEY-----"),
					},
				},
			},
		},
		// Add other fields for the secret as needed.
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
			name: "inject-new-endpoint",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.EndpointType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.EndpointType,
			resourceIn: testEndpoint,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldEndpoint := existing.(*endpointv3.ClusterLoadAssignment)
				newEndpoint := new.(*endpointv3.ClusterLoadAssignment)
				if newEndpoint == nil || oldEndpoint == nil {
					return false
				}
				if oldEndpoint.ClusterName == newEndpoint.ClusterName {
					return true
				}
				return false
			},
			tableOut: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.EndpointType: []types.Resource{testEndpoint},
				},
			},
		},
		{
			name: "replace-endpoint",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.EndpointType: []types.Resource{testEndpoint},
				},
			},
			typeIn:     resourcev3.EndpointType,
			resourceIn: updatedEndpoint,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldEndpoint := existing.(*endpointv3.ClusterLoadAssignment)
				newEndpoint := new.(*endpointv3.ClusterLoadAssignment)
				if newEndpoint == nil || oldEndpoint == nil {
					return false
				}
				if oldEndpoint.ClusterName == newEndpoint.ClusterName {
					return true
				}
				return false
			},
			tableOut: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.EndpointType: []types.Resource{updatedEndpoint},
				},
			},
		},
		{
			name: "inject-new-listener",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{},
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
		{
			name: "inject-route-config",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.RouteType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.RouteType,
			resourceIn: testRouteConfig,
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
					resourcev3.RouteType: []types.Resource{testRouteConfig},
				},
			},
		},
		{
			name: "new-secret",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.SecretType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.SecretType,
			resourceIn: testSecret,
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
					resourcev3.SecretType: []types.Resource{testSecret},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tableIn.AddOrReplaceXdsResource(tc.typeIn, tc.resourceIn, tc.funcIn)
			require.NoError(t, err)
			diff := cmp.Diff(tc.tableOut, tc.tableIn.DeepCopy(), protocmp.Transform())
			require.Empty(t, diff)
		})
	}
}

func TestInvalidAddXdsResource(t *testing.T) {
	invalidListener := &listenerv3.Listener{
		Name: "invalid-listener",
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Address: "",
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: 5000,
					},
					Protocol: corev3.SocketAddress_TCP,
				},
			},
		},
	}
	invalidRouteConfig := &routev3.RouteConfiguration{
		Name: "test-route-config",
		VirtualHosts: []*routev3.VirtualHost{
			{
				Name:    "", // missing name
				Domains: []string{"test.example.com"},
				Routes: []*routev3.Route{
					{
						Match: &routev3.RouteMatch{
							PathSpecifier: &routev3.RouteMatch_Prefix{
								Prefix: "/",
							},
						},
						Action: &routev3.Route_Route{
							Route: &routev3.RouteAction{
								ClusterSpecifier: &routev3.RouteAction_Cluster{
									Cluster: "test-cluster",
								},
							},
						},
					},
				},
			},
		},
	}

	invalidCluster := &clusterv3.Cluster{
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
												Address: "",
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
	invalidEndpoint := &endpointv3.ClusterLoadAssignment{
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
											Address: "",
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
	}

	invalidSecret := &tlsv3.Secret{
		Name: "=*&",
		Type: &tlsv3.Secret_TlsCertificate{
			TlsCertificate: &tlsv3.TlsCertificate{
				CertificateChain: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{
						InlineBytes: []byte("-----BEGIN CERTIFICATE-----\n... Your certificate data ... \n-----END CERTIFICATE-----"),
					},
				},
				PrivateKey: &corev3.DataSource{},
			},
		},
		// Add other fields for the secret as needed.
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
			name: "inject-invalid-listener",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.ListenerType,
			resourceIn: invalidListener,
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
			name: "inject-invalid-route-config",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.RouteType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.RouteType,
			resourceIn: invalidRouteConfig,
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
			tableOut: nil,
		},
		{
			name: "inject-invalid-cluster",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ClusterType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.ClusterType,
			resourceIn: invalidCluster,
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
			tableOut: nil,
		},
		{
			name: "cast-cluster-type",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ClusterType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.ClusterType,
			resourceIn: invalidListener,
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
			tableOut: nil,
		},
		{
			name: "cast-listener-type",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.ListenerType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.ListenerType,
			resourceIn: invalidCluster,
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
			tableOut: nil,
		},
		{
			name: "cast-route-config-type",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.RouteType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.RouteType,
			resourceIn: invalidCluster,
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
			tableOut: nil,
		},
		{
			name: "cast-secret-type",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.SecretType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.SecretType,
			resourceIn: invalidRouteConfig,
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
			tableOut: nil,
		},
		{
			name: "invalid-secret",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.SecretType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.SecretType,
			resourceIn: invalidSecret,
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
			tableOut: nil,
		},
		{
			name: "inject-invalid-endpoint",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.EndpointType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.EndpointType,
			resourceIn: invalidEndpoint,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldEndpoint := existing.(*endpointv3.ClusterLoadAssignment)
				newEndpoint := new.(*endpointv3.ClusterLoadAssignment)
				if newEndpoint == nil || oldEndpoint == nil {
					return false
				}
				if oldEndpoint.ClusterName == newEndpoint.ClusterName {
					return true
				}
				return false
			},
			tableOut: nil,
		},
		{
			name: "cast-endpoint-type",
			tableIn: &ResourceVersionTable{
				XdsResources: XdsResources{
					resourcev3.EndpointType: []types.Resource{},
				},
			},
			typeIn:     resourcev3.EndpointType,
			resourceIn: invalidListener,
			funcIn: func(existing types.Resource, new types.Resource) bool {
				oldEndpoint := existing.(*endpointv3.ClusterLoadAssignment)
				newEndpoint := new.(*endpointv3.ClusterLoadAssignment)
				if newEndpoint == nil || oldEndpoint == nil {
					return false
				}
				if oldEndpoint.ClusterName == newEndpoint.ClusterName {
					return true
				}
				return false
			},
			tableOut: nil,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.tableIn.AddOrReplaceXdsResource(tc.typeIn, tc.resourceIn, tc.funcIn)
			require.Error(t, err)
		})
	}
}
