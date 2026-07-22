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
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
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

func TestPolicyScopeGraphGetDirectChildren(t *testing.T) {
	gatewayNN := types.NamespacedName{Namespace: "default", Name: "gateway"}
	listenerSetNN := types.NamespacedName{Namespace: "default", Name: "listener-set"}
	routeNN := types.NamespacedName{Namespace: "default", Name: "route"}
	gateway := gatewayScope(gatewayNN)
	httpListener := gatewayListenerScope(gatewayNN, gwapiv1.SectionName("http"))
	route := routeScope(routeNN)

	testCases := []struct {
		name     string
		setup    []policyScopeGraphSetup
		parent   policyScope
		expected []policyScope
	}{
		{
			name:   "empty graph",
			parent: gateway,
		},
		{
			name: "returns only direct children",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: gateway, child: httpListener},
				{kind: policyScopeGraphSetupAdd, parent: httpListener, child: route},
			},
			parent:   gateway,
			expected: []policyScope{httpListener},
		},
		{
			name: "ignores registered listener set containment",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: gateway, child: httpListener},
				{kind: policyScopeGraphSetupRegisterListenerSet, listenerSet: listenerSetNN, gateway: gatewayNN},
			},
			parent:   gateway,
			expected: []policyScope{httpListener},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			graph := newTestPolicyScopeGraph(t, tc.setup)

			got := graph.GetDirectChildren(tc.parent)
			requirePolicyScopesEqual(t, got, tc.expected...)
		})
	}
}

