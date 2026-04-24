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
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteHTTPSListenerDetectMisdirectedRequests)
}

var HTTPRouteHTTPSListenerDetectMisdirectedRequests = suite.ConformanceTest{
	ShortName:   "HTTPRouteHTTPSListenerDetectMisdirectedRequests",
	Description: "This's similar to the one(same test name) in upstream conformance tests, but use HTTP/1.1 instead of HTTP/2 to verify the behavior is not broken.",
	Features: []features.FeatureName{
		features.SupportGateway,
	},
	Manifests: []string{"testdata/httproute-https-listener-detect-misdirected-requests.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		routeNNs := []types.NamespacedName{
			{Name: "https-listener-detect-misdirected-requests-test-1", Namespace: ns},
			{Name: "https-listener-detect-misdirected-requests-test-2", Namespace: ns},
			{Name: "https-listener-detect-misdirected-requests-test-3", Namespace: ns},
			{Name: "https-listener-detect-misdirected-requests-test-4", Namespace: ns},
		}
		gwNN := types.NamespacedName{Name: "same-namespace-with-https-listener", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNNs...)
		for _, routeNN := range routeNNs {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
		}

		certNN := types.NamespacedName{Name: "tls-validity-checks-certificate", Namespace: ns}
		serverCertPem, _, _, err := GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}

		cases := []struct {
			host       string
			statusCode int
			backend    string
			serverName string
		}{
			{serverName: "example.org", host: "example.org", statusCode: 200, backend: "infra-backend-v1"},
			{serverName: "example.org", host: "second-example.org", statusCode: 404},
			{serverName: "example.org", host: "unknown-example.org", statusCode: 404},

			{serverName: "second-example.org", host: "second-example.org", statusCode: 200, backend: "infra-backend-v2"},
			{serverName: "second-example.org", host: "example.org", statusCode: 404},
			{serverName: "second-example.org", host: "unknown-example.org", statusCode: 404},

			{serverName: "third-example.wildcard.org", host: "third-example.wildcard.org", statusCode: 200, backend: "infra-backend-v3"},
			{serverName: "third-example.wildcard.org", host: "fith-example.wildcard.org", statusCode: 200, backend: "infra-backend-v3"},
			{serverName: "third-example.wildcard.org", host: "fourth-example.wildcard.org", statusCode: 200, backend: "infra-backend-v3"},
			{serverName: "third-example.wildcard.org", host: "second-example.org", statusCode: 404},
			{serverName: "third-example.wildcard.org", host: "unknown-example.org", statusCode: 404},

			// Note: Since infra-backend-v4 does not exist, infra-backend-v1 is reused for the fourth HTTPRoute
			{serverName: "fourth-example.wildcard.org", host: "fourth-example.wildcard.org", statusCode: 200, backend: "infra-backend-v1"},
			{serverName: "fourth-example.wildcard.org", host: "fith-example.wildcard.org", statusCode: 404},

			{serverName: "unknown-example.org", host: "example.org", statusCode: 200, backend: "infra-backend-v1"},
			{serverName: "unknown-example.org", host: "unknown-example.org", statusCode: 404},
		}

		for i, tc := range cases {
			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: tc.host,
					Path: "/detect-misdirected-requests",
				},
				Response:  http.Response{StatusCodes: []int{tc.statusCode}},
				Backend:   tc.backend,
				Namespace: "gateway-conformance-infra",
			}
			t.Run(expected.GetTestCaseName(i), func(t *testing.T) {
				tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, serverCertPem, nil, nil, tc.serverName, expected)
			})
		}
	},
}
