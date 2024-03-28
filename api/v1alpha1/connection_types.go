// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Connection allows users to configure connection-level settings
type Connection struct {
	// ConnectionLimit defines limits related to connections
	//
	// +optional
	ConnectionLimit *ConnectionLimit `json:"connectionLimit,omitempty"`
	// BufferLimit provides configuration for the maximum buffer size in bytes for each incoming connection.
	// For example, 20Mi, 1Gi, 256Ki etc.
	// Note that when the suffix is not provided, the value is interpreted as bytes.
	// Default: 32768 bytes.
	//
	// +kubebuilder:validation:XValidation:rule="type(self) == string ? self.matches(r\"^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$\") : type(self) == int",message="bufferLimit must be of the format \"^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$\""
	// +optional
	BufferLimit *resource.Quantity `json:"bufferLimit,omitempty"`
}

type ConnectionLimit struct {
	// Value of the maximum concurrent connections limit.
	// When the limit is reached, incoming connections will be closed after the CloseDelay duration.
	// Default: unlimited.
	//
	// +kubebuilder:validation:Minimum=0
	Value int64 `json:"value,omitempty"`

	// CloseDelay defines the delay to use before closing connections that are rejected
	// once the limit value is reached.
	// Default: none.
	//
	// +optional
	CloseDelay *gwapiv1.Duration `json:"closeDelay,omitempty"`
}
