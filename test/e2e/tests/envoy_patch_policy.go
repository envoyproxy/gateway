// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	stdnet "net"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyPatchPolicyTest)
}

var EnvoyPatchPolicyTest = suite.ConformanceTest{
	ShortName:   "EnvoyPatchPolicy",
	Description: "update xds using EnvoyPatchPolicy",
	Manifests:   []string{"testdata/envoy-patch-policy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("With name", func(t *testing.T) {
			testEnvoyPatchPolicy(t, suite, "same-namespace", "http-envoy-patch-policy", "infra-backend-v1")
		})
		t.Run("Without name", func(t *testing.T) {
			testEnvoyPatchPolicy(t, suite, "epp-gateways", "epp-http", "infra-backend-v2")
			testEnvoyPatchPolicyWithPort(t, suite, "epp-gateways", "epp-http-8080", "infra-backend-v3", "8080")
		})
	},
}

func testEnvoyPatchPolicy(t *testing.T, suite *suite.ConformanceTestSuite, gwName, routeName, backendName string) {
	testEnvoyPatchPolicyWithPort(t, suite, gwName, routeName, backendName, "")
}

func testEnvoyPatchPolicyWithPort(t *testing.T, suite *suite.ConformanceTestSuite, gwName, routeName, backendName, gwPort string) {
	ns := "gateway-conformance-infra"
	routeNN := types.NamespacedName{Name: routeName, Namespace: ns}
	gwNN := types.NamespacedName{Name: gwName, Namespace: ns}
	gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
	if gwPort != "" {
		gwAddr = stdnet.JoinHostPort(gwAddr, gwPort)
	}
	okResp := http.ExpectedResponse{
		Request: http.Request{
			Path: "/epp",
		},
		Response: http.Response{
			StatusCodes: []int{200},
		},
		Backend:   backendName,
		Namespace: ns,
	}

	// Send a request to a valid path and expect a successful response
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResp)

	customResp := http.ExpectedResponse{
		Request: http.Request{
			Path: "/not-exist-path",
		},
		Response: http.Response{
			StatusCodes: []int{406},
		},
		Namespace: ns,
	}

	// Send a request to an invalid path and expect a custom response
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, customResp)
}
