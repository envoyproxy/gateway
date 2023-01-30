// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"strconv"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	ratelimit "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	ratelimitfilter "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	wkt "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	ratelimitserviceconfig "github.com/envoyproxy/ratelimit/src/config"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	goyaml "gopkg.in/yaml.v3" // nolint: depguard

	"github.com/envoyproxy/gateway/internal/ir"
)

// patchHCMWithRateLimit builds and appends the Rate Limit Filter to the HTTP connection manager
// if applicable and it does not already exist.
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

// isRateLimitPresent returns true if rate limit config exists for the listener.
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

	rateLimitFilterAny, err := anypb.New(rateLimitFilterProto)
	if err != nil {
		return nil
	}

	rateLimitFilter := &hcm.HttpFilter{
		Name: wkt.HTTPRateLimit,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: rateLimitFilterAny,
		},
	}
	return rateLimitFilter
}

// patchRouteWithRateLimit builds rate limit actions and appends to the route.
func patchRouteWithRateLimit(xdsRouteAction *routev3.RouteAction, irRoute *ir.HTTPRoute) error { //nolint:unparam
	// Return early if no rate limit config exists.
	if irRoute.RateLimit == nil || irRoute.RateLimit.Global == nil {
		return nil
	}

	rateLimits := buildRouteRateLimits(irRoute.Name, irRoute.RateLimit.Global)
	xdsRouteAction.RateLimits = rateLimits
	return nil
}

func buildRouteRateLimits(descriptorPrefix string, global *ir.GlobalRateLimit) []*routev3.RateLimit {
	rateLimits := []*routev3.RateLimit{}
	// Rules are ORed
	for rIdx, rule := range global.Rules {
		rlActions := []*routev3.RateLimit_Action{}
		// Matches are ANDed
		for mIdx, match := range rule.HeaderMatches {
			// Case for distinct match
			if match.Distinct {
				// Setup RequestHeader actions
				descriptorKey := getRateLimitDescriptorKey(descriptorPrefix, rIdx, mIdx)
				action := &routev3.RateLimit_Action{
					ActionSpecifier: &routev3.RateLimit_Action_RequestHeaders_{
						RequestHeaders: &routev3.RateLimit_Action_RequestHeaders{
							HeaderName:    match.Name,
							DescriptorKey: descriptorKey,
						},
					},
				}
				rlActions = append(rlActions, action)
			} else {
				// Setup HeaderValueMatch actions
				descriptorKey := getRateLimitDescriptorKey(descriptorPrefix, rIdx, mIdx)
				descriptorVal := getRateLimitDescriptorValue(descriptorPrefix, rIdx, mIdx)
				headerMatcher := &routev3.HeaderMatcher{
					Name: match.Name,
					HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
						StringMatch: buildXdsStringMatcher(match),
					},
				}
				action := &routev3.RateLimit_Action{
					ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
						HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
							DescriptorKey:   descriptorKey,
							DescriptorValue: descriptorVal,
							ExpectMatch: &wrapperspb.BoolValue{
								Value: true,
							},
							Headers: []*routev3.HeaderMatcher{headerMatcher},
						},
					},
				}
				rlActions = append(rlActions, action)
			}
		}

		// Case when header match is not set and the rate limit is applied
		// to all traffic.
		if len(rule.HeaderMatches) == 0 {
			// Setup GenericKey action
			action := &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_GenericKey_{
					GenericKey: &routev3.RateLimit_Action_GenericKey{
						DescriptorKey:   getRateLimitDescriptorKey(descriptorPrefix, rIdx, -1),
						DescriptorValue: getRateLimitDescriptorValue(descriptorPrefix, rIdx, -1),
					},
				},
			}
			rlActions = append(rlActions, action)
		}

		rateLimit := &routev3.RateLimit{Actions: rlActions}
		rateLimits = append(rateLimits, rateLimit)
	}

	return rateLimits
}

