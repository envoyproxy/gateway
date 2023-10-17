// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"go.opentelemetry.io/otel/attribute"
)

// embed metric implementation.
type embed struct {
	name  string
	attrs []attribute.KeyValue
	m     Metric
}

func (f embed) Name() string {
	return f.name
}

func (f embed) Increment() {
	f.m.Record(1)
}

func (f embed) Decrement() {
	f.m.Record(-1)
}

func (f embed) RecordInt(value int64) {
	f.m.Record(float64(value))
}

// disabled metric implementation.
type disabled struct {
	name string
}

// Decrement implements Metric
func (dm *disabled) Decrement() {}

// Increment implements Metric
func (dm *disabled) Increment() {}

// Name implements Metric
func (dm *disabled) Name() string {
	return dm.name
}

// Record implements Metric
func (dm *disabled) Record(value float64) {}

// RecordInt implements Metric
func (dm *disabled) RecordInt(value int64) {}

// With implements Metric
func (dm *disabled) With(labelValues ...LabelValue) Metric {
	return dm
}

var _ Metric = &disabled{}

func mergeLabelValues(bm embed, labelValues []LabelValue) ([]attribute.KeyValue, attribute.Set) {
	attrs := make([]attribute.KeyValue, 0, len(bm.attrs)+len(labelValues))
	attrs = append(attrs, bm.attrs...)
	for _, v := range labelValues {
		kv := v.(otelLabelValue)
		attrs = append(attrs, kv.keyValue)
	}

	set := attribute.NewSet(attrs...)
	return attrs, set
}
