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
