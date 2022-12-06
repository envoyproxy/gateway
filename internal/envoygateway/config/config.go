// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"errors"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/log"
	"github.com/envoyproxy/gateway/internal/utils/env"
)

const (
	// DefaultNamespace is the default namespace of Envoy Gateway.
	DefaultNamespace = "envoy-gateway-system"
	// EnvoyGatewayServiceName is the name of the Envoy Gateway service.
	EnvoyGatewayServiceName = "envoy-gateway"
	// EnvoyPrefix is the prefix applied to the Envoy ConfigMap, Service, Deployment, and ServiceAccount.
	EnvoyPrefix = "envoy"
)

// Server wraps the EnvoyGateway configuration and additional parameters
// used by Envoy Gateway server.
type Server struct {
	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *v1alpha1.EnvoyGateway
	// Namespace is the namespace that Envoy Gateway runs in.
	Namespace string
	// Logger is the logr implementation used by Envoy Gateway.
	Logger logr.Logger
}

// New returns a Server with default parameters.
func New() (*Server, error) {
	logger, err := log.NewLogger()
	if err != nil {
		return nil, err
	}
	return &Server{
		EnvoyGateway: v1alpha1.DefaultEnvoyGateway(),
		Namespace:    env.Lookup("ENVOY_GATEWAY_NAMESPACE", DefaultNamespace),
		Logger:       logger,
	}, nil
}

// Validate validates a Server config.
func (s *Server) Validate() error {
	switch {
	case s == nil:
		return errors.New("server config is unspecified")
	case s.EnvoyGateway == nil:
		return errors.New("envoy gateway config is unspecified")
	case s.EnvoyGateway.EnvoyGatewaySpec.Gateway == nil:
		return errors.New("gateway is unspecified")
	case len(s.EnvoyGateway.EnvoyGatewaySpec.Gateway.ControllerName) == 0:
		return errors.New("gateway controllerName is unspecified")
	case s.EnvoyGateway.EnvoyGatewaySpec.Provider == nil:
		return errors.New("provider is unspecified")
	case s.EnvoyGateway.EnvoyGatewaySpec.Provider.Type != v1alpha1.ProviderTypeKubernetes:
		return fmt.Errorf("unsupported provider %v", s.EnvoyGateway.EnvoyGatewaySpec.Provider.Type)
	case len(s.Namespace) == 0:
		return errors.New("namespace is empty string")
	}

	return nil
}
