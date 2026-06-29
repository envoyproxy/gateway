// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRunningOnKubernetes(t *testing.T) {
	type args struct {
		provider EnvoyGatewayProvider
	}

	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "Kubernetes provider",
			args: args{
				provider: EnvoyGatewayProvider{
					Type: ProviderTypeKubernetes,
				},
			},
			expected: true,
		},
		{
			name: "Custom provider with no configuration",
			args: args{
				provider: EnvoyGatewayProvider{
					Type:   ProviderTypeCustom,
					Custom: new(EnvoyGatewayCustomProvider{}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with no configuration",
			args: args{
				provider: EnvoyGatewayProvider{
					Type:   ProviderTypeCustom,
					Custom: new(EnvoyGatewayCustomProvider{}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with file configuration",
			args: args{
				provider: EnvoyGatewayProvider{
					Type: ProviderTypeCustom,
					Custom: new(EnvoyGatewayCustomProvider{
						Resource: EnvoyGatewayResourceProvider{
							Type: ResourceProviderTypeFile,
						},
					}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with kubernetes configuration",
			args: args{
				provider: EnvoyGatewayProvider{
					Type: ProviderTypeCustom,
					Custom: new(EnvoyGatewayCustomProvider{
						Resource: EnvoyGatewayResourceProvider{
							Type: ResourceProviderTypeKubernetes,
						},
					}),
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.args.provider.IsRunningOnKubernetes())
		})
	}
}

func TestIsInfraManagedRemotely(t *testing.T) {
	type args struct {
		provider EnvoyGatewayProvider
	}

	tests := []struct {
		name     string
		args     args
		expected bool
	}{
		{
			name: "Kubernetes provider",
			args: args{
				provider: EnvoyGatewayProvider{
					Type: ProviderTypeKubernetes,
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with no configuration",
			args: args{
				provider: EnvoyGatewayProvider{
					Type:   ProviderTypeCustom,
					Custom: new(EnvoyGatewayCustomProvider{}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with no configuration",
			args: args{
				provider: EnvoyGatewayProvider{
					Type:   ProviderTypeCustom,
					Custom: new(EnvoyGatewayCustomProvider{}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with file configuration",
			args: args{
				provider: EnvoyGatewayProvider{
					Type: ProviderTypeCustom,
					Custom: new(EnvoyGatewayCustomProvider{
						Resource: EnvoyGatewayResourceProvider{
							Type: ResourceProviderTypeFile,
						},
					}),
				},
			},
			expected: false,
		},
		{
			name: "Custom provider with remote configuration",
			args: args{
				provider: EnvoyGatewayProvider{
					Type: ProviderTypeCustom,
					Custom: new(EnvoyGatewayCustomProvider{
						Infrastructure: new(EnvoyGatewayInfrastructureProvider{
							Type: InfrastructureProviderTypeRemote,
						}),
					}),
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.args.provider.IsInfraManagedRemotely())
		})
	}
}

func TestWatchesNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		eg       *EnvoyGateway
		expected bool
	}{
		{
			name: "nil provider",
			eg: &EnvoyGateway{
				EnvoyGatewaySpec: EnvoyGatewaySpec{
					Provider: nil,
				},
			},
			expected: false,
		},
		{
			name: "non-kubernetes provider",
			eg: &EnvoyGateway{
				EnvoyGatewaySpec: EnvoyGatewaySpec{
					Provider: &EnvoyGatewayProvider{
						Type: ProviderTypeCustom,
						Custom: &EnvoyGatewayCustomProvider{
							Resource: EnvoyGatewayResourceProvider{
								Type: ResourceProviderTypeFile,
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "kubernetes provider with nil watch",
			eg: &EnvoyGateway{
				EnvoyGatewaySpec: EnvoyGatewaySpec{
					Provider: &EnvoyGatewayProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyGatewayKubernetesProvider{
							EnvoyGatewayKubernetesConfiguration: EnvoyGatewayKubernetesConfiguration{
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
			eg: &EnvoyGateway{
				EnvoyGatewaySpec: EnvoyGatewaySpec{
					Provider: &EnvoyGatewayProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyGatewayKubernetesProvider{
							EnvoyGatewayKubernetesConfiguration: EnvoyGatewayKubernetesConfiguration{
								Watch: &KubernetesWatchMode{
									Type: KubernetesWatchModeTypeNamespaceSelector,
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
			eg: &EnvoyGateway{
				EnvoyGatewaySpec: EnvoyGatewaySpec{
					Provider: &EnvoyGatewayProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyGatewayKubernetesProvider{
							EnvoyGatewayKubernetesConfiguration: EnvoyGatewayKubernetesConfiguration{
								Watch: &KubernetesWatchMode{
									Type:       KubernetesWatchModeTypeNamespaces,
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
			eg: &EnvoyGateway{
				EnvoyGatewaySpec: EnvoyGatewaySpec{
					Provider: &EnvoyGatewayProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyGatewayKubernetesProvider{
							EnvoyGatewayKubernetesConfiguration: EnvoyGatewayKubernetesConfiguration{
								Watch: &KubernetesWatchMode{
									Type:       KubernetesWatchModeTypeNamespaces,
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
			eg: &EnvoyGateway{
				EnvoyGatewaySpec: EnvoyGatewaySpec{
					Provider: &EnvoyGatewayProvider{
						Type: ProviderTypeCustom,
						Custom: &EnvoyGatewayCustomProvider{
							Resource: EnvoyGatewayResourceProvider{
								Type: ResourceProviderTypeKubernetes,
								Kubernetes: &EnvoyGatewayKubernetesCustomProvider{
									EnvoyGatewayKubernetesConfiguration: EnvoyGatewayKubernetesConfiguration{
										Watch: &KubernetesWatchMode{
											Type:       KubernetesWatchModeTypeNamespaces,
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
			assert.Equal(t, tt.expected, tt.eg.WatchesNamespaces())
		})
	}
}
