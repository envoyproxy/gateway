// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const (
	// KindClientTrafficPolicy is the name of the ClientTrafficPolicy kind.
	KindClientTrafficPolicy = "ClientTrafficPolicy"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=envoy-gateway,shortName=ctp
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ClientTrafficPolicy allows the user to configure the behavior of the connection
// between the downstream client and Envoy Proxy listener.
type ClientTrafficPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of ClientTrafficPolicy.
	Spec ClientTrafficPolicySpec `json:"spec"`

	// Status defines the current status of ClientTrafficPolicy.
	Status gwapiv1a2.PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="(has(self.targetRef) && !has(self.targetRefs)) || (!has(self.targetRef) && has(self.targetRefs)) || (has(self.targetSelectors) && self.targetSelectors.size() > 0) ", message="either targetRef or targetRefs must be used"
//
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.group == 'gateway.networking.k8s.io' : true", message="this policy can only have a targetRef.group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRef) ? self.targetRef.kind == 'Gateway' : true", message="this policy can only have a targetRef.kind of Gateway"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.group == 'gateway.networking.k8s.io') : true", message="this policy can only have a targetRefs[*].group of gateway.networking.k8s.io"
// +kubebuilder:validation:XValidation:rule="has(self.targetRefs) ? self.targetRefs.all(ref, ref.kind == 'Gateway') : true", message="this policy can only have a targetRefs[*].kind of Gateway"
//
// ClientTrafficPolicySpec defines the desired state of ClientTrafficPolicy.
type ClientTrafficPolicySpec struct {
	PolicyTargetReferences `json:",inline"`

	// TcpKeepalive settings associated with the downstream client connection.
	// If defined, sets SO_KEEPALIVE on the listener socket to enable TCP Keepalives.
	// Disabled by default.
	//
	// +optional
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty"`
	// EnableProxyProtocol interprets the ProxyProtocol header and adds the
	// Client Address into the X-Forwarded-For header.
	// Note Proxy Protocol must be present when this field is set, else the connection
	// is closed.
	//
	// Deprecated: Use ProxyProtocol instead.
	//
	// +optional
	EnableProxyProtocol *bool `json:"enableProxyProtocol,omitempty"`
	// ProxyProtocol configures the Proxy Protocol settings. When configured,
	// the Proxy Protocol header will be interpreted and the Client Address
	// will be added into the X-Forwarded-For header.
	// If both EnableProxyProtocol and ProxyProtocol are set, ProxyProtocol takes precedence.
	//
	// +optional
	ProxyProtocol *ProxyProtocolSettings `json:"proxyProtocol,omitempty"`
	// ClientIPDetectionSettings provides configuration for determining the original client IP address for requests.
	//
	// +optional
	ClientIPDetection *ClientIPDetectionSettings `json:"clientIPDetection,omitempty"`
	// TLS settings configure TLS termination settings with the downstream client.
	//
	// +optional
	TLS *ClientTLSSettings `json:"tls,omitempty"`
	// Path enables managing how the incoming path set by clients can be normalized.
	//
	// +optional
	Path *PathSettings `json:"path,omitempty"`
	// HeaderSettings provides configuration for header management.
	//
	// +optional
	Headers *HeaderSettings `json:"headers,omitempty"`
	// Timeout settings for the client connections.
	//
	// +optional
	Timeout *ClientTimeout `json:"timeout,omitempty"`
	// Connection includes client connection settings.
	//
	// +optional
	Connection *ClientConnection `json:"connection,omitempty"`
	// HTTP1 provides HTTP/1 configuration on the listener.
	//
	// +optional
	HTTP1 *HTTP1Settings `json:"http1,omitempty"`
	// HTTP2 provides HTTP/2 configuration on the listener.
	//
	// +optional
	HTTP2 *HTTP2Settings `json:"http2,omitempty"`
	// HTTP3 provides HTTP/3 configuration on the listener.
	//
	// +optional
	HTTP3 *HTTP3Settings `json:"http3,omitempty"`
	// HealthCheck provides configuration for determining whether the HTTP/HTTPS listener is healthy.
	//
	// +optional
	HealthCheck *HealthCheckSettings `json:"healthCheck,omitempty"`
}

