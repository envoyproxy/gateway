// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/proxy_protocol/v3"
	rawbufferv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/raw_buffer/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
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
	name           string
	settings       []*ir.DestinationSetting
	tSocket        *corev3.TransportSocket
	endpointType   EndpointType
	loadBalancer   *ir.LoadBalancer
	proxyProtocol  *ir.ProxyProtocol
	circuitBreaker *ir.CircuitBreaker
	healthCheck    *ir.HealthCheck
	enableTrailers bool
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

	// Set Proxy Protocol
	if args.proxyProtocol != nil {
		cluster.TransportSocket = buildProxyProtocolSocket(args.proxyProtocol, args.tSocket)
	} else if args.tSocket != nil {
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
	cluster.TypedExtensionProtocolOptions = buildTypedExtensionProtocolOptions(isHTTP2, args.enableTrailers)

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

	if args.healthCheck != nil {
		hc := &corev3.HealthCheck{
			Timeout:  durationpb.New(args.healthCheck.Timeout.Duration),
			Interval: durationpb.New(args.healthCheck.Interval.Duration),
		}
		if args.healthCheck.UnhealthyThreshold != nil {
			hc.UnhealthyThreshold = wrapperspb.UInt32(*args.healthCheck.UnhealthyThreshold)
		}
		if args.healthCheck.HealthyThreshold != nil {
			hc.HealthyThreshold = wrapperspb.UInt32(*args.healthCheck.HealthyThreshold)
		}
		if args.healthCheck.HTTP != nil {
			httpChecker := &corev3.HealthCheck_HttpHealthCheck{
				Path: args.healthCheck.HTTP.Path,
			}
			if args.healthCheck.HTTP.Method != nil {
				httpChecker.Method = corev3.RequestMethod(corev3.RequestMethod_value[*args.healthCheck.HTTP.Method])
			}
			httpChecker.ExpectedStatuses = buildHTTPStatusRange(args.healthCheck.HTTP.ExpectedStatuses)
			if receive := buildHealthCheckPayload(args.healthCheck.HTTP.ExpectedResponse); receive != nil {
				httpChecker.Receive = append(httpChecker.Receive, receive)
			}
			hc.HealthChecker = &corev3.HealthCheck_HttpHealthCheck_{
				HttpHealthCheck: httpChecker,
			}
		}
		if args.healthCheck.TCP != nil {
			tcpChecker := &corev3.HealthCheck_TcpHealthCheck{
				Send: buildHealthCheckPayload(args.healthCheck.TCP.Send),
			}
			if receive := buildHealthCheckPayload(args.healthCheck.TCP.Receive); receive != nil {
				tcpChecker.Receive = append(tcpChecker.Receive, receive)
			}
			hc.HealthChecker = &corev3.HealthCheck_TcpHealthCheck_{
				TcpHealthCheck: tcpChecker,
			}
		}
		cluster.HealthChecks = []*corev3.HealthCheck{hc}
	}
	if args.circuitBreaker != nil {
		cluster.CircuitBreakers = buildXdsClusterCircuitBreaker(args.circuitBreaker)
	}

	return cluster
}

// buildHTTPStatusRange converts an array of http status to an array of the range of http status.
func buildHTTPStatusRange(irStatuses []ir.HTTPStatus) []*xdstype.Int64Range {
	if len(irStatuses) == 0 {
		return nil
	}
	ranges := []*xdstype.Int64Range{}
	sort.Slice(irStatuses, func(i int, j int) bool {
		return irStatuses[i] < irStatuses[j]
	})
	var start, end int64
	for i := 0; i < len(irStatuses); i++ {
		switch {
		case start == 0:
			start = int64(irStatuses[i])
			end = int64(irStatuses[i] + 1)
		case int64(irStatuses[i]) == end:
			end++
		default:
			ranges = append(ranges, &xdstype.Int64Range{Start: start, End: end})
			start = int64(irStatuses[i])
			end = int64(irStatuses[i] + 1)
		}
	}
	if start != 0 {
		ranges = append(ranges, &xdstype.Int64Range{Start: start, End: end})
	}
	return ranges
}

func buildHealthCheckPayload(irLoad *ir.HealthCheckPayload) *corev3.HealthCheck_Payload {
	if irLoad == nil {
		return nil
	}

	var hcp corev3.HealthCheck_Payload
	if irLoad.Text != nil && *irLoad.Text != "" {
		hcp.Payload = &corev3.HealthCheck_Payload_Text{
			Text: hex.EncodeToString([]byte(*irLoad.Text)),
		}
	}
	if len(irLoad.Binary) > 0 {
		binPayload := &corev3.HealthCheck_Payload_Binary{
			Binary: make([]byte, len(irLoad.Binary)),
		}
		copy(binPayload.Binary, irLoad.Binary)
		hcp.Payload = binPayload
	}
	return &hcp
}

