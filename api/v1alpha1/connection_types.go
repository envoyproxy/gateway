// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// Connection allows users to configure connection-level settings
type Connection struct {
	// Limit defines limits related to connections
	//
	// +optional
	Limit *ConnectionLimit `json:"limit,omitempty"`
}

type ConnectionLimit struct {
	// Value is the maximum concurrent connections limit.
	// When the limit is reached, incoming connections will be closed after the CloseDelay duration.
	// Default: unlimited.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	Value *int64 `json:"value,omitempty"`

	// CloseDelay defines the delay to use before closing connections that are rejected
	// because of of the Value limit.
	// Default: none.
	//
	// +optional
	CloseDelay *gwapiv1.Duration `json:"closeDelay,omitempty"`
}
