package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestExtractTargetRefs(t *testing.T) {

	tests := []struct {
		desc          string
		specInput     map[string]any
		output        []gwv1a2.LocalPolicyTargetReferenceWithSectionName
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
			expectedError: "targetRefs is not an array",
		},
		{
			desc: "invalid targetref",
			specInput: map[string]any{
				"targetRef": map[string]any{
					"someKey": "someValue",
				},
			},
			output:        nil,
			expectedError: "invalid targetRef found: {\"someKey\":\"someValue\"}",
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
			output: []gwv1a2.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: gwv1a2.LocalPolicyTargetReference{
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
			output: []gwv1a2.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: gwv1a2.LocalPolicyTargetReference{
						Group: "some.group",
						Kind:  "SomeKind2",
						Name:  "othername",
					},
				},
				{
					LocalPolicyTargetReference: gwv1a2.LocalPolicyTargetReference{
						Group: "some.group",
						Kind:  "SomeKind",
						Name:  "name",
					},
				},
			},
		},
		{
			desc: "valid multiple targetRefs and targetRef",
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
				"targetRef": map[string]any{
					"group":       "some.group",
					"kind":        "SomeKind",
					"name":        "three",
					"sectionName": "one",
				},
			},
			output: []gwv1a2.LocalPolicyTargetReferenceWithSectionName{
				{
					LocalPolicyTargetReference: gwv1a2.LocalPolicyTargetReference{
						Group: "some.group",
						Kind:  "SomeKind2",
						Name:  "othername",
					},
				},
				{
					LocalPolicyTargetReference: gwv1a2.LocalPolicyTargetReference{
						Group: "some.group",
						Kind:  "SomeKind",
						Name:  "name",
					},
				},
				{
					LocalPolicyTargetReference: gwv1a2.LocalPolicyTargetReference{
						Group: "some.group",
						Kind:  "SomeKind",
						Name:  "three",
					},
					SectionName: ptr.To(gwv1a2.SectionName("one")),
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
				require.Nil(t, err)
				require.Equal(t, currTest.output, targets)
			}
		})
	}

}
