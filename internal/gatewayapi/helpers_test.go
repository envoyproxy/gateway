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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
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

func TestResolvePolicyTargets(t *testing.T) {
	testCases := []struct {
		name       string
		policy     egv1a1.PolicyTargetReferences
		targets    []*unstructured.Unstructured
		namespaces []*corev1.Namespace
		grants     []*gwapiv1b1.ReferenceGrant
		results    []policyTargetReferenceWithSectionName
	}{
		{
			name: "simple",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind:  "Gateway",
						Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
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
			results: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "second",
					Namespace: "default",
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
			results: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "TLSRoute",
					Name:      "third",
					Namespace: "default",
				},
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "second",
					Namespace: "default",
				},
			},
		},
		{
			name: "deduplicated",
			policy: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
			results: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "TLSRoute",
					Name:      "third",
					Namespace: "default",
				},
			},
		},
		{
			name: "bad-group-is-ignored",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind:  "Gateway",
						Group: new(gwapiv1.Group("bad-group")),
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
			results: []policyTargetReferenceWithSectionName{},
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
			results: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "first",
					Namespace: "default",
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
			results: []policyTargetReferenceWithSectionName{},
		},
		{
			name: "namespaces from same",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "Gateway",
						Namespaces: &egv1a1.TargetSelectorNamespaces{
							From: egv1a1.TargetNamespaceFromSame,
						},
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
							"name":      "same-ns",
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
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "other-ns",
							"namespace": "other",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			results: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "same-ns",
					Namespace: "default",
				},
			},
		},
		{
			name: "namespaces from all",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "Gateway",
						Namespaces: &egv1a1.TargetSelectorNamespaces{
							From: egv1a1.TargetNamespaceFromAll,
						},
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
							"name":      "same-ns",
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
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "other-ns",
							"namespace": "other",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			grants: []*gwapiv1b1.ReferenceGrant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-default-btp",
						Namespace: "other",
					},
					Spec: gwapiv1b1.ReferenceGrantSpec{
						From: []gwapiv1b1.ReferenceGrantFrom{
							{
								Group:     gwapiv1b1.Group(egv1a1.GroupVersion.Group),
								Kind:      gwapiv1b1.Kind("BackendTrafficPolicy"),
								Namespace: gwapiv1b1.Namespace("default"),
							},
						},
						To: []gwapiv1b1.ReferenceGrantTo{
							{
								Group: gwapiv1b1.Group(gwapiv1.GroupName),
								Kind:  gwapiv1b1.Kind("Gateway"),
							},
						},
					},
				},
			},
			results: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "same-ns",
					Namespace: "default",
				},
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "other-ns",
					Namespace: "other",
				},
			},
		},
		{
			name: "namespaces from selector",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "Gateway",
						Namespaces: &egv1a1.TargetSelectorNamespaces{
							From: egv1a1.TargetNamespaceFromSelector,
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"team": "blue",
								},
							},
						},
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
							"name":      "selected-ns",
							"namespace": "selected",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "gateway.networking.k8s.io/v1",
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "unselected-ns",
							"namespace": "unselected",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			namespaces: []*corev1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "selected",
						Labels: map[string]string{"team": "blue"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "unselected",
						Labels: map[string]string{"team": "green"},
					},
				},
			},
			grants: []*gwapiv1b1.ReferenceGrant{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "allow-default-btp",
						Namespace: "selected",
					},
					Spec: gwapiv1b1.ReferenceGrantSpec{
						From: []gwapiv1b1.ReferenceGrantFrom{
							{
								Group:     gwapiv1b1.Group(egv1a1.GroupVersion.Group),
								Kind:      gwapiv1b1.Kind("BackendTrafficPolicy"),
								Namespace: gwapiv1b1.Namespace("default"),
							},
						},
						To: []gwapiv1b1.ReferenceGrantTo{
							{
								Group: gwapiv1b1.Group(gwapiv1.GroupName),
								Kind:  gwapiv1b1.Kind("Gateway"),
							},
						},
					},
				},
			},
			results: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "selected-ns",
					Namespace: "selected",
				},
			},
		},
		{
			name: "namespaces from selector requires known namespace labels",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "Gateway",
						Namespaces: &egv1a1.TargetSelectorNamespaces{
							From: egv1a1.TargetNamespaceFromSelector,
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"team": "blue",
								},
							},
						},
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
							"name":      "unknown-ns",
							"namespace": "unknown",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			results: []policyTargetReferenceWithSectionName{},
		},
		{
			name: "namespaces from all cross-namespace requires reference grant",
			policy: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "Gateway",
						Namespaces: &egv1a1.TargetSelectorNamespaces{
							From: egv1a1.TargetNamespaceFromAll,
						},
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
							"name":      "same-ns",
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
						"kind":       "Gateway",
						"metadata": map[string]any{
							"name":      "other-ns",
							"namespace": "other",
							"labels": map[string]any{
								"pick": "me",
							},
						},
					},
				},
			},
			results: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "same-ns",
					Namespace: "default",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			namespaceMap := map[string]*corev1.Namespace{}
			for _, ns := range tc.namespaces {
				namespaceMap[ns.Name] = ns
			}

			results := resolvePolicyTargets(
				tc.policy,
				tc.targets,
				tc.grants,
				egv1a1.GroupName,
				egv1a1.KindBackendTrafficPolicy,
				"default",
				func(name string) *corev1.Namespace {
					return namespaceMap[name]
				},
			)
			require.ElementsMatch(t, results, tc.results)
		})
	}
}

