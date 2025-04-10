// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"fmt"
	"net/url"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ValidateEnvoyGateway validates the provided EnvoyGateway.
func ValidateEnvoyGateway(eg *egv1a1.EnvoyGateway) error {
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
	case egv1a1.ProviderTypeKubernetes:
		if err := validateEnvoyGatewayKubernetesProvider(eg.Provider.Kubernetes); err != nil {
			return err
		}
	case egv1a1.ProviderTypeCustom:
		if err := validateEnvoyGatewayCustomProvider(eg.Provider.Custom); err != nil {
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

func validateEnvoyGatewayKubernetesProvider(provider *egv1a1.EnvoyGatewayKubernetesProvider) error {
	if provider == nil || provider.Watch == nil {
		return nil
	}

	watch := provider.Watch
	switch watch.Type {
	case egv1a1.KubernetesWatchModeTypeNamespaces:
		if len(watch.Namespaces) == 0 {
			return fmt.Errorf("namespaces should be specified when envoy gateway watch mode is 'Namespaces'")
		}
	case egv1a1.KubernetesWatchModeTypeNamespaceSelector:
		if watch.NamespaceSelector == nil {
			return fmt.Errorf("namespaceSelector should be specified when envoy gateway watch mode is 'NamespaceSelector'")
		}
	default:
		return fmt.Errorf("envoy gateway watch mode invalid, should be 'Namespaces' or 'NamespaceSelector'")
	}
	return nil
}

func validateEnvoyGatewayCustomProvider(provider *egv1a1.EnvoyGatewayCustomProvider) error {
	if provider == nil {
		return fmt.Errorf("empty custom provider settings")
	}

	if err := validateEnvoyGatewayCustomResourceProvider(provider.Resource); err != nil {
		return err
	}

	if err := validateEnvoyGatewayCustomInfrastructureProvider(provider.Infrastructure); err != nil {
		return err
	}

	return nil
}

func validateEnvoyGatewayCustomResourceProvider(resource egv1a1.EnvoyGatewayResourceProvider) error {
	switch resource.Type {
	case egv1a1.ResourceProviderTypeFile:
		if resource.File == nil {
			return fmt.Errorf("field 'file' should be specified when resource type is 'File'")
		}

		if len(resource.File.Paths) == 0 {
			return fmt.Errorf("no paths were assigned for file resource provider to watch")
		}
	default:
		return fmt.Errorf("unsupported resource provider: %s", resource.Type)
	}
	return nil
}

func validateEnvoyGatewayCustomInfrastructureProvider(infra *egv1a1.EnvoyGatewayInfrastructureProvider) error {
	if infra == nil {
		return nil
	}

	switch infra.Type {
	case egv1a1.InfrastructureProviderTypeHost:
		if infra.Host == nil {
			return fmt.Errorf("field 'host' should be specified when infrastructure type is 'Host'")
		}
	default:
		return fmt.Errorf("unsupported infrastructure provdier: %s", infra.Type)
	}
	return nil
}

func validateEnvoyGatewayLogging(logging *egv1a1.EnvoyGatewayLogging) error {
	if logging == nil || len(logging.Level) == 0 {
		return nil
	}

	for component, logLevel := range logging.Level {
		switch component {
		case egv1a1.LogComponentGatewayDefault,
			egv1a1.LogComponentProviderRunner,
			egv1a1.LogComponentGatewayAPIRunner,
			egv1a1.LogComponentXdsTranslatorRunner,
			egv1a1.LogComponentXdsServerRunner,
			egv1a1.LogComponentInfrastructureRunner,
			egv1a1.LogComponentGlobalRateLimitRunner:
			switch logLevel {
			case egv1a1.LogLevelTrace, egv1a1.LogLevelDebug, egv1a1.LogLevelError, egv1a1.LogLevelWarn, egv1a1.LogLevelInfo:
			default:
				return fmt.Errorf("envoy gateway logging level invalid. valid options: trace/debug/info/warn/error")
			}
		default:
			return fmt.Errorf("envoy gateway logging components invalid. valid options: system/provider/gateway-api/xds-translator/xds-server/infrastructure")
		}
	}
	return nil
}

func validateEnvoyGatewayRateLimit(rateLimit *egv1a1.RateLimit) error {
	if rateLimit == nil {
		return nil
	}
	if rateLimit.Backend.Type != egv1a1.RedisBackendType {
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

func validateEnvoyGatewayExtensionManager(extensionManager *egv1a1.ExtensionManager) error {
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

	switch {
	case extensionManager.Service.Host == "" && extensionManager.Service.FQDN == nil && extensionManager.Service.Unix == nil && extensionManager.Service.IP == nil:
		return fmt.Errorf("extension service must contain a configured target")

	case extensionManager.Service.FQDN != nil && (extensionManager.Service.IP != nil || extensionManager.Service.Unix != nil || extensionManager.Service.Host != ""),
		extensionManager.Service.IP != nil && (extensionManager.Service.FQDN != nil || extensionManager.Service.Unix != nil || extensionManager.Service.Host != ""),
		extensionManager.Service.Unix != nil && (extensionManager.Service.IP != nil || extensionManager.Service.FQDN != nil || extensionManager.Service.Host != ""):
		return fmt.Errorf("only one backend target can be configured for the extension manager")
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

func validateEnvoyGatewayTelemetry(telemetry *egv1a1.EnvoyGatewayTelemetry) error {
	if telemetry == nil {
		return nil
	}

	if telemetry.Metrics != nil {
		for _, sink := range telemetry.Metrics.Sinks {
			if sink.Type == egv1a1.MetricSinkTypeOpenTelemetry {
				if sink.OpenTelemetry == nil {
					return fmt.Errorf("OpenTelemetry is required when sink Type is OpenTelemetry")
				}
			}
		}
	}
	return nil
}
