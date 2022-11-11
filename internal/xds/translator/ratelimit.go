// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"strconv"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	ratelimitfilter "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	wkt "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
)

func patchHCMWithRateLimit(mgr *hcm.HttpConnectionManager, irListener *ir.HTTPListener) error {
	// Return early if rate limits dont exist
	if !isRateLimitPresent(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == wkt.HTTPRateLimit {
			return nil
		}
	}

	rateLimitFilter := buildRateLimitFilter(irListener)
	// Make sure the router filter is the terminal filter in the chain
	mgr.HttpFilters = append([]*hcm.HttpFilter{rateLimitFilter}, mgr.HttpFilters...)
	return nil
}

func isRateLimitPresent(irListener *ir.HTTPListener) bool {
	// Return true if rate limit config exists.
	for _, route := range irListener.Routes {
		if route.RateLimit != nil && route.RateLimit.Global != nil {
			return true
		}
	}
	return false
}

func buildRateLimitFilter(irListener *ir.HTTPListener) *hcm.HttpFilter {
	rateLimitFilterProto := &ratelimitfilter.RateLimit{
		Domain: getRateLimitDomain(irListener),
		RateLimitService: &ratelimit.RateLimitServiceConfig{
			GrpcService: &core.GrpcService{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
						ClusterName: getRateLimitServiceClusterName(),
					},
				},
			},
			TransportApiVersion: core.ApiVersion_V3,
		},
	}

	any, err := anypb.New(rateLimitFilterProto)
	if err != nil {
		return nil
	}

	rateLimitFilter := &hcm.HttpFilter{
		Name: wkt.HTTPRateLimit,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: any,
		},
	}
	return rateLimitFilter
}

func PatchRouteWithRateLimit(xdsRouteAction *route.RouteAction, irRoute *ir.HTTPRoute) error {
	// Return early if no rate limit config exists.
	if irRoute.RateLimit == nil || irRoute.RateLimit.Global == nil {
		return nil
	}

	rateLimits := buildRouteRateLimits(irRoute.Name, irRoute.RateLimit.Global)
	xdsRouteAction.RateLimits = rateLimits
	return nil
}

func buildRouteRateLimits(descriptorPrefix string, global *ir.GlobalRateLimit) []*route.RateLimit {
	rateLimits := []*route.RateLimit{}
	// Rules are ORed
	for rIdx, rule := range global.Rules {
		rlActions := []*route.RateLimit_Action{}
		// Matches are ANDed
		for mIdx, match := range rule.HeaderMatches {
			if match.Exact == nil && match.Prefix == nil && match.SafeRegex == nil {
				// Setup RequestHeader actions
				descriptorKey := getRateLimitDescriptorKey(descriptorPrefix, rIdx, mIdx)
				action := &route.RateLimit_Action{
					ActionSpecifier: &route.RateLimit_Action_RequestHeaders_{
						RequestHeaders: &route.RateLimit_Action_RequestHeaders{
							HeaderName:    match.Name,
							DescriptorKey: descriptorKey,
						},
					},
				}
				rlActions = append(rlActions, action)
			} else {
				// Setup HeaderValueMatch actions
				descriptorVal := getRateLimitDescriptorValue(descriptorPrefix, rIdx, mIdx)
				headerMatcher := buildHeaderMatcher(match)
				action := &route.RateLimit_Action{
					ActionSpecifier: &route.RateLimit_Action_HeaderValueMatch_{
						HeaderValueMatch: &route.RateLimit_Action_HeaderValueMatch{
							DescriptorValue: descriptorVal,
							ExpectMatch: &wrapperspb.BoolValue{
								Value: true,
							},
							Headers: []*route.HeaderMatcher{headerMatcher},
						},
					},
				}
				rlActions = append(rlActions, action)
			}
		}

		rateLimit := &route.RateLimit{Actions: rlActions}
		rateLimits = append(rateLimits, rateLimit)
	}

	return rateLimits
}

func buildHeaderMatcher(match *ir.StringMatch) *route.HeaderMatcher {
	var stringMatcher *matcher.StringMatcher

	if match.Exact != nil {
		stringMatcher = &matcher.StringMatcher{
			MatchPattern: &matcher.StringMatcher_Exact{
				Exact: *match.Exact,
			},
		}
	}
	if match.Prefix != nil {
		stringMatcher = &matcher.StringMatcher{
			MatchPattern: &matcher.StringMatcher_Prefix{
				Prefix: *match.Prefix,
			},
		}
	}
	if match.SafeRegex != nil {
		stringMatcher = &matcher.StringMatcher{
			MatchPattern: &matcher.StringMatcher_SafeRegex{
				SafeRegex: &matcher.RegexMatcher{
					Regex: *match.SafeRegex,
					EngineType: &matcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
					},
				},
			},
		}
	}

	return &route.HeaderMatcher{
		Name: match.Name,
		HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
			StringMatch: stringMatcher,
		},
	}
}

func buildRateLimitServiceCluster(irListener *ir.HTTPListener) (*cluster.Cluster, error) {
	// Return early if rate limits dont exist.
	if !isRateLimitPresent(irListener) {
		return nil, nil
	}

	clusterName := getRateLimitServiceClusterName()
	host, port := getRateLimitServiceGrpcHostPort()
	rateLimitServerCluster := &cluster.Cluster{
		Name:                 clusterName,
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
		ConnectTimeout:       durationpb.New(10 * time.Second),
		LbPolicy:             cluster.Cluster_RANDOM,
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*endpoint.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpoint.LbEndpoint{
						{
							HostIdentifier: &endpoint.LbEndpoint_Endpoint{
								Endpoint: &endpoint.Endpoint{
									Address: &core.Address{
										Address: &core.Address_SocketAddress{
											SocketAddress: &core.SocketAddress{
												Address:       host,
												PortSpecifier: &core.SocketAddress_PortValue{PortValue: uint32(port)},
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
		Http2ProtocolOptions: &core.Http2ProtocolOptions{},
		DnsRefreshRate:       durationpb.New(30 * time.Second),
		RespectDnsTtl:        true,
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
	}
	return rateLimitServerCluster, nil
}

func getRateLimitDescriptorKey(prefix string, ruleIndex, matchIndex int) string {
	return prefix + "-key-rule-" + strconv.Itoa(ruleIndex) + "-match-" + strconv.Itoa(matchIndex)
}

func getRateLimitDescriptorValue(prefix string, ruleIndex, matchIndex int) string {
	return prefix + "-value-rule-" + strconv.Itoa(ruleIndex) + "-match-" + strconv.Itoa(matchIndex)
}

func getRateLimitServiceClusterName() string {
	return "ratelimit_cluster"
}

func getRateLimitDomain(irListener *ir.HTTPListener) string {
	// Use IR listener name as domain
	return irListener.Name
}

func getRateLimitServiceGrpcHostPort() (string, int) {
	return "TODO", 0
}
