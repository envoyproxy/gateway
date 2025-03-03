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
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

const (
	extensionOptionsKey = "envoy.extensions.upstreams.http.v3.HttpProtocolOptions"
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#envoy-v3-api-field-config-cluster-v3-cluster-per-connection-buffer-limit-bytes
	tcpClusterPerConnectionBufferLimitBytes = 32768
	tcpClusterPerConnectTimeout             = 10 * time.Second
)

type xdsClusterArgs struct {
	name              string
	settings          []*ir.DestinationSetting
	tSocket           *corev3.TransportSocket
	endpointType      EndpointType
	loadBalancer      *ir.LoadBalancer
	proxyProtocol     *ir.ProxyProtocol
	circuitBreaker    *ir.CircuitBreaker
	healthCheck       *ir.HealthCheck
	http1Settings     *ir.HTTP1Settings
	http2Settings     *ir.HTTP2Settings
	timeout           *ir.Timeout
	tcpkeepalive      *ir.TCPKeepalive
	metrics           *ir.Metrics
	backendConnection *ir.BackendConnection
	dns               *ir.DNS
	useClientProtocol bool
	ipFamily          *egv1a1.IPFamily
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
	dnsLookupFamily := clusterv3.Cluster_V4_PREFERRED
	customDNSPolicy := args.dns != nil && args.dns.LookupFamily != nil
	// apply DNS lookup family if custom DNS traffic policy is set
	if customDNSPolicy {
		switch *args.dns.LookupFamily {
		case egv1a1.IPv4DNSLookupFamily:
			dnsLookupFamily = clusterv3.Cluster_V4_ONLY
		case egv1a1.IPv6DNSLookupFamily:
			dnsLookupFamily = clusterv3.Cluster_V6_ONLY
		case egv1a1.IPv6PreferredDNSLookupFamily:
			dnsLookupFamily = clusterv3.Cluster_AUTO
		case egv1a1.IPv4AndIPv6DNSLookupFamily:
			dnsLookupFamily = clusterv3.Cluster_ALL
		}
	}

	// Ensure to override if a specific IP family is set.
	if args.ipFamily != nil {
		switch *args.ipFamily {
		case egv1a1.IPv4:
			dnsLookupFamily = clusterv3.Cluster_V4_ONLY
		case egv1a1.IPv6:
			dnsLookupFamily = clusterv3.Cluster_V6_ONLY
		case egv1a1.DualStack:
			// if a custom DNS policy is set, prefer the custom policy as its more specific.
			if !customDNSPolicy {
				dnsLookupFamily = clusterv3.Cluster_ALL
			}
		}
	}

	cluster := &clusterv3.Cluster{
		Name:            args.name,
		DnsLookupFamily: dnsLookupFamily,
		CommonLbConfig: &clusterv3.Cluster_CommonLbConfig{
			LocalityConfigSpecifier: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{
				LocalityWeightedLbConfig: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig{},
			},
		},
		PerConnectionBufferLimitBytes: buildBackandConnectionBufferLimitBytes(args.backendConnection),
	}

	// 50% is the Envoy default value for panic threshold. No need to explicitly set it in this case.
	if args.healthCheck != nil && args.healthCheck.PanicThreshold != nil && *args.healthCheck.PanicThreshold != 50 {
		cluster.CommonLbConfig.HealthyPanicThreshold = &xdstype.Percent{
			Value: float64(*args.healthCheck.PanicThreshold),
		}
	}

	cluster.ConnectTimeout = buildConnectTimeout(args.timeout)

	// Initialize TrackClusterStats if any metrics are enabled
	if args.metrics != nil && (args.metrics.EnablePerEndpointStats || args.metrics.EnableRequestResponseSizesStats) {
		cluster.TrackClusterStats = &clusterv3.TrackClusterStats{}

		// Set per endpoint stats if enabled
		if args.metrics.EnablePerEndpointStats {
			cluster.TrackClusterStats.PerEndpointStats = args.metrics.EnablePerEndpointStats
		}

		// Set request response sizes stats if enabled
		if args.metrics.EnableRequestResponseSizesStats {
			cluster.TrackClusterStats.RequestResponseSizes = args.metrics.EnableRequestResponseSizesStats
		}
	}

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
		// Dont wait for a health check to determine health and remove these endpoints
		// if the endpoint has been removed via EDS by the control plane
		cluster.IgnoreHealthOnHostRemoval = true
	} else {
		cluster.ClusterDiscoveryType = &clusterv3.Cluster_Type{Type: clusterv3.Cluster_STRICT_DNS}
		cluster.DnsRefreshRate = durationpb.New(30 * time.Second)
		cluster.RespectDnsTtl = true
		if args.dns != nil {
			if args.dns.DNSRefreshRate != nil {
				if args.dns.DNSRefreshRate.Duration > 0 {
					cluster.DnsRefreshRate = durationpb.New(args.dns.DNSRefreshRate.Duration)
				}
			}
			if args.dns.RespectDNSTTL != nil {
				cluster.RespectDnsTtl = ptr.Deref(args.dns.RespectDNSTTL, true)
			}
		}
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
		if args.loadBalancer.RoundRobin.SlowStart != nil && args.loadBalancer.RoundRobin.SlowStart.Window != nil {
			cluster.LbConfig = &clusterv3.Cluster_RoundRobinLbConfig_{
				RoundRobinLbConfig: &clusterv3.Cluster_RoundRobinLbConfig{
					SlowStartConfig: &clusterv3.Cluster_SlowStartConfig{
						SlowStartWindow: durationpb.New(args.loadBalancer.RoundRobin.SlowStart.Window.Duration),
					},
				},
			}
		}
	} else if args.loadBalancer.Random != nil {
		cluster.LbPolicy = clusterv3.Cluster_RANDOM
	} else if args.loadBalancer.ConsistentHash != nil {
		cluster.LbPolicy = clusterv3.Cluster_MAGLEV

		if args.loadBalancer.ConsistentHash.TableSize != nil {
			cluster.LbConfig = &clusterv3.Cluster_MaglevLbConfig_{
				MaglevLbConfig: &clusterv3.Cluster_MaglevLbConfig{
					TableSize: &wrapperspb.UInt64Value{Value: *args.loadBalancer.ConsistentHash.TableSize},
				},
			}
		}
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
	switch {
	case healthcheck.HTTP != nil:
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
	case healthcheck.TCP != nil:
		tcpChecker := &corev3.HealthCheck_TcpHealthCheck{
			Send: buildHealthCheckPayload(healthcheck.TCP.Send),
		}
		if receive := buildHealthCheckPayload(healthcheck.TCP.Receive); receive != nil {
			tcpChecker.Receive = append(tcpChecker.Receive, receive)
		}
		hc.HealthChecker = &corev3.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: tcpChecker,
		}
	case healthcheck.GRPC != nil:
		hc.HealthChecker = &corev3.HealthCheck_GrpcHealthCheck_{
			GrpcHealthCheck: &corev3.HealthCheck_GrpcHealthCheck{
				ServiceName: ptr.Deref(healthcheck.GRPC.Service, ""),
			},
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
			healthStatus := corev3.HealthStatus_UNKNOWN
			if irEp.Draining {
				healthStatus = corev3.HealthStatus_DRAINING
			}
			lbEndpoint := &endpointv3.LbEndpoint{
				Metadata: metadata,
				HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
					Endpoint: &endpointv3.Endpoint{
						Address: buildAddress(irEp),
					},
				},
				HealthStatus: healthStatus,
			}
			// Set default weight of 1 for all endpoints.
			lbEndpoint.LoadBalancingWeight = &wrapperspb.UInt32Value{Value: 1}
			endpoints = append(endpoints, lbEndpoint)
		}

		locality := &endpointv3.LocalityLbEndpoints{
			/*Locality: &corev3.Locality{
				Region: ds.Name,
			},*/
			LbEndpoints: endpoints,
			// Priority:    0,
		}

		// Set locality weight
		var weight uint32
		if ds.Weight != nil {
			weight = *ds.Weight
		} else {
			weight = 1
		}
		locality.LoadBalancingWeight = &wrapperspb.UInt32Value{Value: weight}
		locality.Priority = ptr.Deref(ds.Priority, 0)
		localities = append(localities, locality)
	}
	if len(clusterName) == 0 {
		fmt.Println("xxxxxxxx Cluster name is empty")
		clusterName = "default"
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

	if !(requiresCommonHTTPOptions || requiresHTTP1Options || requiresHTTP2Options || args.useClientProtocol) {
		return nil
	}

	protocolOptions := httpv3.HttpProtocolOptions{}

	if requiresCommonHTTPOptions {
		protocolOptions.CommonHttpProtocolOptions = &corev3.HttpProtocolOptions{}

		if args.timeout != nil && args.timeout.HTTP != nil {
			if args.timeout.HTTP.ConnectionIdleTimeout != nil {
				protocolOptions.CommonHttpProtocolOptions.IdleTimeout = durationpb.New(args.timeout.HTTP.ConnectionIdleTimeout.Duration)
			}

			if args.timeout.HTTP.MaxConnectionDuration != nil {
				protocolOptions.CommonHttpProtocolOptions.MaxConnectionDuration = durationpb.New(args.timeout.HTTP.MaxConnectionDuration.Duration)
			}
		}

		if args.circuitBreaker != nil && args.circuitBreaker.MaxRequestsPerConnection != nil {
			protocolOptions.CommonHttpProtocolOptions.MaxRequestsPerConnection = &wrapperspb.UInt32Value{
				Value: *args.circuitBreaker.MaxRequestsPerConnection,
			}
		}
	}

	http1opts := &corev3.Http1ProtocolOptions{}
	if args.http1Settings != nil {
		http1opts.EnableTrailers = args.http1Settings.EnableTrailers
		if args.http1Settings.PreserveHeaderCase {
			preservecaseAny, _ := proto.ToAnyWithValidation(&preservecasev3.PreserveCaseFormatterConfig{})
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
	}

	// When setting any Typed Extension Protocol Options, UpstreamProtocolOptions are mandatory
	// If translation requires HTTP2 enablement or HTTP1 trailers, set appropriate setting
	// Default to http1 otherwise
	// TODO: If the cluster is TLS enabled, use AutoHTTPConfig instead of ExplicitHttpConfig
	// so that when ALPN is supported then enabling http1 options doesn't force HTTP/1.1
	switch {
	case args.useClientProtocol:
		protocolOptions.UpstreamProtocolOptions = &httpv3.HttpProtocolOptions_UseDownstreamProtocolConfig{
			UseDownstreamProtocolConfig: &httpv3.HttpProtocolOptions_UseDownstreamHttpConfig{
				HttpProtocolOptions:  http1opts,
				Http2ProtocolOptions: buildHTTP2Settings(args.http2Settings),
			},
		}
	case requiresHTTP2Options:
		protocolOptions.UpstreamProtocolOptions = &httpv3.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &httpv3.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
					Http2ProtocolOptions: buildHTTP2Settings(args.http2Settings),
				},
			},
		}
	case requiresHTTP1Options:
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

	anyProtocolOptions, _ := proto.ToAnyWithValidation(&protocolOptions)

	extensionOptions := map[string]*anypb.Any{
		extensionOptionsKey: anyProtocolOptions,
	}

	return extensionOptions
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
		rawCtxAny, err := proto.ToAnyWithValidation(rawCtx)
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

	ppCtxAny, err := proto.ToAnyWithValidation(ppCtx)
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

