// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	"testing"

	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/test"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

// TestResolvedMetricSinksConversion verifies that resolved metric sinks from IR
// are correctly converted to bootstrap format, including TLS configuration.
func TestResolvedMetricSinksConversion(t *testing.T) {
	sni := "otel-collector.example.com"

	testCases := []struct {
		name     string
		irSinks  []ir.ResolvedMetricSink
		expected []bootstrap.MetricSink
	}{
		{
			name:     "no sinks",
			irSinks:  nil,
			expected: []bootstrap.MetricSink{},
		},
		{
			name: "skip sink with no endpoints",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name:     "metrics_otel_0",
						Settings: []*ir.DestinationSetting{{}},
					},
				},
			},
			expected: []bootstrap.MetricSink{},
		},
		{
			name: "sink without TLS",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name: "metrics_otel_0",
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "otel-collector.example.com", Port: 4317},
								},
							},
						},
					},
				},
			},
			expected: []bootstrap.MetricSink{
				{
					Address: "otel-collector.example.com",
					Port:    4317,
				},
			},
		},
		{
			name: "sink with TLS and SNI",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name: "metrics_otel_0",
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "otel-collector.example.com", Port: 443},
								},
								TLS: &ir.TLSUpstreamConfig{
									SNI: &sni,
								},
							},
						},
					},
					Authority: "otel-collector.example.com",
				},
			},
			expected: []bootstrap.MetricSink{
				{
					Address:   "otel-collector.example.com",
					Port:      443,
					Authority: "otel-collector.example.com",
					TLS: &bootstrap.MetricSinkTLS{
						SNI: sni,
					},
				},
			},
		},
		{
			name: "sink with TLS and custom CA",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name: "metrics_otel_0",
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "otel-collector.example.com", Port: 443},
								},
								TLS: &ir.TLSUpstreamConfig{
									SNI: &sni,
									CACertificate: &ir.TLSCACertificate{
										Name:        "custom-ca",
										Certificate: test.TestCACertificate,
									},
								},
							},
						},
					},
					Authority: "otel-collector.example.com",
				},
			},
			expected: []bootstrap.MetricSink{
				{
					Address:   "otel-collector.example.com",
					Port:      443,
					Authority: "otel-collector.example.com",
					TLS: &bootstrap.MetricSinkTLS{
						SNI:           sni,
						CACertificate: test.TestCACertificate,
					},
				},
			},
		},
		{
			name: "sink with headers and deltas",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name: "metrics_otel_0",
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "otel-collector.example.com", Port: 4317},
								},
							},
						},
					},
					Headers: []gwapiv1.HTTPHeader{
						{Name: "Authorization", Value: "Bearer token"},
					},
					ReportCountersAsDeltas:   true,
					ReportHistogramsAsDeltas: true,
				},
			},
			expected: []bootstrap.MetricSink{
				{
					Address:                  "otel-collector.example.com",
					Port:                     4317,
					ReportCountersAsDeltas:   true,
					ReportHistogramsAsDeltas: true,
					Headers: []gwapiv1.HTTPHeader{
						{Name: "Authorization", Value: "Bearer token"},
					},
				},
			},
		},
		{
			name: "sink with resources",
			irSinks: []ir.ResolvedMetricSink{
				{
					Destination: ir.RouteDestination{
						Name: "metrics_otel_0",
						Settings: []*ir.DestinationSetting{
							{
								Endpoints: []*ir.DestinationEndpoint{
									{Host: "otel-collector.example.com", Port: 4317},
								},
							},
						},
					},
					Resources: map[string]string{
						"service.name":           "test-service",
						"deployment.environment": "test",
					},
				},
			},
			expected: []bootstrap.MetricSink{
				{
					Address: "otel-collector.example.com",
					Port:    4317,
					Resources: map[string]string{
						"service.name":           "test-service",
						"deployment.environment": "test",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ConvertResolvedMetricSinks(tc.irSinks)
			require.Equal(t, tc.expected, actual)
		})
	}
}
