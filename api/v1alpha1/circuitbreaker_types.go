// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

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

	// PerEndpoint defines Circuit Breakers that will apply per-endpoint for an upstream cluster
	//
	// +optional
	PerEndpoint *PerEndpointCircuitBreakers `json:"perEndpoint,omitempty"`

	// RetryBudget provides settings for retry budget, which limits the number of retries in a given percentage.
	// RetryBudget take precedence over maxParallelRetries.
	//
	// +optional
	RetryBudget *RetryBudget `json:"retryBudget,omitempty"`
}

// PerEndpointCircuitBreakers defines Circuit Breakers that will apply per-endpoint for an upstream cluster
type PerEndpointCircuitBreakers struct {
	// MaxConnections configures the maximum number of connections that Envoy will establish per-endpoint to the referenced backend defined within a xRoute rule.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +kubebuilder:default=1024
	// +optional
	MaxConnections *int64 `json:"maxConnections,omitempty"`
}

// RetryBudget specifies the details of the retry budget configuration, like
// the percentage of requests in the budget, and the min retry concurrency.
type RetryBudget struct {
	// Percent specifies the limit on concurrent retries as a percentage [0, 100] of
	// the sum of active requests and active pending requests.
	Percent gwapiv1.Fraction `json:"percent"`
	// MinRetryConcurrency specifies the minimum retry concurrency allowed for the retry budget.
	// For example, a budget of 20% with a minimum retry concurrency of 3
	// will allow 5 active retries while there are 25 active requests.
	// If there are 2 active requests, there are still 3 active retries
	// allowed because of the minimum retry concurrency.
	// Defaults to 3.
	//
	// +optional
	MinRetryConcurrency *uint32 `json:"minRetryConcurrency,omitempty"`
}
