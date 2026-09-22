// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExtractTargetRefs(t *testing.T) {
	tests := []struct {
		desc          string
		specInput     map[string]any
		output        egv1a1.PolicyTargetReferences
		expectedError string
	}{
		{
			desc:          "no spec",
			specInput:     nil,
			output:        egv1a1.PolicyTargetReferences{},
			expectedError: "no targets found for the policy",
		},
		{
			desc: "no targetRef",
			specInput: map[string]any{
				"someAttr": "someValue",
			},
			output:        egv1a1.PolicyTargetReferences{},
			expectedError: "no targets found for the policy",
		},
		{
			desc: "targetRefs is not an array",
			specInput: map[string]any{
				"targetRefs": "someValue",
			},
			output:        egv1a1.PolicyTargetReferences{},
			expectedError: "no targets found for the policy",
		},
		{
			desc: "invalid targetref",
			specInput: map[string]any{
				"targetRef": map[string]any{
					"someKey": "someValue",
				},
			},
			output:        egv1a1.PolicyTargetReferences{},
			expectedError: "no targets found for the policy",
		},
		{
			desc: "valid single targetRef",
			specInput: map[string]any{
				"targetRef": map[string]any{
					"group": "some.group",
					"kind":  "SomeKind",
					"name":  "name",
				},
			},
			output: egv1a1.PolicyTargetReferences{
				TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
						Group: "some.group",
						Kind:  "SomeKind",
						Name:  "name",
					},
				},
			},
		},
		{
			desc: "valid targetSelectors",
			specInput: map[string]any{
				"targetSelectors": []any{
					map[string]any{
						"kind": "SomeKind",
						"matchLabels": map[string]any{
							"some": "name",
						},
					},
				},
			},
			output: egv1a1.PolicyTargetReferences{
				TargetSelectors: []egv1a1.TargetSelector{
					{
						Kind: "SomeKind",
						MatchLabels: map[string]string{
							"some": "name",
						},
					},
				},
			},
		},
		{
			desc: "valid multiple targetRefs",
			specInput: map[string]any{
				"targetRefs": []any{
					map[string]any{
						"group": "some.group",
						"kind":  "SomeKind2",
						"name":  "othername",
					},
					map[string]any{
						"group": "some.group",
						"kind":  "SomeKind",
						"name":  "name",
					},
				},
			},
			output: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Group: "some.group",
							Kind:  "SomeKind2",
							Name:  "othername",
						},
					},
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Group: "some.group",
							Kind:  "SomeKind",
							Name:  "name",
						},
					},
				},
			},
		},
	}

	for _, currTest := range tests {
		t.Run(currTest.desc, func(t *testing.T) {
			policy := &unstructured.Unstructured{
				Object: map[string]any{},
			}
			policy.Object["spec"] = currTest.specInput
			targets, err := extractTargetRefs(policy)

			if currTest.expectedError != "" {
				require.EqualError(t, err, currTest.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, currTest.output, targets)
			}
		})
	}
}

func TestAppendUnstructuredRefIfAbsent(t *testing.T) {
	tt := &Translator{}
	p1 := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "p1"}}}
	p2 := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "p2"}}}

	// append nil list
	refs := tt.appendUnstructuredRefIfAbsent(nil, nil, nil, p1)
	require.Len(t, refs, 1)
	require.Same(t, p1, refs[0].Object)

	// append valid list
	refs = tt.appendUnstructuredRefIfAbsent(nil, nil, refs, p2)
	require.Len(t, refs, 2)
	require.Same(t, p1, refs[0].Object)
	require.Same(t, p2, refs[1].Object)

	// append objects that were already added
	refs = tt.appendUnstructuredRefIfAbsent(nil, nil, refs, p1)
	require.Len(t, refs, 2)
	refs = tt.appendUnstructuredRefIfAbsent(nil, nil, refs, p2)
	require.Len(t, refs, 2)

	// existence check if only done using pointers, adding a policy with the same name
	// but a different pointer should work
	p1Copy := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "p1"}}}
	refs = tt.appendUnstructuredRefIfAbsent(nil, nil, refs, p1Copy)
	require.Len(t, refs, 3)
	require.Same(t, p1Copy, refs[2].Object)
}

func TestGetOrCreateExtensionResource(t *testing.T) {
	newObj := func(name string) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "example.io/v1",
			"kind":       "Foo",
			"metadata":   map[string]any{"name": name, "namespace": "ns1"},
		}}
	}
	gatewayCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: "ns1", Name: "gw1"}}}

	t.Run("no gateway context embeds the object directly", func(t *testing.T) {
		tt := &Translator{}
		gwIR := &ir.Xds{}
		obj := newObj("obj1")

		ref := tt.getOrCreateExtensionResource(gwIR, nil, obj)
		require.Same(t, obj, ref.Object)
		require.Empty(t, ref.Name)
		require.Empty(t, gwIR.ExtensionResources)
	})

	t.Run("registers each distinct identity once", func(t *testing.T) {
		tt := &Translator{TranslatorContext: &TranslatorContext{}}
		gwIR := &ir.Xds{}
		obj1 := newObj("obj1")

		ref1 := tt.getOrCreateExtensionResource(gwIR, gatewayCtx, obj1)
		require.Nil(t, ref1.Object)
		require.NotEmpty(t, ref1.Name)
		require.Len(t, gwIR.ExtensionResources, 1)
		require.Same(t, obj1, gwIR.ExtensionResources[0].Object)
		require.Equal(t, ref1.Name, gwIR.ExtensionResources[0].Name)

		// A second, distinct *unstructured.Unstructured with the same GVK+namespace+name
		// identity reuses the existing registry entry rather than adding a new one, but
		// refreshes the canonical object to this newer copy rather than keeping the first
		// one seen (e.g. a copy later mutated with ExtensionServerPolicy status must not be
		// shadowed by an earlier, stale copy registered from route/filter processing).
		obj1Again := newObj("obj1")
		ref2 := tt.getOrCreateExtensionResource(gwIR, gatewayCtx, obj1Again)
		require.Equal(t, ref1.Name, ref2.Name)
		require.Nil(t, ref2.Object)
		require.Len(t, gwIR.ExtensionResources, 1)
		require.Same(t, obj1Again, gwIR.ExtensionResources[0].Object)

		// A resource with a different identity gets its own registry entry.
		ref3 := tt.getOrCreateExtensionResource(gwIR, gatewayCtx, newObj("obj2"))
		require.NotEqual(t, ref1.Name, ref3.Name)
		require.Len(t, gwIR.ExtensionResources, 2)
	})
}

