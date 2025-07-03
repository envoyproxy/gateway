// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// ClientConnection allows users to configure connection-level settings of client
type ClientConnection struct {
	// ConnectionLimit defines limits related to connections
	//
	// +optional
	ConnectionLimit *ConnectionLimit `json:"connectionLimit,omitempty"`
	// BufferLimit provides configuration for the maximum buffer size in bytes for each incoming connection.
	// BufferLimit applies to connection streaming (maybe non-streaming) channel between processes, it's in user space.
	// For example, 20Mi, 1Gi, 256Ki etc.
	// Note that when the suffix is not provided, the value is interpreted as bytes.
	// Default: 32768 bytes.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	BufferLimit *resource.Quantity `json:"bufferLimit,omitempty"`
	// SocketBufferLimit provides configuration for the maximum buffer size in bytes for each incoming socket.
	// SocketBufferLimit applies to socket streaming channel between TCP/IP stacks, it's in kernel space.
	// For example, 20Mi, 1Gi, 256Ki etc.
	// Note that when the suffix is not provided, the value is interpreted as bytes.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	// +notImplementedHide
	SocketBufferLimit *resource.Quantity `json:"socketBufferLimit,omitempty"`

	// MaxAcceptPerSocketEvent provides configuration for the maximum number of connections to accept from the kernel
	// per socket event. If there are more than MaxAcceptPerSocketEvent connections pending accept, connections over
	// this threshold will be accepted in later event loop iterations.
	// Defaults to 1 and can be disabled by setting to 0 for allowing unlimited accepted connections.
	//
	// +optional
	// +kubebuilder:default=1
	MaxAcceptPerSocketEvent *uint32 `json:"maxAcceptPerSocketEvent,omitempty"`
}

// BackendConnection allows users to configure connection-level settings of backend
type BackendConnection struct {
	// BufferLimit Soft limit on size of the cluster’s connections read and write buffers.
	// BufferLimit applies to connection streaming (maybe non-streaming) channel between processes, it's in user space.
	// If unspecified, an implementation defined default is applied (32768 bytes).
	// For example, 20Mi, 1Gi, 256Ki etc.
	// Note: that when the suffix is not provided, the value is interpreted as bytes.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	BufferLimit *resource.Quantity `json:"bufferLimit,omitempty"`
	// SocketBufferLimit provides configuration for the maximum buffer size in bytes for each socket
	// to backend.
	// SocketBufferLimit applies to socket streaming channel between TCP/IP stacks, it's in kernel space.
	// For example, 20Mi, 1Gi, 256Ki etc.
	// Note that when the suffix is not provided, the value is interpreted as bytes.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	// +notImplementedHide
	SocketBufferLimit *resource.Quantity `json:"socketBufferLimit,omitempty"`
}

type ConnectionLimit struct {
	// Value of the maximum concurrent connections limit.
	// When the limit is reached, incoming connections will be closed after the CloseDelay duration.
	//
	// +kubebuilder:validation:Minimum=1
	Value int64 `json:"value"`

	// CloseDelay defines the delay to use before closing connections that are rejected
	// once the limit value is reached.
	// Default: none.
	//
	// +optional
	CloseDelay *gwapiv1.Duration `json:"closeDelay,omitempty"`
}
