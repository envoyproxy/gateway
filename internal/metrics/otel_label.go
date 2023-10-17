// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import "go.opentelemetry.io/otel/attribute"

// NewLabel will attempt to create a new Label.
func NewLabel(key string) Label {
	return otelLabel{attribute.Key(key)}
}

// A otelLabel provides a named dimension for a Metric.
type otelLabel struct {
	key attribute.Key
}

// Value creates a new LabelValue for the Label.
func (l otelLabel) Value(value string) LabelValue {
	return otelLabelValue{l.key.String(value)}
}

// A LabelValue represents a Label with a specific value. It is used to record
// values for a Metric.
type otelLabelValue struct {
	keyValue attribute.KeyValue
}

func (l otelLabelValue) Key() Label {
	return otelLabel{l.keyValue.Key}
}

func (l otelLabelValue) Value() string {
	return l.keyValue.Value.AsString()
}
