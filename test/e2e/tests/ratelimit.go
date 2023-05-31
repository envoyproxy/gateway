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
	ConformanceTests = append(ConformanceTests, RateLimitTest)
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

			// should just send exactly 4 requests, and expect 429
			firstThreeExpResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			firstThreeReq := http.MakeRequest(t, &firstThreeExpResp, gwAddr, "HTTP", "http")
			if err := GotNTimesExpectedResponse(t, 3, suite.RoundTripper, firstThreeReq, firstThreeExpResp); err != nil {
				t.Errorf("fail to get expected response at first three request: %v", err)
			}

			lastFourthExpResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 429,
				},
				Namespace: ns,
			}
			lastFourthReq := http.MakeRequest(t, &lastFourthExpResp, gwAddr, "HTTP", "http")
			if err := GotNTimesExpectedResponse(t, 1, suite.RoundTripper, lastFourthReq, lastFourthExpResp); err != nil {
				t.Errorf("fail to get expected response at last fourth request: %v", err)
			}
		})
	},
}

func GotNTimesExpectedResponse(t *testing.T, n int, r roundtripper.RoundTripper, req roundtripper.Request, resp http.ExpectedResponse) error {
	for i := 0; i < n; i++ {
		cReq, cRes, err := r.CaptureRoundTrip(req)
		if err != nil {
			return err
		}

		if err = http.CompareRequest(t, &req, cReq, cRes, resp); err != nil {
			return err
		}
	}
	return nil
}
