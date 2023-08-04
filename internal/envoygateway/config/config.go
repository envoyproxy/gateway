// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"errors"
	"fmt"
	"net/url"

	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/env"
)

const (
	// DefaultNamespace is the default namespace of Envoy Gateway.
	DefaultNamespace = "envoy-gateway-system"
	// DefaultDNSDomain is the default DNS domain used by k8s services.
	DefaultDNSDomain = "cluster.local"
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
	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string
	// Logger is the logr implementation used by Envoy Gateway.
	Logger logging.Logger
}

// New returns a Server with default parameters.
func New() (*Server, error) {
	return &Server{
		EnvoyGateway: v1alpha1.DefaultEnvoyGateway(),
		Namespace:    env.Lookup("ENVOY_GATEWAY_NAMESPACE", DefaultNamespace),
		DNSDomain:    env.Lookup("KUBERNETES_CLUSTER_DOMAIN", DefaultDNSDomain),
		// the default logger
		Logger: logging.DefaultLogger(v1alpha1.LogLevelInfo),
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
	case s.EnvoyGateway.Logging != nil && len(s.EnvoyGateway.Logging.Level) != 0:
		level := s.EnvoyGateway.Logging.Level
		for component, logLevel := range level {
			switch component {
			case v1alpha1.LogComponentGatewayDefault,
				v1alpha1.LogComponentProviderRunner,
				v1alpha1.LogComponentGatewayAPIRunner,
				v1alpha1.LogComponentXdsTranslatorRunner,
				v1alpha1.LogComponentXdsServerRunner,
				v1alpha1.LogComponentInfrastructureRunner,
				v1alpha1.LogComponentGlobalRateLimitRunner:
				switch logLevel {
				case v1alpha1.LogLevelDebug, v1alpha1.LogLevelError, v1alpha1.LogLevelWarn, v1alpha1.LogLevelInfo:
				default:
					return errors.New("envoy gateway logging level invalid. valid options: info/debug/warn/error")
				}
			default:
				return errors.New("envoy gateway logging components invalid. valid options: system/provider/gateway-api/xds-translator/xds-server/infrastructure")
			}
		}
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
	case s.EnvoyGateway.ExtensionManager != nil:
		if s.EnvoyGateway.ExtensionManager.Hooks == nil || s.EnvoyGateway.ExtensionManager.Hooks.XDSTranslator == nil {
			return fmt.Errorf("registered extension has no hooks specified")
		}

		if len(s.EnvoyGateway.ExtensionManager.Hooks.XDSTranslator.Pre) == 0 && len(s.EnvoyGateway.ExtensionManager.Hooks.XDSTranslator.Post) == 0 {
			return fmt.Errorf("registered extension has no hooks specified")
		}

		if s.EnvoyGateway.ExtensionManager.Service == nil {
			return fmt.Errorf("extension service config is empty")
		}

		if s.EnvoyGateway.ExtensionManager.Service.TLS != nil {
			certificateRefKind := s.EnvoyGateway.ExtensionManager.Service.TLS.CertificateRef.Kind

			if certificateRefKind == nil {
				return fmt.Errorf("certificateRef empty in extension service server TLS settings")
			}

			if *certificateRefKind != gwapiv1b1.Kind("Secret") {
				return fmt.Errorf("unsupported extension server TLS certificateRef %v", certificateRefKind)
			}
		}
	}
	return nil
}
