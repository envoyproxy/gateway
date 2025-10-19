// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdaptiveConcurrency defines the adaptive concurrency configuration for backend traffic.
// This feature dynamically adjusts the number of concurrent requests to optimize
// performance and prevent overload based on downstream response times.
//
// +optional
type AdaptiveConcurrency struct {
	// Enabled enables or disables adaptive concurrency. Defaults to false.
	//
	// +optional
	// +kubebuilder:default=false
	Enabled *bool `json:"enabled,omitempty"`

	// GradientController defines the gradient controller configuration for adaptive concurrency.
	//
	// +optional
	GradientController *AdaptiveConcurrencyGradientController `json:"gradientController,omitempty"`

	// MinRTTCalculation defines the minimum RTT calculation parameters.
	//
	// +optional
	MinRTTCalculation *AdaptiveConcurrencyMinRTTCalculation `json:"minRTTCalculation,omitempty"`

	// ConcurrencyLimitParams defines the concurrency limit parameters.
	//
	// +optional
	ConcurrencyLimitParams *AdaptiveConcurrencyLimitParams `json:"concurrencyLimitParams,omitempty"`
}

// AdaptiveConcurrencyGradientController defines the gradient controller configuration.
//
// +optional
type AdaptiveConcurrencyGradientController struct {
	// SampleAggregatePercentile defines the percentile of request latencies to sample.
	// This should be a value between 0 and 100.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=90
	SampleAggregatePercentile *float64 `json:"sampleAggregatePercentile,omitempty"`

	// ConcurrencyLimitParams defines the concurrency limit parameters.
	//
	// +optional
	ConcurrencyLimitParams *AdaptiveConcurrencyLimitParams `json:"concurrencyLimitParams,omitempty"`

	// MinRTTCalculation defines the minimum RTT calculation parameters.
	//
	// +optional
	MinRTTCalculation *AdaptiveConcurrencyMinRTTCalculation `json:"minRTTCalculation,omitempty"`
}

// AdaptiveConcurrencyLimitParams defines the concurrency limit parameters.
//
// +optional
type AdaptiveConcurrencyLimitParams struct {
	// ConcurrencyUpdateInterval defines the interval at which the concurrency limit is updated.
	//
	// +optional
	// +kubebuilder:default="0.1s"
	ConcurrencyUpdateInterval *metav1.Duration `json:"concurrencyUpdateInterval,omitempty"`

	// MaxConcurrencyLimit defines the maximum concurrency limit.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1000
	MaxConcurrencyLimit *uint32 `json:"maxConcurrencyLimit,omitempty"`

	// MinConcurrencyLimit defines the minimum concurrency limit.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	MinConcurrencyLimit *uint32 `json:"minConcurrencyLimit,omitempty"`

	// ConcurrencyUpdateRatio defines the ratio by which the concurrency limit is updated.
	// This should be a value between 0.0 and 1.0.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:validation:Maximum=1.0
	// +kubebuilder:default=0.1
	ConcurrencyUpdateRatio *float64 `json:"concurrencyUpdateRatio,omitempty"`
}

// AdaptiveConcurrencyMinRTTCalculation defines the minimum RTT calculation parameters.
//
// +optional
type AdaptiveConcurrencyMinRTTCalculation struct {
	// Interval defines the interval at which the minimum RTT is recalculated.
	//
	// +optional
	// +kubebuilder:default="60s"
	Interval *metav1.Duration `json:"interval,omitempty"`

	// RequestCount defines the number of requests to use for minimum RTT calculation.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=50
	RequestCount *uint32 `json:"requestCount,omitempty"`

	// Jitter defines the jitter to apply to the minimum RTT calculation interval.
	// This should be a value between 0 and 100 representing a percentage.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=10
	Jitter *float64 `json:"jitter,omitempty"`

	// Buffer defines the buffer to apply to the minimum RTT calculation.
	// This should be a value between 0 and 100 representing a percentage.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=25
	Buffer *float64 `json:"buffer,omitempty"`
}
