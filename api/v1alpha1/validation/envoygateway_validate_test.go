// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

var (
	TLSSecretKind       = v1.Kind("Secret")
	TLSUnrecognizedKind = v1.Kind("Unrecognized")
)

func TestValidateEnvoyGateway(t *testing.T) {
	eg := v1alpha1.DefaultEnvoyGateway()

	testCases := []struct {
		name   string
		eg     *v1alpha1.EnvoyGateway
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
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: false,
		},
		{
			name: "unspecified provider",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway: v1alpha1.DefaultGateway(),
				},
			},
			expect: false,
		},
		{
			name: "empty gateway controllerName",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  &v1alpha1.Gateway{ControllerName: ""},
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
				},
			},
			expect: false,
		},
		{
			name: "unsupported provider",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: &v1alpha1.EnvoyGatewayProvider{Type: v1alpha1.ProviderTypeFile},
				},
			},
			expect: false,
		},
		{
			name: "empty ratelimit",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:   v1alpha1.DefaultGateway(),
					Provider:  v1alpha1.DefaultEnvoyGatewayProvider(),
					RateLimit: &v1alpha1.RateLimit{},
				},
			},
			expect: false,
		},
		{
			name: "empty ratelimit redis setting",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					RateLimit: &v1alpha1.RateLimit{
						Backend: v1alpha1.RateLimitDatabaseBackend{
							Type:  v1alpha1.RedisBackendType,
							Redis: &v1alpha1.RateLimitRedisSettings{},
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "unknown ratelimit redis url format",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					RateLimit: &v1alpha1.RateLimit{
						Backend: v1alpha1.RateLimitDatabaseBackend{
							Type: v1alpha1.RedisBackendType,
							Redis: &v1alpha1.RateLimitRedisSettings{
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
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					RateLimit: &v1alpha1.RateLimit{
						Backend: v1alpha1.RateLimitDatabaseBackend{
							Type: v1alpha1.RedisBackendType,
							Redis: &v1alpha1.RateLimitRedisSettings{
								URL: "localhost:6376",
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "happy extension settings",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
						Service: &v1alpha1.ExtensionService{
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
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
						Service: &v1alpha1.ExtensionService{
							Host: "foo.extension",
							Port: 443,
							TLS: &v1alpha1.ExtensionTLS{
								CertificateRef: v1.SecretObjectReference{
									Kind: &TLSSecretKind,
									Name: v1.ObjectName("certificate"),
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
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
						Service: &v1alpha1.ExtensionService{
							Host: "foo.extension",
							Port: 443,
							TLS: &v1alpha1.ExtensionTLS{
								CertificateRef: v1.SecretObjectReference{
									Kind: &TLSSecretKind,
									Name: v1.ObjectName("certificate"),
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
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
								},
							},
						},
						Service: &v1alpha1.ExtensionService{
							Host: "foo.extension",
							Port: 8080,
							TLS: &v1alpha1.ExtensionTLS{
								CertificateRef: v1.SecretObjectReference{
									Kind: &TLSUnrecognizedKind,
									Name: v1.ObjectName("certificate"),
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
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Hooks: &v1alpha1.ExtensionHooks{
							XDSTranslator: &v1alpha1.XDSTranslatorHooks{
								Pre: []v1alpha1.XDSTranslatorHook{},
								Post: []v1alpha1.XDSTranslatorHook{
									v1alpha1.XDSHTTPListener,
									v1alpha1.XDSTranslation,
									v1alpha1.XDSRoute,
									v1alpha1.XDSVirtualHost,
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
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					ExtensionManager: &v1alpha1.ExtensionManager{
						Resources: []v1alpha1.GroupVersionKind{
							{
								Group:   "foo.example.io",
								Version: "v1alpha1",
								Kind:    "Foo",
							},
						},
						Service: &v1alpha1.ExtensionService{
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
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelInfo,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level warn",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelWarn,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level error",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelError,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "valid gateway logging level debug",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: v1alpha1.LogLevelDebug,
							v1alpha1.LogComponentProviderRunner: v1alpha1.LogLevelDebug,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "invalid gateway logging level",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Logging: &v1alpha1.EnvoyGatewayLogging{
						Level: map[v1alpha1.EnvoyGatewayLogComponent]v1alpha1.LogLevel{
							v1alpha1.LogComponentGatewayDefault: "inffo",
						},
					},
				},
			},
			expect: false,
		}, {
			name: "valid gateway metrics sink",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Telemetry: &v1alpha1.EnvoyGatewayTelemetry{
						Metrics: &v1alpha1.EnvoyGatewayMetrics{
							Sinks: []v1alpha1.EnvoyGatewayMetricSink{
								{
									Type: v1alpha1.MetricSinkTypeOpenTelemetry,
									OpenTelemetry: &v1alpha1.EnvoyGatewayOpenTelemetrySink{
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
		}, {
			name: "invalid gateway metrics sink",
			eg: &v1alpha1.EnvoyGateway{
				EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
					Gateway:  v1alpha1.DefaultGateway(),
					Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					Telemetry: &v1alpha1.EnvoyGatewayTelemetry{
						Metrics: &v1alpha1.EnvoyGatewayMetrics{
							Sinks: []v1alpha1.EnvoyGatewayMetricSink{
								{
									Type: v1alpha1.MetricSinkTypeOpenTelemetry,
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
		tc := tc
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
