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

type xdsClusterArgs struct {
	name         string
	settings     []*ir.DestinationSetting
	tSocket      *corev3.TransportSocket
	endpointType EndpointType
	loadBalancer *ir.LoadBalancer
}

type EndpointType int

const (
	EndpointTypeDNS EndpointType = iota
	EndpointTypeStatic
)

func buildXdsCluster(args *xdsClusterArgs) *clusterv3.Cluster {
	cluster := &clusterv3.Cluster{
		Name:            args.name,
		ConnectTimeout:  durationpb.New(10 * time.Second),
		DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
		CommonLbConfig: &clusterv3.Cluster_CommonLbConfig{
			LocalityConfigSpecifier: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{
				LocalityWeightedLbConfig: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig{}}},
		OutlierDetection:              &clusterv3.OutlierDetection{},
		PerConnectionBufferLimitBytes: wrapperspb.UInt32(tcpClusterPerConnectionBufferLimitBytes),
	}

	if args.tSocket != nil {
		cluster.TransportSocket = args.tSocket
	}

	if args.endpointType == EndpointTypeStatic {
		cluster.ClusterDiscoveryType = &clusterv3.Cluster_Type{Type: clusterv3.Cluster_EDS}
		cluster.EdsClusterConfig = &clusterv3.Cluster_EdsClusterConfig{
			ServiceName: args.name,
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

	isHTTP2 := false
	for _, ds := range args.settings {
		if ds.Protocol == ir.GRPC ||
			ds.Protocol == ir.HTTP2 {
			isHTTP2 = true
			break
		}
	}
	if isHTTP2 {
		cluster.TypedExtensionProtocolOptions = buildTypedExtensionProtocolOptions()
	}

	// Set Load Balancer policy
	//nolint:gocritic
	if args.loadBalancer == nil {
		cluster.LbPolicy = clusterv3.Cluster_LEAST_REQUEST
	} else if args.loadBalancer.LeastRequest != nil {
		cluster.LbPolicy = clusterv3.Cluster_LEAST_REQUEST
		if args.loadBalancer.LeastRequest.SlowStart != nil {
			if args.loadBalancer.LeastRequest.SlowStart.Window != nil {
				cluster.LbConfig = &clusterv3.Cluster_LeastRequestLbConfig_{
					LeastRequestLbConfig: &clusterv3.Cluster_LeastRequestLbConfig{
						SlowStartConfig: &clusterv3.Cluster_SlowStartConfig{
							SlowStartWindow: durationpb.New(args.loadBalancer.LeastRequest.SlowStart.Window.Duration),
						},
					},
				}
			}
		}
	} else if args.loadBalancer.RoundRobin != nil {
		cluster.LbPolicy = clusterv3.Cluster_ROUND_ROBIN
		if args.loadBalancer.RoundRobin.SlowStart != nil {
			if args.loadBalancer.RoundRobin.SlowStart.Window != nil {
				cluster.LbConfig = &clusterv3.Cluster_RoundRobinLbConfig_{
					RoundRobinLbConfig: &clusterv3.Cluster_RoundRobinLbConfig{
						SlowStartConfig: &clusterv3.Cluster_SlowStartConfig{
							SlowStartWindow: durationpb.New(args.loadBalancer.RoundRobin.SlowStart.Window.Duration),
						},
					},
				}
			}
		}
	} else if args.loadBalancer.Random != nil {
		cluster.LbPolicy = clusterv3.Cluster_RANDOM
	} else if args.loadBalancer.ConsistentHash != nil {
		cluster.LbPolicy = clusterv3.Cluster_MAGLEV
	}

	return cluster
}

func buildXdsClusterLoadAssignment(clusterName string, destSettings []*ir.DestinationSetting) *endpointv3.ClusterLoadAssignment {
	localities := make([]*endpointv3.LocalityLbEndpoints, 0, len(destSettings))
	for i, ds := range destSettings {

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

		// Envoy requires a distinct region to be set for each LocalityLbEndpoints.
		// If we don't do this, Envoy will merge all LocalityLbEndpoints into one.
		// We use the name of the backendRef as a pseudo region name.
		locality := &endpointv3.LocalityLbEndpoints{
			Locality: &corev3.Locality{
				Region: fmt.Sprintf("%s/backend/%d", clusterName, i),
			},
			LbEndpoints: endpoints,
			Priority:    0,
		}

		// Set locality weight
		var weight uint32
		if ds.Weight != nil {
			weight = *ds.Weight
		} else {
			weight = 1
		}
		locality.LoadBalancingWeight = &wrapperspb.UInt32Value{Value: weight}

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
