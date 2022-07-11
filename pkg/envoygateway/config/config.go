package config

import (
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/pkg/log"
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
		EnvoyGateway: &v1alpha1.EnvoyGateway{
			TypeMeta: metav1.TypeMeta{
				APIVersion: v1alpha1.GroupVersion.String(),
				Kind:       v1alpha1.KindEnvoyGateway,
			},
			EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
				Provider: v1alpha1.DefaultProvider(),
				Gateway:  v1alpha1.DefaultGateway(),
			},
		},
		Logger: logger,
	}, nil
}
