// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/pkg/features"
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

func TestSetGatewayClassAcceptedSupportedFeatures(t *testing.T) {
	for _, accepted := range []bool{true, false} {
		gc := &gwapiv1.GatewayClass{}
		SetGatewayClassAccepted(gc, accepted, string(gwapiv1.GatewayClassReasonAccepted), MsgValidGatewayClass)

		got := gc.Status.SupportedFeatures
		require.NotEmpty(t, got)
		// The CRD limits the list to 64 entries and requires it to be sorted by name.
		assert.LessOrEqual(t, len(got), 64)
		assert.True(t, slices.IsSortedFunc(got, func(a, b gwapiv1.SupportedFeature) int {
			return strings.Compare(string(a.Name), string(b.Name))
		}))

		names := sets.New[string]()
		for _, f := range got {
			names.Insert(string(f.Name))
		}
		assert.Len(t, names, len(got), "duplicate feature")
		for _, supported := range []features.FeatureName{features.SupportHTTPRoute, features.SupportTCPRoute, features.SupportUDPRoute} {
			assert.True(t, names.Has(string(supported)), "%s is supported", supported)
		}
		assert.False(t, names.Has(string(features.SupportMesh)), "mesh features are not supported")
		for _, unsupported := range unsupportedFeatures {
			assert.False(t, names.Has(string(unsupported)), "%s is unsupported", unsupported)
		}
	}
}
