// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildXdsCluster(routeName string, tSocket *corev3.TransportSocket, protocol ProtocolType, c EndpointType) *clusterv3.Cluster {
	clusterName := routeName
	cluster := &clusterv3.Cluster{
		Name:            clusterName,
		ConnectTimeout:  durationpb.New(10 * time.Second),
		LbPolicy:        clusterv3.Cluster_ROUND_ROBIN,
		DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
		CommonLbConfig: &clusterv3.Cluster_CommonLbConfig{
			LocalityConfigSpecifier: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{
				LocalityWeightedLbConfig: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig{}}},
		OutlierDetection: &clusterv3.OutlierDetection{},
	}

	if tSocket != nil {
		cluster.TransportSocket = tSocket
	}

	if c == Static {
		cluster.ClusterDiscoveryType = &clusterv3.Cluster_Type{Type: clusterv3.Cluster_EDS}
		cluster.EdsClusterConfig = &clusterv3.Cluster_EdsClusterConfig{
			EdsConfig: &corev3.ConfigSource{
				ResourceApiVersion: resource.DefaultAPIVersion,
				ConfigSourceSpecifier: &corev3.ConfigSource_Ads{
					Ads: &corev3.AggregatedConfigSource{},
				},
			},
		}
	} else {
		cluster.ClusterDiscoveryType = &clusterv3.Cluster_Type{Type: clusterv3.Cluster_STRICT_DNS}
		cluster.DnsRefreshRate = durationpb.New(30 * time.Second)
		cluster.RespectDnsTtl = true
	}

	if protocol == HTTP2 {
		cluster.TypedExtensionProtocolOptions = buildTypedExtensionProtocolOptions()
	}

	return cluster
}

func buildXdsClusterLoadAssignment(clusterName string, destinations []*ir.RouteDestination) *endpointv3.ClusterLoadAssignment {
	endpoints := make([]*endpointv3.LbEndpoint, 0, len(destinations))
	for _, destination := range destinations {
		lbEndpoint := &endpointv3.LbEndpoint{
			HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
				Endpoint: &endpointv3.Endpoint{
					Address: &corev3.Address{
						Address: &corev3.Address_SocketAddress{
							SocketAddress: &corev3.SocketAddress{
								Protocol: corev3.SocketAddress_TCP,
								Address:  destination.Host,
								PortSpecifier: &corev3.SocketAddress_PortValue{
									PortValue: destination.Port,
								},
							},
						},
					},
				},
			},
		}
		if destination.Weight != nil {
			lbEndpoint.LoadBalancingWeight = &wrapperspb.UInt32Value{Value: *destination.Weight}
		}
		endpoints = append(endpoints, lbEndpoint)
	}

	localities := make([]*endpointv3.LocalityLbEndpoints, 0, 1)
	locality := &endpointv3.LocalityLbEndpoints{
		Locality:    &corev3.Locality{},
		LbEndpoints: endpoints,
		Priority:    0,
		// Each locality gets the same weight 1. There is a single locality
		// per priority, so the weight value does not really matter, but some
		// load balancers need the value to be set.
		LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 1}}
	localities = append(localities, locality)

	return &endpointv3.ClusterLoadAssignment{ClusterName: clusterName, Endpoints: localities}
}

func buildTypedExtensionProtocolOptions() map[string]*anypb.Any {
	protocolOptions := httpv3.HttpProtocolOptions{
		UpstreamProtocolOptions: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{},
			},
		},
	}

	anyProtocolOptions, _ := anypb.New(&protocolOptions)

	extensionOptions := map[string]*anypb.Any{
		"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": anyProtocolOptions,
	}

	return extensionOptions
}