// HeaderSettings provides configuration options for headers on the listener.
//
// +kubebuilder:validation:XValidation:rule="!(has(self.preserveXRequestID) && has(self.requestID))",message="preserveXRequestID and requestID cannot both be set."
type HeaderSettings struct {
	// EnableEnvoyHeaders configures Envoy Proxy to add the "X-Envoy-" headers to requests
	// and responses.
	// +optional
	EnableEnvoyHeaders *bool `json:"enableEnvoyHeaders,omitempty"`

	// DisableRateLimitHeaders configures Envoy Proxy to omit the "X-RateLimit-" response headers
	// when rate limiting is enabled.
	// +optional
	DisableRateLimitHeaders *bool `json:"disableRateLimitHeaders,omitempty"`

	// XForwardedClientCert configures how Envoy Proxy handle the x-forwarded-client-cert (XFCC) HTTP header.
	//
	// x-forwarded-client-cert (XFCC) is an HTTP header used to forward the certificate
	// information of part or all of the clients or proxies that a request has flowed through,
	// on its way from the client to the server.
	//
	// Envoy proxy may choose to sanitize/append/forward the XFCC header before proxying the request.
	//
	// If not set, the default behavior is sanitizing the XFCC header.
	// +optional
	XForwardedClientCert *XForwardedClientCert `json:"xForwardedClientCert,omitempty"`

	// WithUnderscoresAction configures the action to take when an HTTP header with underscores
	// is encountered. The default action is to reject the request.
	// +optional
	WithUnderscoresAction *WithUnderscoresAction `json:"withUnderscoresAction,omitempty"`

	// PreserveXRequestID configures Envoy to keep the X-Request-ID header if passed for a request that is edge
	// (Edge request is the request from external clients to front Envoy) and not reset it, which is the current Envoy behaviour.
	// Defaults to false and cannot be combined with RequestID.
	// Deprecated: use RequestID=Preserve instead
	//
	// +optional
	PreserveXRequestID *bool `json:"preserveXRequestID,omitempty"`

	// RequestID configures Envoy's behavior for handling the `X-Request-ID` header.
	// Defaults to `Generate` and builds the `X-Request-ID` for every request and ignores pre-existing values from the edge.
	// (An "edge request" refers to a request from an external client to the Envoy entrypoint.)
	//
	// +optional
	RequestID *RequestIDAction `json:"requestID,omitempty"`

	// EarlyRequestHeaders defines settings for early request header modification, before envoy performs
	// routing, tracing and built-in header manipulation.
	//
	// +optional
	EarlyRequestHeaders *gwapiv1.HTTPHeaderFilter `json:"earlyRequestHeaders,omitempty"`
}

// WithUnderscoresAction configures the action to take when an HTTP header with underscores
// is encountered.
// +kubebuilder:validation:Enum=Allow;RejectRequest;DropHeader
type WithUnderscoresAction string

const (
	// WithUnderscoresActionAllow allows headers with underscores to be passed through.
	WithUnderscoresActionAllow WithUnderscoresAction = "Allow"
	// WithUnderscoresActionRejectRequest rejects the client request. HTTP/1 requests are rejected with
	// the 400 status. HTTP/2 requests end with the stream reset.
	WithUnderscoresActionRejectRequest WithUnderscoresAction = "RejectRequest"
	// WithUnderscoresActionDropHeader drops the client header with name containing underscores. The header
	// is dropped before the filter chain is invoked and as such filters will not see
	// dropped headers.
	WithUnderscoresActionDropHeader WithUnderscoresAction = "DropHeader"
)

// RequestIDAction configures Envoy's behavior for handling the `X-Request-ID` header.
//
// +kubebuilder:validation:Enum=PreserveOrGenerate;Preserve;Generate;Disable
type RequestIDAction string

const (
	// Preserve `X-Request-ID` if already present or generate if empty
	RequestIDActionPreserveOrGenerate RequestIDAction = "PreserveOrGenerate"
	// Preserve `X-Request-ID` if already present, do not generate when empty
	RequestIDActionPreserve RequestIDAction = "Preserve"
	// Always generate `X-Request-ID` header, do not preserve `X-Request-ID`
	// header if it exists. This is the default behavior.
	RequestIDActionGenerate RequestIDAction = "Generate"
	// Do not preserve or generate `X-Request-ID` header
	RequestIDActionDisable RequestIDAction = "Disable"
)

// XForwardedClientCert configures how Envoy Proxy handle the x-forwarded-client-cert (XFCC) HTTP header.
// +kubebuilder:validation:XValidation:rule="(has(self.certDetailsToAdd) && self.certDetailsToAdd.size() > 0) ? (self.mode == 'AppendForward' || self.mode == 'SanitizeSet') : true",message="certDetailsToAdd can only be set when mode is AppendForward or SanitizeSet"
type XForwardedClientCert struct {
	// Mode defines how XFCC header is handled by Envoy Proxy.
	// If not set, the default mode is `Sanitize`.
	// +optional
	Mode *XFCCForwardMode `json:"mode,omitempty"`

	// CertDetailsToAdd specifies the fields in the client certificate to be forwarded in the XFCC header.
	//
	// Hash(the SHA 256 digest of the current client certificate) and By(the Subject Alternative Name)
	// are always included if the client certificate is forwarded.
	//
	// This field is only applicable when the mode is set to `AppendForward` or
	// `SanitizeSet` and the client connection is mTLS.
	// +kubebuilder:validation:MaxItems=5
	// +optional
	CertDetailsToAdd []XFCCCertData `json:"certDetailsToAdd,omitempty"`
}

