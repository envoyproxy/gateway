// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestProcessTracing(t *testing.T) {
	cases := []struct {
		gw    gwapiv1.Gateway
		proxy *egcfgv1a1.EnvoyProxy

		expected *ir.Tracing
	}{
		{},
		{
			gw: gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-gw",
					Namespace: "fake-ns",
				},
			},
			proxy: &egcfgv1a1.EnvoyProxy{
				Spec: egcfgv1a1.EnvoyProxySpec{
					Telemetry: &egcfgv1a1.ProxyTelemetry{
						Tracing: &egcfgv1a1.ProxyTracing{},
					},
				},
			},
			expected: &ir.Tracing{
				ServiceName:  "fake-gw.fake-ns",
				ProxyTracing: egcfgv1a1.ProxyTracing{},
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run("", func(t *testing.T) {
			got := processTracing(&c.gw, c.proxy)
			assert.Equal(t, c.expected, got)
		})
	}
}

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
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := processMetrics(c.proxy)
			assert.Equal(t, c.expected, got)
		})
	}
}
