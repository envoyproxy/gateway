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
	"k8s.io/apimachinery/pkg/api/resource"
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

	if err := validateEnvoyGatewayDebounce(eg.Debounce); err != nil {
		return err
	}

	if eg.ExtensionAPIs != nil && eg.ExtensionAPIs.DisableLua != nil && *eg.ExtensionAPIs.DisableLua == eg.ExtensionAPIs.EnableLua {
		return fmt.Errorf("disableLua and enableLua must not have the same value")
	}

	if eg.EnvoyProxy != nil {
		if err := validateEnvoyProxySpec(eg.EnvoyProxy); err != nil {
			return fmt.Errorf("invalid EnvoyProxy template: %w", err)
		}
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

	if err := validateRateLimitClusterSettings(rateLimit.ServiceSettings); err != nil {
		return fmt.Errorf("invalid rateLimit.serviceSettings: %w", err)
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

// validateRateLimitClusterSettings validates EnvoyGateway.RateLimit.ServiceSettings.
//
// EnvoyGateway is loaded as static configuration rather than admitted as a CRD, so the
// kubebuilder constraints declared on these types (enums, minimums, the Go-duration format)
// are never enforced. Without this check, a malformed value here is accepted at startup and
// only surfaces once a Global rate limit policy is actually used -- either failing translation
// or producing a cluster that Envoy rejects.
func validateRateLimitClusterSettings(cs *egv1a1.RateLimitClusterSettings) error {
	if cs == nil {
		return nil
	}

	if lb := cs.LoadBalancer; lb != nil {
		switch lb.Type {
		case egv1a1.LeastRequestLoadBalancerType, egv1a1.RandomLoadBalancerType, egv1a1.RoundRobinLoadBalancerType:
		default:
			return fmt.Errorf("loadBalancer.type %q is not supported, use LeastRequest, Random or RoundRobin", lb.Type)
		}
		if lb.SlowStart != nil {
			if lb.SlowStart.Window == nil {
				return fmt.Errorf("loadBalancer.slowStart.window is required")
			}
			if err := validateOptionalDuration("loadBalancer.slowStart.window", lb.SlowStart.Window, false); err != nil {
				return err
			}
		}
	}

	if err := validateRateLimitClusterCircuitBreaker(cs.CircuitBreaker); err != nil {
		return err
	}

	if t := cs.Timeout; t != nil {
		if t.TCP != nil {
			if err := validateOptionalDuration("timeout.tcp.connectTimeout", t.TCP.ConnectTimeout, false); err != nil {
				return err
			}
		}
		if t.HTTP != nil {
			if err := validateOptionalDuration("timeout.http.connectionIdleTimeout", t.HTTP.ConnectionIdleTimeout, true); err != nil {
				return err
			}
			if err := validateOptionalDuration("timeout.http.maxConnectionDuration", t.HTTP.MaxConnectionDuration, true); err != nil {
				return err
			}
			if err := validateOptionalDuration("timeout.http.maxStreamDuration", t.HTTP.MaxStreamDuration, true); err != nil {
				return err
			}
		}
	}

	if cs.TCPKeepalive != nil {
		if err := validateOptionalDuration("tcpKeepalive.idleTime", cs.TCPKeepalive.IdleTime, true); err != nil {
			return err
		}
		if err := validateOptionalDuration("tcpKeepalive.interval", cs.TCPKeepalive.Interval, true); err != nil {
			return err
		}
	}

	if cs.DNS != nil {
		if err := validateOptionalDuration("dns.dnsRefreshRate", cs.DNS.DNSRefreshRate, false); err != nil {
			return err
		}
		if lf := cs.DNS.LookupFamily; lf != nil {
			switch *lf {
			case egv1a1.IPv4DNSLookupFamily, egv1a1.IPv6DNSLookupFamily, egv1a1.IPv4PreferredDNSLookupFamily,
				egv1a1.IPv6PreferredDNSLookupFamily, egv1a1.IPv4AndIPv6DNSLookupFamily:
			default:
				return fmt.Errorf("dns.lookupFamily %q is not supported", *lf)
			}
		}
	}

	if err := validateRateLimitClusterConnection(cs.Connection); err != nil {
		return err
	}

	if err := validateRateLimitClusterHTTP2(cs.HTTP2); err != nil {
		return err
	}

	return validateRateLimitClusterHealthCheck(cs.HealthCheck)
}

func validateRateLimitClusterConnection(c *egv1a1.BackendConnection) error {
	if c == nil {
		return nil
	}
	if c.BufferLimit != nil {
		if v, ok := c.BufferLimit.AsInt64(); !ok || v < 1 || v > math.MaxUint32 {
			return fmt.Errorf("connection.bufferLimit value %s is out of range [1, %d]", c.BufferLimit.String(), uint32(math.MaxUint32))
		}
	}
	if pc := c.Preconnect; pc != nil {
		if pc.PerEndpointPercent != nil && (*pc.PerEndpointPercent < 100 || *pc.PerEndpointPercent > 300) {
			return fmt.Errorf("connection.preconnect.perEndpointPercent must be between 100 and 300")
		}
		if pc.PredictivePercent != nil && *pc.PredictivePercent < 100 {
			return fmt.Errorf("connection.preconnect.predictivePercent must be at least 100")
		}
	}
	return nil
}

// HTTP/2 flow-control window bounds enforced by Envoy, mirroring internal/gatewayapi.
const (
	minHTTP2WindowSize = 65535
	maxHTTP2WindowSize = math.MaxInt32
)

func validateRateLimitClusterHTTP2(h *egv1a1.HTTP2Settings) error {
	if h == nil {
		return nil
	}
	for _, w := range []struct {
		name string
		val  *resource.Quantity
	}{
		{"http2.initialStreamWindowSize", h.InitialStreamWindowSize},
		{"http2.initialConnectionWindowSize", h.InitialConnectionWindowSize},
	} {
		if w.val == nil {
			continue
		}
		if v, ok := w.val.AsInt64(); !ok || v < minHTTP2WindowSize || v > maxHTTP2WindowSize {
			return fmt.Errorf("%s value %s is out of range [%d, %d]", w.name, w.val.String(), minHTTP2WindowSize, maxHTTP2WindowSize)
		}
	}
	if h.MaxConcurrentStreams != nil && (*h.MaxConcurrentStreams < 1 || *h.MaxConcurrentStreams > math.MaxInt32) {
		return fmt.Errorf("http2.maxConcurrentStreams must be between 1 and %d", math.MaxInt32)
	}
	if a := h.OnInvalidMessage; a != nil {
		switch *a {
		case egv1a1.InvalidMessageActionTerminateConnection, egv1a1.InvalidMessageActionTerminateStream:
		default:
			return fmt.Errorf("http2.onInvalidMessage %q is not supported", *a)
		}
	}
	if ka := h.ConnectionKeepalive; ka != nil {
		// Envoy requires a keepalive timeout, and every keepalive duration must be at least 1ms.
		if ka.Timeout == nil {
			return fmt.Errorf("http2.connectionKeepalive.timeout is required")
		}
		for _, f := range []struct {
			name string
			val  *gwapiv1.Duration
		}{
			{"http2.connectionKeepalive.interval", ka.Interval},
			{"http2.connectionKeepalive.timeout", ka.Timeout},
			{"http2.connectionKeepalive.idleInterval", ka.IdleInterval},
		} {
			if err := validateOptionalDuration(f.name, f.val, false); err != nil {
				return err
			}
			if f.val != nil && parseDuration(*f.val) < time.Millisecond {
				return fmt.Errorf("%s: duration must be at least 1ms", f.name)
			}
		}
		if ka.Interval != nil && parseDuration(*ka.Timeout) >= parseDuration(*ka.Interval) {
			return fmt.Errorf("http2.connectionKeepalive.timeout must be less than interval")
		}
		if ka.IntervalJitter != nil && *ka.IntervalJitter > 100 {
			return fmt.Errorf("http2.connectionKeepalive.intervalJitter must be between 0 and 100")
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
	if rb := cb.RetryBudget; rb != nil {
		denominator := ptr.Deref(rb.Percent.Denominator, 100)
		if rb.Percent.Numerator < 0 || denominator < 1 || rb.Percent.Numerator > denominator {
			return fmt.Errorf("circuitBreaker.retryBudget.percent must satisfy 0 <= numerator <= denominator and denominator >= 1")
		}
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

func validateRateLimitClusterHealthCheck(hc *egv1a1.RateLimitHealthCheck) error {
	if hc == nil {
		return nil
	}

	if hc.PanicThreshold != nil && *hc.PanicThreshold > 100 {
		return fmt.Errorf("healthCheck.panicThreshold must be between 0 and 100")
	}

	if a := hc.Active; a != nil {
		switch a.Type {
		case egv1a1.ActiveHealthCheckerTypeTCP:
			if a.TCP == nil {
				return fmt.Errorf("healthCheck.active.tcp is required when type is TCP")
			}
			if err := validateHealthCheckPayload("healthCheck.active.tcp.send", a.TCP.Send); err != nil {
				return err
			}
			if err := validateHealthCheckPayload("healthCheck.active.tcp.receive", a.TCP.Receive); err != nil {
				return err
			}
			if a.GRPC != nil {
				return fmt.Errorf("healthCheck.active.grpc can only be set when type is GRPC")
			}
		case egv1a1.ActiveHealthCheckerTypeGRPC:
			if a.TCP != nil {
				return fmt.Errorf("healthCheck.active.tcp can only be set when type is TCP")
			}
		default:
			return fmt.Errorf("healthCheck.active.type %q is not supported, use GRPC or TCP", a.Type)
		}
		// Envoy rejects health checks with zero thresholds.
		if a.UnhealthyThreshold != nil && *a.UnhealthyThreshold == 0 {
			return fmt.Errorf("healthCheck.active.unhealthyThreshold must be at least 1")
		}
		if a.HealthyThreshold != nil && *a.HealthyThreshold == 0 {
			return fmt.Errorf("healthCheck.active.healthyThreshold must be at least 1")
		}
		if err := validateOptionalDuration("healthCheck.active.timeout", a.Timeout, false); err != nil {
			return err
		}
		if err := validateOptionalDuration("healthCheck.active.interval", a.Interval, false); err != nil {
			return err
		}
		if err := validateOptionalDuration("healthCheck.active.initialJitter", a.InitialJitter, true); err != nil {
			return err
		}
	}

	if p := hc.Passive; p != nil {
		if err := validateOptionalDuration("healthCheck.passive.interval", p.Interval, false); err != nil {
			return err
		}
		if err := validateOptionalDuration("healthCheck.passive.baseEjectionTime", p.BaseEjectionTime, false); err != nil {
			return err
		}
		if p.MaxEjectionPercent != nil && (*p.MaxEjectionPercent < 0 || *p.MaxEjectionPercent > 100) {
			return fmt.Errorf("healthCheck.passive.maxEjectionPercent must be between 0 and 100")
		}
		if p.FailurePercentageThreshold != nil && *p.FailurePercentageThreshold > 100 {
			return fmt.Errorf("healthCheck.passive.failurePercentageThreshold must be between 0 and 100")
		}
	}

	return nil
}

func validateHealthCheckPayload(field string, p *egv1a1.ActiveHealthCheckPayload) error {
	if p == nil {
		return nil
	}
	switch p.Type {
	case egv1a1.ActiveHealthCheckPayloadTypeText:
		if p.Text == nil || p.Binary != nil {
			return fmt.Errorf("%s.text must be set, and only text, when type is Text", field)
		}
	case egv1a1.ActiveHealthCheckPayloadTypeBinary:
		if p.Binary == nil || p.Text != nil {
			return fmt.Errorf("%s.binary must be set, and only binary, when type is Binary", field)
		}
	default:
		return fmt.Errorf("%s.type %q is not supported, use Text or Binary", field, p.Type)
	}
	return nil
}

// parseDuration parses a duration already checked by validateOptionalDuration.
func parseDuration(d gwapiv1.Duration) time.Duration {
	duration, _ := time.ParseDuration(string(d))
	return duration
}

// validateOptionalDuration validates parsing and field-specific xDS duration bounds.
func validateOptionalDuration(field string, d *gwapiv1.Duration, allowZero bool) error {
	if d == nil {
		return nil
	}
	duration, err := time.ParseDuration(string(*d))
	if err != nil {
		return fmt.Errorf("%s: invalid duration %q: %w", field, string(*d), err)
	}
	if duration < 0 || (!allowZero && duration == 0) {
		if allowZero {
			return fmt.Errorf("%s: duration must be nonnegative", field)
		}
		return fmt.Errorf("%s: duration must be positive", field)
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

func validateEnvoyGatewayDebounce(debounce *egv1a1.Debounce) error {
	if debounce == nil {
		return nil
	}

	after, maxHold, err := debounce.ResolveDurations()
	if err != nil {
		return err
	}

	if maxHold < after {
		return fmt.Errorf("debounce.max (%s) must be greater than or equal to debounce.after (%s)", maxHold, after)
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
