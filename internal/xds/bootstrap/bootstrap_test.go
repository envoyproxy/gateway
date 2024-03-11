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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestGetRenderedBootstrapConfig(t *testing.T) {
	cases := []struct {
		name         string
		proxyMetrics *egv1a1.ProxyMetrics
	}{
		{
			name: "disable-prometheus",
			proxyMetrics: &egv1a1.ProxyMetrics{
				Prometheus: &egv1a1.ProxyPrometheusProvider{
					Disable: true,
				},
			},
		},
		{
			name: "enable-prometheus",
			proxyMetrics: &egv1a1.ProxyMetrics{
				Prometheus: &egv1a1.ProxyPrometheusProvider{},
			},
		},
		{
			name: "otel-metrics",
			proxyMetrics: &egv1a1.ProxyMetrics{
				Prometheus: &egv1a1.ProxyPrometheusProvider{
					Disable: true,
				},
				Sinks: []egv1a1.ProxyMetricSink{
					{
						Type: egv1a1.MetricSinkTypeOpenTelemetry,
						OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{
							Host: "otel-collector.monitoring.svc",
							Port: 4317,
						},
					},
				},
			},
		},
		{
			name: "custom-stats-matcher",
			proxyMetrics: &egv1a1.ProxyMetrics{
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
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetRenderedBootstrapConfig(tc.proxyMetrics)
			require.NoError(t, err)

			if *overrideTestData {
				// nolint:gosec
				err = os.WriteFile(path.Join("testdata", "render", fmt.Sprintf("%s.yaml", tc.name)), []byte(got), 0644)
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
