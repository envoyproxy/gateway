// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rlv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/common/ratelimit/v3"
	localrlv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestLocalRateLimitPatchHCM(t *testing.T) {
	localRL := &localRateLimit{}

	tests := []struct {
		name       string
		mgr        *hcmv3.HttpConnectionManager
		irListener *ir.HTTPListener
		wantErr    bool
		errMsg     string
		validate   func(t *testing.T, mgr *hcmv3.HttpConnectionManager)
	}{
		{
			name: "nil hcm",
			mgr:  nil,
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{},
			},
			wantErr: true,
			errMsg:  "hcm is nil",
		},
		{
			name:       "nil ir listener",
			mgr:        &hcmv3.HttpConnectionManager{},
			irListener: nil,
			wantErr:    true,
			errMsg:     "ir listener is nil",
		},
		{
			name: "no local rate limit",
			mgr:  &hcmv3.HttpConnectionManager{},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				assert.Empty(t, mgr.HttpFilters)
			},
		},
		{
			name: "single route with local rate limit",
			mgr:  &hcmv3.HttpConnectionManager{},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						Traffic: &ir.TrafficFeatures{
							RateLimit: &ir.RateLimit{
								Local: &ir.LocalRateLimit{
									Default: ir.RateLimitValue{
										Requests: 100,
										Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				require.Len(t, mgr.HttpFilters, 1)
				assert.Equal(t, egv1a1.EnvoyFilterLocalRateLimit.String(), mgr.HttpFilters[0].Name)

				// Verify the config
				localRl := &localrlv3.LocalRateLimit{}
				err := mgr.HttpFilters[0].GetTypedConfig().UnmarshalTo(localRl)
				require.NoError(t, err)
				assert.Equal(t, localRateLimitFilterStatPrefix, localRl.StatPrefix)
				assert.Equal(t, uint32(10000), localRl.MaxDynamicDescriptors.Value)
			},
		},
		{
			name: "filter already exists",
			mgr: &hcmv3.HttpConnectionManager{
				HttpFilters: []*hcmv3.HttpFilter{
					{
						Name: egv1a1.EnvoyFilterLocalRateLimit.String(),
					},
				},
			},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						Traffic: &ir.TrafficFeatures{
							RateLimit: &ir.RateLimit{
								Local: &ir.LocalRateLimit{
									Default: ir.RateLimitValue{
										Requests: 100,
										Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				// Should not add duplicate filter
				require.Len(t, mgr.HttpFilters, 1)
			},
		},
		{
			name: "multiple routes with local rate limit",
			mgr:  &hcmv3.HttpConnectionManager{},
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "route-1",
						Traffic: &ir.TrafficFeatures{
							RateLimit: &ir.RateLimit{
								Local: &ir.LocalRateLimit{
									Default: ir.RateLimitValue{
										Requests: 100,
										Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
									},
								},
							},
						},
					},
					{
						Name: "route-2",
						Traffic: &ir.TrafficFeatures{
							RateLimit: &ir.RateLimit{
								Local: &ir.LocalRateLimit{
									Default: ir.RateLimitValue{
										Requests: 200,
										Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *hcmv3.HttpConnectionManager) {
				// Should only add one filter for all routes
				require.Len(t, mgr.HttpFilters, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := localRL.patchHCM(tt.mgr, tt.irListener)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.mgr)
				}
			}
		})
	}
}

func TestListenerContainsLocalRateLimit(t *testing.T) {
	tests := []struct {
		name       string
		irListener *ir.HTTPListener
		want       bool
	}{
		{
			name:       "nil listener",
			irListener: nil,
			want:       false,
		},
		{
			name: "no routes",
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{},
			},
			want: false,
		},
		{
			name: "route without rate limit",
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
					},
				},
			},
			want: false,
		},
		{
			name: "route with local rate limit",
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "test-route",
						Traffic: &ir.TrafficFeatures{
							RateLimit: &ir.RateLimit{
								Local: &ir.LocalRateLimit{
									Default: ir.RateLimitValue{
										Requests: 100,
										Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
									},
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "multiple routes, one with local rate limit",
			irListener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Name: "route-1",
					},
					{
						Name: "route-2",
						Traffic: &ir.TrafficFeatures{
							RateLimit: &ir.RateLimit{
								Local: &ir.LocalRateLimit{
									Default: ir.RateLimitValue{
										Requests: 100,
										Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
									},
								},
							},
						},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listenerContainsLocalRateLimit(tt.irListener)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRouteContainsLocalRateLimit(t *testing.T) {
	tests := []struct {
		name    string
		irRoute *ir.HTTPRoute
		want    bool
	}{
		{
			name:    "nil route",
			irRoute: nil,
			want:    false,
		},
		{
			name: "route without traffic",
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
			},
			want: false,
		},
		{
			name: "route without rate limit",
			irRoute: &ir.HTTPRoute{
				Name:    "test-route",
				Traffic: &ir.TrafficFeatures{},
			},
			want: false,
		},
		{
			name: "route without local rate limit",
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				Traffic: &ir.TrafficFeatures{
					RateLimit: &ir.RateLimit{},
				},
			},
			want: false,
		},
		{
			name: "route with local rate limit",
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				Traffic: &ir.TrafficFeatures{
					RateLimit: &ir.RateLimit{
						Local: &ir.LocalRateLimit{
							Default: ir.RateLimitValue{
								Requests: 100,
								Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
							},
						},
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routeContainsLocalRateLimit(tt.irRoute)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLocalRateLimitPatchResources(t *testing.T) {
	localRL := &localRateLimit{}

	// patchResources should always return nil
	rvt := &types.ResourceVersionTable{}
	routes := []*ir.HTTPRoute{
		{
			Name: "test-route",
			Traffic: &ir.TrafficFeatures{
				RateLimit: &ir.RateLimit{
					Local: &ir.LocalRateLimit{
						Default: ir.RateLimitValue{
							Requests: 100,
							Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
						},
					},
				},
			},
		},
	}

	err := localRL.patchResources(rvt, routes)
	assert.NoError(t, err)
}

func TestLocalRateLimitPatchRoute(t *testing.T) {
	localRL := &localRateLimit{}

	tests := []struct {
		name         string
		route        *routev3.Route
		irRoute      *ir.HTTPRoute
		httpListener *ir.HTTPListener
		wantErr      bool
		errMsg       string
		validate     func(t *testing.T, route *routev3.Route)
	}{
		{
			name: "no rate limit",
			route: &routev3.Route{
				Action: &routev3.Route_Route{
					Route: &routev3.RouteAction{},
				},
			},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
			},
			httpListener: &ir.HTTPListener{},
			wantErr:      false,
			validate: func(t *testing.T, route *routev3.Route) {
				assert.Nil(t, route.TypedPerFilterConfig)
			},
		},
		{
			name: "basic local rate limit",
			route: &routev3.Route{
				Name: "test-route",
				Action: &routev3.Route_Route{
					Route: &routev3.RouteAction{},
				},
			},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				Traffic: &ir.TrafficFeatures{
					RateLimit: &ir.RateLimit{
						Local: &ir.LocalRateLimit{
							Default: ir.RateLimitValue{
								Requests: 100,
								Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
							},
						},
					},
				},
			},
			httpListener: &ir.HTTPListener{},
			wantErr:      false,
			validate: func(t *testing.T, route *routev3.Route) {
				require.NotNil(t, route.TypedPerFilterConfig)
				filterName := egv1a1.EnvoyFilterLocalRateLimit.String()
				assert.Contains(t, route.TypedPerFilterConfig, filterName)

				// Verify the local rate limit config
				localRl := &localrlv3.LocalRateLimit{}
				err := route.TypedPerFilterConfig[filterName].UnmarshalTo(localRl)
				require.NoError(t, err)

				assert.Equal(t, localRateLimitFilterStatPrefix, localRl.StatPrefix)
				assert.Equal(t, uint32(100), localRl.TokenBucket.MaxTokens)
				assert.Equal(t, uint32(100), localRl.TokenBucket.TokensPerFill.Value)
				assert.False(t, localRl.AlwaysConsumeDefaultTokenBucket.Value)
				assert.Equal(t, rlv3.XRateLimitHeadersRFCVersion_DRAFT_VERSION_03, localRl.EnableXRatelimitHeaders)
			},
		},
		{
			name: "local rate limit with disabled headers",
			route: &routev3.Route{
				Name: "test-route",
				Action: &routev3.Route_Route{
					Route: &routev3.RouteAction{},
				},
			},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				Traffic: &ir.TrafficFeatures{
					RateLimit: &ir.RateLimit{
						Local: &ir.LocalRateLimit{
							Default: ir.RateLimitValue{
								Requests: 50,
								Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
							},
						},
					},
				},
			},
			httpListener: &ir.HTTPListener{
				Headers: &ir.HeaderSettings{
					DisableRateLimitHeaders: true,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, route *routev3.Route) {
				require.NotNil(t, route.TypedPerFilterConfig)
				filterName := egv1a1.EnvoyFilterLocalRateLimit.String()

				localRl := &localrlv3.LocalRateLimit{}
				err := route.TypedPerFilterConfig[filterName].UnmarshalTo(localRl)
				require.NoError(t, err)

				assert.Equal(t, rlv3.XRateLimitHeadersRFCVersion_OFF, localRl.EnableXRatelimitHeaders)
			},
		},
		{
			name: "route already has rate limits",
			route: &routev3.Route{
				Name: "test-route",
				Action: &routev3.Route_Route{
					Route: &routev3.RouteAction{
						RateLimits: []*routev3.RateLimit{
							{},
						},
					},
				},
			},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				Traffic: &ir.TrafficFeatures{
					RateLimit: &ir.RateLimit{
						Local: &ir.LocalRateLimit{
							Default: ir.RateLimitValue{
								Requests: 100,
								Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
							},
						},
					},
				},
			},
			httpListener: &ir.HTTPListener{},
			wantErr:      true,
			errMsg:       "route already contains rate limit config",
		},
		{
			name: "route already has local rate limit filter config",
			route: &routev3.Route{
				Name: "test-route",
				Action: &routev3.Route_Route{
					Route: &routev3.RouteAction{},
				},
				TypedPerFilterConfig: map[string]*anypb.Any{
					egv1a1.EnvoyFilterLocalRateLimit.String(): {},
				},
			},
			irRoute: &ir.HTTPRoute{
				Name: "test-route",
				Traffic: &ir.TrafficFeatures{
					RateLimit: &ir.RateLimit{
						Local: &ir.LocalRateLimit{
							Default: ir.RateLimitValue{
								Requests: 100,
								Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
							},
						},
					},
				},
			},
			httpListener: &ir.HTTPListener{},
			wantErr:      true,
			errMsg:       "route already contains local rate limit filter config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := localRL.patchRoute(tt.route, tt.irRoute, tt.httpListener)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.route)
				}
			}
		})
	}
}

func TestBuildRouteLocalRateLimits(t *testing.T) {
	tests := []struct {
		name                string
		local               *ir.LocalRateLimit
		validateRateLimits  func(t *testing.T, rateLimits []*routev3.RateLimit)
		validateDescriptors func(t *testing.T, descriptors []*rlv3.LocalRateLimitDescriptor)
	}{
		{
			name: "no rules",
			local: &ir.LocalRateLimit{
				Default: ir.RateLimitValue{
					Requests: 100,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
			},
			validateRateLimits: func(t *testing.T, rateLimits []*routev3.RateLimit) {
				assert.Empty(t, rateLimits)
			},
			validateDescriptors: func(t *testing.T, descriptors []*rlv3.LocalRateLimitDescriptor) {
				assert.Empty(t, descriptors)
			},
		},
		{
			name: "single rule with header match",
			local: &ir.LocalRateLimit{
				Default: ir.RateLimitValue{
					Requests: 100,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
				Rules: []*ir.RateLimitRule{
					{
						HeaderMatches: []*ir.StringMatch{
							{
								Name:  "x-user-id",
								Exact: ptr.To("user123"),
							},
						},
						Limit: ir.RateLimitValue{
							Requests: 50,
							Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
						},
					},
				},
			},
			validateRateLimits: func(t *testing.T, rateLimits []*routev3.RateLimit) {
				require.Len(t, rateLimits, 1)
				require.Len(t, rateLimits[0].Actions, 1)
			},
			validateDescriptors: func(t *testing.T, descriptors []*rlv3.LocalRateLimitDescriptor) {
				require.Len(t, descriptors, 1)
				assert.Equal(t, uint32(50), descriptors[0].TokenBucket.MaxTokens)
			},
		},
		{
			name: "rule with CIDR match",
			local: &ir.LocalRateLimit{
				Default: ir.RateLimitValue{
					Requests: 100,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
				Rules: []*ir.RateLimitRule{
					{
						CIDRMatch: &ir.CIDRMatch{
							CIDR:    "192.168.1.0/24",
							MaskLen: 24,
							IsIPv6:  false,
						},
						Limit: ir.RateLimitValue{
							Requests: 200,
							Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
						},
					},
				},
			},
			validateRateLimits: func(t *testing.T, rateLimits []*routev3.RateLimit) {
				require.Len(t, rateLimits, 1)
				require.Len(t, rateLimits[0].Actions, 1)
			},
			validateDescriptors: func(t *testing.T, descriptors []*rlv3.LocalRateLimitDescriptor) {
				require.Len(t, descriptors, 1)
				require.Len(t, descriptors[0].Entries, 1)
				assert.Equal(t, descriptorMaskedRemoteAddress, descriptors[0].Entries[0].Key)
			},
		},
		{
			name: "rule with distinct CIDR match",
			local: &ir.LocalRateLimit{
				Default: ir.RateLimitValue{
					Requests: 100,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
				Rules: []*ir.RateLimitRule{
					{
						CIDRMatch: &ir.CIDRMatch{
							CIDR:     "10.0.0.0/8",
							MaskLen:  8,
							IsIPv6:   false,
							Distinct: true,
						},
						Limit: ir.RateLimitValue{
							Requests: 150,
							Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
						},
					},
				},
			},
			validateRateLimits: func(t *testing.T, rateLimits []*routev3.RateLimit) {
				require.Len(t, rateLimits, 1)
				// Should have 2 actions: MaskedRemoteAddress + RemoteAddress
				require.Len(t, rateLimits[0].Actions, 2)
			},
			validateDescriptors: func(t *testing.T, descriptors []*rlv3.LocalRateLimitDescriptor) {
				require.Len(t, descriptors, 1)
				// Should have 2 entries for distinct CIDR
				require.Len(t, descriptors[0].Entries, 2)
				assert.Equal(t, descriptorMaskedRemoteAddress, descriptors[0].Entries[0].Key)
				assert.Equal(t, descriptorRemoteAddress, descriptors[0].Entries[1].Key)
			},
		},
		{
			name: "rule with method match",
			local: &ir.LocalRateLimit{
				Default: ir.RateLimitValue{
					Requests: 100,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
				Rules: []*ir.RateLimitRule{
					{
						MethodMatches: []*ir.StringMatch{
							{
								Exact: ptr.To("POST"),
							},
						},
						Limit: ir.RateLimitValue{
							Requests: 30,
							Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
						},
					},
				},
			},
			validateRateLimits: func(t *testing.T, rateLimits []*routev3.RateLimit) {
				require.Len(t, rateLimits, 1)
				require.Len(t, rateLimits[0].Actions, 1)
			},
			validateDescriptors: func(t *testing.T, descriptors []*rlv3.LocalRateLimitDescriptor) {
				require.Len(t, descriptors, 1)
			},
		},
		{
			name: "rule with path match",
			local: &ir.LocalRateLimit{
				Default: ir.RateLimitValue{
					Requests: 100,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
				Rules: []*ir.RateLimitRule{
					{
						PathMatch: &ir.StringMatch{
							Prefix: ptr.To("/api/"),
						},
						Limit: ir.RateLimitValue{
							Requests: 75,
							Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
						},
					},
				},
			},
			validateRateLimits: func(t *testing.T, rateLimits []*routev3.RateLimit) {
				require.Len(t, rateLimits, 1)
				require.Len(t, rateLimits[0].Actions, 1)
			},
			validateDescriptors: func(t *testing.T, descriptors []*rlv3.LocalRateLimitDescriptor) {
				require.Len(t, descriptors, 1)
			},
		},
		{
			name: "multiple rules",
			local: &ir.LocalRateLimit{
				Default: ir.RateLimitValue{
					Requests: 100,
					Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
				},
				Rules: []*ir.RateLimitRule{
					{
						HeaderMatches: []*ir.StringMatch{
							{
								Name:  "x-api-key",
								Exact: ptr.To("secret"),
							},
						},
						Limit: ir.RateLimitValue{
							Requests: 200,
							Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
						},
					},
					{
						PathMatch: &ir.StringMatch{
							Prefix: ptr.To("/admin/"),
						},
						Limit: ir.RateLimitValue{
							Requests: 10,
							Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
						},
					},
				},
			},
			validateRateLimits: func(t *testing.T, rateLimits []*routev3.RateLimit) {
				assert.Len(t, rateLimits, 2)
			},
			validateDescriptors: func(t *testing.T, descriptors []*rlv3.LocalRateLimitDescriptor) {
				assert.Len(t, descriptors, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rateLimits, descriptors := buildRouteLocalRateLimits(tt.local)

			if tt.validateRateLimits != nil {
				tt.validateRateLimits(t, rateLimits)
			}
			if tt.validateDescriptors != nil {
				tt.validateDescriptors(t, descriptors)
			}
		})
	}
}

func TestBuildHeaderMatchLocalRateLimitActions(t *testing.T) {
	tests := []struct {
		name            string
		ruleIdx         int
		headerMatches   []*ir.StringMatch
		validateActions func(t *testing.T, actions []*routev3.RateLimit_Action)
		validateEntries func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry)
	}{
		{
			name:          "no header matches",
			ruleIdx:       0,
			headerMatches: []*ir.StringMatch{},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				assert.Empty(t, actions)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				assert.Empty(t, entries)
			},
		},
		{
			name:    "distinct header match",
			ruleIdx: 0,
			headerMatches: []*ir.StringMatch{
				{
					Name:     "x-user-id",
					Distinct: true,
				},
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
				reqHeaders := actions[0].GetRequestHeaders()
				require.NotNil(t, reqHeaders)
				assert.Equal(t, "x-user-id", reqHeaders.HeaderName)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
				// For distinct matches, value should be empty
				assert.Empty(t, entries[0].Value)
			},
		},
		{
			name:    "exact header match",
			ruleIdx: 0,
			headerMatches: []*ir.StringMatch{
				{
					Name:  "x-api-key",
					Exact: ptr.To("secret"),
				},
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
				headerValueMatch := actions[0].GetHeaderValueMatch()
				require.NotNil(t, headerValueMatch)
				assert.True(t, headerValueMatch.ExpectMatch.Value)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
				// For exact matches, value should be set
				assert.NotEmpty(t, entries[0].Value)
			},
		},
		{
			name:    "inverted header match",
			ruleIdx: 0,
			headerMatches: []*ir.StringMatch{
				{
					Name:   "x-internal",
					Exact:  ptr.To("true"),
					Invert: ptr.To(true),
				},
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
				headerValueMatch := actions[0].GetHeaderValueMatch()
				require.NotNil(t, headerValueMatch)
				assert.False(t, headerValueMatch.ExpectMatch.Value)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rlActions []*routev3.RateLimit_Action
			var descriptorEntries []*rlv3.RateLimitDescriptor_Entry

			buildHeaderMatchLocalRateLimitActions(&rlActions, &descriptorEntries, tt.ruleIdx, tt.headerMatches)

			if tt.validateActions != nil {
				tt.validateActions(t, rlActions)
			}
			if tt.validateEntries != nil {
				tt.validateEntries(t, descriptorEntries)
			}
		})
	}
}

func TestBuildPathMatchLocalRateLimitAction(t *testing.T) {
	tests := []struct {
		name            string
		ruleIdx         int
		pathMatch       *ir.StringMatch
		validateActions func(t *testing.T, actions []*routev3.RateLimit_Action)
		validateEntries func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry)
	}{
		{
			name:      "nil path match",
			ruleIdx:   0,
			pathMatch: nil,
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				assert.Empty(t, actions)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				assert.Empty(t, entries)
			},
		},
		{
			name:    "prefix path match",
			ruleIdx: 0,
			pathMatch: &ir.StringMatch{
				Prefix: ptr.To("/api/"),
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
				headerValueMatch := actions[0].GetHeaderValueMatch()
				require.NotNil(t, headerValueMatch)
				assert.True(t, headerValueMatch.ExpectMatch.Value)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
				assert.NotEmpty(t, entries[0].Key)
				assert.NotEmpty(t, entries[0].Value)
			},
		},
		{
			name:    "exact path match",
			ruleIdx: 1,
			pathMatch: &ir.StringMatch{
				Exact: ptr.To("/admin/users"),
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rlActions []*routev3.RateLimit_Action
			var descriptorEntries []*rlv3.RateLimitDescriptor_Entry

			buildPathMatchLocalRateLimitAction(&rlActions, &descriptorEntries, tt.ruleIdx, tt.pathMatch)

			if tt.validateActions != nil {
				tt.validateActions(t, rlActions)
			}
			if tt.validateEntries != nil {
				tt.validateEntries(t, descriptorEntries)
			}
		})
	}
}

func TestBuildCIDRMatchLocalRateLimitActions(t *testing.T) {
	tests := []struct {
		name            string
		cidrMatch       *ir.CIDRMatch
		validateActions func(t *testing.T, actions []*routev3.RateLimit_Action)
		validateEntries func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry)
	}{
		{
			name:      "nil CIDR match",
			cidrMatch: nil,
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				assert.Empty(t, actions)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				assert.Empty(t, entries)
			},
		},
		{
			name: "IPv4 CIDR match",
			cidrMatch: &ir.CIDRMatch{
				CIDR:    "192.168.1.0/24",
				MaskLen: 24,
				IsIPv6:  false,
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
				mra := actions[0].GetMaskedRemoteAddress()
				require.NotNil(t, mra)
				assert.Equal(t, uint32(24), mra.V4PrefixMaskLen.Value)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
				assert.Equal(t, descriptorMaskedRemoteAddress, entries[0].Key)
				assert.Equal(t, "192.168.1.0/24", entries[0].Value)
			},
		},
		{
			name: "IPv6 CIDR match",
			cidrMatch: &ir.CIDRMatch{
				CIDR:    "2001:db8::/32",
				MaskLen: 32,
				IsIPv6:  true,
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
				mra := actions[0].GetMaskedRemoteAddress()
				require.NotNil(t, mra)
				assert.Equal(t, uint32(32), mra.V6PrefixMaskLen.Value)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
				assert.Equal(t, descriptorMaskedRemoteAddress, entries[0].Key)
			},
		},
		{
			name: "distinct CIDR match",
			cidrMatch: &ir.CIDRMatch{
				CIDR:     "10.0.0.0/8",
				MaskLen:  8,
				IsIPv6:   false,
				Distinct: true,
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				// Should have 2 actions: MaskedRemoteAddress + RemoteAddress
				require.Len(t, actions, 2)
				assert.NotNil(t, actions[0].GetMaskedRemoteAddress())
				assert.NotNil(t, actions[1].GetRemoteAddress())
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				// Should have 2 entries
				require.Len(t, entries, 2)
				assert.Equal(t, descriptorMaskedRemoteAddress, entries[0].Key)
				assert.Equal(t, descriptorRemoteAddress, entries[1].Key)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rlActions []*routev3.RateLimit_Action
			var descriptorEntries []*rlv3.RateLimitDescriptor_Entry

			buildCIDRMatchLocalRateLimitActions(&rlActions, &descriptorEntries, tt.cidrMatch)

			if tt.validateActions != nil {
				tt.validateActions(t, rlActions)
			}
			if tt.validateEntries != nil {
				tt.validateEntries(t, descriptorEntries)
			}
		})
	}
}

func TestBuildMethodMatchLocalRateLimitAction(t *testing.T) {
	tests := []struct {
		name            string
		ruleIdx         int
		methodMatch     *ir.StringMatch
		validateActions func(t *testing.T, actions []*routev3.RateLimit_Action)
		validateEntries func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry)
	}{
		{
			name:        "nil method match",
			ruleIdx:     0,
			methodMatch: nil,
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				assert.Empty(t, actions)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				assert.Empty(t, entries)
			},
		},
		{
			name:    "POST method match",
			ruleIdx: 0,
			methodMatch: &ir.StringMatch{
				Exact: ptr.To("POST"),
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
				headerValueMatch := actions[0].GetHeaderValueMatch()
				require.NotNil(t, headerValueMatch)
				assert.True(t, headerValueMatch.ExpectMatch.Value)
				require.Len(t, headerValueMatch.Headers, 1)
				assert.Equal(t, ":method", headerValueMatch.Headers[0].Name)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
				assert.NotEmpty(t, entries[0].Key)
				assert.NotEmpty(t, entries[0].Value)
			},
		},
		{
			name:    "GET method match",
			ruleIdx: 1,
			methodMatch: &ir.StringMatch{
				Exact: ptr.To("GET"),
			},
			validateActions: func(t *testing.T, actions []*routev3.RateLimit_Action) {
				require.Len(t, actions, 1)
			},
			validateEntries: func(t *testing.T, entries []*rlv3.RateLimitDescriptor_Entry) {
				require.Len(t, entries, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rlActions []*routev3.RateLimit_Action
			var descriptorEntries []*rlv3.RateLimitDescriptor_Entry

			buildMethodMatchLocalRateLimitAction(&rlActions, &descriptorEntries, tt.ruleIdx, tt.methodMatch)

			if tt.validateActions != nil {
				tt.validateActions(t, rlActions)
			}
			if tt.validateEntries != nil {
				tt.validateEntries(t, descriptorEntries)
			}
		})
	}
}

func TestLocalRateLimitIntegration(t *testing.T) {
	// Test the complete flow: patchHCM -> patchRoute
	localRL := &localRateLimit{}

	irListener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{
			{
				Name: "test-route",
				Traffic: &ir.TrafficFeatures{
					RateLimit: &ir.RateLimit{
						Local: &ir.LocalRateLimit{
							Default: ir.RateLimitValue{
								Requests: 100,
								Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
							},
							Rules: []*ir.RateLimitRule{
								{
									HeaderMatches: []*ir.StringMatch{
										{
											Name:  "x-api-key",
											Exact: ptr.To("premium"),
										},
									},
									Limit: ir.RateLimitValue{
										Requests: 500,
										Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Step 1: Patch HCM
	mgr := &hcmv3.HttpConnectionManager{}
	err := localRL.patchHCM(mgr, irListener)
	require.NoError(t, err)
	require.Len(t, mgr.HttpFilters, 1)
	assert.Equal(t, egv1a1.EnvoyFilterLocalRateLimit.String(), mgr.HttpFilters[0].Name)

	// Step 2: Patch Route
	route := &routev3.Route{
		Name: "test-route",
		Action: &routev3.Route_Route{
			Route: &routev3.RouteAction{},
		},
	}
	err = localRL.patchRoute(route, irListener.Routes[0], irListener)
	require.NoError(t, err)

	// Verify the route has the local rate limit config
	require.NotNil(t, route.TypedPerFilterConfig)
	filterName := egv1a1.EnvoyFilterLocalRateLimit.String()
	assert.Contains(t, route.TypedPerFilterConfig, filterName)

	// Verify the config details
	localRl := &localrlv3.LocalRateLimit{}
	err = route.TypedPerFilterConfig[filterName].UnmarshalTo(localRl)
	require.NoError(t, err)

	// Verify default token bucket
	assert.Equal(t, uint32(100), localRl.TokenBucket.MaxTokens)

	// Verify descriptors for rules
	assert.Len(t, localRl.Descriptors, 1)
	assert.Equal(t, uint32(500), localRl.Descriptors[0].TokenBucket.MaxTokens)

	// Verify rate limits are set in the local rate limit config
	assert.Len(t, localRl.RateLimits, 1)
}
