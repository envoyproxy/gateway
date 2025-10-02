// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/test"
)

func TestGetRenderedBootstrapConfig(t *testing.T) {
	sds := SdsConfigPath{
		Certificate: "/sds/xds-certificate.json",
		TrustedCA:   "/sds/xds-trusted-ca.json",
	}

	cases := []struct {
		name string
		opts *RenderBootstrapConfigOptions
	}{
		{
			name: "disable-prometheus",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Disable: true,
					},
				},
				SdsConfig:   sds,
				ServiceName: "envoy-gateway",
			},
		},
		{
			name: "enable-prometheus",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{},
				},
				SdsConfig:   sds,
				ServiceName: "envoy-gateway",
			},
		},
		{
			name: "enable-prometheus-gzip-compression",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Compression: &egv1a1.Compression{
							Type: egv1a1.GzipCompressorType,
						},
					},
				},
				SdsConfig:   sds,
				ServiceName: "envoy-gateway",
			},
		},
		{
			name: "enable-prometheus-brotli-compression",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Compression: &egv1a1.Compression{
							Type: egv1a1.BrotliCompressorType,
						},
					},
				},
				SdsConfig:   sds,
				ServiceName: "envoy-gateway",
			},
		},
		{
			name: "otel-metrics",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Disable: true,
					},
					Sinks: []egv1a1.ProxyMetricSink{
						{
							Type: egv1a1.MetricSinkTypeOpenTelemetry,
							OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
								Host: ptr.To("otel-collector.monitoring.svc"),
								Port: 4317,
							},
						},
					},
				},
				SdsConfig:   sds,
				ServiceName: "envoy-gateway",
			},
		},
		{
			name: "otel-metrics-backendref",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Disable: true,
					},
					Sinks: []egv1a1.ProxyMetricSink{
						{
							Type: egv1a1.MetricSinkTypeOpenTelemetry,
							OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
								Host: ptr.To("otel-collector.monitoring.svc"),
								Port: 4317,
								BackendCluster: egv1a1.BackendCluster{
									BackendRefs: []egv1a1.BackendRef{
										{
											BackendObjectReference: gwapiv1.BackendObjectReference{
												Name:      "otel-collector",
												Namespace: ptr.To(gwapiv1.Namespace("monitoring")),
												Port:      ptr.To(gwapiv1.PortNumber(4317)),
											},
										},
									},
								},
							},
						},
					},
				},
				SdsConfig:   sds,
				ServiceName: "envoy-gateway",
			},
		},
		{
			name: "custom-stats-matcher",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Matches: []egv1a1.StringMatch{
						{
							Type:  ptr.To(egv1a1.StringMatchExact),
							Value: "http.foo.bar.cluster.upstream_rq",
						},
						{
							Type:  ptr.To(egv1a1.StringMatchPrefix),
							Value: "http",
						},
						{
							Type:  ptr.To(egv1a1.StringMatchSuffix),
							Value: "upstream_rq",
						},
						{
							Type:  ptr.To(egv1a1.StringMatchRegularExpression),
							Value: "virtual.*",
						},
						{
							Type:  ptr.To(egv1a1.StringMatchPrefix),
							Value: "cluster",
						},
					},
				},
				SdsConfig:   sds,
				ServiceName: "envoy-gateway",
			},
		},
		{
			name: "custom-server-port",
			opts: &RenderBootstrapConfigOptions{
				XdsServerHost:   ptr.To("foo.bar"),
				XdsServerPort:   ptr.To(int32(12345)),
				AdminServerPort: ptr.To(int32(2222)),
				StatsServerPort: ptr.To(int32(3333)),
				SdsConfig:       sds,
				ServiceName:     "envoy-gateway",
			},
		},
		{
			name: "with-max-heap-size-bytes",
			opts: &RenderBootstrapConfigOptions{
				MaxHeapSizeBytes: 1073741824,
				SdsConfig:        sds,
				ServiceName:      "envoy-gateway",
			},
		},
		{
			name: "ipv6",
			opts: &RenderBootstrapConfigOptions{
				IPFamily:    ptr.To(egv1a1.IPv6),
				ServiceName: "envoy-gateway",
			},
		},
		{
			name: "topology-injector-disabled",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Disable: true,
					},
				},
				SdsConfig:                sds,
				TopologyInjectorDisabled: true,
			},
		},
		{
			name: "topology-injector-enabled",
			opts: &RenderBootstrapConfigOptions{
				ProxyMetrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Disable: true,
					},
				},
				SdsConfig:                sds,
				TopologyInjectorDisabled: false,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetRenderedBootstrapConfig(tc.opts)
			require.NoError(t, err)

			if test.OverrideTestData() {
				// nolint:gosec
				err = os.WriteFile(path.Join("testdata", "render", fmt.Sprintf("%s.yaml", tc.name)), []byte(got), 0o644)
				require.NoError(t, err)
				return
			}

			expected, err := readTestData(tc.name)
			require.NoError(t, err)
			assert.Equal(t, expected, got)
		})
	}
}

func readTestData(caseName string) (string, error) {
	filename := path.Join("testdata", "render", fmt.Sprintf("%s.yaml", caseName))

	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
