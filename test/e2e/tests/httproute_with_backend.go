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
	ConformanceTests = append(ConformanceTests, EnvoyGatewayBackendTest)
}

var EnvoyGatewayBackendTest = suite.ConformanceTest{
	ShortName:   "EnvoyGatewayBackend",
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
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("of type IP", func(t *testing.T) {
			svcNN := types.NamespacedName{
				Name:      "infra-backend-v1-clusterip",
				Namespace: "gateway-conformance-infra",
			}
			svc, err := GetService(suite.Client, svcNN)
			if err != nil {
				t.Fatalf("failed to get service %s: %v", svcNN, err)
			}

			backendIPName := "backend-ip"
			ns := "gateway-conformance-infra"
			err = CreateBackend(suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns}, svc.Spec.ClusterIP, 8080)
			if err != nil {
				t.Fatalf("failed to create backend %s: %v", backendIPName, err)
			}
			t.Cleanup(func() {
				if err := DeleteBackend(suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns}); err != nil {
					t.Fatalf("failed to delete backend %s: %v", backendIPName, err)
				}
			})

			routeNN := types.NamespacedName{Name: "httproute-to-backend-ip", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns})

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-ip",
				},
				Response: http.Response{
					StatusCodes: []int{200},
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
					StatusCodes: []int{200},
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
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
