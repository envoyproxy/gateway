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

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

func TestGetRenderedBootstrapConfig(t *testing.T) {
	cases := []struct {
		name         string
		proxyMetrics *egcfgv1a1.ProxyMetrics
	}{
		{
			name: "default",
		},
		{
			name: "enable-prometheus",
			proxyMetrics: &egcfgv1a1.ProxyMetrics{
				Prometheus: &egcfgv1a1.PrometheusProvider{},
			},
		},
		{
			name: "otel-metrics",
			proxyMetrics: &egcfgv1a1.ProxyMetrics{
				Sinks: []egcfgv1a1.ProxyMetricSink{
					{
						Type: egcfgv1a1.MetricSinkTypeOpenTelemetry,
						OpenTelemetry: &egcfgv1a1.OpenTelemetrySink{
							Host: "otel-collector.monitoring.svc",
							Port: 4317,
						},
					},
				},
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
