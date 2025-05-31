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
	dfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/clusters/dynamic_forward_proxy/v3"
	commondfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/dynamic_forward_proxy/v3"
	codecv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/upstream_codec/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	preservecasev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/header_formatters/preserve_case/v3"
	commonv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/common/v3"
	least_requestv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/least_request/v3"
	maglevv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/maglev/v3"
	randomv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/random/v3"
	round_robinv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/round_robin/v3"
	proxyprotocolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/proxy_protocol/v3"
	rawbufferv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/raw_buffer/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
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
	metadata          *ir.ResourceMetadata
	statName          *string
}

type EndpointType int

const (
	EndpointTypeDNS EndpointType = iota
	EndpointTypeStatic
	EndpointTypeDynamicResolver
)

func buildEndpointType(settings []*ir.DestinationSetting) EndpointType {
	// Get endpoint address type for xds cluster by returning the first DestinationSetting's AddressType,
	// since there's no Mixed AddressType among all the DestinationSettings.
	if settings == nil {
		return EndpointTypeStatic
	}

	if settings[0].IsDynamicResolver {
		return EndpointTypeDynamicResolver
	}

	addrType := settings[0].AddressType

	if addrType != nil && *addrType == ir.FQDN {
		return EndpointTypeDNS
	}

	return EndpointTypeStatic
}

type buildClusterResult struct {
	cluster *clusterv3.Cluster
	secrets []*tlsv3.Secret // Secrets used in the cluster filters, we may need to add other types of resources in the future.
}

