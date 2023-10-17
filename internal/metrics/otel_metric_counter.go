// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"

	api "go.opentelemetry.io/otel/metric"
)

var _ Metric = &otelCounter{}

type otelCounter struct {
	embed

	c                api.Float64Counter
	preRecordOptions []api.AddOption
}

func (f *otelCounter) Record(value float64) {
	if f.preRecordOptions != nil {
		f.c.Add(context.Background(), value, f.preRecordOptions...)
	} else {
		f.c.Add(context.Background(), value)
	}
}

func (f *otelCounter) With(labelValues ...LabelValue) Metric {
	attrs, set := mergeLabelValues(f.embed, labelValues)
	m := &otelCounter{
		c:                f.c,
		preRecordOptions: []api.AddOption{api.WithAttributeSet(set)},
	}
	m.embed = embed{
		name:  f.name,
		attrs: attrs,
		m:     m,
	}
	return m
}
