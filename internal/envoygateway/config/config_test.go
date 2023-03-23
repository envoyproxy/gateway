// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
)

var (
	TLSSecretKind       = v1beta1.Kind("Secret")
	TLSUnrecognizedKind = v1beta1.Kind("Unrecognized")
)

func TestValidate(t *testing.T) {
	cfg, err := New()
	require.NoError(t, err)

	testCases := []struct {
		name   string
		cfg    *Server
		expect bool
	}{
		{
			name:   "default",
			cfg:    cfg,
			expect: true,
		},
		{
			name: "empty namespace",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
					},
				},
				Namespace: "",
			},
			expect: false,
		},
		{
			name: "unspecified envoy gateway",
			cfg: &Server{
				Namespace: "test-ns",
				Logger:    logr.Logger{},
			},
			expect: false,
		},
		{
			name: "unspecified gateway",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Provider: v1alpha1.DefaultProvider(),
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "unspecified provider",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway: v1alpha1.DefaultGateway(),
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "empty gateway controllerName",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  &v1alpha1.Gateway{ControllerName: ""},
						Provider: v1alpha1.DefaultProvider(),
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "unsupported provider",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: &v1alpha1.Provider{Type: v1alpha1.ProviderTypeFile},
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "empty ratelimit",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:   v1alpha1.DefaultGateway(),
						Provider:  v1alpha1.DefaultProvider(),
						RateLimit: &v1alpha1.RateLimit{},
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "empty ratelimit redis setting",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						RateLimit: &v1alpha1.RateLimit{
							Backend: v1alpha1.RateLimitDatabaseBackend{
								Type:  v1alpha1.RedisBackendType,
								Redis: &v1alpha1.RateLimitRedisSettings{},
							},
						},
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "unknown ratelimit redis url format",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
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
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "happy ratelimit redis settings",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
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
				Namespace: "test-ns",
			},
			expect: true,
		},
		{
			name: "happy extension settings",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						Extension: &v1alpha1.Extension{
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
				Namespace: "test-ns",
			},
			expect: true,
		},
		{
			name: "happy extension settings tls",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						Extension: &v1alpha1.Extension{
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
									CertificateRef: v1beta1.SecretObjectReference{
										Kind: &TLSSecretKind,
										Name: v1beta1.ObjectName("certificate"),
									},
								},
							},
						},
					},
				},
				Namespace: "test-ns",
			},
			expect: true,
		},
		{
			name: "happy extension settings no resources",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						Extension: &v1alpha1.Extension{
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
									CertificateRef: v1beta1.SecretObjectReference{
										Kind: &TLSSecretKind,
										Name: v1beta1.ObjectName("certificate"),
									},
								},
							},
						},
					},
				},
				Namespace: "test-ns",
			},
			expect: true,
		},
		{
			name: "unknown TLS certificateRef in extension settings",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						Extension: &v1alpha1.Extension{
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
									CertificateRef: v1beta1.SecretObjectReference{
										Kind: &TLSUnrecognizedKind,
										Name: v1beta1.ObjectName("certificate"),
									},
								},
							},
						},
					},
				},
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "empty service in extension settings",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						Extension: &v1alpha1.Extension{
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
				Namespace: "test-ns",
			},
			expect: false,
		},
		{
			name: "empty hooks in extension settings",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultProvider(),
						Extension: &v1alpha1.Extension{
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
				Namespace: "test-ns",
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
