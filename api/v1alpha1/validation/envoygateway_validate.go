// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

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
		if err := validateEnvoyGatewayKubernetesRateLimit(eg.RateLimit); err != nil {
			return err
		}
	case egv1a1.ProviderTypeCustom:
		if err := validateEnvoyGatewayCustomProvider(eg.Provider.Custom); err != nil {
			return err
		}
		if err := validateEnvoyGatewayCustomRateLimit(eg.RateLimit); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported provider type")
	}

	if err := validateEnvoyGatewayLogging(eg.Logging); err != nil {
		return err
	}

	if err := validateEnvoyGatewayExtensionManagers(eg); err != nil {
		return err
	}

	if err := validateEnvoyGatewayTelemetry(eg.Telemetry); err != nil {
		return err
	}

	if err := validateEnvoyGatewayXDSServer(eg.XDSServer); err != nil {
		return err
	}

	if eg.ExtensionAPIs != nil && eg.ExtensionAPIs.DisableLua != nil && *eg.ExtensionAPIs.DisableLua == eg.ExtensionAPIs.EnableLua {
		return fmt.Errorf("disableLua and enableLua must not have the same value")
	}

	return nil
}

// WarnEnvoyGateway returns deprecation warnings for the provided EnvoyGateway configuration.
func WarnEnvoyGateway(eg *egv1a1.EnvoyGateway) []string {
	if eg == nil || eg.ExtensionAPIs == nil {
		return nil
	}
	var warnings []string
	if eg.ExtensionAPIs.DisableLua != nil {
		warnings = append(warnings, "disableLua is deprecated, use enableLua instead")
	}
	return warnings
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

func validateEnvoyGatewayKubernetesProviderCustom(provider *egv1a1.EnvoyGatewayKubernetesCustomProvider) error {
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
	case egv1a1.ResourceProviderTypeKubernetes:
		return validateEnvoyGatewayKubernetesProviderCustom(resource.Kubernetes)
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
	case egv1a1.InfrastructureProviderTypeRemote:
		if infra.Remote == nil {
			return fmt.Errorf("field 'remote' should be specified when infrastructure type is 'Remote'")
		}

		if infra.Remote.Service == nil {
			return fmt.Errorf("field 'service' should be specified when infrastructure type is 'Remote'")
		}
		err := validateExtensionService(infra.Remote.Service)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported infrastructure provider: %s", infra.Type)
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
			egv1a1.LogComponentXdsRunner,
			egv1a1.LogComponentInfrastructureRunner,
			egv1a1.LogComponentGlobalRateLimitRunner:
			switch logLevel {
			case egv1a1.LogLevelDebug, egv1a1.LogLevelError, egv1a1.LogLevelWarn, egv1a1.LogLevelInfo:
			default:
				return fmt.Errorf("envoy gateway logging level invalid. valid options: info/debug/warn/error")
			}
		default:
			return fmt.Errorf("envoy gateway logging components invalid. valid options: system/provider/gateway-api/xds-translator/xds-server/xds/infrastructure")
		}
	}
	return nil
}

func validateEnvoyGatewayKubernetesRateLimit(rateLimit *egv1a1.RateLimit) error {
	if rateLimit == nil {
		return nil
	}
	if rateLimit.Backend.Type != egv1a1.RedisBackendType {
		return fmt.Errorf("unsupported ratelimit backend %v", rateLimit.Backend.Type)
	}
	if rateLimit.Backend.Redis == nil || rateLimit.Backend.Redis.URL == "" {
		return fmt.Errorf("empty ratelimit redis settings")
	}
	redisHosts := strings.Split(rateLimit.Backend.Redis.URL, ",")
	for _, host := range redisHosts {
		if _, err := url.Parse(host); err != nil {
			return fmt.Errorf("unknown ratelimit redis url format: %w", err)
		}
	}
	return nil
}

func validateEnvoyGatewayCustomRateLimit(rateLimit *egv1a1.RateLimit) error {
	if rateLimit == nil {
		return nil
	}
	if rateLimit.URL == nil {
		return fmt.Errorf("empty ratelimit url settings")
	}
	u, err := url.Parse(*rateLimit.URL)
	if err != nil {
		return fmt.Errorf("unknown ratelimit url format: %w", err)
	}
	// The xDS translator expects a grpc:// URL with an explicit port so it can
	// build the rate limit service cluster. Reject anything else here so that
	// configuration validation fails instead of the translator panicking.
	if u.Scheme != "grpc" {
		return fmt.Errorf("ratelimit url must use the grpc:// scheme, got %q", *rateLimit.URL)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("ratelimit url must include a host, got %q", *rateLimit.URL)
	}
	if u.Port() == "" {
		return fmt.Errorf("ratelimit url must include an explicit port, got %q", *rateLimit.URL)
	}
	if _, err := strconv.ParseUint(u.Port(), 10, 32); err != nil {
		return fmt.Errorf("ratelimit url has invalid port %q: %w", u.Port(), err)
	}
	return nil
}

