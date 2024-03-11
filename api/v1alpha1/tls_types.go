// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +kubebuilder:validation:XValidation:rule="has(self.minVersion) && self.minVersion == '1.3' ? !has(self.ciphers) : true", message="setting ciphers has no effect if the minimum possible TLS version is 1.3"
// +kubebuilder:validation:XValidation:rule="has(self.minVersion) && has(self.maxVersion) ? {\"Auto\":0,\"1.0\":1,\"1.1\":2,\"1.2\":3,\"1.3\":4}[self.minVersion] <= {\"1.0\":1,\"1.1\":2,\"1.2\":3,\"1.3\":4,\"Auto\":5}[self.maxVersion] : !has(self.minVersion) && has(self.maxVersion) ? 3 <= {\"1.0\":1,\"1.1\":2,\"1.2\":3,\"1.3\":4,\"Auto\":5}[self.maxVersion] : true", message="minVersion must be smaller or equal to maxVersion"
type TLSSettings struct {

	// Min specifies the minimal TLS protocol version to allow.
	// The default is TLS 1.2 if this is not specified.
	//
	// +optional
	MinVersion *TLSVersion `json:"minVersion,omitempty"`

	// Max specifies the maximal TLS protocol version to allow
	// The default is TLS 1.3 if this is not specified.
	//
	// +optional
	MaxVersion *TLSVersion `json:"maxVersion,omitempty"`

	// Ciphers specifies the set of cipher suites supported when
	// negotiating TLS 1.0 - 1.2. This setting has no effect for TLS 1.3.
	// In non-FIPS Envoy Proxy builds the default cipher list is:
	// - [ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]
	// - [ECDHE-RSA-AES128-GCM-SHA256|ECDHE-RSA-CHACHA20-POLY1305]
	// - ECDHE-ECDSA-AES256-GCM-SHA384
	// - ECDHE-RSA-AES256-GCM-SHA384
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
	// exposed by the listener. By default h2 and http/1.1 are enabled.
	// Supported values are:
	// - http/1.0
	// - http/1.1
	// - h2
	//
	// +optional
	ALPNProtocols []ALPNProtocol `json:"alpnProtocols,omitempty"`

	// ClientValidation specifies the configuration to validate the client
	// initiating the TLS connection to the Gateway listener.
	// +optional
	ClientValidation *ClientValidationContext `json:"clientValidation,omitempty"`
}

// ALPNProtocol specifies the protocol to be negotiated using ALPN
// +kubebuilder:validation:Enum=http/1.0;http/1.1;h2
type ALPNProtocol string

// When adding ALPN constants, they must be values that are defined
// in the IANA registry for ALPN identification sequences
// https://www.iana.org/assignments/tls-extensiontype-values/tls-extensiontype-values.xhtml#alpn-protocol-ids
const (
	// HTTPProtocolVersion1_0 specifies that HTTP/1.0 should be negotiable with ALPN
	HTTPProtocolVersion1_0 ALPNProtocol = "http/1.0"
	// HTTPProtocolVersion1_1 specifies that HTTP/1.1 should be negotiable with ALPN
	HTTPProtocolVersion1_1 ALPNProtocol = "http/1.1"
	// HTTPProtocolVersion2 specifies that HTTP/2 should be negotiable with ALPN
	HTTPProtocolVersion2 ALPNProtocol = "h2"
)

// TLSVersion specifies the TLS version
// +kubebuilder:validation:Enum=Auto;"1.0";"1.1";"1.2";"1.3"
type TLSVersion string

const (
	// TLSAuto allows Envoy to choose the optimal TLS Version
	TLSAuto TLSVersion = "Auto"
	// TLS1.0 specifies TLS version 1.0
	TLSv10 TLSVersion = "1.0"
	// TLS1.1 specifies TLS version 1.1
	TLSv11 TLSVersion = "1.1"
	// TLSv1.2 specifies TLS version 1.2
	TLSv12 TLSVersion = "1.2"
	// TLSv1.3 specifies TLS version 1.3
	TLSv13 TLSVersion = "1.3"
)

// ClientValidationContext holds configuration that can be used to validate the client initiating the TLS connection
// to the Gateway.
// By default, no client specific configuration is validated.
type ClientValidationContext struct {
	// CACertificateRefs contains one or more references to
	// Kubernetes objects that contain TLS certificates of
	// the Certificate Authorities that can be used
	// as a trust anchor to validate the certificates presented by the client.
	//
	// A single reference to a Kubernetes ConfigMap or a Kubernetes Secret,
	// with the CA certificate in a key named `ca.crt` is currently supported.
	//
	// References to a resource in different namespace are invalid UNLESS there
	// is a ReferenceGrant in the target namespace that allows the certificate
	// to be attached.
	//
	// +kubebuilder:validation:MaxItems=8
	// +optional
	CACertificateRefs []gwapiv1.SecretObjectReference `json:"caCertificateRefs,omitempty"`
}
