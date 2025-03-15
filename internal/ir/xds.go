// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"cmp"
	"crypto/tls"
	"crypto/x509"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"reflect"

	"golang.org/x/exp/slices"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const (
	EmptyPath = ""
)

var (
	ErrListenerNameEmpty                        = errors.New("field Name must be specified")
	ErrListenerAddressInvalid                   = errors.New("field Address must be a valid IP address")
	ErrListenerPortInvalid                      = errors.New("field Port specified is invalid")
	ErrHTTPListenerHostnamesEmpty               = errors.New("field Hostnames must be specified with at least a single hostname entry")
	ErrTCPRouteSNIsEmpty                        = errors.New("field SNIs must be specified with at least a single server name entry")
	ErrTLSServerCertEmpty                       = errors.New("field ServerCertificate must be specified")
	ErrTLSPrivateKey                            = errors.New("field PrivateKey must be specified")
	ErrRouteNameEmpty                           = errors.New("field Name must be specified")
	ErrHTTPRouteHostnameEmpty                   = errors.New("field Hostname must be specified")
	ErrRouteDestinationsFQDNMixed               = errors.New("mixed endpoints address type for the same route destination is not supported")
	ErrDestinationNameEmpty                     = errors.New("field Name must be specified")
	ErrDestEndpointHostInvalid                  = errors.New("field Address must be a valid IP or FQDN address")
	ErrDestEndpointPortInvalid                  = errors.New("field Port specified is invalid")
	ErrDestEndpointUDSPortInvalid               = errors.New("field Port must not be specified for Unix Domain Socket address")
	ErrDestEndpointUDSHostInvalid               = errors.New("field Host must not be specified for Unix Domain Socket address")
	ErrStringMatchConditionInvalid              = errors.New("only one of the Exact, Prefix, SafeRegex or Distinct fields must be set")
	ErrStringMatchInvertDistinctInvalid         = errors.New("only one of the Invert or Distinct fields can be set")
	ErrStringMatchNameIsEmpty                   = errors.New("field Name must be specified")
	ErrDirectResponseStatusInvalid              = errors.New("only HTTP status codes 100 - 599 are supported for DirectResponse")
	ErrRedirectUnsupportedStatus                = errors.New("only HTTP status codes 301 and 302 are supported for redirect filters")
	ErrRedirectUnsupportedScheme                = errors.New("only http and https are supported for the scheme in redirect filters")
	ErrHTTPPathModifierDoubleReplace            = errors.New("redirect filter cannot have a path modifier that supplies more than one of fullPathReplace, prefixMatchReplace and regexMatchReplace")
	ErrHTTPPathModifierNoReplace                = errors.New("redirect filter cannot have a path modifier that does not supply either fullPathReplace, prefixMatchReplace or regexMatchReplace")
	ErrHTTPPathRegexModifierNoSetting           = errors.New("redirect filter cannot have a path modifier that does not supply either fullPathReplace, prefixMatchReplace or regexMatchReplace")
	ErrHTTPHostModifierDoubleReplace            = errors.New("redirect filter cannot have a host modifier that supplies more than one of Hostname, Header and Backend")
	ErrAddHeaderEmptyName                       = errors.New("header modifier filter cannot configure a header without a name to be added")
	ErrAddHeaderDuplicate                       = errors.New("header modifier filter attempts to add the same header more than once (case insensitive)")
	ErrRemoveHeaderDuplicate                    = errors.New("header modifier filter attempts to remove the same header more than once (case insensitive)")
	ErrLoadBalancerInvalid                      = errors.New("loadBalancer setting is invalid, only one setting can be set")
	ErrHealthCheckTimeoutInvalid                = errors.New("field HealthCheck.Timeout must be specified")
	ErrHealthCheckIntervalInvalid               = errors.New("field HealthCheck.Interval must be specified")
	ErrHealthCheckUnhealthyThresholdInvalid     = errors.New("field HealthCheck.UnhealthyThreshold should be greater than 0")
	ErrHealthCheckHealthyThresholdInvalid       = errors.New("field HealthCheck.HealthyThreshold should be greater than 0")
	ErrHealthCheckerInvalid                     = errors.New("health checker setting is invalid, only one health checker can be set")
	ErrHCHTTPHostInvalid                        = errors.New("field HTTPHealthChecker.Host should be specified")
	ErrHCHTTPPathInvalid                        = errors.New("field HTTPHealthChecker.Path should be specified")
	ErrHCHTTPMethodInvalid                      = errors.New("only one of the GET, HEAD, POST, DELETE, OPTIONS, TRACE, PATCH of HTTPHealthChecker.Method could be set")
	ErrHCHTTPExpectedStatusesInvalid            = errors.New("field HTTPHealthChecker.ExpectedStatuses should be specified")
	ErrHealthCheckPayloadInvalid                = errors.New("one of Text, Binary fields must be set in payload")
	ErrHTTPStatusInvalid                        = errors.New("HTTPStatus should be in [200,600)")
	ErrOutlierDetectionBaseEjectionTimeInvalid  = errors.New("field OutlierDetection.BaseEjectionTime must be specified")
	ErrOutlierDetectionIntervalInvalid          = errors.New("field OutlierDetection.Interval must be specified")
	ErrBothXForwardedForAndCustomHeaderInvalid  = errors.New("only one of ClientIPDetection.XForwardedFor and ClientIPDetection.CustomHeader must be set")
	ErrBothNumTrustedHopsAndTrustedCIDRsInvalid = errors.New("only one of ClientIPDetection.XForwardedFor.NumTrustedHops and ClientIPDetection.XForwardedFor.TrustedCIDRs must be set")
	ErrPanicThresholdInvalid                    = errors.New("PanicThreshold value is outside of 0-100 range")

	redacted = []byte("[redacted]")
)

// PrivateBytes implements a custom []byte type so that we can override the
// various string-ish printing functions to redact the contents.
type PrivateBytes []byte

var (
	_ fmt.Stringer           = PrivateBytes{}
	_ encoding.TextMarshaler = PrivateBytes{}
)

// MarshalText redacts the contents of the PrivateBytes type.
// Note that MarshalJSON will call MarshalText if it exists, so we don't
// need to implement MarshalJSON.
func (p PrivateBytes) MarshalText() ([]byte, error) {
	if len(p) == 0 {
		return nil, nil
	}
	return redacted, nil
}

// String redacts the contents of the PrivateBytes type.
func (p PrivateBytes) String() string {
	if len(p) == 0 {
		return ""
	}
	return string(redacted)
}

func (p *PrivateBytes) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*p = nil
		return nil
	}
	if string(data) == `"`+string(redacted)+`"` {
		*p = redacted
		return nil
	}
	var b []byte
	err := json.Unmarshal(data, &b)
	if err != nil {
		return fmt.Errorf("UnmarshalJSON failed: %w, %q", err, string(data))
	}
	*p = b
	return err
}

