// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package apihelpers holds the defaulting and accessor logic for the
// gateway.envoyproxy.io/v1alpha1 types. It lives outside the API package so
// that package only carries the API itself.
package apihelpers

import (
	"fmt"
	"net"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// DefaultEnvoyGateway returns a new EnvoyGateway with default configuration parameters.
func DefaultEnvoyGateway() *egv1a1.EnvoyGateway {
	return &egv1a1.EnvoyGateway{
		TypeMeta: metav1.TypeMeta{
			Kind:       egv1a1.KindEnvoyGateway,
			APIVersion: egv1a1.GroupVersion.String(),
		},
		EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
			Gateway:   DefaultGateway(),
			Provider:  DefaultEnvoyGatewayProvider(),
			Logging:   DefaultEnvoyGatewayLogging(),
			Admin:     DefaultEnvoyGatewayAdmin(),
			Telemetry: DefaultEnvoyGatewayTelemetry(),
			XDSServer: DefaultXDSServer(),
		},
	}
}

// SetEnvoyGatewayDefaults sets default EnvoyGateway configuration parameters.
func SetEnvoyGatewayDefaults(e *egv1a1.EnvoyGateway) {
	if e.Kind == "" {
		e.Kind = egv1a1.KindEnvoyGateway
	}
	if e.APIVersion == "" {
		e.APIVersion = egv1a1.GroupVersion.String()
	}
	if e.Provider == nil {
		e.Provider = DefaultEnvoyGatewayProvider()
	}

	if e.Provider.Kubernetes == nil {
		e.Provider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{
			EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
				LeaderElection: DefaultLeaderElection(),
			},
		}
	}
	if e.Provider.Kubernetes.LeaderElection == nil {
		e.Provider.Kubernetes.LeaderElection = DefaultLeaderElection()
	}

	if e.Provider.Kubernetes.Client == nil {
		e.Provider.Kubernetes.Client = DefaultKubernetesClient()
	}

	if e.Provider != nil && e.Provider.Custom != nil && e.Provider.Custom.Resource.Type == egv1a1.ResourceProviderTypeKubernetes {
		if e.Provider.Custom.Resource.Kubernetes == nil {
			e.Provider.Custom.Resource.Kubernetes = &egv1a1.EnvoyGatewayKubernetesCustomProvider{}
		}
		if e.Provider.Custom.Resource.Kubernetes.LeaderElection == nil {
			e.Provider.Custom.Resource.Kubernetes.LeaderElection = DefaultLeaderElection()
		}

		if e.Provider.Custom.Resource.Kubernetes.Client == nil {
			e.Provider.Custom.Resource.Kubernetes.Client = DefaultKubernetesClient()
		}
	}

	if e.Gateway == nil {
		e.Gateway = DefaultGateway()
	}
	if e.Logging == nil {
		e.Logging = DefaultEnvoyGatewayLogging()
	}
	if e.Admin == nil {
		e.Admin = DefaultEnvoyGatewayAdmin()
	}
	if e.Telemetry == nil {
		e.Telemetry = DefaultEnvoyGatewayTelemetry()
	}
	if e.XDSServer == nil {
		e.XDSServer = DefaultXDSServer()
	}
}

// GetEnvoyGatewayAdmin returns the EnvoyGatewayAdmin of EnvoyGateway or a default EnvoyGatewayAdmin if unspecified.
func GetEnvoyGatewayAdmin(e *egv1a1.EnvoyGateway) *egv1a1.EnvoyGatewayAdmin {
	if e.Admin != nil {
		if e.Admin.Address == nil {
			e.Admin.Address = DefaultEnvoyGatewayAdminAddress()
		}
		return e.Admin
	}
	e.Admin = DefaultEnvoyGatewayAdmin()

	return e.Admin
}

// GetEnvoyGatewayAdminAddress returns the EnvoyGateway Admin Address.
func GetEnvoyGatewayAdminAddress(e *egv1a1.EnvoyGateway) string {
	address := GetEnvoyGatewayAdmin(e).Address
	if address != nil {
		return net.JoinHostPort(address.Host, strconv.Itoa(address.Port))
	}

	return ""
}