func buildXdsClusterCircuitBreaker(circuitBreaker *ir.CircuitBreaker) *clusterv3.CircuitBreakers {
	cbt := &clusterv3.CircuitBreakers_Thresholds{
		Priority: corev3.RoutingPriority_DEFAULT,
	}

	if circuitBreaker.MaxConnections != nil {
		cbt.MaxConnections = &wrapperspb.UInt32Value{
			Value: *circuitBreaker.MaxConnections,
		}
	}

	if circuitBreaker.MaxPendingRequests != nil {
		cbt.MaxPendingRequests = &wrapperspb.UInt32Value{
			Value: *circuitBreaker.MaxPendingRequests,
		}
	}

	if circuitBreaker.MaxParallelRequests != nil {
		cbt.MaxRequests = &wrapperspb.UInt32Value{
			Value: *circuitBreaker.MaxParallelRequests,
		}
	}

	ecb := &clusterv3.CircuitBreakers{
		Thresholds: []*clusterv3.CircuitBreakers_Thresholds{cbt},
	}

	return ecb
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

func buildTypedExtensionProtocolOptions(http2, http1Trailers bool) map[string]*anypb.Any {
	var anyProtocolOptions *anypb.Any

	if http2 {
		protocolOptions := httpv3.HttpProtocolOptions{
			UpstreamProtocolOptions: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{},
				},
			},
		}

		anyProtocolOptions, _ = anypb.New(&protocolOptions)
	} else if http1Trailers {
		// TODO: If the cluster is TLS enabled, use AutoHTTPConfig instead of ExplicitHttpConfig
		// so that when ALPN is supported enabling trailers doesn't force HTTP/1.1
		protocolOptions := httpv3.HttpProtocolOptions{
			UpstreamProtocolOptions: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
				ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
					ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{
						HttpProtocolOptions: &corev3.Http1ProtocolOptions{
							EnableTrailers: http1Trailers,
						},
					},
				},
			},
		}
		anyProtocolOptions, _ = anypb.New(&protocolOptions)
	}

	if anyProtocolOptions != nil {
		extensionOptions := map[string]*anypb.Any{
			extensionOptionsKey: anyProtocolOptions,
		}

		return extensionOptions
	}
	return nil
}

// buildClusterName returns a cluster name for the given `host` and `port`.
// The format is: <type>|<host>|<port>, where type is "accesslog" for access logs.
// It's easy to distinguish when debugging.
func buildClusterName(prefix string, host string, port uint32) string {
	return fmt.Sprintf("%s|%s|%d", prefix, host, port)
}

// buildProxyProtocolSocket builds the ProxyProtocol transport socket.
func buildProxyProtocolSocket(proxyProtocol *ir.ProxyProtocol, tSocket *corev3.TransportSocket) *corev3.TransportSocket {
	if proxyProtocol == nil {
		return nil
	}

	ppCtx := &proxyprotocolv3.ProxyProtocolUpstreamTransport{}

	switch proxyProtocol.Version {
	case ir.ProxyProtocolVersionV1:
		ppCtx.Config = &corev3.ProxyProtocolConfig{
			Version: corev3.ProxyProtocolConfig_V1,
		}
	case ir.ProxyProtocolVersionV2:
		ppCtx.Config = &corev3.ProxyProtocolConfig{
			Version: corev3.ProxyProtocolConfig_V2,
		}
	}

	// If existing transport socket does not exist wrap around raw buffer
	if tSocket == nil {
		rawCtx := &rawbufferv3.RawBuffer{}
		rawCtxAny, err := anypb.New(rawCtx)
		if err != nil {
			return nil
		}
		rawSocket := &corev3.TransportSocket{
			Name: wellknown.TransportSocketRawBuffer,
			ConfigType: &corev3.TransportSocket_TypedConfig{
				TypedConfig: rawCtxAny,
			},
		}
		ppCtx.TransportSocket = rawSocket
	} else {
		ppCtx.TransportSocket = tSocket
	}

	ppCtxAny, err := anypb.New(ppCtx)
	if err != nil {
		return nil
	}

	return &corev3.TransportSocket{
		Name: "envoy.transport_sockets.upstream_proxy_protocol",
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: ppCtxAny,
		},
	}
}
