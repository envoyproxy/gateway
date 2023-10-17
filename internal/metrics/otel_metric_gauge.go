// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"sync"

	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
)

var _ Metric = &otelGauge{}

type otelGauge struct {
	embed

	g       api.Float64ObservableGauge
	mutex   *sync.RWMutex
	stores  map[attribute.Set]*otelGaugeValues
	current *otelGaugeValues
}

type otelGaugeValues struct {
	val float64
	opt []api.ObserveOption
}

func (f *otelGauge) Record(value float64) {
	f.mutex.Lock()
	if f.current == nil {
		f.current = &otelGaugeValues{}
		f.stores[attribute.NewSet()] = f.current
	}
	f.current.val = value
	f.mutex.Unlock()
}

func (f *otelGauge) With(labelValues ...LabelValue) Metric {
	attrs, set := mergeLabelValues(f.embed, labelValues)
	m := &otelGauge{
		g:      f.g,
		mutex:  f.mutex,
		stores: f.stores,
	}
	if _, f := m.stores[set]; !f {
		m.stores[set] = &otelGaugeValues{
			opt: []api.ObserveOption{api.WithAttributeSet(set)},
		}
	}
	m.current = m.stores[set]
	m.embed = embed{
		name:  f.name,
		attrs: attrs,
		m:     m,
	}
	return m
}
