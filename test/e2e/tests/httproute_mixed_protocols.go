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
	ConformanceTests = append(ConformanceTests, HTTPRouteMixedProtocols)
}

var HTTPRouteMixedProtocols = suite.ConformanceTest{
	ShortName:   "HTTPRouteMixedProtocols",
	Description: "HTTPRoute with mixed protocols: HTTP and HTTPS",
	Manifests:   []string{"testdata/httproute-mixed-protocol-backend.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "mixed-protocols", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		// Wait for the BackendTLSPolicy to be accepted before making requests to ensure the TLS backend is ready.
		kubernetes.BackendTLSPolicyMustHaveAcceptedConditionTrue(t, suite.Client, suite.TimeoutConfig, types.NamespacedName{
			Name:      "policy-btls",
			Namespace: "gateway-conformance-infra",
		}, gwNN)

		// Make sure both backends are healthy before making requests to verify traffic is routed correctly.
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
			Request: http.Request{
				Path: "/infra-backend-v1",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
			Backend:   "infra-backend-v1",
		})

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
			Request: http.Request{
				Path: "/tls-backend-2",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
			Backend:   "tls-backend-2",
		})

		// Make requests to both backends to verify traffic is routed correctly.
		responses := []http.ExpectedResponse{
			{
				Request: http.Request{
					Path: "/mixed-protocols",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
				Backend:   "infra-backend-v1",
			},
			{
				Request: http.Request{
					Path: "/mixed-protocols",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
				Backend:   "tls-backend-2",
			},
		}

		for _, res := range responses {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, res)
		}
	},
}
