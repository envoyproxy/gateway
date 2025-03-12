// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from upstream gateway-api, it will be moved to upstream.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, TCPRouteBackendFQDNTest)
	ConformanceTests = append(ConformanceTests, TCPRouteBackendIPTest)
}

var TCPRouteBackendFQDNTest = suite.ConformanceTest{
	ShortName:   "TCPRouteBackendFQDNTest",
	Description: "TCPRoutes with a backend ref to a FQDN Backend",
	Manifests:   []string{"testdata/tcproute-to-backend-fqdn.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("tcp-route-1", func(t *testing.T) {
			testTCPRouteWithBackend(t, suite, "backend-fqdn")
		})
	},
}

var TCPRouteBackendIPTest = suite.ConformanceTest{
	ShortName:   "TCPRouteBackendIPTest",
	Description: "TCPRoutes with a backend ref to an IP Backend",
	Manifests:   []string{"testdata/tcproute-to-backend-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("tcp-route-1", func(t *testing.T) {
			svcNN := types.NamespacedName{
				Name:      "infra-backend-v1",
				Namespace: "gateway-conformance-infra",
			}
			svc, err := GetService(suite.Client, svcNN)
			if err != nil {
				t.Fatalf("failed to get service %s: %v", svcNN, err)
			}

			backendIPName := "backend-ip"
			ns := "gateway-conformance-infra"
			err = CreateBackend(suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns}, svc.Spec.ClusterIP, 8080)
			if err != nil {
				t.Fatalf("failed to create backend %s: %v", backendIPName, err)
			}
			t.Cleanup(func() {
				if err := DeleteBackend(suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns}); err != nil {
					t.Fatalf("failed to delete backend %s: %v", backendIPName, err)
				}
			})
			testTCPRouteWithBackend(t, suite, backendIPName)
		})
	},
}

func testTCPRouteWithBackend(t *testing.T, suite *suite.ConformanceTestSuite, backendName string) {
	ns := "gateway-conformance-infra"
	routeNN := types.NamespacedName{Name: "tcp-app-1", Namespace: ns}
	gwNN := types.NamespacedName{Name: "my-tcp-gateway", Namespace: ns}
	gwAddr := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), routeNN)
	BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: backendName, Namespace: ns})
	OkResp := http.ExpectedResponse{
		Request: http.Request{
			Path: "/",
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}

	// Send a request to an valid path and expect a successful response
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
}
