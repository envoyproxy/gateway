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
	ConformanceTests = append(ConformanceTests, HTTPRouteHostnameMatch)
}

var HTTPRouteHostnameMatch = suite.ConformanceTest{
	ShortName:   "HTTPRouteHostnameMatch",
	Description: "Request to match the \"catch all\" route",
	Manifests:   []string{"testdata/httproute-hostname-match.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		routeHost := types.NamespacedName{Name: "hostname-match", Namespace: ns}
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeHost, gwNN)

		routePathThrough := types.NamespacedName{Name: "hostname-match-passthrough", Namespace: ns}
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routePathThrough, gwNN)

		gwAddrs := []string{
			kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeHost),
			kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routePathThrough),
		}

		testCases := []http.ExpectedResponse{
			{
				Request:         http.Request{Host: "www.host.org", Path: "/pathmatch"},
				ExpectedRequest: &http.ExpectedRequest{Request: http.Request{Path: "/pathmatch"}},
				Backend:         "infra-backend-v1",
				Namespace:       ns,
			},
			{
				Request:         http.Request{Host: "www.host.org", Path: "/otherpath"},
				ExpectedRequest: &http.ExpectedRequest{Request: http.Request{Host: "www.host.org", Path: "/otherpath"}},
				Backend:         "infra-backend-v2",
				Namespace:       ns,
			},
		}
		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			gwAddr := gwAddrs[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
