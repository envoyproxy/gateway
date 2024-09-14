// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyGatewayBackendTest)
}

var EnvoyGatewayBackendTest = suite.ConformanceTest{
	ShortName:   "EnvoyGatewayBackendTest",
	Description: "Routes with a backend ref to a backend",
	Manifests: []string{
		"testdata/httproute-to-backend-fqdn.yaml",
		"testdata/httproute-to-backend-ip.yaml",
		"testdata/httproute-to-backend-fqdn-http2.yaml",
		"testdata/httproute-to-backend-fqdn-tls.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("of type FQDN", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "httproute-to-backend-fqdn", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-fqdn", Namespace: ns})

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-fqdn",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("of type IP", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "httproute-to-backend-ip", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-ip", Namespace: ns})

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-ip",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("of type FQDN with HTTP2", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "httproute-to-backend-fqdn-http2", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-fqdn-http2", Namespace: ns})

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-fqdn-http2",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("of type FQDN with a backend TLS Policy", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "httproute-to-fqdn-backend-tls", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-fqdn-tls", Namespace: ns})

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-fqdn-tls",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
