// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultEnvoyGateway returns a new EnvoyGateway with default configuration parameters.
func DefaultEnvoyGateway() *EnvoyGateway {
	return &EnvoyGateway{
		metav1.TypeMeta{
			Kind:       KindEnvoyGateway,
			APIVersion: GroupVersion.String(),
		},
		EnvoyGatewaySpec{
			Gateway:  DefaultGateway(),
			Provider: DefaultEnvoyGatewayProvider(),
			Logging:  DefaultEnvoyGatewayLogging(),
			Admin:    DefaultEnvoyGatewayAdmin(),
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
	if e.Gateway == nil {
		e.Gateway = DefaultGateway()
	}
	if e.Logging == nil {
		e.Logging = DefaultEnvoyGatewayLogging()
	}
	if e.Admin == nil {
		e.Admin = DefaultEnvoyGatewayAdmin()
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

// DefaultEnvoyGatewayProvider returns a new EnvoyGatewayProvider with default configuration parameters.
func DefaultEnvoyGatewayProvider() *EnvoyGatewayProvider {
	return &EnvoyGatewayProvider{
		Type: ProviderTypeKubernetes,
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
		Debug:   false,
		Address: DefaultEnvoyGatewayAdminAddress(),
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
		return r.Kubernetes
	}

	if r.Kubernetes.RateLimitDeployment == nil {
		r.Kubernetes.RateLimitDeployment = DefaultKubernetesDeployment(DefaultRateLimitImage)
	}

	r.Kubernetes.RateLimitDeployment.defaultKubernetesDeploymentSpec(DefaultRateLimitImage)

	return r.Kubernetes
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
