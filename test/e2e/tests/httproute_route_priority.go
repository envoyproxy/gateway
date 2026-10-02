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
	ConformanceTests = append(ConformanceTests, HTTPRouteRoutePriorityTest)
}

var HTTPRouteRoutePriorityTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteRoutePriority",
	Description: "HTTPRoute route-priority annotation orders matches across HTTPRoutes",
	Manifests:   []string{"testdata/httproute-route-priority.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Higher priority catch-all wins over a more specific HTTPRoute", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			catchAllNN := types.NamespacedName{Name: "http-route-priority-catch-all", Namespace: ns}
			specificNN := types.NamespacedName{Name: "http-route-priority-specific", Namespace: ns}
			gwNN := types.NamespacedName{Name: "route-priority-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, catchAllNN, specificNN)

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/specific/foo",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/specific/foo",
						Headers: map[string]string{
							"matched-route": "catch-all",
						},
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})
	},
}
