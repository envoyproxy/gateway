// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// CircuitBreaker defines the Circuit Breaker configuration.
type CircuitBreaker struct {
	// The maximum number of connections that Envoy will establish to the referenced backend defined within a xRoute rule.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +kubebuilder:default=1024
	// +optional
	MaxConnections *int64 `json:"maxConnections,omitempty"`

	// The maximum number of pending requests that Envoy will queue to the referenced backend defined within a xRoute rule.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +kubebuilder:default=1024
	// +optional
	MaxPendingRequests *int64 `json:"maxPendingRequests,omitempty"`

	// The maximum number of parallel requests that Envoy will make to the referenced backend defined within a xRoute rule.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +kubebuilder:default=1024
	// +optional
	MaxParallelRequests *int64 `json:"maxParallelRequests,omitempty"`

	// The maximum number of parallel retries that Envoy will make to the referenced backend defined within a xRoute rule.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +kubebuilder:default=1024
	// +optional
	MaxParallelRetries *int64 `json:"maxParallelRetries,omitempty"`

	// The maximum number of requests that Envoy will make over a single connection to the referenced backend defined within a xRoute rule.
	// Default: unlimited.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +optional
	MaxRequestsPerConnection *int64 `json:"maxRequestsPerConnection,omitempty"`

	// Defines per-host Circuit Breaker thresholds
	// +optional
	// +notImplementedHide
	PerHostThresholds *PerHostCircuitBreakers `json:"perHostThresholds,omitempty"`
}

// PerHostCircuitBreakers defines the per-host Circuit Breaker configuration.
type PerHostCircuitBreakers struct {
	// The maximum number of connections that Envoy will establish per-host to the referenced backend defined within a xRoute rule.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +optional
	// +notImplementedHide
	MaxConnections *int64 `json:"maxConnections,omitempty"`
}
