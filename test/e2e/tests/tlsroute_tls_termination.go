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
	tlsutils "sigs.k8s.io/gateway-api/conformance/utils/tls"
)

func init() {
	ConformanceTests = append(ConformanceTests, TLSRouteTLSTerminationTest)
}

var TLSRouteTLSTerminationTest = suite.ConformanceTest{
	ShortName:   "TLSRouteTLSTermination",
	Description: "TLSRoute with TLS Termination and SNI-based routing",
	Manifests: []string{
		"testdata/tlsroute-tls-termination.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "tlsroute-termination-gateway", Namespace: ns}

		t.Run("TLSRoute with TLS termination - route 1 (foo.example.com)", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "tlsroute-terminate-1", Namespace: ns}
			gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls"), routeNN)

			certNN := types.NamespacedName{Name: "tls-termination-certificate", Namespace: ns}
			serverCertificate, _, _, err := GetTLSSecret(suite.Client, certNN)
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "foo.example.com",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
				gwAddr, serverCertificate, nil, nil, "foo.example.com", expected)
		})

		t.Run("TLSRoute with TLS termination - route 2 (bar.example.com)", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "tlsroute-terminate-2", Namespace: ns}
			gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls"), routeNN)

			certNN := types.NamespacedName{Name: "tls-termination-certificate", Namespace: ns}
			serverCertificate, _, _, err := GetTLSSecret(suite.Client, certNN)
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "bar.example.com",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
				gwAddr, serverCertificate, nil, nil, "baz.example.com", expected)
		})

		t.Run("TLSRoute with TLS termination - route 3 (baz.example.com)", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "tlsroute-terminate-3", Namespace: ns}
			gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls"), routeNN)

			certNN := types.NamespacedName{Name: "tls-termination-certificate", Namespace: ns}
			serverCertificate, _, _, err := GetTLSSecret(suite.Client, certNN)
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "baz.example.com",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
				gwAddr, serverCertificate, nil, nil, "baz.example.com", expected)
		})
	},
}
