// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, RegexRedirectTest)
}

var RegexRedirectTest = suite.ConformanceTest{
	ShortName:   "RegexRedirect",
	Description: "Regex redirect paths compose with native redirects in either filter order",
	Manifests:   []string{"testdata/regex-redirect.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		route := types.NamespacedName{Name: "regex-redirect", Namespace: ns}
		gateway := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		address := kubernetes.GatewayAndRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName, kubernetes.NewGatewayRef(gateway), &gwapiv1.HTTPRoute{}, false, route)
		for _, tc := range []struct {
			path, location string
			code           int
		}{
			{"/blogs/123", "https://example.com/post-123", 301},
			{"/blogs/123?utm_source=email&x=1", "https://example.com/post-123?utm_source=email&x=1", 301},
			{"/blogs/not-a-number", "https://example.com/blogs/not-a-number", 301},
			{"/blogs/123/", "https://example.com/blogs/123/", 301},
			{"/articles/456", "https://example.com:8443/post-456", 302},
			{"/articles/456?x=%2F", "https://example.com:8443/post-456?x=%2F", 302},
		} {
			t.Run(tc.path, func(t *testing.T) {
				expected := http.ExpectedResponse{
					Request:   http.Request{Path: tc.path, UnfollowRedirect: true},
					Response:  http.Response{StatusCode: tc.code, Headers: map[string]string{"Location": tc.location}},
					Namespace: ns,
				}
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, address, expected)
			})
		}
	},
}