func buildAddress(irEp *ir.DestinationEndpoint) *corev3.Address {
	if irEp.Path != nil {
		return &corev3.Address{
			Address: &corev3.Address_Pipe{
				Pipe: &corev3.Pipe{
					Path: *irEp.Path,
				},
			},
		}
	}
	return &corev3.Address{
		Address: &corev3.Address_SocketAddress{
			SocketAddress: &corev3.SocketAddress{
				Protocol: corev3.SocketAddress_TCP,
				Address:  irEp.Host,
				PortSpecifier: &corev3.SocketAddress_PortValue{
					PortValue: irEp.Port,
				},
			},
		},
	}
}

func buildBackandConnectionBufferLimitBytes(bc *ir.BackendConnection) *wrappers.UInt32Value {
	if bc != nil && bc.BufferLimitBytes != nil {
		return wrapperspb.UInt32(*bc.BufferLimitBytes)
	}

	return wrapperspb.UInt32(tcpClusterPerConnectionBufferLimitBytes)
}

type ExtraArgs struct {
	metrics       *ir.Metrics
	http1Settings *ir.HTTP1Settings
	http2Settings *ir.HTTP2Settings
	ipFamily      *egv1a1.IPFamily
}

type clusterArgs interface {
	asClusterArgs(name string, settings []*ir.DestinationSetting, extras *ExtraArgs) *xdsClusterArgs
}

