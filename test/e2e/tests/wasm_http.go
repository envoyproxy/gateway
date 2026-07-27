// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPWasmTest)
}

// HTTPWasmTest tests Wasm extension for an http route with HTTP Wasm configured.
var HTTPWasmTest = suite.ConformanceTest{
	ShortName:   "WasmHTTPCodeSource",
	Description: "Test Wasm extension that adds response headers",
	Manifests:   []string{"testdata/wasm-http.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with http wasm source", func(t *testing.T) {
			testWasmHTTPCodeSource(t, suite, "http-with-http-wasm-source", "http-wasm-source-test", "/wasm-http")
		})

		t.Run("http route with http wasm source no sha", func(t *testing.T) {
			testWasmHTTPCodeSource(t, suite, "http-with-http-wasm-source-no-sha", "http-wasm-source-test-no-sha", "/wasm-http-no-sha")
		})

		t.Run("http route without wasm", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-without-wasm", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "http-wasm-source-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/no-wasm",
				},
				Response: http.Response{
					StatusCodes:   []int{200},
					AbsentHeaders: []string{"x-wasm-custom"},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// Unlike the shared/bounded Lua VM count (see the "lua vm count stays bounded" check in
		// lua_http.go), wasmConfig() in internal/xds/translator/wasm.go sets VmId to a name that
		// is unique per EnvoyExtensionPolicy wasm entry ("Do not share VMs across different
		// filters"), so each of the 2 wasm-configured routes above gets its own dedicated Wasm
		// filter/VM even though both reference the same underlying .wasm module. The
		// process-wide "wasm.wasm_vm_count" gauge should therefore read exactly 3* worker thread.
		t.Run("wasm vm count is per-route", func(t *testing.T) {
			promQL := `sum(envoy_wasm_wasm_vm_count{app_kubernetes_io_component="proxy", app_kubernetes_io_managed_by="envoy-gateway", app_kubernetes_io_name="envoy", gateway_envoyproxy_io_owning_gateway_name="same-namespace"})`
			// 3 is the count of routes
			expectedCount := model.SampleValue(3 * runtime.NumCPU())
			tlog.Logf(t, "expected to got %v", expectedCount)
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					v, err := prometheus.QueryPrometheus(suite.Client, promQL)
					if err != nil {
						tlog.Logf(t, "failed to query prometheus: %v", err)
						return false, nil
					}
					if v != nil {
						vectorVal := v.(model.Vector)
						if len(vectorVal) == 1 && vectorVal[0].Value == expectedCount {
							tlog.Logf(t, "got expected wasm_vm_count value: %v", vectorVal[0].Value)
							return true, nil
						} else {
							tlog.Logf(t, "got metric: %v", vectorVal)
						}
					}

					return false, nil
				}); err != nil {
				t.Errorf("failed to get expected wasm_vm_count metric: %v", err)
			}
		})
	},
}

func testWasmHTTPCodeSource(t *testing.T, suite *suite.ConformanceTestSuite, route, eep, path string) {
	ns := "gateway-conformance-infra"
	routeNN := types.NamespacedName{Name: route, Namespace: ns}
	gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
	gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

	ancestorRef := gwapiv1.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}
	EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: eep, Namespace: ns}, suite.ControllerName, ancestorRef)

	expectedResponse := http.ExpectedResponse{
		Request: http.Request{
			Host: "www.example.com",
			Path: path,
		},

		// Set the expected request properties to empty strings.
		// This is a workaround to avoid the test failure.
		// These values can't be extracted from the json format response
		// body because the test wasm code appends a "Hello, world" text
		// to the response body, invalidating the json format.
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
				"x-wasm-custom": "FOO", // response header added by wasm
			},
		},
	}

	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
}
