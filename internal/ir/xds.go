// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

import (
	"errors"
	"net"

	"github.com/tetratelabs/multierror"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
)

var (
	ErrListenerNameEmpty             = errors.New("field Name must be specified")
	ErrListenerAddressInvalid        = errors.New("field Address must be a valid IP address")
	ErrListenerPortInvalid           = errors.New("field Port specified is invalid")
	ErrHTTPListenerHostnamesEmpty    = errors.New("field Hostnames must be specified with at least a single hostname entry")
	ErrTCPListenesSNIsEmpty          = errors.New("field SNIs must be specified with at least a single server name entry")
	ErrTLSServerCertEmpty            = errors.New("field ServerCertificate must be specified")
	ErrTLSPrivateKey                 = errors.New("field PrivateKey must be specified")
	ErrHTTPRouteNameEmpty            = errors.New("field Name must be specified")
	ErrHTTPRouteMatchEmpty           = errors.New("either PathMatch, HeaderMatches or QueryParamMatches fields must be specified")
	ErrRouteDestinationHostInvalid   = errors.New("field Address must be a valid IP address")
	ErrRouteDestinationPortInvalid   = errors.New("field Port specified is invalid")
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
	ErrRequestAuthenRequiresJwt      = errors.New("jwt field is required when request authentication is set")
)

// Xds holds the intermediate representation of a Gateway and is
// used by the xDS Translator to convert it into xDS resources.
// +k8s:deepcopy-gen=true
type Xds struct {
	// HTTP listeners exposed by the gateway.
	HTTP []*HTTPListener
	// TCP Listeners exposed by the gateway.
	TCP []*TCPListener
	// UDP Listeners exposed by the gateway.
	UDP []*UDPListener
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
	}
	return out
}

