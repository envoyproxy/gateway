// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics_test

import (
	"context"
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

func TestCounter(t *testing.T) {

	writer := os.Stdout
	if true {
		// nolint:gosec
		f, err := os.OpenFile("testdata/counter_metric.yaml", os.O_CREATE|os.O_RDWR, 0644)
		require.NoError(t, err)
		writer = f
	}

	p, err := initCounterProvider(writer)
	require.NoError(t, err)

	// simulate a function that builds an indicator and changes its value
	metricsFunc := []func(){
		xdsIRCounter,
		watchableSubCounter,
	}
	for _, f := range metricsFunc {
		f()
	}

	// Ensure that metrics data can be flushed by closing the provider
	err = p.Shutdown(context.TODO())
	require.NoError(t, err)
}

func xdsIRCounter() {
	irCounter := metrics.NewCounter(
		"ir_updates_total",
		"Number of IR updates, by ir type",
	)
	// increment on every xds ir update
	irCounter.With(metrics.NewLabel("ir-type").Value("xds")).Increment()
	// xds ir updates double
	irCounter.With(metrics.NewLabel("ir-type").Value("xds")).Add(2)
}

func watchableSubCounter() {
	subCounter := metrics.NewCounter(
		"watchable_subscribed_total",
		"Total number of subscribed watchable.",
	)
	// increment on every xds ir subscribed
	subCounter.With(metrics.NewLabel("ir-type").Value("xds")).Add(2)
	// xds ir updates double
	subCounter.With(metrics.NewLabel("ir-type").Value("xds")).Add(5)
}

func initCounterProvider(writer io.Writer) (*metric.MeterProvider, error) {
	stdExp, err := stdoutmetric.New(
		stdoutmetric.WithPrettyPrint(),
		stdoutmetric.WithWriter(writer),
	)
	if err != nil {
		return nil, err
	}

	p := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(stdExp)),
		metric.WithResource(resource.NewSchemaless(
			semconv.ServiceName("test"),
			attribute.String("metric-type", "Counter"),
		)),
	)
	otel.SetMeterProvider(p)

	return p, nil
}
