// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// TCPKeepalive define the TCP Keepalive configuration.
type TCPKeepalive struct {
	// The total number of unacknowledged probes to send before deciding
	// the connection is dead.
	// Defaults to 9.
	//
	// +optional
	Probes *uint32 `json:"probes,omitempty"`
	// The duration a connection needs to be idle before keep-alive
	// probes start being sent.
	// The duration format is
	// Defaults to `7200s`.
	//
	// +optional
	IdleTime *gwapiv1.Duration `json:"idleTime,omitempty"`
	// The duration between keep-alive probes.
	// Defaults to `75s`.
	//
	// +optional
	Interval *gwapiv1.Duration `json:"interval,omitempty"`
}
