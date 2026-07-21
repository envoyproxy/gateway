// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, WasmEnvoyProxyModuleTest)
}

// WasmEnvoyProxyModuleTest loads Wasm from a local path registered on EnvoyProxy.
// The Envoy image (envoyproxy/gateway-wasm-module-test) embeds the example .wasm.
var WasmEnvoyProxyModuleTest = suite.ConformanceTest{
	ShortName:   "WasmEnvoyProxyModule",
	Description: "Test EnvoyProxyModule Wasm source that loads a local module and adds response headers",
	Manifests:   []string{"testdata/wasm-envoyproxy-module.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with envoy proxy module wasm", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-wasm-module", Namespace: ns}
			gwNN := types.NamespacedName{Name: "wasm-module-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "wasm-module-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			// Wait for the Envoy proxy pods to be running and ready.
			gwPodNamespace := GetGatewayResourceNamespace()
			WaitForPods(t, suite.Client, gwPodNamespace, map[string]string{
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}, corev1.PodRunning, &PodReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/wasm-module",
				},
				// Empty ExpectedRequest: the example Wasm appends "Hello, world" to the
				// response body, which invalidates the JSON format used to extract request
				// properties (same workaround as the HTTP Wasm e2e).
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host:    "",
						Method:  "",
						Path:    "",
						Headers: nil,
					},
				},
				Namespace: "",
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-wasm-custom": "FOO",
					},
				},
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
