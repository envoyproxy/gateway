// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
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

func TestGetSupportedFeatures(t *testing.T) {
	testCases := []struct {
		name           string
		gatewaySuite   suite.ConformanceOptions
		skippedTests   []suite.ConformanceTest
		expectedResult []gwapiv1.SupportedFeature
	}{
		{
			name: "No exempt features",
			gatewaySuite: suite.ConformanceOptions{
				SupportedFeatures: sets.New[features.SupportedFeature]("Gateway", "HTTPRoute"),
				ExemptFeatures:    sets.New[features.SupportedFeature](),
			},
			expectedResult: []gwapiv1.SupportedFeature{"Gateway", "HTTPRoute"},
		},
		{
			name: "All features exempt",
			gatewaySuite: suite.ConformanceOptions{
				SupportedFeatures: sets.New[features.SupportedFeature]("Gateway", "HTTPRoute"),
				ExemptFeatures:    sets.New[features.SupportedFeature]("Gateway", "HTTPRoute"),
			},
			expectedResult: []gwapiv1.SupportedFeature{},
		},
		{
			name: "Some features exempt",
			gatewaySuite: suite.ConformanceOptions{
				SupportedFeatures: sets.New[features.SupportedFeature]("Gateway", "HTTPRoute", "GRPCRoute"),
				ExemptFeatures:    sets.New[features.SupportedFeature]("GRPCRoute"),
			},
			expectedResult: []gwapiv1.SupportedFeature{"Gateway", "HTTPRoute"},
		},
		{
			name: "Some features exempt with skipped tests",
			gatewaySuite: suite.ConformanceOptions{
				SupportedFeatures: sets.New[features.SupportedFeature]("Gateway", "HTTPRoute", "GRPCRoute"),
				ExemptFeatures:    sets.New[features.SupportedFeature]("GRPCRoute"),
			},
			skippedTests: []suite.ConformanceTest{
				{
					Features: []features.SupportedFeature{"HTTPRoute"},
				},
			},
			expectedResult: []gwapiv1.SupportedFeature{"Gateway"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getSupportedFeatures(tc.gatewaySuite, tc.skippedTests)

			assert.ElementsMatch(t, tc.expectedResult, result, "The result should match the expected output for the test case.")
		})
	}
}
