// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package apihelpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestIsRunningOnKubernetes(t *testing.T) {
	type args struct {
		provider egv1a1.EnvoyGatewayProvider
	}

	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "Kubernetes provider",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
				},
			},
			expected: true,
		},
		{
			name: "Custom provider with no configuration",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type:   egv1a1.ProviderTypeCustom,
					Custom: new(egv1a1.EnvoyGatewayCustomProvider{}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with no configuration",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type:   egv1a1.ProviderTypeCustom,
					Custom: new(egv1a1.EnvoyGatewayCustomProvider{}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with file configuration",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeCustom,
					Custom: new(egv1a1.EnvoyGatewayCustomProvider{
						Resource: egv1a1.EnvoyGatewayResourceProvider{
							Type: egv1a1.ResourceProviderTypeFile,
						},
					}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with kubernetes configuration",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeCustom,
					Custom: new(egv1a1.EnvoyGatewayCustomProvider{
						Resource: egv1a1.EnvoyGatewayResourceProvider{
							Type: egv1a1.ResourceProviderTypeKubernetes,
						},
					}),
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRunningOnKubernetes(&tt.args.provider))
		})
	}
}

func TestIsInfraManagedRemotely(t *testing.T) {
	type args struct {
		provider egv1a1.EnvoyGatewayProvider
	}

	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "Kubernetes provider",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with no configuration",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type:   egv1a1.ProviderTypeCustom,
					Custom: new(egv1a1.EnvoyGatewayCustomProvider{}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with no configuration",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type:   egv1a1.ProviderTypeCustom,
					Custom: new(egv1a1.EnvoyGatewayCustomProvider{}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with file configuration",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeCustom,
					Custom: new(egv1a1.EnvoyGatewayCustomProvider{
						Resource: egv1a1.EnvoyGatewayResourceProvider{
							Type: egv1a1.ResourceProviderTypeFile,
						},
					}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with remote configuration",
			args: args{
				provider: egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeCustom,
					Custom: new(egv1a1.EnvoyGatewayCustomProvider{
						Infrastructure: new(egv1a1.EnvoyGatewayInfrastructureProvider{
							Type: egv1a1.InfrastructureProviderTypeRemote,
						}),
					}),
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsInfraManagedRemotely(&tt.args.provider))
		})
	}
}

func TestWatchesNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		eg       *egv1a1.EnvoyGateway
		expected bool
	}{
		{
			name: "nil provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: nil,
				},
			},
			expected: false,
		},
		{
			name: "non-kubernetes provider",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
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
			expected: false,
		},
		{
			name: "kubernetes provider with nil watch",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
								Watch: nil,
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "kubernetes provider with namespace selector watch mode",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
								Watch: &egv1a1.KubernetesWatchMode{
									Type: egv1a1.KubernetesWatchModeTypeNamespaceSelector,
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "kubernetes provider with namespaces watch mode but empty namespaces",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
								Watch: &egv1a1.KubernetesWatchMode{
									Type:       egv1a1.KubernetesWatchModeTypeNamespaces,
									Namespaces: []string{},
								},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "kubernetes provider with namespaces watch mode and namespaces set",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
							EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
								Watch: &egv1a1.KubernetesWatchMode{
									Type:       egv1a1.KubernetesWatchModeTypeNamespaces,
									Namespaces: []string{"ns-a", "ns-b"},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "custom kubernetes resource provider with namespaces watch mode and namespaces set",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Provider: &egv1a1.EnvoyGatewayProvider{
						Type: egv1a1.ProviderTypeCustom,
						Custom: &egv1a1.EnvoyGatewayCustomProvider{
							Resource: egv1a1.EnvoyGatewayResourceProvider{
								Type: egv1a1.ResourceProviderTypeKubernetes,
								Kubernetes: &egv1a1.EnvoyGatewayKubernetesCustomProvider{
									EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
										Watch: &egv1a1.KubernetesWatchMode{
											Type:       egv1a1.KubernetesWatchModeTypeNamespaces,
											Namespaces: []string{"ns-a"},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, WatchesNamespaces(tt.eg))
		})
	}
}

func TestGetKubernetesConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		provider egv1a1.EnvoyGatewayProvider
		expected egv1a1.EnvoyGatewayKubernetesConfiguration
	}{
		{
			name: "Kubernetes provider with nil sub-provider does not panic",
			provider: egv1a1.EnvoyGatewayProvider{
				Type:       egv1a1.ProviderTypeKubernetes,
				Kubernetes: nil,
			},
			expected: egv1a1.EnvoyGatewayKubernetesConfiguration{},
		},
		{
			name: "Kubernetes provider with populated sub-provider returns config",
			provider: egv1a1.EnvoyGatewayProvider{
				Type: egv1a1.ProviderTypeKubernetes,
				Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
					EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
						Watch: &egv1a1.KubernetesWatchMode{Type: egv1a1.KubernetesWatchModeTypeNamespaces},
					},
				},
			},
			expected: egv1a1.EnvoyGatewayKubernetesConfiguration{
				Watch: &egv1a1.KubernetesWatchMode{Type: egv1a1.KubernetesWatchModeTypeNamespaces},
			},
		},
		{
			name: "Custom provider with kubernetes resource type but nil sub-provider does not panic",
			provider: egv1a1.EnvoyGatewayProvider{
				Type: egv1a1.ProviderTypeCustom,
				Custom: &egv1a1.EnvoyGatewayCustomProvider{
					Resource: egv1a1.EnvoyGatewayResourceProvider{
						Type:       egv1a1.ResourceProviderTypeKubernetes,
						Kubernetes: nil,
					},
				},
			},
			expected: egv1a1.EnvoyGatewayKubernetesConfiguration{},
		},
		{
			name: "Custom provider with populated kubernetes resource returns config",
			provider: egv1a1.EnvoyGatewayProvider{
				Type: egv1a1.ProviderTypeCustom,
				Custom: &egv1a1.EnvoyGatewayCustomProvider{
					Resource: egv1a1.EnvoyGatewayResourceProvider{
						Type: egv1a1.ResourceProviderTypeKubernetes,
						Kubernetes: &egv1a1.EnvoyGatewayKubernetesCustomProvider{
							EnvoyGatewayKubernetesConfiguration: egv1a1.EnvoyGatewayKubernetesConfiguration{
								Watch: &egv1a1.KubernetesWatchMode{Type: egv1a1.KubernetesWatchModeTypeNamespaceSelector},
							},
						},
					},
				},
			},
			expected: egv1a1.EnvoyGatewayKubernetesConfiguration{
				Watch: &egv1a1.KubernetesWatchMode{Type: egv1a1.KubernetesWatchModeTypeNamespaceSelector},
			},
		},
		{
			name: "Custom provider with file resource type returns zero value",
			provider: egv1a1.EnvoyGatewayProvider{
				Type: egv1a1.ProviderTypeCustom,
				Custom: &egv1a1.EnvoyGatewayCustomProvider{
					Resource: egv1a1.EnvoyGatewayResourceProvider{
						Type: egv1a1.ResourceProviderTypeFile,
					},
				},
			},
			expected: egv1a1.EnvoyGatewayKubernetesConfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got egv1a1.EnvoyGatewayKubernetesConfiguration
			assert.NotPanics(t, func() {
				got = GetKubernetesConfiguration(&tt.provider)
			})
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetKubernetesInfrastructureConfiguration(t *testing.T) {
	deploy := &egv1a1.KubernetesDeployMode{}
	tests := []struct {
		name     string
		provider egv1a1.EnvoyGatewayProvider
		expected egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration
	}{
		{
			name: "Kubernetes provider with nil sub-provider does not panic",
			provider: egv1a1.EnvoyGatewayProvider{
				Type:       egv1a1.ProviderTypeKubernetes,
				Kubernetes: nil,
			},
			expected: egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration{},
		},
		{
			name: "Kubernetes provider with populated sub-provider returns config",
			provider: egv1a1.EnvoyGatewayProvider{
				Type: egv1a1.ProviderTypeKubernetes,
				Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
					EnvoyGatewayKubernetesInfrastructureConfiguration: egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration{
						Deploy: deploy,
					},
				},
			},
			expected: egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration{Deploy: deploy},
		},
		{
			name: "Custom provider returns zero value",
			provider: egv1a1.EnvoyGatewayProvider{
				Type: egv1a1.ProviderTypeCustom,
				Custom: &egv1a1.EnvoyGatewayCustomProvider{
					Resource: egv1a1.EnvoyGatewayResourceProvider{
						Type: egv1a1.ResourceProviderTypeKubernetes,
					},
				},
			},
			expected: egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got egv1a1.EnvoyGatewayKubernetesInfrastructureConfiguration
			assert.NotPanics(t, func() {
				got = GetKubernetesInfrastructureConfiguration(&tt.provider)
			})
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDebounceDefaultsToDisabled(t *testing.T) {
	eg := &egv1a1.EnvoyGateway{}
	SetEnvoyGatewayDefaults(eg)

	// Debouncing is opt in, so defaulting must not define the config, and defining
	// it is what turns debouncing on.
	require.Nil(t, eg.Debounce)
	require.False(t, IsDebounceEnabled(eg.Debounce))
	require.True(t, IsDebounceEnabled(&egv1a1.Debounce{}))

	// A config that sets no durations falls back to the advertised defaults.
	after, maxHold, err := ResolveDebounceDurations(&egv1a1.Debounce{})
	require.NoError(t, err)
	require.Equal(t, DefaultDebounceAfter, after)
	require.Equal(t, DefaultDebounceMax, maxHold)
}

func TestEnvoyProxyPatchRuntimeFlag(t *testing.T) {
	tests := []struct {
		name     string
		eg       egv1a1.EnvoyGateway
		expected bool
	}{
		{
			name:     "enabled by default",
			eg:       egv1a1.EnvoyGateway{},
			expected: true,
		},
		{
			name: "explicitly disabled",
			eg: egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					RuntimeFlags: &egv1a1.RuntimeFlags{Disabled: []egv1a1.RuntimeFlag{egv1a1.EnvoyProxyPatch}},
				},
			},
			expected: false,
		},
		{
			name: "disabled takes precedence",
			eg: egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					RuntimeFlags: &egv1a1.RuntimeFlags{
						Enabled:  []egv1a1.RuntimeFlag{egv1a1.EnvoyProxyPatch},
						Disabled: []egv1a1.RuntimeFlag{egv1a1.EnvoyProxyPatch},
					},
				},
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRuntimeFlagEnabled(tt.eg.RuntimeFlags, egv1a1.EnvoyProxyPatch))
		})
	}
}

func TestOIDCTokenCookieChunkingRuntimeFlag(t *testing.T) {
	tests := []struct {
		name     string
		eg       egv1a1.EnvoyGateway
		expected bool
	}{
		{
			name:     "enabled by default",
			eg:       egv1a1.EnvoyGateway{},
			expected: true,
		},
		{
			name: "explicitly disabled",
			eg: egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					RuntimeFlags: &egv1a1.RuntimeFlags{Disabled: []egv1a1.RuntimeFlag{egv1a1.OIDCTokenCookieChunking}},
				},
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsRuntimeFlagEnabled(tt.eg.RuntimeFlags, egv1a1.OIDCTokenCookieChunking))
		})
	}
}
