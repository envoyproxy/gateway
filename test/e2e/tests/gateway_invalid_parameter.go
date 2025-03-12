// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayInvalidParameterTest)
}

var GatewayInvalidParameterTest = suite.ConformanceTest{
	ShortName:   "GatewayInvalidParameterTest",
	Description: "Gateway with invalid parameters should not be accepted",
	Manifests:   []string{"testdata/gateway-invalid-parameter.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Positive", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gateway-invalid-parameter", Namespace: "gateway-conformance-infra"}
			kubernetes.GatewayMustHaveLatestConditions(
				t,
				suite.Client,
				suite.TimeoutConfig,
				gwNN)
			kubernetes.GatewayMustHaveCondition(
				t,
				suite.Client,
				suite.TimeoutConfig,
				gwNN,
				metav1.Condition{
					Type:   string(gwapiv1.GatewayConditionAccepted),
					Status: metav1.ConditionFalse,
					Reason: string(gwapiv1.GatewayReasonInvalidParameters),
				})
			gw := &gwapiv1.Gateway{}
			err := suite.Client.Get(context.Background(), gwNN, gw)
			if err != nil {
				t.Fatalf("Failed to get gateway: %v", err)
			}
			require.Empty(t, gw.Status.Addresses)
			require.Empty(t, gw.Status.Listeners)
		})
	},
}