func TestPolicyScopeGraphGetWithDescendants(t *testing.T) {
	gatewayNN := types.NamespacedName{Namespace: "default", Name: "gateway"}
	otherGatewayNN := types.NamespacedName{Namespace: "default", Name: "other-gateway"}
	listenerSetNN := types.NamespacedName{Namespace: "default", Name: "listener-set"}

	gateway := gatewayScope(gatewayNN)
	otherGateway := gatewayScope(otherGatewayNN)
	httpListener := gatewayListenerScope(gatewayNN, gwapiv1.SectionName("http"))
	httpsListener := gatewayListenerScope(gatewayNN, gwapiv1.SectionName("https"))
	otherGatewayListener := gatewayListenerScope(otherGatewayNN, gwapiv1.SectionName("http"))
	listenerSet := listenerSetScope(listenerSetNN)
	listenerSetHTTPListener := listenerSetListenerScope(listenerSetNN, gwapiv1.SectionName("ls-http"))
	listenerSetHTTPSListener := listenerSetListenerScope(listenerSetNN, gwapiv1.SectionName("ls-https"))

	gatewayRoute := routeScope(types.NamespacedName{Namespace: "default", Name: "gateway-route"})
	gatewayListenerRoute := routeScope(types.NamespacedName{Namespace: "default", Name: "gateway-listener-route"})
	listenerSetRoute := routeScope(types.NamespacedName{Namespace: "default", Name: "listener-set-route"})
	listenerSetListenerRoute := routeScope(types.NamespacedName{Namespace: "default", Name: "listener-set-listener-route"})
	otherGatewayRoute := routeScope(types.NamespacedName{Namespace: "default", Name: "other-gateway-route"})

	testCases := []struct {
		name       string
		setup      []policyScopeGraphSetup
		parent     policyScope
		expected   []policyScope
		unexpected []policyScope
	}{
		{
			name:   "empty graph",
			parent: gateway,
		},
		{
			name: "resource scope returns direct children",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: gateway, child: httpsListener},
				{kind: policyScopeGraphSetupAdd, parent: gateway, child: gatewayRoute},
			},
			parent:   gateway,
			expected: []policyScope{httpsListener, gatewayRoute},
		},
		{
			name: "gateway includes gateway listener descendants",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: httpListener, child: gatewayListenerRoute},
			},
			parent:   gateway,
			expected: []policyScope{gatewayListenerRoute},
		},
		{
			name: "gateway includes nested listener set descendants",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupRegisterListenerSet, listenerSet: listenerSetNN, gateway: gatewayNN},
				{kind: policyScopeGraphSetupAdd, parent: listenerSet, child: listenerSetRoute},
				{kind: policyScopeGraphSetupAdd, parent: listenerSetHTTPListener, child: listenerSetListenerRoute},
			},
			parent:   gateway,
			expected: []policyScope{listenerSetRoute, listenerSetListenerRoute},
		},
		{
			name: "listener set includes listener descendants",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: listenerSet, child: listenerSetRoute},
				{kind: policyScopeGraphSetupAdd, parent: listenerSetHTTPListener, child: listenerSetListenerRoute},
			},
			parent:   listenerSet,
			expected: []policyScope{listenerSetRoute, listenerSetListenerRoute},
		},
		{
			name: "gateway listener includes direct children and resource-level routes only",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: gateway, child: gatewayRoute},
				{kind: policyScopeGraphSetupAdd, parent: gateway, child: httpsListener},
				{kind: policyScopeGraphSetupAdd, parent: httpListener, child: gatewayListenerRoute},
			},
			parent:     httpListener,
			expected:   []policyScope{gatewayRoute, gatewayListenerRoute},
			unexpected: []policyScope{httpsListener},
		},
		{
			name: "listener set listener includes direct children and resource-level routes only",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: listenerSet, child: listenerSetRoute},
				{kind: policyScopeGraphSetupAdd, parent: listenerSet, child: listenerSetHTTPSListener},
				{kind: policyScopeGraphSetupAdd, parent: listenerSetHTTPListener, child: listenerSetListenerRoute},
			},
			parent:     listenerSetHTTPListener,
			expected:   []policyScope{listenerSetRoute, listenerSetListenerRoute},
			unexpected: []policyScope{listenerSetHTTPSListener},
		},
		{
			name: "gateway ignores unregistered listener set containment",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: listenerSet, child: listenerSetRoute},
			},
			parent: gateway,
		},
		{
			name: "gateway ignores other gateway descendants",
			setup: []policyScopeGraphSetup{
				{kind: policyScopeGraphSetupAdd, parent: otherGateway, child: otherGatewayRoute},
				{kind: policyScopeGraphSetupAdd, parent: otherGatewayListener, child: gatewayListenerRoute},
			},
			parent: gateway,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			graph := newTestPolicyScopeGraph(t, tc.setup)

			got := graph.GetWithDescendants(tc.parent)
			requirePolicyScopesEqual(t, got, tc.expected...)
			for _, scope := range tc.unexpected {
				require.False(t, got.Has(scope), "unexpected scope %v", scope)
			}
		})
	}
}

type policyScopeGraphSetupKind string

const (
	policyScopeGraphSetupAdd                 policyScopeGraphSetupKind = "add"
	policyScopeGraphSetupRegisterListenerSet policyScopeGraphSetupKind = "registerListenerSet"
)

type policyScopeGraphSetup struct {
	kind policyScopeGraphSetupKind

	parent policyScope
	child  policyScope

	listenerSet types.NamespacedName
	gateway     types.NamespacedName
}

func newTestPolicyScopeGraph(t *testing.T, setup []policyScopeGraphSetup) policyScopeGraph {
	t.Helper()

	graph := newPolicyScopeGraph()
	for i := range setup {
		step := &setup[i]
		switch step.kind {
		case policyScopeGraphSetupAdd:
			graph.Add(step.parent, step.child)
		case policyScopeGraphSetupRegisterListenerSet:
			graph.RegisterListenerSet(step.listenerSet, step.gateway)
		default:
			require.FailNowf(t, "unknown policy scope graph setup kind", "kind: %s", step.kind)
		}
	}
	return graph
}

func requirePolicyScopesEqual(t *testing.T, actual sets.Set[policyScope], expected ...policyScope) {
	t.Helper()

	require.Equal(t, len(expected), actual.Len())
	for _, scope := range expected {
		require.True(t, actual.Has(scope), "expected scope %v", scope)
	}
}

