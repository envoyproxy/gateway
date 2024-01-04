// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"cmp"
	"errors"
	"net"
	"reflect"

	"github.com/tetratelabs/multierror"
	"golang.org/x/exp/slices"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/validation"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	egv1a1validation "github.com/envoyproxy/gateway/api/v1alpha1/validation"
)

var (
	ErrListenerNameEmpty             = errors.New("field Name must be specified")
	ErrListenerAddressInvalid        = errors.New("field Address must be a valid IP address")
	ErrListenerPortInvalid           = errors.New("field Port specified is invalid")
	ErrHTTPListenerHostnamesEmpty    = errors.New("field Hostnames must be specified with at least a single hostname entry")
	ErrTCPListenerSNIsEmpty          = errors.New("field SNIs must be specified with at least a single server name entry")
	ErrTLSServerCertEmpty            = errors.New("field ServerCertificate must be specified")
	ErrTLSPrivateKey                 = errors.New("field PrivateKey must be specified")
	ErrHTTPRouteNameEmpty            = errors.New("field Name must be specified")
	ErrHTTPRouteHostnameEmpty        = errors.New("field Hostname must be specified")
	ErrDestinationNameEmpty          = errors.New("field Name must be specified")
	ErrDestEndpointHostInvalid       = errors.New("field Address must be a valid IP or FQDN address")
	ErrDestEndpointPortInvalid       = errors.New("field Port specified is invalid")
	ErrStringMatchConditionInvalid   = errors.New("only one of the Exact, Prefix, SafeRegex or Distinct fields must be set")
	ErrStringMatchNameIsEmpty        = errors.New("field Name must be specified")
	ErrDirectResponseStatusInvalid   = errors.New("only HTTP status codes 100 - 599 are supported for DirectResponse")
	ErrRedirectUnsupportedStatus     = errors.New("only HTTP status codes 301 and 302 are supported for redirect filters")
	ErrRedirectUnsupportedScheme     = errors.New("only http and https are supported for the scheme in redirect filters")
	ErrHTTPPathModifierDoubleReplace = errors.New("redirect filter cannot have a path modifier that supplies both fullPathReplace and prefixMatchReplace")
	ErrHTTPPathModifierNoReplace     = errors.New("redirect filter cannot have a path modifier that does not supply either fullPathReplace or prefixMatchReplace")
	ErrAddHeaderEmptyName            = errors.New("header modifier filter cannot configure a header without a name to be added")
	ErrAddHeaderDuplicate            = errors.New("header modifier filter attempts to add the same header more than once (case insensitive)")
	ErrRemoveHeaderDuplicate         = errors.New("header modifier filter attempts to remove the same header more than once (case insensitive)")
	ErrLoadBalancerInvalid           = errors.New("loadBalancer setting is invalid, only one setting can be set")
)