// WatchesNamespaces returns true when Envoy Gateway is configured to watch specific Kubernetes namespaces.
func WatchesNamespaces(e *egv1a1.EnvoyGateway) bool {
	if e.Provider == nil || !IsRunningOnKubernetes(e.Provider) {
		return false
	}

	cfg := GetKubernetesConfiguration(e.Provider)

	return cfg.Watch != nil &&
		cfg.Watch.Type == egv1a1.KubernetesWatchModeTypeNamespaces &&
		len(cfg.Watch.Namespaces) > 0
}

// GatewayNamespaceMode returns true if controller uses gateway namespace mode for infra deployments.
func GatewayNamespaceMode(e *egv1a1.EnvoyGateway) bool {
	return e.Provider != nil &&
		e.Provider.Kubernetes != nil &&
		e.Provider.Kubernetes.Deploy != nil &&
		e.Provider.Kubernetes.Deploy.Type != nil &&
		*e.Provider.Kubernetes.Deploy.Type == egv1a1.KubernetesDeployModeTypeGatewayNamespace
}

// TopologyInjectorDisabled checks whether the provided EnvoyGateway disables TopologyInjector
func TopologyInjectorDisabled(e *egv1a1.EnvoyGateway) bool {
	if e.Provider != nil &&
		e.Provider.Kubernetes != nil &&
		e.Provider.Kubernetes.TopologyInjector != nil {
		return ptr.Deref(e.Provider.Kubernetes.TopologyInjector.Disable, false)
	}
	return false
}

// LuaDisabled returns true if Lua EnvoyExtensionPolicies should be disabled.
// EnableLua takes precedence over the deprecated DisableLua field.
// When neither is set, Lua is disabled by default.
func LuaDisabled(e *egv1a1.ExtensionAPISettings) bool {
	if e == nil {
		return true
	}
	if e.EnableLua {
		return false
	}
	if e.DisableLua != nil {
		return *e.DisableLua
	}
	// Default: Lua is disabled
	return true
}

// GetEnvoyProxyDefaultSpec returns the default EnvoyProxySpec if specified,
// otherwise returns nil.
func GetEnvoyProxyDefaultSpec(e *egv1a1.EnvoyGateway) *egv1a1.EnvoyProxySpec {
	return e.EnvoyProxy
}

// GetRateLimitServiceSettings returns the settings of the rate limit service cluster, or nil if unset.
func GetRateLimitServiceSettings(r *egv1a1.RateLimit) *egv1a1.ClusterSettings {
	if r == nil {
		return nil
	}
	return r.ServiceSettings
}

// defaultRuntimeFlags are the default runtime flags for Envoy Gateway.
var defaultRuntimeFlags = map[egv1a1.RuntimeFlag]bool{
	egv1a1.XDSNameSchemeV2:         false,
	egv1a1.EndpointSliceIndex:      true,
	egv1a1.EnvoyProxyPatch:         true,
	egv1a1.OIDCTokenCookieChunking: true,
}

// IsGatewayAPIEnabled checks if an experimental Gateway API is enabled in the EnvoyGateway configuration.
func IsGatewayAPIEnabled(f *egv1a1.GatewayAPISettings, api egv1a1.GatewayAPI) bool {
	if f != nil {
		for _, enable := range f.Enabled {
			if enable == api {
				return true
			}
		}
	}

	return false
}

// IsRuntimeFlagEnabled checks if a runtime flag is enabled in the EnvoyGateway configuration.
func IsRuntimeFlagEnabled(f *egv1a1.RuntimeFlags, flag egv1a1.RuntimeFlag) bool {
	if f != nil {
		for _, disable := range f.Disabled {
			if disable == flag {
				return false
			}
		}
		for _, enable := range f.Enabled {
			if enable == flag {
				return true
			}
		}
	}

	if defaultValue, found := defaultRuntimeFlags[flag]; found {
		return defaultValue
	}
	return false
}

