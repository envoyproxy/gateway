// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// ProxyProtocol defines the configuration related to the proxy protocol
// when communicating with the backend.
type ProxyProtocol struct {
	// Version of ProxyProtol
	// Valid ProxyProtocolVersion values are
	// "V1"
	// "V2"
	Version ProxyProtocolVersion `json:"version"`
}

// ProxyProtocolVersion defines the version of the Proxy Protocol to use.
// +kubebuilder:validation:Enum=V1;V2
type ProxyProtocolVersion string

const (
	// ProxyProtocolVersionV1 is the PROXY protocol version 1 (human readable format).
	ProxyProtocolVersionV1 ProxyProtocolVersion = "V1"
	// ProxyProtocolVersionV2 is the PROXY protocol version 2 (binary format).
	ProxyProtocolVersionV2 ProxyProtocolVersion = "V2"
)