func TestResolvePolicyTargetsFromReferences(t *testing.T) {
	testCases := []struct {
		name            string
		targetRefs      egv1a1.PolicyTargetReferences
		policyNamespace string
		expected        []policyTargetReferenceWithSectionName
	}{
		{
			name: "target ref",
			targetRefs: egv1a1.PolicyTargetReferences{
				TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "eg",
					},
					SectionName: SectionNamePtr("http"),
				},
			},
			policyNamespace: "default",
			expected: []policyTargetReferenceWithSectionName{
				{
					Group:       "gateway.networking.k8s.io",
					Kind:        "Gateway",
					Name:        "eg",
					Namespace:   "default",
					SectionName: SectionNamePtr("http"),
				},
			},
		},
		{
			name: "target refs",
			targetRefs: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Group: "gateway.networking.k8s.io",
							Kind:  "Gateway",
							Name:  "first",
						},
					},
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Group: "gateway.networking.k8s.io",
							Kind:  "Gateway",
							Name:  "second",
						},
					},
				},
			},
			policyNamespace: "default",
			expected: []policyTargetReferenceWithSectionName{
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "first",
					Namespace: "default",
				},
				{
					Group:     "gateway.networking.k8s.io",
					Kind:      "Gateway",
					Name:      "second",
					Namespace: "default",
				},
			},
		},
		{
			name: "empty target ref is ignored",
			targetRefs: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{},
				},
			},
			policyNamespace: "default",
			expected:        []policyTargetReferenceWithSectionName{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := resolvePolicyTargetsFromReferences(tc.targetRefs, tc.policyNamespace)
			require.Equal(t, tc.expected, actual)
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
					IPFamilyPolicy: new(corev1.IPFamilyPolicyRequireDualStack),
				},
			},
			expected: new(egv1a1.DualStack),
		},
		{
			name: "multiple ip families",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
				},
			},
			expected: new(egv1a1.DualStack),
		},
		{
			name: "ipv4 only",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
				},
			},
			expected: new(egv1a1.IPv4),
		},
		{
			name: "ipv6 only",
			service: &corev1.Service{
				Spec: corev1.ServiceSpec{
					IPFamilies: []corev1.IPFamily{corev1.IPv6Protocol},
				},
			},
			expected: new(egv1a1.IPv6),
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
			got, found := getOrFirstFromData(tc.cm.Data, CACertKey)
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
			got, found := getOrFirstFromData(tc.s.Data, CACertKey)
			require.Equal(t, tc.expectedFound, found)
			require.Equal(t, tc.expected, string(got))
		})
	}
}

func TestIrStringMatch(t *testing.T) {
	const stringMatchUnknown egv1a1.StringMatchType = "Unknown"
	matchName := "Name"
	matchValue := "Value"

	testCases := []struct {
		name     string
		match    egv1a1.StringMatch
		expected *ir.StringMatch
	}{
		{
			name: "Exact by default",
			match: egv1a1.StringMatch{
				Type:  nil,
				Value: matchValue,
			},
			expected: &ir.StringMatch{
				Name:  matchName,
				Exact: &matchValue,
			},
		},
		{
			name: "Exact",
			match: egv1a1.StringMatch{
				Type:  new(egv1a1.StringMatchExact),
				Value: matchValue,
			},
			expected: &ir.StringMatch{
				Name:  matchName,
				Exact: &matchValue,
			},
		},
		{
			name: "Prefix",
			match: egv1a1.StringMatch{
				Type:  new(egv1a1.StringMatchPrefix),
				Value: matchValue,
			},
			expected: &ir.StringMatch{
				Name:   matchName,
				Prefix: &matchValue,
			},
		},
		{
			name: "Suffix",
			match: egv1a1.StringMatch{
				Type:  new(egv1a1.StringMatchSuffix),
				Value: matchValue,
			},
			expected: &ir.StringMatch{
				Name:   matchName,
				Suffix: &matchValue,
			},
		},
		{
			name: "RegularExpression",
			match: egv1a1.StringMatch{
				Type:  new(egv1a1.StringMatchRegularExpression),
				Value: matchValue,
			},
			expected: &ir.StringMatch{
				Name:      matchName,
				SafeRegex: &matchValue,
			},
		},
		{
			name: "Unknown",
			match: egv1a1.StringMatch{
				Type:  new(stringMatchUnknown),
				Value: matchValue,
			},
			expected: &ir.StringMatch{
				Name:  matchName,
				Exact: &matchValue,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := irStringMatch(matchName, tc.match)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestWildcardHostnameMatchesHostname(t *testing.T) {
	testCases := []struct {
		name     string
		wildcard string
		hostname string
		expected bool
	}{
		{
			name:     "*.com matches *.example.com",
			wildcard: "*.com",
			hostname: "*.example.com",
			expected: true,
		},
		{
			name:     "*.example.com matches *.foo.example.com",
			wildcard: "*.example.com",
			hostname: "*.foo.example.com",
			expected: true,
		},
		{
			name:     "*.com does not match *.net",
			wildcard: "*.com",
			hostname: "*.net",
			expected: false,
		},
		{
			name:     "*.example.com does not match *.other.com",
			wildcard: "*.example.com",
			hostname: "*.other.com",
			expected: false,
		},
		{
			name:     "*.foo.example.com does not match *.example.com",
			wildcard: "*.foo.example.com",
			hostname: "*.example.com",
			expected: false,
		},
		{
			name:     "*.example.com match foo.example.com",
			wildcard: "*.example.com",
			hostname: "foo.example.com",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := wildcardHostnameMatchesHostname(tc.wildcard, tc.hostname)
			require.Equal(t, tc.expected, result)
		})
	}
}
