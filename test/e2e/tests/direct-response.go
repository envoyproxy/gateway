// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
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

			// Test /inline with custom header
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{Path: "/inline"},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"Content-Type":    "text/plain",
						"X-Custom-Header": "custom-value",
					},
				},
				ResponseBody: "Oops! Your request is not found.",
				Namespace:    ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}
			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}

			verifyCustomResponse(t, suite.TimeoutConfig, gwAddr, "/value-ref", "application/json", `{"error": "Internal Server Error"}`, 200)
			verifyCustomResponse(t, suite.TimeoutConfig, gwAddr, "/401", "", ``, 401)
		})
	},
}
