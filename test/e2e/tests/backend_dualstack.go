// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendDualStackTest)
}

var BackendDualStackTest = suite.ConformanceTest{
	ShortName:   "BackendDualStack",
	Description: "Test IPv6 and Dual Stack support for backends",
	Manifests:   []string{"testdata/backend-dualstack.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ipFamily := os.Getenv("IP_FAMILY")
		if ipFamily == "ipv4" {
			t.Skip("Skipping IPv6 and Dual Stack test as IP_FAMILY is set to 'ipv4'")
			return
		}

		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		t.Run("IPv6 Backend", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "httproute-to-backend-ipv6", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-ipv6",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("Dual Stack Backend", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "httproute-to-backend-dualstack", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-dualstack",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
