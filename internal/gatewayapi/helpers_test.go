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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		{
			name: "match expression",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "Gateway",
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "environment",
								Operator: "In",
								Values:   []string{"prod", "staging"},
							},
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
								"environment": "prod",
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
								"environment": "dev",
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
						Name:  "first",
					},
				},
			},
		},
		{
			name: "match expression - bad expression matches nothing",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "Gateway",
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "environment",
								Operator: "Foo",
								Values:   []string{"prod", "staging"},
							},
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
								"environment": "prod",
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
								"environment": "dev",
							},
						},
					},
				},
			},
			results: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results := getPolicyTargetRefs(tc.policy, tc.targets)
			require.ElementsMatch(t, results, tc.results)
		})
	}
}

func TestIsRefToGateway(t *testing.T) {
	cases := []struct {
		name           string
		routeNamespace gwapiv1.Namespace
		parentRef      gwapiv1.ParentReference
		gatewayNN      types.NamespacedName
		expected       bool
	}{
		{
			name:           "match without namespace-true",
			routeNamespace: gwapiv1.Namespace("ns1"),
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
			name:           "match without namespace-false",
			routeNamespace: gwapiv1.Namespace("ns1"),
			parentRef: gwapiv1.ParentReference{
				Name: gwapiv1.ObjectName("eg"),
			},
			gatewayNN: types.NamespacedName{
				Name:      "eg",
				Namespace: "ns2",
			},
			expected: false,
		},
		{
			name:           "match with namespace-true",
			routeNamespace: gwapiv1.Namespace("ns1"),
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
			name:           "match without namespace2-false",
			routeNamespace: gwapiv1.Namespace("ns1"),
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
			got := IsRefToGateway(tc.routeNamespace, tc.parentRef, tc.gatewayNN)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestGetServiceIPFamily(t *testing.T) {
	testCases := []struct {
		name     string
		service  *corev1.Service
		expected *egv1a1.IPFamily
	}{
		{
			name:     "nil service",
			service:  nil,
			expected: nil,
		},
		{
			name: "require dual stack",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					IPFamilyPolicy: ptr.To(corev1.IPFamilyPolicyRequireDualStack),
				},
			},
			expected: ptr.To(egv1a1.DualStack),
		},
		{
			name: "multiple ip families",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
				},
			},
			expected: ptr.To(egv1a1.DualStack),
		},
		{
			name: "ipv4 only",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				},
			},
			expected: ptr.To(egv1a1.IPv4),
		},
		{
			name: "ipv6 only",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv6Protocol},
				},
			},
			expected: ptr.To(egv1a1.IPv6),
		},
		{
			name: "no ip family specified",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{},
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getServiceIPFamily(tc.service)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestGetCaCertFromConfigMap(t *testing.T) {
	cases := []struct {
		name          string
		cm            *corev1.ConfigMap
		expectedFound bool
		expected      string
	}{
		{
			name: "get from ca.crt",
			cm: &corev1.ConfigMap{
				Data: map[string]string{
					"ca.crt":        "fake-cert",
					"root-cert.pem": "fake-root",
				},
			},
			expectedFound: true,
			expected:      "fake-cert",
		},
		{
			name: "get from first key",
			cm: &corev1.ConfigMap{
				Data: map[string]string{
					"root-cert.pem": "fake-root",
				},
			},
			expectedFound: true,
			expected:      "fake-root",
		},
		{
			name: "not found",
			cm: &corev1.ConfigMap{
				Data: map[string]string{},
			},
			expectedFound: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, found := getCaCertFromConfigMap(tc.cm)
			require.Equal(t, tc.expectedFound, found)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestGetCaCertFromSecret(t *testing.T) {
	cases := []struct {
		name          string
		s             *corev1.Secret
		expectedFound bool
		expected      string
	}{
		{
			name: "get from ca.crt",
			s: &corev1.Secret{
				Data: map[string][]byte{
					"ca.crt":        []byte("fake-cert"),
					"root-cert.pem": []byte("fake-root"),
				},
			},
			expectedFound: true,
			expected:      "fake-cert",
		},
		{
			name: "get from first key",
			s: &corev1.Secret{
				Data: map[string][]byte{
					"root-cert.pem": []byte("fake-root"),
				},
			},
			expectedFound: true,
			expected:      "fake-root",
		},
		{
			name: "not found",
			s: &corev1.Secret{
				Data: map[string][]byte{},
			},
			expectedFound: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, found := getCaCertFromSecret(tc.s)
			require.Equal(t, tc.expectedFound, found)
			require.Equal(t, tc.expected, string(got))
		})
	}
}
