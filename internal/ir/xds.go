package ir

import (
	"errors"
	"net"

	"github.com/tetratelabs/multierror"
)

var (
	ErrHTTPListenerNameEmpty         = errors.New("field Name must be specified")
	ErrHTTPListenerAddressInvalid    = errors.New("field Address must be a valid IP address")
	ErrHTTPListenerPortInvalid       = errors.New("field Port specified is invalid")
	ErrHTTPListenerHostnamesEmpty    = errors.New("field Hostnames must be specified with at least a single hostname entry")
	ErrTLSServerCertEmpty            = errors.New("field ServerCertificate must be specified")
	ErrTLSPrivateKey                 = errors.New("field PrivateKey must be specified")
	ErrHTTPRouteNameEmpty            = errors.New("field Name must be specified")
	ErrHTTPRouteMatchEmpty           = errors.New("either PathMatch, HeaderMatches or QueryParamMatches fields must be specified")
	ErrRouteDestinationHostInvalid   = errors.New("field Address must be a valid IP address")
	ErrRouteDestinationPortInvalid   = errors.New("field Port specified is invalid")
	ErrStringMatchConditionInvalid   = errors.New("only one of the Exact, Prefix or SafeRegex fields must be specified")
	ErrDirectResponseStatusInvalid   = errors.New("only HTTP status codes 100 - 599 are supported for DirectResponse")
	ErrRedirectUnsupportedStatus     = errors.New("only HTTP status codes 301 and 302 are supported for redirect filters")
	ErrRedirectUnsupportedScheme     = errors.New("only http and https are supported for the scheme in redirect filters")
	ErrHTTPPathModifierDoubleReplace = errors.New("redirect filter cannot have a path modifier that supplies both fullPathReplace and prefixMatchReplace")
	ErrHTTPPathModifierNoReplace     = errors.New("redirect filter cannot have a path modifier that does not supply either fullPathReplace or prefixMatchReplace")
	ErrAddHeaderEmptyName            = errors.New("header modifier filter cannot configure a header without a name to be added")
	ErrAddHeaderDuplicate            = errors.New("header modifier filter attempts to add the same header more than once (case insensitive)")
	ErrRemoveHeaderDuplicate         = errors.New("header modifier filter attempts to remove the same header more than once (case insensitive)")
)

// Xds holds the intermediate representation of a Gateway and is
// used by the xDS Translator to convert it into xDS resources.
// +k8s:deepcopy-gen=true
type Xds struct {
	// HTTP listeners exposed by the gateway.
	HTTP []*HTTPListener
}

// Validate the fields within the Xds structure.
func (x Xds) Validate() error {
	var errs error
	for _, http := range x.HTTP {
		if err := http.Validate(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
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
}

func (x Xds) GetListener(name string) *HTTPListener {
	for _, listener := range x.HTTP {
		if listener.Name == name {
			return listener
		}
	}
	return nil
}

// Validate the fields within the HTTPListener structure
func (h HTTPListener) Validate() error {
	var errs error
	if h.Name == "" {
		errs = multierror.Append(errs, ErrHTTPListenerNameEmpty)
	}
	if ip := net.ParseIP(h.Address); ip == nil {
		errs = multierror.Append(errs, ErrHTTPListenerAddressInvalid)
	}
	if h.Port == 0 {
		errs = multierror.Append(errs, ErrHTTPListenerPortInvalid)
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
	// AddRequestHeaders defines header/value sets to be added to the headers of requests.
	AddRequestHeaders []AddHeader
	// RemoveRequestHeaders defines a list of headers to be removed from requests.
	RemoveRequestHeaders []string
	// Direct responses to be returned for this route. Takes precedence over Destinations and Redirect.
	DirectResponse *DirectResponse
	// Redirections to be returned for this route. Takes precedence over Destinations.
	Redirect *Redirect
	// Destinations associated with this matched route.
	Destinations []*RouteDestination
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
	return errs
}

// RouteDestination holds the destination details associated with the route
type RouteDestination struct {
	// Host refers to the FQDN or IP address of the backend service.
	Host string
	// Port on the service to forward the request to.
	Port uint32
	// Weight associated with this destination.
	Weight uint32
}

// Add header configures a headder to be added to a request.
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

// StringMatch holds the various match conditions.
// Only one of Exact, Prefix or SafeRegex can be set.
// +k8s:deepcopy-gen=true
type StringMatch struct {
	// Name of the field to match on.
	Name string
	// Exact match condition.
	Exact *string
	// Prefix match condition.
	Prefix *string
	// SafeRegex match condition.
	SafeRegex *string
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
	if s.SafeRegex != nil {
		matchCount++
	}

	if matchCount != 1 {
		errs = multierror.Append(errs, ErrStringMatchConditionInvalid)
	}

	return errs
}
