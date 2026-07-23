// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
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

// maxExpectedLuaVMCount bounds the "lua.lua_vm_count" gauge (added in
// https://github.com/envoyproxy/envoy/pull/45871) for this test's fixture. lua-http.yaml
// configures 4 distinct Lua scripts (gateway-level, route1, route2, route4), and each
// configured script accounts for at most (concurrency + 1) VMs. This is a generous cap on
// concurrency to catch a real blow-up (e.g. a VM created per-route or per-request) without
// being sensitive to the worker thread count of the CI machine.
const maxExpectedLuaVMCount = 4 * 33

func init() {
	ConformanceTests = append(ConformanceTests, HTTPLuaTest)
}

// HTTPLuaTest tests Lua extension for a http route with HTTP Lua configured.
var HTTPLuaTest = suite.ConformanceTest{
	ShortName:   "LuaHTTP",
	Description: "Test Lua extension that adds response headers",
	Manifests:   []string{"testdata/lua-http.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with lua filter 1", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "example-route-1-with-lua", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "example-lua-1", Namespace: ns}, suite.ControllerName, ancestorRef)
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "example-lua-2", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/route1",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"X-Custom-Lua-Header": "lua_value_1",
					},
					AbsentHeaders: []string{
						"X-Custom-Response-Header", // gateway policy never took effect
						"X-Custom-Lua-Another-Header",
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("http route with lua filter 2", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "example-route-2-with-lua", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "example-lua-1", Namespace: ns}, suite.ControllerName, ancestorRef)
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "example-lua-2", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/route2",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"X-Custom-Lua-Header":         "lua_value_2",
						"X-Custom-Lua-Another-Header": "lua_another_value",
					},
					AbsentHeaders: []string{
						"X-Custom-Response-Header", // gateway policy never took effect
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("http route fallback to gateway policy", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "example-route-3-without-lua", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "example-lua-1", Namespace: ns}, suite.ControllerName, ancestorRef)
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "example-lua-2", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/route3",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"X-Custom-Response-Header": "gateway", // fallback to gateway policy
					},
					AbsentHeaders: []string{
						"X-Custom-Lua-Header", "X-Custom-Lua-Another-Header",
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("http route with lua filter context", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "example-route-4-with-lua-filter-context", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "example-lua-3-filter-context", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/route4",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"X-Lua-Filter-Context": "hello_from_filter_context",
					},
					AbsentHeaders: []string{
						"X-Custom-Response-Header", // gateway policy never took effect
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("http route without lua filter", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "example-route-3-without-lua", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "example-lua-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/route3",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					AbsentHeaders: []string{
						"X-Custom-Response-Header", // no policy for all-namespaces gatweway
						"X-Custom-Lua-Header", "X-Custom-Lua-Another-Header",
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// Regression check for https://github.com/envoyproxy/envoy/issues/9355-style VM blow-ups:
		// the fixture above configures 4 distinct Lua scripts (gateway-level + 3 route-level),
		// so the "lua.lua_vm_count" gauge should settle at a small, bounded value rather than
		// growing per-route or per-request.
		t.Run("lua vm count stays bounded", func(t *testing.T) {
			// Sum across all "same-namespace" proxy replicas/pods so multi-replica setups
			// don't produce more than one time series.
			promQL := `sum(envoy_lua_lua_vm_count{app_kubernetes_io_component="proxy", app_kubernetes_io_managed_by="envoy-gateway", app_kubernetes_io_name="envoy", gateway_envoyproxy_io_owning_gateway_name="same-namespace"})`

			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					v, err := prometheus.QueryPrometheus(suite.Client, promQL)
					if err != nil {
						tlog.Logf(t, "failed to query prometheus: %v", err)
						return false, nil
					}
					if v != nil && v.Type() == model.ValVector {
						vectorVal := v.(model.Vector)
						// Wait for the gauge to appear (present and non-zero) before judging it,
						// since it's only populated once Envoy has loaded the Lua filter config.
						if len(vectorVal) == 1 && vectorVal[0].Value > 0 {
							tlog.Logf(t, "got lua_vm_count value: %v", vectorVal[0].Value)
							if vectorVal[0].Value > maxExpectedLuaVMCount {
								// Fail outright instead of retrying: once the gauge is reporting,
								// exceeding the bound means VMs are leaking/duplicating, and more
								// polling won't make that false.
								return false, fmt.Errorf("lua_vm_count %v exceeds expected bound %d", vectorVal[0].Value, maxExpectedLuaVMCount)
							}
							return true, nil
						}
					}
					return false, nil
				}); err != nil {
				t.Errorf("failed to get expected lua_vm_count metric: %v", err)
			}
		})
	},
}
