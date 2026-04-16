// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

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
											Port: ptr.To(gwapiv1.PortNumber(8080)),
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
											Port: ptr.To(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
							Filters: []gwapiv1.GRPCRouteFilter{
								{
									Type: gwapiv1.GRPCRouteFilterRequestMirror,
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Kind: ptr.To(gwapiv1.Kind("Service")),
											Name: "mirror-service",
											Port: ptr.To(gwapiv1.PortNumber(8080)),
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
											Port: ptr.To(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
							Filters: []gwapiv1.GRPCRouteFilter{
								{
									Type: gwapiv1.GRPCRouteFilterRequestMirror,
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Kind:      ptr.To(gwapiv1.Kind("Service")),
											Namespace: ptr.To(gwapiv1.Namespace("other-ns")),
											Name:      "mirror-service",
											Port:      ptr.To(gwapiv1.PortNumber(8080)),
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
											Port: ptr.To(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
							Filters: []gwapiv1.GRPCRouteFilter{
								{
									Type: gwapiv1.GRPCRouteFilterRequestMirror,
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Kind: ptr.To(gwapiv1.Kind("Service")),
											Name: "mirror-1",
											Port: ptr.To(gwapiv1.PortNumber(8080)),
										},
									},
								},
								{
									Type: gwapiv1.GRPCRouteFilterRequestMirror,
									RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
										BackendRef: gwapiv1.BackendObjectReference{
											Kind: ptr.To(gwapiv1.Kind("Service")),
											Name: "mirror-2",
											Port: ptr.To(gwapiv1.PortNumber(8080)),
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
											Port: ptr.To(gwapiv1.PortNumber(8080)),
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
											Kind: ptr.To(gwapiv1.Kind(egv1a1.KindBackend)),
											Name: "backend-1",
											Port: ptr.To(gwapiv1.PortNumber(8080)),
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
