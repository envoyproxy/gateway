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
