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

	// Preconnect configures proactive upstream connections to reduce latency by establishing
	// connections before they’re needed and avoiding connection establishment overhead.
	//
	// If unset, Envoy will fetch connections as needed to serve in-flight requests.
	//
	// +optional
	Preconnect *PreconnectPolicy `json:"preconnect,omitempty"`
}

// Preconnect configures proactive upstream connections to avoid
// connection establishment overhead and reduce latency.
type PreconnectPolicy struct {
	// PerEndpointPercent configures how many additional connections to maintain per
	// upstream endpoint, useful for high-QPS or latency sensitive services. Expressed as a
	// percentage of the connections required by active streams
	// (e.g. 100 = preconnect disabled, 105 = 1.05x connections per-endpoint, 200 = 2.00×).
	//
	// Allowed value range is between 100-300. When both PerEndpointPercent and
	// PredictivePercent are set, Envoy ensures both are satisfied (max of the two).
	//
	// +kubebuilder:validation:Minimum=100
	// +kubebuilder:validation:Maximum=300
	// +optional
	PerEndpointPercent *uint32 `json:"perEndpointPercent,omitempty"`

	// PredictivePercent configures how many additional connections to maintain
	// across the cluster by anticipating which upstream endpoint the load balancer
	// will select next, useful for low-QPS services. Relies on deterministic
	// loadbalancing and is only supported with Random or RoundRobin.
	// Expressed as a percentage of the connections required by active streams
	// (e.g. 100 = 1.0 (no preconnect), 105 = 1.05× connections across the cluster, 200 = 2.00×).
	//
	// Minimum allowed value is 100. When both PerEndpointPercent and PredictivePercent are
	// set Envoy ensures both are satisfied per host (max of the two).
	//
	// +kubebuilder:validation:Minimum=100
	// +optional
	PredictivePercent *uint32 `json:"predictivePercent,omitempty"`
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
	// MaxConnectionDuration is the maximum amount of time a connection can remain established
	// (usually via TCP/HTTP Keepalive packets) before being drained and/or closed.
	// If not specified, there is no limit.
	//
	// +optional
	MaxConnectionDuration *gwapiv1.Duration `json:"maxConnectionDuration,omitempty"`
	// MaxRequestsPerConnection defines the maximum number of requests allowed over a single connection.
	// If not specified, there is no limit. Setting this parameter to 1 will effectively disable keep alive.
	//
	// +optional
	MaxRequestsPerConnection *uint32 `json:"maxRequestsPerConnection,omitempty"`
	// MaxStreamDuration is the maximum amount of time to keep alive an http stream. When the limit is reached
	// the stream will be reset independent of any other timeouts. If not specified, no value is set.
	//
	// +optional
	MaxStreamDuration *gwapiv1.Duration `json:"maxStreamDuration,omitempty"`
}
