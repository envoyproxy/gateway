// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

// XdsIRWithContext structs with differing context values should be Equal
func TestXdsWithContextEqual(t *testing.T) {
	xdsIR := &ir.Xds{
		HTTP: []*ir.HTTPListener{
			{
				CoreListenerDetails: ir.CoreListenerDetails{
					Name: fmt.Sprintf("default/%s/listener-0", "gwName"),
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
				},
			},
		},
	}

	c1 := context.Background()
	c2 := context.TODO()

	x1 := &message.XdsIRWithContext{
		XdsIR:   xdsIR,
		Context: c1,
	}
	x2 := &message.XdsIRWithContext{
		XdsIR:   xdsIR,
		Context: c2,
	}

	assert.True(t, x1.Equal(x2))
	assert.True(t, x2.Equal(x1))
}

// TestXdsNACKsRoundTrip verifies that NACK facts can be stored, observed via a
// subscription, loaded, and deleted on the XdsNACKs message map.
func TestXdsNACKsRoundTrip(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nacks := new(message.XdsNACKs)
	sub := nacks.Subscribe(ctx)

	nack := &message.XdsNACK{
		Rejections: map[message.XdsNACKKey]message.XdsNACKError{
			{NodeID: "pod-1", TypeURL: "type.googleapis.com/envoy.config.listener.v3.Listener"}: {
				Code:    13,
				Message: "invalid access log format",
			},
		},
	}
	nacks.Store("default/eg", nack)

	// The Store should be observable on the subscription.
	gotUpdate := false
	for snapshot := range sub {
		for k, v := range snapshot.State {
			assert.Equal(t, "default/eg", k)
			assert.Equal(t, nack, v)
			gotUpdate = true
		}
		if gotUpdate {
			break
		}
	}
	assert.True(t, gotUpdate)

	got, ok := nacks.Load("default/eg")
	assert.True(t, ok)
	assert.Equal(t, nack, got)

	nacks.Delete("default/eg")
	_, ok = nacks.Load("default/eg")
	assert.False(t, ok)
}
