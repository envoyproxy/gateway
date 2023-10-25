// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import "go.opentelemetry.io/otel/attribute"

// A Label provides a named dimension for a Metric.
type Label struct {
	key attribute.Key
}

// NewLabel will attempt to create a new Label.
func NewLabel(key string) Label {
	return Label{attribute.Key(key)}
}

// Value creates a new LabelValue for the Label.
func (l Label) Value(value string) LabelValue {
	return LabelValue{l.key.String(value)}
}

// A LabelValue represents a Label with a specific value. It is used to record
// values for a Metric.
type LabelValue struct {
	keyValue attribute.KeyValue
}

func (l LabelValue) Key() Label {
	return Label{l.keyValue.Key}
}

func (l LabelValue) Value() string {
	return l.keyValue.Value.AsString()
}

func mergeLabelValues(attrs []attribute.KeyValue, labelValues []LabelValue) ([]attribute.KeyValue, attribute.Set) {
	mergedAttrs := make([]attribute.KeyValue, 0, len(attrs)+len(labelValues))
	mergedAttrs = append(mergedAttrs, attrs...)
	for _, v := range labelValues {
		kv := v
		mergedAttrs = append(mergedAttrs, kv.keyValue)
	}

	return mergedAttrs, attribute.NewSet(mergedAttrs...)
}
