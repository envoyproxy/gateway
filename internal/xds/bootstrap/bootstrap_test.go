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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestGetRenderedBootstrapConfig(t *testing.T) {
	cases := []struct {
		name         string
		proxyMetrics *egv1a1.ProxyMetrics
	}{
		{
			name: "default",
		},
		{
			name: "enable-prometheus",
			proxyMetrics: &egv1a1.ProxyMetrics{
				Prometheus: &egv1a1.PrometheusProvider{},
			},
		},
		{
			name: "otel-metrics",
			proxyMetrics: &egv1a1.ProxyMetrics{
				Sinks: []egv1a1.MetricSink{
					{
						Type: egv1a1.MetricSinkTypeOpenTelemetry,
						OpenTelemetry: &egv1a1.OpenTelemetrySink{
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
				Matches: []egv1a1.Match{
					{
						Type:  egv1a1.Prefix,
						Value: "http",
					},
					{
						Type:  egv1a1.Suffix,
						Value: "upstream_rq",
					},
					{
						Type:  egv1a1.RegularExpression,
						Value: "virtual.*",
					},
					{
						Type:  egv1a1.Prefix,
						Value: "cluster",
					},
				},
				Prometheus: &egv1a1.PrometheusProvider{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetRenderedBootstrapConfig(tc.proxyMetrics)
			assert.NoError(t, err)
			expected, err := readTestData(tc.name)
			assert.NoError(t, err)
			assert.Equal(t, expected, got)
		})
	}
}

func readTestData(caseName string) (string, error) {
	filename := path.Join("testdata", fmt.Sprintf("%s.yaml", caseName))

	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
