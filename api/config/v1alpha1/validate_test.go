// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	// Register embed
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//go:embed testdata/valid-user-bootstrap.yaml
	validUserBootstrap string
	//go:embed testdata/missing-admin-address-user-bootstrap.yaml
	missingAdminAddressUserBootstrap string
	//go:embed testdata/different-dynamic-resources-user-bootstrap.yaml
	differentDynamicResourcesUserBootstrap string
	//go:embed testdata/different-xds-cluster-address-bootstrap.yaml
	differentXdsClusterAddressBootstrap string
)

func TestValidateEnvoyProxy(t *testing.T) {
	testCases := []struct {
		name     string
		obj      *EnvoyProxy
		expected bool
	}{
		{
			name:     "nil envoyproxy",
			obj:      nil,
			expected: false,
		},
		{
			name: "nil provider",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: nil,
				},
			},
			expected: true,
		},
		{
			name: "unsupported provider",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeFile,
					},
				},
			},
			expected: false,
		},
		{
			name: "nil envoy service",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: nil,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "unsupported envoy service type \"\" ",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: &KubernetesServiceSpec{
								Type: GetKubernetesServiceType(""),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "unsupported envoy service type 'NodePort'",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: &KubernetesServiceSpec{
								Type: GetKubernetesServiceType(ServiceType(corev1.ServiceTypeNodePort)),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "valid envoy service type 'LoadBalancer'",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: &KubernetesServiceSpec{
								Type: GetKubernetesServiceType(ServiceTypeLoadBalancer),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid envoy service type 'ClusterIP'",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Provider: &EnvoyProxyProvider{
						Type: ProviderTypeKubernetes,
						Kubernetes: &EnvoyProxyKubernetesProvider{
							EnvoyService: &KubernetesServiceSpec{
								Type: GetKubernetesServiceType(ServiceTypeClusterIP),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid user bootstrap",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Bootstrap: &validUserBootstrap,
				},
			},
			expected: true,
		},
		{
			name: "user bootstrap with missing admin address",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Bootstrap: &missingAdminAddressUserBootstrap,
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different dynamic resources",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Bootstrap: &differentDynamicResourcesUserBootstrap,
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different xds_cluster endpoint",
			obj: &EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: EnvoyProxySpec{
					Bootstrap: &differentXdsClusterAddressBootstrap,
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := tc.obj.Validate()
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
