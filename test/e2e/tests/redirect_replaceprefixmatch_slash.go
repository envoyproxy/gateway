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
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, RedirectTrailingSlashTest)

}

// RedirectTrailingSlashTest tests that only one slash in the redirect URL
// See https://github.com/envoyproxy/gateway/issues/2976
var RedirectTrailingSlashTest = suite.ConformanceTest{
	ShortName:   "RedirectTrailingSlash",
	Description: "Test that only one slash in the redirect URL",
	Manifests:   []string{"testdata/redirect-replaceprefixmatch-slash.yaml"},

	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		testCases := []struct {
			name             string
			path             string
			statusCode       int
			expectedLocation string
		}{
			// Test cases for the HTTPRoute match /api/foo/
			{
				name:             "match: /api/foo/, request: /api/foo/redirect",
				path:             "/api/foo/redirect",
				statusCode:       302,
				expectedLocation: "/redirect",
			},
			{
				name:             "match: /api/foo/, request: /api/foo/",
				path:             "/api/foo/",
				statusCode:       302,
				expectedLocation: "/",
			},
			{
				name:             "match: /api/foo/, request: /api/foo",
				path:             "/api/foo",
				statusCode:       302,
				expectedLocation: "/",
			},
			{
				name:       "match: /api/foo/, request: /api/foo-bar",
				path:       "/api/foo-bar",
				statusCode: 404,
			},

			// Test cases for the HTTPRoute match /api/bar
			{
				name:             "match: /api/bar, request: /api/bar/redirect",
				path:             "/api/bar/redirect",
				statusCode:       302,
				expectedLocation: "/redirect",
			},
			{
				name:             "match: /api/bar, request: /api/bar/",
				path:             "/api/bar/",
				statusCode:       302,
				expectedLocation: "/",
			},
			{
				name:             "match: /api/bar, request: /api/bar",
				path:             "/api/bar",
				statusCode:       302,
				expectedLocation: "/",
			},
			{
				name:       "match: /api/bar, request: /api/bar-foo",
				path:       "/api/bar-foo",
				statusCode: 404,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				ns := "gateway-conformance-infra"
				routeNN := types.NamespacedName{Name: "redirect-replaceprefixmatch-slash", Namespace: ns}
				gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
				gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path:             testCase.path,
						UnfollowRedirect: true,
					},
					Response: http.Response{
						StatusCode: testCase.statusCode,
					},
					Namespace: ns,
				}
				if testCase.expectedLocation != "" {
					expectedResponse.Response.Headers = map[string]string{
						"Location": testCase.expectedLocation,
					}
				}

				req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Errorf("failed to get expected response: %v", err)
				}

				if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
					t.Errorf("failed to compare request and response: %v", err)
				}
			})
		}
	},
}