// XFCCForwardMode defines how XFCC header is handled by Envoy Proxy.
// +kubebuilder:validation:Enum=Sanitize;ForwardOnly;AppendForward;SanitizeSet;AlwaysForwardOnly
type XFCCForwardMode string

const (
	// XFCCForwardModeSanitize removes the XFCC header from the request. This is the default mode.
	XFCCForwardModeSanitize XFCCForwardMode = "Sanitize"

	// XFCCForwardModeForwardOnly forwards the XFCC header in the request if the client connection is mTLS.
	XFCCForwardModeForwardOnly XFCCForwardMode = "ForwardOnly"

	// XFCCForwardModeAppendForward appends the client certificate information to the request’s XFCC header and forward it if the client connection is mTLS.
	XFCCForwardModeAppendForward XFCCForwardMode = "AppendForward"

	// XFCCForwardModeSanitizeSet resets the XFCC header with the client certificate information and forward it if the client connection is mTLS.
	// The existing certificate information in the XFCC header is removed.
	XFCCForwardModeSanitizeSet XFCCForwardMode = "SanitizeSet"

	// XFCCForwardModeAlwaysForwardOnly always forwards the XFCC header in the request, regardless of whether the client connection is mTLS.
	XFCCForwardModeAlwaysForwardOnly XFCCForwardMode = "AlwaysForwardOnly"
)

// XFCCCertData specifies the fields in the client certificate to be forwarded in the XFCC header.
// +kubebuilder:validation:Enum=Subject;Cert;Chain;DNS;URI
type XFCCCertData string

const (
	// XFCCCertDataSubject is the Subject field of the current client certificate.
	XFCCCertDataSubject XFCCCertData = "Subject"
	// XFCCCertDataCert is the entire client certificate in URL encoded PEM format.
	XFCCCertDataCert XFCCCertData = "Cert"
	// XFCCCertDataChain is the entire client certificate chain (including the leaf certificate) in URL encoded PEM format.
	XFCCCertDataChain XFCCCertData = "Chain"
	// XFCCCertDataDNS is the DNS type Subject Alternative Name field of the current client certificate.
	XFCCCertDataDNS XFCCCertData = "DNS"
	// XFCCCertDataURI is the URI type Subject Alternative Name field of the current client certificate.
	XFCCCertDataURI XFCCCertData = "URI"
)

// ClientIPDetectionSettings provides configuration for determining the original client IP address for requests.
//
// +kubebuilder:validation:XValidation:rule="!(has(self.xForwardedFor) && has(self.customHeader))",message="customHeader cannot be used in conjunction with xForwardedFor"
type ClientIPDetectionSettings struct {
	// XForwardedForSettings provides configuration for using X-Forwarded-For headers for determining the client IP address.
	//
	// +optional
	XForwardedFor *XForwardedForSettings `json:"xForwardedFor,omitempty"`
	// CustomHeader provides configuration for determining the client IP address for a request based on
	// a trusted custom HTTP header. This uses the custom_header original IP detection extension.
	// Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/http/original_ip_detection/custom_header/v3/custom_header.proto
	// for more details.
	//
	// +optional
	CustomHeader *CustomHeaderExtensionSettings `json:"customHeader,omitempty"`
}

// XForwardedForSettings provides configuration for using X-Forwarded-For headers for determining the client IP address.
// Refer to https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for
// for more details.
// +kubebuilder:validation:XValidation:rule="(has(self.numTrustedHops) && !has(self.trustedCIDRs)) || (!has(self.numTrustedHops) && has(self.trustedCIDRs))", message="only one of numTrustedHops or trustedCIDRs must be set"
type XForwardedForSettings struct {
	// NumTrustedHops controls the number of additional ingress proxy hops from the right side of XFF HTTP
	// headers to trust when determining the origin client's IP address.
	// Only one of NumTrustedHops and TrustedCIDRs must be set.
	//
	// +optional
	NumTrustedHops *uint32 `json:"numTrustedHops,omitempty"`

	// TrustedCIDRs is a list of CIDR ranges to trust when evaluating
	// the remote IP address to determine the original client’s IP address.
	// When the remote IP address matches a trusted CIDR and the x-forwarded-for header was sent,
	// each entry in the x-forwarded-for header is evaluated from right to left
	// and the first public non-trusted address is used as the original client address.
	// If all addresses in x-forwarded-for are within the trusted list, the first (leftmost) entry is used.
	// Only one of NumTrustedHops and TrustedCIDRs must be set.
	//
	// +optional
	// +kubebuilder:validation:MinItems=1
	TrustedCIDRs []CIDR `json:"trustedCIDRs,omitempty"`
}

