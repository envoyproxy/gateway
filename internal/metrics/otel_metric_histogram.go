// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
)

type Histogram struct {
	name  string
	attrs []attribute.KeyValue

	d                api.Float64Histogram
	preRecordOptions []api.RecordOption
}

func (f *Histogram) Record(value float64) {
	if f.preRecordOptions != nil {
		f.d.Record(context.Background(), value, f.preRecordOptions...)
	} else {
		f.d.Record(context.Background(), value)
	}
}

func (f *Histogram) With(labelValues ...LabelValue) *Histogram {
	attrs, set := mergeLabelValues(f.attrs, labelValues)
	m := &Histogram{
		name:             f.name,
		attrs:            attrs,
		d:                f.d,
		preRecordOptions: []api.RecordOption{api.WithAttributeSet(set)},
	}

	return m
}
