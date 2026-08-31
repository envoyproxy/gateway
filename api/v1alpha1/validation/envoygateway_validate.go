// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

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

// WarnEnvoyGateway returns non-fatal warnings for the provided EnvoyGateway configuration:
// deprecated fields, and fields that are accepted but have no effect.
func WarnEnvoyGateway(eg *egv1a1.EnvoyGateway) []string {
	if eg == nil {
		return nil
	}

	var warnings []string

	if eg.ExtensionAPIs != nil && eg.ExtensionAPIs.DisableLua != nil {
		warnings = append(warnings, "disableLua is deprecated, use enableLua instead")
	}

	warnings = append(warnings, warnRateLimitClusterSettings(eg.RateLimit)...)

	return warnings
}

// warnRateLimitClusterSettings warns about RateLimit.ClusterSettings members that are accepted
// by validateRateLimitClusterSettings but have no effect on the rate limit service cluster: it
// has no associated route, so ir.TrafficFeatures.ClusterFeatures() drops Retry entirely, and
// ir.Timeout.ClusterOnly() strips HTTP.RequestTimeout/HTTP.StreamIdleTimeout, before the CDS
// cluster is built.
func warnRateLimitClusterSettings(rateLimit *egv1a1.RateLimit) []string {
	if rateLimit == nil || rateLimit.ClusterSettings == nil {
		return nil
	}
	cs := rateLimit.ClusterSettings

	var warnings []string
	if cs.Retry != nil {
		warnings = append(warnings, "rateLimit.clusterSettings.retry has no effect: the rate limit service cluster has no associated route")
	}
	if cs.Timeout != nil && cs.Timeout.HTTP != nil {
		if cs.Timeout.HTTP.RequestTimeout != nil {
			warnings = append(warnings, "rateLimit.clusterSettings.timeout.http.requestTimeout has no effect: the rate limit service cluster has no associated route")
		}
		if cs.Timeout.HTTP.StreamIdleTimeout != nil {
			warnings = append(warnings, "rateLimit.clusterSettings.timeout.http.streamIdleTimeout has no effect: the rate limit service cluster has no associated route")
		}
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

func validateEnvoyGatewayRateLimit(rateLimit *egv1a1.RateLimit) error {
	if rateLimit == nil {
		return nil
	}

	if err := validateRateLimitClusterSettings(rateLimit.ClusterSettings); err != nil {
		return fmt.Errorf("invalid rateLimit.clusterSettings: %w", err)
	}

	if rateLimit.Backend.Type != egv1a1.RedisBackendType {
		return fmt.Errorf("unsupported ratelimit backend %v", rateLimit.Backend.Type)
	}
	redis := rateLimit.Backend.Redis
	if redis == nil {
		return fmt.Errorf("empty ratelimit redis settings")
	}

	hasURL := ptr.Deref(redis.URL, "") != ""
	hasURLRef := redis.URLRef != nil
	if hasURL == hasURLRef {
		return fmt.Errorf("exactly one of ratelimit redis url or urlRef must be set")
	}

	if hasURLRef {
		ref := redis.URLRef.SecretKeyRef
		if ref == nil || ref.Name == "" || ref.Key == "" {
			return fmt.Errorf("ratelimit redis urlRef.secretKeyRef must set both name and key")
		}
		// The Secret is required: the rate limit pod must wait for an
		// externally provisioned Secret rather than start with an unset
		// REDIS_URL. An optional reference would let the container boot
		// nonfunctional, so reject optional=true.
		if ref.Optional != nil && *ref.Optional {
			return fmt.Errorf("ratelimit redis urlRef.secretKeyRef.optional must not be true; the Secret is required")
		}
		return nil
	}

	return ValidateRedisURL(*redis.URL)
}

// ValidateRedisURL validates a ratelimit Redis URL string, which may be a single
// host or a comma-delimited list of hosts for Sentinel and Cluster deployments.
func ValidateRedisURL(redisURL string) error {
	if redisURL == "" {
		return fmt.Errorf("ratelimit redis url is empty")
	}
	for _, host := range strings.Split(redisURL, ",") {
		if _, err := url.Parse(host); err != nil {
			return fmt.Errorf("unknown ratelimit redis url format: %w", err)
		}
	}
	return nil
}

// validateRateLimitClusterSettings validates EnvoyGateway.RateLimit.ClusterSettings.
//
// EnvoyGateway is loaded as static configuration rather than admitted as a CRD, so the
// kubebuilder/CEL constraints declared on ClusterSettings (e.g. Minimum=0 on circuit breaker
// fields, the Go-duration format on timeout fields) are never enforced. Without this check, a
// malformed value here is accepted at startup and only surfaces once a Global rate limit policy
// is actually used and ProcessGlobalResources fails to translate it -- by which point the runner
// has already begun publishing IR built from the rest of the (valid) configuration.
//
// The rate limit service cluster has no associated route, so route-scoped ClusterSettings
// members have nowhere to apply: ir.TrafficFeatures.ClusterFeatures() drops Retry entirely, and
// ir.Timeout.ClusterOnly() strips HTTP.RequestTimeout/HTTP.StreamIdleTimeout before the CDS
// cluster is built. Rather than rejecting the whole configuration because of them, those members
// are accepted here -- only the fields that actually apply to a cluster are validated below --
// and WarnEnvoyGateway surfaces a non-fatal warning that they have no effect.
func validateRateLimitClusterSettings(cs *egv1a1.ClusterSettings) error {
	if cs == nil {
		return nil
	}

	if err := validateRateLimitClusterCircuitBreaker(cs.CircuitBreaker); err != nil {
		return err
	}

	if err := validateRateLimitClusterTimeout(cs.Timeout); err != nil {
		return err
	}

	if cs.TCPKeepalive != nil {
		if err := validateOptionalDuration("tcpKeepalive.idleTime", cs.TCPKeepalive.IdleTime); err != nil {
			return err
		}
		if err := validateOptionalDuration("tcpKeepalive.interval", cs.TCPKeepalive.Interval); err != nil {
			return err
		}
	}

	if cs.DNS != nil {
		if err := validateOptionalDuration("dns.dnsRefreshRate", cs.DNS.DNSRefreshRate); err != nil {
			return err
		}
	}

	return nil
}

func validateRateLimitClusterCircuitBreaker(cb *egv1a1.CircuitBreaker) error {
	if cb == nil {
		return nil
	}

	fields := []struct {
		name string
		val  *int64
	}{
		{"circuitBreaker.maxConnections", cb.MaxConnections},
		{"circuitBreaker.maxPendingRequests", cb.MaxPendingRequests},
		{"circuitBreaker.maxParallelRequests", cb.MaxParallelRequests},
		{"circuitBreaker.maxParallelRetries", cb.MaxParallelRetries},
		{"circuitBreaker.maxRequestsPerConnection", cb.MaxRequestsPerConnection},
	}
	if cb.PerEndpoint != nil {
		fields = append(fields, struct {
			name string
			val  *int64
		}{"circuitBreaker.perEndpoint.maxConnections", cb.PerEndpoint.MaxConnections})
	}

	for _, f := range fields {
		if f.val == nil {
			continue
		}
		if *f.val < 0 || *f.val > math.MaxUint32 {
			return fmt.Errorf("%s value %d is out of range [0, %d]", f.name, *f.val, uint32(math.MaxUint32))
		}
	}

	return nil
}

func validateRateLimitClusterTimeout(t *egv1a1.Timeout) error {
	if t == nil {
		return nil
	}

	if t.TCP != nil {
		if err := validateOptionalDuration("timeout.tcp.connectTimeout", t.TCP.ConnectTimeout); err != nil {
			return err
		}
	}

	if t.HTTP != nil {
		// RequestTimeout and StreamIdleTimeout only ever take effect on a route, and the rate
		// limit service cluster has none -- ir.Timeout.ClusterOnly() already drops them before
		// the CDS cluster is built, so there's nothing to validate here. WarnEnvoyGateway warns
		// about them separately (see warnRateLimitClusterSettings).

		if err := validateOptionalDuration("timeout.http.connectionIdleTimeout", t.HTTP.ConnectionIdleTimeout); err != nil {
			return err
		}
		if err := validateOptionalDuration("timeout.http.maxConnectionDuration", t.HTTP.MaxConnectionDuration); err != nil {
			return err
		}
		if err := validateOptionalDuration("timeout.http.maxStreamDuration", t.HTTP.MaxStreamDuration); err != nil {
			return err
		}
	}

	return nil
}

// validateOptionalDuration parses d, when set, the same way IR translation does
// (time.ParseDuration), so a malformed value is rejected at config-load time instead of at
// first use.
func validateOptionalDuration(field string, d *gwapiv1.Duration) error {
	if d == nil {
		return nil
	}
	if _, err := time.ParseDuration(string(*d)); err != nil {
		return fmt.Errorf("%s: invalid duration %q: %w", field, string(*d), err)
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

	if xdsServer.MaxReceiveMessageSize != nil {
		v, ok := xdsServer.MaxReceiveMessageSize.AsInt64()
		if !ok || v <= 0 {
			return fmt.Errorf("xdsServer.maxReceiveMessageSize must be greater than zero")
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
