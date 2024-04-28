// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/stretchr/testify/assert"
)

func TestProcessMetrics(t *testing.T) {
	cases := []struct {
		name  string
		proxy *egcfgv1a1.EnvoyProxy

		expected *ir.Metrics
	}{
		{
			name: "nil proxy config",
		},
		{
			name: "virtual host stats enabled",
			proxy: &egcfgv1a1.EnvoyProxy{
				Spec: egcfgv1a1.EnvoyProxySpec{
					Telemetry: &egcfgv1a1.ProxyTelemetry{
						Metrics: &egcfgv1a1.ProxyMetrics{
							EnableVirtualHostStats: true,
						},
					},
				},
			},
			expected: &ir.Metrics{
				EnableVirtualHostStats: true,
			},
		},
		{
			name: "peer endpoint stats enabled",
			proxy: &egcfgv1a1.EnvoyProxy{
				Spec: egcfgv1a1.EnvoyProxySpec{
					Telemetry: &egcfgv1a1.ProxyTelemetry{
						Metrics: &egcfgv1a1.ProxyMetrics{
							EnablePerEndpointStats: true,
						},
					},
				},
			},
			expected: &ir.Metrics{
				EnablePerEndpointStats: true,
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := processMetrics(c.proxy)
			assert.Equal(t, c.expected, got)
		})
	}
}
