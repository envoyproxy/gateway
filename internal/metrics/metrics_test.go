// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	overrideTestData = flag.Bool("override-testdata", false, "if override the test output data.")
)

func TestCounter(t *testing.T) {

	name := "counter_metric"

	var writer io.ReadWriter = bytes.NewBuffer(nil)
	writer, err := exporterWriter(name, writer)
	require.NoError(t, err)

	counterProvider, err := newTestMetricsProvider("Counter", writer)
	require.NoError(t, err)

	// simulate a function that builds an indicator and changes its value
	metricsFunc := []func(){
		func() {
			metricName := "ir_updates_total"
			description := "Total Number of IR updates, by ir type"

			irCounter := NewCounter(
				metricName,
				description,
			)

			// increment on every xds ir update
			irCounter.With(NewLabel("ir-type").Value("xds")).Increment()
			// xds ir updates double
			irCounter.With(NewLabel("ir-type").Value("xds")).Add(2)
		},
		func() {
			name := "watchable_subscribed_total"
			description := "Total Number of IR updates, by ir type"

			subCounter := NewCounter(
				name,
				description,
			)

			// increment on every xds ir subscribed
			subCounter.With(NewLabel("ir-type").Value("xds")).Add(2)
			// xds ir updates double
			subCounter.With(NewLabel("ir-type").Value("xds")).Add(5)
		},
	}
	for _, f := range metricsFunc {
		f()
	}

	// Ensure that metrics data can be flushed by closing the provider
	err = counterProvider.Shutdown(context.Background())
	require.NoError(t, err)

	loadMetricsFile(t, name, writer)
}

func TestGauge(t *testing.T) {

	name := "gauge_metric"

	var writer io.ReadWriter = bytes.NewBuffer(nil)
	writer, err := exporterWriter(name, writer)
	require.NoError(t, err)

	gaugeProvider, err := newTestMetricsProvider("Gauge", writer)
	require.NoError(t, err)

	// simulate a function that builds an indicator and changes its value
	metricsFunc := []func(){
		func() {
			metricName := "current_irs_queue_num"
			description := "current number of ir in queue, by ir type"

			currentIRsNum := NewGauge(
				metricName,
				description,
			)

			// only the last recorded value (2) will be exported for this gauge
			currentIRsNum.With(NewLabel("ir-type").Value("xds")).Record(1)
			currentIRsNum.With(NewLabel("ir-type").Value("xds")).Record(3)
			currentIRsNum.With(NewLabel("ir-type").Value("xds")).Record(2)

			currentIRsNum.With(NewLabel("ir-type").Value("xds")).Record(1)
			currentIRsNum.With(NewLabel("ir-type").Value("xds")).Record(3)
			currentIRsNum.With(NewLabel("ir-type").Value("xds")).Record(2)
		},
	}
	for _, f := range metricsFunc {
		f()
	}

	// Ensure that metrics data can be flushed by closing the provider
	err = gaugeProvider.Shutdown(context.Background())
	require.NoError(t, err)

	loadMetricsFile(t, name, writer)
}

func TestHistogram(t *testing.T) {

	name := "histogram_metric"

	var writer io.ReadWriter = bytes.NewBuffer(nil)
	writer, err := exporterWriter(name, writer)
	require.NoError(t, err)

	histogramProvider, err := newTestMetricsProvider("Histogram", writer)
	require.NoError(t, err)

	// simulate a function that builds an indicator and changes its value
	metricsFunc := []func(){
		func() {
			metricName := "sent_bytes_total"
			description := "Histogram of sent bytes by method"

			sentBytes := NewHistogram(
				metricName,
				description,
				[]float64{10, 50, 100, 1000, 10000},
				WithUnit(Bytes),
			)

			// This will hit Bounds of 25,500 and 2500
			sentBytes.With(NewLabel("method").Value("/request/path/1")).Record(20)
			sentBytes.With(NewLabel("method").Value("/request/path/1")).Record(458)
			sentBytes.With(NewLabel("method").Value("/request/path/1")).Record(2000)
		},
	}
	for _, f := range metricsFunc {
		f()
	}

	// Ensure that metrics data can be flushed by closing the provider
	err = histogramProvider.Shutdown(context.Background())
	require.NoError(t, err)

	loadMetricsFile(t, name, writer)
}

// newTestMetricsProvider Create an OTEL Metrics Provider for testing use only
func newTestMetricsProvider(metricType string, writer io.Writer) (*metric.MeterProvider, error) {
	enc := json.NewEncoder(writer)
	enc.SetIndent("", "  ")

	stdExp, err := stdoutmetric.New(
		stdoutmetric.WithEncoder(
			&jsonEncoderWithoutTime{
				encoder: enc,
			},
		),
	)
	if err != nil {
		return nil, err
	}

	p := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(stdExp)),
		metric.WithResource(resource.NewSchemaless(
			semconv.ServiceName("test"),
			attribute.String("metric.type", metricType),
		)),
	)
	otel.SetMeterProvider(p)

	return p, nil
}

func loadMetricsFile(t *testing.T, name string, reader io.Reader) {
	if !*overrideTestData {
		fname := fmt.Sprintf("testdata/%s.json", name)

		// nolint:gosec
		f, err := os.ReadFile(fname)
		require.NoError(t, err)

		actual := reader.(*bytes.Buffer).String()
		// we set the json encoder indent, so we need to remove the "\r" from the read file
		expect := strings.ReplaceAll(string(f), "\r", "")

		require.Equal(t, expect, actual)
	}
}

func exporterWriter(name string, origin io.ReadWriter) (io.ReadWriter, error) {
	if *overrideTestData {
		fname := fmt.Sprintf("testdata/%s.json", name)

		// nolint:gosec
		f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
		return f, nil
	}

	return origin, nil
}

// jsonEncoderWithoutTime Is a JSON Encoder that zeroed out the ResourceMetrics time
// WARNING: This can only be used in the ResourceMetrics serialization when testing a metric!
type jsonEncoderWithoutTime struct {
	encoder *json.Encoder
}

func (enc *jsonEncoderWithoutTime) Encode(v any) error {
	data, ok := v.(*metricdata.ResourceMetrics)
	if !ok {
		return fmt.Errorf("object of type %T is not ResourceMetrics", data)
	}

	// Since to the presence of time information in metrics, it prevents us from performing comparisons.
	// In practice, when testing whether metrics are output as expected,
	// we are not overly concerned with the value of time,
	// but rather focus on the attributes and values of the metrics.
	// In serialization, we always set the Time/StartTime fields to zero value.
	for _, sm := range data.ScopeMetrics {
		for _, m := range sm.Metrics {
			val := reflect.ValueOf(m.Data).FieldByName("DataPoints")
			for i := 0; i < val.Len(); i++ {
				field := val.Index(i)
				if exist := func(reflect.Value) bool {
					var updated bool
					if !field.FieldByName("Time").IsZero() && field.IsValid() && field.CanSet() {
						field.FieldByName("Time").Set(reflect.ValueOf(time.Time{}))
						updated = true
					}
					if !field.FieldByName("StartTime").IsZero() && field.IsValid() && field.CanSet() {
						field.FieldByName("StartTime").Set(reflect.ValueOf(time.Time{}))
						updated = true
					}
					return updated
				}(field); !exist {
					return errors.New("failed to set the Time or StartTime field value")
				}

			}
		}
	}

	return enc.encoder.Encode(data)
}