// Xds holds the intermediate representation of a Gateway and is
// used by the xDS Translator to convert it into xDS resources.
// +k8s:deepcopy-gen=true
type Xds struct {
	// ReadyListener is the listener that is ready to accept traffic.
	ReadyListener *ReadyListener `json:"readyListener,omitempty" yaml:"readyListener,omitempty"`
	// AccessLog configuration for the gateway.
	AccessLog *AccessLog `json:"accessLog,omitempty" yaml:"accessLog,omitempty"`
	// Tracing configuration for the gateway.
	Tracing *Tracing `json:"tracing,omitempty" yaml:"tracing,omitempty"`
	// Metrics configuration for the gateway.
	Metrics *Metrics `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	// HTTP listeners exposed by the gateway.
	HTTP []*HTTPListener `json:"http,omitempty" yaml:"http,omitempty"`
	// TCP Listeners exposed by the gateway.
	TCP []*TCPListener `json:"tcp,omitempty" yaml:"tcp,omitempty"`
	// UDP Listeners exposed by the gateway.
	UDP []*UDPListener `json:"udp,omitempty" yaml:"udp,omitempty"`
	// EnvoyPatchPolicies is the intermediate representation of the EnvoyPatchPolicy resource
	EnvoyPatchPolicies []*EnvoyPatchPolicy `json:"envoyPatchPolicies,omitempty" yaml:"envoyPatchPolicies,omitempty"`
	// FilterOrder holds the custom order of the HTTP filters
	FilterOrder []egv1a1.FilterPosition `json:"filterOrder,omitempty" yaml:"filterOrder,omitempty"`
}

// Equal implements the Comparable interface used by watchable.DeepEqual to skip unnecessary updates.
func (x *Xds) Equal(y *Xds) bool {
	// Deep copy to avoid modifying the original ordering.
	x = x.DeepCopy()
	x.sort()
	y = y.DeepCopy()
	y.sort()
	return reflect.DeepEqual(x, y)
}

// sort ensures the listeners are in a consistent order.
func (x *Xds) sort() {
	slices.SortFunc(x.HTTP, func(l1, l2 *HTTPListener) int {
		return cmp.Compare(l1.Name, l2.Name)
	})
	for _, l := range x.HTTP {
		slices.SortFunc(l.Routes, func(r1, r2 *HTTPRoute) int {
			return cmp.Compare(r1.Name, r2.Name)
		})
	}
	slices.SortFunc(x.TCP, func(l1, l2 *TCPListener) int {
		return cmp.Compare(l1.Name, l2.Name)
	})
	slices.SortFunc(x.UDP, func(l1, l2 *UDPListener) int {
		return cmp.Compare(l1.Name, l2.Name)
	})
}

// Validate the fields within the Xds structure.
func (x *Xds) Validate() error {
	var errs error
	for _, http := range x.HTTP {
		if err := http.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	for _, tcp := range x.TCP {
		if err := tcp.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	for _, udp := range x.UDP {
		if err := udp.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (x *Xds) GetHTTPListener(name string) *HTTPListener {
	for _, listener := range x.HTTP {
		if listener.Name == name {
			return listener
		}
	}
	return nil
}

func (x *Xds) GetTCPListener(name string) *TCPListener {
	for _, listener := range x.TCP {
		if listener.Name == name {
			return listener
		}
	}
	return nil
}

func (x *Xds) GetUDPListener(name string) *UDPListener {
	for _, listener := range x.UDP {
		if listener.Name == name {
			return listener
		}
	}
	return nil
}

func (x *Xds) YAMLString() string {
	y, _ := yaml.Marshal(x)
	return string(y)
}

func (x *Xds) JSONString() string {
	j, _ := json.Marshal(x)
	return string(j)
}

type Listener interface {
	GetName() string
	GetAddress() string
	GetPort() uint32
	GetExtensionRefs() []*UnstructuredRef
}

// +k8s:deepcopy-gen=true
type CoreListenerDetails struct {
	// Name of the HttpListener
	Name string `json:"name" yaml:"name"`
	// Address that the listener should listen on.
	Address string `json:"address" yaml:"address"`
	// Port on which the service can be expected to be accessed by clients.
	Port uint32 `json:"port" yaml:"port"`
	// ExtensionRefs holds unstructured resources that were introduced by an extension policy
	ExtensionRefs []*UnstructuredRef `json:"extensionRefs,omitempty" yaml:"extensionRefs,omitempty"`
	// Metadata is used to enrich envoy resource metadata with user and provider-specific information
	Metadata *ResourceMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// IPFamily specifies the IP address family used by the Gateway for its listening ports.
	IPFamily *egv1a1.IPFamily `json:"ipFamily,omitempty" yaml:"ipFamily,omitempty"`
}

func (l CoreListenerDetails) GetName() string {
	return l.Name
}

func (l CoreListenerDetails) GetAddress() string {
	return l.Address
}

func (l CoreListenerDetails) GetPort() uint32 {
	return l.Port
}

func (l CoreListenerDetails) GetExtensionRefs() []*UnstructuredRef {
	return l.ExtensionRefs
}

// HTTPListener holds the listener configuration.
// +k8s:deepcopy-gen=true
type HTTPListener struct {
	CoreListenerDetails `json:",inline" yaml:",inline"`
	// Hostnames (Host/Authority header value) with which the service can be expected to be accessed by clients.
	// This field is required. Wildcard hosts are supported in the suffix or prefix form.
	// Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-virtualhost
	// for more info.
	Hostnames []string `json:"hostnames" yaml:"hostnames"`
	// Tls configuration. If omitted, the gateway will expose a plain text HTTP server.
	TLS *TLSConfig `json:"tls,omitempty" yaml:"tls,omitempty"`
	// Routes associated with HTTP traffic to the service.
	Routes []*HTTPRoute `json:"routes,omitempty" yaml:"routes,omitempty"`
	// IsHTTP2 is set if the listener is configured to serve HTTP2 traffic,
	// grpc-web and grpc-stats are also enabled if this is set.
	IsHTTP2 bool `json:"isHTTP2" yaml:"isHTTP2"`
	// TCPKeepalive configuration for the listener
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty" yaml:"tcpKeepalive,omitempty"`
	// Headers configures special header management for the listener
	Headers *HeaderSettings `json:"headers,omitempty" yaml:"headers,omitempty"`
	// EnableProxyProtocol enables the listener to interpret proxy protocol header
	EnableProxyProtocol bool `json:"enableProxyProtocol,omitempty" yaml:"enableProxyProtocol,omitempty"`
	// ClientIPDetection controls how the original client IP address is determined for requests.
	ClientIPDetection *ClientIPDetectionSettings `json:"clientIPDetection,omitempty" yaml:"clientIPDetection,omitempty"`
	// Path contains settings for path URI manipulations
	Path PathSettings `json:"path,omitempty"`
	// HTTP1 provides HTTP/1 configuration on the listener
	// +optional
	HTTP1 *HTTP1Settings `json:"http1,omitempty" yaml:"http1,omitempty"`
	// HTTP2 provides HTTP/2 configuration on the listener
	// +optional
	HTTP2 *HTTP2Settings `json:"http2,omitempty" yaml:"http2,omitempty"`
	// HTTP3 provides HTTP/3 configuration on the listener.
	// +optional
	HTTP3 *HTTP3Settings `json:"http3,omitempty"`
	// HealthCheck provides configuration for determining whether the HTTP/HTTPS listener is healthy.
	HealthCheck *HealthCheckSettings `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	// ClientTimeout sets the timeout configuration for downstream connections
	Timeout *ClientTimeout `json:"timeout,omitempty" yaml:"clientTimeout,omitempty"`
	// Connection settings
	Connection *ClientConnection `json:"connection,omitempty" yaml:"connection,omitempty"`
	// PreserveRouteOrder determines if routes should be sorted according to GW-API specs
	PreserveRouteOrder bool `json:"preserveRouteOrder,omitempty" yaml:"preserveRouteOrder,omitempty"`
}

// Validate the fields within the HTTPListener structure
func (h HTTPListener) Validate() error {
	var errs error
	if h.Name == "" {
		errs = errors.Join(errs, ErrListenerNameEmpty)
	}
	if _, err := netip.ParseAddr(h.Address); err != nil {
		errs = errors.Join(errs, ErrListenerAddressInvalid)
	}
	if h.Port == 0 {
		errs = errors.Join(errs, ErrListenerPortInvalid)
	}
	if len(h.Hostnames) == 0 {
		errs = errors.Join(errs, ErrHTTPListenerHostnamesEmpty)
	}
	if h.TLS != nil {
		if err := h.TLS.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	for _, route := range h.Routes {
		if err := route.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if h.ClientIPDetection != nil {
		if h.ClientIPDetection.XForwardedFor != nil && h.ClientIPDetection.CustomHeader != nil {
			errs = errors.Join(errs, ErrBothXForwardedForAndCustomHeaderInvalid)
		} else if h.ClientIPDetection.XForwardedFor != nil {
			if h.ClientIPDetection.XForwardedFor.NumTrustedHops != nil && h.ClientIPDetection.XForwardedFor.TrustedCIDRs != nil {
				errs = errors.Join(errs, ErrBothNumTrustedHopsAndTrustedCIDRsInvalid)
			}
		}
	}
	return errs
}

type TLSVersion egv1a1.TLSVersion

const (
	// TLSAuto allows Envoy to choose the optimal TLS Version
	TLSAuto = TLSVersion(egv1a1.TLSAuto)
	// TLSv10 specifies TLS version 1.0
	TLSv10 = TLSVersion(egv1a1.TLSv10)
	// TLSv11 specifies TLS version 1.1
	TLSv11 = TLSVersion(egv1a1.TLSv11)
	// TLSv12 specifies TLS version 1.2
	TLSv12 = TLSVersion(egv1a1.TLSv12)
	// TLSv13 specifies TLS version 1.3
	TLSv13 = TLSVersion(egv1a1.TLSv13)
)

func (t TLSVersion) Int() uint16 {
	switch t {
	case TLSAuto:
		return tls.VersionTLS13
	case TLSv10:
		return tls.VersionTLS10
	case TLSv11:
		return tls.VersionTLS11
	case TLSv12:
		return tls.VersionTLS12
	case TLSv13:
		return tls.VersionTLS13
	default:
		return tls.VersionTLS13
	}
}

// TLSConfig holds the configuration for downstream TLS context.
// +k8s:deepcopy-gen=true
type TLSConfig struct {
	// Certificates contains the set of certificates associated with this listener
	Certificates []TLSCertificate `json:"certificates,omitempty" yaml:"certificates,omitempty"`
	// ClientCertificates contains the set of certificates used for mtls as client side certificates
	ClientCertificates []TLSCertificate `json:"clientCertificates,omitempty" yaml:"clientCertificates,omitempty"`
	// CACertificate to verify the client
	CACertificate *TLSCACertificate `json:"caCertificate,omitempty" yaml:"caCertificate,omitempty"`
	// RequireClientCertificate to enforce client certificate
	RequireClientCertificate bool `json:"requireClientCertificate,omitempty" yaml:"requireClientCertificate,omitempty"`
	// MinVersion defines the minimal version of the TLS protocol supported by this listener.
	MinVersion *TLSVersion `json:"minVersion,omitempty" yaml:"version,omitempty"`
	// MaxVersion defines the maximal version of the TLS protocol supported by this listener.
	MaxVersion *TLSVersion `json:"maxVersion,omitempty" yaml:"version,omitempty"`
	// CipherSuites supported by this listener
	Ciphers []string `json:"ciphers,omitempty" yaml:"ciphers,omitempty"`
	// EDCHCurves supported by this listener
	ECDHCurves []string `json:"ecdhCurves,omitempty" yaml:"ecdhCurves,omitempty"`
	// SignatureAlgorithms supported by this listener
	SignatureAlgorithms []string `json:"signatureAlgorithms,omitempty" yaml:"signatureAlgorithms,omitempty"`
	// ALPNProtocols exposed by this listener
	ALPNProtocols []string `json:"alpnProtocols" yaml:"alpnProtocols"`
	// StatelessSessionResumption determines if stateless (session-ticket based) session resumption is enabled
	StatelessSessionResumption bool `json:"statelessSessionResumption,omitempty" yaml:"statelessSessionResumption,omitempty"`
	// StatefulSessionResumption determines if stateful (session-id based) session resumption is enabled
	StatefulSessionResumption bool `json:"statefulSessionResumption,omitempty" yaml:"statefulSessionResumption,omitempty"`
}

// TLSCertificate holds a single certificate's details
// +k8s:deepcopy-gen=true
type TLSCertificate struct {
	// Name of the Secret object.
	Name string `json:"name" yaml:"name"`
	// Certificate can be either a client or server certificate.
	Certificate []byte `json:"serverCertificate,omitempty" yaml:"serverCertificate,omitempty"`
	// PrivateKey for the server.
	PrivateKey PrivateBytes `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`
}

// TLSCACertificate holds CA Certificate to validate clients
// +k8s:deepcopy-gen=true
type TLSCACertificate struct {
	// Name of the Secret object.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Certificate content.
	Certificate []byte `json:"certificate,omitempty" yaml:"certificate,omitempty"`
}

func (t TLSCertificate) Validate() error {
	var errs error
	if len(t.Certificate) == 0 {
		errs = errors.Join(errs, ErrTLSServerCertEmpty)
	}
	if len(t.PrivateKey) == 0 {
		errs = errors.Join(errs, ErrTLSPrivateKey)
	}
	return errs
}

// Validate the fields within the TLSListenerConfig structure
func (t TLSConfig) Validate() error {
	var errs error
	for _, cert := range t.Certificates {
		if err := cert.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	// Correct values for cipher suites, ECDH curves, and signature algorithms are
	// dependent on the version of EnvoyProxy being used - different values are valid
	// depending if Envoy was compiled against BoringSSL or OpenSSL, or even the exact version
	// of each of these libraries.
	// Validation for TLS versions was done with CEL
	return errs
}

type PathEscapedSlashAction egv1a1.PathEscapedSlashAction

const (
	KeepUnchangedAction = PathEscapedSlashAction(egv1a1.KeepUnchangedAction)
	RejectRequestAction = PathEscapedSlashAction(egv1a1.RejectRequestAction)
	UnescapeAndRedirect = PathEscapedSlashAction(egv1a1.UnescapeAndRedirect)
	UnescapeAndForward  = PathEscapedSlashAction(egv1a1.UnescapeAndForward)
)

// PathSettings holds configuration for path URI manipulations
// +k8s:deepcopy-gen=true
type PathSettings struct {
	MergeSlashes         bool                   `json:"mergeSlashes" yaml:"mergeSlashes"`
	EscapedSlashesAction PathEscapedSlashAction `json:"escapedSlashesAction" yaml:"escapedSlashesAction"`
}

type WithUnderscoresAction egv1a1.WithUnderscoresAction

const (
	WithUnderscoresActionAllow         = WithUnderscoresAction(egv1a1.WithUnderscoresActionAllow)
	WithUnderscoresActionRejectRequest = WithUnderscoresAction(egv1a1.WithUnderscoresActionRejectRequest)
	WithUnderscoresActionDropHeader    = WithUnderscoresAction(egv1a1.WithUnderscoresActionDropHeader)
)

type RequestIDAction egv1a1.RequestIDAction

const (
	RequestIDActionPreserveOrGenerate = RequestIDAction(egv1a1.RequestIDActionPreserveOrGenerate)
	RequestIDActionPreserve           = RequestIDAction(egv1a1.RequestIDActionPreserve)
	RequestIDActionGenerate           = RequestIDAction(egv1a1.RequestIDActionGenerate)
	RequestIDActionDisable            = RequestIDAction(egv1a1.RequestIDActionDisable)
)

// Configure Envoy proxy how to handle the x-forwarded-client-cert (XFCC) HTTP header.
// +k8s:deepcopy-gen=true
type XForwardedClientCert struct {
	// Envoy Proxy mode how to handle the x-forwarded-client-cert (XFCC) HTTP header.
	Mode egv1a1.XFCCForwardMode `json:"mode,omitempty" yaml:"mode,omitempty"`
	// Specifies the fields in the client certificate to be forwarded on the x-forwarded-client-cert (XFCC) HTTP header
	CertDetailsToAdd []egv1a1.XFCCCertData `json:"certDetailsToAdd,omitempty" yaml:"certDetailsToAdd,omitempty"`
}

// ClientIPDetectionSettings provides configuration for determining the original client IP address for requests.
// +k8s:deepcopy-gen=true
type ClientIPDetectionSettings egv1a1.ClientIPDetectionSettings

// BackendWeights stores the weights of valid and invalid backends for the route so that 500 error responses can be returned in the same proportions
type BackendWeights struct {
	Name    string `json:"name" yaml:"name"`
	Valid   uint32 `json:"valid" yaml:"valid"`
	Invalid uint32 `json:"invalid" yaml:"invalid"`
}

// HTTP1Settings provides HTTP/1 configuration on the listener.
// +k8s:deepcopy-gen=true
type HTTP1Settings struct {
	EnableTrailers     bool            `json:"enableTrailers,omitempty" yaml:"enableTrailers,omitempty"`
	PreserveHeaderCase bool            `json:"preserveHeaderCase,omitempty" yaml:"preserveHeaderCase,omitempty"`
	HTTP10             *HTTP10Settings `json:"http10,omitempty" yaml:"http10,omitempty"`
}

// HTTP10Settings provides HTTP/1.0 configuration on the listener.
// +k8s:deepcopy-gen=true
type HTTP10Settings struct {
	// defaultHost is set to the default host that should be injected for HTTP10. If the hostname shouldn't
	// be set, then defaultHost will be nil
	DefaultHost *string `json:"defaultHost,omitempty" yaml:"defaultHost,omitempty"`
}

// HTTP2Settings provides HTTP/2 configuration on the listener.
// +k8s:deepcopy-gen=true
type HTTP2Settings struct {
	// InitialStreamWindowSize is the initial window size for a stream.
	InitialStreamWindowSize *uint32 `json:"initialConnectionWindowSize,omitempty" yaml:"initialConnectionWindowSize,omitempty"`
	// InitialConnectionWindowSize is the initial window size for a connection.
	InitialConnectionWindowSize *uint32 `json:"initialStreamWindowSize,omitempty" yaml:"initialStreamWindowSize,omitempty"`
	// MaxConcurrentStreams is the maximum number of concurrent streams that can be opened on a connection.
	MaxConcurrentStreams *uint32 `json:"maxConcurrentStreams,omitempty" yaml:"maxConcurrentStreams,omitempty"`
	// ResetStreamOnError determines if a stream or connection is reset on messaging error.
	ResetStreamOnError *bool `json:"resetStreamOnError,omitempty" yaml:"resetStreamOnError,omitempty"`
}

// ResponseOverride defines the configuration to override specific responses with a custom one.
// +k8s:deepcopy-gen=true
type ResponseOverride struct {
	// Name is a unique name for a ResponseOverride configuration.
	// The xds translator only generates one CustomResponse filter for each unique name.
	Name string `json:"name" yaml:"name"`

	// Rules contains the list of rules to override responses.
	Rules []ResponseOverrideRule `json:"rules,omitempty"`
}

// ResponseOverrideRule defines the configuration for overriding a response.
// +k8s:deepcopy-gen=true
type ResponseOverrideRule struct {
	// Name is a generated name for the rule.
	Name string `json:"name"`
	// Match configuration.
	Match CustomResponseMatch `json:"match"`
	// Response configuration.
	Response CustomResponse `json:"response"`
}

// CustomResponseMatch defines the configuration for matching a user response to return a custom one.
// +k8s:deepcopy-gen=true
type CustomResponseMatch struct {
	// Status code to match on. The match evaluates to true if any of the matches are successful.
	StatusCodes []StatusCodeMatch `json:"statusCodes"`
}

// StatusCodeMatch defines the configuration for matching a status code.
// +k8s:deepcopy-gen=true
type StatusCodeMatch struct {
	// Value contains the value of the status code.
	Value *int `json:"value,omitempty"`

	// Range contains a range of status codes.
	Range *StatusCodeRange `json:"range,omitempty"`
}

// StatusCodeRange defines the configuration for define a range of status codes.
type StatusCodeRange struct {
	// Start of the range, including the start value.
	Start int `json:"start"`
	// End of the range, including the end value.
	End int `json:"end"`
}

// CustomResponse defines the configuration for returning a custom response.
// +k8s:deepcopy-gen=true
type CustomResponse struct {
	// Content Type of the response. This will be set in the Content-Type header.
	ContentType *string `json:"contentType,omitempty"`

	// Body of the Custom Response
	Body *string `json:"body,omitempty"`

	// StatusCode will be used for the response's status code.
	StatusCode *uint32 `json:"statusCode,omitempty"`
}

// Validate the fields within the CustomResponse structure
func (r *CustomResponse) Validate() error {
	var errs error
	if status := r.StatusCode; status != nil && (*status > 599 || *status < 100) {
		errs = errors.Join(errs, ErrDirectResponseStatusInvalid)
	}

	return errs
}

// HealthCheckSettings provides HealthCheck configuration on the HTTP/HTTPS listener.
// +k8s:deepcopy-gen=true
type HealthCheckSettings egv1a1.HealthCheckSettings

// HeaderSettings provides configuration related to header processing on the listener.
// +k8s:deepcopy-gen=true
type HeaderSettings struct {
	// EnableEnvoyHeaders controls if "x-envoy-" headers are added by the HTTP Router filter.
	// The default is to suppress these headers.
	// Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/router/v3/router.proto#extensions-filters-http-router-v3-router
	EnableEnvoyHeaders bool `json:"enableEnvoyHeaders,omitempty" yaml:"enableEnvoyHeaders,omitempty"`

	// DisableRateLimitHeaders controls if "x-ratelimit-" headers are added by the HTTP Router filter.
	// The default is to emit these headers.
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ratelimit/v3/rate_limit.proto#extensions-filters-http-ratelimit-v3-ratelimit
	DisableRateLimitHeaders bool `json:"disableRateLimitHeaders,omitempty" yaml:"disableRateLimitHeaders,omitempty"`

	// Configure Envoy proxy how to handle the x-forwarded-client-cert (XFCC) HTTP header.
	// refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/http_connection_manager/v3/http_connection_manager.proto#envoy-v3-api-enum-extensions-filters-network-http-connection-manager-v3-httpconnectionmanager-forwardclientcertdetails
	XForwardedClientCert *XForwardedClientCert `json:"xForwardedClientCert,omitempty" yaml:"xForwardedClientCert,omitempty"`

	// WithUnderscoresAction configures the action to take when an HTTP header with underscores
	// is encountered. The default action is to reject the request.
	// Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-enum-config-core-v3-httpprotocoloptions-headerswithunderscoresaction
	WithUnderscoresAction WithUnderscoresAction `json:"withUnderscoresAction,omitempty" yaml:"withUnderscoresAction,omitempty"`

	// RequestID configures Envoy's behavior for handling the `X-Request-ID` header.
	// Defaults to `Generate` and builds the `X-Request-ID` for every request and ignores pre-existing values from the edge.
	// (An "edge request" refers to a request from an external client to the Envoy entrypoint.)
	RequestID *RequestIDAction `json:"requestID,omitempty" yaml:"requestID,omitempty"`

	// EarlyAddRequestHeaders defines headers that would be added before envoy request processing.
	EarlyAddRequestHeaders []AddHeader `json:"earlyAddRequestHeaders,omitempty" yaml:"earlyAddRequestHeaders,omitempty"`

	// EarlyRemoveRequestHeaders defines headers that would be removed before envoy request processing.
	EarlyRemoveRequestHeaders []string `json:"earlyRemoveRequestHeaders,omitempty" yaml:"earlyRemoveRequestHeaders,omitempty"`
}

// ClientTimeout sets the timeout configuration for downstream connections
// +k8s:deepcopy-gen=true
type ClientTimeout struct {
	// Timeout settings for TCP (not HTTP).
	TCP *TCPClientTimeout `json:"tcp,omitempty" yaml:"tcp,omitempty"`

	// Timeout settings for HTTP.
	HTTP *HTTPClientTimeout `json:"http,omitempty" yaml:"http,omitempty"`
}

// TCPClientTimeout set the configuration for client TCP (not HTTP).
// +k8s:deepcopy-gen=true
type TCPClientTimeout struct {
	// IdleTimeout for a TCP connection. Idle time is defined as a period in which there are no
	// bytes sent or received on either the upstream or downstream connection.
	IdleTimeout *metav1.Duration `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
}

