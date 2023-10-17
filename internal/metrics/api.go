// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

// A Metric collects numerical observations.
type Metric interface {
	// Name returns the name value of a Metric.
	Name() string

	// Record makes an observation of the provided value for the given measure.
	Record(value float64)

	// RecordInt makes an observation of the provided value for the measure.
	RecordInt(value int64)

	// Increment records a value of 1 for the current measure.
	// For Counters, this is equivalent to adding 1 to the current value.
	// For Gauges, this is equivalent to setting the value to 1.
	// For Histograms, this is equivalent to making an observation of value 1.
	Increment()

	// Decrement records a value of -1 for the current measure.
	// For Counters, this is equivalent to subtracting -1 to the current value.
	// For Gauges, this is equivalent to setting the value to -1.
	// For Histograms, this is equivalent to making an observation of value -1.
	Decrement()

	// With creates a new Metric, with the LabelValues provided.
	// This allows creating a set of pre-dimensioned data for recording purposes.
	// This is primarily used for documentation and convenience.
	// Metrics created with this method do not need to be registered (they share the registration of their parent Metric).
	With(labelValues ...LabelValue) Metric
}

// Label holds a metric dimension which can be operated on using the interface
// methods.
type Label interface {
	// Value will set the provided value for the Label.
	Value(value string) LabelValue
}

// LabelValue holds an action to take on a metric dimension's value.
type LabelValue interface {
	// Key will get the key of the Label.
	Key() Label
	// Value will get the value of the Label.
	Value() string
}
