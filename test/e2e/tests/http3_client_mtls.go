// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	tlsutils "sigs.k8s.io/gateway-api/conformance/utils/tls"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/e2e/utils"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTP3ClientMTLSTest)
}

var HTTP3ClientMTLSTest = suite.ConformanceTest{
	ShortName:   "HTTP3ClientMTLS",
	Description: "Client certificate validation on an HTTP/3 listener",
	Manifests:   []string{"testdata/http3-client-mtls.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		routeNN := types.NamespacedName{Name: "http3-mtls-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "http3-mtls-gateway", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "http3-mtls-ctp", Namespace: ns},
			suite.ControllerName, ancestorRef)

		// The same self-signed cert is used as the server cert, the client cert and the trusted CA.
		certificate, certificateKey, _, err := GetTLSSecret(suite.Client, types.NamespacedName{Name: "http3-mtls-certificate", Namespace: ns})
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}

		quicRoundTripper := &utils.QuicRoundTripper{
			Debug:         suite.Debug,
			TimeoutConfig: suite.TimeoutConfig,
		}

		t.Run("client certificate is validated over HTTP/3", func(t *testing.T) {
			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "mtls.example.com",
					Path: "/http3-mtls",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "mtls.example.com",
						Path: "/http3-mtls",
						Headers: map[string]string{
							"X-Forwarded-Client-Cert": "Hash=ac77d86dd638969a0a39b4e0743370e860d1b70da58b1b08ce950417b6386a8b;Subject=\"CN=mtls.example.com,OU=Gateway,O=EnvoyProxy,L=SomeCity,ST=VA,C=US\"",
						},
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, quicRoundTripper, suite.TimeoutConfig,
				gwAddr, certificate, certificate, certificateKey, "mtls.example.com", expected)
		})

		t.Run("connection without a client certificate is rejected over HTTP/3", func(t *testing.T) {
			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "mtls.example.com",
					Path: "/http3-mtls",
				},
				Namespace: ns,
			}
			tlsutils.MakeTLSRequestAndExpectFailureResponse(t, quicRoundTripper,
				gwAddr, certificate, nil, nil, "mtls.example.com", expected)
		})
	},
}
