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
	preservecasev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/header_formatters/preserve_case/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/proxy_protocol/v3"
	rawbufferv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/raw_buffer/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	extensionOptionsKey = "envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#envoy-v3-api-field-config-cluster-v3-cluster-per-connection-buffer-limit-bytes
	tcpClusterPerConnectionBufferLimitBytes = 32768
	tcpClusterPerConnectTimeout             = 10 * time.Second
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
	http1Settings  *ir.HTTP1Settings
	timeout        *ir.Timeout
	tcpkeepalive   *ir.TCPKeepalive
}

type EndpointType int

const (
	EndpointTypeDNS EndpointType = iota
	EndpointTypeStatic
)

func buildEndpointType(settings []*ir.DestinationSetting) EndpointType {
	// Get endpoint address type for xds cluster by returning the first DestinationSetting's AddressType,
	// since there's no Mixed AddressType among all the DestinationSettings.
	if settings == nil {
		return EndpointTypeStatic
	}

	addrType := settings[0].AddressType

	if addrType != nil && *addrType == ir.FQDN {
		return EndpointTypeDNS
	}

	return EndpointTypeStatic
}

func buildXdsCluster(args *xdsClusterArgs) *clusterv3.Cluster {
	cluster := &clusterv3.Cluster{
		Name:            args.name,
		DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
		CommonLbConfig: &clusterv3.Cluster_CommonLbConfig{
			LocalityConfigSpecifier: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{
				LocalityWeightedLbConfig: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig{}}},
		OutlierDetection:              &clusterv3.OutlierDetection{},
		PerConnectionBufferLimitBytes: wrapperspb.UInt32(tcpClusterPerConnectionBufferLimitBytes),
	}

	cluster.ConnectTimeout = buildConnectTimeout(args.timeout)

	// Set Proxy Protocol
	if args.proxyProtocol != nil {
		cluster.TransportSocket = buildProxyProtocolSocket(args.proxyProtocol, args.tSocket)
	} else if args.tSocket != nil {
		cluster.TransportSocket = args.tSocket
	}

	for i, ds := range args.settings {
		if ds.TLS != nil {
			socket, err := buildXdsUpstreamTLSSocketWthCert(ds.TLS)
			if err != nil {
				// TODO: Log something here
				return nil
			}
			if args.proxyProtocol != nil {
				socket = buildProxyProtocolSocket(args.proxyProtocol, socket)
			}
			matchName := fmt.Sprintf("%s/tls/%d", args.name, i)
			cluster.TransportSocketMatches = append(cluster.TransportSocketMatches, &clusterv3.Cluster_TransportSocketMatch{
				Name: matchName,
				Match: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"name": structpb.NewStringValue(matchName),
					},
				},
				TransportSocket: socket,
			})
		}
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

	// build common, HTTP/1 and HTTP/2  protocol options for cluster
	epo := buildTypedExtensionProtocolOptions(args)
	if epo != nil {
		cluster.TypedExtensionProtocolOptions = epo
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

	if args.healthCheck != nil && args.healthCheck.Active != nil {
		cluster.HealthChecks = buildXdsHealthCheck(args.healthCheck.Active)
	}

	if args.healthCheck != nil && args.healthCheck.Passive != nil {
		cluster.OutlierDetection = buildXdsOutlierDetection(args.healthCheck.Passive)

	}

	cluster.CircuitBreakers = buildXdsClusterCircuitBreaker(args.circuitBreaker)

	if args.tcpkeepalive != nil {
		cluster.UpstreamConnectionOptions = buildXdsClusterUpstreamOptions(args.tcpkeepalive)
	}
	return cluster
}