// GetRateLimitServiceConfigStr returns the YAML string for the rate limit service configuration.
func GetRateLimitServiceConfigStr(yamlRoot *ratelimitserviceconfig.YamlRoot) (string, error) {
	var buf bytes.Buffer
	enc := goyaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err := enc.Encode(*yamlRoot)
	return buf.String(), err
}

// BuildRateLimitServiceConfig builds the rate limit service configuration based on
// https://github.com/envoyproxy/ratelimit#the-configuration-format
func BuildRateLimitServiceConfig(irListener *ir.HTTPListener) *ratelimitserviceconfig.YamlRoot {
	yamlDescs := make([]ratelimitserviceconfig.YamlDescriptor, 0, 1)

	for _, route := range irListener.Routes {
		if route.RateLimit != nil && route.RateLimit.Global != nil {
			descs := buildRateLimitServiceDescriptors(route.Name, route.RateLimit.Global)
			yamlDescs = append(yamlDescs, descs...)
		}
	}

	if len(yamlDescs) == 0 {
		return nil
	}

	return &ratelimitserviceconfig.YamlRoot{
		Domain:      getRateLimitDomain(irListener),
		Descriptors: yamlDescs,
	}
}

// buildRateLimitServiceDescriptors creates the rate limit service yaml descriptors based on the global rate limit IR config.
func buildRateLimitServiceDescriptors(descriptorPrefix string, global *ir.GlobalRateLimit) []ratelimitserviceconfig.YamlDescriptor {
	yamlDescs := make([]ratelimitserviceconfig.YamlDescriptor, 0, 1)

	for rIdx, rule := range global.Rules {
		var head, cur *ratelimitserviceconfig.YamlDescriptor
		if len(rule.HeaderMatches) == 0 {
			yamlDesc := new(ratelimitserviceconfig.YamlDescriptor)
			// GenericKey case
			yamlDesc.Key = getRateLimitDescriptorKey(descriptorPrefix, rIdx, -1)
			yamlDesc.Value = getRateLimitDescriptorValue(descriptorPrefix, rIdx, -1)
			rateLimit := ratelimitserviceconfig.YamlRateLimit{
				RequestsPerUnit: uint32(rule.Limit.Requests),
				Unit:            string(rule.Limit.Unit),
			}
			yamlDesc.RateLimit = &rateLimit

			head = yamlDesc
			cur = head
		}

		for mIdx, match := range rule.HeaderMatches {
			yamlDesc := new(ratelimitserviceconfig.YamlDescriptor)
			// Case for distinct match
			if match.Distinct {
				// RequestHeader case
				yamlDesc.Key = getRateLimitDescriptorKey(descriptorPrefix, rIdx, mIdx)
			} else {
				// HeaderValueMatch case
				yamlDesc.Key = getRateLimitDescriptorKey(descriptorPrefix, rIdx, mIdx)
				yamlDesc.Value = getRateLimitDescriptorValue(descriptorPrefix, rIdx, mIdx)

			}

			// Add the ratelimit values to the last descriptor
			if mIdx == len(rule.HeaderMatches)-1 {
				rateLimit := ratelimitserviceconfig.YamlRateLimit{
					RequestsPerUnit: uint32(rule.Limit.Requests),
					Unit:            string(rule.Limit.Unit),
				}
				yamlDesc.RateLimit = &rateLimit
			}

			if mIdx == 0 {
				head = yamlDesc
			} else {
				cur.Descriptors = []ratelimitserviceconfig.YamlDescriptor{*yamlDesc}
			}

			cur = yamlDesc
		}

		yamlDescs = append(yamlDescs, *head)
	}

	return yamlDescs
}

func buildRateLimitServiceCluster(irListener *ir.HTTPListener) *cluster.Cluster {
	// Return early if rate limits dont exist.
	if !isRateLimitPresent(irListener) {
		return nil
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
	return rateLimitServerCluster
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
