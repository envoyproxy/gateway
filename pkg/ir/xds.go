package ir

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

// Xds holds the intermediate representation of a Gateway and is
// used by the xDS Translator to convert it into xDS resources.
type Xds struct {
	// Name of the Xds IR.
	Name string
	// HTTP listeners exposed by the gateway.
	HTTP []HTTPListener
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
	TLS TLSListenerConfig
	// Routes associated with HTTP traffic to the service.
	Routes []HTTPRoute
}

// HTTPRoute holds the route information associated with the HTTP Route
type HTTPRoute struct {
	// Name of the HttpRoute
	Name string
	// Matches define the match conditions for this route.
	Matches []route.HeaderMatcher
	// Destinations associated with this matched route.
	Destinations []RouteDestination
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

// TLSMode describes how authentication is performed as part of establishing a TLS connection.
type TLSMode string

const (
	// SimpleTLS denotes that only the server is authenticated.
	SimpleTLS TLSMode = "SimpleTLS"
)

// String returns the string literal for the TLS Mode
func (m TLSMode) String() string {
	return string(m)
}

// TLSListenerConfig holds the configuration for downstream TLS context.
type TLSListenerConfig struct {
	// Mode for TLS Authentication. Set this to SIMPLE for one-way TLS.
	Mode TLSMode
	// ServerCertificate of the server.
	ServerCertificate []byte
	// PrivateKey for the server.
	PrivateKey []byte
}