// DefaultLeaderElection returns a new LeaderElection with default configuration parameters.
func DefaultLeaderElection() *egv1a1.LeaderElection {
	return &egv1a1.LeaderElection{
		RenewDeadline: new(gwapiv1.Duration("10s")),
		RetryPeriod:   new(gwapiv1.Duration("2s")),
		LeaseDuration: new(gwapiv1.Duration("15s")),
		Disable:       new(false),
	}
}

// DefaultKubernetesClient returns a new DefaultKubernetesClient with default parameters.
func DefaultKubernetesClient() *egv1a1.KubernetesClient {
	return &egv1a1.KubernetesClient{
		RateLimit: &egv1a1.KubernetesClientRateLimit{
			QPS:   new(egv1a1.DefaultKubernetesClientQPS),
			Burst: new(egv1a1.DefaultKubernetesClientBurst),
		},
	}
}

// DefaultGateway returns a new Gateway with default configuration parameters.
func DefaultGateway() *egv1a1.Gateway {
	return &egv1a1.Gateway{
		ControllerName: egv1a1.GatewayControllerName,
	}
}

// DefaultXDSServer returns a new XDSServer with default configuration parameters.
func DefaultXDSServer() *egv1a1.XDSServer {
	return &egv1a1.XDSServer{}
}

const (
	// DefaultDebounceAfter is the default quiet period before a pending batch of
	// resource changes is flushed.
	DefaultDebounceAfter = 100 * time.Millisecond

	// DefaultDebounceMax is the default upper bound on how long a resource change
	// may be held before a flush is forced.
	DefaultDebounceMax = 10 * time.Second
)

// IsDebounceEnabled reports whether debouncing of resource changes is turned on.
// Debouncing is opt in, so defining the config is what turns it on.
func IsDebounceEnabled(d *egv1a1.Debounce) bool {
	return d != nil
}

// ResolveDebounceDurations returns the quiet period and the maximum hold time, substituting
// the defaults for any field left unset. It is the single place these fields are
// parsed, so every consumer agrees on what a given configuration means.
func ResolveDebounceDurations(d *egv1a1.Debounce) (after, maxHold time.Duration, err error) {
	after, maxHold = DefaultDebounceAfter, DefaultDebounceMax
	if d == nil {
		return after, maxHold, nil
	}

	if d.After != nil {
		if after, err = time.ParseDuration(string(*d.After)); err != nil {
			return 0, 0, fmt.Errorf("invalid debounce.after: %w", err)
		}
		if after <= 0 {
			return 0, 0, fmt.Errorf("debounce.after must be greater than zero")
		}
	}

	if d.Max != nil {
		if maxHold, err = time.ParseDuration(string(*d.Max)); err != nil {
			return 0, 0, fmt.Errorf("invalid debounce.max: %w", err)
		}
		if maxHold <= 0 {
			return 0, 0, fmt.Errorf("debounce.max must be greater than zero")
		}
	}

	return after, maxHold, nil
}

// DefaultEnvoyGatewayLogging returns a new EnvoyGatewayLogging with default configuration parameters.
func DefaultEnvoyGatewayLogging() *egv1a1.EnvoyGatewayLogging {
	return &egv1a1.EnvoyGatewayLogging{
		Level: map[egv1a1.EnvoyGatewayLogComponent]egv1a1.LogLevel{
			egv1a1.LogComponentGatewayDefault: egv1a1.LogLevelInfo,
		},
	}
}

// GetEnvoyGatewayTelemetry returns the EnvoyGatewayTelemetry of EnvoyGateway or a default EnvoyGatewayTelemetry if unspecified.
func GetEnvoyGatewayTelemetry(e *egv1a1.EnvoyGateway) *egv1a1.EnvoyGatewayTelemetry {
	if e.Telemetry != nil {
		if e.Telemetry.Metrics == nil {
			e.Telemetry.Metrics = DefaultEnvoyGatewayMetrics()
		}
		if e.Telemetry.Metrics.Prometheus == nil {
			e.Telemetry.Metrics.Prometheus = DefaultEnvoyGatewayPrometheus()
		}
		return e.Telemetry
	}
	e.Telemetry = DefaultEnvoyGatewayTelemetry()

	return e.Telemetry
}

