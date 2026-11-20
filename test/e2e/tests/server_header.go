// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"net"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, ServerHeaderTest)
}

var ServerHeaderTest = suite.ConformanceTest{
	ShortName:   "ServerHeader",
	Description: "Test that the ClientTrafficPolicy API implementation supports configuring the Server response header",
	Manifests:   []string{"testdata/server-header.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Overwrite replaces the Server header set by the backend", func(t *testing.T) {
			gwAddr := serverHeaderGatewayAddr(t, suite,
				"server-header-overwrite-gateway", "http-with-server-header-overwrite", "server-header-overwrite-ctp")

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/server-header",
				},
				BackendSetResponseHeaders: map[string]string{
					"Server": "backend",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"Server": "envoy-gateway-e2e",
					},
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})

		t.Run("AppendIfAbsent keeps the Server header set by the backend", func(t *testing.T) {
			gwAddr := serverHeaderGatewayAddr(t, suite,
				"server-header-append-gateway", "http-with-server-header-append", "server-header-append-ctp")

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/server-header",
				},
				BackendSetResponseHeaders: map[string]string{
					"Server": "backend",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"Server": "backend",
					},
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})

		t.Run("AppendIfAbsent adds the Server header when the backend omits it", func(t *testing.T) {
			gwAddr := serverHeaderGatewayAddr(t, suite,
				"server-header-append-gateway", "http-with-server-header-append", "server-header-append-ctp")

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/server-header",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"Server": "envoy-gateway-e2e",
					},
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})

		t.Run("PassThrough leaves the Server header untouched", func(t *testing.T) {
			gwAddr := serverHeaderGatewayAddr(t, suite,
				"server-header-passthrough-gateway", "http-with-server-header-passthrough", "server-header-passthrough-ctp")

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/server-header",
				},
				BackendSetResponseHeaders: map[string]string{
					"Server": "backend",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"Server": "backend",
					},
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})

		t.Run("PassThrough does not add a Server header when the backend omits it", func(t *testing.T) {
			gwAddr := serverHeaderGatewayAddr(t, suite,
				"server-header-passthrough-gateway", "http-with-server-header-passthrough", "server-header-passthrough-ctp")

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/server-header",
				},
				Response: http.Response{
					StatusCodes:   []int{200},
					AbsentHeaders: []string{"Server"},
				},
				Namespace: ConformanceInfraNamespace,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})
	},
}

func serverHeaderGatewayAddr(t *testing.T, suite *suite.ConformanceTestSuite, gateway, route, policy string) string {
	t.Helper()

	gwNN := types.NamespacedName{Name: gateway, Namespace: ConformanceInfraNamespace}
	routeNN := types.NamespacedName{Name: route, Namespace: ConformanceInfraNamespace}
	gwHost := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
		kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

	ancestorRef := gwapiv1.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}
	ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: policy, Namespace: ConformanceInfraNamespace},
		suite.ControllerName, ancestorRef)

	return net.JoinHostPort(gwHost, "80")
}