type UDPRouteTranslator struct {
	*ir.UDPRoute
}

func (route *UDPRouteTranslator) asClusterArgs(name string,
	settings []*ir.DestinationSetting,
	extra *ExtraArgs,
) *xdsClusterArgs {
	return &xdsClusterArgs{
		name:         name,
		settings:     settings,
		loadBalancer: route.LoadBalancer,
		endpointType: buildEndpointType(route.Destination.Settings),
		metrics:      extra.metrics,
		dns:          route.DNS,
		ipFamily:     extra.ipFamily,
	}
}

type TCPRouteTranslator struct {
	*ir.TCPRoute
}

func (route *TCPRouteTranslator) asClusterArgs(name string,
	settings []*ir.DestinationSetting,
	extra *ExtraArgs,
) *xdsClusterArgs {
	return &xdsClusterArgs{
		name:              name,
		settings:          settings,
		loadBalancer:      route.LoadBalancer,
		proxyProtocol:     route.ProxyProtocol,
		circuitBreaker:    route.CircuitBreaker,
		tcpkeepalive:      route.TCPKeepalive,
		healthCheck:       route.HealthCheck,
		timeout:           route.Timeout,
		endpointType:      buildEndpointType(route.Destination.Settings),
		metrics:           extra.metrics,
		backendConnection: route.BackendConnection,
		dns:               route.DNS,
		ipFamily:          extra.ipFamily,
	}
}

