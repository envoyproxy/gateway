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

type Counter struct {
	name             string
	attrs            []attribute.KeyValue
	c                api.Float64Counter
	preRecordOptions []api.AddOption
}

func (f *Counter) Add(value float64) {
	if f.preRecordOptions != nil {
		f.c.Add(context.Background(), value, f.preRecordOptions...)
	} else {
		f.c.Add(context.Background(), value)
	}
}

func (f *Counter) Increment() {
	f.Add(1)
}

func (f *Counter) Decrement() {
	f.Add(-1)
}

func (f *Counter) With(labelValues ...LabelValue) *Counter {
	mergedAttrs := mergeLabelValues(f.attrs, labelValues)
	m := &Counter{
		c:                f.c,
		preRecordOptions: []api.AddOption{api.WithAttributes(mergedAttrs...)},
		name:             f.name,
		attrs:            mergedAttrs,
	}

	return m
}

func (f *Counter) WithStatus(status string, labelValues ...LabelValue) *Counter {
	labelValues = append(labelValues, statusLabel.Value(status))
	return f.With(labelValues...)
}

func (f *Counter) WithSuccess(labelValues ...LabelValue) *Counter {
	if len(labelValues) > 0 {
		labelValues = append(labelValues, statusLabel.Value(StatusSuccess))
	} else {
		labelValues = []LabelValue{statusLabel.Value(StatusSuccess)}
	}
	return f.With(labelValues...)
}

func (f *Counter) WithFailure(reason string, labelValues ...LabelValue) *Counter {
	if len(labelValues) > 0 {
		labelValues = append(labelValues, statusLabel.Value(StatusFailure), reasonLabel.Value(reason))
	} else {
		labelValues = []LabelValue{statusLabel.Value(StatusFailure), reasonLabel.Value(reason)}
	}
	return f.With(labelValues...)
}