func buildXdsHealthCheck(healthcheck *ir.ActiveHealthCheck) []*corev3.HealthCheck {
	hc := &corev3.HealthCheck{
		Timeout:  durationpb.New(healthcheck.Timeout.Duration),
		Interval: durationpb.New(healthcheck.Interval.Duration),
	}
	if healthcheck.UnhealthyThreshold != nil {
		hc.UnhealthyThreshold = wrapperspb.UInt32(*healthcheck.UnhealthyThreshold)
	}
	if healthcheck.HealthyThreshold != nil {
		hc.HealthyThreshold = wrapperspb.UInt32(*healthcheck.HealthyThreshold)
	}
	if healthcheck.HTTP != nil {
		httpChecker := &corev3.HealthCheck_HttpHealthCheck{
			Host: healthcheck.HTTP.Host,
			Path: healthcheck.HTTP.Path,
		}
		if healthcheck.HTTP.Method != nil {
			httpChecker.Method = corev3.RequestMethod(corev3.RequestMethod_value[*healthcheck.HTTP.Method])
		}
		httpChecker.ExpectedStatuses = buildHTTPStatusRange(healthcheck.HTTP.ExpectedStatuses)
		if receive := buildHealthCheckPayload(healthcheck.HTTP.ExpectedResponse); receive != nil {
			httpChecker.Receive = append(httpChecker.Receive, receive)
		}
		hc.HealthChecker = &corev3.HealthCheck_HttpHealthCheck_{
			HttpHealthCheck: httpChecker,
		}
	}
	if healthcheck.TCP != nil {
		tcpChecker := &corev3.HealthCheck_TcpHealthCheck{
			Send: buildHealthCheckPayload(healthcheck.TCP.Send),
		}
		if receive := buildHealthCheckPayload(healthcheck.TCP.Receive); receive != nil {
			tcpChecker.Receive = append(tcpChecker.Receive, receive)
		}
		hc.HealthChecker = &corev3.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: tcpChecker,
		}
	}
	return []*corev3.HealthCheck{hc}
}