func TestMergeAncestorsForExtensionServerPolicies(t *testing.T) {
	tests := []struct {
		aggStatus *gwapiv1.PolicyStatus
		newStatus *gwapiv1.PolicyStatus
		noStatus  bool
	}{
		{
			aggStatus: &gwapiv1.PolicyStatus{
				Ancestors: []gwapiv1.PolicyAncestorStatus{
					{
						AncestorRef: gwapiv1.ParentReference{
							Name: "gateway-1",
						},
					},
				},
			},
			newStatus: &gwapiv1.PolicyStatus{
				Ancestors: []gwapiv1.PolicyAncestorStatus{
					{
						AncestorRef: gwapiv1.ParentReference{
							Name: "gateway-2",
						},
					},
				},
			},
		},
		{
			aggStatus: &gwapiv1.PolicyStatus{},
			newStatus: &gwapiv1.PolicyStatus{
				Ancestors: []gwapiv1.PolicyAncestorStatus{
					{
						AncestorRef: gwapiv1.ParentReference{
							Name: "gateway-2",
						},
					},
				},
			},
		},
		{
			aggStatus: &gwapiv1.PolicyStatus{
				Ancestors: []gwapiv1.PolicyAncestorStatus{
					{
						AncestorRef: gwapiv1.ParentReference{
							Name: "gateway-1",
						},
					},
				},
			},
			newStatus: &gwapiv1.PolicyStatus{},
		},
		{
			aggStatus: &gwapiv1.PolicyStatus{},
			newStatus: &gwapiv1.PolicyStatus{},
		},
		{
			aggStatus: nil,
			newStatus: &gwapiv1.PolicyStatus{
				Ancestors: []gwapiv1.PolicyAncestorStatus{
					{
						AncestorRef: gwapiv1.ParentReference{
							Name: "gateway-1",
						},
					},
				},
			},
		},
		{
			aggStatus: &gwapiv1.PolicyStatus{
				Ancestors: []gwapiv1.PolicyAncestorStatus{
					{
						AncestorRef: gwapiv1.ParentReference{
							Name: "gateway-1",
						},
					},
				},
			},
			newStatus: nil,
		},
		{
			aggStatus: nil,
			newStatus: nil,
		},
	}

	for _, test := range tests {
		aggPolicy := unstructured.Unstructured{Object: make(map[string]interface{})}
		newPolicy := unstructured.Unstructured{Object: make(map[string]interface{})}
		desiredMergedStatus := gwapiv1.PolicyStatus{}

		// aggStatus == nil, means simulate not setting status at all within the policy.
		if test.aggStatus != nil {
			aggPolicy.Object["status"] = PolicyStatusToUnstructured(*test.aggStatus)
			desiredMergedStatus.Ancestors = append(desiredMergedStatus.Ancestors, test.aggStatus.Ancestors...)
		}

		// newStatus == nil, means simulate not setting status at all within the policy.
		if test.newStatus != nil {
			newPolicy.Object["status"] = PolicyStatusToUnstructured(*test.newStatus)
			desiredMergedStatus.Ancestors = append(desiredMergedStatus.Ancestors, test.newStatus.Ancestors...)
		}

		mergeAncestorsForExtensionServerPolicies(&aggPolicy, &newPolicy)

		// The product object will always have an existing `status`, even if with 0 ancestors.
		newAggPolicy := ExtServerPolicyStatusAsPolicyStatus(&aggPolicy)
		require.Len(t, newAggPolicy.Ancestors, len(desiredMergedStatus.Ancestors))
		for i := range newAggPolicy.Ancestors {
			require.Equal(t, desiredMergedStatus.Ancestors[i].AncestorRef.Name, newAggPolicy.Ancestors[i].AncestorRef.Name)
		}
	}
}

// Appends status ancestors from newPolicy into aggregatedPolicy's list of ancestors.
func mergeAncestorsForExtensionServerPolicies(aggregatedPolicy, newPolicy *unstructured.Unstructured) {
	aggStatus := ExtServerPolicyStatusAsPolicyStatus(aggregatedPolicy)
	newStatus := ExtServerPolicyStatusAsPolicyStatus(newPolicy)
	aggStatus.Ancestors = append(aggStatus.Ancestors, newStatus.Ancestors...)
	aggregatedPolicy.Object["status"] = PolicyStatusToUnstructured(aggStatus)
}
