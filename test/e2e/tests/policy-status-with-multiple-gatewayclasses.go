// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func init() {
	ConformanceTests = append(ConformanceTests, PolicyStatusWithMultipleGatewayClassesTest)
}

// PolicyStatusWithMultipleGatewayClassesTest tests the policy status contains all the Ancestors when
// targeting resources from multiple GatewayClasses.
var PolicyStatusWithMultipleGatewayClassesTest = suite.ConformanceTest{
	ShortName:   "PolicyStatusWithMultipleGatewayClasses",
	Description: "Test policy status contains all the Ancestors when targeting resources from multiple GatewayClasses",
	Manifests:   []string{"testdata/policy-status-with-multiple-gatewayclasses.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("BackendTrafficPolicy targets HTTPRoutes that are associated with Gateways from two different GatewayClasses", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			_ = kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(types.NamespacedName{Name: "gateway-1", Namespace: ns}),
				types.NamespacedName{Name: "httproute-1", Namespace: ns})
			_ = kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(types.NamespacedName{Name: "gateway-2", Namespace: ns}),
				types.NamespacedName{Name: "httproute-2", Namespace: ns})

			// BackendTrafficPolicy should have 2 ancestors from both GatewayClasses
			require.Eventually(t, func() bool {
				btp := &egv1a1.BackendTrafficPolicy{}
				err := suite.Client.Get(
					context.Background(),
					types.NamespacedName{Name: "btp-target-multiple-gateway-classes-success", Namespace: ns}, btp)
				return err == nil && len(btp.Status.Ancestors) == 2
			}, suite.TimeoutConfig.GetTimeout, time.Second)
			require.Eventually(t, func() bool {
				btp := &egv1a1.BackendTrafficPolicy{}
				err := suite.Client.Get(context.Background(), types.NamespacedName{Name: "btp-target-multiple-gateway-classes-failure", Namespace: ns}, btp)
				return err == nil && len(btp.Status.Ancestors) == 2
			}, suite.TimeoutConfig.GetTimeout, time.Second)
		})
	},
}
