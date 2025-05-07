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
	ConformanceTests = append(ConformanceTests, DynamicResolverBackendTest)
}

var DynamicResolverBackendTest = suite.ConformanceTest{
	ShortName:   "DynamicResolverBackend",
	Description: "Routes with a backend ref to a dynamic resolver backend",
	Manifests: []string{
		"testdata/httproute-with-dynamic-resolver-backend.yaml",
		"testdata/httproute-with-dynamic-resolver-backend-with-tls.yaml",
		"testdata/httproute-with-dynamic-resolver-backend-with-tls-system-ca.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "httproute-with-dynamic-resolver-backend", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-dynamic-resolver", Namespace: ns})

		t.Run("route to service foo", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "test-service-foo.gateway-conformance-infra",
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
		t.Run("route to service bar", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "test-service-bar.gateway-conformance-infra",
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
		t.Run("route to service with TLS", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "httproute-with-dynamic-resolver-backend-tls", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-dynamic-resolver-tls", Namespace: ns})

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "backend-dynamic-resolver-tls.gateway-conformance-infra.svc.cluster.local:443",
					Path: "/with-tls",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
		t.Run("route to service with TLS using system CA", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "httproute-with-dynamic-resolver-backend-tls-system-trust-store", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-dynamic-resolver-tls-system-trust-store", Namespace: ns})

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "gateway.envoyproxy.io:443",
					Path: "/with-tls-system-trust-store",
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
	},
}