// DisablePrometheus returns if disable prometheus.
func DisablePrometheus(e *egv1a1.EnvoyGateway) bool {
	return GetEnvoyGatewayTelemetry(e).Metrics.Prometheus.Disable
}

// DefaultEnvoyGatewayTelemetry returns a new EnvoyGatewayTelemetry with default configuration parameters.
func DefaultEnvoyGatewayTelemetry() *egv1a1.EnvoyGatewayTelemetry {
	return &egv1a1.EnvoyGatewayTelemetry{
		Metrics: DefaultEnvoyGatewayMetrics(),
	}
}

// DefaultEnvoyGatewayMetrics returns a new EnvoyGatewayMetrics with default configuration parameters.
func DefaultEnvoyGatewayMetrics() *egv1a1.EnvoyGatewayMetrics {
	return &egv1a1.EnvoyGatewayMetrics{
		Prometheus: DefaultEnvoyGatewayPrometheus(),
	}
}

// DefaultEnvoyGatewayPrometheus returns a new EnvoyGatewayMetrics with default configuration parameters.
func DefaultEnvoyGatewayPrometheus() *egv1a1.EnvoyGatewayPrometheusProvider {
	return &egv1a1.EnvoyGatewayPrometheusProvider{
		// Enable prometheus pull by default.
		Disable: false,
	}
}

// DefaultEnvoyGatewayProvider returns a new EnvoyGatewayProvider with default configuration parameters.
func DefaultEnvoyGatewayProvider() *egv1a1.EnvoyGatewayProvider {
	return &egv1a1.EnvoyGatewayProvider{
		Type: egv1a1.ProviderTypeKubernetes,
		Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
			EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
				LeaderElection: DefaultLeaderElection(),
				Client:         DefaultKubernetesClient(),
			},
		},
	}
}

// GetEnvoyGatewayProvider returns the EnvoyGatewayProvider of EnvoyGateway or a default EnvoyGatewayProvider if unspecified.
func GetEnvoyGatewayProvider(e *egv1a1.EnvoyGateway) *egv1a1.EnvoyGatewayProvider {
	if e.Provider != nil {
		return e.Provider
	}
	e.Provider = DefaultEnvoyGatewayProvider()

	return e.Provider
}

// DefaultEnvoyGatewayKubeProvider returns a new EnvoyGatewayKubernetesProvider with default settings.
func DefaultEnvoyGatewayKubeProvider() *egv1a1.EnvoyGatewayKubernetesProvider {
	return &egv1a1.EnvoyGatewayKubernetesProvider{
		EnvoyGatewayKubernetesInfrastructureConfiguration: egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration{
			RateLimitDeployment: DefaultKubernetesDeployment(egv1a1.DefaultRateLimitImage),
		},
	}
}

// DefaultEnvoyGatewayAdmin returns a new EnvoyGatewayAdmin with default configuration parameters.
func DefaultEnvoyGatewayAdmin() *egv1a1.EnvoyGatewayAdmin {
	return &egv1a1.EnvoyGatewayAdmin{
		Address:          DefaultEnvoyGatewayAdminAddress(),
		EnableDumpConfig: false,
		EnablePprof:      false,
	}
}

// DefaultEnvoyGatewayAdminAddress returns a new EnvoyGatewayAdminAddress with default configuration parameters.
func DefaultEnvoyGatewayAdminAddress() *egv1a1.EnvoyGatewayAdminAddress {
	return &egv1a1.EnvoyGatewayAdminAddress{
		Port: egv1a1.GatewayAdminPort,
		Host: egv1a1.GatewayAdminHost,
	}
}

