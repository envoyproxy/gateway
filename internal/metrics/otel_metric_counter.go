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
	attrs, set := mergeLabelValues(f.attrs, labelValues)
	m := &Counter{
		c:                f.c,
		preRecordOptions: []api.AddOption{api.WithAttributeSet(set)},
		name:             f.name,
		attrs:            attrs,
	}

	return m
}
