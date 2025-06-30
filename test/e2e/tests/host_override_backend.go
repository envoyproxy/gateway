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
	ConformanceTests = append(ConformanceTests, HostOverrideBackendTest)
}

var HostOverrideBackendTest = suite.ConformanceTest{
	ShortName:   "HostOverrideBackend",
	Description: "Routes with host override backend that selects endpoints based on headers or metadata",
	Manifests: []string{
		"testdata/host-override-backend-header.yaml",
		"testdata/host-override-backend-metadata.yaml",
		"testdata/host-override-backend-multiple-sources.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Host override with header source", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "httproute-host-override-header", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-host-override-header", Namespace: ns})

			// Test with target-pod header pointing to a valid backend service
			// Use the cluster IP of the infra-backend-v1 service
			svcNN := types.NamespacedName{
				Name:      "infra-backend-v1",
				Namespace: "gateway-conformance-infra",
			}
			svc, err := GetService(suite.Client, svcNN)
			if err != nil {
				t.Fatalf("failed to get service %s: %v", svcNN, err)
			}

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/header-override",
					Headers: map[string]string{
						"target-pod": svc.Spec.ClusterIP + ":8080",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// Test without header - should fail or use fallback
			expectedResponseNoHeader := http.ExpectedResponse{
				Request: http.Request{
					Path: "/header-override",
				},
				Response: http.Response{
					StatusCode: 503, // Service unavailable when no override host is provided
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponseNoHeader)
		})

		t.Run("Host override with metadata source", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "httproute-host-override-metadata", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-host-override-metadata", Namespace: ns})

			// Note: Testing metadata-based override is more complex as it requires
			// setting up dynamic metadata through filters or extensions.
			// For now, we test that the backend is accepted and the route is configured.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/metadata-override",
				},
				Response: http.Response{
					StatusCode: 503, // Expected when no metadata is set
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("Host override with multiple sources", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "httproute-host-override-multiple", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-host-override-multiple", Namespace: ns})

			// Get the service cluster IP for testing
			svcNN := types.NamespacedName{
				Name:      "infra-backend-v1",
				Namespace: "gateway-conformance-infra",
			}
			svc, err := GetService(suite.Client, svcNN)
			if err != nil {
				t.Fatalf("failed to get service %s: %v", svcNN, err)
			}

			// Test with header source (first in priority)
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/multiple-override",
					Headers: map[string]string{
						"target-pod": svc.Spec.ClusterIP + ":8080",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// Test without header - should fall back to metadata or fail
			expectedResponseNoHeader := http.ExpectedResponse{
				Request: http.Request{
					Path: "/multiple-override",
				},
				Response: http.Response{
					StatusCode: 503, // Expected when no override sources provide valid host
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponseNoHeader)
		})
	},
}
