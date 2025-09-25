// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, DirectResponseTest)
}

var DirectResponseTest = suite.ConformanceTest{
	ShortName:   "DirectResponse",
	Description: "Direct",
	Manifests:   []string{"testdata/direct-response.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("direct response", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "direct-response", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

			// Test inline response with add and set headers
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, "/inline", "text/plain", "Oops! Your request is not found.", 200, map[string]string{
				"X-Add-Header":    "added-value",
				"X-Set-Header":    "set-value",
				"X-Response-Type": "direct",
				"Cache-Control":   "no-cache",
			})

			// Test value-ref response with add and set headers
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, "/value-ref", "application/json", `{"error": "Internal Server Error"}`, 200, map[string]string{
				"X-Add-Header": "added-json",
				"X-Set-Header": "set-json",
			})

			// Test status-only response with add and set headers
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, "/401", "", ``, 401, map[string]string{
				"X-Add-Header": "added-401",
				"X-Set-Header": "set-401",
			})
		})
	},
}
