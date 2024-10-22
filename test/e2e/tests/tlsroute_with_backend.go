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
	ConformanceTests = append(ConformanceTests, TLSRouteBackendFQDNTest)
	ConformanceTests = append(ConformanceTests, TLSRouteBackendIPTest)
}

var TLSRouteBackendFQDNTest = suite.ConformanceTest{
	ShortName:   "TLSRouteBackendFQDNTest",
	Description: "TLSRoutes with a backend ref to a Backend",
	Manifests: []string{
		"testdata/tlsroute-to-backend-fqdn.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("TLSRoute with a FQDN type Backend", func(t *testing.T) {
			testTLSRouteWithBackend(t, suite, "tlsroute-to-backend-fqdn", "backend-fqdn")
		})
	},
}

var TLSRouteBackendIPTest = suite.ConformanceTest{
	ShortName:   "TLSRouteBackendIPTest",
	Description: "TLSRoutes with a backend ref to a Backend",
	Manifests: []string{
		"testdata/tlsroute-to-backend-ip.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("TLSRoute with a IP type Backend", func(t *testing.T) {
			testTLSRouteWithBackend(t, suite, "tlsroute-to-backend-ip", "backend-ip")
		})
	},
}

func testTLSRouteWithBackend(t *testing.T, suite *suite.ConformanceTestSuite, route, backend string) {
	ns := "gateway-conformance-infra"
	routeNN := types.NamespacedName{Name: route, Namespace: ns}
	gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
	gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls"), routeNN)
	certNN := types.NamespacedName{Name: "backend-tls-certificate", Namespace: ns}

	BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: backend, Namespace: ns})

	expected := http.ExpectedResponse{
		Request: http.Request{
			Host: "example.com",
			Path: "/",
		},
		Response: http.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}

	req := http.MakeRequest(t, &expected, gwAddr, "HTTPS", "https")

	// This test uses the same key/cert pair as both a client cert and server cert
	// Both backend and client treat the self-signed cert as a trusted CA
	cPem, keyPem, err := GetTLSSecret(suite.Client, certNN)
	if err != nil {
		t.Fatalf("unexpected error finding TLS secret: %v", err)
	}

	WaitForConsistentMTLSResponse(
		t,
		suite.RoundTripper,
		req,
		expected,
		suite.TimeoutConfig.RequiredConsecutiveSuccesses,
		suite.TimeoutConfig.MaxTimeToConsistency,
		cPem,
		keyPem,
		"example.com")
}