type HTTPRouteTranslator struct {
	*ir.HTTPRoute
}

func (httpRoute *HTTPRouteTranslator) asClusterArgs(name string,
	settings []*ir.DestinationSetting,
	extra *ExtraArgs,
) *xdsClusterArgs {
	clusterArgs := &xdsClusterArgs{
		name:              name,
		settings:          settings,
		tSocket:           nil,
		endpointType:      buildEndpointType(httpRoute.Destination.Settings),
		metrics:           extra.metrics,
		http1Settings:     extra.http1Settings,
		http2Settings:     extra.http2Settings,
		useClientProtocol: ptr.Deref(httpRoute.UseClientProtocol, false),
		ipFamily:          extra.ipFamily,
	}

	// Populate traffic features.
	bt := httpRoute.Traffic
	if bt != nil {
		clusterArgs.loadBalancer = bt.LoadBalancer
		clusterArgs.proxyProtocol = bt.ProxyProtocol
		clusterArgs.circuitBreaker = bt.CircuitBreaker
		clusterArgs.healthCheck = bt.HealthCheck
		clusterArgs.timeout = bt.Timeout
		clusterArgs.tcpkeepalive = bt.TCPKeepalive
		clusterArgs.backendConnection = bt.BackendConnection
		clusterArgs.dns = bt.DNS
	}

	return clusterArgs
}

func buildHTTP2Settings(opts *ir.HTTP2Settings) *corev3.Http2ProtocolOptions {
	if opts == nil {
		opts = &ir.HTTP2Settings{}
	}

	// defaults based on https://www.envoyproxy.io/docs/envoy/latest/configuration/best_practices/edge
	out := &corev3.Http2ProtocolOptions{
		InitialStreamWindowSize: &wrapperspb.UInt32Value{
			Value: ptr.Deref(opts.InitialStreamWindowSize, http2InitialStreamWindowSize),
		},
		InitialConnectionWindowSize: &wrapperspb.UInt32Value{
			Value: ptr.Deref(opts.InitialConnectionWindowSize, http2InitialConnectionWindowSize),
		},
	}

	if opts.MaxConcurrentStreams != nil {
		out.MaxConcurrentStreams = &wrapperspb.UInt32Value{
			Value: *opts.MaxConcurrentStreams,
		}
	}

	if opts.ResetStreamOnError != nil {
		out.OverrideStreamErrorOnInvalidHttpMessage = &wrapperspb.BoolValue{
			Value: *opts.ResetStreamOnError,
		}
	}

	return out
}
