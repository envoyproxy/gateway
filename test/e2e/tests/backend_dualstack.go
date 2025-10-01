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

// If the environment is not dual, the IPv6 manifest cannot be applied, so the test will be skipped.
func init() {
	ConformanceTests = append(ConformanceTests, BackendDualStackTest)
}

var BackendDualStackTest = suite.ConformanceTest{
	ShortName:   "BackendDualStack",
	Description: "Test IPv6 and Dual Stack support for backends",
	Manifests:   []string{"testdata/backend-dualstack.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "dualstack-gateway", Namespace: ns}

		t.Run("IPv6 Backend", func(t *testing.T) {
			runBackendDualStackTest(t, suite, ns, gwNN, "infra-backend-v1-route-ipv6", "/backend-ipv6")
		})
		t.Run("Dual Stack Backend", func(t *testing.T) {
			runBackendDualStackTest(t, suite, ns, gwNN, "infra-backend-v1-route-dualstack", "/backend-dualstack")
		})
		t.Run("IPv4 Backend", func(t *testing.T) {
			runBackendDualStackTest(t, suite, ns, gwNN, "infra-backend-v1-route-ipv4", "/backend-ipv4")
		})
	},
}

func runBackendDualStackTest(t *testing.T, suite *suite.ConformanceTestSuite, ns string, gwNN types.NamespacedName, routeName, path string) {
	routeNN := types.NamespacedName{Name: routeName, Namespace: ns}
	gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

	expectedResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
		},
		Response: http.Response{
			StatusCodes: []int{200},
		},
		Namespace: ns,
	}

	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
}
