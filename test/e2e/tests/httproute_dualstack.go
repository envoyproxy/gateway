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

// If the environment is not dual, the IPv6 manifest cannot be applied, so the test will be skipped.
func init() {
	if os.Getenv("IP_FAMILY") == "dual" {
		ConformanceTests = append(ConformanceTests, HTTPRouteDualStackTest)
	} else {
		ConformanceTests = append(ConformanceTests, SkipHTTPRouteDualStackTest)
	}
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
	},
}

func runHTTPRouteTest(t *testing.T, suite *suite.ConformanceTestSuite, ns string, gwNN types.NamespacedName, routeName, path string) {
	routeNN := types.NamespacedName{Name: routeName, Namespace: ns}
	gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

	expectedResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}

	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
}

var SkipHTTPRouteDualStackTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteDualStack",
	Description: "Skipping HTTPRouteDualStack test as IP_FAMILY is not dual",
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Skip("Skipping HTTPRouteDualStack test as IP_FAMILY is not dual")
	},
}