// GetEnvoyGatewayKubeProvider returns the EnvoyGatewayKubernetesProvider of Provider or
// a default EnvoyGatewayKubernetesProvider if unspecified. If EnvoyGatewayProvider is not of
// type "Kubernetes", a nil EnvoyGatewayKubernetesProvider is returned.
func GetEnvoyGatewayKubeProvider(r *egv1a1.EnvoyGatewayProvider) *egv1a1.EnvoyGatewayKubernetesProvider {
	if !IsRunningOnKubernetes(r) {
		return nil
	}

	if r.Kubernetes == nil {
		r.Kubernetes = DefaultEnvoyGatewayKubeProvider()
		if r.Kubernetes.LeaderElection == nil {
			r.Kubernetes.LeaderElection = DefaultLeaderElection()
		}
		if r.Kubernetes.Client == nil {
			r.Kubernetes.Client = DefaultKubernetesClient()
		}
		return r.Kubernetes
	}

	if r.Kubernetes.LeaderElection == nil {
		r.Kubernetes.LeaderElection = DefaultLeaderElection()
	}

	if r.Kubernetes.Client == nil {
		r.Kubernetes.Client = DefaultKubernetesClient()
	}

	if r.Kubernetes.RateLimitDeployment == nil {
		r.Kubernetes.RateLimitDeployment = DefaultKubernetesDeployment(egv1a1.DefaultRateLimitImage)
	}

	defaultKubernetesDeploymentSpec(r.Kubernetes.RateLimitDeployment, egv1a1.DefaultRateLimitImage)

	if r.Kubernetes.RateLimitHpa != nil {
		setHPADefaults(r.Kubernetes.RateLimitHpa)
	}

	if r.Kubernetes.ShutdownManager == nil {
		r.Kubernetes.ShutdownManager = &egv1a1.ShutdownManager{Image: new(egv1a1.DefaultShutdownManagerImage)}
	}

	return r.Kubernetes
}

// IsRunningOnKubernetes returns true if Envoy Gateway reads its resources from Kubernetes.
func IsRunningOnKubernetes(r *egv1a1.EnvoyGatewayProvider) bool {
	if r.Type == egv1a1.ProviderTypeKubernetes {
		return true
	}
	if r.Type == egv1a1.ProviderTypeCustom && r.Custom != nil {
		return r.Custom.Resource.Type == egv1a1.ResourceProviderTypeKubernetes
	}
	return false
}

// IsInfraManagedRemotely returns true if a custom provider hands infrastructure to a remote manager.
func IsInfraManagedRemotely(r *egv1a1.EnvoyGatewayProvider) bool {
	if r.Type != egv1a1.ProviderTypeCustom {
		return false
	}

	if r.Custom == nil || r.Custom.Infrastructure == nil {
		return false
	}

	return r.Custom.Infrastructure.Type == egv1a1.InfrastructureProviderTypeRemote
}

// IsRunningOnHost returns true if a custom provider runs the infrastructure on the host.
func IsRunningOnHost(r *egv1a1.EnvoyGatewayProvider) bool {
	return r.Type == egv1a1.ProviderTypeCustom &&
		r.Custom.Infrastructure != nil &&
		r.Custom.Infrastructure.Type == egv1a1.InfrastructureProviderTypeHost
}

// GetKubernetesConfiguration returns the Kubernetes configuration of the provider, or an empty one if unset.
func GetKubernetesConfiguration(r *egv1a1.EnvoyGatewayProvider) egv1a1.EnvoyGatewayKubernetesConfiguration {
	if r.Type == egv1a1.ProviderTypeKubernetes && r.Kubernetes != nil {
		return r.Kubernetes.EnvoyGatewayKubernetesConfiguration
	}

	if r.Type == egv1a1.ProviderTypeCustom && r.Custom != nil &&
		r.Custom.Resource.Type == egv1a1.ResourceProviderTypeKubernetes &&
		r.Custom.Resource.Kubernetes != nil {
		return r.Custom.Resource.Kubernetes.EnvoyGatewayKubernetesConfiguration
	}
	return egv1a1.EnvoyGatewayKubernetesConfiguration{}
}

// GetKubernetesInfrastructureConfiguration returns the Kubernetes infrastructure configuration
// of the provider, or an empty one if unset.
func GetKubernetesInfrastructureConfiguration(r *egv1a1.EnvoyGatewayProvider) egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration {
	if r.Type == egv1a1.ProviderTypeKubernetes && r.Kubernetes != nil {
		return r.Kubernetes.EnvoyGatewayKubernetesInfrastructureConfiguration
	}
	return egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration{}
}

