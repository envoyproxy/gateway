// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestBuildLoadBalancer_DynamicModule(t *testing.T) {
	envoyProxy := &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			DynamicModules: []egv1a1.DynamicModuleEntry{
				{
					Name: "my-module",
					Source: egv1a1.DynamicModuleSource{
						Type: ptr.To(egv1a1.LocalDynamicModuleSourceType),
						Local: &egv1a1.LocalDynamicModuleSource{
							Path: "/usr/local/lib/my-module.so",
						},
					},
				},
			},
		},
	}

	policy := &egv1a1.ClusterSettings{
		LoadBalancer: &egv1a1.LoadBalancer{
			Type: egv1a1.DynamicModuleLoadBalancerType,
			DynamicModule: &egv1a1.DynamicModuleLBPolicy{
				DynamicModuleRef: egv1a1.DynamicModuleRef{
					Name:               "my-module",
					ImplementationName: ptr.To("round-robin-v2"),
					Config:             &apiextensionsv1.JSON{Raw: []byte(`{"key":"value"}`)},
				},
			},
		},
	}

	lb, err := buildLoadBalancer(policy, envoyProxy)
	require.NoError(t, err)
	require.NotNil(t, lb)
	require.NotNil(t, lb.DynamicModuleLB)
	require.Equal(t, "my-module", lb.DynamicModuleLB.Name)
	require.Equal(t, "round-robin-v2", lb.DynamicModuleLB.ImplementationName)
	require.NotNil(t, lb.DynamicModuleLB.Config)
	require.Equal(t, "/usr/local/lib/my-module.so", lb.DynamicModuleLB.Path)
}

func TestBuildLoadBalancer_DynamicModule_UnregisteredModule(t *testing.T) {
	envoyProxy := &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			DynamicModules: []egv1a1.DynamicModuleEntry{
				{
					Name: "other-module",
					Source: egv1a1.DynamicModuleSource{
						Type: ptr.To(egv1a1.LocalDynamicModuleSourceType),
						Local: &egv1a1.LocalDynamicModuleSource{
							Path: "/usr/local/lib/other.so",
						},
					},
				},
			},
		},
	}

	policy := &egv1a1.ClusterSettings{
		LoadBalancer: &egv1a1.LoadBalancer{
			Type: egv1a1.DynamicModuleLoadBalancerType,
			DynamicModule: &egv1a1.DynamicModuleLBPolicy{
				DynamicModuleRef: egv1a1.DynamicModuleRef{
					Name:               "my-module",
					ImplementationName: ptr.To("round-robin-v2"),
				},
			},
		},
	}

	_, err := buildLoadBalancer(policy, envoyProxy)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not registered")
}

func TestBuildLoadBalancer_DynamicModule_NilEnvoyProxy(t *testing.T) {
	policy := &egv1a1.ClusterSettings{
		LoadBalancer: &egv1a1.LoadBalancer{
			Type: egv1a1.DynamicModuleLoadBalancerType,
			DynamicModule: &egv1a1.DynamicModuleLBPolicy{
				DynamicModuleRef: egv1a1.DynamicModuleRef{
					Name:               "my-module",
					ImplementationName: ptr.To("round-robin-v2"),
				},
			},
		},
	}

	_, err := buildLoadBalancer(policy, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "EnvoyProxy")
}

func TestBuildLoadBalancer_DynamicModule_Remote(t *testing.T) {
	envoyProxy := &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			DynamicModules: []egv1a1.DynamicModuleEntry{
				{
					Name: "remote-module",
					Source: egv1a1.DynamicModuleSource{
						Type: ptr.To(egv1a1.RemoteDynamicModuleSourceType),
						Remote: &egv1a1.RemoteDynamicModuleSource{
							URL:    "https://example.com/module.so",
							SHA256: "abc123def456",
						},
					},
				},
			},
		},
	}

	policy := &egv1a1.ClusterSettings{
		LoadBalancer: &egv1a1.LoadBalancer{
			Type: egv1a1.DynamicModuleLoadBalancerType,
			DynamicModule: &egv1a1.DynamicModuleLBPolicy{
				DynamicModuleRef: egv1a1.DynamicModuleRef{
					Name:               "remote-module",
					ImplementationName: ptr.To("custom-lb"),
				},
			},
		},
	}

	lb, err := buildLoadBalancer(policy, envoyProxy)
	require.NoError(t, err)
	require.NotNil(t, lb)
	require.NotNil(t, lb.DynamicModuleLB)
	require.NotNil(t, lb.DynamicModuleLB.Remote)
	require.Equal(t, "https://example.com/module.so", lb.DynamicModuleLB.Remote.URL)
	require.Equal(t, "abc123def456", lb.DynamicModuleLB.Remote.SHA256)
}
