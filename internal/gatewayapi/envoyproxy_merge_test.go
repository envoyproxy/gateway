// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestMergeEnvoyProxyConfigs(t *testing.T) {
	testCases := []struct {
		name              string
		defaultSpec       *egv1a1.EnvoyProxySpec
		gatewayClassProxy *egv1a1.EnvoyProxy
		gatewayProxy      *egv1a1.EnvoyProxy
		expectedSpec      *egv1a1.EnvoyProxySpec
		expectError       bool
	}{
		{
			name:         "no configs provided",
			expectedSpec: nil,
		},
		{
			name: "only default spec",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
		},
		{
			name: "replace mode - gatewayclass overrides default",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(8)),
			},
		},
		{
			name: "replace mode - gateway overrides all",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(16)),
			},
		},
		{
			name: "gateway mergeType controls gateway-over-gatewayclass step",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   new(egv1a1.StrategicMerge),
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(16)),
			},
		},
		{
			name: "gateway nil mergeType - step1 Replace discards gatewayclass fields",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(16)),
			},
		},
		{
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(8)),
			},
		},
		{
			name: "gateway StrategicMerge - merges gateway+gatewayclass and defaults",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   new(egv1a1.StrategicMerge),
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(16)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
		},
		{
			name: "gateway StrategicMerge propagates to defaults merge - preserves default-only fields",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   new(egv1a1.StrategicMerge),
					Concurrency: new(int32(16)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(16)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
		},
		{
			name: "gatewayclass StrategicMerge - merges result with defaults",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: new(egv1a1.StrategicMerge),
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentAdmin: egv1a1.LogLevelDebug,
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
						egv1a1.LogComponentAdmin:   egv1a1.LogLevelDebug,
					},
				},
			},
		},
		{
			name: "defaultSpec mergeType has no effect on merge strategy",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   new(egv1a1.StrategicMerge),
				Concurrency: new(int32(4)),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: new(int32(8)),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: new(int32(8)),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := MergeEnvoyProxyConfigs(tc.defaultSpec, tc.gatewayClassProxy, tc.gatewayProxy)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tc.expectedSpec == nil {
				require.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			require.Equal(t, tc.expectedSpec.MergeType, result.Spec.MergeType)
			require.Equal(t, tc.expectedSpec.Concurrency, result.Spec.Concurrency)

			if len(tc.expectedSpec.Logging.Level) > 0 {
				require.Equal(t, tc.expectedSpec.Logging.Level, result.Spec.Logging.Level)
			}
		})
	}
}