// DefaultEnvoyGatewayLoggingLevel returns a new EnvoyGatewayLogging with default configuration parameters.
// When v1alpha1.LogComponentGatewayDefault specified, all other logging components are ignored.
func DefaultEnvoyGatewayLoggingLevel(logging *egv1a1.EnvoyGatewayLogging, level egv1a1.LogLevel) egv1a1.LogLevel {
	if level != "" {
		return level
	}

	if logging.Level[egv1a1.LogComponentGatewayDefault] != "" {
		return logging.Level[egv1a1.LogComponentGatewayDefault]
	}

	return egv1a1.LogLevelInfo
}

// SetEnvoyGatewayLoggingDefaults sets default EnvoyGatewayLogging configuration parameters.
func SetEnvoyGatewayLoggingDefaults(logging *egv1a1.EnvoyGatewayLogging) {
	if logging != nil && logging.Level != nil && logging.Level[egv1a1.LogComponentGatewayDefault] == "" {
		logging.Level[egv1a1.LogComponentGatewayDefault] = egv1a1.LogLevelInfo
	}
}

// GetQPSAndBurst returns the Kubernetes client QPS and burst, falling back to the defaults.
func GetQPSAndBurst(kcr *egv1a1.KubernetesClientRateLimit) (float32, int) {
	if kcr == nil {
		return float32(egv1a1.DefaultKubernetesClientQPS), int(egv1a1.DefaultKubernetesClientBurst)
	}
	qps := ptr.Deref(kcr.QPS, egv1a1.DefaultKubernetesClientQPS)
	burst := ptr.Deref(kcr.Burst, egv1a1.DefaultKubernetesClientBurst)
	return float32(qps), int(burst)
}

// GetExtensionManagers normalizes the singular ExtensionManager and plural ExtensionManagers
// fields into a single list. The plural field takes precedence. If only the singular field
// is set, it is returned as a single-element list. Returns nil if neither is set.
func GetExtensionManagers(e *egv1a1.EnvoyGatewaySpec) []egv1a1.ExtensionManager {
	if len(e.ExtensionManagers) > 0 {
		return e.ExtensionManagers
	}
	if e.ExtensionManager != nil {
		return []egv1a1.ExtensionManager{*e.ExtensionManager}
	}
	return nil
}

// ShouldIncludeClusters returns true if clusters should be included in the translation hook.
// When TranslationConfig is nil, defaults to true for backward compatibility.
// When TranslationConfig is explicitly set, uses the configuration.
func ShouldIncludeClusters(tc *egv1a1.TranslationConfig) bool {
	if tc == nil || tc.Cluster == nil {
		// Default behavior: include clusters for backward compatibility
		return true
	}

	return ptr.Deref(tc.Cluster.IncludeAll, true)
}

// ShouldIncludeSecrets returns true if secrets should be included in the translation hook.
// When TranslationConfig is nil, defaults to true for backward compatibility.
// When TranslationConfig is explicitly set, uses the configuration.
func ShouldIncludeSecrets(tc *egv1a1.TranslationConfig) bool {
	if tc == nil || tc.Secret == nil {
		// Default behavior: include secrets for backward compatibility
		return true
	}

	return ptr.Deref(tc.Secret.IncludeAll, true)
}

// ShouldIncludeListeners returns true if listeners should be included in the translation hook.
// When TranslationConfig is nil, defaults to false for backward compatibility.
// When TranslationConfig is explicitly set, uses the configuration.
func ShouldIncludeListeners(tc *egv1a1.TranslationConfig) bool {
	if tc == nil || tc.Listener == nil {
		// Default behavior: exclude listeners for backward compatibility
		return false
	}

	return ptr.Deref(tc.Listener.IncludeAll, false)
}

// ShouldIncludeRoutes returns true if routes should be included in the translation hook.
// When TranslationConfig is nil, defaults to false for backward compatibility.
// When TranslationConfig is explicitly set, uses the configuration.
func ShouldIncludeRoutes(tc *egv1a1.TranslationConfig) bool {
	if tc == nil || tc.Route == nil {
		// Default behavior: exclude routes for backward compatibility
		return false
	}

	return ptr.Deref(tc.Route.IncludeAll, false)
}
