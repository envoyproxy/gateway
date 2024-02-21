// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	configv3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rlv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	localrlv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	localRateLimitFilter           = "envoy.filters.http.local_ratelimit"
	localRateLimitFilterStatPrefix = "http_local_rate_limiter"
	descriptorMaskedRemoteAddress  = "masked_remote_address"
)

func init() {
	registerHTTPFilter(&localRateLimit{})
}

type localRateLimit struct {
}

var _ httpFilter = &localRateLimit{}

// patchHCM builds and appends the local rate limit filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
func (*localRateLimit) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}
	if !listenerContainsLocalRateLimit(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == localRateLimitFilter {
			return nil
		}
	}

	localRl := &localrlv3.LocalRateLimit{
		StatPrefix: localRateLimitFilterStatPrefix,
	}

	localRlAny, err := anypb.New(localRl)
	if err != nil {
		return err
	}

	// The local rate limit filter at the HTTP connection manager level is an
	// empty filter. The real configuration is done at the route level.
	// See https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/local_rate_limit_filter
	filter := &hcmv3.HttpFilter{
		Name: localRateLimitFilter,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: localRlAny,
		},
	}

	mgr.HttpFilters = append(mgr.HttpFilters, filter)
	return nil
}

func listenerContainsLocalRateLimit(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if routeContainsLocalRateLimit(route) {
			return true
		}
	}

	return false
}

func routeContainsLocalRateLimit(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil || irRoute.RateLimit == nil || irRoute.RateLimit.Local == nil {
		return false
	}

	return true
}

func (*localRateLimit) patchResources(*types.ResourceVersionTable,
	[]*ir.HTTPRoute) error {
	return nil
}

func (*localRateLimit) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	routeAction := route.GetRoute()

	// Return early if no rate limit config exists.
	if irRoute.RateLimit == nil || irRoute.RateLimit.Local == nil || routeAction == nil {
		return nil
	}

	if routeAction.RateLimits != nil {
		// This should not happen since this is the only place where the rate limit
		// config is added in a route.
		return fmt.Errorf(
			"route already contains rate limit config:  %s",
			route.Name)
	}

	local := irRoute.RateLimit.Local

	rateLimits, descriptors, err := buildRouteLocalRateLimits(local)
	if err != nil {
		return err
	}
	routeAction.RateLimits = rateLimits

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[localRateLimitFilter]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf(
			"route already contains local rate limit filter config:  %s",
			route.Name)
	}

	localRl := &localrlv3.LocalRateLimit{
		StatPrefix: localRateLimitFilterStatPrefix,
		TokenBucket: &typev3.TokenBucket{
			MaxTokens: uint32(local.Default.Requests),
			TokensPerFill: &wrapperspb.UInt32Value{
				Value: uint32(local.Default.Requests),
			},
			FillInterval: ratelimitUnitToDuration(local.Default.Unit),
		},
		FilterEnabled: &configv3.RuntimeFractionalPercent{
			DefaultValue: &typev3.FractionalPercent{
				Numerator:   100,
				Denominator: typev3.FractionalPercent_HUNDRED,
			},
		},
		FilterEnforced: &configv3.RuntimeFractionalPercent{
			DefaultValue: &typev3.FractionalPercent{
				Numerator:   100,
				Denominator: typev3.FractionalPercent_HUNDRED,
			},
		},
		Descriptors: descriptors,
		// By setting AlwaysConsumeDefaultTokenBucket to false, the descriptors
		// won't consume the default token bucket. This means that a request only
		// counts towards the default token bucket if it does not match any of the
		// descriptors.
		AlwaysConsumeDefaultTokenBucket: &wrappers.BoolValue{
			Value: false,
		},
	}

	localRlAny, err := anypb.New(localRl)
	if err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[localRateLimitFilter] = localRlAny
	return nil
}

func buildRouteLocalRateLimits(local *ir.LocalRateLimit) (
	[]*routev3.RateLimit, []*rlv3.LocalRateLimitDescriptor, error) {
	var rateLimits []*routev3.RateLimit
	var descriptors []*rlv3.LocalRateLimitDescriptor

	// Rules are ORed
	for rIdx, rule := range local.Rules {
		var rlActions []*routev3.RateLimit_Action
		var descriptorEntries []*rlv3.RateLimitDescriptor_Entry

		// HeaderMatches
		for mIdx, match := range rule.HeaderMatches {
			if match.Distinct {
				// This is a sanity check. This should never happen because Gateway
				// API translator should have already validated this.
				if rule.CIDRMatch.Distinct {
					return nil, nil, errors.New("local rateLimit does not support distinct HeaderMatch")
				}
			}

			// Setup HeaderValueMatch actions
			descriptorKey := getRouteRuleDescriptor(rIdx, mIdx)
			descriptorVal := getRouteRuleDescriptor(rIdx, mIdx)
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
			entry := &rlv3.RateLimitDescriptor_Entry{
				Key:   descriptorKey,
				Value: descriptorVal,
			}
			rlActions = append(rlActions, action)
			descriptorEntries = append(descriptorEntries, entry)
		}

		// Source IP CIDRMatch
		if rule.CIDRMatch != nil {
			// This is a sanity check. This should never happen because Gateway
			// API translator should have already validated this.
			if rule.CIDRMatch.Distinct {
				return nil, nil, errors.New("local rateLimit does not support distinct CIDRMatch")
			}

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
			entry := &rlv3.RateLimitDescriptor_Entry{
				Key:   descriptorMaskedRemoteAddress,
				Value: rule.CIDRMatch.CIDR,
			}
			descriptorEntries = append(descriptorEntries, entry)
			rlActions = append(rlActions, action)
		}

		rateLimit := &routev3.RateLimit{Actions: rlActions}
		rateLimits = append(rateLimits, rateLimit)

		descriptor := &rlv3.LocalRateLimitDescriptor{
			Entries: descriptorEntries,
			TokenBucket: &typev3.TokenBucket{
				MaxTokens: uint32(rule.Limit.Requests),
				TokensPerFill: &wrapperspb.UInt32Value{
					Value: uint32(rule.Limit.Requests),
				},
				FillInterval: ratelimitUnitToDuration(rule.Limit.Unit),
			},
		}
		descriptors = append(descriptors, descriptor)
	}

	return rateLimits, descriptors, nil
}

func ratelimitUnitToDuration(unit ir.RateLimitUnit) *duration.Duration {
	var seconds int64

	switch egv1a1.RateLimitUnit(unit) {
	case egv1a1.RateLimitUnitSecond:
		seconds = 1
	case egv1a1.RateLimitUnitMinute:
		seconds = 60
	case egv1a1.RateLimitUnitHour:
		seconds = 60 * 60
	case egv1a1.RateLimitUnitDay:
		seconds = 60 * 60 * 24
	}
	return &duration.Duration{
		Seconds: seconds,
	}
}
