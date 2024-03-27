// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/envoyproxy/gateway/internal/metrics"
)

func TestGauge(t *testing.T) {
	writer := os.Stdout
	if true {
		// nolint:gosec
		f, err := os.OpenFile("testdata/gauge_metric.yaml", os.O_CREATE|os.O_RDWR, 0644)
		require.NoError(t, err)
		writer = f
	}
	p, err := initGaugeProvider(writer)
	require.NoError(t, err)

	// simulate a function that builds an indicator and changes its value
	metricsFunc := []func(){
		xdsIRGauge,
	}
	for _, f := range metricsFunc {
		f()
	}

	// Ensure that metrics data can be flushed by closing the provider
	err = p.Shutdown(context.TODO())
	require.NoError(t, err)
}

func xdsIRGauge() {
	currentIRsNum := metrics.NewGauge(
		"current_irs_queue_num",
		"current number of ir in queue, by ir type",
	)

	// only the last recorded value (2) will be exported for this gauge
	currentIRsNum.With(metrics.NewLabel("ir-type").Value("xds")).Record(1)
	currentIRsNum.With(metrics.NewLabel("ir-type").Value("xds")).Record(3)
	currentIRsNum.With(metrics.NewLabel("ir-type").Value("xds")).Record(2)

	currentIRsNum.With(metrics.NewLabel("ir-type").Value("xds")).Record(1)
	currentIRsNum.With(metrics.NewLabel("ir-type").Value("xds")).Record(3)
	currentIRsNum.With(metrics.NewLabel("ir-type").Value("xds")).Record(2)
}

func initGaugeProvider(writer io.Writer) (*metric.MeterProvider, error) {
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")

	stdExp, err := stdoutmetric.New(
		stdoutmetric.WithEncoder(enc),
	)
	if err != nil {
		return nil, err
	}

	p := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(stdExp)),
		metric.WithResource(resource.NewSchemaless(
			semconv.ServiceName("test"),
			attribute.String("metric-type", "Gauge"),
		)),
	)
	otel.SetMeterProvider(p)

	return p, nil
}
