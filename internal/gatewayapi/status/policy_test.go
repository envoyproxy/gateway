// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestSetConditionForPolicyAncestorsTruncatesMessages(t *testing.T) {
	longMsg := strings.Repeat("x", conditionMessageMaxLength+5)
	policyStatus := &gwapiv1.PolicyStatus{}
	ancestorRef := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName("example")}

	SetConditionForPolicyAncestors(policyStatus, []*gwapiv1.ParentReference{ancestorRef}, "example.com/controller",
		gwapiv1.PolicyConditionAccepted, metav1.ConditionTrue, gwapiv1.PolicyReasonAccepted, longMsg, 1)

	if assert.Len(t, policyStatus.Ancestors, 1) {
		ancestor := policyStatus.Ancestors[0]
		if assert.Len(t, ancestor.Conditions, 1) {
			gotMsg := ancestor.Conditions[0].Message
			assert.Len(t, gotMsg, conditionMessageMaxLength)
			prefixLen := conditionMessageMaxLength - len(conditionMessageTruncationSuffix)
			expectedPrefix := strings.Repeat("x", prefixLen)
			assert.True(t, strings.HasSuffix(gotMsg, conditionMessageTruncationSuffix))
			assert.Equal(t, expectedPrefix, gotMsg[:prefixLen])
		}
	}
}

func TestBuildDeprecationWarningMessage(t *testing.T) {
	tests := []struct {
		name             string
		deprecatedFields map[string]string
		expected         string
	}{
		{
			name:             "empty map",
			deprecatedFields: map[string]string{},
			expected:         "",
		},
		{
			name: "two entries with deterministic ordering",
			deprecatedFields: map[string]string{
				"spec.targetRef":   "spec.targetRefs",
				"spec.compression": "spec.compressor",
			},
			expected: "spec.compression is deprecated, use spec.compressor instead; spec.targetRef is deprecated, use spec.targetRefs instead",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDeprecationWarningMessage(tt.deprecatedFields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetWarningForPolicyAncestorMergesWarnings(t *testing.T) {
	policyStatus := &gwapiv1.PolicyStatus{}
	ancestorRef := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName("example")}

	SetWarningForPolicyAncestor(policyStatus, ancestorRef, "example.com/controller",
		egv1a1.PolicyReasonDeprecatedField, "deprecated field warning", 1)
	SetWarningForPolicyAncestor(policyStatus, ancestorRef, "example.com/controller",
		PolicyReasonUnsupportedHTTP3ClientValidation, "http3 warning", 1)

	if assert.Len(t, policyStatus.Ancestors, 1) {
		conditions := policyStatus.Ancestors[0].Conditions
		if assert.Len(t, conditions, 1) {
			assert.Equal(t, string(egv1a1.PolicyConditionWarning), conditions[0].Type)
			assert.Equal(t, string(PolicyReasonMultipleWarnings), conditions[0].Reason)
			assert.Equal(t, "deprecated field warning; http3 warning", conditions[0].Message)
		}
	}
}
