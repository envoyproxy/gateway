// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"context"

	api "go.opentelemetry.io/otel/metric"
)

var _ Metric = &otelHistogram{}

type otelHistogram struct {
	embed

	d                api.Float64Histogram
	preRecordOptions []api.RecordOption
}

func (f *otelHistogram) Record(value float64) {
	if f.preRecordOptions != nil {
		f.d.Record(context.Background(), value, f.preRecordOptions...)
	} else {
		f.d.Record(context.Background(), value)
	}
}

func (f *otelHistogram) With(labelValues ...LabelValue) Metric {
	attrs, set := mergeLabelValues(f.embed, labelValues)
	m := &otelHistogram{
		d:                f.d,
		preRecordOptions: []api.RecordOption{api.WithAttributeSet(set)},
	}
	m.embed = embed{
		name:  f.name,
		attrs: attrs,
		m:     m,
	}
	return m
}
