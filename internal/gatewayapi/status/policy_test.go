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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestSetConditionForPolicyAncestorsTruncatesMessages(t *testing.T) {
	longMsg := strings.Repeat("x", conditionMessageMaxLength+5)
	policyStatus := &gwapiv1a2.PolicyStatus{}
	ancestorRef := &gwapiv1a2.ParentReference{Name: gwapiv1a2.ObjectName("example")}

	SetConditionForPolicyAncestors(policyStatus, []*gwapiv1a2.ParentReference{ancestorRef}, "example.com/controller",
		gwapiv1a2.PolicyConditionAccepted, metav1.ConditionTrue, gwapiv1a2.PolicyReasonAccepted, longMsg, 1)

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
