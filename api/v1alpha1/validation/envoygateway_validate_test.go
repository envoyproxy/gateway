// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

var (
	TLSSecretKind       = gwapiv1.Kind("Secret")
	TLSUnrecognizedKind = gwapiv1.Kind("Unrecognized")
)

func TestValidateEnvoyGateway(t *testing.T) {
	eg := egv1a1.DefaultEnvoyGateway()

	testCases := []struct {
		name   string
		eg     *egv1a1.EnvoyGateway
		expect bool
	}{
		{
			name:   "default",
			eg:     eg,
			expect: true,
		},
		{
			name:   "unspecified envoy gateway",
			eg:     nil,
			expect: false,
		},
		{
			name: "unspecified gateway",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: false,
		},
		{
			name: "unspecified provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			name: "empty gateway controllerName",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  &egv1a1.Gateway{ControllerName: ""},
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: false,
		},
		{
			name: "nil custom provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type:   egv1a1.ProviderTypeCustom,
						Custom: nil,
					},
				},
			},
			expect: false,
		},
		{
			name: "empty custom provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type:   egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{},
					},
				},
			},
			expect: false,
		},
		{
			name: "custom provider with file resource provider and host infra provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{
							Resource: egv1a1.EnvoyGatewayResourceProvider{
								Type: egv1a1.ResourceProviderTypeFile,
								File: &egv1a1.EnvoyGatewayFileResourceProvider{
									Paths: []string{"foo", "bar"},
								},
							},
							Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
								Type: egv1a1.InfrastructureProviderTypeHost,
								Host: &egv1a1.EnvoyGatewayHostInfrastructureProvider{},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "custom provider with file provider and k8s infra provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{
							Resource: egv1a1.EnvoyGatewayResourceProvider{
								Type: egv1a1.ResourceProviderTypeFile,
								File: &egv1a1.EnvoyGatewayFileResourceProvider{
									Paths: []string{"foo", "bar"},
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "custom provider with unsupported resource provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{
							Resource: egv1a1.EnvoyGatewayResourceProvider{
								Type: "foobar",
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "custom provider with file provider but no file struct",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{
							Resource: egv1a1.EnvoyGatewayResourceProvider{
								Type: egv1a1.ResourceProviderTypeFile,
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "custom provider with file provider and host infra provider but no host struct",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{
							Resource: egv1a1.EnvoyGatewayResourceProvider{
								Type: egv1a1.ResourceProviderTypeFile,
								File: &egv1a1.EnvoyGatewayFileResourceProvider{
									Paths: []string{"a", "b"},
								},
							},
							Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
								Type: egv1a1.InfrastructureProviderTypeHost,
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "custom provider with file provider and unsupported infra provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{
							Resource: egv1a1.EnvoyGatewayResourceProvider{
								Type: egv1a1.ResourceProviderTypeFile,
								File: &egv1a1.EnvoyGatewayFileResourceProvider{
									Paths: []string{"a", "b"},
								},
							},
							Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
								Type: "foobar",
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "custom provider with file provider and host infra provider but no paths assign in resource",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{
							Resource: egv1a1.EnvoyGatewayResourceProvider{
								Type: egv1a1.ResourceProviderTypeFile,
								File: &egv1a1.EnvoyGatewayFileResourceProvider{},
							},
							Infrastructure: &egv1a1.EnvoyGatewayInfrastructureProvider{
								Type: egv1a1.InfrastructureProviderTypeHost,
								Host: &egv1a1.EnvoyGatewayHostInfrastructureProvider{},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "empty ratelimit",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:   egv1a1.DefaultGateway(),
					Provider:  egv1a1.DefaultEnvoyGatewayProvider(),
					RateLimit: &egv1a1.RateLimit{},
				},
			},
			expect: false,
		},
		{
			name: "empty ratelimit redis setting",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					RateLimit: &egv1a1.RateLimit{
						Backend: egv1a1.RateLimitDatabaseBackend{
							Type:  egv1a1.RedisBackendType,
							Redis: &egv1a1.RateLimitRedisSettings{},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "unknown ratelimit redis url format",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					RateLimit: &egv1a1.RateLimit{
						Backend: egv1a1.RateLimitDatabaseBackend{
							Type: egv1a1.RedisBackendType,
							Redis: &egv1a1.RateLimitRedisSettings{
								URL: ":foo",
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "happy ratelimit redis settings",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					RateLimit: &egv1a1.RateLimit{
						Backend: egv1a1.RateLimitDatabaseBackend{
							Type: egv1a1.RedisBackendType,
							Redis: &egv1a1.RateLimitRedisSettings{
								URL: "localhost:6376",
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy ratelimit redis sentinel settings",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					RateLimit: &egv1a1.RateLimit{
						Backend: egv1a1.RateLimitDatabaseBackend{
							Type: egv1a1.RedisBackendType,
							Redis: &egv1a1.RateLimitRedisSettings{
								URL: "primary_.-,node-0:26379,node-1:26379",
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy ratelimit redis cluster settings",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					RateLimit: &egv1a1.RateLimit{
						Backend: egv1a1.RateLimitDatabaseBackend{
							Type: egv1a1.RedisBackendType,
							Redis: &egv1a1.RateLimitRedisSettings{
								URL: "node-0:6376,node-1:6376,node-2:6376",
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy extension settings",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &egv1a1.ExtensionHooks{
							XDSTranslator: &egv1a1.XDSTranslatorHooks{
								Pre: []egv1a1.XDSTranslatorHook{},
								Post: []egv1a1.XDSTranslatorHook{
									egv1a1.XDSHTTPListener,
									egv1a1.XDSTranslation,
									egv1a1.XDSRoute,
									egv1a1.XDSVirtualHost,
								},
							},
						},
						Service: &egv1a1.ExtensionService{
							Host: "foo.extension",
							Port: 80,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy extension settings tls",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &egv1a1.ExtensionHooks{
							XDSTranslator: &egv1a1.XDSTranslatorHooks{
								Pre: []egv1a1.XDSTranslatorHook{},
								Post: []egv1a1.XDSTranslatorHook{
									egv1a1.XDSHTTPListener,
									egv1a1.XDSTranslation,
									egv1a1.XDSRoute,
									egv1a1.XDSVirtualHost,
								},
							},
						},
						Service: &egv1a1.ExtensionService{
							Host: "foo.extension",
							Port: 443,
							TLS: &egv1a1.ExtensionTLS{
								CertificateRef: gwapiv1.SecretObjectReference{
									Kind: &TLSSecretKind,
									Name: gwapiv1.ObjectName("certificate"),
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy extension settings no resources",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Hooks: &egv1a1.ExtensionHooks{
							XDSTranslator: &egv1a1.XDSTranslatorHooks{
								Pre: []egv1a1.XDSTranslatorHook{},
								Post: []egv1a1.XDSTranslatorHook{
									egv1a1.XDSHTTPListener,
									egv1a1.XDSTranslation,
									egv1a1.XDSRoute,
									egv1a1.XDSVirtualHost,
								},
							},
						},
						Service: &egv1a1.ExtensionService{
							Host: "foo.extension",
							Port: 443,
							TLS: &egv1a1.ExtensionTLS{
								CertificateRef: gwapiv1.SecretObjectReference{
									Kind: &TLSSecretKind,
									Name: gwapiv1.ObjectName("certificate"),
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "unknown TLS certificateRef in extension settings",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &egv1a1.ExtensionHooks{
							XDSTranslator: &egv1a1.XDSTranslatorHooks{
								Pre: []egv1a1.XDSTranslatorHook{},
								Post: []egv1a1.XDSTranslatorHook{
									egv1a1.XDSHTTPListener,
									egv1a1.XDSTranslation,
									egv1a1.XDSRoute,
									egv1a1.XDSVirtualHost,
								},
							},
						},
						Service: &egv1a1.ExtensionService{
							Host: "foo.extension",
							Port: 8080,
							TLS: &egv1a1.ExtensionTLS{
								CertificateRef: gwapiv1.SecretObjectReference{
									Kind: &TLSUnrecognizedKind,
									Name: gwapiv1.ObjectName("certificate"),
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "empty service in extension settings",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &egv1a1.ExtensionHooks{
							XDSTranslator: &egv1a1.XDSTranslatorHooks{
								Pre: []egv1a1.XDSTranslatorHook{},
								Post: []egv1a1.XDSTranslatorHook{
									egv1a1.XDSHTTPListener,
									egv1a1.XDSTranslation,
									egv1a1.XDSRoute,
									egv1a1.XDSVirtualHost,
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "empty hooks in extension settings",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Service: &egv1a1.ExtensionService{
							Host: "foo.extension",
							Port: 8080,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "valid gateway logging level info",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Logging: &egv1a1.EnvoyGatewayLogging{
						Level: map[egv1a1.EnvoyGatewayLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentGatewayDefault: egv1a1.LogLevelInfo,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level warn",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Logging: &egv1a1.EnvoyGatewayLogging{
						Level: map[egv1a1.EnvoyGatewayLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentGatewayDefault: egv1a1.LogLevelWarn,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level error",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Logging: &egv1a1.EnvoyGatewayLogging{
						Level: map[egv1a1.EnvoyGatewayLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentGatewayDefault: egv1a1.LogLevelError,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level debug",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Logging: &egv1a1.EnvoyGatewayLogging{
						Level: map[egv1a1.EnvoyGatewayLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentGatewayDefault: egv1a1.LogLevelDebug,
							egv1a1.LogComponentProviderRunner: egv1a1.LogLevelDebug,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "invalid gateway logging level",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Logging: &egv1a1.EnvoyGatewayLogging{
						Level: map[egv1a1.EnvoyGatewayLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentGatewayDefault: "inffo",
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "valid gateway metrics sink",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Telemetry: &egv1a1.EnvoyGatewayTelemetry{
						Metrics: &egv1a1.EnvoyGatewayMetrics{
							Sinks: []egv1a1.EnvoyGatewayMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.EnvoyGatewayOpenTelemetrySink{
										Host:     "x.x.x.x",
										Port:     4317,
										Protocol: "grpc",
									},
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "invalid gateway metrics sink",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					Telemetry: &egv1a1.EnvoyGatewayTelemetry{
						Metrics: &egv1a1.EnvoyGatewayMetrics{
							Sinks: []egv1a1.EnvoyGatewayMetricSink{
								{
									Type: egv1a1.MetricSinkTypeOpenTelemetry,
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "invalid gateway watch mode",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Watch: &egv1a1.KubernetesWatchMode{
								Type: "foobar",
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "happy namespaces must be set when watch mode is Namespaces",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Watch: &egv1a1.KubernetesWatchMode{
								Type:       egv1a1.KubernetesWatchModeTypeNamespaces,
								Namespaces: []string{"foo"},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "fail namespaces is not be set when watch mode is Namespaces",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Watch: &egv1a1.KubernetesWatchMode{
								Type:              egv1a1.KubernetesWatchModeTypeNamespaces,
								NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": ""}},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "happy namespaceSelector must be set when watch mode is NamespaceSelector",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Watch: &egv1a1.KubernetesWatchMode{
								Type:              egv1a1.KubernetesWatchModeTypeNamespaceSelector,
								NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"foo": ""}},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "fail namespaceSelector is not be set when watch mode is NamespaceSelector",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway: egv1a1.DefaultGateway(),
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							Watch: &egv1a1.KubernetesWatchMode{
								Type: egv1a1.KubernetesWatchModeTypeNamespaceSelector,
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no extension server target set",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Service: &egv1a1.ExtensionService{
							Port: 8080,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "both host and path targets are set for extension server",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Service: &egv1a1.ExtensionService{
							BackendEndpoint: egv1a1.BackendEndpoint{
								FQDN: &egv1a1.FQDNEndpoint{
									Hostname: "foo.example.com",
									Port:     8080,
								},
								Unix: &egv1a1.UnixSocket{
									Path: "/some/path",
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "multiple backend targets are set for extension server",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Service: &egv1a1.ExtensionService{
							BackendEndpoint: egv1a1.BackendEndpoint{
								FQDN: &egv1a1.FQDNEndpoint{
									Hostname: "foo.example.com",
									Port:     8080,
								},
								IP: &egv1a1.IPEndpoint{
									Address: "10.9.8.7",
									Port:    8080,
								},
							},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "both host and path targets are set for extension server",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &egv1a1.ExtensionManager{
						Resources: []egv1a1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Service: &egv1a1.ExtensionService{
							Host: "foo.example.com",
							Port: 8080,
							BackendEndpoint: egv1a1.BackendEndpoint{
								FQDN: &egv1a1.FQDNEndpoint{
									Hostname: "foo.example.com",
									Port:     8080,
								},
							},
						},
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEnvoyGateway(tc.eg)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEnvoyGateway(t *testing.T) {
	envoyGateway := egv1a1.DefaultEnvoyGateway()
	assert.NotNil(t, envoyGateway.Provider)
	assert.NotNil(t, envoyGateway.Gateway)
	assert.NotNil(t, envoyGateway.Logging)
	envoyGateway.SetEnvoyGatewayDefaults()
	assert.Equal(t, envoyGateway.Logging, egv1a1.DefaultEnvoyGatewayLogging())

	logging := egv1a1.DefaultEnvoyGatewayLogging()
	assert.NotNil(t, logging)
	assert.Equal(t, egv1a1.LogLevelInfo, logging.Level[egv1a1.LogComponentGatewayDefault])

	gatewayLogging := &egv1a1.EnvoyGatewayLogging{
		Level: logging.Level,
	}
	gatewayLogging.SetEnvoyGatewayLoggingDefaults()
	assert.NotNil(t, gatewayLogging)
	assert.Equal(t, egv1a1.LogLevelInfo, gatewayLogging.Level[egv1a1.LogComponentGatewayDefault])
}

func TestDefaultEnvoyGatewayLoggingLevel(t *testing.T) {
	type args struct {
		component string
		level     egv1a1.LogLevel
	}
	tests := []struct {
		name string
		args args
		want egv1a1.LogLevel
	}{
		{
			name: "test default info level for empty level",
			args: args{component: "", level: ""},
			want: egv1a1.LogLevelInfo,
		},
		{
			name: "test default info level for empty level",
			args: args{component: string(egv1a1.LogComponentGatewayDefault), level: ""},
			want: egv1a1.LogLevelInfo,
		},
		{
			name: "test default info level for info level",
			args: args{component: string(egv1a1.LogComponentGatewayDefault), level: egv1a1.LogLevelInfo},
			want: egv1a1.LogLevelInfo,
		},
		{
			name: "test default error level for error level",
			args: args{component: string(egv1a1.LogComponentGatewayDefault), level: egv1a1.LogLevelError},
			want: egv1a1.LogLevelError,
		},
		{
			name: "test gateway-api error level for error level",
			args: args{component: string(egv1a1.LogComponentGatewayAPIRunner), level: egv1a1.LogLevelError},
			want: egv1a1.LogLevelError,
		},
		{
			name: "test gateway-api info level for info level",
			args: args{component: string(egv1a1.LogComponentGatewayAPIRunner), level: egv1a1.LogLevelInfo},
			want: egv1a1.LogLevelInfo,
		},
		{
			name: "test default gateway-api warn level for info level",
			args: args{component: string(egv1a1.LogComponentGatewayAPIRunner), level: ""},
			want: egv1a1.LogLevelInfo,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logging := &egv1a1.EnvoyGatewayLogging{}
			if got := logging.DefaultEnvoyGatewayLoggingLevel(tt.args.level); got != tt.want {
				t.Errorf("defaultLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvoyGatewayProvider(t *testing.T) {
	envoyGateway := &egv1a1.EnvoyGateway{
		TypeMeta:         metav1.TypeMeta{},
		EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{Provider: egv1a1.DefaultEnvoyGatewayProvider()},
	}
	assert.NotNil(t, envoyGateway.Provider)

	envoyGatewayProvider := envoyGateway.GetEnvoyGatewayProvider()
	assert.NotNil(t, envoyGatewayProvider.Kubernetes)
	assert.Equal(t, envoyGateway.Provider, envoyGatewayProvider)

	envoyGatewayProvider.Kubernetes = egv1a1.DefaultEnvoyGatewayKubeProvider()
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, egv1a1.DefaultKubernetesDeployment(egv1a1.DefaultRateLimitImage))

	envoyGatewayProvider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{}
	assert.Nil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment)

	envoyGatewayProvider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: &egv1a1.KubernetesDeploymentSpec{
			Replicas:  nil,
			Pod:       nil,
			Container: nil,
		},
	}
	assert.Nil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Replicas)
	assert.Nil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod)
	assert.Nil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container)
	envoyGatewayKubeProvider := envoyGatewayProvider.GetEnvoyGatewayKubeProvider()

	envoyGatewayProvider.Kubernetes = &egv1a1.EnvoyGatewayKubernetesProvider{
		RateLimitDeployment: &egv1a1.KubernetesDeploymentSpec{
			Pod: nil,
			Container: &egv1a1.KubernetesContainerSpec{
				Resources:       nil,
				SecurityContext: nil,
				Image:           nil,
			},
		},
	}
	assert.Nil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources)
	envoyGatewayProvider.GetEnvoyGatewayKubeProvider()

	assert.NotNil(t, envoyGatewayProvider.Kubernetes)
	assert.Equal(t, envoyGatewayProvider.Kubernetes, envoyGatewayKubeProvider)

	assert.NotNil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment, egv1a1.DefaultKubernetesDeployment(egv1a1.DefaultRateLimitImage))
	assert.NotNil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Pod, egv1a1.DefaultKubernetesPod())
	assert.NotNil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container, egv1a1.DefaultKubernetesContainer(egv1a1.DefaultRateLimitImage))
	assert.NotNil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Resources, egv1a1.DefaultResourceRequirements())
	assert.NotNil(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image)
	assert.Equal(t, envoyGatewayProvider.Kubernetes.RateLimitDeployment.Container.Image, egv1a1.DefaultKubernetesContainerImage(egv1a1.DefaultRateLimitImage))
}

func TestEnvoyGatewayAdmin(t *testing.T) {
	// default envoygateway config admin should not be nil
	eg := egv1a1.DefaultEnvoyGateway()
	assert.NotNil(t, eg.Admin)

	// get default admin config from envoygateway
	// values should be set in default
	egAdmin := eg.GetEnvoyGatewayAdmin()
	assert.NotNil(t, egAdmin)
	assert.Equal(t, egv1a1.GatewayAdminPort, egAdmin.Address.Port)
	assert.Equal(t, egv1a1.GatewayAdminHost, egAdmin.Address.Host)
	assert.False(t, egAdmin.EnableDumpConfig)
	assert.False(t, egAdmin.EnablePprof)

	// override the admin config
	// values should be updated
	eg.Admin = &egv1a1.EnvoyGatewayAdmin{
		Address: &egv1a1.EnvoyGatewayAdminAddress{
			Host: "0.0.0.0",
			Port: 19010,
		},
		EnableDumpConfig: true,
		EnablePprof:      true,
	}

	assert.Equal(t, 19010, eg.GetEnvoyGatewayAdmin().Address.Port)
	assert.Equal(t, "0.0.0.0", eg.GetEnvoyGatewayAdmin().Address.Host)
	assert.True(t, eg.GetEnvoyGatewayAdmin().EnableDumpConfig)
	assert.True(t, eg.GetEnvoyGatewayAdmin().EnablePprof)

	// set eg defaults when admin is nil
	// the admin should not be nil
	eg.Admin = nil
	eg.SetEnvoyGatewayDefaults()
	assert.NotNil(t, eg.Admin)
	assert.Equal(t, egv1a1.GatewayAdminPort, eg.Admin.Address.Port)
	assert.Equal(t, egv1a1.GatewayAdminHost, eg.Admin.Address.Host)
	assert.False(t, eg.Admin.EnableDumpConfig)
	assert.False(t, eg.Admin.EnablePprof)
}

func TestEnvoyGatewayTelemetry(t *testing.T) {
	// default envoygateway config telemetry should not be nil
	eg := egv1a1.DefaultEnvoyGateway()
	assert.NotNil(t, eg.Telemetry)

	// get default telemetry config from envoygateway
	// values should be set in default
	egTelemetry := eg.GetEnvoyGatewayTelemetry()
	assert.NotNil(t, egTelemetry)
	assert.NotNil(t, egTelemetry.Metrics)
	assert.False(t, egTelemetry.Metrics.Prometheus.Disable)
	assert.Nil(t, egTelemetry.Metrics.Sinks)

	// override the telemetry config
	// values should be updated
	eg.Telemetry.Metrics = &egv1a1.EnvoyGatewayMetrics{
		Prometheus: &egv1a1.EnvoyGatewayPrometheusProvider{
			Disable: true,
		},
		Sinks: []egv1a1.EnvoyGatewayMetricSink{
			{
				Type: egv1a1.MetricSinkTypeOpenTelemetry,
				OpenTelemetry: &egv1a1.EnvoyGatewayOpenTelemetrySink{
					Host:     "otel-collector.monitoring.svc.cluster.local",
					Protocol: "grpc",
					Port:     4317,
				},
			}, {
				Type: egv1a1.MetricSinkTypeOpenTelemetry,
				OpenTelemetry: &egv1a1.EnvoyGatewayOpenTelemetrySink{
					Host:     "otel-collector.monitoring.svc.cluster.local",
					Protocol: "http",
					Port:     4318,
				},
			},
		},
	}

	assert.True(t, eg.GetEnvoyGatewayTelemetry().Metrics.Prometheus.Disable)
	assert.Len(t, eg.GetEnvoyGatewayTelemetry().Metrics.Sinks, 2)
	assert.Equal(t, egv1a1.MetricSinkTypeOpenTelemetry, eg.GetEnvoyGatewayTelemetry().Metrics.Sinks[0].Type)

	// set eg defaults when telemetry is nil
	// the telemetry should not be nil
	eg.Telemetry = nil
	eg.SetEnvoyGatewayDefaults()
	assert.NotNil(t, eg.Telemetry)
	assert.NotNil(t, eg.Telemetry.Metrics)
	assert.False(t, eg.Telemetry.Metrics.Prometheus.Disable)
	assert.Nil(t, eg.Telemetry.Metrics.Sinks)
}
