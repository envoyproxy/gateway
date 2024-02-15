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
	ConformanceTests = append(ConformanceTests, RateLimitCIDRMatchTest)
	ConformanceTests = append(ConformanceTests, RateLimitHeaderMatchTest)
	ConformanceTests = append(ConformanceTests, RateLimitBasedJwtClaimsTest)
}

var RateLimitCIDRMatchTest = suite.ConformanceTest{
	ShortName:   "RateLimitCIDRMatch",
	Description: "Limit all requests that match CIDR",
	Manifests:   []string{"testdata/ratelimit-cidr-match.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("block all ips", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "cidr-ratelimit", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			ratelimitHeader := make(map[string]string)
			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
					Headers:    ratelimitHeader,
				},
				Namespace: ns,
			}
			expectOkResp.Response.Headers["X-Ratelimit-Limit"] = "3, 3;w=3600"
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

			// fire the rest of requests
			if err := GotExactExpectedResponse(t, 2, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("failed to get expected response for the first three requests: %v", err)
			}
			if err := GotExactExpectedResponse(t, 1, suite.RoundTripper, expectLimitReq, expectLimitResp); err != nil {
				t.Errorf("failed to get expected response for the last (fourth) request: %v", err)
			}
		})
	},
}

var RateLimitHeaderMatchTest = suite.ConformanceTest{
	ShortName:   "RateLimitHeaderMatch",
	Description: "Limit all requests that match headers",
	Manifests:   []string{"testdata/ratelimit-header-match.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "header-ratelimit", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("all matched headers can got limited", func(t *testing.T) {
			requestHeaders := map[string]string{
				"x-user-id":  "one",
				"x-user-org": "acme",
			}

			ratelimitHeader := make(map[string]string)
			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path:    "/get",
					Headers: requestHeaders,
				},
				Response: http.Response{
					StatusCode: 200,
					Headers:    ratelimitHeader,
				},
				Namespace: ns,
			}
			expectOkResp.Response.Headers["X-Ratelimit-Limit"] = "3, 3;w=3600"
			expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

			expectLimitResp := http.ExpectedResponse{
				Request: http.Request{
					Path:    "/get",
					Headers: requestHeaders,
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

			// fire the rest of the requests
			if err := GotExactExpectedResponse(t, 2, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("failed to get expected response for the first three requests: %v", err)
			}
			if err := GotExactExpectedResponse(t, 1, suite.RoundTripper, expectLimitReq, expectLimitResp); err != nil {
				t.Errorf("failed to get expected response for the last (fourth) request: %v", err)
			}
		})

		t.Run("only one matched header cannot got limited", func(t *testing.T) {
			requestHeaders := map[string]string{
				"x-user-id": "one",
			}

			// it does not require any rate limit header, since this request never be rate limited.
			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path:    "/get",
					Headers: requestHeaders,
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

			// send exactly 4 requests, and still expect 200

			// keep sending requests till get 200 first, that will cost one 200
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

			// fire the rest of the requests
			if err := GotExactExpectedResponse(t, 3, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("failed to get expected responses for the request: %v", err)
			}
		})
	},
}

