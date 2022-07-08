package envoygateway

import (
	"github.com/envoyproxy/gateway/api/config/v1alpha1"

	"github.com/go-logr/logr"
)

// Config wraps the EnvoyGateway configuration and additional parameters.
type Config struct {
	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *v1alpha1.EnvoyGateway
	// Logger is the logr implementation used by Envoy Gateway.
	Logger logr.Logger
}
