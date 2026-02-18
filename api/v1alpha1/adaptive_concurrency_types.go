// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// AdaptiveConcurrency defines the configuration for Envoy's adaptive concurrency
// filter, which dynamically adjusts the allowed request concurrency limit based
// on sampled latencies.
//
// +k8s:deepcopy-gen=true
type AdaptiveConcurrency struct {
	// GradientController configures the gradient-based concurrency controller.
	// This is the core algorithm that adjusts the concurrency limit based on
	// latency measurements.
	//
	// +optional
	GradientController *GradientController `json:"gradientController,omitempty"`

	// ConcurrencyLimitExceededStatus sets the HTTP status code returned to
	// downstream clients when the concurrency limit is exceeded.
	// Defaults to 503 (Service Unavailable).
	//
	// +optional
	// +kubebuilder:validation:Minimum=400
	// +kubebuilder:validation:Maximum=599
	ConcurrencyLimitExceededStatus *int32 `json:"concurrencyLimitExceededStatus,omitempty"`
}

// GradientController configures the gradient-based concurrency controller algorithm.
// The controller periodically samples request latencies and adjusts the concurrency
// limit using a gradient calculated from the current vs minimum RTT.
//
// +k8s:deepcopy-gen=true
type GradientController struct {
	// SampleAggregatePercentile specifies the percentile to use when
	// summarizing aggregated latency samples. Defaults to 50 (p50).
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	SampleAggregatePercentile *float64 `json:"sampleAggregatePercentile,omitempty"`

	// ConcurrencyLimitParams controls the periodic concurrency limit
	// recalculation.
	//
	// +optional
	ConcurrencyLimitParams *ConcurrencyLimitParams `json:"concurrencyLimitParams,omitempty"`

	// MinRTTCalcParams controls the periodic minRTT recalculation.
	//
	// +optional
	MinRTTCalcParams *MinRTTCalcParams `json:"minRTTCalcParams,omitempty"`
}

// ConcurrencyLimitParams controls parameters for the periodic recalculation
// of the concurrency limit from sampled request latencies.
//
// +k8s:deepcopy-gen=true
type ConcurrencyLimitParams struct {
	// MaxConcurrencyLimit sets the upper bound on the concurrency limit.
	// Defaults to 1000.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	MaxConcurrencyLimit *uint32 `json:"maxConcurrencyLimit,omitempty"`

	// ConcurrencyUpdateInterval is the period of time between
	// recalculations of the concurrency limit.
	// Defaults to 100ms.
	//
	// +optional
	ConcurrencyUpdateInterval *gwapiv1.Duration `json:"concurrencyUpdateInterval,omitempty"`
}

// MinRTTCalcParams controls parameters for the periodic minRTT recalculation.
//
// +k8s:deepcopy-gen=true
type MinRTTCalcParams struct {
	// Interval is the time interval between recalculating the minimum RTT.
	// Defaults to 30s. Setting this to 0s disables dynamic minRTT sampling
	// (use FixedMinRTT in that case).
	//
	// +optional
	Interval *gwapiv1.Duration `json:"interval,omitempty"`

	// FixedMinRTT sets a fixed minimum RTT value instead of dynamically
	// sampling it. This is required when Interval is set to 0s.
	//
	// +optional
	FixedMinRTT *gwapiv1.Duration `json:"fixedMinRTT,omitempty"`

	// RequestCount is the number of requests to aggregate during the minRTT
	// recalculation window. Defaults to 50.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	RequestCount *uint32 `json:"requestCount,omitempty"`

	// Jitter adds a randomized delay to the start of the minRTT calculation
	// window, expressed as a percentage of the interval. Defaults to 15%.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Jitter *float64 `json:"jitter,omitempty"`

	// MinConcurrency sets the concurrency limit used while measuring the
	// minRTT. Defaults to 3.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	MinConcurrency *uint32 `json:"minConcurrency,omitempty"`

	// Buffer is the amount added to the measured minRTT to add stability
	// to the concurrency limit, expressed as a percentage of the measured
	// value. Defaults to 25%.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Buffer *float64 `json:"buffer,omitempty"`
}
