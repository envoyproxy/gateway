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
	ConformanceTests = append(ConformanceTests, TLSRouteDynamicResolverBackendTest)
}

// TLSRouteDynamicResolverBackendTest verifies that a TLSRoute referencing a DynamicResolver Backend
// forwards the connection based on the SNI (SNI based dynamic forward proxy). The gateway resolves
// the upstream host from the SNI of the incoming connection and forwards the TLS bytes unchanged.
var TLSRouteDynamicResolverBackendTest = suite.ConformanceTest{
	ShortName:   "TLSRouteDynamicResolverBackend",
	Description: "A TLSRoute with a dynamic resolver backend forwards based on the SNI",
	Manifests: []string{
		"testdata/tlsroute-with-dynamic-resolver-backend.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "tlsroute-with-dynamic-resolver-backend", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls"), routeNN)
		BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-dynamic-resolver", Namespace: ns})

		// The SNI must match the backend service in-cluster FQDN so that the dynamic forward proxy can
		// resolve the upstream from it. As this is TLS passthrough, the client validates the backend
		// certificate directly, so the same FQDN must be covered by the backend certificate's SAN.
		sni := "tls-dfp-backend.gateway-conformance-infra.svc.cluster.local"

		serverCertificate, _, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: "tls-dfp-backend-cert", Namespace: ns})
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}

		expected := http.ExpectedResponse{
			Request: http.Request{
				Host: sni,
				Path: "/",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig,
			gwAddr, serverCertificate, nil, nil, sni, expected)
	},
}
