// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestComputeGatewayClassAcceptedCondition(t *testing.T) {
	testCases := []struct {
		name     string
		accepted bool
		expect   metav1.Condition
	}{
		{
			name:     "accepted gatewayclass",
			accepted: true,
			expect: metav1.Condition{
				Type:    string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:  metav1.ConditionTrue,
				Reason:  string(gwapiv1.GatewayClassReasonAccepted),
				Message: MsgValidGatewayClass,
			},
		},
		{
			name:     "not accepted gatewayclass",
			accepted: false,
			expect: metav1.Condition{
				Type:    string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:  metav1.ConditionFalse,
				Reason:  string(ReasonOlderGatewayClassExists),
				Message: MsgOlderGatewayClassExists,
			},
		},
		{
			name:     "invalid parameters gatewayclass",
			accepted: false,
			expect: metav1.Condition{
				Type:    string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:  metav1.ConditionFalse,
				Reason:  string(gwapiv1.GatewayClassReasonInvalidParameters),
				Message: MsgGatewayClassInvalidParams,
			},
		},
	}

	for _, tc := range testCases {
		gc := &gwapiv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 7,
			},
		}

		got := computeGatewayClassAcceptedCondition(gc, tc.accepted, tc.expect.Reason, tc.expect.Message)

		assert.Equal(t, tc.expect.Type, got.Type)
		assert.Equal(t, tc.expect.Status, got.Status)
		assert.Equal(t, tc.expect.Reason, got.Reason)
		assert.Equal(t, gc.Generation, got.ObservedGeneration)
	}
}
