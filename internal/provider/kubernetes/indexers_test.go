// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestEndpointSliceIndexFuncs(t *testing.T) {
	testCases := []struct {
		name                  string
		labels                map[string]string
		expectedService       []string
		expectedServiceImport []string
	}{
		{
			name: "service endpointslice",
			labels: map[string]string{
				discoveryv1.LabelServiceName: "backend",
			},
			expectedService: []string{"backend"},
		},
		{
			name: "serviceimport endpointslice",
			labels: map[string]string{
				mcsapiv1a1.LabelServiceName: "imported-backend",
			},
			expectedServiceImport: []string{"imported-backend"},
		},
		{
			name: "unlabeled endpointslice",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			endpointSlice := &discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Labels: tc.labels,
				},
			}

			require.Equal(t, tc.expectedService, serviceEndpointSliceIndexFunc(endpointSlice))
			require.Equal(t, tc.expectedServiceImport, serviceImportEndpointSliceIndexFunc(endpointSlice))
		})
	}
}

func TestBackendGRPCRouteIndexFunc(t *testing.T) {
	testCases := []struct {
		name     string
		route    *gwapiv1.GRPCRoute
		expected []string
	}{
		{
			name: "no filters, single backend",
			route: &gwapiv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grpcroute-1",
					Namespace: "default",
				},
				Spec: gwapiv1.GRPCRouteSpec{
					Rules: []gwapiv1.GRPCRouteRule{
						{
							BackendRefs: []gwapiv1.GRPCBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "service-1",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"default/service-1"},
		},
		{
			name: "request mirror filter includes mirror backendRef",
			route: &gwapiv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grpcroute-mirror",
					Namespace: "default",
				},
				Spec: gwapiv1.GRPCRouteSpec{
					Rules: []gwapiv1.GRPCRouteRule{
						{
							BackendRefs: []gwapiv1.GRPCBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "service-1",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
							Filters: []gwapiv1.GRPCRouteFilter{
								{
									Type: gwapiv1.GRPCRouteFilterRequestMirror,
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Kind: new(gwapiv1.Kind("Service")),
											Name: "mirror-service",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"default/service-1", "default/mirror-service"},
		},
		{
			name: "mirror filter with cross-namespace backendRef",
			route: &gwapiv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grpcroute-cross-ns",
					Namespace: "default",
				},
				Spec: gwapiv1.GRPCRouteSpec{
					Rules: []gwapiv1.GRPCRouteRule{
						{
							BackendRefs: []gwapiv1.GRPCBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "service-1",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
							Filters: []gwapiv1.GRPCRouteFilter{
								{
									Type: gwapiv1.GRPCRouteFilterRequestMirror,
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Kind:      new(gwapiv1.Kind("Service")),
											Namespace: new(gwapiv1.Namespace("other-ns")),
											Name:      "mirror-service",
											Port:      new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"default/service-1", "other-ns/mirror-service"},
		},
		{
			name: "multiple mirror filters in same rule",
			route: &gwapiv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grpcroute-multi-mirror",
					Namespace: "default",
				},
				Spec: gwapiv1.GRPCRouteSpec{
					Rules: []gwapiv1.GRPCRouteRule{
						{
							BackendRefs: []gwapiv1.GRPCBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "service-1",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
							Filters: []gwapiv1.GRPCRouteFilter{
								{
									Type: gwapiv1.GRPCRouteFilterRequestMirror,
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Kind: new(gwapiv1.Kind("Service")),
											Name: "mirror-1",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
								{
									Type: gwapiv1.GRPCRouteFilterRequestMirror,
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Kind: new(gwapiv1.Kind("Service")),
											Name: "mirror-2",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"default/service-1", "default/mirror-1", "default/mirror-2"},
		},
		{
			name: "non-mirror filter is ignored",
			route: &gwapiv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grpcroute-header-filter",
					Namespace: "default",
				},
				Spec: gwapiv1.GRPCRouteSpec{
					Rules: []gwapiv1.GRPCRouteRule{
						{
							BackendRefs: []gwapiv1.GRPCBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "service-1",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
							Filters: []gwapiv1.GRPCRouteFilter{
								{
									Type: gwapiv1.GRPCRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gwapiv1.HTTPHeaderFilter{
										Set: []gwapiv1.HTTPHeader{
											{Name: "x-custom", Value: "value"},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"default/service-1"},
		},
		{
			name: "backend with Backend kind is indexed",
			route: &gwapiv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grpcroute-backend-kind",
					Namespace: "default",
				},
				Spec: gwapiv1.GRPCRouteSpec{
					Rules: []gwapiv1.GRPCRouteRule{
						{
							BackendRefs: []gwapiv1.GRPCBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Kind: new(gwapiv1.Kind(egv1a1.KindBackend)),
											Name: "backend-1",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"default/backend-1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := backendGRPCRouteIndexFunc(tc.route)
			require.ElementsMatch(t, tc.expected, result)
		})
	}
}

func TestEnvoyProxyDynamicModuleIndexFuncs(t *testing.T) {
	imageModule := func(name string, image *egv1a1.ImageDynamicModuleSource) egv1a1.DynamicModuleEntry {
		return egv1a1.DynamicModuleEntry{
			Name: name,
			Source: egv1a1.DynamicModuleSource{
				Type:  new(egv1a1.ImageDynamicModuleSourceType),
				Image: image,
			},
		}
	}
	caRef := func(kind, namespace, name string) *egv1a1.WasmCodeSourceTLSConfig {
		ref := gwapiv1.SecretObjectReference{Name: gwapiv1.ObjectName(name)}
		if kind != "" {
			ref.Kind = new(gwapiv1.Kind(kind))
		}
		if namespace != "" {
			ref.Namespace = new(gwapiv1.Namespace(namespace))
		}
		return &egv1a1.WasmCodeSourceTLSConfig{CACertificateRef: ref}
	}

	ep := &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "proxy"},
		Spec: egv1a1.EnvoyProxySpec{
			DynamicModules: []egv1a1.DynamicModuleEntry{
				{
					Name: "local-module",
					Source: egv1a1.DynamicModuleSource{
						Local: &egv1a1.LocalDynamicModuleSource{Path: "/lib/liblocal.so"},
					},
				},
				imageModule("no-tls", &egv1a1.ImageDynamicModuleSource{URL: "registry.example.com/no-tls"}),
				imageModule("pull-secret", &egv1a1.ImageDynamicModuleSource{
					URL:           "registry.example.com/private",
					PullSecretRef: &gwapiv1.SecretObjectReference{Name: "pull-secret"},
				}),
				imageModule("ca-secret-default-kind", &egv1a1.ImageDynamicModuleSource{
					URL: "registry.example.com/a",
					TLS: caRef("", "", "ca-secret"),
				}),
				imageModule("ca-configmap", &egv1a1.ImageDynamicModuleSource{
					URL: "registry.example.com/b",
					TLS: caRef("ConfigMap", "other", "ca-configmap"),
				}),
				imageModule("ca-cluster-trust-bundle", &egv1a1.ImageDynamicModuleSource{
					URL: "registry.example.com/c",
					TLS: caRef("ClusterTrustBundle", "", "ca-bundle"),
				}),
			},
		},
	}

	require.ElementsMatch(t, []string{"default/pull-secret", "default/ca-secret"}, secretEnvoyProxyIndexFunc(ep))
	require.Equal(t, []string{"other/ca-configmap"}, configMapEnvoyProxyIndexFunc(ep))
	require.Equal(t, []string{"ca-bundle"}, clusterTrustBundleEnvoyProxyIndexFunc(ep))
}
