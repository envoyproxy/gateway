package config

import (
	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// Server wraps the EnvoyGateway configuration and additional parameters
// used by Envoy Gateway server.
type Server struct {
	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *v1alpha1.EnvoyGateway
	// Logger is the logr implementation used by Envoy Gateway.
	Logger logr.Logger
}