// HTTPClientTimeout set the configuration for client HTTP.
// +k8s:deepcopy-gen=true
type HTTPClientTimeout struct {
	// The duration envoy waits for the complete request reception. This timer starts upon request
	// initiation and stops when either the last byte of the request is sent upstream or when the response begins.
	RequestReceivedTimeout *metav1.Duration `json:"requestReceivedTimeout,omitempty" yaml:"requestReceivedTimeout,omitempty"`
	// IdleTimeout for an HTTP connection. Idle time is defined as a period in which there are no active requests in the connection.
	IdleTimeout *metav1.Duration `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
}

// HTTPRoute holds the route information associated with the HTTP Route
// +k8s:deepcopy-gen=true
type HTTPRoute struct {
	// Name of the HTTPRoute
	Name string `json:"name" yaml:"name"`
	// Hostname that the route matches against
	Hostname string `json:"hostname" yaml:"hostname,omitempty"`
	// IsHTTP2 is set if the route is configured to serve HTTP2 traffic
	IsHTTP2 bool `json:"isHTTP2" yaml:"isHTTP2"`
	// PathMatch defines the match conditions on the path.
	PathMatch *StringMatch `json:"pathMatch,omitempty" yaml:"pathMatch,omitempty"`
	// HeaderMatches define the match conditions on the request headers for this route.
	HeaderMatches []*StringMatch `json:"headerMatches,omitempty" yaml:"headerMatches,omitempty"`
	// QueryParamMatches define the match conditions on the query parameters.
	QueryParamMatches []*StringMatch `json:"queryParamMatches,omitempty" yaml:"queryParamMatches,omitempty"`
	// AddRequestHeaders defines header/value sets to be added to the headers of requests.
	AddRequestHeaders []AddHeader `json:"addRequestHeaders,omitempty" yaml:"addRequestHeaders,omitempty"`
	// RemoveRequestHeaders defines a list of headers to be removed from requests.
	RemoveRequestHeaders []string `json:"removeRequestHeaders,omitempty" yaml:"removeRequestHeaders,omitempty"`
	// AddResponseHeaders defines header/value sets to be added to the headers of response.
	AddResponseHeaders []AddHeader `json:"addResponseHeaders,omitempty" yaml:"addResponseHeaders,omitempty"`
	// RemoveResponseHeaders defines a list of headers to be removed from response.
	RemoveResponseHeaders []string `json:"removeResponseHeaders,omitempty" yaml:"removeResponseHeaders,omitempty"`
	// Direct responses to be returned for this route. Takes precedence over Destinations and Redirect.
	DirectResponse *CustomResponse `json:"directResponse,omitempty" yaml:"directResponse,omitempty"`
	// Redirections to be returned for this route. Takes precedence over Destinations.
	Redirect *Redirect `json:"redirect,omitempty" yaml:"redirect,omitempty"`
	// Destination that requests to this HTTPRoute will be mirrored to
	Mirrors []*MirrorPolicy `json:"mirrors,omitempty" yaml:"mirrors,omitempty"`
	// Destination associated with this matched route.
	Destination *RouteDestination `json:"destination,omitempty" yaml:"destination,omitempty"`
	// Rewrite to be changed for this route.
	URLRewrite *URLRewrite `json:"urlRewrite,omitempty" yaml:"urlRewrite,omitempty"`
	// ExtensionRefs holds unstructured resources that were introduced by an extension and used on the HTTPRoute as extensionRef filters
	ExtensionRefs []*UnstructuredRef `json:"extensionRefs,omitempty" yaml:"extensionRefs,omitempty"`
	// Traffic holds the features associated with BackendTrafficPolicy
	Traffic *TrafficFeatures `json:"traffic,omitempty" yaml:"traffic,omitempty"`
	// Security holds the features associated with SecurityPolicy
	Security *SecurityFeatures `json:"security,omitempty" yaml:"security,omitempty"`
	// EnvoyExtension holds the features associated with EnvoyExtensionPolicy
	EnvoyExtensions *EnvoyExtensionFeatures `json:"envoyExtensions,omitempty" yaml:"envoyExtensions,omitempty"`
	// UseClientProtocol enables using the same protocol upstream that was used downstream
	UseClientProtocol *bool `json:"useClientProtocol,omitempty" yaml:"useClientProtocol,omitempty"`
	// Metadata is used to enrich envoy route metadata with user and provider-specific information
	Metadata *ResourceMetadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	// SessionPersistence holds the configuration for session persistence.
	SessionPersistence *SessionPersistence `json:"sessionPersistence,omitempty" yaml:"sessionPersistence,omitempty"`
	// Timeout is the time until which entire response is received from the upstream.
	Timeout *metav1.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Retry defines the retry policy for the route.
	// This is derived from the core Gateway API, and should take precedence over Traffic.Retry.
	Retry *Retry `json:"retry,omitempty" yaml:"retry,omitempty"`
}

func (h *HTTPRoute) GetRetry() *Retry {
	if h.Retry != nil {
		return h.Retry
	}

	if h.Traffic != nil {
		return h.Traffic.Retry
	}

	return nil
}

// DNS contains configuration options for DNS resolution.
// +k8s:deepcopy-gen=true
type DNS struct {
	// DNSRefreshRate specifies the rate at which DNS records should be refreshed.
	DNSRefreshRate *metav1.Duration `json:"dnsRefreshRate,omitempty"`
	// RespectDNSTTL indicates whether the DNS Time-To-Live (TTL) should be respected.
	RespectDNSTTL *bool `json:"respectDnsTtl,omitempty"`
	// LookupFamily allows to configure the dns lookup policy
	LookupFamily *egv1a1.DNSLookupFamily `json:"lookupFamily,omitempty"`
}

// SessionPersistence defines the desired state of SessionPersistence.
// +k8s:deepcopy-gen=true
type SessionPersistence struct {
	// Cookie defines the configuration for cookie-based session persistence.
	// Either Cookie or Header must be non-empty.
	Cookie *CookieBasedSessionPersistence `json:"cookie,omitempty" yaml:"cookie,omitempty"`
	// Header defines the configuration for header-based session persistence.
	// Either Cookie or Header must be non-empty.
	Header *HeaderBasedSessionPersistence `json:"header,omitempty" yaml:"header,omitempty"`
}

// CookieBasedSessionPersistence defines the configuration for cookie-based session persistence.
// +k8s:deepcopy-gen=true
type CookieBasedSessionPersistence struct {
	// Name defines the name of the persistent session token.
	Name string `json:"name"`

	TTL *metav1.Duration `json:"ttl,omitempty" yaml:"ttl,omitempty"`
}

// HeaderBasedSessionPersistence defines the configuration for header-based session persistence.
// +k8s:deepcopy-gen=true
type HeaderBasedSessionPersistence struct {
	// Name defines the name of the persistent session token.
	Name string `json:"name"`
}

// Compression holds the configuration for HTTP compression.
// +k8s:deepcopy-gen=true
type Compression struct {
	// Type of compression to be used.
	Type egv1a1.CompressorType `json:"type" yaml:"type"`
}

// TrafficFeatures holds the information associated with the Backend Traffic Policy.
// +k8s:deepcopy-gen=true
type TrafficFeatures struct {
	// RateLimit defines the more specific match conditions as well as limits for ratelimiting
	// the requests on this route.
	RateLimit *RateLimit `json:"rateLimit,omitempty" yaml:"rateLimit,omitempty"`
	// load balancer policy to use when routing to the backend endpoints.
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty"`
	// Proxy Protocol Settings
	ProxyProtocol *ProxyProtocol `json:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty"`
	// HealthCheck defines the configuration for health checking on the upstream.
	HealthCheck *HealthCheck `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	// FaultInjection defines the schema for injecting faults into HTTP requests.
	FaultInjection *FaultInjection `json:"faultInjection,omitempty" yaml:"faultInjection,omitempty"`
	// Circuit Breaker Settings
	CircuitBreaker *CircuitBreaker `json:"circuitBreaker,omitempty" yaml:"circuitBreaker,omitempty"`
	// Request and connection timeout settings
	Timeout *Timeout `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// TcpKeepalive settings associated with the upstream client connection.
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty" yaml:"tcpKeepalive,omitempty"`
	// Retry settings
	Retry *Retry `json:"retry,omitempty" yaml:"retry,omitempty"`
	// settings of upstream connection
	BackendConnection *BackendConnection `json:"backendConnection,omitempty" yaml:"backendConnection,omitempty"`
	// HTTP2 provides HTTP/2 configuration for clusters
	// +optional
	HTTP2 *HTTP2Settings `json:"http2,omitempty" yaml:"http2,omitempty"`
	// DNS is used to configure how DNS resolution is handled by the Envoy Proxy cluster
	DNS *DNS `json:"dns,omitempty" yaml:"dns,omitempty"`
	// ResponseOverride defines the schema for overriding the response.
	ResponseOverride *ResponseOverride `json:"responseOverride,omitempty" yaml:"responseOverride,omitempty"`
	// Compression settings for HTTP Response
	Compression []*Compression `json:"compression,omitempty" yaml:"compression,omitempty"`
	// HTTPUpgrade defines the schema for upgrading the HTTP protocol.
	HTTPUpgrade []string `json:"httpUpgrade,omitempty" yaml:"httpUpgrade,omitempty"`
}