// Xds holds the intermediate representation of a Gateway and is
// used by the xDS Translator to convert it into xDS resources.
// +k8s:deepcopy-gen=true
type Xds struct {
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
func (x Xds) Validate() error {
	var errs error
	for _, http := range x.HTTP {
		if err := http.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	for _, tcp := range x.TCP {
		if err := tcp.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	for _, udp := range x.UDP {
		if err := udp.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (x Xds) GetHTTPListener(name string) *HTTPListener {
	for _, listener := range x.HTTP {
		if listener.Name == name {
			return listener
		}
	}
	return nil
}

func (x Xds) GetTCPListener(name string) *TCPListener {
	for _, listener := range x.TCP {
		if listener.Name == name {
			return listener
		}
	}
	return nil
}

func (x Xds) GetUDPListener(name string) *UDPListener {
	for _, listener := range x.UDP {
		if listener.Name == name {
			return listener
		}
	}
	return nil
}

// Printable returns a deep copy of the resource that can be safely logged.
func (x Xds) Printable() *Xds {
	out := x.DeepCopy()
	for _, listener := range out.HTTP {
		// Omit field
		listener.TLS = nil

		for _, route := range listener.Routes {
			// Omit field
			if route.OIDC != nil {
				route.OIDC.ClientSecret = []byte{}
			}
			if route.BasicAuth != nil {
				route.BasicAuth.Users = []byte{}
			}
		}
	}
	return out
}

// HTTPListener holds the listener configuration.
// +k8s:deepcopy-gen=true
type HTTPListener struct {
	// Name of the HttpListener
	Name string `json:"name" yaml:"name"`
	// Address that the listener should listen on.
	Address string `json:"address" yaml:"address"`
	// Port on which the service can be expected to be accessed by clients.
	Port uint32 `json:"port" yaml:"port"`
	// Hostnames (Host/Authority header value) with which the service can be expected to be accessed by clients.
	// This field is required. Wildcard hosts are supported in the suffix or prefix form.
	// Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-virtualhost
	// for more info.
	Hostnames []string `json:"hostnames" yaml:"hostnames"`
	// Tls certificate info. If omitted, the gateway will expose a plain text HTTP server.
	TLS []*TLSListenerConfig `json:"tls,omitempty" yaml:"tls,omitempty"`
	// Routes associated with HTTP traffic to the service.
	Routes []*HTTPRoute `json:"routes,omitempty" yaml:"routes,omitempty"`
	// IsHTTP2 is set if the listener is configured to serve HTTP2 traffic,
	// grpc-web and grpc-stats are also enabled if this is set.
	IsHTTP2 bool `json:"isHTTP2" yaml:"isHTTP2"`
	// TCPKeepalive configuration for the listener
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty" yaml:"tcpKeepalive,omitempty"`
	// SuppressEnvoyHeaders controls if "x-envoy-" headers are suppressed by the HTTP Router filter
	// Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/router/v3/router.proto#extensions-filters-http-router-v3-router
	SuppressEnvoyHeaders bool `json:"suppressEnvoyHeaders,omitempty" yaml:"suppressEnvoyHeaders,omitempty"`
	// EnableProxyProtocol enables the listener to interpret proxy protocol header
	EnableProxyProtocol bool `json:"enableProxyProtocol,omitempty" yaml:"enableProxyProtocol,omitempty"`
	// HTTP3 provides HTTP/3 configuration on the listener.
	// +optional
	HTTP3 *HTTP3Settings `json:"http3,omitempty"`
}

// Validate the fields within the HTTPListener structure
func (h HTTPListener) Validate() error {
	var errs error
	if h.Name == "" {
		errs = multierror.Append(errs, ErrListenerNameEmpty)
	}
	if ip := net.ParseIP(h.Address); ip == nil {
		errs = multierror.Append(errs, ErrListenerAddressInvalid)
	}
	if h.Port == 0 {
		errs = multierror.Append(errs, ErrListenerPortInvalid)
	}
	if len(h.Hostnames) == 0 {
		errs = multierror.Append(errs, ErrHTTPListenerHostnamesEmpty)
	}
	if h.TLS != nil {
		for t := range h.TLS {
			if err := h.TLS[t].Validate(); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	for _, route := range h.Routes {
		if err := route.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

// TLSListenerConfig holds the configuration for downstream TLS context.
// +k8s:deepcopy-gen=true
type TLSListenerConfig struct {
	// Name of the Secret object.
	Name string `json:"name" yaml:"name"`
	// ServerCertificate of the server.
	ServerCertificate []byte `json:"serverCertificate,omitempty" yaml:"serverCertificate,omitempty"`
	// PrivateKey for the server.
	PrivateKey []byte `json:"privateKey,omitempty" yaml:"privateKey,omitempty"`
}

// Validate the fields within the TLSListenerConfig structure
func (t TLSListenerConfig) Validate() error {
	var errs error
	if len(t.ServerCertificate) == 0 {
		errs = multierror.Append(errs, ErrTLSServerCertEmpty)
	}
	if len(t.PrivateKey) == 0 {
		errs = multierror.Append(errs, ErrTLSPrivateKey)
	}
	return errs
}

// BackendWeights stores the weights of valid and invalid backends for the route so that 500 error responses can be returned in the same proportions
type BackendWeights struct {
	Valid   uint32 `json:"valid" yaml:"valid"`
	Invalid uint32 `json:"invalid" yaml:"invalid"`
}

// HTTPRoute holds the route information associated with the HTTP Route
// +k8s:deepcopy-gen=true
type HTTPRoute struct {
	// Name of the HTTPRoute
	Name string `json:"name" yaml:"name"`
	// Hostname that the route matches against
	Hostname string `json:"hostname" yaml:"hostname,omitempty"`
	// PathMatch defines the match conditions on the path.
	PathMatch *StringMatch `json:"pathMatch,omitempty" yaml:"pathMatch,omitempty"`
	// HeaderMatches define the match conditions on the request headers for this route.
	HeaderMatches []*StringMatch `json:"headerMatches,omitempty" yaml:"headerMatches,omitempty"`
	// QueryParamMatches define the match conditions on the query parameters.
	QueryParamMatches []*StringMatch `json:"queryParamMatches,omitempty" yaml:"queryParamMatches,omitempty"`
	// DestinationWeights stores the weights of valid and invalid backends for the route so that 500 error responses can be returned in the same proportions
	BackendWeights BackendWeights `json:"backendWeights" yaml:"backendWeights"`
	// AddRequestHeaders defines header/value sets to be added to the headers of requests.
	AddRequestHeaders []AddHeader `json:"addRequestHeaders,omitempty" yaml:"addRequestHeaders,omitempty"`
	// RemoveRequestHeaders defines a list of headers to be removed from requests.
	RemoveRequestHeaders []string `json:"removeRequestHeaders,omitempty" yaml:"removeRequestHeaders,omitempty"`
	// AddResponseHeaders defines header/value sets to be added to the headers of response.
	AddResponseHeaders []AddHeader `json:"addResponseHeaders,omitempty" yaml:"addResponseHeaders,omitempty"`
	// RemoveResponseHeaders defines a list of headers to be removed from response.
	RemoveResponseHeaders []string `json:"removeResponseHeaders,omitempty" yaml:"removeResponseHeaders,omitempty"`
	// Direct responses to be returned for this route. Takes precedence over Destinations and Redirect.
	DirectResponse *DirectResponse `json:"directResponse,omitempty" yaml:"directResponse,omitempty"`
	// Redirections to be returned for this route. Takes precedence over Destinations.
	Redirect *Redirect `json:"redirect,omitempty" yaml:"redirect,omitempty"`
	// Destination that requests to this HTTPRoute will be mirrored to
	Mirrors []*RouteDestination `json:"mirrors,omitempty" yaml:"mirrors,omitempty"`
	// Destination associated with this matched route.
	Destination *RouteDestination `json:"destination,omitempty" yaml:"destination,omitempty"`
	// Rewrite to be changed for this route.
	URLRewrite *URLRewrite `json:"urlRewrite,omitempty" yaml:"urlRewrite,omitempty"`
	// RateLimit defines the more specific match conditions as well as limits for ratelimiting
	// the requests on this route.
	RateLimit *RateLimit `json:"rateLimit,omitempty" yaml:"rateLimit,omitempty"`
	// Timeout is the time until which entire response is received from the upstream.
	Timeout *metav1.Duration `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	// load balancer policy to use when routing to the backend endpoints.
	LoadBalancer *LoadBalancer `json:"loadBalancer,omitempty" yaml:"loadBalancer,omitempty"`
	// CORS policy for the route.
	CORS *CORS `json:"cors,omitempty" yaml:"cors,omitempty"`
	// JWT defines the schema for authenticating HTTP requests using JSON Web Tokens (JWT).
	JWT *JWT `json:"jwt,omitempty" yaml:"jwt,omitempty"`
	// OIDC defines the schema for authenticating HTTP requests using OpenID Connect (OIDC).
	OIDC *OIDC `json:"oidc,omitempty" yaml:"oidc,omitempty"`
	// Proxy Protocol Settings
	ProxyProtocol *ProxyProtocol `json:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty"`
	// BasicAuth defines the schema for the HTTP Basic Authentication.
	BasicAuth *BasicAuth `json:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
	// FaultInjection defines the schema for injecting faults into HTTP requests.
	FaultInjection *FaultInjection `json:"faultInjection,omitempty" yaml:"faultInjection,omitempty"`
	// ExtensionRefs holds unstructured resources that were introduced by an extension and used on the HTTPRoute as extensionRef filters
	ExtensionRefs []*UnstructuredRef `json:"extensionRefs,omitempty" yaml:"extensionRefs,omitempty"`
	// Circuit Breaker Settings
	CircuitBreaker *CircuitBreaker `json:"circuitBreaker,omitempty" yaml:"circuitBreaker,omitempty"`
}

// UnstructuredRef holds unstructured data for an arbitrary k8s resource introduced by an extension
// Envoy Gateway does not need to know about the resource types in order to store and pass the data for these objects
// to an extension.
//
// +k8s:deepcopy-gen=true
type UnstructuredRef struct {
	Object *unstructured.Unstructured `json:"object,omitempty" yaml:"object,omitempty"`
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
	// Providers defines a list of JSON Web Token (JWT) authentication providers.
	Providers []egv1a1.JWTProvider `json:"providers,omitempty" yaml:"providers,omitempty"`
}

// OIDC defines the schema for authenticating HTTP requests using
// OpenID Connect (OIDC).
//
// +k8s:deepcopy-gen=true
type OIDC struct {
	// The OIDC Provider configuration.
	Provider OIDCProvider `json:"provider" yaml:"provider"`

	// The OIDC client ID to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	ClientID string `json:"clientID" yaml:"clientID"`

	// The OIDC client secret to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// This is an Opaque secret. The client secret should be stored in the key "client-secret".

	ClientSecret []byte `json:"clientSecret,omitempty" yaml:"clientSecret,omitempty"`

	// The OIDC scopes to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	Scopes []string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// BasicAuth defines the schema for the HTTP Basic Authentication.
//
// +k8s:deepcopy-gen=true
type BasicAuth struct {
	// The username-password pairs in htpasswd format.
	Users []byte `json:"users,omitempty" yaml:"users,omitempty"`
}

type OIDCProvider struct {
	// The OIDC Provider's [authorization endpoint](https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint).
	AuthorizationEndpoint string `json:"authorizationEndpoint,omitempty"`

	// The OIDC Provider's [token endpoint](https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint).
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`
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

// Validate the fields within the HTTPRoute structure
func (h HTTPRoute) Validate() error {
	var errs error
	if h.Name == "" {
		errs = multierror.Append(errs, ErrHTTPRouteNameEmpty)
	}
	if h.Hostname == "" {
		errs = multierror.Append(errs, ErrHTTPRouteHostnameEmpty)
	}
	if h.PathMatch != nil {
		if err := h.PathMatch.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	for _, hMatch := range h.HeaderMatches {
		if err := hMatch.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	for _, qMatch := range h.QueryParamMatches {
		if err := qMatch.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if h.Destination != nil {
		if err := h.Destination.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if h.Redirect != nil {
		if err := h.Redirect.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if h.DirectResponse != nil {
		if err := h.DirectResponse.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if h.URLRewrite != nil {
		if err := h.URLRewrite.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if h.Mirrors != nil {
		for _, mirror := range h.Mirrors {
			if err := mirror.Validate(); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	if len(h.AddRequestHeaders) > 0 {
		occurred := map[string]bool{}
		for _, header := range h.AddRequestHeaders {
			if err := header.Validate(); err != nil {
				errs = multierror.Append(errs, err)
			}
			if !occurred[header.Name] {
				occurred[header.Name] = true
			} else {
				errs = multierror.Append(errs, ErrAddHeaderDuplicate)
				break
			}
		}
	}
	if len(h.RemoveRequestHeaders) > 0 {
		occurred := map[string]bool{}
		for _, header := range h.RemoveRequestHeaders {
			if !occurred[header] {
				occurred[header] = true
			} else {
				errs = multierror.Append(errs, ErrRemoveHeaderDuplicate)
				break
			}
		}
	}
	if len(h.AddResponseHeaders) > 0 {
		occurred := map[string]bool{}
		for _, header := range h.AddResponseHeaders {
			if err := header.Validate(); err != nil {
				errs = multierror.Append(errs, err)
			}
			if !occurred[header.Name] {
				occurred[header.Name] = true
			} else {
				errs = multierror.Append(errs, ErrAddHeaderDuplicate)
				break
			}
		}
	}
	if len(h.RemoveResponseHeaders) > 0 {
		occurred := map[string]bool{}
		for _, header := range h.RemoveResponseHeaders {
			if !occurred[header] {
				occurred[header] = true
			} else {
				errs = multierror.Append(errs, ErrRemoveHeaderDuplicate)
				break
			}
		}
	}
	if h.LoadBalancer != nil {
		if err := h.LoadBalancer.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	if h.JWT != nil {
		if err := h.JWT.validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

func (j *JWT) validate() error {
	var errs error

	if err := egv1a1validation.ValidateJWTProvider(j.Providers); err != nil {
		errs = multierror.Append(errs, err)
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
func (r RouteDestination) Validate() error {
	var errs error
	if len(r.Name) == 0 {
		errs = multierror.Append(errs, ErrDestinationNameEmpty)
	}
	for _, s := range r.Settings {
		if err := s.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}

// DestinationSetting holds the settings associated with the destination
// +kubebuilder:object:generate=true
type DestinationSetting struct {
	// Weight associated with this destination.
	// Note: Weight is not used in TCP/UDP route.
	Weight *uint32 `json:"weight,omitempty" yaml:"weight,omitempty"`
	// Protocol associated with this destination/port.
	Protocol  AppProtocol            `json:"protocol" yaml:"protocol"`
	Endpoints []*DestinationEndpoint `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	// AddressTypeState specifies the state of DestinationEndpoint address type.
	AddressType *DestinationAddressType `json:"addressType,omitempty" yaml:"addressType,omitempty"`
}

// Validate the fields within the RouteDestination structure
func (d DestinationSetting) Validate() error {
	var errs error
	for _, ep := range d.Endpoints {
		if err := ep.Validate(); err != nil {
			errs = multierror.Append(errs, err)
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
}

// Validate the fields within the DestinationEndpoint structure
func (d DestinationEndpoint) Validate() error {
	var errs error

	err := validation.IsDNS1123Subdomain(d.Host)
	ip := net.ParseIP(d.Host)

	if err != nil && ip == nil {
		errs = multierror.Append(errs, ErrDestEndpointHostInvalid)
	}

	if d.Port == 0 {
		errs = multierror.Append(errs, ErrDestEndpointPortInvalid)
	}

	return errs
}

// NewDestEndpoint creates a new DestinationEndpoint.
func NewDestEndpoint(host string, port uint32) *DestinationEndpoint {
	return &DestinationEndpoint{
		Host: host,
		Port: port,
	}
}

// AddHeader configures a header to be added to a request or response.
// +k8s:deepcopy-gen=true
type AddHeader struct {
	Name   string `json:"name" yaml:"name"`
	Value  string `json:"value" yaml:"value"`
	Append bool   `json:"append" yaml:"append"`
}

// Validate the fields within the AddHeader structure
func (h AddHeader) Validate() error {
	var errs error
	if h.Name == "" {
		errs = multierror.Append(errs, ErrAddHeaderEmptyName)
	}

	return errs
}

// DirectResponse holds the details for returning a body and status code for a route.
// +k8s:deepcopy-gen=true
type DirectResponse struct {
	// Body configures the body of the direct response. Currently only a string response
	// is supported, but in the future a config.core.v3.DataSource may replace it.
	Body *string `json:"body,omitempty" yaml:"body,omitempty"`
	// StatusCode will be used for the direct response's status code.
	StatusCode uint32 `json:"statusCode" yaml:"statusCode"`
}

// Validate the fields within the DirectResponse structure
func (r DirectResponse) Validate() error {
	var errs error
	if status := r.StatusCode; status > 599 || status < 100 {
		errs = multierror.Append(errs, ErrDirectResponseStatusInvalid)
	}

	return errs
}

// URLRewrite holds the details for how to rewrite a request
// +k8s:deepcopy-gen=true
type URLRewrite struct {
	// Path contains config for rewriting the path of the request.
	Path *HTTPPathModifier `json:"path,omitempty" yaml:"path,omitempty"`
	// Hostname configures the replacement of the request's hostname.
	Hostname *string `json:"hostname,omitempty" yaml:"hostname,omitempty"`
}

// Validate the fields within the URLRewrite structure
func (r URLRewrite) Validate() error {
	var errs error

	if r.Path != nil {
		if err := r.Path.Validate(); err != nil {
			errs = multierror.Append(errs, err)
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
			errs = multierror.Append(errs, ErrRedirectUnsupportedScheme)
		}
	}

	if r.Path != nil {
		if err := r.Path.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if r.StatusCode != nil {
		if *r.StatusCode != 301 && *r.StatusCode != 302 {
			errs = multierror.Append(errs, ErrRedirectUnsupportedStatus)
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
		errs = multierror.Append(errs, ErrHTTPPathModifierDoubleReplace)
	}

	if r.FullReplace == nil && r.PrefixMatchReplace == nil {
		errs = multierror.Append(errs, ErrHTTPPathModifierNoReplace)
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
			errs = multierror.Append(errs, ErrStringMatchNameIsEmpty)
		}
		matchCount++
	}

	if matchCount != 1 {
		errs = multierror.Append(errs, ErrStringMatchConditionInvalid)
	}

	return errs
}

// TCPListener holds the TCP listener configuration.
// +k8s:deepcopy-gen=true
type TCPListener struct {
	// Name of the TCPListener
	Name string `json:"name" yaml:"name"`
	// Address that the listener should listen on.
	Address string `json:"address" yaml:"address"`
	// Port on which the service can be expected to be accessed by clients.
	Port uint32 `json:"port" yaml:"port"`
	// TLS holds information for configuring TLS on a listener
	TLS *TLS `json:"tls,omitempty" yaml:"tls,omitempty"`
	// Destinations associated with TCP traffic to the service.
	Destination *RouteDestination `json:"destination,omitempty" yaml:"destination,omitempty"`
	// TCPKeepalive configuration for the listener
	TCPKeepalive *TCPKeepalive `json:"tcpKeepalive,omitempty" yaml:"tcpKeepalive,omitempty"`
}

// TLS holds information for configuring TLS on a listener
// +k8s:deepcopy-gen=true
type TLS struct {
	// TLS information required for TLS Passthrough, If provided, incoming
	// connections' server names are inspected and routed to backends accordingly.
	Passthrough *TLSInspectorConfig `json:"passthrough,omitempty" yaml:"passthrough,omitempty"`
	// TLS information required for TLS Termination
	Terminate []*TLSListenerConfig `json:"terminate,omitempty" yaml:"terminate,omitempty"`
}

// Validate the fields within the TCPListener structure
func (h TCPListener) Validate() error {
	var errs error
	if h.Name == "" {
		errs = multierror.Append(errs, ErrListenerNameEmpty)
	}
	if ip := net.ParseIP(h.Address); ip == nil {
		errs = multierror.Append(errs, ErrListenerAddressInvalid)
	}
	if h.Port == 0 {
		errs = multierror.Append(errs, ErrListenerPortInvalid)
	}
	if h.TLS != nil && h.TLS.Passthrough != nil {
		if err := h.TLS.Passthrough.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if h.TLS != nil && h.TLS.Terminate != nil {
		for t := range h.TLS.Terminate {
			if err := h.TLS.Terminate[t].Validate(); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	if h.Destination != nil {
		if err := h.Destination.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

// TLSInspectorConfig holds the configuration required for inspecting TLS
// passthrough connections.
// +k8s:deepcopy-gen=true
type TLSInspectorConfig struct {
	// Server names that are compared against the server names of a new connection.
	// Wildcard hosts are supported in the prefix form. Partial wildcards are not
	// supported, and values like *w.example.com are invalid.
	// SNIs are used only in case of TLS Passthrough.
	SNIs []string `json:"snis,omitempty" yaml:"snis,omitempty"`
}

func (t TLSInspectorConfig) Validate() error {
	var errs error
	if len(t.SNIs) == 0 {
		errs = multierror.Append(errs, ErrTCPListenerSNIsEmpty)
	}
	return errs
}

// UDPListener holds the UDP listener configuration.
// +k8s:deepcopy-gen=true
type UDPListener struct {
	// Name of the UDPListener
	Name string `json:"name" yaml:"name"`
	// Address that the listener should listen on.
	Address string `json:"address" yaml:"address"`
	// Port on which the service can be expected to be accessed by clients.
	Port uint32 `json:"port" yaml:"port"`
	// Destination associated with UDP traffic to the service.
	Destination *RouteDestination `json:"destination,omitempty" yaml:"destination,omitempty"`
}

// Validate the fields within the UDPListener structure
func (h UDPListener) Validate() error {
	var errs error
	if h.Name == "" {
		errs = multierror.Append(errs, ErrListenerNameEmpty)
	}
	if ip := net.ParseIP(h.Address); ip == nil {
		errs = multierror.Append(errs, ErrListenerAddressInvalid)
	}
	if h.Port == 0 {
		errs = multierror.Append(errs, ErrListenerPortInvalid)
	}
	if h.Destination != nil {
		if err := h.Destination.Validate(); err != nil {
			errs = multierror.Append(errs, err)
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
}

type CIDRMatch struct {
	CIDR    string `json:"cidr" yaml:"cidr"`
	IPv6    bool   `json:"ipv6" yaml:"ipv6"`
	MaskLen int    `json:"maskLen" yaml:"maskLen"`
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

// AccessLog holds the access logging configuration.
// +k8s:deepcopy-gen=true
type AccessLog struct {
	Text          []*TextAccessLog          `json:"text,omitempty" yaml:"text,omitempty"`
	JSON          []*JSONAccessLog          `json:"json,omitempty" yaml:"json,omitempty"`
	OpenTelemetry []*OpenTelemetryAccessLog `json:"openTelemetry,omitempty" yaml:"openTelemetry,omitempty"`
}

// TextAccessLog holds the configuration for text access logging.
// +k8s:deepcopy-gen=true
type TextAccessLog struct {
	Format *string `json:"format,omitempty" yaml:"format,omitempty"`
	Path   string  `json:"path" yaml:"path"`
}

// JSONAccessLog holds the configuration for JSON access logging.
// +k8s:deepcopy-gen=true
type JSONAccessLog struct {
	JSON map[string]string `json:"json,omitempty" yaml:"json,omitempty"`
	Path string            `json:"path" yaml:"path"`
}

// OpenTelemetryAccessLog holds the configuration for OpenTelemetry access logging.
// +k8s:deepcopy-gen=true
type OpenTelemetryAccessLog struct {
	Text       *string           `json:"text,omitempty" yaml:"text,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Host       string            `json:"host" yaml:"host"`
	Port       uint32            `json:"port" yaml:"port"`
	Resources  map[string]string `json:"resources,omitempty" yaml:"resources,omitempty"`
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
	Status *egv1a1.EnvoyPatchPolicyStatus `json:"status,omitempty" yaml:"status,omitempty"`
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

// JSONPatchOperation defines the JSON Patch Operation as defined in
// https://datatracker.ietf.org/doc/html/rfc6902
// +k8s:deepcopy-gen=true
type JSONPatchOperation struct {
	// Op is the type of operation to perform
	Op string `json:"op" yaml:"op"`
	// Path is the location of the target document/field where the operation will be performed
	// Refer to https://datatracker.ietf.org/doc/html/rfc6901 for more details.
	Path string `json:"path" yaml:"path"`
	// Value is the new value of the path location.
	Value apiextensionsv1.JSON `json:"value" yaml:"value"`
}

// Tracing defines the configuration for tracing a Envoy xDS Resource
// +k8s:deepcopy-gen=true
type Tracing struct {
	ServiceName string `json:"serviceName"`

	egv1a1.ProxyTracing
}

// Metrics defines the configuration for metrics generated by Envoy
// +k8s:deepcopy-gen=true
type Metrics struct {
	EnableVirtualHostStats bool `json:"enableVirtualHostStats" yaml:"enableVirtualHostStats"`
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
		errs = multierror.Append(errs, ErrLoadBalancerInvalid)
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
	SourceIP *bool `json:"sourceIP,omitempty" yaml:"sourceIP,omitempty"`
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
}
