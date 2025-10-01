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
	ConformanceTests = append(ConformanceTests, PreserveRouteOrderTest)
}

var PreserveRouteOrderTest = suite.ConformanceTest{
	ShortName:   "PreserveRouteOrder",
	Description: "Route order should be preserved",
	Manifests:   []string{"testdata/preserve-route-order.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Request will match the less-specific rule due to order preservation", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-preserved-route-order", Namespace: ns}
			gwNN := types.NamespacedName{Name: "preserve-route-order-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/specific/foo",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/specific/foo",
						Headers: map[string]string{
							"matched-rule-prefix": "/", // the less specific route rule is used because the user-defined order is preserved
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
