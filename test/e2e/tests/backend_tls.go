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
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSTest, BackendClusterTrustBundleTest)
}

var BackendTLSTest = suite.ConformanceTest{
	ShortName:   "BackendTLS",
	Description: "Connect to backend with TLS",
	Manifests:   []string{"testdata/backend-tls.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ConformanceInfraNamespace}
		t.Run("with a backend TLS Policy", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-with-backend-tls", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-tls",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("with a backend TLS Policy using Truststore", func(t *testing.T) {
			// the upstream used is the eg site which doesn't support IPv6 at this time
			if IPFamily == "ipv6" {
				t.Skip("Skipping test as IP_FAMILY is IPv6")
			}
			routeNN := types.NamespacedName{Name: "http-with-backend-tls-system-trust-store", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
					Host: "gateway.envoyproxy.io",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("without a backend TLS Policy", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-without-backend-tls", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-tls-without-policy",
				},
				Response: http.Response{
					StatusCode: 400, // Bad Request: Client sent an HTTP request to an HTTPS server
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("with CA mismatch and skip tls verify", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-with-backend-insecure-skip-verify-and-mismatch-ca", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-tls-skip-verify-and-mismatch-ca",
				},
				Response: http.Response{
					StatusCode: 200, // Bad Request: Client sent an HTTP request to an HTTPS server
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("without BackendTLSPolicy and skip tls verify", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-with-backend-insecure-skip-verify-without-backend-tls-policy", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-tls-skip-verify-without-backend-tls-policy",
				},
				Response: http.Response{
					StatusCode: 200, // Bad Request: Client sent an HTTP request to an HTTPS server
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("With BackendTLSPolicy and Backend.TLS.AutoSNI", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-with-backend-tls-auto-sni", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "example.com",
					Path: "/backend-auto-sni",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ConformanceInfraNamespace,
			}

			// make assertion
			// Since BTLSP uses a Hostname that's incorrect, the request will only succeed if:
			// SNI is rewritten to example.com and DNS SAN validation is done according to this SNI
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}

var BackendClusterTrustBundleTest = suite.ConformanceTest{
	ShortName:   "BackendTLSClusterTrustBundle",
	Description: "Connect to backend with TLS",
	Manifests: []string{
		"testdata/backend-tls-clustertrustbundle.yaml",
	},
	Features: []features.FeatureName{
		ClusterTrustBundleFeature,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{Name: AllNamespacesGateway, Namespace: ConformanceInfraNamespace}
		t.Run("with ClusterTrustBundle", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-with-backend-tls-trust-bundle", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cluster-trust-bundle",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
