package ir

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

// Xds holds the intermediate representation of a Gateway and is
// used by the xDS Translator to convert it into xDS resources
type Xds struct {
	Name string
	// One or more HTTP or HTTPS listeners exposed by the gateway
	Http []HttpListener
}

type HttpListener struct {
	// Port on which the service can be expected to be accessed by clients.
	Port uint32

	// Hostnames (Host/Authority header value) with which the service can be expected to be accessed by clients.
	Hostnames []string

	// TLS certificate info. If omitted, the gateway will expose a plain text HTTP server.
	Tls ServerTLSSettings

	// Routing rules associated with HTTP traffic to the service.
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
	// Priority of the destination.
	Priority uint32
}

// TLSMode Describes how authentication is performed as part of establishing TLS connection.
type TLSMode int32

const (
	INVALID TLSMode = 0
	// Only the server is authenticated.
	SIMPLE = 1
	// Both the peers in the communication must present their certificate for TLS authentication.
	MUTUAL = 2
)

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
