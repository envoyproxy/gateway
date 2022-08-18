package ir

import (
	"errors"
	"net"

	"github.com/tetratelabs/multierror"
)

var (
	ErrHTTPListenerNameEmpty       = errors.New("field Name must be specified")
	ErrHTTPListenerAddressInvalid  = errors.New("field Address must be a valid IP address")
	ErrHTTPListenerPortInvalid     = errors.New("field Port specified is invalid")
	ErrHTTPListenerHostnamesEmpty  = errors.New("field Hostnames must be specified with at least a single hostname entry")
	ErrTLSServerCertEmpty          = errors.New("field ServerCertificate must be specified")
	ErrTLSPrivateKey               = errors.New("field PrivateKey must be specified")
	ErrHTTPRouteNameEmpty          = errors.New("field Name must be specified")
	ErrHTTPRouteMatchEmpty         = errors.New("either PathMatch, HeaderMatches or QueryParamMatches fields must be specified")
	ErrRouteDestinationHostInvalid = errors.New("field Address must be a valid IP address")
	ErrRouteDestinationPortInvalid = errors.New("field Port specified is invalid")
	ErrStringMatchConditionInvalid = errors.New("only one of the Exact, Prefix or SafeRegex fields must be specified")
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
