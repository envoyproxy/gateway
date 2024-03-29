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

func TestHistogram(t *testing.T) {
	writer := os.Stdout
	if true {
		// nolint:gosec
		f, err := os.OpenFile("testdata/histogram_metric.yaml", os.O_CREATE|os.O_RDWR, 0644)
		require.NoError(t, err)
		writer = f
	}
	p, err := initHistogramProvider(writer)
	require.NoError(t, err)

	// simulate a function that builds an indicator and changes its value
	metricsFunc := []func(){
		methodHistogram,
	}
	for _, f := range metricsFunc {
		f()
	}

	// Ensure that metrics data can be flushed by closing the provider
	err = p.Shutdown(context.TODO())
	require.NoError(t, err)
}

func methodHistogram() {
	sentBytes := metrics.NewHistogram(
		"sent_bytes_total",
		"Histogram of sent bytes by method",
		[]float64{10, 50, 100, 1000, 10000},
		metrics.WithUnit(metrics.Bytes),
	)

	// This will hit Bounds of 25,500 and 2500
	sentBytes.With(metrics.NewLabel("method").Value("/request/path/1")).Record(20)
	sentBytes.With(metrics.NewLabel("method").Value("/request/path/1")).Record(458)
	sentBytes.With(metrics.NewLabel("method").Value("/request/path/1")).Record(2000)
}

func initHistogramProvider(writer io.Writer) (*metric.MeterProvider, error) {
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
			attribute.String("metric-type", "Histogram"),
		)),
	)
	otel.SetMeterProvider(p)

	return p, nil
}
