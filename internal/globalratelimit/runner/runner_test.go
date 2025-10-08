// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	rlsconfv3 "github.com/envoyproxy/go-control-plane/ratelimit/config/ratelimit/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

func Test_subscribeAndTranslate(t *testing.T) {
	t.Parallel()

	testxds := func(gwName string) *ir.Xds {
		return &ir.Xds{
			HTTP: []*ir.HTTPListener{
				{
					CoreListenerDetails: ir.CoreListenerDetails{
						Name: fmt.Sprintf("default/%s/listener-0", gwName),
					},
					Routes: []*ir.HTTPRoute{
						{
							Name: "route-0",
							Traffic: &ir.TrafficFeatures{
								RateLimit: &ir.RateLimit{
									Global: &ir.GlobalRateLimit{
										Rules: []*ir.RateLimitRule{
											{
												HeaderMatches: []*ir.StringMatch{
													{
														Name:     "x-user-id",
														Distinct: true,
													},
												},
												Limit: ir.RateLimitValue{
													Requests: 100,
													Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
												},
											},
											{
												HeaderMatches: []*ir.StringMatch{
													{
														Name:     "x-another-user-id",
														Distinct: true,
													},
												},
												Limit: ir.RateLimitValue{
													Requests: 10,
													Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
												},
											},
										},
									},
								},
							},
						},
						{
							Name: "route-1",
							Traffic: &ir.TrafficFeatures{
								RateLimit: &ir.RateLimit{
									Global: &ir.GlobalRateLimit{
										Rules: []*ir.RateLimitRule{
											{
												HeaderMatches: []*ir.StringMatch{
													{
														Name:     "x-user-id",
														Distinct: true,
													},
												},
												Limit: ir.RateLimitValue{
													Requests: 100,
													Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitMinute),
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
		}
	}

	testRateLimitConfig := func(gwName string) *rlsconfv3.RateLimitConfig {
		return &rlsconfv3.RateLimitConfig{
			Name:   fmt.Sprintf("default/%s/listener-0", gwName),
			Domain: fmt.Sprintf("default/%s/listener-0", gwName),
			Descriptors: []*rlsconfv3.RateLimitDescriptor{
				{
					Key:   "route-0",
					Value: "route-0",
					Descriptors: []*rlsconfv3.RateLimitDescriptor{
						{
							Key: "rule-0-match-0",
							RateLimit: &rlsconfv3.RateLimitPolicy{
								Unit:            rlsconfv3.RateLimitUnit_MINUTE,
								RequestsPerUnit: 100,
							},
						},
						{
							Key: "rule-1-match-0",
							RateLimit: &rlsconfv3.RateLimitPolicy{
								Unit:            rlsconfv3.RateLimitUnit_SECOND,
								RequestsPerUnit: 10,
							},
						},
					},
				},
				{
					Key:   "route-1",
					Value: "route-1",
					Descriptors: []*rlsconfv3.RateLimitDescriptor{
						{
							Key: "rule-0-match-0",
							RateLimit: &rlsconfv3.RateLimitPolicy{
								Unit:            rlsconfv3.RateLimitUnit_MINUTE,
								RequestsPerUnit: 100,
							},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		name string
		// xdsIRs contains a list of xds updates that the runner will receive.
		xdsIRs               []message.Update[string, *ir.Xds]
		wantRateLimitConfigs map[string]cachetypes.Resource
	}{
		{
			name: "one xds is added",
			xdsIRs: []message.Update[string, *ir.Xds]{
				{
					Key:   "gw0",
					Value: testxds("gw0"),
				},
			},
			wantRateLimitConfigs: map[string]cachetypes.Resource{
				"default/gw0/listener-0": testRateLimitConfig("gw0"),
			},
		},
		{
			name: "two xds are added",
			xdsIRs: []message.Update[string, *ir.Xds]{
				{
					Key:   "gw0",
					Value: testxds("gw0"),
				},
				{
					Key:   "gw1",
					Value: testxds("gw1"),
				},
			},
			wantRateLimitConfigs: map[string]cachetypes.Resource{
				"default/gw0/listener-0": testRateLimitConfig("gw0"),
				"default/gw1/listener-0": testRateLimitConfig("gw1"),
			},
		},
		{
			name: "one xds is deleted",
			xdsIRs: []message.Update[string, *ir.Xds]{
				{
					Key:   "gw0",
					Value: testxds("gw0"),
				},
				{
					Key:   "gw1",
					Value: testxds("gw1"),
				},
				{
					Key:    "gw0",
					Delete: true,
				},
			},
			wantRateLimitConfigs: map[string]cachetypes.Resource{
				"default/gw1/listener-0": testRateLimitConfig("gw1"),
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			xdsIR := new(message.XdsIR)
			defer xdsIR.Close()
			cfg, err := config.New(os.Stdout)
			require.NoError(t, err)

			r := New(&Config{
				Server: *cfg,
				XdsIR:  xdsIR,
				cache:  cachev3.NewSnapshotCache(false, cachev3.IDHash{}, nil),
			})

			c := xdsIR.Subscribe(ctx)
			go r.translateFromSubscription(ctx, c)

			for _, xds := range tt.xdsIRs {
				if xds.Delete {
					xdsIR.Delete(xds.Key)
					continue
				}
				xdsIR.Store(xds.Key, xds.Value)
			}

			diff := ""
			if !assert.Eventually(t, func() bool {
				rs, err := r.cache.GetSnapshot(ratelimit.InfraName)
				if err != nil {
					t.Logf("failed to get snapshot: %v", err)
					return false
				}

				diff = cmp.Diff(tt.wantRateLimitConfigs, rs.GetResources(resourcev3.RateLimitConfigType), cmpopts.IgnoreUnexported(rlsconfv3.RateLimitConfig{}, rlsconfv3.RateLimitDescriptor{}, rlsconfv3.RateLimitPolicy{}))
				return diff == ""
			}, time.Second*10, time.Second) {
				t.Fatalf("snapshot mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