func (b *TrafficFeatures) Validate() error {
	var errs error

	if b.LoadBalancer != nil {
		if err := b.LoadBalancer.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if b.HealthCheck != nil {
		if err := b.HealthCheck.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// SecurityFeatures holds the information associated with the Security Policy.
// +k8s:deepcopy-gen=true
type SecurityFeatures struct {
	// CORS policy for the route.
	CORS *CORS `json:"cors,omitempty" yaml:"cors,omitempty"`
	// JWT defines the schema for authenticating HTTP requests using JSON Web Tokens (JWT).
	JWT *JWT `json:"jwt,omitempty" yaml:"jwt,omitempty"`
	// OIDC defines the schema for authenticating HTTP requests using OpenID Connect (OIDC).
	OIDC *OIDC `json:"oidc,omitempty" yaml:"oidc,omitempty"`
	// APIKeyAuth defines the schema for the API Key Authentication.
	APIKeyAuth *APIKeyAuth `json:"apiKeyAuth,omitempty" yaml:"apiKeyAuth,omitempty"`
	// BasicAuth defines the schema for the HTTP Basic Authentication.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
	// ExtAuth defines the schema for the external authorization.
	ExtAuth *ExtAuth `json:"extAuth,omitempty" yaml:"extAuth,omitempty"`
	// Authorization defines the schema for the authorization.
	Authorization *Authorization `json:"authorization,omitempty" yaml:"authorization,omitempty"`
}

// EnvoyExtensionFeatures holds the information associated with the Envoy Extension Policy.
// +k8s:deepcopy-gen=true
type EnvoyExtensionFeatures struct {
	// External Processing extensions
	ExtProcs []ExtProc `json:"extProcs,omitempty" yaml:"extProcs,omitempty"`
	// Wasm extensions
	Wasms []Wasm `json:"wasms,omitempty" yaml:"wasms,omitempty"`
	// Lua extensions
	Luas []Lua `json:"luas,omitempty" yaml:"luas,omitempty"`
}

// UnstructuredRef holds unstructured data for an arbitrary k8s resource introduced by an extension
// Envoy Gateway does not need to know about the resource types in order to store and pass the data for these objects
// to an extension.
//
// +k8s:deepcopy-gen=true
type UnstructuredRef struct {
	Object *unstructured.Unstructured `json:"object,omitempty" yaml:"object,omitempty"`
}

// RegexMatchReplace defines the schema for modifying HTTP request path using regex.
//
// +k8s:deepcopy-gen=true
type RegexMatchReplace struct {
	// Pattern matches a regular expression against the value of the HTTP Path.The regex string must
	// adhere to the syntax documented in https://github.com/google/re2/wiki/Syntax.
	Pattern string `json:"pattern" yaml:"pattern"`
	// Substitution is an expression that replaces the matched portion.The expression may include numbered
	// capture groups that adhere to syntax documented in https://github.com/google/re2/wiki/Syntax.
	Substitution string `json:"substitution" yaml:"substitution"`
}

// CORS holds the Cross-Origin Resource Sharing (CORS) policy for the route.
//
// +k8s:deepcopy-gen=true
type CORS struct {
	// AllowOrigins defines the origins that are allowed to make requests.
	AllowOrigins []*StringMatch `json:"allowOrigins,omitempty" yaml:"allowOrigins,omitempty"`
	// AllowMethods defines the methods that are allowed to make requests.
	AllowMethods []string `json:"allowMethods,omitempty" yaml:"allowMethods,omitempty"`
	// AllowHeaders defines the headers that are allowed to be sent with requests.
	AllowHeaders []string `json:"allowHeaders,omitempty" yaml:"allowHeaders,omitempty"`
	// ExposeHeaders defines the headers that can be exposed in the responses.
	ExposeHeaders []string `json:"exposeHeaders,omitempty" yaml:"exposeHeaders,omitempty"`
	// MaxAge defines how long the results of a preflight request can be cached.
	MaxAge *metav1.Duration `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
	// AllowCredentials indicates whether a request can include user credentials.
	AllowCredentials bool `json:"allowCredentials,omitempty" yaml:"allowCredentials,omitempty"`
}

// JWT defines the schema for authenticating HTTP requests using
// JSON Web Tokens (JWT).
//
// +k8s:deepcopy-gen=true
type JWT struct {
	// AllowMissing determines whether a missing JWT is acceptable.
	AllowMissing bool `json:"allowMissing,omitempty" yaml:"allowMissing,omitempty"`

	// Providers defines a list of JSON Web Token (JWT) authentication providers.
	Providers []JWTProvider `json:"providers,omitempty" yaml:"providers,omitempty"`
}

// JWTProvider defines the schema for the JWT Provider.
//
// +k8s:deepcopy-gen=true
type JWTProvider struct {
	// Name defines a unique name for the JWT provider. A name can have a variety of forms,
	// including RFC1123 subdomains, RFC 1123 labels, or RFC 1035 labels.
	Name string `json:"name"`

	// Issuer is the principal that issued the JWT and takes the form of a URL or email address.
	Issuer string `json:"issuer,omitempty"`

	// Audiences is a list of JWT audiences allowed access. For additional details, see
	// https://tools.ietf.org/html/rfc7519#section-4.1.3. If not provided, JWT audiences
	// are not checked.
	Audiences []string `json:"audiences,omitempty"`

	// RemoteJWKS defines how to fetch and cache JSON Web Key Sets (JWKS) from a remote
	// HTTP/HTTPS endpoint.
	RemoteJWKS RemoteJWKS `json:"remoteJWKS"`

	// ClaimToHeaders is a list of JWT claims that must be extracted into HTTP request headers
	// For examples, following config:
	// The claim must be of type; string, int, double, bool. Array type claims are not supported
	ClaimToHeaders []egv1a1.ClaimToHeader `json:"claimToHeaders,omitempty"`

	// RecomputeRoute clears the route cache and recalculates the routing decision.
	// This field must be enabled if the headers generated from the claim are used for
	// route matching decisions. If the recomputation selects a new route, features targeting
	// the new matched route will be applied.
	RecomputeRoute *bool `json:"recomputeRoute,omitempty"`

	// ExtractFrom defines different ways to extract the JWT token from HTTP request.
	// If empty, it defaults to extract JWT token from the Authorization HTTP request header using Bearer schema
	// or access_token from query parameters.
	ExtractFrom *egv1a1.JWTExtractor `json:"extractFrom,omitempty"`
}

// RemoteJWKSBackend holds the configuration for a remote JWKS backend.
//
// +k8s:deepcopy-gen=true
type RemoteJWKS struct {
	// Destination defines the destination for the OIDC Provider.
	Destination *RouteDestination `json:"destination,omitempty"`

	// Traffic contains configuration for traffic features for the OIDC Provider
	Traffic *TrafficFeatures `json:"traffic,omitempty"`

	// URI is the HTTPS URI to fetch the JWKS. Envoy's system trust bundle is used to validate the server certificate.
	// If a custom trust bundle is needed, it can be specified in a BackendTLSConfig resource and target the BackendRefs.
	URI string `json:"uri"`
}

// OIDC defines the schema for authenticating HTTP requests using
// OpenID Connect (OIDC).
//
// +k8s:deepcopy-gen=true
type OIDC struct {
	// Name is a unique name for an OIDC configuration.
	// The xds translator only generates one OAuth2 filter for each unique name.
	Name string `json:"name" yaml:"name"`

	// The OIDC Provider configuration.
	Provider OIDCProvider `json:"provider" yaml:"provider"`

	// The OIDC client ID to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	ClientID string `json:"clientID" yaml:"clientID"`

	// The OIDC client secret to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// This is an Opaque secret. The client secret should be stored in the key "client-secret".
	ClientSecret PrivateBytes `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`

	// HMACSecret is the secret used to sign the HMAC of the OAuth2 filter cookies.
	HMACSecret PrivateBytes `json:"hmacSecret,omitempty" yaml:"hmacSecret,omitempty"`

	// The OIDC scopes to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	Scopes []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`

	// The OIDC resources to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	Resources []string `json:"resources,omitempty" yaml:"resources,omitempty"`

	// The redirect URL to be used in the OIDC
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	RedirectURL string `json:"redirectURL,omitempty"`

	// The path part of the redirect URL
	RedirectPath string `json:"redirectPath,omitempty"`

	// The path to log a user out, clearing their credential cookies.
	LogoutPath string `json:"logoutPath,omitempty"`

	// ForwardAccessToken indicates whether the Envoy should forward the access token
	// via the Authorization header Bearer scheme to the upstream.
	ForwardAccessToken bool `json:"forwardAccessToken,omitempty"`

	// DefaultTokenTTL is the default lifetime of the id token and access token.
	DefaultTokenTTL *metav1.Duration `json:"defaultTokenTTL,omitempty"`

	// RefreshToken indicates whether the Envoy should automatically refresh the
	// id token and access token when they expire.
	RefreshToken bool `json:"refreshToken,omitempty"`

	// DefaultRefreshTokenTTL is the default lifetime of the refresh token.
	DefaultRefreshTokenTTL *metav1.Duration `json:"defaultRefreshTokenTTL,omitempty"`

	// CookieSuffix will be added to the name of the cookies set by the oauth filter.
	// Adding a suffix avoids multiple oauth filters from overwriting each other's cookies.
	// These cookies are set by the oauth filter, including: AccessToken,
	// OauthHMAC, OauthExpires, IdToken, and RefreshToken.
	CookieSuffix string `json:"cookieSuffix,omitempty"`

	// CookieNameOverrides can optionally override the generated name of the cookies set by the oauth filter.
	CookieNameOverrides *egv1a1.OIDCCookieNames `json:"cookieNameOverrides,omitempty"`

	// CookieDomain sets the domain of the cookies set by the oauth filter.
	CookieDomain *string `json:"cookieDomain,omitempty"`
}

// OIDCProvider defines the schema for the OIDC Provider.
//
// +k8s:deepcopy-gen=true
type OIDCProvider struct {
	// Destination defines the destination for the OIDC Provider.
	Destination *RouteDestination `json:"destination,omitempty"`

	// Traffic contains configuration for traffic features for the OIDC Provider
	Traffic *TrafficFeatures `json:"traffic,omitempty"`

	// The OIDC Provider's [authorization endpoint](https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint).
	AuthorizationEndpoint string `json:"authorizationEndpoint,omitempty"`

	// The OIDC Provider's [token endpoint](https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint).
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`
}

// BasicAuth defines the schema for the HTTP Basic Authentication.
//
// +k8s:deepcopy-gen=true
type BasicAuth struct {
	// Name is a unique name for an BasicAuth configuration.
	// The xds translator only generates one basic auth filter for each unique name.
	Name string `json:"name" yaml:"name"`

	// The username-password pairs in htpasswd format.
	Users PrivateBytes `json:"users,omitempty" yaml:"users,omitempty"`

	// This field specifies the header name to forward a successfully authenticated user to
	// the backend. The header will be added to the request with the username as the value.
	//
	// If it is not specified, the username will not be forwarded.
	ForwardUsernameHeader *string `json:"forwardUsernameHeader,omitempty" yaml:"forwardUsernameHeader,omitempty"`
}

// APIKeyAuth defines the schema for the API Key Authentication.
//
// +k8s:deepcopy-gen=true
type APIKeyAuth struct {
	// The API key to be used for authentication.
	// Key is the client id and the value is the API key to be used for authentication.
	Credentials map[string]PrivateBytes `json:"credentials,omitempty" yaml:"credentials,omitempty"`

	// ExtractFrom is where to fetch the key from the coming request.
	// The value from the first source that has a key will be used.
	ExtractFrom []*ExtractFrom `json:"extractFrom"`
}

// ExtractFrom defines the source of the key.
//
// +k8s:deepcopy-gen=true
type ExtractFrom struct {
	// Headers is the names of the header to fetch the key from.
	// If multiple headers are specified, envoy will look for the api key in the order of the list.
	// This field is optional, but only one of headers, params or cookies is supposed to be specified.
	//
	// +optional
	Headers []string `json:"headers,omitempty"`
	// Params is the names of the query parameter to fetch the key from.
	// If multiple params are specified, envoy will look for the api key in the order of the list.
	// This field is optional, but only one of headers, params or cookies is supposed to be specified.
	//
	// +optional
	Params []string `json:"params,omitempty"`
	// Cookies is the names of the cookie to fetch the key from.
	// If multiple cookies are specified, envoy will look for the api key in the order of the list.
	// This field is optional, but only one of headers, params or cookies is supposed to be specified.
	//
	// +optional
	Cookies []string `json:"cookies,omitempty"`
}

// ExtAuth defines the schema for the external authorization.
//
// +k8s:deepcopy-gen=true
type ExtAuth struct {
	// Name is a unique name for an ExtAuth configuration.
	// The xds translator only generates one external authorization filter for each unique name.
	Name string `json:"name" yaml:"name"`

	// GRPC defines the gRPC External Authorization service.
	// Only one of GRPCService or HTTPService may be specified.
	GRPC *GRPCExtAuthService `json:"grpc,omitempty"`

	// HTTP defines the HTTP External Authorization service.
	// Only one of GRPCService or HTTPService may be specified.
	HTTP *HTTPExtAuthService `json:"http,omitempty"`

	// Traffic contains configuration for traffic features for the ExtAuth service
	Traffic *TrafficFeatures `json:"traffic,omitempty"`

	// HeadersToExtAuth defines the client request headers that will be included
	// in the request to the external authorization service.
	// Note: If not specified, the default behavior for gRPC and HTTP external
	// authorization services is different due to backward compatibility reasons.
	// All headers will be included in the check request to a gRPC authorization server.
	// Only the following headers will be included in the check request to an HTTP
	// authorization server: Host, Method, Path, Content-Length, and Authorization.
	// And these headers will always be included to the check request to an HTTP
	// authorization server by default, no matter whether they are specified
	// in HeadersToExtAuth or not.
	// +optional
	HeadersToExtAuth []string `json:"headersToExtAuth,omitempty"`

	// BodyToExtAuth defines the Body to Ext Auth configuration.
	// +optional
	BodyToExtAuth *BodyToExtAuth `json:"bodyToExtAuth,omitempty"`

	// FailOpen is a switch used to control the behavior when a response from the External Authorization service cannot be obtained.
	// If FailOpen is set to true, the system allows the traffic to pass through.
	// Otherwise, if it is set to false or not set (defaulting to false),
	// the system blocks the traffic and returns a HTTP 5xx error, reflecting a fail-closed approach.
	// This setting determines whether to prioritize accessibility over strict security in case of authorization service failure.
	// +optional
	FailOpen *bool `json:"failOpen,omitempty"`

	// RecomputeRoute clears the route cache and recalculates the routing decision.
	// This field must be enabled if the headers generated from the claim are used for
	// route matching decisions. If the recomputation selects a new route, features targeting
	// the new matched route will be applied.
	// +optional
	RecomputeRoute *bool `json:"recomputeRoute,omitempty"`
}

// BodyToExtAuth defines the Body to Ext Auth configuration
// +k8s:deepcopy-gen=true
type BodyToExtAuth struct {
	// MaxRequestBytes is the maximum size of a message body that the filter will hold in memory.
	// Envoy will return HTTP 413 and will not initiate the authorization process when buffer
	// reaches the number set in this field.
	// Note that this setting will have precedence over failOpen mode.
	MaxRequestBytes uint32 `json:"maxRequestBytes"`
}

// HTTPExtAuthService defines the HTTP External Authorization service
// +k8s:deepcopy-gen=true
type HTTPExtAuthService struct {
	// Destination defines the destination for the HTTP External Authorization service.
	Destination RouteDestination `json:"destination"`

	// Authority is the hostname:port of the HTTP External Authorization service.
	Authority string `json:"authority"`

	// Path is the path of the HTTP External Authorization service.
	// If path is not empty, the authorization request will be sent to that path,
	// or else the authorization request will be sent to the root path.
	Path string `json:"path"`

	// HeadersToBackend are the authorization response headers that will be added
	// to the original client request before sending it to the backend server.
	// Note that coexisting headers will be overridden.
	// If not specified, no authorization response headers will be added to the
	// original client request.
	// +optional
	HeadersToBackend []string `json:"headersToBackend,omitempty"`
}

// GRPCExtAuthService defines the gRPC External Authorization service
// The authorization request message is defined in
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto
// +k8s:deepcopy-gen=true
type GRPCExtAuthService struct {
	// Destination defines the destination for the gRPC External Authorization service.
	Destination RouteDestination `json:"destination"`

	// Authority is the hostname:port of the gRPC External Authorization service.
	Authority string `json:"authority"`
}

// Authorization defines the schema for the authorization.
//
// +k8s:deepcopy-gen=true
type Authorization struct {
	// Rules defines the authorization rules.
	Rules []*AuthorizationRule `json:"rules,omitempty"`

	// DefaultAction defines the default action to be taken if no rules match.
	DefaultAction egv1a1.AuthorizationAction `json:"defaultAction"`
}

// AuthorizationRule defines the schema for the authorization rule.
//
// +k8s:deepcopy-gen=true
type AuthorizationRule struct {
	// Name is a user-defined name for the rule.
	// If not specified, a name will be generated by EG.
	Name string `json:"name"`

	// Action defines the action to be taken if the rule matches.
	Action egv1a1.AuthorizationAction `json:"action"`

	// Operation specifies the operation of a request, such as HTTP methods.
	Operation *egv1a1.Operation `json:"operation,omitempty"`

	// Principal defines the principal to be matched.
	Principal Principal `json:"principal"`
}

// Principal defines the schema for the principal.
//
// +k8s:deepcopy-gen=true
type Principal struct {
	// ClientCIDRs defines the client CIDRs to be matched.
	ClientCIDRs []*CIDRMatch `json:"clientCIDRs,omitempty"`
	// JWT defines the JWT principal to be matched.
	JWT *egv1a1.JWTPrincipal `json:"jwt,omitempty"`
	// Headers defines the headers to be matched.
	Headers []egv1a1.AuthorizationHeaderMatch `json:"headers,omitempty"`
}

// FaultInjection defines the schema for injecting faults into requests.
//
// +k8s:deepcopy-gen=true
type FaultInjection struct {
	// Delay defines the fault injection delay.
	Delay *FaultInjectionDelay `json:"delay,omitempty" yaml:"delay,omitempty"`
	// Abort defines the fault injection abort.
	Abort *FaultInjectionAbort `json:"abort,omitempty" yaml:"abort,omitempty"`
}

// FaultInjectionDelay defines the schema for injecting delay into requests.
//
// +k8s:deepcopy-gen=true
type FaultInjectionDelay struct {
	// FixedDelay defines the fixed delay duration.
	FixedDelay *metav1.Duration `json:"fixedDelay,omitempty" yaml:"fixedDelay,omitempty"`
	// Percentage defines the percentage of requests to be delayed.
	Percentage *float32 `json:"percentage,omitempty" yaml:"percentage,omitempty"`
}

// FaultInjectionAbort defines the schema for injecting abort into requests.
//
// +k8s:deepcopy-gen=true
type FaultInjectionAbort struct {
	// HTTPStatus defines the HTTP status code to be returned.
	HTTPStatus *int32 `json:"httpStatus,omitempty" yaml:"httpStatus,omitempty"`
	// GrpcStatus defines the gRPC status code to be returned.
	GrpcStatus *int32 `json:"grpcStatus,omitempty" yaml:"grpcStatus,omitempty"`
	// Percentage defines the percentage of requests to be aborted.
	Percentage *float32 `json:"percentage,omitempty" yaml:"percentage,omitempty"`
}

// MirrorPolicy specifies a destination to mirror traffic in addition
// to the original destination
//
// +kubebuilder:object:generate=true
type MirrorPolicy struct {
	// Destination defines the target where the request will be mirrored.
	Destination *RouteDestination `json:"destination" yaml:"destination"`
	// Percentage of the traffic to be mirrored by the `destination` field.
	// When absent, all the traffic (100%) will be mirrored.
	// Values are in the range of [0.0, 100.0].
	Percentage *float32 `json:"percentage,omitempty" yaml:"percentage,omitempty"`
}

// Validate the fields within the HTTPRoute structure
func (h *HTTPRoute) Validate() error {
	var errs error
	if h.Name == "" {
		errs = errors.Join(errs, ErrRouteNameEmpty)
	}
	if h.Hostname == "" {
		errs = errors.Join(errs, ErrHTTPRouteHostnameEmpty)
	}
	if h.PathMatch != nil {
		if err := h.PathMatch.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	for _, hMatch := range h.HeaderMatches {
		if err := hMatch.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	for _, qMatch := range h.QueryParamMatches {
		if err := qMatch.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if h.Destination != nil {
		if err := h.Destination.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if h.Redirect != nil {
		if err := h.Redirect.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if h.DirectResponse != nil {
		if err := h.DirectResponse.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if h.URLRewrite != nil {
		if err := h.URLRewrite.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if h.Mirrors != nil {
		for _, mirror := range h.Mirrors {
			if err := mirror.Destination.Validate(); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	if len(h.AddRequestHeaders) > 0 {
		occurred := sets.NewString()
		for _, header := range h.AddRequestHeaders {
			if err := header.Validate(); err != nil {
				errs = errors.Join(errs, err)
			}
			if occurred.Has(header.Name) {
				errs = errors.Join(errs, ErrAddHeaderDuplicate)
				break
			}
			occurred.Insert(header.Name)
		}
	}
	if len(h.RemoveRequestHeaders) > 0 {
		occurred := sets.NewString()
		for _, header := range h.RemoveRequestHeaders {
			if occurred.Has(header) {
				errs = errors.Join(errs, ErrRemoveHeaderDuplicate)
				break
			}
			occurred.Insert(header)
		}
	}
	if len(h.AddResponseHeaders) > 0 {
		occurred := sets.NewString()
		for _, header := range h.AddResponseHeaders {
			if err := header.Validate(); err != nil {
				errs = errors.Join(errs, err)
			}
			if occurred.Has(header.Name) {
				errs = errors.Join(errs, ErrAddHeaderDuplicate)
				break
			}
			occurred.Insert(header.Name)
		}
	}
	if len(h.RemoveResponseHeaders) > 0 {
		occurred := sets.NewString()
		for _, header := range h.RemoveResponseHeaders {
			if occurred.Has(header) {
				errs = errors.Join(errs, ErrRemoveHeaderDuplicate)
				break
			}
			occurred.Insert(header)
		}
	}
	if h.Traffic != nil {
		if err := h.Traffic.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// RouteDestination holds the destination details associated with the route
// +kubebuilder:object:generate=true
type RouteDestination struct {
	// Name of the destination. This field allows the xds layer
	// to check if this route destination already exists and can be
	// reused
	Name     string                `json:"name" yaml:"name"`
	Settings []*DestinationSetting `json:"settings,omitempty" yaml:"settings,omitempty"`
}

// Validate the fields within the RouteDestination structure
func (r *RouteDestination) Validate() error {
	var errs error
	if len(r.Name) == 0 {
		errs = errors.Join(errs, ErrDestinationNameEmpty)
	}
	routeHasAddressTypes := make(map[DestinationAddressType]bool)
	for _, s := range r.Settings {
		if err := s.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
		if s.AddressType != nil {
			routeHasAddressTypes[*s.AddressType] = true
		}
	}
	if len(routeHasAddressTypes) > 1 || (len(routeHasAddressTypes) == 1 && routeHasAddressTypes[MIXED]) {
		errs = errors.Join(ErrRouteDestinationsFQDNMixed)
	}

	return errs
}

func (r *RouteDestination) ToBackendWeights() *BackendWeights {
	w := &BackendWeights{
		Name: r.Name,
	}

	for _, s := range r.Settings {
		if s.Weight == nil {
			continue
		}

		if len(s.Endpoints) > 0 {
			w.Valid += *s.Weight
		} else {
			w.Invalid += *s.Weight
		}
	}

	return w
}

// DestinationSetting holds the settings associated with the destination
// +kubebuilder:object:generate=true
type DestinationSetting struct {
	// Name of the setting
	Name string `json:"name" yaml:"name"`
	// Weight associated with this destination,
	// invalid endpoints are represents with a
	// non-zero weight with an empty endpoints list
	Weight *uint32 `json:"weight,omitempty" yaml:"weight,omitempty"`
	// Priority default to priority 0, the highest level.
	// If multiple destinations share the same priority, they will all be utilized.
	// Lower priority endpoints will be used only if higher priority levels are unavailable.
	Priority *uint32 `json:"priority,omitempty"`
	// Protocol associated with this destination/port.
	Protocol  AppProtocol            `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	Endpoints []*DestinationEndpoint `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	// AddressTypeState specifies the state of DestinationEndpoint address type.
	AddressType *DestinationAddressType `json:"addressType,omitempty" yaml:"addressType,omitempty"`
	// IPFamily specifies the IP family (IPv4 or IPv6) to use for this destination's endpoints.
	// This is derived from the backend service and endpoint slice information.
	IPFamily *egv1a1.IPFamily    `json:"ipFamily,omitempty" yaml:"ipFamily,omitempty"`
	TLS      *TLSUpstreamConfig  `json:"tls,omitempty" yaml:"tls,omitempty"`
	Filters  *DestinationFilters `json:"filters,omitempty" yaml:"filters,omitempty"`
	// ZoneAwareRoutingEnabled specifies whether to enable Zone Aware Routing for this destination's endpoints.
	// This is derived from the backend service and depends on having Kubernetes Topology Aware Routing or Traffic Distribution enabled.
	ZoneAwareRoutingEnabled bool `json:"zoneAwareRoutingEnabled,omitempty" yaml:"zoneAwareRoutingEnabled,omitempty"`
}

// Validate the fields within the DestinationSetting structure
func (d *DestinationSetting) Validate() error {
	var errs error
	for _, ep := range d.Endpoints {
		if err := ep.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// DestinationAddressType describes the address type state for a group of DestinationEndpoint
type DestinationAddressType string

const (
	IP    DestinationAddressType = "IP"
	FQDN  DestinationAddressType = "FQDN"
	MIXED DestinationAddressType = "Mixed"
)

// DestinationEndpoint holds the endpoint details associated with the destination
// +kubebuilder:object:generate=true
type DestinationEndpoint struct {
	// Host refers to the FQDN or IP address of the backend service.
	Host string `json:"host" yaml:"host"`
	// Port on the service to forward the request to.
	Port uint32 `json:"port" yaml:"port"`
	// Path refers to the Unix Domain Socket
	Path *string `json:"path,omitempty" yaml:"path,omitempty"`
	// Draining is true if this endpoint should be drained
	Draining bool `json:"draining,omitempty" yaml:"draining,omitempty"`
	// Zone refers to the topology zone the Endpoint resides in
	Zone *string `json:"zone,omitempty" yaml:"zone,omitempty"`
}

// Validate the fields within the DestinationEndpoint structure
func (d DestinationEndpoint) Validate() error {
	var errs error

	// unix domain socket
	if d.Path != nil {
		if d.Port != 0 {
			errs = errors.Join(errs, ErrDestEndpointUDSPortInvalid)
		}
		if d.Host != "" {
			errs = errors.Join(errs, ErrDestEndpointUDSHostInvalid)
		}
	} else { // IP or FQDN
		err := validation.IsDNS1123Subdomain(d.Host)
		_, pErr := netip.ParseAddr(d.Host)

		if err != nil && pErr != nil {
			errs = errors.Join(errs, ErrDestEndpointHostInvalid)
		}

		if d.Port == 0 {
			errs = errors.Join(errs, ErrDestEndpointPortInvalid)
		}
	}

	return errs
}

// NewDestEndpoint creates a new DestinationEndpoint.
func NewDestEndpoint(host string, port uint32, draining bool, zone *string) *DestinationEndpoint {
	return &DestinationEndpoint{
		Host:     host,
		Port:     port,
		Draining: draining,
		Zone:     zone,
	}
}

// AddHeader configures a header to be added to a request or response.
// +k8s:deepcopy-gen=true
type AddHeader struct {
	Name   string   `json:"name" yaml:"name"`
	Value  []string `json:"value" yaml:"value"`
	Append bool     `json:"append" yaml:"append"`
}

// Validate the fields within the AddHeader structure
func (h AddHeader) Validate() error {
	var errs error
	if h.Name == "" {
		errs = errors.Join(errs, ErrAddHeaderEmptyName)
	}

	return errs
}

// URLRewrite holds the details for how to rewrite a request
// +k8s:deepcopy-gen=true
type URLRewrite struct {
	// Path contains config for rewriting the path of the request.
	Path *ExtendedHTTPPathModifier `json:"path,omitempty" yaml:"path,omitempty"`
	// Host configures the replacement of the request's host header.
	Host *HTTPHostModifier `json:"host,omitempty" yaml:"host,omitempty"`
}

// Validate the fields within the URLRewrite structure
func (r URLRewrite) Validate() error {
	var errs error

	if r.Path != nil {
		if err := r.Path.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if r.Host != nil {
		if err := r.Host.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// Redirect holds the details for how and where to redirect a request
// +k8s:deepcopy-gen=true
type Redirect struct {
	// Scheme configures the replacement of the request's scheme.
	Scheme *string `json:"scheme" yaml:"scheme"`
	// Hostname configures the replacement of the request's hostname.
	Hostname *string `json:"hostname" yaml:"hostname"`
	// Path contains config for rewriting the path of the request.
	Path *HTTPPathModifier `json:"path" yaml:"path"`
	// Port configures the replacement of the request's port.
	Port *uint32 `json:"port" yaml:"port"`
	// Status code configures the redirection response's status code.
	StatusCode *int32 `json:"statusCode" yaml:"statusCode"`
}

// Validate the fields within the Redirect structure
func (r Redirect) Validate() error {
	var errs error

	if r.Scheme != nil {
		if *r.Scheme != "http" && *r.Scheme != "https" {
			errs = errors.Join(errs, ErrRedirectUnsupportedScheme)
		}
	}

	if r.Path != nil {
		if err := r.Path.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if r.StatusCode != nil {
		if *r.StatusCode != 301 && *r.StatusCode != 302 {
			errs = errors.Join(errs, ErrRedirectUnsupportedStatus)
		}
	}

	return errs
}

// HTTPPathModifier holds instructions for how to modify the path of a request on a redirect response
// +k8s:deepcopy-gen=true
type HTTPPathModifier struct {
	// FullReplace provides a string to replace the full path of the request.
	FullReplace *string `json:"fullReplace" yaml:"fullReplace"`
	// PrefixMatchReplace provides a string to replace the matched prefix of the request.
	PrefixMatchReplace *string `json:"prefixMatchReplace" yaml:"prefixMatchReplace"`
}

// Validate the fields within the HTTPPathModifier structure
func (r HTTPPathModifier) Validate() error {
	var errs error

	if r.FullReplace != nil && r.PrefixMatchReplace != nil {
		errs = errors.Join(errs, ErrHTTPPathModifierDoubleReplace)
	}

	if r.FullReplace == nil && r.PrefixMatchReplace == nil {
		errs = errors.Join(errs, ErrHTTPPathModifierNoReplace)
	}

	return errs
}

// ExtendedHTTPPathModifier holds instructions for how to modify the path of a request on a redirect response
// with both core gateway-api and extended envoy gateway capabilities
// +k8s:deepcopy-gen=true
type ExtendedHTTPPathModifier struct {
	HTTPPathModifier `json:",inline" yaml:",inline"`
	// RegexMatchReplace provides a regex to match an a replacement to perform on the path.
	RegexMatchReplace *RegexMatchReplace `json:"regexMatchReplace,omitempty" yaml:"regexMatchReplace,omitempty"`
}

// Validate the fields within the HTTPPathModifier structure
func (r ExtendedHTTPPathModifier) Validate() error {
	var errs error

	rewrites := []bool{r.RegexMatchReplace != nil, r.PrefixMatchReplace != nil, r.FullReplace != nil}
	rwc := 0
	for _, rw := range rewrites {
		if rw {
			rwc++
		}
	}

	if rwc > 1 {
		errs = errors.Join(errs, ErrHTTPPathModifierDoubleReplace)
	}

	if r.FullReplace == nil && r.PrefixMatchReplace == nil && r.RegexMatchReplace == nil {
		errs = errors.Join(errs, ErrHTTPPathModifierNoReplace)
	}

	if r.RegexMatchReplace != nil && (r.RegexMatchReplace.Pattern == "" || r.RegexMatchReplace.Substitution == "") {
		errs = errors.Join(errs, ErrHTTPPathModifierNoReplace)
	}

	return errs
}

// HTTPHostModifier holds instructions for how to modify the host of a request
// with both core gateway-api and extended envoy gateway capabilities
// +k8s:deepcopy-gen=true
type HTTPHostModifier struct {
	Name    *string `json:"name,omitempty" yaml:"name,omitempty"`
	Header  *string `json:"header,omitempty" yaml:"header,omitempty"`
	Backend *bool   `json:"backend,omitempty" yaml:"backend,omitempty"`
}

// Validate the fields within the HTTPPathModifier structure
func (r HTTPHostModifier) Validate() error {
	var errs error

	rewrites := []bool{r.Name != nil, r.Header != nil, r.Backend != nil}
	rwc := 0
	for _, rw := range rewrites {
		if rw {
			rwc++
		}
	}

	if rwc > 1 {
		errs = errors.Join(errs, ErrHTTPHostModifierDoubleReplace)
	}

	return errs
}

// StringMatch holds the various match conditions.
// Only one of Exact, Prefix, SafeRegex or Distinct can be set.
// +k8s:deepcopy-gen=true
type StringMatch struct {
	// Name of the field to match on.
	Name string `json:"name" yaml:"name"`
	// Exact match condition.
	Exact *string `json:"exact,omitempty" yaml:"exact,omitempty"`
	// Prefix match condition.
	Prefix *string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	// Suffix match condition.
	Suffix *string `json:"suffix,omitempty" yaml:"suffix,omitempty"`
	// SafeRegex match condition.
	SafeRegex *string `json:"safeRegex,omitempty" yaml:"safeRegex,omitempty"`
	// Distinct match condition.
	// Used to match any and all possible unique values encountered within the Name field.
	Distinct bool `json:"distinct" yaml:"distinct"`
	// Invert inverts the final match decision
	Invert *bool `json:"invert,omitempty" yaml:"invert,omitempty"`
}

// Validate the fields within the StringMatch structure
func (s StringMatch) Validate() error {
	var errs error
	matchCount := 0
	if s.Exact != nil {
		matchCount++
	}
	if s.Prefix != nil {
		matchCount++
	}
	if s.Suffix != nil {
		matchCount++
	}
	if s.SafeRegex != nil {
		matchCount++
	}
	if s.Distinct {
		if s.Name == "" {
			errs = errors.Join(errs, ErrStringMatchNameIsEmpty)
		}
		if s.Invert != nil && *s.Invert {
			errs = errors.Join(errs, ErrStringMatchInvertDistinctInvalid)
		}
		matchCount++
	}

	if matchCount != 1 {
		errs = errors.Join(errs, ErrStringMatchConditionInvalid)
	}

	return errs
}

// TCPListener holds the TCP listener configuration.
// +k8s:deepcopy-gen=true
type TCPListener struct {
	CoreListenerDetails `json:",inline" yaml:",inline"`
	// TLS holds information for configuring TLS on a listener.
	TLS *TLSConfig `json:"tls,omitempty" yaml:"tls,omitempty"`
	// TCPKeepalive configuration for the listener
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty" yaml:"tcpKeepalive,omitempty"`
	// EnableProxyProtocol enables the listener to interpret proxy protocol header
	EnableProxyProtocol bool `json:"enableProxyProtocol,omitempty" yaml:"enableProxyProtocol,omitempty"`
	// ClientTimeout sets the timeout configuration for downstream connections.
	Timeout *ClientTimeout `json:"timeout,omitempty" yaml:"clientTimeout,omitempty"`
	// Connection settings for clients
	Connection *ClientConnection `json:"connection,omitempty" yaml:"connection,omitempty"`
	// Routes associated with TCP traffic to the listener.
	Routes []*TCPRoute `json:"routes,omitempty" yaml:"routes,omitempty"`
}

// TCPRoute holds the route information associated with the TCP Route
// +k8s:deepcopy-gen=true
type TCPRoute struct {
	// Name of the TCPRoute.
	Name string `json:"name" yaml:"name"`
	// TLS holds information for configuring TLS on a listener
	TLS *TLS `json:"tls,omitempty" yaml:"tls,omitempty"`
	// Destinations associated with TCP traffic to the service.
	Destination *RouteDestination `json:"destination,omitempty" yaml:"destination,omitempty"`
	// TCPKeepalive settings associated with the upstream client connection.
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty" yaml:"tcpKeepalive,omitempty"`
	// load balancer policy to use when routing to the backend endpoints.
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty"`
	// Request and connection timeout settings
	Timeout *Timeout `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// Retry settings
	CircuitBreaker *CircuitBreaker `json:"circuitBreaker,omitempty" yaml:"circuitBreaker,omitempty"`
	// HealthCheck defines the configuration for health checking on the upstream.
	HealthCheck *HealthCheck `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	// Proxy Protocol Settings
	ProxyProtocol *ProxyProtocol `json:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty"`
	// settings of upstream connection
	BackendConnection *BackendConnection `json:"backendConnection,omitempty" yaml:"backendConnection,omitempty"`
	// DNS is used to configure how DNS resolution is handled for the route
	DNS *DNS `json:"dns,omitempty" yaml:"dns,omitempty"`
}

// TLS holds information for configuring TLS on a listener
// +k8s:deepcopy-gen=true
type TLS struct {
	// TLS information required for TLS Passthrough, If provided, incoming
	// connections' server names are inspected and routed to backends accordingly.
	TLSInspectorConfig *TLSInspectorConfig `json:"inspector,omitempty" yaml:"inspector,omitempty"`
	// TLS information required for TLS Termination
	Terminate *TLSConfig `json:"terminate,omitempty" yaml:"terminate,omitempty"`
}

// Validate the fields within the TCPListener structure
func (t TCPListener) Validate() error {
	var errs error
	if t.Name == "" {
		errs = errors.Join(errs, ErrListenerNameEmpty)
	}
	if _, err := netip.ParseAddr(t.Address); err != nil {
		errs = errors.Join(errs, ErrListenerAddressInvalid)
	}
	if t.Port == 0 {
		errs = errors.Join(errs, ErrListenerPortInvalid)
	}
	for _, route := range t.Routes {
		if err := route.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

func (t TCPRoute) Validate() error {
	var errs error

	if t.Name == "" {
		errs = errors.Join(errs, ErrRouteNameEmpty)
	}

	if t.TLS != nil && t.TLS.TLSInspectorConfig != nil {
		if err := t.TLS.TLSInspectorConfig.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if t.TLS != nil && t.TLS.Terminate != nil {
		if err := t.TLS.Terminate.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if t.Destination != nil {
		if err := t.Destination.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if t.LoadBalancer != nil {
		if err := t.LoadBalancer.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	if t.HealthCheck != nil {
		if err := t.HealthCheck.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// TLSInspectorConfig holds the configuration required for inspecting TLS
// connections.
// +k8s:deepcopy-gen=true
type TLSInspectorConfig struct {
	// Server names that are compared against the server names of a new connection.
	// Wildcard hosts are supported in the prefix form. Partial wildcards are not
	// supported, and values like *w.example.com are invalid.
	SNIs []string `json:"snis,omitempty" yaml:"snis,omitempty"`
}

func (t TLSInspectorConfig) Validate() error {
	var errs error
	if len(t.SNIs) == 0 {
		errs = errors.Join(errs, ErrTCPRouteSNIsEmpty)
	}
	return errs
}

// UDPListener holds the UDP listener configuration.
// +k8s:deepcopy-gen=true
type UDPListener struct {
	CoreListenerDetails `json:",inline" yaml:",inline"`
	// Route associated with UDP traffic to the listener.
	Route *UDPRoute `json:"route,omitempty" yaml:"route,omitempty"`
}

// UDPRoute holds the route information associated with the UDP Route.
// +k8s:deepcopy-gen=true
type UDPRoute struct {
	// Name of the UDPRoute.
	Name string `json:"name" yaml:"name"`
	// Destination associated with UDP traffic to the service.
	Destination *RouteDestination `json:"destination,omitempty" yaml:"destination,omitempty"`
	// load balancer policy to use when routing to the backend endpoints.
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty"`
	// DNS is used to configure how DNS resolution is handled by the Envoy Proxy cluster
	DNS *DNS `json:"dns,omitempty" yaml:"dns,omitempty"`
}

// Validate the fields within the UDPListener structure
func (h UDPListener) Validate() error {
	var errs error
	if h.Name == "" {
		errs = errors.Join(errs, ErrListenerNameEmpty)
	}
	if _, err := netip.ParseAddr(h.Address); err != nil {
		errs = errors.Join(errs, ErrListenerAddressInvalid)
	}
	if h.Port == 0 {
		errs = errors.Join(errs, ErrListenerPortInvalid)
	}

	if h.Route != nil {
		if err := h.Route.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (u UDPRoute) Validate() error {
	var errs error

	if u.Name == "" {
		errs = errors.Join(errs, ErrRouteNameEmpty)
	}

	if u.Destination != nil {
		if err := u.Destination.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// RateLimit holds the rate limiting configuration.
// +k8s:deepcopy-gen=true
type RateLimit struct {
	// Global rate limit settings.
	Global *GlobalRateLimit `json:"global,omitempty" yaml:"global,omitempty"`

	// Local rate limit settings.
	Local *LocalRateLimit `json:"local,omitempty" yaml:"local,omitempty"`
}

// GlobalRateLimit holds the global rate limiting configuration.
// +k8s:deepcopy-gen=true
type GlobalRateLimit struct {
	// TODO zhaohuabing: add default values for Global rate limiting.

	// Rules for rate limiting.
	Rules []*RateLimitRule `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// LocalRateLimit holds the local rate limiting configuration.
// +k8s:deepcopy-gen=true
type LocalRateLimit struct {
	// Default rate limiting values.
	// If a request does not match any of the rules, the default values are used.
	Default RateLimitValue `json:"default,omitempty" yaml:"default,omitempty"`

	// Rules for rate limiting.
	Rules []*RateLimitRule `json:"rules,omitempty" yaml:"rules,omitempty"`
}

// RateLimitRule holds the match and limit configuration for ratelimiting.
// +k8s:deepcopy-gen=true
type RateLimitRule struct {
	// HeaderMatches define the match conditions on the request headers for this route.
	HeaderMatches []*StringMatch `json:"headerMatches" yaml:"headerMatches"`
	// CIDRMatch define the match conditions on the source IP's CIDR for this route.
	CIDRMatch *CIDRMatch `json:"cidrMatch,omitempty" yaml:"cidrMatch,omitempty"`
	// Limit holds the rate limit values.
	Limit RateLimitValue `json:"limit,omitempty" yaml:"limit,omitempty"`
	// RequestCost specifies the cost of the request.
	RequestCost *RateLimitCost `json:"requestCost,omitempty" yaml:"requestCost,omitempty"`
	// ResponseCost specifies the cost of the response.
	ResponseCost *RateLimitCost `json:"responseCost,omitempty" yaml:"responseCost,omitempty"`
}

// RateLimitCost specifies the cost of the request or response.
// +k8s:deepcopy-gen=true
type RateLimitCost struct {
	Number *uint64 `json:"number,omitempty" yaml:"number,omitempty"`
	Format *string `json:"format,omitempty" yaml:"format,omitempty"`
}

type CIDRMatch struct {
	CIDR    string `json:"cidr" yaml:"cidr"`
	IP      string `json:"ip" yaml:"ip"`
	MaskLen uint32 `json:"maskLen" yaml:"maskLen"`
	IsIPv6  bool   `json:"isIPv6" yaml:"isIPv6"`
	// Distinct means that each IP Address within the specified Source IP CIDR is treated as a distinct client selector
	// and uses a separate rate limit bucket/counter.
	Distinct bool `json:"distinct" yaml:"distinct"`
}

// TODO zhaohuabing: remove this function
func (r *RateLimitRule) IsMatchSet() bool {
	return len(r.HeaderMatches) != 0 || r.CIDRMatch != nil
}

type RateLimitUnit egv1a1.RateLimitUnit

// RateLimitValue holds the
// +k8s:deepcopy-gen=true
type RateLimitValue struct {
	// Requests are the number of requests that need to be rate limited.
	Requests uint `json:"requests" yaml:"requests"`
	// Unit of rate limiting.
	Unit RateLimitUnit `json:"unit" yaml:"unit"`
}

type ProxyAccessLogType egv1a1.ProxyAccessLogType

const (
	ProxyAccessLogTypeRoute    = ProxyAccessLogType(egv1a1.ProxyAccessLogTypeRoute)
	ProxyAccessLogTypeListener = ProxyAccessLogType(egv1a1.ProxyAccessLogTypeListener)
)

// ReadyListener holds the configuration for ready listener.
// +k8s:deepcopy-gen=true
type ReadyListener struct {
	IPFamily egv1a1.IPFamily `json:"ipFamily" yaml:"ipFamily"`
	Address  string          `json:"address" yaml:"address"`
	Port     uint32          `json:"port" yaml:"port"`
	Path     string          `json:"path" yaml:"path"`
}

// AccessLog holds the access logging configuration.
// +k8s:deepcopy-gen=true
type AccessLog struct {
	Text          []*TextAccessLog          `json:"text,omitempty" yaml:"text,omitempty"`
	JSON          []*JSONAccessLog          `json:"json,omitempty" yaml:"json,omitempty"`
	ALS           []*ALSAccessLog           `json:"als,omitempty" yaml:"als,omitempty"`
	OpenTelemetry []*OpenTelemetryAccessLog `json:"openTelemetry,omitempty" yaml:"openTelemetry,omitempty"`
}

// TextAccessLog holds the configuration for text access logging.
// +k8s:deepcopy-gen=true
type TextAccessLog struct {
	CELMatches []string            `json:"celMatches,omitempty" yaml:"celMatches,omitempty"`
	Format     *string             `json:"format,omitempty" yaml:"format,omitempty"`
	Path       string              `json:"path" yaml:"path"`
	LogType    *ProxyAccessLogType `json:"logType,omitempty" yaml:"logType,omitempty"`
}

// JSONAccessLog holds the configuration for JSON access logging.
// +k8s:deepcopy-gen=true
type JSONAccessLog struct {
	CELMatches []string            `json:"celMatches,omitempty" yaml:"celMatches,omitempty"`
	JSON       map[string]string   `json:"json,omitempty" yaml:"json,omitempty"`
	Path       string              `json:"path" yaml:"path"`
	LogType    *ProxyAccessLogType `json:"logType,omitempty" yaml:"logType,omitempty"`
}

// ALSAccessLog holds the configuration for gRPC ALS access logging.
// +k8s:deepcopy-gen=true
type ALSAccessLog struct {
	CELMatches  []string                          `json:"celMatches,omitempty" yaml:"celMatches,omitempty"`
	LogName     string                            `json:"name" yaml:"name"`
	Destination RouteDestination                  `json:"destination,omitempty" yaml:"destination,omitempty"`
	Traffic     *TrafficFeatures                  `json:"traffic,omitempty" yaml:"traffic,omitempty"`
	Type        egv1a1.ALSEnvoyProxyAccessLogType `json:"type" yaml:"type"`
	Text        *string                           `json:"text,omitempty" yaml:"text,omitempty"`
	Attributes  map[string]string                 `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	HTTP        *ALSAccessLogHTTP                 `json:"http,omitempty" yaml:"http,omitempty"`
	LogType     *ProxyAccessLogType               `json:"logType,omitempty" yaml:"logType,omitempty"`
}

// ALSAccessLogHTTP holds the configuration for HTTP ALS access logging.
// +k8s:deepcopy-gen=true
type ALSAccessLogHTTP struct {
	RequestHeaders   []string `json:"requestHeaders,omitempty" yaml:"requestHeaders,omitempty"`
	ResponseHeaders  []string `json:"responseHeaders,omitempty" yaml:"responseHeaders,omitempty"`
	ResponseTrailers []string `json:"responseTrailers,omitempty" yaml:"responseTrailers,omitempty"`
}

// OpenTelemetryAccessLog holds the configuration for OpenTelemetry access logging.
// +k8s:deepcopy-gen=true
type OpenTelemetryAccessLog struct {
	CELMatches  []string            `json:"celMatches,omitempty" yaml:"celMatches,omitempty"`
	Authority   string              `json:"authority,omitempty" yaml:"authority,omitempty"`
	Text        *string             `json:"text,omitempty" yaml:"text,omitempty"`
	Attributes  map[string]string   `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Resources   map[string]string   `json:"resources,omitempty" yaml:"resources,omitempty"`
	Destination RouteDestination    `json:"destination,omitempty" yaml:"destination,omitempty"`
	Traffic     *TrafficFeatures    `json:"traffic,omitempty" yaml:"traffic,omitempty"`
	LogType     *ProxyAccessLogType `json:"logType,omitempty" yaml:"logType,omitempty"`
}

// EnvoyPatchPolicy defines the intermediate representation of the EnvoyPatchPolicy resource.
// +k8s:deepcopy-gen=true
type EnvoyPatchPolicy struct {
	EnvoyPatchPolicyStatus
	// JSONPatches are the JSON Patches that
	// are to be applied to generated Xds linked to the gateway.
	JSONPatches []*JSONPatchConfig `json:"jsonPatches,omitempty" yaml:"jsonPatches,omitempty"`
}

// EnvoyPatchPolicyStatus defines the status reference for the EnvoyPatchPolicy resource
// +k8s:deepcopy-gen=true
type EnvoyPatchPolicyStatus struct {
	Name      string `json:"name,omitempty" yaml:"name"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace"`
	// Status of the EnvoyPatchPolicy
	Status *gwapiv1a2.PolicyStatus `json:"status,omitempty" yaml:"status,omitempty"`
}

// JSONPatchConfig defines the configuration for patching a Envoy xDS Resource
// using JSONPatch semantics
// +k8s:deepcopy-gen=true
type JSONPatchConfig struct {
	// Type is the typed URL of the Envoy xDS Resource
	Type string `json:"type" yaml:"type"`
	// Name is the name of the resource
	Name string `json:"name" yaml:"name"`
	// Patch defines the JSON Patch Operation
	Operation JSONPatchOperation `json:"operation" yaml:"operation"`
}

type JSONPatchOp string

const (
	JSONPatchOpAdd     JSONPatchOp = "add"
	JSONPatchOpRemove  JSONPatchOp = "remove"
	JSONPatchOpReplace JSONPatchOp = "replace"
	JSONPatchOpCopy    JSONPatchOp = "copy"
	JSONPatchOpMove    JSONPatchOp = "move"
	JSONPatchOpTest    JSONPatchOp = "test"
)

func TranslateJSONPatchOp(op egv1a1.JSONPatchOperationType) JSONPatchOp {
	switch op {
	case "add":
		return JSONPatchOpAdd
	case "remove":
		return JSONPatchOpRemove
	case "replace":
		return JSONPatchOpReplace
	case "move":
		return JSONPatchOpMove
	case "copy":
		return JSONPatchOpCopy
	case "test":
		return JSONPatchOpTest
	default:
		return ""
	}
}

// JSONPatchOperation defines the JSON Patch Operation as defined in
// https://datatracker.ietf.org/doc/html/rfc6902
// +k8s:deepcopy-gen=true
type JSONPatchOperation struct {
	// Op is the type of operation to perform
	Op JSONPatchOp `json:"op" yaml:"op"`
	// Path is the location of the target document/field where the operation will be performed
	// Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details.
	// +optional
	Path *string `json:"path,omitempty"  yaml:"path,omitempty"`
	// JSONPath specifies the locations of the target document/field where the operation will be performed
	// Refer to https://datatracker.ietf.org/doc/rfc9535/ for more details.
	// +optional
	JSONPath *string `json:"jsonPath,omitempty"  yaml:"jsonPath,omitempty"`
	// From is the source location of the value to be copied or moved. Only valid
	// for move or copy operations
	// Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details.
	// +optional
	From *string `json:"from,omitempty" yaml:"from,omitempty"`
	// Value is the new value of the path location.
	Value *apiextensionsv1.JSON `json:"value,omitempty" yaml:"value,omitempty"`
}

func (o *JSONPatchOperation) IsPathNilOrEmpty() bool {
	return o.Path == nil || *o.Path == EmptyPath
}

func (o *JSONPatchOperation) IsJSONPathNilOrEmpty() bool {
	return o.JSONPath == nil || *o.JSONPath == EmptyPath
}

// Validate ensures that the appropriate fields are set for each operation type according to RFC 6902:
// https://www.rfc-editor.org/rfc/rfc6902.html
func (o *JSONPatchOperation) Validate() error {
	if o.Path == nil && o.JSONPath == nil {
		return fmt.Errorf("a patch operation must specify a path or jsonPath")
	}
	switch o.Op {
	case JSONPatchOpAdd, JSONPatchOpReplace, JSONPatchOpTest:
		if o.Value == nil {
			return fmt.Errorf("the %s operation requires a value", o.Op)
		}
		if o.From != nil {
			return fmt.Errorf("the %s operation doesn't support a from attribute", o.Op)
		}
	case JSONPatchOpRemove:
		if o.From != nil || o.Value != nil {
			return fmt.Errorf("value and from can't be specified with the remove operation")
		}
	case JSONPatchOpMove, JSONPatchOpCopy:
		if o.From == nil {
			return fmt.Errorf("the %s operation requires a valid from attribute", o.Op)
		}
		if o.Value != nil {
			return fmt.Errorf("the %s operation doesn't support a value attribute", o.Op)
		}
	default:
		return fmt.Errorf("unsupported JSONPatch operation")
	}
	return nil
}

// Tracing defines the configuration for tracing a Envoy xDS Resource
// +k8s:deepcopy-gen=true
type Tracing struct {
	ServiceName  string                      `json:"serviceName"`
	Authority    string                      `json:"authority,omitempty"`
	SamplingRate float64                     `json:"samplingRate,omitempty"`
	CustomTags   map[string]egv1a1.CustomTag `json:"customTags,omitempty"`
	Destination  RouteDestination            `json:"destination,omitempty"`
	Traffic      *TrafficFeatures            `json:"traffic,omitempty"`
	Provider     egv1a1.TracingProvider      `json:"provider"`
}

// Metrics defines the configuration for metrics generated by Envoy
// +k8s:deepcopy-gen=true
type Metrics struct {
	EnableVirtualHostStats          bool `json:"enableVirtualHostStats" yaml:"enableVirtualHostStats"`
	EnablePerEndpointStats          bool `json:"enablePerEndpointStats" yaml:"enablePerEndpointStats"`
	EnableRequestResponseSizesStats bool `json:"enableRequestResponseSizesStats" yaml:"enableRequestResponseSizesStats"`
}

// TCPKeepalive define the TCP Keepalive configuration.
// +k8s:deepcopy-gen=true
type TCPKeepalive struct {
	// The total number of unacknowledged probes to send before deciding
	// the connection is dead.
	// Defaults to 9.
	Probes *uint32 `json:"probes,omitempty" yaml:"probes,omitempty"`
	// The duration, in seconds, a connection needs to be idle before keep-alive
	// probes start being sent.
	// Defaults to `7200s`.
	IdleTime *uint32 `json:"idleTime,omitempty" yaml:"idleTime,omitempty"`
	// The duration, in seconds, between keep-alive probes.
	// Defaults to `75s`.
	Interval *uint32 `json:"interval,omitempty" yaml:"interval,omitempty"`
}

// LoadBalancer defines the load balancer settings.
// +k8s:deepcopy-gen=true
type LoadBalancer struct {
	// RoundRobin load balacning policy
	RoundRobin *RoundRobin `json:"roundRobin,omitempty" yaml:"roundRobin,omitempty"`
	// LeastRequest load balancer policy
	LeastRequest *LeastRequest `json:"leastRequest,omitempty" yaml:"leastRequest,omitempty"`
	// Random load balancer policy
	Random *Random `json:"random,omitempty" yaml:"random,omitempty"`
	// ConsistentHash load balancer policy
	ConsistentHash *ConsistentHash `json:"consistentHash,omitempty" yaml:"consistentHash,omitempty"`
}

// Validate the fields within the LoadBalancer structure
func (l *LoadBalancer) Validate() error {
	var errs error
	matchCount := 0
	if l.RoundRobin != nil {
		matchCount++
	}
	if l.LeastRequest != nil {
		matchCount++
	}
	if l.Random != nil {
		matchCount++
	}
	if l.ConsistentHash != nil {
		matchCount++
	}
	if matchCount != 1 {
		errs = errors.Join(errs, ErrLoadBalancerInvalid)
	}

	return errs
}

// RoundRobin load balancer settings
// +k8s:deepcopy-gen=true
type RoundRobin struct {
	// SlowStart defines the slow start configuration.
	// If set, slow start mode is enabled for newly added hosts in the cluster.
	SlowStart *SlowStart `json:"slowStart,omitempty" yaml:"slowStart,omitempty"`
}

// LeastRequest load balancer settings
// +k8s:deepcopy-gen=true
type LeastRequest struct {
	// SlowStart defines the slow start configuration.
	// If set, slow start mode is enabled for newly added hosts in the cluster.
	SlowStart *SlowStart `json:"slowStart,omitempty" yaml:"slowStart,omitempty"`
}

// Random load balancer settings
// +k8s:deepcopy-gen=true
type Random struct{}

// ConsistentHash load balancer settings
// +k8s:deepcopy-gen=true
type ConsistentHash struct {
	// Hash based on the Source IP Address
	SourceIP  *bool          `json:"sourceIP,omitempty" yaml:"sourceIP,omitempty"`
	Header    *Header        `json:"header,omitempty" yaml:"header,omitempty"`
	Cookie    *egv1a1.Cookie `json:"cookie,omitempty" yaml:"cookie,omitempty"`
	TableSize *uint64        `json:"tableSize,omitempty" yaml:"tableSize,omitempty"`
}

// Header consistent hash type settings
type Header struct {
	Name string `json:"name" yaml:"name"`
}

type ProxyProtocolVersion string

const (
	// ProxyProtocolVersionV1 is the PROXY protocol version 1 (human readable format).
	ProxyProtocolVersionV1 ProxyProtocolVersion = "V1"
	// ProxyProtocolVersionV2 is the PROXY protocol version 2 (binary format).
	ProxyProtocolVersionV2 ProxyProtocolVersion = "V2"
)

// ProxyProtocol upstream settings
// +k8s:deepcopy-gen=true
type ProxyProtocol struct {
	// Version of proxy protocol to use
	Version ProxyProtocolVersion `json:"version,omitempty" yaml:"version,omitempty"`
}

// SlowStart defines the slow start configuration.
// +k8s:deepcopy-gen=true
type SlowStart struct {
	// Window defines the duration of the warm up period for newly added host.
	Window *metav1.Duration `json:"window" yaml:"window"`
}

// Backend CircuitBreaker settings for the DEFAULT routing priority
// +k8s:deepcopy-gen=true
type CircuitBreaker struct {
	// The maximum number of connections that Envoy will establish.
	MaxConnections *uint32 `json:"maxConnections,omitempty" yaml:"maxConnections,omitempty"`

	// The maximum number of pending requests that Envoy will queue.
	MaxPendingRequests *uint32 `json:"maxPendingRequests,omitempty" yaml:"maxPendingRequests,omitempty"`

	// The maximum number of parallel requests that Envoy will make.
	MaxParallelRequests *uint32 `json:"maxParallelRequests,omitempty" yaml:"maxParallelRequests,omitempty"`

	// The maximum number of parallel requests that Envoy will make.
	MaxRequestsPerConnection *uint32 `json:"maxRequestsPerConnection,omitempty" yaml:"maxRequestsPerConnection,omitempty"`

	// The maximum number of parallel retries that Envoy will make.
	MaxParallelRetries *uint32 `json:"maxParallelRetries,omitempty" yaml:"maxParallelRetries,omitempty"`

	// PerEndpoint defines per-endpoint Circuit Breakers
	PerEndpoint *PerEndpointCircuitBreakers `json:"perEndpoint,omitempty"`
}

// PerEndpointCircuitBreakers defines the per-endpoint Circuit Breaker configuration.
// +k8s:deepcopy-gen=true
type PerEndpointCircuitBreakers struct {
	// MaxConnections configures the maximum number of connections that Envoy will establish per-endpoint to the referenced backend defined within a xRoute rule.
	MaxConnections *uint32 `json:"maxConnections,omitempty"`
}

// HealthCheck defines health check settings
// +k8s:deepcopy-gen=true
type HealthCheck struct {
	Active *ActiveHealthCheck `json:"active,omitempty" yaml:"active,omitempty"`

	Passive *OutlierDetection `json:"passive,omitempty" yaml:"passive,omitempty"`

	PanicThreshold *uint32 `json:"panicThreshold,omitempty" yaml:"panicThreshold,omitempty"`
}

// OutlierDetection defines passive health check settings
// +k8s:deepcopy-gen=true
type OutlierDetection struct {
	// Interval defines the time between passive health checks.
	Interval *metav1.Duration `json:"interval,omitempty"`
	// SplitExternalLocalOriginErrors enables splitting of errors between external and local origin.
	SplitExternalLocalOriginErrors *bool `json:"splitExternalLocalOriginErrors,omitempty" yaml:"splitExternalLocalOriginErrors,omitempty"`
	// ConsecutiveLocalOriginFailures sets the number of consecutive local origin failures triggering ejection.
	ConsecutiveLocalOriginFailures *uint32 `json:"consecutiveLocalOriginFailures,omitempty" yaml:"consecutiveLocalOriginFailures,omitempty"`
	// ConsecutiveGatewayErrors sets the number of consecutive gateway errors triggering ejection.
	ConsecutiveGatewayErrors *uint32 `json:"consecutiveGatewayErrors,omitempty" yaml:"consecutiveGatewayErrors,omitempty"`
	// Consecutive5xxErrors sets the number of consecutive 5xx errors triggering ejection.
	Consecutive5xxErrors *uint32 `json:"consecutive5XxErrors,omitempty" yaml:"consecutive5XxErrors,omitempty"`
	// BaseEjectionTime defines the base duration for which a host will be ejected on consecutive failures.
	BaseEjectionTime *metav1.Duration `json:"baseEjectionTime,omitempty" yaml:"baseEjectionTime,omitempty"`
	// MaxEjectionPercent sets the maximum percentage of hosts in a cluster that can be ejected.
	MaxEjectionPercent *int32 `json:"maxEjectionPercent,omitempty" yaml:"maxEjectionPercent,omitempty"`
}

// ActiveHealthCheck defines active health check settings
// +k8s:deepcopy-gen=true
type ActiveHealthCheck struct {
	// Timeout defines the time to wait for a health check response.
	Timeout *metav1.Duration `json:"timeout"`
	// Interval defines the time between active health checks.
	Interval *metav1.Duration `json:"interval"`
	// UnhealthyThreshold defines the number of unhealthy health checks required before a backend host is marked unhealthy.
	UnhealthyThreshold *uint32 `json:"unhealthyThreshold"`
	// HealthyThreshold defines the number of healthy health checks required before a backend host is marked healthy.
	HealthyThreshold *uint32 `json:"healthyThreshold"`
	// HTTP defines the configuration of http health checker.
	HTTP *HTTPHealthChecker `json:"http,omitempty" yaml:"http,omitempty"`
	// TCP defines the configuration of tcp health checker.
	TCP *TCPHealthChecker `json:"tcp,omitempty" yaml:"tcp,omitempty"`
	// GRPC defines if the GRPC healthcheck service should be used
	GRPC *GRPCHealthChecker `json:"grpc,omitempty" yaml:"grpc,omitempty"`
}

func (h *HealthCheck) SetHTTPHostIfAbsent(host string) {
	if h != nil && h.Active != nil && h.Active.HTTP != nil && h.Active.HTTP.Host == "" {
		h.Active.HTTP.Host = host
	}
}

// Validate the fields within the HealthCheck structure.
func (h *HealthCheck) Validate() error {
	var errs error
	if h.Active != nil {
		if h.Active.Timeout != nil && h.Active.Timeout.Duration == 0 {
			errs = errors.Join(errs, ErrHealthCheckTimeoutInvalid)
		}
		if h.Active.Interval != nil && h.Active.Interval.Duration == 0 {
			errs = errors.Join(errs, ErrHealthCheckIntervalInvalid)
		}
		if h.Active.UnhealthyThreshold != nil && *h.Active.UnhealthyThreshold == 0 {
			errs = errors.Join(errs, ErrHealthCheckUnhealthyThresholdInvalid)
		}
		if h.Active.HealthyThreshold != nil && *h.Active.HealthyThreshold == 0 {
			errs = errors.Join(errs, ErrHealthCheckHealthyThresholdInvalid)
		}

		matchCount := 0
		if h.Active.HTTP != nil {
			matchCount++
		}
		if h.Active.TCP != nil {
			matchCount++
		}
		if h.Active.GRPC != nil {
			matchCount++
		}
		if matchCount > 1 {
			errs = errors.Join(errs, ErrHealthCheckerInvalid)
		}

		if h.Active.HTTP != nil {
			if err := h.Active.HTTP.Validate(); err != nil {
				errs = errors.Join(errs, err)
			}
		}
		if h.Active.TCP != nil {
			if err := h.Active.TCP.Validate(); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	if h.Passive != nil {
		if h.Passive.BaseEjectionTime != nil && h.Passive.BaseEjectionTime.Duration == 0 {
			errs = errors.Join(errs, ErrOutlierDetectionBaseEjectionTimeInvalid)
		}

		if h.Passive.Interval != nil && h.Passive.Interval.Duration == 0 {
			errs = errors.Join(errs, ErrOutlierDetectionIntervalInvalid)
		}
	}

	if h.PanicThreshold != nil {
		if *h.PanicThreshold > 100 {
			errs = errors.Join(errs, ErrPanicThresholdInvalid)
		}
	}

	return errs
}

// HTTPHealthChecker defines the settings of http health check.
// +k8s:deepcopy-gen=true
type HTTPHealthChecker struct {
	// Host defines the value of the host header in the HTTP health check request.
	Host string `json:"host" yaml:"host"`
	// Path defines the HTTP path that will be requested during health checking.
	Path string `json:"path" yaml:"path"`
	// Method defines the HTTP method used for health checking.
	Method *string `json:"method,omitempty" yaml:"method,omitempty"`
	// ExpectedStatuses defines a list of HTTP response statuses considered healthy.
	ExpectedStatuses []HTTPStatus `json:"expectedStatuses,omitempty" yaml:"expectedStatuses,omitempty"`
	// ExpectedResponse defines a list of HTTP expected responses to match.
	ExpectedResponse *HealthCheckPayload `json:"expectedResponse,omitempty" yaml:"expectedResponses,omitempty"`
}

// Validate the fields within the HTTPHealthChecker structure.
func (c *HTTPHealthChecker) Validate() error {
	var errs error
	if c.Host == "" {
		errs = errors.Join(errs, ErrHCHTTPHostInvalid)
	}
	if c.Path == "" {
		errs = errors.Join(errs, ErrHCHTTPPathInvalid)
	}
	if c.Method != nil {
		switch *c.Method {
		case http.MethodGet:
		case http.MethodHead:
		case http.MethodPost:
		case http.MethodPut:
		case http.MethodDelete:
		case http.MethodOptions:
		case http.MethodTrace:
		case http.MethodPatch:
		case "":
		default:
			errs = errors.Join(errs, ErrHCHTTPMethodInvalid)
		}
	}
	if len(c.ExpectedStatuses) == 0 {
		errs = errors.Join(errs, ErrHCHTTPExpectedStatusesInvalid)
	}
	for _, r := range c.ExpectedStatuses {
		if err := r.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if c.ExpectedResponse != nil {
		if err := c.ExpectedResponse.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

// HTTPStatus represents http status code.
type HTTPStatus int

func (h HTTPStatus) Validate() error {
	if h < 100 || h >= 600 {
		return ErrHTTPStatusInvalid
	}
	return nil
}

// GRPCHealthChecker defines the settings of the gRPC health check.
// +k8s:deepcopy-gen=true
type GRPCHealthChecker struct {
	// Service is the name of a specific service hosted by the server for
	// which the health check should be requested. If not specified, then the default
	// is to send a health check request for the entire server.
	Service *string `json:"service,omitempty" yaml:"service,omitempty"`
}

// TCPHealthChecker defines the settings of tcp health check.
// +k8s:deepcopy-gen=true
type TCPHealthChecker struct {
	Send    *HealthCheckPayload `json:"send,omitempty" yaml:"send,omitempty"`
	Receive *HealthCheckPayload `json:"receive,omitempty" yaml:"receive,omitempty"`
}

// Validate the fields within the TCPHealthChecker structure.
func (c *TCPHealthChecker) Validate() error {
	var errs error
	if c.Send != nil {
		if err := c.Send.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	if c.Receive != nil {
		if err := c.Receive.Validate(); err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}

// HealthCheckPayload defines the encoding of the payload bytes in the payload.
// +k8s:deepcopy-gen=true
type HealthCheckPayload struct {
	// Text payload in plain text.
	Text *string `json:"text,omitempty" yaml:"text,omitempty"`
	// Binary payload base64 encoded
	Binary []byte `json:"binary,omitempty" yaml:"binary,omitempty"`
}

// Validate the fields in the HealthCheckPayload.
func (p *HealthCheckPayload) Validate() error {
	var errs error
	matchCount := 0
	if p.Text != nil && *p.Text != "" {
		matchCount++
	}
	if len(p.Binary) > 0 {
		matchCount++
	}
	if matchCount != 1 {
		errs = errors.Join(errs, ErrHealthCheckPayloadInvalid)
	}
	return errs
}

// Backend connection timeout settings
// +k8s:deepcopy-gen=true
type Timeout struct {
	// Timeout settings for TCP.
	TCP *TCPTimeout `json:"tcp,omitempty" yaml:"tcp,omitempty"`

	// Timeout settings for HTTP.
	HTTP *HTTPTimeout `json:"http,omitempty" yaml:"tcp,omitempty"`
}

// +k8s:deepcopy-gen=true
type TCPTimeout struct {
	// The timeout for network connection establishment, including TCP and TLS handshakes.
	ConnectTimeout *metav1.Duration `json:"connectTimeout,omitempty" yaml:"connectTimeout,omitempty"`
}

// +k8s:deepcopy-gen=true
type HTTPTimeout struct {
	// RequestTimeout is the time until which entire response is received from the upstream.
	RequestTimeout *metav1.Duration `json:"requestTimeout,omitempty" yaml:"requestTimeout,omitempty"`

	// The idle timeout for an HTTP connection. Idle time is defined as a period in which there are no active requests in the connection.
	ConnectionIdleTimeout *metav1.Duration `json:"connectionIdleTimeout,omitempty" yaml:"connectionIdleTimeout,omitempty"`

	// The maximum duration of an HTTP connection.
	MaxConnectionDuration *metav1.Duration `json:"maxConnectionDuration,omitempty" yaml:"maxConnectionDuration,omitempty"`
}

// Retry define the retry policy configuration.
// +k8s:deepcopy-gen=true
type Retry struct {
	// NumRetries is the number of retries to be attempted. Defaults to 2.
	NumRetries *uint32 `json:"numRetries,omitempty"`

	// RetryOn specifies the retry trigger condition.
	RetryOn *RetryOn `json:"retryOn,omitempty"`

	// PerRetry is the retry policy to be applied per retry attempt.
	PerRetry *PerRetryPolicy `json:"perRetry,omitempty"`
}

type TriggerEnum egv1a1.TriggerEnum

const (
	Error5XX             = TriggerEnum(egv1a1.Error5XX)
	GatewayError         = TriggerEnum(egv1a1.GatewayError)
	Reset                = TriggerEnum(egv1a1.Reset)
	ConnectFailure       = TriggerEnum(egv1a1.ConnectFailure)
	Retriable4XX         = TriggerEnum(egv1a1.Retriable4XX)
	RefusedStream        = TriggerEnum(egv1a1.RefusedStream)
	RetriableStatusCodes = TriggerEnum(egv1a1.RetriableStatusCodes)
	Cancelled            = TriggerEnum(egv1a1.Cancelled)
	DeadlineExceeded     = TriggerEnum(egv1a1.DeadlineExceeded)
	Internal             = TriggerEnum(egv1a1.Internal)
	ResourceExhausted    = TriggerEnum(egv1a1.ResourceExhausted)
	Unavailable          = TriggerEnum(egv1a1.Unavailable)
)

// RetryOn specifies the retry policy.
// +k8s:deepcopy-gen=true
type RetryOn struct {
	// Triggers specifies the retry trigger condition(Http/Grpc).
	Triggers []TriggerEnum `json:"triggers,omitempty"`

	// HttpStatusCodes specifies the http status codes to be retried.
	HTTPStatusCodes []HTTPStatus `json:"httpStatusCodes,omitempty"`
}

// +k8s:deepcopy-gen=true
type PerRetryPolicy struct {
	// Timeout is the timeout per retry attempt.
	Timeout *metav1.Duration `json:"timeout,omitempty"`
	// Backoff is the backoff policy to be applied per retry attempt.
	BackOff *BackOffPolicy `json:"backOff,omitempty"`
}

// +k8s:deepcopy-gen=true
type BackOffPolicy struct {
	// BaseInterval is the base interval between retries.
	BaseInterval *metav1.Duration `json:"baseInterval,omitempty"`
	// MaxInterval is the maximum interval between retries.
	MaxInterval *metav1.Duration `json:"maxInterval,omitempty"`
}

// TLSUpstreamConfig contains sni and ca file in []byte format.
// +k8s:deepcopy-gen=true
type TLSUpstreamConfig struct {
	SNI                 string            `json:"sni,omitempty" yaml:"sni,omitempty"`
	UseSystemTrustStore bool              `json:"useSystemTrustStore,omitempty" yaml:"useSystemTrustStore,omitempty"`
	CACertificate       *TLSCACertificate `json:"caCertificate,omitempty" yaml:"caCertificate,omitempty"`
	TLSConfig           `json:",inline"`
}

func (t *TLSUpstreamConfig) ToTLSConfig() (*tls.Config, error) {
	// nolint:gosec
	tlsConfig := &tls.Config{
		ServerName: t.SNI,
	}
	if t.MinVersion != nil {
		tlsConfig.MinVersion = t.MinVersion.Int()
	}
	if t.MaxVersion != nil {
		tlsConfig.MaxVersion = t.MaxVersion.Int()
	}
	if t.CACertificate != nil {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(t.CACertificate.Certificate)
		tlsConfig.RootCAs = caCertPool
	}
	for _, cert := range t.ClientCertificates {
		cert, err := tls.X509KeyPair(cert.Certificate, cert.PrivateKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}
	return tlsConfig, nil
}

// BackendConnection settings for upstream connections
// +k8s:deepcopy-gen=true
type BackendConnection struct {
	// BufferLimitBytes is the maximum number of bytes that can be buffered for a connection.
	BufferLimitBytes *uint32 `json:"bufferLimit,omitempty" yaml:"bufferLimit,omitempty"`
}

// ClientConnection settings for downstream connections
// +k8s:deepcopy-gen=true
type ClientConnection struct {
	// ConnectionLimit is the limit of number of connections
	ConnectionLimit *ConnectionLimit `json:"limit,omitempty" yaml:"limit,omitempty"`
	// BufferLimitBytes is the maximum number of bytes that can be buffered for a connection.
	BufferLimitBytes *uint32 `json:"bufferLimit,omitempty" yaml:"bufferLimit,omitempty"`
}

// ConnectionLimit contains settings for downstream connection limits
// +k8s:deepcopy-gen=true
type ConnectionLimit struct {
	// Value of the maximum concurrent connections limit.
	// When the limit is reached, incoming connections will be closed after the CloseDelay duration.
	Value *uint64 `json:"value,omitempty" yaml:"value,omitempty"`

	// CloseDelay defines the delay to use before closing connections that are rejected
	// once the limit value is reached.
	CloseDelay *metav1.Duration `json:"closeDelay,omitempty" yaml:"closeDelay,omitempty"`
}

type ExtProcBodyProcessingMode egv1a1.ExtProcBodyProcessingMode

const (
	// ExtProcBodyStreamed sets the streamed body processing mode
	ExtProcBodyStreamed = ExtProcBodyProcessingMode(egv1a1.StreamedExtProcBodyProcessingMode)
	// ExtProcBodyBuffered sets the buffered body processing mode
	ExtProcBodyBuffered = ExtProcBodyProcessingMode(egv1a1.BufferedExtProcBodyProcessingMode)
	// ExtProcBodyBufferedPartial sets the partial buffered body processing mode
	ExtProcBodyBufferedPartial = ExtProcBodyProcessingMode(egv1a1.BufferedPartialExtBodyHeaderProcessingMode)
	// ExtProcBodyFullDuplexStreamed sets the full duplex streamed processing mode
	ExtProcBodyFullDuplexStreamed = ExtProcBodyProcessingMode(egv1a1.FullDuplexStreamedExtBodyProcessingMode)
)

// ExtProc holds the information associated with the ExtProc extensions.
// +k8s:deepcopy-gen=true
type ExtProc struct {
	// Name is a unique name for an ExtProc configuration.
	// The xds translator only generates one ExtProc filter for each unique name.
	Name string `json:"name" yaml:"name"`

	// Destination defines the destination for the gRPC External Processing service.
	Destination RouteDestination `json:"destination" yaml:"destination"`

	// Traffic holds the features associated with traffic management
	Traffic *TrafficFeatures `json:"traffic,omitempty" yaml:"traffic,omitempty"`

	// Authority is the hostname:port of the HTTP External Processing service.
	Authority string `json:"authority" yaml:"authority"`

	// MessageTimeout is the timeout for a response to be returned from the external processor
	MessageTimeout *metav1.Duration `json:"messageTimeout,omitempty" yaml:"messageTimeout,omitempty"`

	// FailOpen defines if requests or responses that cannot be processed due to connectivity to the
	// external processor are terminated or passed-through.
	FailOpen *bool `json:"failOpen,omitempty" yaml:"failOpen,omitempty"`

	// RequestHeaderProcessing Defines if request headers are processed
	RequestHeaderProcessing bool `json:"requestHeaderProcessing,omitempty" yaml:"requestHeaderProcessing,omitempty"`

	// RequestBodyProcessingMode Defines request body processing
	RequestBodyProcessingMode *ExtProcBodyProcessingMode `json:"requestBodyProcessingMode,omitempty" yaml:"requestBodyProcessingMode,omitempty"`

	// ResponseHeaderProcessingMode Defines if response headers are processed
	ResponseHeaderProcessing bool `json:"responseHeaderProcessing,omitempty" yaml:"responseHeaderProcessing,omitempty"`

	// ResponseBodyProcessingMode Defines response body processing
	ResponseBodyProcessingMode *ExtProcBodyProcessingMode `json:"responseBodyProcessingMode,omitempty" yaml:"responseBodyProcessingMode,omitempty"`

	// RequestAttributes defines which envoy attributes are provided as context to external processor
	// when processing requests
	RequestAttributes []string `json:"requestAttributes,omitempty" yaml:"requestAttributes,omitempty"`

	// ResponseAttributes defines which envoy attributes are provided as context to external processor
	// when processing responses
	ResponseAttributes []string `json:"responseAttributes,omitempty" yaml:"responseAttributes,omitempty"`

	// ForwardingMetadataNamespaces are metadata namespaces forwarded to external processor
	ForwardingMetadataNamespaces []string `json:"forwardingMetadataNamespaces,omitempty" yaml:"forwardingMetadataNamespaces,omitempty"`

	// ReceivingMetadataNamespaces are metadata namespaces updatable by external processor
	ReceivingMetadataNamespaces []string `json:"receivingMetadataNamespaces,omitempty" yaml:"receivingMetadataNamespaces,omitempty"`

	// AllowModeOverride allows the external processor to modify the processing mode.
	AllowModeOverride bool `json:"allowModeOverride,omitempty" yaml:"allowModeOverride,omitempty"`
}

// Lua holds the information associated with Lua extensions
// +k8s:deepcopy-gen=true
type Lua struct {
	// Name is a unique name for the LUa configuration.
	// The xds translator only generates one Lua filter for each unique name
	Name string
	// Code is the Lua source code
	Code *string
}

// Wasm holds the information associated with the Wasm extensions.
// +k8s:deepcopy-gen=true
type Wasm struct {
	// Name is a unique name for an Wasm configuration.
	// The xds translator only generates one Wasm filter for each unique name.
	Name string `json:"name"`

	// RootID is a unique ID for a set of extensions in a VM which will share a
	// RootContext and Contexts if applicable (e.g., an Wasm HttpFilter and an Wasm AccessLog).
	// If left blank, all extensions with a blank root_id with the same vm_id will share Context(s).
	RootID *string `json:"rootID,omitempty"`

	// WasmName is used to identify the Wasm extension if multiple extensions are
	// handled by the same vm_id and root_id.
	// It's also used for logging/debugging.
	WasmName string `json:"wasmName"`

	// Config is the configuration for the Wasm extension.
	// This configuration will be passed as a JSON string to the Wasm extension.
	Config *apiextensionsv1.JSON `json:"config"`

	// FailOpen is a switch used to control the behavior when a fatal error occurs
	// during the initialization or the execution of the Wasm extension.
	FailOpen bool `json:"failOpen"`

	// Code is the HTTP Wasm code source.
	// Envoy only supports HTTP Wasm code source. EG downloads the Wasm code from the
	// original URL(either an HTTP URL or an OCI image) and serves it through the
	// local HTTP server.
	Code *HTTPWasmCode `json:"httpWasmCode,omitempty"`

	// HostKeys is a list of keys for environment variables from the host envoy process
	// that should be passed into the Wasm VM.
	HostKeys []string `json:"hostKeys,omitempty"`
}

// HTTPWasmCode holds the information associated with the HTTP Wasm code source.
// +k8s:deepcopy-gen=true
type HTTPWasmCode struct {
	// ServingURL is the URL of the Wasm code served by the local EG HTTP server.
	ServingURL string `json:"servingURL"`

	// SHA256 checksum that will be used by the Envoy to verify the Wasm code.
	// It's different from the digest of the OCI image.
	SHA256 string `json:"sha256"`

	// OriginalURL is the original downloading URL of the Wasm code.
	// Note: This field is just used for testing. It's not used to generate the Envoy configuration.
	OriginalURL string `json:"originalDownloadingURL"`
}

// DestinationFilters contains HTTP filters that will be used with the DestinationSetting.
// +k8s:deepcopy-gen=true
type DestinationFilters struct {
	// AddRequestHeaders defines header/value sets to be added to the headers of requests.
	AddRequestHeaders []AddHeader `json:"addRequestHeaders,omitempty" yaml:"addRequestHeaders,omitempty"`
	// RemoveRequestHeaders defines a list of headers to be removed from requests.
	RemoveRequestHeaders []string `json:"removeRequestHeaders,omitempty" yaml:"removeRequestHeaders,omitempty"`
	// AddResponseHeaders defines header/value sets to be added to the headers of response.
	AddResponseHeaders []AddHeader `json:"addResponseHeaders,omitempty" yaml:"addResponseHeaders,omitempty"`
	// RemoveResponseHeaders defines a list of headers to be removed from response.
	RemoveResponseHeaders []string `json:"removeResponseHeaders,omitempty" yaml:"removeResponseHeaders,omitempty"`
}

// ResourceMetadata is metadata from the provider resource that is translated to an envoy resource
// +k8s:deepcopy-gen=true
type ResourceMetadata struct {
	// Kind is the kind of the resource
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	// Name is the name of the resource
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Namespace is the namespace of the resource
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	// Annotations are the annotations of the resource
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// SectionName is the name of a section of a resource
	SectionName string `json:"sectionName,omitempty" yaml:"sectionName,omitempty"`
}