// HTTPListener holds the listener configuration.
// +k8s:deepcopy-gen=true
type HTTPListener struct {
	// Name of the HttpListener
	Name string
	// Address that the listener should listen on.
	Address string
	// Port on which the service can be expected to be accessed by clients.
	Port uint32
	// Hostnames (Host/Authority header value) with which the service can be expected to be accessed by clients.
	// This field is required. Wildcard hosts are supported in the suffix or prefix form.
	// Refer to https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-virtualhost
	// for more info.
	Hostnames []string
	// Tls certificate info. If omitted, the gateway will expose a plain text HTTP server.
	TLS *TLSListenerConfig
	// Routes associated with HTTP traffic to the service.
	Routes []*HTTPRoute
	// IsHTTP2 is set if the upstream client as well as the downstream server are configured to serve HTTP2 traffic.
	IsHTTP2 bool
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
		if err := h.TLS.Validate(); err != nil {
			errs = multierror.Append(errs, err)
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
	// ServerCertificate of the server.
	ServerCertificate []byte
	// PrivateKey for the server.
	PrivateKey []byte
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

// DestinationWeights stores the weights of valid and invalid backends for the route so that 500 error responses can be returned in the same proportions
type BackendWeights struct {
	Valid   uint32
	Invalid uint32
}

// HTTPRoute holds the route information associated with the HTTP Route
// +k8s:deepcopy-gen=true
type HTTPRoute struct {
	// Name of the HTTPRoute
	Name string
	// PathMatch defines the match conditions on the path.
	PathMatch *StringMatch
	// HeaderMatches define the match conditions on the request headers for this route.
	HeaderMatches []*StringMatch
	// QueryParamMatches define the match conditions on the query parameters.
	QueryParamMatches []*StringMatch
	// DestinationWeights stores the weights of valid and invalid backends for the route so that 500 error responses can be returned in the same proportions
	BackendWeights BackendWeights
	// AddRequestHeaders defines header/value sets to be added to the headers of requests.
	AddRequestHeaders []AddHeader
	// RemoveRequestHeaders defines a list of headers to be removed from requests.
	RemoveRequestHeaders []string
	// AddResponseHeaders defines header/value sets to be added to the headers of response.
	AddResponseHeaders []AddHeader
	// RemoveResponseHeaders defines a list of headers to be removed from response.
	RemoveResponseHeaders []string
	// Direct responses to be returned for this route. Takes precedence over Destinations and Redirect.
	DirectResponse *DirectResponse
	// Redirections to be returned for this route. Takes precedence over Destinations.
	Redirect *Redirect
	// Destinations that requests to this HTTPRoute will be mirrored to
	Mirrors []*RouteDestination
	// Destinations associated with this matched route.
	Destinations []*RouteDestination
	// Rewrite to be changed for this route.
	URLRewrite *URLRewrite
	// RateLimit defines the more specific match conditions as well as limits for ratelimiting
	// the requests on this route.
	RateLimit *RateLimit
	// RequestAuthentication defines the schema for authenticating HTTP requests.
	RequestAuthentication *RequestAuthentication
}

// RequestAuthentication defines the schema for authenticating HTTP requests.
// Only one of "jwt" can be specified.
//
// TODO: Add support for additional request authentication providers, i.e. OIDC.
//
// +k8s:deepcopy-gen=true
type RequestAuthentication struct {
	// JWT defines the schema for authenticating HTTP requests using JSON Web Tokens (JWT).
	JWT *JwtRequestAuthentication
}

// JwtRequestAuthentication defines the schema for authenticating HTTP requests using
// JSON Web Tokens (JWT).
//
// +k8s:deepcopy-gen=true
type JwtRequestAuthentication struct {
	// Providers defines a list of JSON Web Token (JWT) authentication providers.
	Providers []egv1a1.JwtAuthenticationFilterProvider
}

// Validate the fields within the HTTPRoute structure
func (h HTTPRoute) Validate() error {
	var errs error
	if h.Name == "" {
		errs = multierror.Append(errs, ErrHTTPRouteNameEmpty)
	}
	if h.PathMatch == nil && (len(h.HeaderMatches) == 0) && (len(h.QueryParamMatches) == 0) {
		errs = multierror.Append(errs, ErrHTTPRouteMatchEmpty)
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
	for _, dest := range h.Destinations {
		if err := dest.Validate(); err != nil {
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
	for _, mirror := range h.Mirrors {
		if err := mirror.Validate(); err != nil {
			errs = multierror.Append(errs, err)
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
	if h.RequestAuthentication != nil {
		switch {
		case h.RequestAuthentication.JWT == nil:
			errs = multierror.Append(errs, ErrRequestAuthenRequiresJwt)
		default:
			if err := h.RequestAuthentication.JWT.Validate(); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	return errs
}

func (j *JwtRequestAuthentication) Validate() error {
	var errs error

	if err := validation.ValidateJwtProviders(j.Providers); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs
}

// RouteDestination holds the destination details associated with the route
type RouteDestination struct {
	// Host refers to the FQDN or IP address of the backend service.
	Host string
	// Port on the service to forward the request to.
	Port uint32
	// Weight associated with this destination.
	// Note: Weight is not used in UDP route.
	Weight uint32
}

// Validate the fields within the RouteDestination structure
func (r RouteDestination) Validate() error {
	var errs error
	// Only support IP hosts for now
	if ip := net.ParseIP(r.Host); ip == nil {
		errs = multierror.Append(errs, ErrRouteDestinationHostInvalid)
	}
	if r.Port == 0 {
		errs = multierror.Append(errs, ErrRouteDestinationPortInvalid)
	}

	return errs
}

// Add header configures a header to be added to a request or response.
// +k8s:deepcopy-gen=true
type AddHeader struct {
	Name   string
	Value  string
	Append bool
}

// Validate the fields within the AddHeader structure
func (h AddHeader) Validate() error {
	var errs error
	if h.Name == "" {
		errs = multierror.Append(errs, ErrAddHeaderEmptyName)
	}

	return errs
}

// Direct response holds the details for returning a body and status code for a route.
// +k8s:deepcopy-gen=true
type DirectResponse struct {
	// Body configures the body of the direct response. Currently only a string response
	// is supported, but in the future a config.core.v3.DataSource may replace it.
	Body *string
	// StatusCode will be used for the direct response's status code.
	StatusCode uint32
}

// Validate the fields within the DirectResponse structure
func (r DirectResponse) Validate() error {
	var errs error
	if status := r.StatusCode; status > 599 || status < 100 {
		errs = multierror.Append(errs, ErrDirectResponseStatusInvalid)
	}

	return errs
}

// Re holds the details for how to rewrite a request
// +k8s:deepcopy-gen=true
type URLRewrite struct {
	// Path contains config for rewriting the path of the request.
	Path *HTTPPathModifier
	// Hostname configures the replacement of the request's hostname.
	Hostname *string
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
	Scheme *string
	// Hostname configures the replacement of the request's hostname.
	Hostname *string
	// Path contains config for rewriting the path of the request.
	Path *HTTPPathModifier
	// Port configures the replacement of the request's port.
	Port *uint32
	// Status code configures the redirection response's status code.
	StatusCode *int32
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
	FullReplace *string
	// PrefixMatchReplace provides a string to replace the matched prefix of the request.
	PrefixMatchReplace *string
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
	Name string
	// Exact match condition.
	Exact *string
	// Prefix match condition.
	Prefix *string
	// Suffix match condition.
	Suffix *string
	// SafeRegex match condition.
	SafeRegex *string
	// Distinct match condition.
	// Used to match any and all possible unique values encountered within the Name field.
	Distinct bool
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
	Name string
	// Address that the listener should listen on.
	Address string
	// Port on which the service can be expected to be accessed by clients.
	Port uint32
	// TLS information required for TLS Passthrough, If provided, incoming
	// connections' server names are inspected and routed to backends accordingly.
	TLS *TLSInspectorConfig
	// Destinations associated with TCP traffic to the service.
	Destinations []*RouteDestination
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
	if h.TLS != nil {
		if err := h.TLS.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	for _, route := range h.Destinations {
		if err := route.Validate(); err != nil {
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
	SNIs []string
}

func (t TLSInspectorConfig) Validate() error {
	var errs error
	if len(t.SNIs) == 0 {
		errs = multierror.Append(errs, ErrTCPListenesSNIsEmpty)
	}
	return errs
}

// UDPListener holds the UDP listener configuration.
// +k8s:deepcopy-gen=true
type UDPListener struct {
	// Name of the UDPListener
	Name string
	// Address that the listener should listen on.
	Address string
	// Port on which the service can be expected to be accessed by clients.
	Port uint32
	// Destinations associated with UDP traffic to the service.
	Destinations []*RouteDestination
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
	for _, route := range h.Destinations {
		if err := route.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

// RateLimit holds the rate limiting configuration.
// +k8s:deepcopy-gen=true
type RateLimit struct {
	// Global rate limit settings.
	Global *GlobalRateLimit
}

// GlobalRateLimit holds the global rate limiting configuration.
// +k8s:deepcopy-gen=true
type GlobalRateLimit struct {
	// Rules for rate limiting.
	Rules []*RateLimitRule
}

// RateLimitRule holds the match and limit configuration for ratelimiting.
// +k8s:deepcopy-gen=true
type RateLimitRule struct {
	// HeaderMatches define the match conditions on the request headers for this route.
	HeaderMatches []*StringMatch
	// Limit holds the rate limit values.
	Limit *RateLimitValue
}

type RateLimitUnit egv1a1.RateLimitUnit

// RateLimitValue holds the
// +k8s:deepcopy-gen=true
type RateLimitValue struct {
	// Requests are the number of requests that need to be rate limited.
	Requests uint
	// Unit of rate limiting.
	Unit RateLimitUnit
}