var RateLimitBasedJwtClaimsTest = suite.ConformanceTest{
	ShortName:   "RateLimitBasedJwtClaims",
	Description: "Limit based jwt claims",
	Manifests:   []string{"testdata/ratelimit-based-jwt-claims.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("ratelimit based on jwt claims", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-ratelimit-based-jwt-claims", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/foo",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			expectLimitResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/foo",
				},
				Response: http.Response{
					StatusCode: 429,
				},
				Namespace: ns,
			}

			// Just to construct the request that carries a jwt token that can be limited
			ratelimitHeader := make(map[string]string)
			TokenHeader := make(map[string]string)
			JwtOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path:    "/foo",
					Headers: TokenHeader,
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/foo",
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers:    ratelimitHeader,
				},
				Namespace: ns,
			}
			JwtOkResp.Request.Headers["Authorization"] = "Bearer " + "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.NHVaYe26MbtOYhSKkoKYdFVomg4i8ZJd8_-RU8VNbftc4TSMb4bXP3l3YlNWACwyXPGffz5aXHc6lty1Y2t4SWRqGteragsVdZufDn5BlnJl9pdR_kdVFUsra2rWKEofkZeIC4yWytE58sMIihvo9H1ScmmVwBcQP6XETqYd0aSHp1gOa9RdUPDvoXQ5oqygTqVtxaDr6wUFKrKItgBMzWIdNZ6y7O9E0DhEPTbE9rfBo6KTFsHAZnMg4k68CDp2woYIaXbmYTWcvbzIuHO7_37GT79XdIwkm95QJ7hYC9RiwrV7mesbY4PAahERJawntho0my942XheVLmGwLMBkQ"
			JwtOkResp.Response.Headers["X-Ratelimit-Limit"] = "3, 3;w=3600"

			JwtReq := http.MakeRequest(t, &JwtOkResp, gwAddr, "HTTP", "http")

			// Just to construct the request that carries a jwt token that can not be limited
			DifTokenHeader := make(map[string]string)
			difJwtOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path:    "/foo",
					Headers: DifTokenHeader,
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			difJwtOkResp.Request.Headers["Authorization"] = "Bearer " + "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRvbSIsImFkbWluIjp0cnVlLCJpYXQiOjE1MTYyMzkwMjJ9.kyzDDSo7XpweSPU1lxoI9IHzhTBrRNlnmcW9lmCbloZELShg-8isBx4AFoM4unXZTHpS_Y24y0gmd4nDQxgUE-CgjVSnGCb0Xhy3WO1gm9iChoKDyyQ3kHp98EmKxTyxKG2X9GyKcDFNBDjH12OBD7TcJUaBEvLf6Jw1SG2A7FakUPWeK04DQ916-ROylzI6qKyaZ0OpfYIbijvyAQxlQRxxs2XHlAkLdJhfVcUqJBwsFTbwHYARC-WNgd2_etAk1GWdwwZ_NoTmRzZAMryrYJpHY9KPlbnZ93Ye3o9h2viBQ_XRb7JBkWnAGYO4_KswpJWE_7ROUVj8iOJo2jfY6w"

			difJwtReq := http.MakeRequest(t, &difJwtOkResp, gwAddr, "HTTP", "http")

			// make sure the gateway is available
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/bar",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// keep sending requests till get 200 first to make sure the gateway is available
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

			// should just send exactly 4 requests, and expect 429

			// keep sending requests till get 200 first, that will cost one 200
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, JwtOkResp)

			// fire the rest of requests
			if err := GotExactExpectedResponse(t, 2, suite.RoundTripper, JwtReq, JwtOkResp); err != nil {
				t.Errorf("failed to get expected response at third request: %v", err)
			}
			if err := GotExactExpectedResponse(t, 1, suite.RoundTripper, JwtReq, expectLimitResp); err != nil {
				t.Errorf("failed to get expected response at the fourth request: %v", err)
			}

			// Carrying different jwt claims will not be limited
			if err := GotExactExpectedResponse(t, 4, suite.RoundTripper, difJwtReq, expectOkResp); err != nil {
				t.Errorf("failed to get expected response for the request with a different jwt: %v", err)
			}

			// make sure the request with no token is rejected
			noTokenResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/foo",
				},
				Response: http.Response{
					StatusCode: 401,
				},
				Namespace: ns,
			}
			noTokenReq := http.MakeRequest(t, &noTokenResp, gwAddr, "HTTP", "http")
			if err := GotExactExpectedResponse(t, 1, suite.RoundTripper, noTokenReq, noTokenResp); err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

		})
	},
}

func GotExactExpectedResponse(t *testing.T, n int, r roundtripper.RoundTripper, req roundtripper.Request, resp http.ExpectedResponse) error {
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
