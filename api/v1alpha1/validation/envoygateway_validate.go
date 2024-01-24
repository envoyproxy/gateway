// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"errors"
	"fmt"
	"net/url"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

// ValidateEnvoyGateway validates the provided EnvoyGateway.
func ValidateEnvoyGateway(eg *v1alpha1.EnvoyGateway) error {
	switch {
	case eg == nil:
		return errors.New("envoy gateway config is unspecified")
	case eg.Gateway == nil:
		return errors.New("gateway is unspecified")
	case len(eg.Gateway.ControllerName) == 0:
		return errors.New("gateway controllerName is unspecified")
	case eg.Provider == nil:
		return errors.New("provider is unspecified")
	case eg.Provider.Type != v1alpha1.ProviderTypeKubernetes:
		return fmt.Errorf("unsupported provider %v", eg.Provider.Type)
	case eg.Provider.Kubernetes != nil && eg.Provider.Kubernetes.Watch != nil:
		watch := eg.Provider.Kubernetes.Watch
		switch watch.Type {
		case v1alpha1.KubernetesWatchModeTypeNamespaces:
			if len(watch.Namespaces) == 0 {
				return errors.New("namespaces should be specified when envoy gateway watch mode is 'Namespaces'")
			}
		case v1alpha1.KubernetesWatchModeTypeNamespaceSelector:
			if watch.NamespaceSelector == nil {
				return errors.New("namespaceSelector should be specified when envoy gateway watch mode is 'NamespaceSelector'")
			}
		default:
			return errors.New("envoy gateway watch mode invalid, should be 'Namespaces' or 'NamespaceSelector'")
		}
	case eg.Logging != nil && len(eg.Logging.Level) != 0:
		level := eg.Logging.Level
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
	case eg.RateLimit != nil:
		if eg.RateLimit.Backend.Type != v1alpha1.RedisBackendType {
			return fmt.Errorf("unsupported ratelimit backend %v", eg.RateLimit.Backend.Type)
		}
		if eg.RateLimit.Backend.Redis == nil || eg.RateLimit.Backend.Redis.URL == "" {
			return fmt.Errorf("empty ratelimit redis settings")
		}
		if _, err := url.Parse(eg.RateLimit.Backend.Redis.URL); err != nil {
			return fmt.Errorf("unknown ratelimit redis url format: %w", err)
		}
	case eg.ExtensionManager != nil:
		if eg.ExtensionManager.Hooks == nil || eg.ExtensionManager.Hooks.XDSTranslator == nil {
			return fmt.Errorf("registered extension has no hooks specified")
		}

		if len(eg.ExtensionManager.Hooks.XDSTranslator.Pre) == 0 && len(eg.ExtensionManager.Hooks.XDSTranslator.Post) == 0 {
			return fmt.Errorf("registered extension has no hooks specified")
		}

		if eg.ExtensionManager.Service == nil {
			return fmt.Errorf("extension service config is empty")
		}

		if eg.ExtensionManager.Service.TLS != nil {
			certificateRefKind := eg.ExtensionManager.Service.TLS.CertificateRef.Kind

			if certificateRefKind == nil {
				return fmt.Errorf("certificateRef empty in extension service server TLS settings")
			}

			if *certificateRefKind != gwapiv1.Kind("Secret") {
				return fmt.Errorf("unsupported extension server TLS certificateRef %v", certificateRefKind)
			}
		}
	case eg.Telemetry != nil:
		if eg.Telemetry.Metrics != nil {
			for _, sink := range eg.Telemetry.Metrics.Sinks {
				if sink.Type == v1alpha1.MetricSinkTypeOpenTelemetry {
					if sink.OpenTelemetry == nil {
						return fmt.Errorf("OpenTelemetry is required when sink Type is OpenTelemetry")
					}
				}
			}
		}
	}
	return nil
}
