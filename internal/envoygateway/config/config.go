// Copyright 2022 Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.
package config

import (
	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/log"
)

const (
	// EnvoyGatewayNamespace is the namespace where envoy-gateway is running.
	EnvoyGatewayNamespace = "envoy-gateway-system"
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
