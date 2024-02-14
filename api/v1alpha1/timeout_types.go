// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// Timeout defines configuration for timeouts related to connections.
type Timeout struct {
	// Timeout settings for TCP.
	//
	// +optional
	TCP *TCPTimeout `json:"tcp,omitempty"`

	// Timeout settings for HTTP.
	//
	// +optional
	HTTP *HTTPTimeout `json:"http,omitempty"`
}

type TCPTimeout struct {
	// The timeout for network connection establishment, including TCP and TLS handshakes.
	// Default: 10 seconds.
	//
	// +optional
	ConnectTimeout *gwapiv1.Duration `json:"connectTimeout,omitempty"`
}

type HTTPTimeout struct {
	// The idle timeout for an HTTP connection. Idle time is defined as a period in which there are no active requests in the connection.
	// Default: 1 hour.
	//
	// +optional
	ConnectionIdleTimeout *gwapiv1.Duration `json:"connectionIdleTimeout,omitempty"`

	// The maximum duration of an HTTP connection.
	// Default: unlimited.
	//
	// +optional
	MaxConnectionDuration *gwapiv1.Duration `json:"maxConnectionDuration,omitempty"`
}

type ClientTimeout struct {
	// Timeout settings for HTTP.
	//
	// +optional
	HTTP *HTTPClientTimeout `json:"http,omitempty"`
}

type HTTPClientTimeout struct {
	// The timeout for processing the client request. Starting upon request initiation and ending when the last byte is sent to backend.
	// Default: 300 seconds.
	//
	// +optional
	RequestProcessTimeout *gwapiv1.Duration `json:"requestProcessTimeout"`
}
