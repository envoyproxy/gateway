// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// ConnectionTimeouts defines configuration for timeouts related to connections.
type ConnectionTimeouts struct {
	// The timeout for new network connection establishment, including TCP and TLS handshakes.
	//
	// +optional
	ConnectTimeout *gwapiv1.Duration `json:"connectTimeout,omitempty"`

	// The idle timeout for connections. The idle time is defined as a period in which there are no active requests in the connection.
	//
	// +optional
	HTTPIdleTimeout *gwapiv1.Duration `json:"httpIdleTimeout,omitempty"`

	// The maximum duration of a connection.
	//
	// +optional
	HTTPMaxConnectionDuration *gwapiv1.Duration `json:"httpMaxConnectionDuration,omitempty"`
}
