// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"fmt"
	"net/url"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

// ValidateEnvoyGateway validates the provided EnvoyGateway.
func ValidateEnvoyGateway(eg *v1alpha1.EnvoyGateway) error {
	if eg == nil {
		return fmt.Errorf("envoy gateway config is unspecified")
	}

	if eg.Gateway == nil {
		return fmt.Errorf("gateway is unspecified")
	}

	if len(eg.Gateway.ControllerName) == 0 {
		return fmt.Errorf("gateway controllerName is unspecified")
	}

	if eg.Provider == nil {
		return fmt.Errorf("provider is unspecified")
	}

	switch eg.Provider.Type {
	case v1alpha1.ProviderTypeKubernetes:
		if err := validateEnvoyGatewayKubernetesProvider(eg.Provider.Kubernetes); err != nil {
			return err
		}
	case v1alpha1.ProviderTypeFile:
		if err := validateEnvoyGatewayFileProvider(eg.Provider.Custom); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported provider type")
	}

	if err := validateEnvoyGatewayLogging(eg.Logging); err != nil {
		return err
	}

	if err := validateEnvoyGatewayRateLimit(eg.RateLimit); err != nil {
		return err
	}

	if err := validateEnvoyGatewayExtensionManager(eg.ExtensionManager); err != nil {
		return err
	}

	if err := validateEnvoyGatewayTelemetry(eg.Telemetry); err != nil {
		return err
	}

	return nil
}

func validateEnvoyGatewayKubernetesProvider(provider *v1alpha1.EnvoyGatewayKubernetesProvider) error {
	if provider == nil {
		return nil
	}

	watch := provider.Watch
	if watch == nil {
		return nil
	}

	switch watch.Type {
	case v1alpha1.KubernetesWatchModeTypeNamespaces:
		if len(watch.Namespaces) == 0 {
			return fmt.Errorf("namespaces should be specified when envoy gateway watch mode is 'Namespaces'")
		}
	case v1alpha1.KubernetesWatchModeTypeNamespaceSelector:
		if watch.NamespaceSelector == nil {
			return fmt.Errorf("namespaceSelector should be specified when envoy gateway watch mode is 'NamespaceSelector'")
		}
	default:
		return fmt.Errorf("envoy gateway watch mode invalid, should be 'Namespaces' or 'NamespaceSelector'")
	}
	return nil
}

func validateEnvoyGatewayFileProvider(provider *v1alpha1.EnvoyGatewayCustomProvider) error {
	if provider == nil {
		return nil
	}

	rType, iType := provider.Resource.Type, provider.Infrastructure.Type
	if rType != v1alpha1.ResourceProviderTypeFile || iType != v1alpha1.InfrastructureProviderTypeHost {
		return fmt.Errorf("file provider only supports 'File' resource type and 'Host' infra type")
	}

	if provider.Resource.File == nil {
		return fmt.Errorf("field 'file' should be specified when resource type is 'File'")
	}

	if provider.Infrastructure.Host == nil {
		return fmt.Errorf("field 'host' should be specified when infrastructure type is 'Host'")
	}

	// TODO(sh2): add more validations for infra.host

	return nil
}

func validateEnvoyGatewayLogging(logging *v1alpha1.EnvoyGatewayLogging) error {
	if logging == nil || len(logging.Level) == 0 {
		return nil
	}

	for component, logLevel := range logging.Level {
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
				return fmt.Errorf("envoy gateway logging level invalid. valid options: info/debug/warn/error")
			}
		default:
			return fmt.Errorf("envoy gateway logging components invalid. valid options: system/provider/gateway-api/xds-translator/xds-server/infrastructure")
		}
	}
	return nil
}

func validateEnvoyGatewayRateLimit(rateLimit *v1alpha1.RateLimit) error {
	if rateLimit == nil {
		return nil
	}
	if rateLimit.Backend.Type != v1alpha1.RedisBackendType {
		return fmt.Errorf("unsupported ratelimit backend %v", rateLimit.Backend.Type)
	}
	if rateLimit.Backend.Redis == nil || rateLimit.Backend.Redis.URL == "" {
		return fmt.Errorf("empty ratelimit redis settings")
	}
	if _, err := url.Parse(rateLimit.Backend.Redis.URL); err != nil {
		return fmt.Errorf("unknown ratelimit redis url format: %w", err)
	}
	return nil
}

func validateEnvoyGatewayExtensionManager(extensionManager *v1alpha1.ExtensionManager) error {
	if extensionManager == nil {
		return nil
	}

	if extensionManager.Hooks == nil || extensionManager.Hooks.XDSTranslator == nil {
		return fmt.Errorf("registered extension has no hooks specified")
	}

	if len(extensionManager.Hooks.XDSTranslator.Pre) == 0 && len(extensionManager.Hooks.XDSTranslator.Post) == 0 {
		return fmt.Errorf("registered extension has no hooks specified")
	}

	if extensionManager.Service == nil {
		return fmt.Errorf("extension service config is empty")
	}

	if extensionManager.Service.TLS != nil {
		certificateRefKind := extensionManager.Service.TLS.CertificateRef.Kind

		if certificateRefKind == nil {
			return fmt.Errorf("certificateRef empty in extension service server TLS settings")
		}

		if *certificateRefKind != "Secret" {
			return fmt.Errorf("unsupported extension server TLS certificateRef %v", certificateRefKind)
		}
	}
	return nil
}

func validateEnvoyGatewayTelemetry(telemetry *v1alpha1.EnvoyGatewayTelemetry) error {
	if telemetry == nil {
		return nil
	}

	if telemetry.Metrics != nil {
		for _, sink := range telemetry.Metrics.Sinks {
			if sink.Type == v1alpha1.MetricSinkTypeOpenTelemetry {
				if sink.OpenTelemetry == nil {
					return fmt.Errorf("OpenTelemetry is required when sink Type is OpenTelemetry")
				}
			}
		}
	}
	return nil
}
