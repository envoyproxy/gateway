// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, JAFingerpintTest)
}

var JAFingerpintTest = suite.ConformanceTest{
	ShortName:   "JAFingerprint",
	Description: "JAFingerprint tests ensure that Envoy Gateway supports TLS JAX fingerprint features.",
	Manifests:   []string{"testdata/ja-fingerprint.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Only JA3 header in response (foo.example.com)", func(t *testing.T) {
			testJAFingerprint(t, suite, "http-route-ja3-foo", "foo-com-tls", "foo.example.com", 443, true, false)
		})

		t.Run("JA3 and JA4 headers in response (bar.example.com)", func(t *testing.T) {
			testJAFingerprint(t, suite, "http-route-ja3-and-ja4-bar", "bar-com-tls", "bar.example.com", 444, true, true)
		})
	},
}

func testJAFingerprint(t *testing.T, suite *suite.ConformanceTestSuite, routeName, secretName, serverName string, port int, expectJA3Fingerprint, expectJA4Fingerprint bool) {
	routeNN := types.NamespacedName{Name: routeName, Namespace: ConformanceInfraNamespace}
	gwNN := types.NamespacedName{Name: "https-fingerprint-gw", Namespace: ConformanceInfraNamespace}
	gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

	expected := http.ExpectedResponse{
		Request: http.Request{
			Host: serverName,
			Path: "/",
		},
		Response: http.Response{
			StatusCodes: []int{200},
		},
		Namespace: ConformanceInfraNamespace,
	}

	certPem, keyPem, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: secretName, Namespace: ConformanceInfraNamespace})
	if err != nil {
		t.Fatalf("unexpected error finding TLS secret: %v", err)
	}

	req := http.MakeRequest(t, &expected, fmt.Sprintf("%s:%d", gwAddr, port), "HTTPS", "https")
	req.Server = serverName
	// Use the certificate and key for TLS setup, CA cert for validation (self-signed cert)
	req.CertPem = certPem
	req.KeyPem = keyPem

	http.AwaitConvergence(t, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
		cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
		if err != nil {
			tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
			return false
		}

		if err := http.CompareRoundTrip(t, &req, cReq, cRes, expected); err != nil {
			tlog.Logf(t, "Response expectation failed for request: %+v not ready yet: %v (after %v)", req, err, elapsed)
			return false
		}

		ja3HeaderValue, hasJA3 := cRes.Headers["Ja3-Fingerprint"]
		if expectJA3Fingerprint && !hasJA3 {
			tlog.Logf(t, "JA3 fingerprint header missing (expected to be present) after %v", elapsed)
			return false
		} else if !expectJA3Fingerprint && hasJA3 {
			tlog.Logf(t, "JA3 fingerprint header unexpectedly present: %s after %v", ja3HeaderValue, elapsed)
			return false
		}

		ja4HeaderValue, hasJA4 := cRes.Headers["Ja4-Fingerprint"]
		if expectJA4Fingerprint && !hasJA4 {
			tlog.Logf(t, "JA4 fingerprint header missing (expected to be present) after %v", elapsed)
			return false
		} else if !expectJA4Fingerprint && hasJA4 {
			tlog.Logf(t, "JA4 fingerprint header unexpectedly present: %s after %v", ja4HeaderValue, elapsed)
			return false
		}

		return true
	})

	tlog.Logf(t, "JA fingerprint test completed successfully for route '%s' on server '%s:%d'", routeName, serverName, port)
}
