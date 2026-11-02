// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, DynamicModuleTest)
}

var DynamicModuleTest = suite.ConformanceTest{
	ShortName:   "DynamicModule",
	Description: "Test dynamic module headers and shared backend callouts from Lua, Wasm and dynamic modules",
	Manifests:   []string{"testdata/dynamic-module.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with dynamic module filter", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-dynamic-module", Namespace: ns}
			gwNN := types.NamespacedName{Name: "dynamic-module-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "dynamic-module-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			// Wait for the Envoy proxy pods to be running and ready.
			gwPodNamespace := GetGatewayResourceNamespace()
			WaitForPods(t, suite.Client, gwPodNamespace, map[string]string{
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}, corev1.PodRunning, &PodReady)

			WaitForPods(t, suite.Client, ns, map[string]string{"app": "callout-failure"}, corev1.PodRunning, &PodReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/dynamic-module",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-dynamic-module": "true",
						"x-module-callout": "true",
						"x-lua-callout":    "true",
						"x-wasm-callout":   "true",
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// The child rebinds the alias while inheriting the same filter instances.
			// A failed callout distinguishes its backend from the echo service.
			// Test each runtime separately so an earlier filter cannot mask its result.
			original := &egv1a1.EnvoyExtensionPolicy{}
			require.NoError(t, suite.Client.Get(t.Context(), types.NamespacedName{Name: "dynamic-module-test", Namespace: ns}, original))
			for _, runtime := range []struct{ name, header string }{
				{"lua", "x-lua-callout"},
				{"wasm", "x-wasm-callout"},
				{"dynamic-module", "x-module-callout"},
			} {
				t.Run(runtime.name+" route isolation", func(t *testing.T) {
					policy := &egv1a1.EnvoyExtensionPolicy{}
					require.NoError(t, suite.Client.Get(t.Context(), types.NamespacedName{Name: "dynamic-module-test", Namespace: ns}, policy))
					policy.Spec.Lua = nil
					policy.Spec.Wasm = nil
					policy.Spec.DynamicModule = nil
					switch runtime.name {
					case "lua":
						policy.Spec.Lua = original.Spec.Lua
					case "wasm":
						policy.Spec.Wasm = original.Spec.Wasm
					case "dynamic-module":
						policy.Spec.DynamicModule = original.Spec.DynamicModule
					}
					require.NoError(t, suite.Client.Update(t.Context(), policy))

					absentHeaders := []string{}
					for _, header := range []string{"x-lua-callout", "x-wasm-callout", "x-module-callout"} {
						if header != runtime.header {
							absentHeaders = append(absentHeaders, header)
						}
					}
					for range 3 {
						http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
							Request:   http.Request{Path: "/dynamic-module"},
							Response:  http.Response{StatusCodes: []int{200}, Headers: map[string]string{runtime.header: "true"}, AbsentHeaders: absentHeaders},
							Namespace: ns,
						})

						http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
							Request:   http.Request{Path: "/dynamic-module-isolation"},
							Response:  http.Response{StatusCodes: []int{502}},
							Namespace: ns,
						})
					}
				})
			}
		})
	},
}