func TestIrBackendClusterName(t *testing.T) {
	tests := []struct {
		name          string
		key           *BackendClusterKey
		mergeGateways bool
		want          string
	}{
		{
			name: "service with port, no protocol",
			key:  &BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080},
			want: "backend/service/default/service-1/8080",
		},
		{
			name: "backend kind, http protocol",
			key:  &BackendClusterKey{Kind: "Backend", Namespace: "ns", Name: "be", Port: 443, Protocol: ir.HTTP},
			want: "backend/backend/ns/be/443/http",
		},
		{
			name: "service with grpc protocol differs from http",
			key:  &BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080, Protocol: ir.GRPC},
			want: "backend/service/default/service-1/8080/grpc",
		},
		{
			name:          "mergeGateways appends the owning Gateway's identity",
			key:           &BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080, GatewayIRKey: "envoy-gateway/gateway-1"},
			mergeGateways: true,
			want:          "backend/service/default/service-1/8080/envoy-gateway/gateway-1",
		},
		{
			name: "GatewayIRKey is ignored when mergeGateways is false",
			key:  &BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080, GatewayIRKey: "envoy-gateway/gateway-1"},
			want: "backend/service/default/service-1/8080",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, irBackendClusterName(tc.key, tc.mergeGateways))
		})
	}
}

func TestIsMergeBackendsEnabled(t *testing.T) {
	enabled := &egv1a1.MergeBackendsConfig{Enabled: new(true)}
	disabled := &egv1a1.MergeBackendsConfig{Enabled: new(false)}

	tests := []struct {
		name string
		res  *resource.Resources
		want bool
	}{
		{
			name: "gatewayclass envoyproxy set",
			res: &resource.Resources{
				EnvoyProxyForGatewayClass: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{MergeBackends: enabled}},
			},
			want: true,
		},
		{
			name: "default spec set",
			res: &resource.Resources{
				EnvoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{MergeBackends: enabled},
			},
			want: true,
		},
		{
			name: "gatewayclass envoyproxy takes precedence over default spec",
			res: &resource.Resources{
				EnvoyProxyForGatewayClass: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{MergeBackends: disabled}},
				EnvoyProxyDefaultSpec:     &egv1a1.EnvoyProxySpec{MergeBackends: enabled},
			},
			want: false,
		},
		{
			name: "gatewayclass envoyproxy set but MergeBackends nil falls back to default spec",
			res: &resource.Resources{
				EnvoyProxyForGatewayClass: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{}},
				EnvoyProxyDefaultSpec:     &egv1a1.EnvoyProxySpec{MergeBackends: enabled},
			},
			want: true,
		},
		{
			name: "unset",
			res:  &resource.Resources{},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsMergeBackendsEnabled(tc.res))
		})
	}
}

// structFieldNames returns t's field names, flattening one level of anonymous/embedded struct
// fields via Go's own field promotion, skipping any name in skip.
func structFieldNames(t reflect.Type, skip map[string]bool) []string {
	var names []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if skip[f.Name] {
			continue
		}
		if f.Anonymous {
			names = append(names, structFieldNames(f.Type, skip)...)
			continue
		}
		names = append(names, f.Name)
	}
	return names
}

// structWithFieldSet returns a zero-value *T with only the named field set to a representative
// non-nil/non-empty value, for behaviorally testing a "does this field affect X" classifier in
// isolation from every other field.
func structWithFieldSet[T any](fieldName string) *T {
	specPtr := new(T)
	v := reflect.ValueOf(specPtr).Elem()
	field := v.FieldByName(fieldName)
	switch field.Kind() {
	case reflect.Ptr:
		field.Set(reflect.New(field.Type().Elem()))
	case reflect.Slice:
		field.Set(reflect.MakeSlice(field.Type(), 1, 1))
	default:
		panic(fmt.Sprintf("structWithFieldSet: unsupported field kind %s for field %q", field.Kind(), fieldName))
	}
	return specPtr
}
