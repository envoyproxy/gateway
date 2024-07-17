// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestValidateGRPCFilterRef(t *testing.T) {
	testCases := []struct {
		name     string
		filter   *gwapiv1.GRPCRouteFilter
		expected bool
	}{
		{
			name: "request mirror filter",
			filter: &gwapiv1.GRPCRouteFilter{
				Type: gwapiv1.GRPCRouteFilterRequestMirror,
			},
			expected: true,
		},
		{
			name: "request header modifier filter",
			filter: &gwapiv1.GRPCRouteFilter{
				Type: gwapiv1.GRPCRouteFilterRequestHeaderModifier,
			},
			expected: true,
		},
		{
			name: "response header modifier filter",
			filter: &gwapiv1.GRPCRouteFilter{
				Type: gwapiv1.GRPCRouteFilterResponseHeaderModifier,
			},
			expected: true,
		},
		{
			name: "valid extension resource",
			filter: &gwapiv1.GRPCRouteFilter{
				Type: gwapiv1.GRPCRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "example.io",
					Kind:  "Foo",
					Name:  "test",
				},
			},
			expected: true,
		},
		{
			name: "unsupported extended filter",
			filter: &gwapiv1.GRPCRouteFilter{
				Type: gwapiv1.GRPCRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "UnsupportedGroup",
					Kind:  "UnsupportedKind",
					Name:  "test",
				},
			},
			expected: false,
		},
		{
			name: "empty extended filter",
			filter: &gwapiv1.GRPCRouteFilter{
				Type: gwapiv1.GRPCRouteFilterExtensionRef,
			},
			expected: false,
		},
		{
			name: "invalid filter type",
			filter: &gwapiv1.GRPCRouteFilter{
				Type: "Invalid",
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "example.io",
					Kind:  "Foo",
					Name:  "test",
				},
			},
			expected: false,
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGRPCRouteFilter(tc.filter, schema.GroupKind{Group: "example.io", Kind: "Foo"})
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestValidateHTTPFilterRef(t *testing.T) {
	testCases := []struct {
		name     string
		filter   *gwapiv1.HTTPRouteFilter
		expected bool
	}{
		{
			name: "request mirror filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterRequestMirror,
			},
			expected: true,
		},
		{
			name: "url rewrite filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterURLRewrite,
			},
			expected: true,
		},
		{
			name: "request header modifier filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterRequestHeaderModifier,
			},
			expected: true,
		},
		{
			name: "request redirect filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterRequestRedirect,
			},
			expected: true,
		},
		{
			name: "unsupported extended filter",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "UnsupportedGroup",
					Kind:  "UnsupportedKind",
					Name:  "test",
				},
			},
			expected: false,
		},
		{
			name: "extended filter with missing reference",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterExtensionRef,
			},
			expected: false,
		},
		{
			name: "valid extension resource",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "example.io",
					Kind:  "Foo",
					Name:  "test",
				},
			},
			expected: true,
		},
		{
			name: "invalid filter type",
			filter: &gwapiv1.HTTPRouteFilter{
				Type: "Invalid",
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: "example.io",
					Kind:  "Foo",
					Name:  "test",
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateHTTPRouteFilter(tc.filter, schema.GroupKind{Group: "example.io", Kind: "Foo"})
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetPolicyTargetRefs(t *testing.T) {
	testCases := []struct {
		name    string
		policy  egv1a1.PolicyTargetReferences
		targets []*unstructured.Unstructured
		results []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName
	}{
		{
			name: "simple",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind:  "Gateway",
						Group: ptr.To(gwapiv1.Group("gateway.networking.k8s.io")),
						MatchLabels: map[string]string{
							"pick": "me",
						},
					},
				},
			},
			targets: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "first",
							"namespace": "default",
							"labels": map[string]any{
								"some": "random label",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "second",
							"namespace": "default",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "TLSRoute",
						"metadata": map[string]any{
							"name":      "third",
							"namespace": "default",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			results: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "second",
					},
				},
			},
		},
		{
			name: "multiple selectors",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "TLSRoute",
						MatchLabels: map[string]string{
							"pick": "me",
						},
					},
					{
						Kind: "Gateway",
						MatchLabels: map[string]string{
							"pick": "me",
						},
					},
				},
			},
			targets: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "first",
							"namespace": "default",
							"labels": map[string]any{
								"some": "random label",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "second",
							"namespace": "default",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "TLSRoute",
						"metadata": map[string]any{
							"name":      "third",
							"namespace": "default",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			results: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
						Group: "gateway.networking.k8s.io",
						Kind:  "TLSRoute",
						Name:  "third",
					},
				},
				{
					LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "second",
					},
				},
			},
		},
		{
			name: "deduplicated",
			policy: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
							Group: "gateway.networking.k8s.io",
							Kind:  "TLSRoute",
							Name:  "third",
						},
					},
				},
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "TLSRoute",
						MatchLabels: map[string]string{
							"pick": "me",
						},
					},
				},
			},
			targets: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "first",
							"namespace": "default",
							"labels": map[string]any{
								"some": "random label",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "second",
							"namespace": "default",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "TLSRoute",
						"metadata": map[string]any{
							"name":      "third",
							"namespace": "default",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			results: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
						Group: "gateway.networking.k8s.io",
						Kind:  "TLSRoute",
						Name:  "third",
					},
				},
			},
		},
		{
			name: "bad-group-is-ignored",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind:  "Gateway",
						Group: ptr.To(gwapiv1.Group("bad-group")),
						MatchLabels: map[string]string{
							"pick": "me",
						},
					},
				},
			},
			targets: []*unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "first",
							"namespace": "default",
							"labels": map[string]any{
								"some": "random label",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "second",
							"namespace": "default",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "TLSRoute",
						"metadata": map[string]any{
							"name":      "third",
							"namespace": "default",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			results: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			results := getPolicyTargetRefs(tc.policy, tc.targets)
			require.ElementsMatch(t, results, tc.results)
		})
	}
}

func TestIsRefToGateway(t *testing.T) {
	cases := []struct {
		name      string
		parentRef gwapiv1.ParentReference
		gatewayNN types.NamespacedName
		expected  bool
	}{
		{
			name: "match without namespace",
			parentRef: gwapiv1.ParentReference{
				Name: gwapiv1.ObjectName("eg"),
			},
			gatewayNN: types.NamespacedName{
				Name:      "eg",
				Namespace: "ns1",
			},
			expected: true,
		},
		{
			name: "match without namespace2",
			parentRef: gwapiv1.ParentReference{
				Name: gwapiv1.ObjectName("eg"),
			},
			gatewayNN: types.NamespacedName{
				Name:      "eg",
				Namespace: "ns2",
			},
			expected: true,
		},
		{
			name: "match with namespace",
			parentRef: gwapiv1.ParentReference{
				Name:      gwapiv1.ObjectName("eg"),
				Namespace: NamespacePtr("ns1"),
			},
			gatewayNN: types.NamespacedName{
				Name:      "eg",
				Namespace: "ns1",
			},
			expected: true,
		},
		{
			name: "match without namespace2",
			parentRef: gwapiv1.ParentReference{
				Name:      gwapiv1.ObjectName("eg"),
				Namespace: NamespacePtr("ns2"),
			},
			gatewayNN: types.NamespacedName{
				Name:      "eg",
				Namespace: "ns1",
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsRefToGateway(tc.parentRef, tc.gatewayNN)
			require.Equal(t, tc.expected, got)
		})
	}
}
