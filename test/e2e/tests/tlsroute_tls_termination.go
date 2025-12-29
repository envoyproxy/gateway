// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
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
			cPem, keyPem, caCertPem, err := GetTLSSecret(suite.Client, certNN)
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

			req := http.MakeRequest(t, &expected, gwAddr, "HTTPS", "https")

			// Use the CA cert to verify server certificate, cert is self-signed so it's also the CA
			WaitForConsistentResponseWithCA(
				t,
				suite.RoundTripper,
				&req,
				&expected,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				cPem,
				keyPem,
				caCertPem,
				"foo.example.com")
		})

		t.Run("TLSRoute with TLS termination - route 2 (bar.example.com)", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "tlsroute-terminate-2", Namespace: ns}
			gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls"), routeNN)

			certNN := types.NamespacedName{Name: "tls-termination-certificate", Namespace: ns}
			cPem, keyPem, caCertPem, err := GetTLSSecret(suite.Client, certNN)
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

			req := http.MakeRequest(t, &expected, gwAddr, "HTTPS", "https")

			WaitForConsistentResponseWithCA(
				t,
				suite.RoundTripper,
				&req,
				&expected,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				cPem,
				keyPem,
				caCertPem,
				"bar.example.com")
		})

		t.Run("TLSRoute with TLS termination - route 3 (baz.example.com)", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "tlsroute-terminate-3", Namespace: ns}
			gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls"), routeNN)

			certNN := types.NamespacedName{Name: "tls-termination-certificate", Namespace: ns}
			cPem, keyPem, caCertPem, err := GetTLSSecret(suite.Client, certNN)
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

			req := http.MakeRequest(t, &expected, gwAddr, "HTTPS", "https")

			WaitForConsistentResponseWithCA(
				t,
				suite.RoundTripper,
				&req,
				&expected,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				cPem,
				keyPem,
				caCertPem,
				"baz.example.com")
		})
	},
}

// WaitForConsistentResponseWithCA makes requests with TLS using a CA certificate to verify the server
func WaitForConsistentResponseWithCA(t *testing.T, r roundtripper.RoundTripper, req *roundtripper.Request, expected *http.ExpectedResponse, threshold int, maxTimeToConsistency time.Duration, certPem, keyPem, caCertPem []byte, serverName string) {
	if req == nil {
		t.Fatalf("request cannot be nil")
	}
	if expected == nil {
		t.Fatalf("expected response cannot be nil")
	}

	http.AwaitConvergence(t, threshold, maxTimeToConsistency, func(elapsed time.Duration) bool {
		updatedReq := *req
		updatedReq.Server = serverName
		// Use the certificate and key for TLS setup, CA cert for validation (self-signed cert)
		updatedReq.CertPem = caCertPem
		updatedReq.KeyPem = keyPem

		cReq, cRes, err := r.CaptureRoundTrip(updatedReq)
		if err != nil {
			tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
			return false
		}

		if err := http.CompareRoundTrip(t, &updatedReq, cReq, cRes, *expected); err != nil {
			tlog.Logf(t, "Response expectation failed for request: %+v not ready yet: %v (after %v)", updatedReq, err, elapsed)
			return false
		}

		return true
	})
	tlog.Logf(t, "Request passed")
}
