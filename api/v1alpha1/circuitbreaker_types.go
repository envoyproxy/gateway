// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// CircuitBreakers defines the Circuit Breakers configuration.
type CircuitBreakers struct {
	// List of Circuit Breaker Thresholds
	// At most one Thresholds resource is supported.
	//
	// +kubebuilder:validation:MaxItems:=1
	// +optional
	Thresholds []Thresholds `json:"thresholds,omitempty"`
}

type Thresholds struct {
	// The maximum number of connections that Envoy will make to the referenced backend (per xRoute).
	// Default: 1024
	//
	// +optional
	MaxConnections *uint32 `json:"maxConnections,omitempty"`

	// The maximum number of pending requests that Envoy will allow to the referenced backend (per xRoute).
	// Default: 1024
	//
	// +optional
	MaxPendingRequests *uint32 `json:"maxPendingRequests,omitempty"`

	// The maximum number of parallel requests that Envoy will make to the referenced backend (per xRoute).
	// Default: 1024
	//
	// +optional
	MaxRequests *uint32 `json:"maxParallelRequests,omitempty"`

	// The maximum number of parallel retries that Envoy will allow to the referenced backend (per xRoute).
	// Default: 3
	//
	// +optional
	MaxRetries *uint32 `json:"maxRetries,omitempty"`
}

