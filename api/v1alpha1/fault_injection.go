// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// FaultInjection defines the fault injection policy to be applied. Support delays and aborts.
type FaultInjection struct {

	// If specified, the delay will inject a fixed delay into the request
	//
	// +optional
	Delay *DelayConfig `json:"delay,omitempty"`

	// If specified, the abort will abort the request with the specified HTTP status code
	//
	// +optional
	Abort *AbortConfig `json:"abort,omitempty"`
}

// DelayConfig defines the delay fault injection configuration
type DelayConfig struct {
	// FixedDelay specifies the fixed delay duration
	//
	// +required
	FixedDelay *metav1.Duration `json:"fixedDelay"`

	// Percentage specifies the percentage of requests to be delayed. Default 100%, if set 0, no requests will be delayed.
	// +optional
	// +kubebuilder:default=100
	Percentage *int `json:"percentage"`
}

// AbortConfig defines the abort fault injection configuration
type AbortConfig struct {
	// StatusCode specifies the HTTP/GRPC status code to be returned
	//
	// +required
	StatusCode *int `json:"statusCode"`

	// Percentage specifies the percentage of requests to be aborted. Default 100%, if set 0, no requests will be aborted.
	// +optional
	// +kubebuilder:default=100
	Percentage *int `json:"percentage"`
}
