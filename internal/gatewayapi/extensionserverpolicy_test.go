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
)

func TestExtractTargetRefs(t *testing.T) {
	tests := []struct {
		desc          string
		specInput     map[string]any
		output        []gwapiv1.LocalPolicyTargetReferenceWithSectionName
		expectedError string
	}{
		{
			desc:          "no spec",
			specInput:     nil,
			output:        nil,
			expectedError: "no targets found for the policy",
		},
		{
			desc: "no targetRef",
			specInput: map[string]any{
				"someAttr": "someValue",
			},
			output:        nil,
			expectedError: "no targets found for the policy",
		},
		{
			desc: "targetRefs is not an array",
			specInput: map[string]any{
				"targetRefs": "someValue",
			},
			output:        nil,
			expectedError: "no targets found for the policy",
		},
		{
			desc: "invalid targetref",
			specInput: map[string]any{
				"targetRef": map[string]any{
					"someKey": "someValue",
				},
			},
			output:        nil,
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
			output: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
						Group: "some.group",
						Kind:  "SomeKind",
						Name:  "name",
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
			output: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
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
	}

	for _, currTest := range tests {
		t.Run(currTest.desc, func(t *testing.T) {
			policy := &unstructured.Unstructured{
				Object: map[string]any{},
			}
			policy.Object["spec"] = currTest.specInput
			targets, err := extractTargetRefs(policy, []*GatewayContext{})

			if currTest.expectedError != "" {
				require.EqualError(t, err, currTest.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, currTest.output, targets)
			}
		})
	}
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

		MergeAncestorsForExtensionServerPolicies(&aggPolicy, &newPolicy)

		// The product object will always have an existing `status`, even if with 0 ancestors.
		newAggPolicy := ExtServerPolicyStatusAsPolicyStatus(&aggPolicy)
		require.Len(t, newAggPolicy.Ancestors, len(desiredMergedStatus.Ancestors))
		for i := range newAggPolicy.Ancestors {
			require.Equal(t, desiredMergedStatus.Ancestors[i].AncestorRef.Name, newAggPolicy.Ancestors[i].AncestorRef.Name)
		}
	}
}
