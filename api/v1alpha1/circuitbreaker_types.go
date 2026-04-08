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

	// RetryBudget specifies a limit on concurrent retries in relation to the number of active requests.
	// If this field is set, the retry budget will override any configured retry circuit breaker (MaxParallelRetries).
	//
	// +optional
	RetryBudget *RetryBudget `json:"retryBudget,omitempty"`

	// PerEndpoint defines Circuit Breakers that will apply per-endpoint for an upstream cluster
	//
	// +optional
	PerEndpoint *PerEndpointCircuitBreakers `json:"perEndpoint,omitempty"`
}

// RetryBudget specifies a limit on concurrent retries in relation to the number of active requests.
type RetryBudget struct {
	// BudgetPercent specifies the limit on concurrent retries as a percentage of the sum
	// of active requests and active pending requests. For example, if there are 100 active
	// requests and the budget_percent is set to 25, there may be 25 active retries.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:default=20.0
	// +optional
	BudgetPercent *float64 `json:"budgetPercent,omitempty"`

	// MinRetryConcurrency specifies the minimum retry concurrency allowed for the retry budget.
	// The limit on the number of active retries may never go below this number.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +kubebuilder:default=3
	// +optional
	MinRetryConcurrency *int64 `json:"minRetryConcurrency,omitempty"`
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
