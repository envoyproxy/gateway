// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
	p1 := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "p1"}}}
	p2 := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "p2"}}}

	// append nil list
	refs := appendUnstructuredRefIfAbsent(nil, p1)
	require.Len(t, refs, 1)
	require.Same(t, p1, refs[0].Object)

	// append valid list
	refs = appendUnstructuredRefIfAbsent(refs, p2)
	require.Len(t, refs, 2)
	require.Same(t, p1, refs[0].Object)
	require.Same(t, p2, refs[1].Object)

	// append objects that were already added
	refs = appendUnstructuredRefIfAbsent(refs, p1)
	require.Len(t, refs, 2)
	refs = appendUnstructuredRefIfAbsent(refs, p2)
	require.Len(t, refs, 2)

	// existence check if only done using pointers, adding a policy with the same name
	// but a different pointer should work
	p1Copy := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "p1"}}}
	refs = appendUnstructuredRefIfAbsent(refs, p1Copy)
	require.Len(t, refs, 3)
	require.Same(t, p1Copy, refs[2].Object)
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
