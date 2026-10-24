// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// TestSetConditionForPolicyAncestorBounded verifies the ancestor list is capped on the fly as
// entries are added (see #9539) and that the highest-priority (not-accepted) ancestors are kept.
func TestSetConditionForPolicyAncestorBounded(t *testing.T) {
	const controller = "example.com/controller"
	policyStatus := &gwapiv1.PolicyStatus{}

	// A few not-accepted (rank 0) ancestors, then many accepted (rank 2) ones.
	badNames := []string{"bad-0", "bad-1", "bad-2"}
	for _, name := range badNames {
		ref := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName(name)}
		SetConditionForPolicyAncestor(policyStatus, ref, controller,
			gwapiv1.PolicyConditionAccepted, metav1.ConditionFalse, gwapiv1.PolicyReasonConflicted, "conflicted", 1)
	}
	for i := range 100 {
		ref := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName(fmt.Sprintf("gw-%03d", i))}
		SetConditionForPolicyAncestor(policyStatus, ref, controller,
			gwapiv1.PolicyConditionAccepted, metav1.ConditionTrue, gwapiv1.PolicyReasonAccepted, "accepted", 1)
	}

	// Bounded on the fly at the soft cap, never growing to 103.
	assert.Len(t, policyStatus.Ancestors, policyAncestorsSoftCap)
	// The not-accepted (rank 0) ancestors are preserved; only rank-2 entries are evicted.
	for _, name := range badNames {
		found := false
		for _, a := range policyStatus.Ancestors {
			if string(a.AncestorRef.Name) == name {
				found = true
			}
		}
		assert.True(t, found, "not-accepted ancestor %q must be preserved", name)
	}

	// Final truncation still cuts to the CRD limit.
	TruncatePolicyAncestors(policyStatus, controller, 1)
	assert.Len(t, policyStatus.Ancestors, maxPolicyAncestors)
}

func BenchmarkSetConditionForPolicyAncestorsAtScale(b *testing.B) {
	const controller = "example.com/controller"
	for b.Loop() {
		policyStatus := &gwapiv1.PolicyStatus{}
		for i := range 10000 {
			ref := &gwapiv1.ParentReference{Name: gwapiv1.ObjectName(fmt.Sprintf("gw-%05d", i))}
			SetConditionForPolicyAncestor(policyStatus, ref, controller,
				gwapiv1.PolicyConditionAccepted, metav1.ConditionTrue, gwapiv1.PolicyReasonAccepted, "accepted", 1)
		}
	}
}

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

func TestSetNotMergedForPolicyAncestor(t *testing.T) {
	const controller = "example.com/controller"
	ref := &gwapiv1.ParentReference{Name: "gw-1"}

	tests := []struct {
		name     string
		existing *metav1.Condition
		want     []string
	}{
		{
			name:     "no Accepted condition is not accepted",
			existing: nil,
			want: []string{
				"Accepted/False/NotMerged",
			},
		},
		{
			name: "existing Accepted=True is left alone",
			existing: &metav1.Condition{
				Type:   string(gwapiv1.PolicyConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.PolicyReasonAccepted),
			},
			want: []string{
				"Accepted/True/Accepted",
			},
		},
		{
			name: "existing Accepted=False error is left alone",
			existing: &metav1.Condition{
				Type:   string(gwapiv1.PolicyConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: string(gwapiv1.PolicyReasonInvalid),
			},
			want: []string{
				"Accepted/False/Invalid",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policyStatus := &gwapiv1.PolicyStatus{}
			if tc.existing != nil {
				SetConditionForPolicyAncestor(policyStatus, ref, controller,
					gwapiv1.PolicyConditionType(tc.existing.Type), tc.existing.Status, gwapiv1.PolicyConditionReason(tc.existing.Reason), "existing", 1)
			}

			SetNotMergedForPolicyAncestor(policyStatus, ref, controller, 1, "not merged")

			require.Len(t, policyStatus.Ancestors, 1)
			conditions := policyStatus.Ancestors[0].Conditions
			got := make([]string, 0, len(conditions))
			for _, c := range conditions {
				got = append(got, fmt.Sprintf("%s/%s/%s", c.Type, c.Status, c.Reason))
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
