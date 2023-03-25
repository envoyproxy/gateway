// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation_test

import (
	// Register embed
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/api/config/v1alpha1/validation"
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
		obj      *egcfgv1a1.EnvoyProxy
		expected bool
	}{
		{
			name:     "nil envoyproxy",
			obj:      nil,
			expected: false,
		},
		{
			name: "nil provider",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: nil,
				},
			},
			expected: true,
		},
		{
			name: "unsupported provider",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.ResourceProvider{
						Type: egcfgv1a1.ProviderTypeFile,
					},
				},
			},
			expected: false,
		},
		{
			name: "nil envoy service",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.ResourceProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.KubernetesResourceProvider{
							EnvoyService: nil,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "unsupported envoy service type \"\" ",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.ResourceProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.KubernetesResourceProvider{
							EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
								Type: egcfgv1a1.GetKubernetesServiceType(""),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "unsupported envoy service type 'NodePort'",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.ResourceProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.KubernetesResourceProvider{
							EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
								Type: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceType(corev1.ServiceTypeNodePort)),
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "valid envoy service type 'LoadBalancer'",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.ResourceProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.KubernetesResourceProvider{
							EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
								Type: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeLoadBalancer),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid envoy service type 'ClusterIP'",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.ResourceProvider{
						Type: egcfgv1a1.ProviderTypeKubernetes,
						Kubernetes: &egcfgv1a1.KubernetesResourceProvider{
							EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
								Type: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeClusterIP),
							},
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "valid user bootstrap",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &validUserBootstrap,
				},
			},
			expected: true,
		},
		{
			name: "user bootstrap with missing admin address",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &missingAdminAddressUserBootstrap,
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different dynamic resources",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &differentDynamicResourcesUserBootstrap,
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different xds_cluster endpoint",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: &differentXdsClusterAddressBootstrap,
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := validation.ValidateEnvoyProxy(tc.obj)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
