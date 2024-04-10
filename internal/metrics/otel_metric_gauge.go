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

type Gauge struct {
	name  string
	attrs []attribute.KeyValue

	g       api.Float64ObservableGauge
	mutex   *sync.RWMutex
	stores  map[attribute.Set]*GaugeValues
	current *GaugeValues
}

type GaugeValues struct {
	val float64
	opt []api.ObserveOption
}

func (f *Gauge) Record(value float64) {
	f.mutex.Lock()
	if f.current == nil {
		f.current = &GaugeValues{}
		f.stores[attribute.NewSet()] = f.current
	}
	f.current.val = value
	f.mutex.Unlock()
}

func (f *Gauge) With(labelValues ...LabelValue) *Gauge {
	attrs, set := mergeLabelValues(f.attrs, labelValues)
	m := &Gauge{
		g:      f.g,
		mutex:  f.mutex,
		stores: f.stores,
		name:   f.name,
		attrs:  attrs,
	}
	if _, f := m.stores[set]; !f {
		m.stores[set] = &GaugeValues{
			opt: []api.ObserveOption{api.WithAttributeSet(set)},
		}
	}
	m.current = m.stores[set]

	return m
}
