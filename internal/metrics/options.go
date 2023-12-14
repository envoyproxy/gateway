// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

// Options encode changes to the options passed to a Metric at creation time.
type MetricOption func(*MetricOptions)

type MetricOptions struct {
	Unit        Unit
	Name        string
	Description string
}

// WithUnit provides configuration options for a new Metric, providing unit of measure
// information for a new Metric.
func WithUnit(unit Unit) MetricOption {
	return func(opts *MetricOptions) {
		opts.Unit = unit
	}
}

func metricOptions(name, description string, opts ...MetricOption) MetricOptions {
	o := MetricOptions{Unit: None, Name: name, Description: description}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
