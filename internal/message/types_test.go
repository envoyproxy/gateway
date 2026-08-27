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
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

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

func TestParentContextPreservesFallbackLifecycle(t *testing.T) {
	tracer := sdktrace.NewTracerProvider().Tracer("test")
	producerCtx, producerSpan := tracer.Start(t.Context(), "producer")
	defer producerSpan.End()
	fallback, cancel := context.WithCancel(t.Context())
	defer cancel()

	for _, value := range []interface {
		ParentContext(context.Context) context.Context
	}{
		&message.XdsIRWithContext{Context: producerCtx},
		&message.InfraIRWithContext{Context: producerCtx},
	} {
		consumerCtx := value.ParentContext(fallback)
		require.Equal(t, trace.SpanContextFromContext(producerCtx), trace.SpanContextFromContext(consumerCtx))
		cancel()
		require.ErrorIs(t, consumerCtx.Err(), context.Canceled)
	}
}
