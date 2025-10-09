// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type ClientTLSSettings struct {
	// ClientValidation specifies the configuration to validate the client
	// initiating the TLS connection to the Gateway listener.
	// +optional
	ClientValidation *ClientValidationContext `json:"clientValidation,omitempty"`
	TLSSettings      `json:",inline"`

	// Session defines settings related to TLS session management.
	// +optional
	Session *Session `json:"session,omitempty"`
}

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
	// exposed by the listener or used by the proxy to connect to the backend.
	// Defaults:
	// 1. HTTPS Routes: h2 and http/1.1 are enabled in listener context.
	// 2. Other Routes: ALPN is disabled.
	// 3. Backends: proxy uses the appropriate ALPN options for the backend protocol.
	// When an empty list is provided, the ALPN TLS extension is disabled.
	//
	// Defaults to [h2, http/1.1] if not specified.
	//
	// Typical Supported values are:
	// - http/1.0
	// - http/1.1
	// - h2
	//
	// +optional
	ALPNProtocols []ALPNProtocol `json:"alpnProtocols,omitempty"`
}

// ALPNProtocol specifies the protocol to be negotiated using ALPN
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
	// Optional set to true accepts connections even when a client doesn't present a certificate.
	// Defaults to false, which rejects connections without a valid client certificate.
	// +optional
	Optional bool `json:"optional,omitempty"`

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

	// An optional list of base64-encoded SHA-256 hashes. If specified, Envoy will
	// verify that the SHA-256 of the DER-encoded Subject Public Key Information
	// (SPKI) of the presented certificate matches one of the specified values.
	// +optional
	SPKIHashes []string `json:"spkiHashes,omitempty"`

	// An optional list of hex-encoded SHA-256 hashes. If specified, Envoy will
	// verify that the SHA-256 of the DER-encoded presented certificate matches
	// one of the specified values.
	// +optional
	CertificateHashes []string `json:"certificateHashes,omitempty"`

	// An optional list of Subject Alternative name matchers. If specified, Envoy
	// will verify that the Subject Alternative Name of the presented certificate
	// matches one of the specified matchers
	// +optional
	SubjectAltNames *SubjectAltNames `json:"subjectAltNames,omitempty"`

	// Crl specifies the crl configuration that can be used to validate the client initiating the TLS connection
	// +optional
	// +notImplementedHide
	Crl *CrlContext `json:"crl,omitempty"`
}

// CrlContext holds certificate revocation list configuration that can be used to validate the client initiating the TLS connection
type CrlContext struct {
	// Refs contains one or more references to a Kubernetes ConfigMap or a Kubernetes Secret,
	// containing the certificate revocation list in PEM format
	// Expects the content in a key named `ca.crl`.
	//
	// References to a resource in different namespace are invalid UNLESS there
	// is a ReferenceGrant in the target namespace that allows the crl
	// to be attached.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	Refs []gwapiv1.SecretObjectReference `json:"refs"`

	// If this option is set to true,  Envoy will only verify the certificate at the end of the certificate chain against the CRL.
	// Defaults to false, which will verify the entire certificate chain against the CRL.
	// +optional
	OnlyVerifyLeafCertificate *bool `json:"onlyVerifyLeafCertificate,omitempty"`
}

type SubjectAltNames struct {
	// DNS names matchers
	// +optional
	DNSNames []StringMatch `json:"dnsNames,omitempty"`

	// Email addresses matchers
	// +optional
	EmailAddresses []StringMatch `json:"emailAddresses,omitempty"`

	// IP addresses matchers
	// +optional
	IPAddresses []StringMatch `json:"ipAddresses,omitempty"`

	// URIs matchers
	// +optional
	URIs []StringMatch `json:"uris,omitempty"`

	// Other names matchers
	// +optional
	OtherNames []OtherSANMatch `json:"otherNames,omitempty"`
}

type OtherSANMatch struct {
	// OID Value
	Oid         string `json:"oid"`
	StringMatch `json:",inline"`
}

// Session defines settings related to TLS session management.
type Session struct {
	// Resumption determines the proxy's supported TLS session resumption option.
	// By default, Envoy Gateway does not enable session resumption. Use sessionResumption to
	// enable stateful and stateless session resumption. Users should consider security impacts
	// of different resumption methods. Performance gains from resumption are diminished when
	// Envoy proxy is deployed with more than one replica.
	// +optional
	Resumption *SessionResumption `json:"resumption,omitempty"`
}

// SessionResumption defines supported tls session resumption methods and their associated configuration.
type SessionResumption struct {
	// Stateless defines setting for stateless (session-ticket based) session resumption
	// +optional
	Stateless *StatelessTLSSessionResumption `json:"stateless,omitempty"`

	// Stateful defines setting for stateful (session-id based) session resumption
	// +optional
	Stateful *StatefulTLSSessionResumption `json:"stateful,omitempty"`
}

// StatefulTLSSessionResumption defines the stateful (session-id based) type of TLS session resumption.
// Note: When Envoy Proxy is deployed with more than one replica, session caches are not synchronized
// between instances, possibly leading to resumption failures.
// Envoy does not re-validate client certificates upon session resumption.
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-routematch-tlscontextmatchoptions
type StatefulTLSSessionResumption struct{}

// StatelessTLSSessionResumption defines the stateless (session-ticket based) type of TLS session resumption.
// Note: When Envoy Proxy is deployed with more than one replica, session ticket encryption keys are not
// synchronized between instances, possibly leading to resumption failures.
// In-memory session ticket encryption keys are rotated every 48 hours.
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/tls/v3/common.proto#extensions-transport-sockets-tls-v3-tlssessionticketkeys
// https://commondatastorage.googleapis.com/chromium-boringssl-docs/ssl.h.html#Session-tickets
type StatelessTLSSessionResumption struct{}
