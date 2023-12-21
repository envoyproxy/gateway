// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// +kubebuilder:validation:XValidation:rule="has(self.minVersion) && self.minVersion == 'v1_3' ? !has(self.ciphers) : true", message="setting ciphers has no effect if the minimum possible TLS version is 1.3"
// +kubebuilder:validation:XValidation:rule="has(self.minVersion) && has(self.maxVersion) ? {\"Auto\":0,\"v1_1\":1,\"v1_2\":2,\"v1_3\":3}[self.minVersion] <= {\"v1_1\":1,\"v1_2\":2,\"v1_3\":3,\"Auto\":4}[self.maxVersion] : !has(self.minVersion) && has(self.maxVersion) ? 2 <= {\"v1_1\":1,\"v1_2\":2,\"v1_3\":3,\"Auto\":4}[self.maxVersion] : true", message="minVersion must be smaller or equal to maxVersion"
type TLSSettings struct {

	// Min specifies the minimal TLS protocol version to allow.
	//
	// The default is TLS 1.2 if this is not specified.
	// +optional
	MinVersion *TLSVersion `json:"minVersion,omitempty"`

	// Max specifies the maximal TLS protocol version to allow
	//
	// The default is TLS 1.3 if this is not specified.
	// +optional
	MaxVersion *TLSVersion `json:"maxVersion,omitempty"`

	// Ciphers specifies the set of cipher suites supported when
	// negotiating TLS 1.0 - 1.2. This setting has no effect for TLS 1.3.
	//
	// In non-FIPS Envoy Proxy builds the default cipher list is:
	// - [ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]
	// - [ECDHE-RSA-AES128-GCM-SHA256|ECDHE-RSA-CHACHA20-POLY1305]
	// - ECDHE-ECDSA-AES256-GCM-SHA384
	// - ECDHE-RSA-AES256-GCM-SHA384
	//
	// In builds using BoringSSL FIPS the default cipher list is:
	// - ECDHE-ECDSA-AES128-GCM-SHA256
	// - ECDHE-RSA-AES128-GCM-SHA256
	// - ECDHE-ECDSA-AES256-GCM-SHA384
	// - ECDHE-RSA-AES256-GCM-SHA384
	//
	// +optional
	Ciphers []string `json:"ciphers,omitempty"`

	// ECDHCurves specifies the set of supported ECDH curves.
	// In non-FIPS Envoy Proxy builds the default curves are:
	// - X25519
	// - P-256
	//
	// In builds using BoringSSL FIPS the default curve is:
	// - P-256
	//
	// +optional
	ECDHCurves []string `json:"ecdhCurves,omitempty"`

	// SignatureAlgorithms specifies which signature algorithms the listener should
	// support.
	//
	// +optional
	SignatureAlgorithms []string `json:"signatureAlgorithms,omitempty"`

	// ALPNProtocols supplies the list of ALPN protocols that should be
	// exposed by the listener. By default http/2 and http/1.1 are enabled.
	//
	// Supported values are:
	// - http/1.0
	// - http/1.1
	// - http/2
	//
	// +optional
	ALPNProtocols []ALPNProtocol `json:"alpnProtocols,omitempty"`
}

// ALPNProtocol specifies the protocol to be negotiated using ALPN
// +kubebuilder:validation:Enum=http/1.0;http/1.1;http/2
type ALPNProtocol string

const (
	// HTTPProtocolVersion1_0 specifies that HTTP/1.0 should be negotiable with ALPN
	HTTPProtocolVersion1_0 ALPNProtocol = "http/1.0"
	// HTTPProtocolVersion1_1 specifies that HTTP/1.1 should be negotiable with ALPN
	HTTPProtocolVersion1_1 ALPNProtocol = "http/1.1"
	// HTTPProtocolVersion2 specifies that HTTP/2 should be negotiable with ALPN
	HTTPProtocolVersion2 ALPNProtocol = "http/2"
)

// TLSVersion specifies the TLS version
// +kubebuilder:validation:Enum=Auto;v1_0;v1_1;v1_2;v1_3
type TLSVersion string

const (
	// TLSAuto allows Envoy to choose the optimal TLS Version
	TLSAuto TLSVersion = "Auto"
	// TLSv1_0 specifies TLS version 1.0
	TLSv10 TLSVersion = "v1_0"
	// TLSv1_1 specifies TLS version 1.1
	TLSv11 TLSVersion = "v1_1"
	// TLSv1.2 specifies TLS version 1.2
	TLSv12 TLSVersion = "v1_2"
	// TLSv1.3 specifies TLS version 1.3
	TLSv13 TLSVersion = "v1_3"
)
