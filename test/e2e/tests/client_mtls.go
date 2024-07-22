// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, ClientMTLSTest)
}

var ClientMTLSTest = suite.ConformanceTest{
	ShortName:   "ClientMTLS",
	Description: "Use Gateway with Client MTLS policy",
	Manifests:   []string{"testdata/client-mtls.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Use Client MTLS", func(t *testing.T) {
			depNS := "envoy-gateway-system"
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-client-mtls", Namespace: ns}
			gwNN := types.NamespacedName{Name: "client-mtls-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{depNS})
			certNN := types.NamespacedName{Name: "client-mtls-certificate", Namespace: ns}

			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "mtls.example.com",
					Path: "/client-mtls",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "mtls.example.com",
						Path: "/client-mtls",
						Headers: map[string]string{
							"X-Forwarded-Client-Cert": "Hash=ac77d86dd638969a0a39b4e0743370e860d1b70da58b1b08ce950417b6386a8b;Subject=\"CN=mtls.example.com,OU=Gateway,O=EnvoyProxy,L=SomeCity,ST=VA,C=US\"",
						},
					},
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

			WaitForConsistentMTLSResponse(t, suite.RoundTripper, req, expected, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency, cPem, keyPem, "mtls.example.com")
		})
	},
}

func WaitForConsistentMTLSResponse(t *testing.T, r roundtripper.RoundTripper, req roundtripper.Request, expected http.ExpectedResponse, threshold int, maxTimeToConsistency time.Duration, cPem, keyPem []byte, server string) {
	http.AwaitConvergence(t, threshold, maxTimeToConsistency, func(elapsed time.Duration) bool {
		req.KeyPem = keyPem
		req.CertPem = cPem
		req.Server = server

		cReq, cRes, err := r.CaptureRoundTrip(req)
		if err != nil {
			tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
			return false
		}

		if err := http.CompareRequest(t, &req, cReq, cRes, expected); err != nil {
			tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
			return false
		}

		return true
	})
	tlog.Logf(t, "Request passed")
}

// GetTLSSecret fetches the named Secret and converts both cert and key to []byte
func GetTLSSecret(client client.Client, secretName types.NamespacedName) ([]byte, []byte, error) {
	var cert, key []byte

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	secret := &corev1.Secret{}
	err := client.Get(ctx, secretName, secret)
	if err != nil {
		return cert, key, fmt.Errorf("error fetching TLS Secret: %w", err)
	}
	cert = secret.Data["tls.crt"]
	key = secret.Data["tls.key"]

	return cert, key, nil
}
