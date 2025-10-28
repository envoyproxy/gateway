// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import "go.opentelemetry.io/otel/attribute"

var (
	// statusLabel defines a label to indicate the status of current metric,
	// e.g. is a SUCCESS or FAILURE status.
	statusLabel = NewLabel("status")

	// reasonLabel defines a label to indicate the reason of failure status metric,
	// it's an optional label.
	reasonLabel = NewLabel("reason")
)

const (
	StatusSuccess = "success"
	StatusFailure = "failure"

	ReasonError    = "error"
	ReasonConflict = "conflict"
)

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

func mergeLabelValues(attrs []attribute.KeyValue, labelValues []LabelValue) []attribute.KeyValue {
	mergedAttrs := make([]attribute.KeyValue, len(attrs)+len(labelValues))
	copy(mergedAttrs, attrs)
	for i, v := range labelValues {
		mergedAttrs[len(attrs)+i] = v.keyValue
	}
	return mergedAttrs
}
