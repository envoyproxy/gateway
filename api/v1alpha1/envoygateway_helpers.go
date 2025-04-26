// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"net"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// DefaultEnvoyGateway returns a new EnvoyGateway with default configuration parameters.
func DefaultEnvoyGateway() *EnvoyGateway {
	return &EnvoyGateway{
		metav1.TypeMeta{
			Kind:       KindEnvoyGateway,
			APIVersion: GroupVersion.String(),
		},
		EnvoyGatewaySpec{
			Gateway:   DefaultGateway(),
			Provider:  DefaultEnvoyGatewayProvider(),
			Logging:   DefaultEnvoyGatewayLogging(),
			Admin:     DefaultEnvoyGatewayAdmin(),
			Telemetry: DefaultEnvoyGatewayTelemetry(),
		},
	}
}

// SetEnvoyGatewayDefaults sets default EnvoyGateway configuration parameters.
func (e *EnvoyGateway) SetEnvoyGatewayDefaults() {
	if e.TypeMeta.Kind == "" {
		e.TypeMeta.Kind = KindEnvoyGateway
	}
	if e.TypeMeta.APIVersion == "" {
		e.TypeMeta.APIVersion = GroupVersion.String()
	}
	if e.Provider == nil {
		e.Provider = DefaultEnvoyGatewayProvider()
	}
	if e.Provider.Kubernetes == nil {
		e.Provider.Kubernetes = &EnvoyGatewayKubernetesProvider{
			LeaderElection: DefaultLeaderElection(),
		}
	}
	if e.Provider.Kubernetes.LeaderElection == nil {
		e.Provider.Kubernetes.LeaderElection = DefaultLeaderElection()
	}

	if e.Provider.Kubernetes.Client == nil {
		e.Provider.Kubernetes.Client = DefaultKubernetesClient()
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
}

// GetEnvoyGatewayAdmin returns the EnvoyGatewayAdmin of EnvoyGateway or a default EnvoyGatewayAdmin if unspecified.
func (e *EnvoyGateway) GetEnvoyGatewayAdmin() *EnvoyGatewayAdmin {
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
func (e *EnvoyGateway) GetEnvoyGatewayAdminAddress() string {
	address := e.GetEnvoyGatewayAdmin().Address
	if address != nil {
		return net.JoinHostPort(address.Host, strconv.Itoa(address.Port))
	}

	return ""
}

// NamespaceMode returns if uses namespace mode.
func (e *EnvoyGateway) NamespaceMode() bool {
	return e.Provider != nil &&
		e.Provider.Kubernetes != nil &&
		e.Provider.Kubernetes.Watch != nil &&
		e.Provider.Kubernetes.Watch.Type == KubernetesWatchModeTypeNamespaces &&
		len(e.Provider.Kubernetes.Watch.Namespaces) > 0
}

// DefaultLeaderElection returns a new LeaderElection with default configuration parameters.
func DefaultLeaderElection() *LeaderElection {
	return &LeaderElection{
		RenewDeadline: ptr.To(gwapiv1.Duration("10s")),
		RetryPeriod:   ptr.To(gwapiv1.Duration("2s")),
		LeaseDuration: ptr.To(gwapiv1.Duration("15s")),
		Disable:       ptr.To(false),
	}
}

// DefaultKubernetesClient returns a new DefaultKubernetesClient with default parameters.
func DefaultKubernetesClient() *KubernetesClient {
	return &KubernetesClient{
		RateLimit: &KubernetesClientRateLimit{
			QPS:   ptr.To(int32(50)),
			Burst: ptr.To(int32(100)),
		},
	}
}

// DefaultGateway returns a new Gateway with default configuration parameters.
func DefaultGateway() *Gateway {
	return &Gateway{
		ControllerName: GatewayControllerName,
	}
}

// DefaultEnvoyGatewayLogging returns a new EnvoyGatewayLogging with default configuration parameters.
func DefaultEnvoyGatewayLogging() *EnvoyGatewayLogging {
	return &EnvoyGatewayLogging{
		Level: map[EnvoyGatewayLogComponent]LogLevel{
			LogComponentGatewayDefault: LogLevelInfo,
		},
	}
}

// GetEnvoyGatewayTelemetry returns the EnvoyGatewayTelemetry of EnvoyGateway or a default EnvoyGatewayTelemetry if unspecified.
func (e *EnvoyGateway) GetEnvoyGatewayTelemetry() *EnvoyGatewayTelemetry {
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
func (e *EnvoyGateway) DisablePrometheus() bool {
	return e.GetEnvoyGatewayTelemetry().Metrics.Prometheus.Disable
}

// DefaultEnvoyGatewayTelemetry returns a new EnvoyGatewayTelemetry with default configuration parameters.
func DefaultEnvoyGatewayTelemetry() *EnvoyGatewayTelemetry {
	return &EnvoyGatewayTelemetry{
		Metrics: DefaultEnvoyGatewayMetrics(),
	}
}

// DefaultEnvoyGatewayMetrics returns a new EnvoyGatewayMetrics with default configuration parameters.
func DefaultEnvoyGatewayMetrics() *EnvoyGatewayMetrics {
	return &EnvoyGatewayMetrics{
		Prometheus: DefaultEnvoyGatewayPrometheus(),
	}
}

// DefaultEnvoyGatewayPrometheus returns a new EnvoyGatewayMetrics with default configuration parameters.
func DefaultEnvoyGatewayPrometheus() *EnvoyGatewayPrometheusProvider {
	return &EnvoyGatewayPrometheusProvider{
		// Enable prometheus pull by default.
		Disable: false,
	}
}

// DefaultEnvoyGatewayProvider returns a new EnvoyGatewayProvider with default configuration parameters.
func DefaultEnvoyGatewayProvider() *EnvoyGatewayProvider {
	return &EnvoyGatewayProvider{
		Type: ProviderTypeKubernetes,
		Kubernetes: &EnvoyGatewayKubernetesProvider{
			LeaderElection: DefaultLeaderElection(),
		},
	}
}

// GetEnvoyGatewayProvider returns the EnvoyGatewayProvider of EnvoyGateway or a default EnvoyGatewayProvider if unspecified.
func (e *EnvoyGateway) GetEnvoyGatewayProvider() *EnvoyGatewayProvider {
	if e.Provider != nil {
		return e.Provider
	}
	e.Provider = DefaultEnvoyGatewayProvider()

	return e.Provider
}

// DefaultEnvoyGatewayKubeProvider returns a new EnvoyGatewayKubernetesProvider with default settings.
func DefaultEnvoyGatewayKubeProvider() *EnvoyGatewayKubernetesProvider {
	return &EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: DefaultKubernetesDeployment(DefaultRateLimitImage),
	}
}

// DefaultEnvoyGatewayAdmin returns a new EnvoyGatewayAdmin with default configuration parameters.
func DefaultEnvoyGatewayAdmin() *EnvoyGatewayAdmin {
	return &EnvoyGatewayAdmin{
		Address:          DefaultEnvoyGatewayAdminAddress(),
		EnableDumpConfig: false,
		EnablePprof:      false,
	}
}

// DefaultEnvoyGatewayAdminAddress returns a new EnvoyGatewayAdminAddress with default configuration parameters.
func DefaultEnvoyGatewayAdminAddress() *EnvoyGatewayAdminAddress {
	return &EnvoyGatewayAdminAddress{
		Port: GatewayAdminPort,
		Host: GatewayAdminHost,
	}
}

// GetEnvoyGatewayKubeProvider returns the EnvoyGatewayKubernetesProvider of Provider or
// a default EnvoyGatewayKubernetesProvider if unspecified. If EnvoyGatewayProvider is not of
// type "Kubernetes", a nil EnvoyGatewayKubernetesProvider is returned.
func (r *EnvoyGatewayProvider) GetEnvoyGatewayKubeProvider() *EnvoyGatewayKubernetesProvider {
	if r.Type != ProviderTypeKubernetes {
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

	if r.Kubernetes.RateLimitDeployment == nil {
		r.Kubernetes.RateLimitDeployment = DefaultKubernetesDeployment(DefaultRateLimitImage)
	}

	r.Kubernetes.RateLimitDeployment.defaultKubernetesDeploymentSpec(DefaultRateLimitImage)

	if r.Kubernetes.RateLimitHpa != nil {
		r.Kubernetes.RateLimitHpa.setDefault()
	}

	if r.Kubernetes.ShutdownManager == nil {
		r.Kubernetes.ShutdownManager = &ShutdownManager{Image: ptr.To(DefaultShutdownManagerImage)}
	}

	return r.Kubernetes
}

func (r *EnvoyGatewayProvider) IsRunningOnKubernetes() bool {
	return r.Type == ProviderTypeKubernetes
}

func (r *EnvoyGatewayProvider) IsRunningOnHost() bool {
	return r.Type == ProviderTypeCustom &&
		r.Custom.Infrastructure != nil &&
		r.Custom.Infrastructure.Type == InfrastructureProviderTypeHost
}

// DefaultEnvoyGatewayLoggingLevel returns a new EnvoyGatewayLogging with default configuration parameters.
// When v1alpha1.LogComponentGatewayDefault specified, all other logging components are ignored.
func (logging *EnvoyGatewayLogging) DefaultEnvoyGatewayLoggingLevel(level LogLevel) LogLevel {
	if level != "" {
		return level
	}

	if logging.Level[LogComponentGatewayDefault] != "" {
		return logging.Level[LogComponentGatewayDefault]
	}

	return LogLevelInfo
}

// SetEnvoyGatewayLoggingDefaults sets default EnvoyGatewayLogging configuration parameters.
func (logging *EnvoyGatewayLogging) SetEnvoyGatewayLoggingDefaults() {
	if logging != nil && logging.Level != nil && logging.Level[LogComponentGatewayDefault] == "" {
		logging.Level[LogComponentGatewayDefault] = LogLevelInfo
	}
}
