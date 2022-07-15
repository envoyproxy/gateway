package ir

import (
	"errors"
	"net"
)

var (
	ErrXdsNameEmpty                = errors.New("Name must be specified.")
	ErrHTTPListenerNameEmpty       = errors.New("Name must be specified.")
	ErrHTTPListenerAddressInvalid  = errors.New("Address must be a valid IP address.")
	ErrHTTPListenerPortInvalid     = errors.New("Port specified is invalid.")
	ErrHTTPListenerHostnamesEmpty  = errors.New("Hostnames must be specified with atleast a single hostname entry.")
	ErrTLSServerCertEmpty          = errors.New("ServerCertificate must be specified.")
	ErrTLSPrivateKey               = errors.New("PrivateKey must be specified.")
	ErrHTTPRouteNameEmpty          = errors.New("Name must be specified.")
	ErrHTTPRouteMatchEmpty         = errors.New("Either PathMatch, HeaderMatches or QueryParamMatches fields must be specified.")
	ErrRouteDestinationHostInvalid = errors.New("Address must be a valid IP address.")
	ErrRouteDestinationPortInvalid = errors.New("Port specified is invalid.")
	ErrStringMatchConditionInvalid = errors.New("Only one of the Exact, Prefix or SafeRegex fields must be specified.")
)

// Xds holds the intermediate representation of a Gateway and is
// used by the xDS Translator to convert it into xDS resources.
type Xds struct {
	// Name of the Xds IR.
	Name string
	// HTTP listeners exposed by the gateway.
	HTTP []*HTTPListener
}

// Validate the fields within the Xds structure
func (x *Xds) Validate() error {
	if x.Name == "" {
		return ErrXdsNameEmpty
	}
	for _, http := range x.HTTP {
		if err := http.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// HTTPListener holds the listener configuration.
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

func (x *Xds) GetListener(name string) *HTTPListener {
	for _, listener := range x.HTTP {
		if listener.Name == name {
			return listener
		}
	}
}

// Validate the fields within the HTTPListener structure
func (h *HTTPListener) Validate() error {
	if h.Name == "" {
		return ErrHTTPListenerNameEmpty
	}
	if ip := net.ParseIP(h.Address); ip == nil {
		return ErrHTTPListenerAddressInvalid
	}
	if h.Port == 0 {
		return ErrHTTPListenerPortInvalid
	}
	if len(h.Hostnames) == 0 {
		return ErrHTTPListenerHostnamesEmpty
	}
	if h.TLS != nil {
		if err := h.TLS.Validate(); err != nil {
			return err
		}
	}
	for _, route := range h.Routes {
		if err := route.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// TLSListenerConfig holds the configuration for downstream TLS context.
type TLSListenerConfig struct {
	// ServerCertificate of the server.
	ServerCertificate []byte
	// PrivateKey for the server.
	PrivateKey []byte
}

// Validate the fields within the TLSListenerConfig structure
func (t *TLSListenerConfig) Validate() error {
	if len(t.ServerCertificate) == 0 {
		return ErrTLSServerCertEmpty
	}
	if len(t.PrivateKey) == 0 {
		return ErrTLSPrivateKey
	}
	return nil
}

// HTTPRoute holds the route information associated with the HTTP Route
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
func (h *HTTPRoute) Validate() error {
	if h.Name == "" {
		return ErrHTTPRouteNameEmpty
	}
	if h.PathMatch == nil && (len(h.HeaderMatches) == 0) && (len(h.QueryParamMatches) == 0) {
		return ErrHTTPRouteMatchEmpty
	}
	if h.PathMatch != nil {
		if err := h.PathMatch.Validate(); err != nil {
			return err
		}
	}
	for _, hMatch := range h.HeaderMatches {
		if err := hMatch.Validate(); err != nil {
			return err
		}
	}
	for _, qMatch := range h.QueryParamMatches {
		if err := qMatch.Validate(); err != nil {
			return err
		}
	}
	for _, dest := range h.Destinations {
		if err := dest.Validate(); err != nil {
			return err
		}
	}
	return nil
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
func (r *RouteDestination) Validate() error {
	// Only support IP hosts for now
	if ip := net.ParseIP(r.Host); ip == nil {
		return ErrRouteDestinationHostInvalid
	}
	if r.Port == 0 {
		return ErrRouteDestinationPortInvalid
	}

	return nil
}

// StringMatch holds the various match conditions.
// Only one of Exact, Prefix or SafeRegex can be set.
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
func (s *StringMatch) Validate() error {
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
		return ErrStringMatchConditionInvalid
	}

	return nil
}
