// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

// Options encode changes to the options passed to a Metric at creation time.
type MetricOption func(*MetricOptions)

type MetricOptions struct {
	EnabledCondition func() bool
	Unit             Unit
	Name             string
	Description      string
}

// WithUnit provides configuration options for a new Metric, providing unit of measure
// information for a new Metric.
func WithUnit(unit Unit) MetricOption {
	return func(opts *MetricOptions) {
		opts.Unit = unit
	}
}

// WithEnabled allows a metric to be condition enabled if the provided function returns true.
// If disabled, metric operations will do nothing.
func WithEnabled(enabled func() bool) MetricOption {
	return func(opts *MetricOptions) {
		opts.EnabledCondition = enabled
	}
}

func metricOptions(name, description string, opts ...MetricOption) (MetricOptions, Metric) {
	o := MetricOptions{Unit: None, Name: name, Description: description}
	for _, opt := range opts {
		opt(&o)
	}
	if o.EnabledCondition != nil && !o.EnabledCondition() {
		return o, &disabled{name: name}
	}
	return o, nil
}
