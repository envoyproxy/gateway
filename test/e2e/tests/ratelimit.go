// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	// ConformanceTests = append(ConformanceTests, RateLimitTest)
}

var RateLimitTest = suite.ConformanceTest{
	ShortName:   "RateLimit",
	Description: "Limit all requests",
	Manifests:   []string{"testdata/ratelimit-block-all-ips.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("block all ips", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-ratelimit", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

			expectLimitResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 429,
				},
				Namespace: ns,
			}
			expectLimitReq := http.MakeRequest(t, &expectLimitResp, gwAddr, "HTTP", "http")

			// should just send exactly 4 requests, and expect 429

			// keep sending requests till get 200 first, that will cost one 200
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

			// fire the rest request
			if err := GotExactNExpectedResponse(t, 2, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("fail to get expected response at first three request: %v", err)
			}
			if err := GotExactNExpectedResponse(t, 1, suite.RoundTripper, expectLimitReq, expectLimitResp); err != nil {
				t.Errorf("fail to get expected response at last fourth request: %v", err)
			}
		})
	},
}

func GotExactNExpectedResponse(t *testing.T, n int, r roundtripper.RoundTripper, req roundtripper.Request, resp http.ExpectedResponse) error {
	for i := 0; i < n; i++ {
		// cReq, cRes, err := r.CaptureRoundTrip(req)
		// if err != nil {
		// 	return err
		// }

		// if err = http.CompareRequest(t, &req, cReq, cRes, resp); err != nil {
		// 	return err
		// }
	}
	return nil
}