// CustomHeaderExtensionSettings provides configuration for determining the client IP address for a request based on
// a trusted custom HTTP header. This uses the the custom_header original IP detection extension.
// Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/http/original_ip_detection/custom_header/v3/custom_header.proto
// for more details.
type CustomHeaderExtensionSettings struct {
	// Name of the header containing the original downstream remote address, if present.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Pattern="^[A-Za-z0-9-]+$"
	//
	Name string `json:"name"`
	// FailClosed is a switch used to control the flow of traffic when client IP detection
	// fails. If set to true, the listener will respond with 403 Forbidden when the client
	// IP address cannot be determined.
	//
	// +optional
	FailClosed *bool `json:"failClosed,omitempty"`
}

// HTTP3Settings provides HTTP/3 configuration on the listener.
type HTTP3Settings struct{}

// HTTP1Settings provides HTTP/1 configuration on the listener.
type HTTP1Settings struct {
	// EnableTrailers defines if HTTP/1 trailers should be proxied by Envoy.
	// +optional
	EnableTrailers *bool `json:"enableTrailers,omitempty"`
	// PreserveHeaderCase defines if Envoy should preserve the letter case of headers.
	// By default, Envoy will lowercase all the headers.
	// +optional
	PreserveHeaderCase *bool `json:"preserveHeaderCase,omitempty"`
	// HTTP10 turns on support for HTTP/1.0 and HTTP/0.9 requests.
	// +optional
	HTTP10 *HTTP10Settings `json:"http10,omitempty"`
	// DisableSafeMaxConnectionDuration controls the close behavior for HTTP/1 connections.
	// By default, connection closure is delayed until the next request arrives after maxConnectionDuration is exceeded.
	// It then adds a Connection: close header and gracefully closes the connection after the response completes.
	// When set to true (disabled), Envoy uses its default drain behavior, closing the connection shortly after maxConnectionDuration elapses.
	// Has no effect unless maxConnectionDuration is set.
	//
	// +optional
	// +notImplementedHide
	DisableSafeMaxConnectionDuration bool `json:"disableSafeMaxConnectionDuration,omitempty"`
}

// HTTP10Settings provides HTTP/1.0 configuration on the listener.
type HTTP10Settings struct {
	// UseDefaultHost defines if the HTTP/1.0 request is missing the Host header,
	// then the hostname associated with the listener should be injected into the
	// request.
	// If this is not set and an HTTP/1.0 request arrives without a host, then
	// it will be rejected.
	// +optional
	UseDefaultHost *bool `json:"useDefaultHost,omitempty"`
}

// HealthCheckSettings provides HealthCheck configuration on the HTTP/HTTPS listener.
type HealthCheckSettings struct {
	// Path specifies the HTTP path to match on for health check requests.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Path string `json:"path"`
}

// ProxyProtocolSettings configures the Proxy Protocol settings. When configured,
// the Proxy Protocol header will be interpreted and the Client Address
// will be added into the X-Forwarded-For header.
// If both EnableProxyProtocol and ProxyProtocol are set, ProxyProtocol takes precedence.
//
// +kubebuilder:validation:MinProperties=0
type ProxyProtocolSettings struct {
	// Optional allows requests without a Proxy Protocol header to be proxied.
	// If set to true, the listener will accept requests without a Proxy Protocol header.
	// If set to false, the listener will reject requests without a Proxy Protocol header.
	// If not set, the default behavior is to reject requests without a Proxy Protocol header.
	// Warning: Optional breaks conformance with the specification. Only enable if ALL traffic to the listener comes from a trusted source.
	// For more information on security implications, see haproxy.org/download/2.1/doc/proxy-protocol.txt
	//
	//
	// +optional
	Optional *bool `json:"optional,omitempty"`
}

//+kubebuilder:object:root=true

// ClientTrafficPolicyList contains a list of ClientTrafficPolicy resources.
type ClientTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClientTrafficPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClientTrafficPolicy{}, &ClientTrafficPolicyList{})
}
