// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"net/url"
	"strconv"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	ratelimitfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	ratelimitserviceconfig "github.com/envoyproxy/ratelimit/src/config"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	goyaml "gopkg.in/yaml.v3" // nolint: depguard

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// patchHCMWithRateLimit builds and appends the Rate Limit Filter to the HTTP connection manager
// if applicable and it does not already exist.
func (t *Translator) patchHCMWithRateLimit(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) {
	// Return early if rate limits dont exist
	if !t.isRateLimitPresent(irListener) {
		return
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == wellknown.HTTPRateLimit {
			return
		}
	}

	rateLimitFilter := buildRateLimitFilter(irListener)
	// Make sure the router filter is the terminal filter in the chain.
	mgr.HttpFilters = append([]*hcmv3.HttpFilter{rateLimitFilter}, mgr.HttpFilters...)
}

// isRateLimitPresent returns true if rate limit config exists for the listener.
func (t *Translator) isRateLimitPresent(irListener *ir.HTTPListener) bool {
	// Return false if global ratelimiting is disabled.
	if t.GlobalRateLimit == nil {
		return false
	}
	// Return true if rate limit config exists.
	for _, route := range irListener.Routes {
		if route.RateLimit != nil && route.RateLimit.Global != nil {
			return true
		}
	}
	return false
}

func buildRateLimitFilter(irListener *ir.HTTPListener) *hcmv3.HttpFilter {
	rateLimitFilterProto := &ratelimitfilterv3.RateLimit{
		Domain: getRateLimitDomain(irListener),
		RateLimitService: &ratelimitv3.RateLimitServiceConfig{
			GrpcService: &corev3.GrpcService{
				TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
						ClusterName: getRateLimitServiceClusterName(),
					},
				},
			},
			TransportApiVersion: corev3.ApiVersion_V3,
		},
	}

	rateLimitFilterAny, err := anypb.New(rateLimitFilterProto)
	if err != nil {
		return nil
	}

	rateLimitFilter := &hcmv3.HttpFilter{
		Name: wellknown.HTTPRateLimit,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
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

		if rule.CIDRMatch != nil {
			// Setup MaskedRemoteAddress action
			mra := &routev3.RateLimit_Action_MaskedRemoteAddress{}
			maskLen := &wrapperspb.UInt32Value{Value: uint32(rule.CIDRMatch.MaskLen)}
			if rule.CIDRMatch.IPv6 {
				mra.V6PrefixMaskLen = maskLen
			} else {
				mra.V4PrefixMaskLen = maskLen
			}
			action := &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_MaskedRemoteAddress_{
					MaskedRemoteAddress: mra,
				},
			}
			rlActions = append(rlActions, action)
		}

		// Case when header match is not set and the rate limit is applied
		// to all traffic.
		if !rule.IsMatchSet() {
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
	yamlDescriptors := make([]ratelimitserviceconfig.YamlDescriptor, 0, 1)

	for _, route := range irListener.Routes {
		if route.RateLimit != nil && route.RateLimit.Global != nil {
			serviceDescriptors := buildRateLimitServiceDescriptors(route.Name, route.RateLimit.Global)
			yamlDescriptors = append(yamlDescriptors, serviceDescriptors...)
		}
	}

	if len(yamlDescriptors) == 0 {
		return nil
	}

	return &ratelimitserviceconfig.YamlRoot{
		Domain:      getRateLimitDomain(irListener),
		Descriptors: yamlDescriptors,
	}
}

// buildRateLimitServiceDescriptors creates the rate limit service yaml descriptors based on the global rate limit IR config.
func buildRateLimitServiceDescriptors(descriptorPrefix string, global *ir.GlobalRateLimit) []ratelimitserviceconfig.YamlDescriptor {
	yamlDescriptors := make([]ratelimitserviceconfig.YamlDescriptor, 0, 1)

	for rIdx, rule := range global.Rules {
		var head, cur *ratelimitserviceconfig.YamlDescriptor
		if !rule.IsMatchSet() {
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

		if rule.CIDRMatch != nil {
			// MaskedRemoteAddress case
			yamlDesc := new(ratelimitserviceconfig.YamlDescriptor)
			yamlDesc.Key = "masked_remote_address"
			yamlDesc.Value = rule.CIDRMatch.CIDR
			rateLimit := ratelimitserviceconfig.YamlRateLimit{
				RequestsPerUnit: uint32(rule.Limit.Requests),
				Unit:            string(rule.Limit.Unit),
			}
			yamlDesc.RateLimit = &rateLimit

			head = yamlDesc
			cur = head
		}

		yamlDescriptors = append(yamlDescriptors, *head)
	}

	return yamlDescriptors
}

func (t *Translator) createRateLimitServiceCluster(tCtx *types.ResourceVersionTable, irListener *ir.HTTPListener) error {
	// Return early if rate limits dont exist.
	if !t.isRateLimitPresent(irListener) {
		return nil
	}
	clusterName := getRateLimitServiceClusterName()
	if rlCluster := findXdsCluster(tCtx, clusterName); rlCluster == nil {
		// Create cluster if it does not exist
		host, port := t.getRateLimitServiceGrpcHostPort()
		routeDestinations := []*ir.RouteDestination{ir.NewRouteDest(host, uint32(port))}
		addXdsCluster(tCtx, addXdsClusterArgs{
			name:         clusterName,
			destinations: routeDestinations,
			tSocket:      nil,
			protocol:     HTTP2,
			endpoint:     DefaultEndpointType,
		})
	}
	return nil
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

func (t *Translator) getRateLimitServiceGrpcHostPort() (string, int) {
	u, err := url.Parse(t.GlobalRateLimit.ServiceURL)
	if err != nil {
		panic(err)
	}
	p, err := strconv.Atoi(u.Port())
	if err != nil {
		panic(err)
	}
	return u.Hostname(), p
}