func buildXdsOutlierDetection(outlierDetection *ir.OutlierDetection) *clusterv3.OutlierDetection {
	od := &clusterv3.OutlierDetection{
		BaseEjectionTime: durationpb.New(outlierDetection.BaseEjectionTime.Duration),
		Interval:         durationpb.New(outlierDetection.Interval.Duration),
	}
	if outlierDetection.SplitExternalLocalOriginErrors != nil {
		od.SplitExternalLocalOriginErrors = *outlierDetection.SplitExternalLocalOriginErrors
	}

	if outlierDetection.MaxEjectionPercent != nil && *outlierDetection.MaxEjectionPercent > 0 {
		od.MaxEjectionPercent = wrapperspb.UInt32(uint32(*outlierDetection.MaxEjectionPercent))
	}

	if outlierDetection.ConsecutiveLocalOriginFailures != nil {
		od.ConsecutiveLocalOriginFailure = wrapperspb.UInt32(*outlierDetection.ConsecutiveLocalOriginFailures)
	}

	if outlierDetection.Consecutive5xxErrors != nil {
		od.Consecutive_5Xx = wrapperspb.UInt32(*outlierDetection.Consecutive5xxErrors)
	}

	if outlierDetection.ConsecutiveGatewayErrors != nil {
		od.ConsecutiveGatewayFailure = wrapperspb.UInt32(*outlierDetection.ConsecutiveGatewayErrors)
	}

	return od
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
	// Always allow the same amount of retries as regular requests to handle surges in retries
	// related to pod restarts
	cbt := &clusterv3.CircuitBreakers_Thresholds{
		Priority: corev3.RoutingPriority_DEFAULT,
		MaxRetries: &wrapperspb.UInt32Value{
			Value: uint32(1024),
		},
	}

	if circuitBreaker != nil {
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

		if circuitBreaker.MaxParallelRetries != nil {
			cbt.MaxRetries = &wrapperspb.UInt32Value{
				Value: *circuitBreaker.MaxParallelRetries,
			}
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

		var metadata *corev3.Metadata
		if ds.TLS != nil {
			metadata = &corev3.Metadata{
				FilterMetadata: map[string]*structpb.Struct{
					"envoy.transport_socket_match": {
						Fields: map[string]*structpb.Value{
							"name": structpb.NewStringValue(fmt.Sprintf("%s/tls/%d", clusterName, i)),
						},
					},
				},
			}
		}

		for _, irEp := range ds.Endpoints {
			lbEndpoint := &endpointv3.LbEndpoint{
				Metadata: metadata,
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

func buildTypedExtensionProtocolOptions(args *xdsClusterArgs) map[string]*anypb.Any {
	requiresHTTP2Options := false
	for _, ds := range args.settings {
		if ds.Protocol == ir.GRPC ||
			ds.Protocol == ir.HTTP2 {
			requiresHTTP2Options = true
			break
		}
	}

	requiresCommonHTTPOptions := (args.timeout != nil && args.timeout.HTTP != nil &&
		(args.timeout.HTTP.MaxConnectionDuration != nil || args.timeout.HTTP.ConnectionIdleTimeout != nil)) ||
		(args.circuitBreaker != nil && args.circuitBreaker.MaxRequestsPerConnection != nil)

	requiresHTTP1Options := args.http1Settings != nil && (args.http1Settings.EnableTrailers || args.http1Settings.PreserveHeaderCase || args.http1Settings.HTTP10 != nil)

	if !(requiresCommonHTTPOptions || requiresHTTP1Options || requiresHTTP2Options) {
		return nil
	}

	protocolOptions := httpv3.HttpProtocolOptions{}

	if requiresCommonHTTPOptions {
		protocolOptions.CommonHttpProtocolOptions = &corev3.HttpProtocolOptions{}

		if args.timeout != nil && args.timeout.HTTP != nil {
			if args.timeout.HTTP.ConnectionIdleTimeout != nil {
				protocolOptions.CommonHttpProtocolOptions.IdleTimeout =
					durationpb.New(args.timeout.HTTP.ConnectionIdleTimeout.Duration)
			}

			if args.timeout.HTTP.MaxConnectionDuration != nil {
				protocolOptions.CommonHttpProtocolOptions.MaxConnectionDuration =
					durationpb.New(args.timeout.HTTP.MaxConnectionDuration.Duration)
			}
		}

		if args.circuitBreaker != nil && args.circuitBreaker.MaxRequestsPerConnection != nil {
			protocolOptions.CommonHttpProtocolOptions.MaxRequestsPerConnection = &wrapperspb.UInt32Value{
				Value: *args.circuitBreaker.MaxRequestsPerConnection,
			}
		}

	}

	// When setting any Typed Extension Protocol Options, UpstreamProtocolOptions are mandatory
	// If translation requires HTTP2 enablement or HTTP1 trailers, set appropriate setting
	// Default to http1 otherwise
	// TODO: If the cluster is TLS enabled, use AutoHTTPConfig instead of ExplicitHttpConfig
	// so that when ALPN is supported then enabling http1 options doesn't force HTTP/1.1
	switch {
	case requiresHTTP2Options:
		protocolOptions.UpstreamProtocolOptions = &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{},
			},
		}
	case requiresHTTP1Options:
		http1opts := &corev3.Http1ProtocolOptions{
			EnableTrailers: args.http1Settings.EnableTrailers,
		}
		if args.http1Settings.PreserveHeaderCase {
			preservecaseAny, _ := anypb.New(&preservecasev3.PreserveCaseFormatterConfig{})
			http1opts.HeaderKeyFormat = &corev3.Http1ProtocolOptions_HeaderKeyFormat{
				HeaderFormat: &corev3.Http1ProtocolOptions_HeaderKeyFormat_StatefulFormatter{
					StatefulFormatter: &corev3.TypedExtensionConfig{
						Name:        "preserve_case",
						TypedConfig: preservecaseAny,
					},
				},
			}
		}
		if args.http1Settings.HTTP10 != nil {
			http1opts.AcceptHttp_10 = true
			http1opts.DefaultHostForHttp_10 = ptr.Deref(args.http1Settings.HTTP10.DefaultHost, "")
		}
		protocolOptions.UpstreamProtocolOptions = &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{
					HttpProtocolOptions: http1opts,
				},
			},
		}
	default:
		protocolOptions.UpstreamProtocolOptions = &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{},
			},
		}
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

func buildConnectTimeout(to *ir.Timeout) *durationpb.Duration {
	if to != nil && to.TCP != nil && to.TCP.ConnectTimeout != nil {
		return durationpb.New(to.TCP.ConnectTimeout.Duration)
	}
	return durationpb.New(tcpClusterPerConnectTimeout)
}

func buildXdsClusterUpstreamOptions(tcpkeepalive *ir.TCPKeepalive) *clusterv3.UpstreamConnectionOptions {
	if tcpkeepalive == nil {
		return nil
	}

	ka := &clusterv3.UpstreamConnectionOptions{
		TcpKeepalive: &corev3.TcpKeepalive{},
	}

	if tcpkeepalive.Probes != nil {
		ka.TcpKeepalive.KeepaliveProbes = wrapperspb.UInt32(*tcpkeepalive.Probes)
	}

	if tcpkeepalive.IdleTime != nil {
		ka.TcpKeepalive.KeepaliveTime = wrapperspb.UInt32(*tcpkeepalive.IdleTime)
	}

	if tcpkeepalive.Interval != nil {
		ka.TcpKeepalive.KeepaliveInterval = wrapperspb.UInt32(*tcpkeepalive.Interval)
	}

	return ka

}
