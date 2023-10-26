// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
)

// NewCounter creates a new Counter Metric (the values will be cumulative).
// That means that data collected by the new Metric will be summed before export.
func NewCounter(name, description string, opts ...MetricOption) *Counter {
	stores.register(Metadata{
		Name:        name,
		Type:        CounterType,
		Description: description,
	})
	o := metricOptions(name, description, opts...)

	return newCounter(o)
}

// NewGauge creates a new Gauge Metric. That means that data collected by the new
// Metric will export only the last recorded value.
func NewGauge(name, description string, opts ...MetricOption) *Gauge {
	stores.register(Metadata{
		Name:        name,
		Type:        GaugeType,
		Description: description,
	})
	o := metricOptions(name, description, opts...)

	return newGauge(o)
}

// NewHistogram creates a new Metric with an aggregation type of Histogram.
// This means that the data collected by the Metric will be collected and exported as a histogram, with the specified bounds.
func NewHistogram(name, description string, bounds []float64, opts ...MetricOption) *Histogram {
	stores.register(Metadata{
		Name:        name,
		Type:        HistogramType,
		Description: description,
		Bounds:      bounds,
	})
	o := metricOptions(name, description, opts...)

	return newHistogram(o)
}

func newCounter(o MetricOptions) *Counter {
	c, err := meter().Float64Counter(o.Name,
		api.WithDescription(o.Description),
		api.WithUnit(string(o.Unit)))
	if err != nil {
		metricsLogger.Error(err, "failed to create otel Counter")
	}
	m := &Counter{c: c, name: o.Name}

	return m
}

func newGauge(o MetricOptions) *Gauge {
	r := &Gauge{mutex: &sync.RWMutex{}, name: o.Name}
	r.stores = map[attribute.Set]*GaugeValues{}
	g, err := meter().Float64ObservableGauge(o.Name,
		api.WithFloat64Callback(func(ctx context.Context, observer api.Float64Observer) error {
			r.mutex.Lock()
			defer r.mutex.Unlock()
			for _, gv := range r.stores {
				observer.Observe(gv.val, gv.opt...)
			}
			return nil
		}),
		api.WithDescription(o.Description),
		api.WithUnit(string(o.Unit)))
	if err != nil {
		metricsLogger.Error(err, "failed to create otel Gauge")
	}
	r.g = g

	return r
}

func newHistogram(o MetricOptions) *Histogram {
	d, err := meter().Float64Histogram(o.Name,
		api.WithDescription(o.Description),
		api.WithUnit(string(o.Unit)))
	if err != nil {
		metricsLogger.Error(err, "failed to create otel Histogram")
	}
	m := &Histogram{d: d, name: o.Name}

	return m
}
