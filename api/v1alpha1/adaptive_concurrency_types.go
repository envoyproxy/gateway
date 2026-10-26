// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// AdaptiveConcurrency configures Envoy's adaptive concurrency HTTP filter, which
// dynamically adjusts the allowed request concurrency based on sampled latencies.
// This protects backends from overload by shedding load when latency increases.
//
// Each BackendTrafficPolicy gets its own filter instance so concurrency state is
// not shared across unrelated routes. Empty configuration (`{}`) enables the
// filter with Envoy's gradient-controller defaults. Only the knobs operators
// commonly need to tune are exposed; advanced minRTT sampling settings stay on
// Envoy defaults.
//
// +kubebuilder:validation:XValidation:rule="!has(self.concurrencyUpdateInterval) || duration(self.concurrencyUpdateInterval) > duration('0s')",message="concurrencyUpdateInterval must be greater than 0s"
// +kubebuilder:validation:XValidation:rule="!has(self.minConcurrencyLimit) || !has(self.maxConcurrencyLimit) || self.minConcurrencyLimit <= self.maxConcurrencyLimit",message="minConcurrencyLimit must be less than or equal to maxConcurrencyLimit"
type AdaptiveConcurrency struct {
	// SampleAggregatePercentile is the latency percentile used when summarizing
	// samples for the concurrency controller. Defaults to 50 (p50).
	//
	// +optional
	// +kubebuilder:default=50
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	SampleAggregatePercentile *float64 `json:"sampleAggregatePercentile,omitempty"`

	// MinConcurrencyLimit is the lower bound on the calculated concurrency limit.
	// Defaults to 3. Raise this when a backend needs a higher in-flight floor so
	// load-based autoscaling can still observe enough traffic.
	//
	// +optional
	// +kubebuilder:default=3
	// +kubebuilder:validation:Minimum=1
	MinConcurrencyLimit *uint32 `json:"minConcurrencyLimit,omitempty"`

	// MaxConcurrencyLimit is the upper bound on the calculated concurrency limit.
	// Defaults to 1000.
	//
	// +optional
	// +kubebuilder:default=1000
	// +kubebuilder:validation:Minimum=1
	MaxConcurrencyLimit *uint32 `json:"maxConcurrencyLimit,omitempty"`

	// ConcurrencyUpdateInterval is how often the concurrency limit is recalculated.
	// Envoy requires a value greater than 0s. Defaults to 100ms.
	//
	// +optional
	// +kubebuilder:default="100ms"
	ConcurrencyUpdateInterval *gwapiv1.Duration `json:"concurrencyUpdateInterval,omitempty"`

	// ConcurrencyLimitExceededStatus sets the HTTP status code returned to
	// downstream clients when the concurrency limit is exceeded.
	// Defaults to 503 (Service Unavailable).
	//
	// +optional
	// +kubebuilder:default=503
	// +kubebuilder:validation:Minimum=400
	// +kubebuilder:validation:Maximum=599
	ConcurrencyLimitExceededStatus *int32 `json:"concurrencyLimitExceededStatus,omitempty"`
}