func buildXdsCluster(args *xdsClusterArgs) (*buildClusterResult, error) {
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
		Metadata:                      buildXdsMetadata(args.metadata),
	}

	if args.statName != nil {
		cluster.AltStatName = *args.statName
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
		// If zone aware routing is enabled we update the cluster lb config
		if ds.ZoneAwareRoutingEnabled {
			cluster.CommonLbConfig.LocalityConfigSpecifier = &clusterv3.Cluster_CommonLbConfig_ZoneAwareLbConfig_{ZoneAwareLbConfig: &clusterv3.Cluster_CommonLbConfig_ZoneAwareLbConfig{}}
		}

		if ds.TLS != nil {
			socket, err := buildXdsUpstreamTLSSocketWthCert(ds.TLS)
			if err != nil {
				// TODO: Log something here
				return nil, err
			}
			if args.proxyProtocol != nil {
				socket = buildProxyProtocolSocket(args.proxyProtocol, socket)
			}
			matchName := fmt.Sprintf("%s/tls/%d", args.name, i)

			// Dynamic resolver clusters have no endpoints, so we need to set the transport socket directly.
			if args.endpointType == EndpointTypeDynamicResolver {
				cluster.TransportSocket = socket
			} else {
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
	}

	// build common, HTTP/1 and HTTP/2  protocol options for cluster
	epo, secrets, err := buildTypedExtensionProtocolOptions(args)
	if err != nil {
		return nil, err
	}
	if epo != nil {
		cluster.TypedExtensionProtocolOptions = epo
	}

	// Set Load Balancer policy
	//nolint:gocritic
	switch {
	case args.loadBalancer == nil:
		cluster.LbPolicy = clusterv3.Cluster_LEAST_REQUEST
	case args.loadBalancer.LeastRequest != nil:
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
	case args.loadBalancer.RoundRobin != nil:
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
	case args.loadBalancer.Random != nil:
		cluster.LbPolicy = clusterv3.Cluster_RANDOM
	case args.loadBalancer.ConsistentHash != nil:
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

	switch args.endpointType {
	case EndpointTypeDynamicResolver:
		dnsCacheConfig := &commondfpv3.DnsCacheConfig{
			Name:            args.name,
			DnsLookupFamily: dnsLookupFamily,
		}
		dnsCacheConfig.DnsRefreshRate = durationpb.New(30 * time.Second)
		if args.dns != nil {
			if args.dns.DNSRefreshRate != nil {
				if args.dns.DNSRefreshRate.Duration > 0 {
					dnsCacheConfig.DnsRefreshRate = durationpb.New(args.dns.DNSRefreshRate.Duration)
				}
			}
		}

		dfp := &dfpv3.ClusterConfig{
			ClusterImplementationSpecifier: &dfpv3.ClusterConfig_DnsCacheConfig{
				DnsCacheConfig: dnsCacheConfig,
			},
		}
		dfpAny, err := proto.ToAnyWithValidation(dfp)
		if err != nil {
			return nil, err
		}
		cluster.ClusterDiscoveryType = &clusterv3.Cluster_ClusterType{ClusterType: &clusterv3.Cluster_CustomClusterType{
			Name:        args.name,
			TypedConfig: dfpAny,
		}}
		cluster.LbPolicy = clusterv3.Cluster_CLUSTER_PROVIDED

	case EndpointTypeStatic:
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
	default:
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

	if args.endpointType != EndpointTypeDynamicResolver {
		for _, ds := range args.settings {
			buildZoneAwareRoutingCluster(ds.ZoneAwareRoutingEnabled, cluster, args.loadBalancer)
		}
	}

	return &buildClusterResult{
		cluster: cluster,
		secrets: secrets,
	}, nil
}

// buildZoneAwareRoutingCluster configures an xds cluster with Zone Aware Routing configuration. It overrides
// cluster.LbPolicy and cluster.CommonLbConfig with cluster.LoadBalancingPolicy.
// TODO: Remove cluster.LbPolicy along with clustercommongLbConfig and switch to cluster.LoadBalancingPolicy
// everywhere as the preferred and more feature-rich configuration field.
func buildZoneAwareRoutingCluster(enabled bool, cluster *clusterv3.Cluster, lb *ir.LoadBalancer) {
	if !enabled {
		return
	}

	// Remove CommonLbConfig.LocalityConfigSpecifier and instead configure via cluster.LoadBalancingPolicy
	cluster.CommonLbConfig.LocalityConfigSpecifier = nil

	localityLbConfig := &commonv3.LocalityLbConfig{
		LocalityConfigSpecifier: &commonv3.LocalityLbConfig_ZoneAwareLbConfig_{
			ZoneAwareLbConfig: &commonv3.LocalityLbConfig_ZoneAwareLbConfig{
				// Future enhancement: differentiate between topology-aware-routing and trafficDistribution
				// once https://github.com/envoyproxy/envoy/pull/39058 is merged
				MinClusterSize:             wrapperspb.UInt64(1),
				ForceLocalityDirectRouting: true,
			},
		},
	}

	// Default to least request LoadBalancingPolicy
	leastRequest := &least_requestv3.LeastRequest{
		LocalityLbConfig: localityLbConfig,
	}
	typedLeastRequest, _ := anypb.New(leastRequest)
	cluster.LoadBalancingPolicy = &clusterv3.LoadBalancingPolicy{
		Policies: []*clusterv3.LoadBalancingPolicy_Policy{{
			TypedExtensionConfig: &corev3.TypedExtensionConfig{
				Name:        "envoy.load_balancing_policies.least_request",
				TypedConfig: typedLeastRequest,
			},
		}},
	}

	if lb != nil {
		switch cluster.LbPolicy {
		case clusterv3.Cluster_LEAST_REQUEST:
			if lb.LeastRequest != nil && lb.LeastRequest.SlowStart != nil && lb.LeastRequest.SlowStart.Window != nil {
				leastRequest.SlowStartConfig = &commonv3.SlowStartConfig{
					SlowStartWindow: durationpb.New(lb.LeastRequest.SlowStart.Window.Duration),
				}
			}
			cluster.LoadBalancingPolicy.Policies[0].TypedExtensionConfig.TypedConfig, _ = anypb.New(leastRequest)
		case clusterv3.Cluster_ROUND_ROBIN:
			roundRobin := &round_robinv3.RoundRobin{
				LocalityLbConfig: localityLbConfig,
			}
			if lb.RoundRobin.SlowStart != nil && lb.RoundRobin.SlowStart.Window != nil {
				roundRobin.SlowStartConfig = &commonv3.SlowStartConfig{
					SlowStartWindow: durationpb.New(lb.RoundRobin.SlowStart.Window.Duration),
				}
			}
			typedRoundRobin, _ := anypb.New(roundRobin)
			cluster.LoadBalancingPolicy = &clusterv3.LoadBalancingPolicy{
				Policies: []*clusterv3.LoadBalancingPolicy_Policy{{
					TypedExtensionConfig: &corev3.TypedExtensionConfig{
						Name:        "envoy.load_balancing_policies.round_robin",
						TypedConfig: typedRoundRobin,
					},
				}},
			}

		case clusterv3.Cluster_RANDOM:
			random := &randomv3.Random{
				LocalityLbConfig: localityLbConfig,
			}
			typeRandom, _ := anypb.New(random)
			cluster.LoadBalancingPolicy = &clusterv3.LoadBalancingPolicy{
				Policies: []*clusterv3.LoadBalancingPolicy_Policy{{
					TypedExtensionConfig: &corev3.TypedExtensionConfig{
						Name:        "envoy.load_balancing_policies.random",
						TypedConfig: typeRandom,
					},
				}},
			}
		case clusterv3.Cluster_MAGLEV:
			consistentHash := &maglevv3.Maglev{}
			if lb.ConsistentHash.TableSize != nil {
				consistentHash.TableSize = wrapperspb.UInt64(*lb.ConsistentHash.TableSize)
			}
			typedConsistentHash, _ := anypb.New(consistentHash)
			cluster.LoadBalancingPolicy = &clusterv3.LoadBalancingPolicy{
				Policies: []*clusterv3.LoadBalancingPolicy_Policy{{
					TypedExtensionConfig: &corev3.TypedExtensionConfig{
						Name:        "envoy.load_balancing_policies.maglev",
						TypedConfig: typedConsistentHash,
					},
				}},
			}
		}
	}
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
	sort.Slice(irStatuses, func(i, j int) bool {
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
	cbtPerEndpoint := []*clusterv3.CircuitBreakers_Thresholds{}

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

		if circuitBreaker.PerEndpoint != nil {
			if circuitBreaker.PerEndpoint.MaxConnections != nil {
				cbtPerEndpoint = []*clusterv3.CircuitBreakers_Thresholds{
					{
						MaxConnections: &wrapperspb.UInt32Value{
							Value: *circuitBreaker.PerEndpoint.MaxConnections,
						},
					},
				}
			}
		}
	}

	ecb := &clusterv3.CircuitBreakers{
		Thresholds:        []*clusterv3.CircuitBreakers_Thresholds{cbt},
		PerHostThresholds: cbtPerEndpoint,
	}

	return ecb
}

func buildXdsClusterLoadAssignment(clusterName string, destSettings []*ir.DestinationSetting) *endpointv3.ClusterLoadAssignment {
	localities := make([]*endpointv3.LocalityLbEndpoints, 0, len(destSettings))
	for i, ds := range destSettings {

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

		// If zone aware routing is enabled for a backendRefs we include endpoint zone info in localities.
		// Note: The locality.LoadBalancingWeight field applies only when using locality-weighted LB config
		// so in order to support traffic splitting we rely on weighted clusters defined at the route level
		// if multiple backendRefs exist. This pushes part of the routing logic higher up the stack which can
		// limit host selection controls during retries and session affinity.
		// For more details see https://github.com/envoyproxy/gateway/issues/5307#issuecomment-2688767482
		if ds.ZoneAwareRoutingEnabled {
			localities = append(localities, buildZonalLocalities(metadata, ds)...)
		} else {
			localities = append(localities, buildWeightedLocalities(metadata, ds))
		}
	}
	return &endpointv3.ClusterLoadAssignment{ClusterName: clusterName, Endpoints: localities}
}

func buildZonalLocalities(metadata *corev3.Metadata, ds *ir.DestinationSetting) []*endpointv3.LocalityLbEndpoints {
	var localities []*endpointv3.LocalityLbEndpoints
	zonalEndpoints := make(map[string][]*endpointv3.LbEndpoint)
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
			LoadBalancingWeight: wrapperspb.UInt32(1),
			HealthStatus:        healthStatus,
		}

		zone := ptr.Deref(irEp.Zone, "")
		zonalEndpoints[zone] = append(zonalEndpoints[zone], lbEndpoint)
	}

	for zone, endPts := range zonalEndpoints {
		locality := &endpointv3.LocalityLbEndpoints{
			Locality: &corev3.Locality{
				Zone: zone,
			},
			LbEndpoints: endPts,
			Priority:    ptr.Deref(ds.Priority, 0),
			Metadata:    buildXdsMetadata(ds.Metadata),
		}
		localities = append(localities, locality)
	}
	// Sort localities by zone, so that the order is deterministic.
	sort.Slice(localities, func(i, j int) bool {
		return localities[i].Locality.Zone < localities[j].Locality.Zone
	})

	return localities
}

func buildWeightedLocalities(metadata *corev3.Metadata, ds *ir.DestinationSetting) *endpointv3.LocalityLbEndpoints {
	endpoints := make([]*endpointv3.LbEndpoint, 0, len(ds.Endpoints))

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
		Locality: &corev3.Locality{
			Region: ds.Name,
		},
		LbEndpoints: endpoints,
		Priority:    0,
		Metadata:    buildXdsMetadata(ds.Metadata),
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
	return locality
}

func buildTypedExtensionProtocolOptions(args *xdsClusterArgs) (map[string]*anypb.Any, []*tlsv3.Secret, error) {
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

	requiresHTTP1Options := args.http1Settings != nil &&
		(args.http1Settings.EnableTrailers || args.http1Settings.PreserveHeaderCase || args.http1Settings.HTTP10 != nil)

	requiresHTTPFilters := len(args.settings) > 0 && args.settings[0].Filters != nil && args.settings[0].Filters.CredentialInjection != nil

	if !requiresCommonHTTPOptions && !requiresHTTP1Options && !requiresHTTP2Options && !args.useClientProtocol && !requiresHTTPFilters {
		return nil, nil, nil
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

	// Envoy proxy requires AutoSNI and AutoSanValidation to be set to true for dynamic_forward_proxy clusters
	if args.endpointType == EndpointTypeDynamicResolver {
		protocolOptions.UpstreamHttpProtocolOptions = &corev3.UpstreamHttpProtocolOptions{
			AutoSni:           true,
			AutoSanValidation: true,
		}
	}

	var (
		filters []*hcmv3.HttpFilter
		secrets []*tlsv3.Secret
		err     error
	)
	if requiresHTTPFilters {
		filters, secrets, err = buildClusterHTTPFilters(args)
		if err != nil {
			return nil, nil, err
		}
		if len(filters) > 0 {
			protocolOptions.HttpFilters = filters
		}
	}

	anyProtocolOptions, _ := proto.ToAnyWithValidation(&protocolOptions)

	extensionOptions := map[string]*anypb.Any{
		extensionOptionsKey: anyProtocolOptions,
	}

	return extensionOptions, secrets, nil
}

// buildClusterHTTPFilters builds the HTTP filters for the cluster.
// EG only supports credential injector filter for now, more filters can be added in the future.
func buildClusterHTTPFilters(args *xdsClusterArgs) ([]*hcmv3.HttpFilter, []*tlsv3.Secret, error) {
	filters := make([]*hcmv3.HttpFilter, 0)
	secrets := make([]*tlsv3.Secret, 0)
	if len(args.settings) > 0 {
		// There is only one setting in the settings slice because EG creates one cluster per backendRef
		// if there are backend filters.
		if args.settings[0].Filters != nil && args.settings[0].Filters.CredentialInjection != nil {
			filter, err := buildHCMCredentialInjectorFilter(args.settings[0].Filters.CredentialInjection)
			filter.Disabled = false
			if err != nil {
				return nil, nil, err
			}
			secret := buildCredentialSecret(args.settings[0].Filters.CredentialInjection)
			filters = append(filters, filter)
			secrets = append(secrets, secret)
		}
	}

	// UpstreamCodec filter is required as the terminal filter for the upstream HTTP filters.
	if len(filters) > 0 {
		upstreamCodec, err := buildUpstreamCodecFilter()
		if err != nil {
			return nil, nil, err
		}
		filters = append(filters, upstreamCodec)
	}
	// We may need to add more Cluster filters in the future, so we return a slice of filters.
	return filters, secrets, nil
}

func buildUpstreamCodecFilter() (*hcmv3.HttpFilter, error) {
	codec := &codecv3.UpstreamCodec{}
	codecAny, err := proto.ToAnyWithValidation(codec)
	if err != nil {
		return nil, err
	}
	return &hcmv3.HttpFilter{
		Name: "envoy.extensions.filters.http.upstream_codec.v3.UpstreamCodec",
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: codecAny,
		},
	}, nil
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
	statName      *string
}

type clusterArgs interface {
	asClusterArgs(name string, settings []*ir.DestinationSetting, extras *ExtraArgs, metadata *ir.ResourceMetadata) *xdsClusterArgs
}

type UDPRouteTranslator struct {
	*ir.UDPRoute
}

func (route *UDPRouteTranslator) asClusterArgs(name string,
	settings []*ir.DestinationSetting,
	extra *ExtraArgs,
	metadata *ir.ResourceMetadata,
) *xdsClusterArgs {
	return &xdsClusterArgs{
		name:         name,
		settings:     settings,
		loadBalancer: route.LoadBalancer,
		endpointType: buildEndpointType(route.Destination.Settings),
		metrics:      extra.metrics,
		dns:          route.DNS,
		ipFamily:     extra.ipFamily,
		metadata:     metadata,
	}
}

type TCPRouteTranslator struct {
	*ir.TCPRoute
}

func (route *TCPRouteTranslator) asClusterArgs(name string,
	settings []*ir.DestinationSetting,
	extra *ExtraArgs,
	metadata *ir.ResourceMetadata,
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
		metadata:          metadata,
	}
}

type HTTPRouteTranslator struct {
	*ir.HTTPRoute
}

func (httpRoute *HTTPRouteTranslator) asClusterArgs(name string,
	settings []*ir.DestinationSetting,
	extra *ExtraArgs,
	metadata *ir.ResourceMetadata,
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
		metadata:          metadata,
		statName:          extra.statName,
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
