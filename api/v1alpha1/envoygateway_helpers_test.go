// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
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

func TestGetKubernetesConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		provider EnvoyGatewayProvider
		expected EnvoyGatewayKubernetesConfiguration
	}{
		{
			name: "Kubernetes provider with nil sub-provider does not panic",
			provider: EnvoyGatewayProvider{
				Type:       ProviderTypeKubernetes,
				Kubernetes: nil,
			},
			expected: EnvoyGatewayKubernetesConfiguration{},
		},
		{
			name: "Kubernetes provider with populated sub-provider returns config",
			provider: EnvoyGatewayProvider{
				Type: ProviderTypeKubernetes,
				Kubernetes: &EnvoyGatewayKubernetesProvider{
					EnvoyGatewayKubernetesConfiguration: EnvoyGatewayKubernetesConfiguration{
						Watch: &KubernetesWatchMode{Type: KubernetesWatchModeTypeNamespaces},
					},
				},
			},
			expected: EnvoyGatewayKubernetesConfiguration{
				Watch: &KubernetesWatchMode{Type: KubernetesWatchModeTypeNamespaces},
			},
		},
		{
			name: "Custom provider with kubernetes resource type but nil sub-provider does not panic",
			provider: EnvoyGatewayProvider{
				Type: ProviderTypeCustom,
				Custom: &EnvoyGatewayCustomProvider{
					Resource: EnvoyGatewayResourceProvider{
						Type:       ResourceProviderTypeKubernetes,
						Kubernetes: nil,
					},
				},
			},
			expected: EnvoyGatewayKubernetesConfiguration{},
		},
		{
			name: "Custom provider with populated kubernetes resource returns config",
			provider: EnvoyGatewayProvider{
				Type: ProviderTypeCustom,
				Custom: &EnvoyGatewayCustomProvider{
					Resource: EnvoyGatewayResourceProvider{
						Type: ResourceProviderTypeKubernetes,
						Kubernetes: &EnvoyGatewayKubernetesCustomProvider{
							EnvoyGatewayKubernetesConfiguration: EnvoyGatewayKubernetesConfiguration{
								Watch: &KubernetesWatchMode{Type: KubernetesWatchModeTypeNamespaceSelector},
							},
						},
					},
				},
			},
			expected: EnvoyGatewayKubernetesConfiguration{
				Watch: &KubernetesWatchMode{Type: KubernetesWatchModeTypeNamespaceSelector},
			},
		},
		{
			name: "Custom provider with file resource type returns zero value",
			provider: EnvoyGatewayProvider{
				Type: ProviderTypeCustom,
				Custom: &EnvoyGatewayCustomProvider{
					Resource: EnvoyGatewayResourceProvider{
						Type: ResourceProviderTypeFile,
					},
				},
			},
			expected: EnvoyGatewayKubernetesConfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got EnvoyGatewayKubernetesConfiguration
			assert.NotPanics(t, func() {
				got = tt.provider.GetKubernetesConfiguration()
			})
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestGetKubernetesInfrastructureConfiguration(t *testing.T) {
	deploy := &KubernetesDeployMode{}
	tests := []struct {
		name     string
		provider EnvoyGatewayProvider
		expected EnvoyGatewayKubernetesInfrastructureConfiguration
	}{
		{
			name: "Kubernetes provider with nil sub-provider does not panic",
			provider: EnvoyGatewayProvider{
				Type:       ProviderTypeKubernetes,
				Kubernetes: nil,
			},
			expected: EnvoyGatewayKubernetesInfrastructureConfiguration{},
		},
		{
			name: "Kubernetes provider with populated sub-provider returns config",
			provider: EnvoyGatewayProvider{
				Type: ProviderTypeKubernetes,
				Kubernetes: &EnvoyGatewayKubernetesProvider{
					EnvoyGatewayKubernetesInfrastructureConfiguration: EnvoyGatewayKubernetesInfrastructureConfiguration{
						Deploy: deploy,
					},
				},
			},
			expected: EnvoyGatewayKubernetesInfrastructureConfiguration{Deploy: deploy},
		},
		{
			name: "Custom provider returns zero value",
			provider: EnvoyGatewayProvider{
				Type: ProviderTypeCustom,
				Custom: &EnvoyGatewayCustomProvider{
					Resource: EnvoyGatewayResourceProvider{
						Type: ResourceProviderTypeKubernetes,
					},
				},
			},
			expected: EnvoyGatewayKubernetesInfrastructureConfiguration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got EnvoyGatewayKubernetesInfrastructureConfiguration
			assert.NotPanics(t, func() {
				got = tt.provider.GetKubernetesInfrastructureConfiguration()
			})
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestDebounceDefaultsToDisabled(t *testing.T) {
	eg := &EnvoyGateway{}
	eg.SetEnvoyGatewayDefaults()

	require.NotNil(t, eg.Debounce)
	require.False(t, ptr.Deref(eg.Debounce.Enable, true))
	require.Equal(t, gwapiv1.Duration("100ms"), *eg.Debounce.After)
	require.Equal(t, gwapiv1.Duration("10s"), *eg.Debounce.Max)

	enabled, after, maxDelay, err := eg.DebounceSettings()
	require.NoError(t, err)
	require.False(t, enabled, "debouncing must be opt-in")
	require.Equal(t, DefaultDebounceAfter, after)
	require.Equal(t, DefaultDebounceMax, maxDelay)
}

func TestDebounceSettings(t *testing.T) {
	duration := func(s string) *gwapiv1.Duration {
		d := gwapiv1.Duration(s)
		return &d
	}

	tests := []struct {
		name        string
		debounce    *Debounce
		wantEnabled bool
		wantAfter   time.Duration
		wantMax     time.Duration
		wantErr     bool
	}{
		{
			name:      "nil debounce",
			wantAfter: DefaultDebounceAfter,
			wantMax:   DefaultDebounceMax,
		},
		{
			name:      "explicitly disabled ignores durations",
			debounce:  &Debounce{Enable: new(false), After: duration("1s"), Max: duration("2s")},
			wantAfter: DefaultDebounceAfter,
			wantMax:   DefaultDebounceMax,
		},
		{
			name:        "enabled with defaults",
			debounce:    &Debounce{Enable: new(true)},
			wantEnabled: true,
			wantAfter:   DefaultDebounceAfter,
			wantMax:     DefaultDebounceMax,
		},
		{
			name:        "enabled with overrides",
			debounce:    &Debounce{Enable: new(true), After: duration("250ms"), Max: duration("30s")},
			wantEnabled: true,
			wantAfter:   250 * time.Millisecond,
			wantMax:     30 * time.Second,
		},
		{
			name:     "enabled with malformed after",
			debounce: &Debounce{Enable: new(true), After: duration("nope")},
			wantErr:  true,
		},
		{
			name:     "enabled with non positive max",
			debounce: &Debounce{Enable: new(true), Max: duration("0s")},
			wantErr:  true,
		},
		{
			name:     "enabled with max shorter than after",
			debounce: &Debounce{Enable: new(true), After: duration("5s"), Max: duration("1s")},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eg := &EnvoyGateway{EnvoyGatewaySpec: EnvoyGatewaySpec{Debounce: tt.debounce}}

			enabled, after, maxDelay, err := eg.DebounceSettings()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantEnabled, enabled)
			require.Equal(t, tt.wantAfter, after)
			require.Equal(t, tt.wantMax, maxDelay)
		})
	}
}