// TestMergeEnvoyProxyConfigsTelemetryBackendRefNamespace checks that telemetry
// backendRefs inherited from the GatewayClass-level EnvoyProxy keep resolving in
// that EnvoyProxy's namespace after the Gateway-level EnvoyProxy is merged over it.
func TestMergeEnvoyProxyConfigsTelemetryBackendRefNamespace(t *testing.T) {
	const (
		classNS   = "envoy-gateway-system"
		gatewayNS = "app-ns"
	)

	ref := func(name string, ns *string) egv1a1.BackendRef {
		r := egv1a1.BackendRef{
			BackendObjectReference: gwapiv1.BackendObjectReference{
				Name: gwapiv1.ObjectName(name),
				Port: new(gwapiv1.PortNumber(4317)),
			},
		}
		if ns != nil {
			r.Namespace = new(gwapiv1.Namespace(*ns))
		}
		return r
	}
	nsOf := func(r egv1a1.BackendRef) string {
		if r.Namespace == nil {
			return ""
		}
		return string(*r.Namespace)
	}

	newClassProxy := func() *egv1a1.EnvoyProxy {
		return &egv1a1.EnvoyProxy{
			ObjectMeta: metav1.ObjectMeta{Namespace: classNS, Name: "class"},
			Spec: egv1a1.EnvoyProxySpec{
				Telemetry: &egv1a1.ProxyTelemetry{
					Tracing: &egv1a1.ProxyTracing{
						Provider: egv1a1.TracingProvider{
							Type: new(egv1a1.TracingProviderTypeDatadog),
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									ref("datadog-agent", nil),
									ref("otel-collector", new("monitoring")),
								},
							},
						},
					},
					AccessLog: &egv1a1.ProxyAccessLog{
						Settings: []egv1a1.ProxyAccessLogSetting{{
							Sinks: []egv1a1.ProxyAccessLogSink{
								{
									Type: egv1a1.ProxyAccessLogSinkTypeALS,
									ALS: &egv1a1.ALSEnvoyProxyAccessLog{
										Type:           egv1a1.ALSEnvoyProxyAccessLogTypeHTTP,
										BackendCluster: egv1a1.BackendCluster{BackendRefs: []egv1a1.BackendRef{ref("als", nil)}},
									},
								},
								{
									Type: egv1a1.ProxyAccessLogSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{
										BackendCluster: egv1a1.BackendCluster{BackendRefs: []egv1a1.BackendRef{ref("otel-logs", nil)}},
									},
								},
							},
						}},
					},
					Metrics: &egv1a1.ProxyMetrics{
						Sinks: []egv1a1.ProxyMetricSink{{
							Type: egv1a1.MetricSinkTypeOpenTelemetry,
							OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
								BackendCluster: egv1a1.BackendCluster{BackendRefs: []egv1a1.BackendRef{ref("otel-metrics", nil)}},
							},
						}},
					},
				},
			},
		}
	}
	newGatewayProxy := func(mergeType *egv1a1.MergeType, tracingRefs ...egv1a1.BackendRef) *egv1a1.EnvoyProxy {
		return &egv1a1.EnvoyProxy{
			ObjectMeta: metav1.ObjectMeta{Namespace: gatewayNS, Name: "gateway"},
			Spec: egv1a1.EnvoyProxySpec{
				MergeType: mergeType,
				Telemetry: &egv1a1.ProxyTelemetry{
					Tracing: &egv1a1.ProxyTracing{
						Provider: egv1a1.TracingProvider{
							ServiceName:    new("my-custom-service"),
							BackendCluster: egv1a1.BackendCluster{BackendRefs: tracingRefs},
						},
					},
				},
			},
		}
	}

	for _, mergeType := range []egv1a1.MergeType{egv1a1.StrategicMerge, egv1a1.JSONMerge} {
		t.Run(string(mergeType)+" pins inherited refs to the GatewayClass EnvoyProxy namespace", func(t *testing.T) {
			classProxy := newClassProxy()
			merged, err := MergeEnvoyProxyConfigs(nil, classProxy, newGatewayProxy(new(mergeType)))
			require.NoError(t, err)

			require.Equal(t, gatewayNS, merged.Namespace)
			telemetry := merged.Spec.Telemetry
			require.Equal(t, "my-custom-service", *telemetry.Tracing.Provider.ServiceName)
			require.Equal(t, egv1a1.TracingProviderTypeDatadog, *telemetry.Tracing.Provider.Type)

			tracingRefs := telemetry.Tracing.Provider.BackendRefs
			require.Len(t, tracingRefs, 2)
			require.Equal(t, classNS, nsOf(tracingRefs[0]), "inherited ref without a namespace")
			require.Equal(t, "monitoring", nsOf(tracingRefs[1]), "explicit namespace is kept")
			require.Equal(t, classNS, nsOf(telemetry.AccessLog.Settings[0].Sinks[0].ALS.BackendRefs[0]))
			require.Equal(t, classNS, nsOf(telemetry.AccessLog.Settings[0].Sinks[1].OpenTelemetry.BackendRefs[0]))
			require.Equal(t, classNS, nsOf(telemetry.Metrics.Sinks[0].OpenTelemetry.BackendRefs[0]))

			// The GatewayClass EnvoyProxy is shared across Gateways and must not be modified.
			require.Equal(t, newClassProxy(), classProxy)
		})
	}

	t.Run("refs set by the Gateway EnvoyProxy keep resolving in its own namespace", func(t *testing.T) {
		gatewayProxy := newGatewayProxy(new(egv1a1.StrategicMerge), ref("gateway-collector", nil))
		merged, err := MergeEnvoyProxyConfigs(nil, newClassProxy(), gatewayProxy)
		require.NoError(t, err)

		tracingRefs := merged.Spec.Telemetry.Tracing.Provider.BackendRefs
		require.Len(t, tracingRefs, 1)
		require.Equal(t, "gateway-collector", string(tracingRefs[0].Name))
		require.Empty(t, nsOf(tracingRefs[0]))
		// Sinks the Gateway EnvoyProxy did not touch are still inherited and pinned.
		require.Equal(t, classNS, nsOf(merged.Spec.Telemetry.Metrics.Sinks[0].OpenTelemetry.BackendRefs[0]))
	})

	t.Run("replace mode leaves the Gateway EnvoyProxy untouched", func(t *testing.T) {
		gatewayProxy := newGatewayProxy(nil, ref("gateway-collector", nil))
		merged, err := MergeEnvoyProxyConfigs(nil, newClassProxy(), gatewayProxy)
		require.NoError(t, err)

		require.Same(t, gatewayProxy, merged)
		require.Empty(t, nsOf(merged.Spec.Telemetry.Tracing.Provider.BackendRefs[0]))
	})

	t.Run("no Gateway EnvoyProxy leaves the GatewayClass EnvoyProxy untouched", func(t *testing.T) {
		classProxy := newClassProxy()
		merged, err := MergeEnvoyProxyConfigs(nil, classProxy, nil)
		require.NoError(t, err)

		require.Same(t, classProxy, merged)
		require.Empty(t, nsOf(merged.Spec.Telemetry.Tracing.Provider.BackendRefs[0]))
	})

	t.Run("default spec refs have no namespace to pin", func(t *testing.T) {
		defaultSpec := &egv1a1.EnvoyProxySpec{
			Telemetry: &egv1a1.ProxyTelemetry{
				Tracing: &egv1a1.ProxyTracing{
					Provider: egv1a1.TracingProvider{
						BackendCluster: egv1a1.BackendCluster{BackendRefs: []egv1a1.BackendRef{ref("default-collector", nil)}},
					},
				},
			},
		}
		classProxy := &egv1a1.EnvoyProxy{
			ObjectMeta: metav1.ObjectMeta{Namespace: classNS, Name: "class"},
			Spec: egv1a1.EnvoyProxySpec{
				MergeType: new(egv1a1.StrategicMerge),
				Telemetry: &egv1a1.ProxyTelemetry{
					Tracing: &egv1a1.ProxyTracing{
						Provider: egv1a1.TracingProvider{ServiceName: new("class-service")},
					},
				},
			},
		}
		merged, err := MergeEnvoyProxyConfigs(defaultSpec, classProxy, nil)
		require.NoError(t, err)

		tracingRefs := merged.Spec.Telemetry.Tracing.Provider.BackendRefs
		require.Len(t, tracingRefs, 1)
		require.Empty(t, nsOf(tracingRefs[0]))
	})
}
