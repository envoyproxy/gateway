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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteDualStackTest)
}

var HTTPRouteDualStackTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteDualStack",
	Description: "Test HTTPRoute support for IPv6 only, dual-stack, and IPv4 only services",
	Manifests:   []string{"testdata/httproute-dualstack.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "dualstack-gateway", Namespace: ns}

		t.Run("HTTPRoute to IPv6 only service", func(t *testing.T) {
			runHTTPRouteTest(t, suite, ns, gwNN, "infra-backend-v1-httproute-ipv6", "/ipv6-only")
		})
		t.Run("HTTPRoute to Dual-stack service", func(t *testing.T) {
			runHTTPRouteTest(t, suite, ns, gwNN, "infra-backend-v1-httproute-dualstack", "/dual-stack")
		})
		t.Run("HTTPRoute to IPv4 only service", func(t *testing.T) {
			runHTTPRouteTest(t, suite, ns, gwNN, "infra-backend-v1-httproute-ipv4", "/ipv4-only")
		})
		t.Run("HTTPRoute to All-stacks services", func(t *testing.T) {
			runHTTPRouteTest(t, suite, ns, gwNN, "infra-backend-v1-httproute-all-stacks", "/all-stacks")
		})
	},
}

func runHTTPRouteTest(t *testing.T, suite *suite.ConformanceTestSuite, ns string, gwNN types.NamespacedName, routeName, path string) {
	routeNN := types.NamespacedName{Name: routeName, Namespace: ns}
	gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

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
