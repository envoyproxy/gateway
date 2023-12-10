// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type TLSSettings struct {
	// Version details the minimum/maximum TLS protocol verison that
	// should be supported by this listener.
	// +optional
	Version *TLSVersions `json:"version,omitempty"`

	// CipherSuites specifies the set of cipher suites supported when
	// negotiating TLS 1.0 - 1.2. This setting has no effect for TLS 1.3.
	// +optional
	CipherSuites []string `json:"ciphers,omitempty"`

	// ECDHCurves specifies the set of supported ECDH curves.
	// +optional
	ECDHCurves []string `json:"ecdhCurves,omitempty"`

	// SignatureAlgorithms specifies which signature algorithms the listener should
	// support.
	// +optional
	SignatureAlgorithms []string `json:"signatureAlgorithms,omitempty"`

	// ALPNProtocols supplies the list of ALPN protocols that should be
	// exposed by the listener. If left empty, ALPN will not be exposed.
	// +optional
	ALPNProtocols []string `json:"alpnProtocols,omitempty"`
}

// TLSVersion specifies the TLS version
// +kubebuilder:validation:Enum=TLS_Auto;TLSv1_0;TLSv1_2;TLSv1_3
type TLSVersion string

const (
	// TLSAuto allows Envoy to choose the optimal TLS Version
	TLSAuto TLSVersion = "TLS_Auto"
	// TLSv1_0 specifies TLS version 1.0
	TLSv10 TLSVersion = "TLSv1_0"
	// TLSv1_1 specifies TLS version 1.1
	TLSv11 TLSVersion = "TLSv1_1"
	// TLSv1.2 specifies TLS version 1.2
	TLSv12 TLSVersion = "TLSv1_2"
	// TLSv1.3 specifies TLS version 1.3
	TLSv13 TLSVersion = "TLSv1_3"
)

type TLSVersions struct {
	// Min specifies the minimal TLS verison to use
	// +optional
	Min TLSVersion `json:"min,omitempty"`

	// Max specifies the maximal TLS verison to use
	// +optional
	Max TLSVersion `json:"max,omitempty"`
}
