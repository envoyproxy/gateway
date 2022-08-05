package config

import (
	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/log"
)

// Server wraps the EnvoyGateway configuration and additional parameters
// used by Envoy Gateway server.
type Server struct {
	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *v1alpha1.EnvoyGateway
	// Logger is the logr implementation used by Envoy Gateway.
	Logger logr.Logger
}

// NewDefaultServer returns a Server with default parameters.
func NewDefaultServer() (*Server, error) {
	logger, err := log.NewLogger()
	if err != nil {
		return nil, err
	}
	return &Server{
		EnvoyGateway: v1alpha1.DefaultEnvoyGateway(),
		Logger:       logger,
	}, nil
}
