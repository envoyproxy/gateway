// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestNamedNonSharedGlobalRateLimitRuleUsesRuleName(t *testing.T) {
	rule := &ir.RateLimitRule{
		Name: "default/policy/rule/by-org",
		HeaderMatches: []*ir.StringMatch{
			{
				Name:  "x-org",
				Exact: ptr.To("acme"),
			},
		},
		Limit: ir.RateLimitValue{
			Requests: 10,
			Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
		},
	}
	route := &ir.HTTPRoute{
		Name: "route-a",
		Traffic: &ir.TrafficFeatures{
			RateLimit: &ir.RateLimit{
				Global: &ir.GlobalRateLimit{
					Rules: []*ir.RateLimitRule{rule},
				},
			},
		},
	}

	routeRateLimits := buildRouteRateLimits("listener-a", route)
	actions := routeRateLimits["listener-a"][0].Actions
	require.Len(t, actions, 3)
	require.Equal(t, "route-a", actions[0].GetGenericKey().DescriptorKey)
	require.Equal(t, "default/policy/rule/by-org", actions[1].GetGenericKey().DescriptorKey)
	require.Equal(t, "rule-0-match-0", actions[2].GetHeaderValueMatch().DescriptorKey)

	configs := BuildRateLimitServiceConfig([]*ir.HTTPListener{{
		CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-a"},
		Routes:              []*ir.HTTPRoute{route},
	}})
	require.Len(t, configs, 1)
	require.Equal(t, "route-a", configs[0].Descriptors[0].Key)
	require.Len(t, configs[0].Descriptors[0].Descriptors, 1)
	require.Equal(t, "default/policy/rule/by-org", configs[0].Descriptors[0].Descriptors[0].Key)
	require.Len(t, configs[0].Descriptors[0].Descriptors[0].Descriptors, 1)
	require.Equal(t, "rule-0-match-0", configs[0].Descriptors[0].Descriptors[0].Descriptors[0].Key)
}

func TestUnnamedNonSharedGlobalRateLimitRuleKeepsIndexDescriptorShape(t *testing.T) {
	rule := &ir.RateLimitRule{
		Name: "default/policy/rule/0",
		HeaderMatches: []*ir.StringMatch{
			{
				Name:  "x-org",
				Exact: ptr.To("acme"),
			},
		},
		Limit: ir.RateLimitValue{
			Requests: 10,
			Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
		},
	}
	route := &ir.HTTPRoute{
		Name: "route-a",
		Traffic: &ir.TrafficFeatures{
			RateLimit: &ir.RateLimit{
				Global: &ir.GlobalRateLimit{
					Rules: []*ir.RateLimitRule{rule},
				},
			},
		},
	}

	routeRateLimits := buildRouteRateLimits("listener-a", route)
	actions := routeRateLimits["listener-a"][0].Actions
	require.Len(t, actions, 3)
	require.Equal(t, "route-a", actions[0].GetGenericKey().DescriptorKey)
	require.Equal(t, "default/policy/rule/0", actions[1].GetGenericKey().DescriptorKey)
	require.Equal(t, "rule-0-match-0", actions[2].GetHeaderValueMatch().DescriptorKey)

	configs := BuildRateLimitServiceConfig([]*ir.HTTPListener{{
		CoreListenerDetails: ir.CoreListenerDetails{Name: "listener-a"},
		Routes:              []*ir.HTTPRoute{route},
	}})
	require.Len(t, configs, 1)
	require.Equal(t, "route-a", configs[0].Descriptors[0].Key)
	require.Len(t, configs[0].Descriptors[0].Descriptors, 1)
	require.Equal(t, "default/policy/rule/0", configs[0].Descriptors[0].Descriptors[0].Key)
	require.Len(t, configs[0].Descriptors[0].Descriptors[0].Descriptors, 1)
	require.Equal(t, "rule-0-match-0", configs[0].Descriptors[0].Descriptors[0].Descriptors[0].Key)
}

func TestNamedLocalRateLimitRuleUsesRuleName(t *testing.T) {
	local := &ir.LocalRateLimit{
		Rules: []*ir.RateLimitRule{
			{
				Name: "default/policy/rule/by-client",
				HeaderMatches: []*ir.StringMatch{
					{
						Name:  "x-client",
						Exact: ptr.To("mobile"),
					},
				},
				Limit: ir.RateLimitValue{
					Requests: 5,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
			},
			{
				Name: "default/policy/rule/1",
				HeaderMatches: []*ir.StringMatch{
					{
						Name:  "x-client",
						Exact: ptr.To("web"),
					},
				},
				Limit: ir.RateLimitValue{
					Requests: 5,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
			},
		},
	}

	rateLimits, descriptors := buildRouteLocalRateLimits(local)

	require.Len(t, rateLimits, 2)
	require.Equal(t, "default/policy/rule/by-client", rateLimits[0].Actions[0].GetGenericKey().DescriptorKey)
	require.Equal(t, "rule-0-match-0", rateLimits[0].Actions[1].GetHeaderValueMatch().DescriptorKey)
	require.Equal(t, "default/policy/rule/1", rateLimits[1].Actions[0].GetGenericKey().DescriptorKey)
	require.Equal(t, "rule-1-match-0", rateLimits[1].Actions[1].GetHeaderValueMatch().DescriptorKey)

	require.Len(t, descriptors, 2)
	require.Equal(t, "default/policy/rule/by-client", descriptors[0].Entries[0].Key)
	require.Equal(t, "rule-0-match-0", descriptors[0].Entries[1].Key)
	require.Equal(t, "default/policy/rule/1", descriptors[1].Entries[0].Key)
	require.Equal(t, "rule-1-match-0", descriptors[1].Entries[1].Key)
}
