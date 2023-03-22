// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/go-logr/logr"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

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
	case s.EnvoyGateway.Gateway == nil:
		return errors.New("gateway is unspecified")
	case len(s.EnvoyGateway.Gateway.ControllerName) == 0:
		return errors.New("gateway controllerName is unspecified")
	case s.EnvoyGateway.Provider == nil:
		return errors.New("provider is unspecified")
	case s.EnvoyGateway.Provider.Type != v1alpha1.ProviderTypeKubernetes:
		return fmt.Errorf("unsupported provider %v", s.EnvoyGateway.Provider.Type)
	case len(s.Namespace) == 0:
		return errors.New("namespace is empty string")
	case s.EnvoyGateway.RateLimit != nil:
		if s.EnvoyGateway.RateLimit.Backend.Type != v1alpha1.RedisBackendType {
			return fmt.Errorf("unsupported ratelimit backend %v", s.EnvoyGateway.RateLimit.Backend.Type)
		}
		if s.EnvoyGateway.RateLimit.Backend.Redis == nil || s.EnvoyGateway.RateLimit.Backend.Redis.URL == "" {
			return fmt.Errorf("empty ratelimit redis settings")
		}
		if _, err := url.Parse(s.EnvoyGateway.RateLimit.Backend.Redis.URL); err != nil {
			return fmt.Errorf("unknown ratelimit redis url format: %w", err)
		}
	case s.EnvoyGateway.Extension != nil:
		if s.EnvoyGateway.Extension.Hooks == nil || s.EnvoyGateway.Extension.Hooks.XDSTranslator == nil {
			return fmt.Errorf("registered extension has no hooks specified")
		}

		if len(s.EnvoyGateway.Extension.Hooks.XDSTranslator.Pre) == 0 && len(s.EnvoyGateway.Extension.Hooks.XDSTranslator.Post) == 0 {
			return fmt.Errorf("registered extension has no hooks specified")
		}

		if s.EnvoyGateway.Extension.Service == nil {
			return fmt.Errorf("extension service config is empty")
		}

		if s.EnvoyGateway.Extension.Service.TLS != nil {
			certifcateRefKind := s.EnvoyGateway.Extension.Service.TLS.CertificateRef.Kind

			if certifcateRefKind == nil {
				return fmt.Errorf("certificateRef empty in extension service server TLS settings")
			}

			if *certifcateRefKind != gwapiv1b1.Kind("Secret") {
				return fmt.Errorf("unsupported extension server TLS certificateRef %v", certifcateRefKind)
			}
		}
	}
	return nil
}
