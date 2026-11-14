// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestWarnEnvoyGateway(t *testing.T) {
	eg := egv1a1.DefaultEnvoyGateway()

	testCases := []struct {
		name     string
		eg       *egv1a1.EnvoyGateway
		expected []string
	}{
		{
			name:     "nil EnvoyGateway",
			eg:       nil,
			expected: nil,
		},
		{
			name:     "nil ExtensionAPIs",
			eg:       eg,
			expected: nil,
		},
		{
			name: "disableLua is set",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionAPIs: &egv1a1.ExtensionAPISettings{
						DisableLua: new(true),
					},
				},
			},
			expected: []string{"disableLua is deprecated, use enableLua instead"},
		},
		{
			name: "enableLua is set",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionAPIs: &egv1a1.ExtensionAPISettings{
						EnableLua: true,
					},
				},
			},
			expected: nil,
		},
		{
			name: "extension service host is set",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Service: &egv1a1.ExtensionService{Host: "extension.default.svc", Port: 8080},
					},
					ExtensionManagers: []egv1a1.ExtensionManager{
						{Service: &egv1a1.ExtensionService{BackendEndpoint: egv1a1.BackendEndpoint{
							FQDN: &egv1a1.FQDNEndpoint{Hostname: "extension.default.svc", Port: 8080},
						}}},
						{Service: &egv1a1.ExtensionService{Host: "extension.default.svc", Port: 8080}},
					},
				},
			},
			expected: []string{
				"extensionManager.service.host and extensionManager.service.port are deprecated and will be removed in v1.11, use extensionManager.service.fqdn, extensionManager.service.ip or extensionManager.service.unix instead",
				"extensionManagers[1].service.host and extensionManagers[1].service.port are deprecated and will be removed in v1.11, use extensionManagers[1].service.fqdn, extensionManagers[1].service.ip or extensionManagers[1].service.unix instead",
			},
		},
		{
			name: "default envoyProxy uses deprecated fields",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					EnvoyProxy: &egv1a1.EnvoyProxySpec{
						Telemetry: &egv1a1.ProxyTelemetry{
							Tracing: &egv1a1.ProxyTracing{Provider: egv1a1.TracingProvider{
								Host: new("otel-collector.monitoring.svc"),
								Port: 4317,
							}},
						},
					},
				},
			},
			expected: []string{
				"envoyProxy.telemetry.tracing.provider.host is deprecated and will be removed in v1.11, use envoyProxy.telemetry.tracing.provider.backendRefs instead",
			},
		},
		{
			name: "default envoyProxy keeps explicit legacy ports with backendRefs",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					EnvoyProxy: &egv1a1.EnvoyProxySpec{
						Telemetry: &egv1a1.ProxyTelemetry{
							Tracing: &egv1a1.ProxyTracing{Provider: egv1a1.TracingProvider{Port: 4317}},
							Metrics: &egv1a1.ProxyMetrics{Sinks: []egv1a1.ProxyMetricSink{
								{OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{Port: 4317}},
								{OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{}},
							}},
						},
					},
				},
			},
			expected: []string{
				"envoyProxy.telemetry.metrics.sinks[0].openTelemetry.port is deprecated and will be removed in v1.11, use envoyProxy.telemetry.metrics.sinks[0].openTelemetry.backendRefs instead",
				"envoyProxy.telemetry.tracing.provider.port is deprecated and will be removed in v1.11, use envoyProxy.telemetry.tracing.provider.backendRefs instead",
			},
		},
		{
			name: "extension service port is set with a new endpoint",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Service: &egv1a1.ExtensionService{
							BackendEndpoint: egv1a1.BackendEndpoint{
								FQDN: &egv1a1.FQDNEndpoint{Hostname: "extension.default.svc", Port: 8080},
							},
							Port: 8080,
						},
					},
				},
			},
			expected: []string{
				"extensionManager.service.port is deprecated and will be removed in v1.11, use the port of extensionManager.service.fqdn or extensionManager.service.ip instead",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			warnings := WarnEnvoyGateway(tc.eg)
			assert.Equal(t, tc.expected, warnings)
		})
	}
}
