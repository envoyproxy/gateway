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
	ConformanceTests = append(ConformanceTests, HTTPRouteBackendEndpointHostname)
}

var HTTPRouteBackendEndpointHostname = suite.ConformanceTest{
	ShortName:   "HTTPRouteBackendEndpointHostname",
	Description: "An HTTPRoute with backend host rewrite uses BackendTrafficPolicy endpoint hostnames for Service backends",
	Manifests:   []string{"testdata/httproute-backend-endpoint-hostname.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		withBTPRouteNN := types.NamespacedName{Name: "backend-endpoint-hostname-with-btp", Namespace: ns}
		withoutBTPRouteNN := types.NamespacedName{Name: "backend-endpoint-hostname-without-btp", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, withBTPRouteNN, withoutBTPRouteNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, withBTPRouteNN, gwNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, withoutBTPRouteNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				Request: http.Request{
					Host: "example.com",
					Path: "/backend-endpoint-hostname-with-btp",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/backend-endpoint-hostname-with-btp",
						Host: "infra-backend-v1.gateway-conformance-infra.svc.cluster.local",
					},
				},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request: http.Request{
					Host: "example.com",
					Path: "/backend-endpoint-hostname-without-btp",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/backend-endpoint-hostname-without-btp",
						Host: "example.com",
					},
				},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
		}
		for i := range testCases {
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