func validateEnvoyGatewayExtensionManagers(eg *egv1a1.EnvoyGateway) error {
	if eg.ExtensionManager != nil && len(eg.ExtensionManagers) > 0 {
		return fmt.Errorf("extensionManager and extensionManagers are mutually exclusive")
	}

	// Mirror +kubebuilder:validation:MinItems=1 for EnvoyGatewaySpec.ExtensionManagers:
	// reject an explicitly-set-but-empty list. A nil slice means the field was omitted.
	if eg.ExtensionManagers != nil && len(eg.ExtensionManagers) == 0 {
		return fmt.Errorf("extensionManagers must contain at least one entry when specified")
	}

	if len(eg.ExtensionManagers) > 0 {
		names := make(map[string]struct{})
		for i, em := range eg.ExtensionManagers {
			if em.Name == "" {
				return fmt.Errorf("extension manager at index %d: name is required", i)
			}
			if _, exists := names[em.Name]; exists {
				return fmt.Errorf("extension manager at index %d: duplicate name %q", i, em.Name)
			}
			names[em.Name] = struct{}{}
			if err := validateEnvoyGatewayExtensionManager(&eg.ExtensionManagers[i]); err != nil {
				return fmt.Errorf("extension manager %q: %w", em.Name, err)
			}
		}
		return nil
	}

	return validateEnvoyGatewayExtensionManager(eg.ExtensionManager)
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

	err := validateExtensionService(extensionManager.Service)
	if err != nil {
		return err
	}

	return nil
}

func validateEnvoyGatewayXDSServer(xdsServer *egv1a1.XDSServer) error {
	if xdsServer == nil {
		return nil
	}

	if xdsServer.MaxConnectionAge != nil {
		d, err := time.ParseDuration(string(*xdsServer.MaxConnectionAge))
		if err != nil {
			return fmt.Errorf("invalid xdsServer.maxConnectionAge: %w", err)
		}
		if d <= 0 {
			return fmt.Errorf("xdsServer.maxConnectionAge must be greater than zero")
		}
	}

	if xdsServer.MaxConnectionAgeGrace != nil {
		d, err := time.ParseDuration(string(*xdsServer.MaxConnectionAgeGrace))
		if err != nil {
			return fmt.Errorf("invalid xdsServer.maxConnectionAgeGrace: %w", err)
		}
		if d <= 0 {
			return fmt.Errorf("xdsServer.maxConnectionAgeGrace must be greater than zero")
		}
	}

	return nil
}

func validateEnvoyGatewayOpenTelemetrySink(sink *egv1a1.EnvoyGatewayOpenTelemetrySink) error {
	if sink.Protocol != egv1a1.GRPCProtocol && sink.Protocol != egv1a1.HTTPProtocol {
		return fmt.Errorf("unsupported protocol %s for OpenTelemetry sink, only 'grpc' and 'http' are supported", sink.Protocol)
	}
	if sink.ExportInterval != nil {
		d, err := time.ParseDuration(string(*sink.ExportInterval))
		if err != nil {
			return fmt.Errorf("invalid exportInterval: %w", err)
		}
		if d <= 0 {
			return fmt.Errorf("exportInterval must be greater than zero")
		}
	}
	if sink.ExportTimeout != nil {
		d, err := time.ParseDuration(string(*sink.ExportTimeout))
		if err != nil {
			return fmt.Errorf("invalid exportTimeout: %w", err)
		}
		if d <= 0 {
			return fmt.Errorf("exportTimeout must be greater than zero")
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
				if err := validateEnvoyGatewayOpenTelemetrySink(sink.OpenTelemetry); err != nil {
					return err
				}
			}
		}
	}

	if telemetry.Traces != nil {
		if telemetry.Traces.Sink.OpenTelemetry == nil {
			return fmt.Errorf("OpenTelemetry is required when trace sink Type is OpenTelemetry")
		}
		if err := validateEnvoyGatewayOpenTelemetrySink(telemetry.Traces.Sink.OpenTelemetry); err != nil {
			return err
		}
	}
	return nil
}

func validateExtensionService(extensionService *egv1a1.ExtensionService) error {
	if extensionService == nil {
		return fmt.Errorf("extension service config is empty")
	}

	switch {
	case extensionService.Host == "" && extensionService.FQDN == nil && extensionService.Unix == nil && extensionService.IP == nil:
		return fmt.Errorf("extension service must contain a configured target")

	case extensionService.FQDN != nil && (extensionService.IP != nil || extensionService.Unix != nil || extensionService.Host != ""),
		extensionService.IP != nil && (extensionService.FQDN != nil || extensionService.Unix != nil || extensionService.Host != ""),
		extensionService.Unix != nil && (extensionService.IP != nil || extensionService.FQDN != nil || extensionService.Host != ""):
		return fmt.Errorf("only one backend target can be configured for the extension manager")
	}

	if extensionService.TLS != nil {
		certRef := &extensionService.TLS.CertificateRef
		if (certRef.Group != nil && *certRef.Group != corev1.GroupName) ||
			(certRef.Kind != nil && *certRef.Kind != "Secret") {
			return fmt.Errorf("unsupported extension server TLS certificateRef group/kind")
		}

		if extensionService.TLS.ClientCertificateRef != nil {
			clientCertRef := extensionService.TLS.ClientCertificateRef
			if (clientCertRef.Group != nil && *clientCertRef.Group != corev1.GroupName) ||
				(clientCertRef.Kind != nil && *clientCertRef.Kind != "Secret") {
				return fmt.Errorf("unsupported extension server mTLS clientCertificateRef group/kind")
			}
		}
	}
	return nil
}
