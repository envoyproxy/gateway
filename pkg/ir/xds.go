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
	Http []HttpListener
}

// HttpListener holds the listener configuration.
type HttpListener struct {
	// Port on which the service can be expected to be accessed by clients.
	Port uint32

	// Hostnames (Host/Authority header value) with which the service can be expected to be accessed by clients.
	Hostnames []string

	// Tls certificate info. If omitted, the gateway will expose a plain text HTTP server.
	Tls ServerTLSSettings

	// Routes associated with HTTP traffic to the service.
	Routes []HttpRoute
}

type HttpRoute struct {
	// Match condition.
	Matchers     []route.HeaderMatcher
	Destinations []RouteDestination
}

type RouteDestination struct {
	// FQDN or IP address of the backend service.
	Host string
	// The port on the service to forward the request to.
	Port uint32
	// Weight associated with this destination.
	Weight uint32
}

// TLSMode describes how authentication is performed as part of establishing a TLS connection.
type TLSMode string

const (
	// Only the server is authenticated.
	SimpleTLS TLSMode = "simple-tls"
)

// String returns the string literal for the TLS Mode
func (m TLSMode) String() string {
	return string(m)
}

type ServerTLSSettings struct {
	// Set this to SIMPLE, or MUTUAL for one-way TLS, mutual TLS respectively.
	Mode TLSMode
	// The server certificate.
	ServerCertificate []byte
	// The server private key.
	PrivateKey []byte
	// The CA certificates for authenticating clients when using TLS mode "MUTUAL".
	CaCertificates []byte
}
