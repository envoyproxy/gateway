// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
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

const (
	extensionOptionsKey = "envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#envoy-v3-api-field-config-cluster-v3-cluster-per-connection-buffer-limit-bytes
	tcpClusterPerConnectionBufferLimitBytes = 32768
)

func buildXdsCluster(clusterName string, tSocket *corev3.TransportSocket, protocol ProtocolType, endpointType EndpointType) *clusterv3.Cluster {
	cluster := &clusterv3.Cluster{
		Name:            clusterName,
		ConnectTimeout:  durationpb.New(10 * time.Second),
		LbPolicy:        clusterv3.Cluster_LEAST_REQUEST,
		DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
		CommonLbConfig: &clusterv3.Cluster_CommonLbConfig{
			LocalityConfigSpecifier: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{
				LocalityWeightedLbConfig: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig{}}},
		OutlierDetection:              &clusterv3.OutlierDetection{},
		PerConnectionBufferLimitBytes: wrapperspb.UInt32(tcpClusterPerConnectionBufferLimitBytes),
	}

	if tSocket != nil {
		cluster.TransportSocket = tSocket
	}

	if endpointType == Static {
		cluster.ClusterDiscoveryType = &clusterv3.Cluster_Type{Type: clusterv3.Cluster_EDS}
		cluster.EdsClusterConfig = &clusterv3.Cluster_EdsClusterConfig{
			ServiceName: clusterName,
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

func buildXdsClusterLoadAssignment(clusterName string, destSettings []*ir.DestinationSetting) *endpointv3.ClusterLoadAssignment {
	localities := make([]*endpointv3.LocalityLbEndpoints, 0, len(destSettings))
	for _, ds := range destSettings {

		endpoints := make([]*endpointv3.LbEndpoint, 0, len(ds.Endpoints))

		for _, irEp := range ds.Endpoints {
			lbEndpoint := &endpointv3.LbEndpoint{
				HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
					Endpoint: &endpointv3.Endpoint{
						Address: &corev3.Address{
							Address: &corev3.Address_SocketAddress{
								SocketAddress: &corev3.SocketAddress{
									Protocol: corev3.SocketAddress_TCP,
									Address:  irEp.Host,
									PortSpecifier: &corev3.SocketAddress_PortValue{
										PortValue: irEp.Port,
									},
								},
							},
						},
					},
				},
			}
			// Set default weight of 1 for all endpoints.
			lbEndpoint.LoadBalancingWeight = &wrapperspb.UInt32Value{Value: 1}
			endpoints = append(endpoints, lbEndpoint)
		}

		locality := &endpointv3.LocalityLbEndpoints{
			Locality:            &corev3.Locality{},
			LbEndpoints:         endpoints,
			Priority:            0,
			LoadBalancingWeight: &wrapperspb.UInt32Value{Value: *ds.Weight},
		}
		localities = append(localities, locality)
	}
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
		extensionOptionsKey: anyProtocolOptions,
	}

	return extensionOptions
}

// buildClusterName returns a cluster name for the given `host` and `port`.
// The format is: <type>|<host>|<port>, where type is "accesslog" for access logs.
// It's easy to distinguish when debugging.
func buildClusterName(prefix string, host string, port uint32) string {
	return fmt.Sprintf("%s|%s|%d", prefix, host, port)
}
