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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, ExtProcTest)
}

// ExtProcTest tests ExtProc authentication for an http route with ExtProc configured.
var ExtProcTest = suite.ConformanceTest{
	ShortName:   "ExtProc",
	Description: "Test ExtProc service that adds request and response headers",
	Manifests:   []string{"testdata/ext-proc-service.yaml", "testdata/ext-proc-envoyextensionpolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with ext proc", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-proc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the grpc ext auth service pod to be ready
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "grpc-ext-proc"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/processor",
					Headers: map[string]string{
						"x-request-client-header": "original", // add a request header that will be mutated by ext-proc
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/processor",
						Headers: map[string]string{
							// header added by ext-processor to backend-bound request
							"x-request-ext-processed": "true",
							// this is the original client header preserved by ext-proc in a new header
							"x-request-client-header-received": "original",
							// this is the mutated value expected to reach upstream
							"x-request-client-header": "mutated",
							// header added by ext-processor to request based on the xds.route_name attribute
							"x-request-xds-route-name": "httproute/gateway-conformance-infra/http-with-ext-proc/rule/0/match/0/www_example_com",
							// header added by router based on metadata emitted by the external processor to
							// the io.envoyproxy.gateway.e2e namespace with key ext-proc-emitted-metadata
							"x-request-from-ext-proc-metadata": "received",
						},
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						// header added by ext-processor to client-bound response
						"x-response-ext-processed": "true",
						// header added by ext-processor to response based on the xds.cluster_name attribute
						"x-response-xds-route-name": "httproute/gateway-conformance-infra/http-with-ext-proc/rule/0/match/0/www_example_com",
						// header added by ext-processor to response based on envoy.filters.http.rbac.enforced_engine_result
						// dynamic metadata emitted by RBAC filter
						"x-response-rbac-result-metadata": "allowed",
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// now check policy attached to a route rule which adds one additional header
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-test-route-rule", Namespace: ns}, suite.ControllerName, ancestorRef)
			expectedResponse.Request.Path = "/processor-route-rule-name"
			expectedResponse.ExpectedRequest.Path = "/processor-route-rule-name"
			expectedResponse.ExpectedRequest.Headers["x-request-xds-route-name"] = "httproute/gateway-conformance-infra/http-with-ext-proc/rule/1/match/0/www_example_com"
			expectedResponse.Response.Headers["x-response-xds-route-name"] = "httproute/gateway-conformance-infra/http-with-ext-proc/rule/1/match/0/www_example_com"
			expectedResponse.Response.Headers["x-response-request-path"] = "/processor-route-rule-name"

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("http route without proc mode", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-without-procmode", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-no-procmode-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the grpc ext auth service pod to be ready
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "grpc-ext-proc"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/no-processor",
					Headers: map[string]string{
						"x-request-client-header": "original", // add a request header that will be mutated by ext-proc
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/no-processor",
						Headers: map[string]string{
							"x-request-client-header": "original", // this is the original value expected to reach upstream
						},
					},
				},
				Response: http.Response{
					StatusCode:    200,
					AbsentHeaders: []string{"x-response-ext-processed"},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("http route with uds ext proc", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-proc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-uds-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the grpc ext auth service pod to be ready
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "grpc-ext-proc"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/uds-processor",
					Headers: map[string]string{
						"x-request-client-header": "original", // add a request header that will be mutated by ext-proc
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/uds-processor",
						Headers: map[string]string{
							"x-request-ext-processed":          "true",     // header added by ext-processor to backend-bound request
							"x-request-client-header-received": "original", // this is the original client header preserved by ext-proc in a new header
							"x-request-client-header":          "mutated",  // this is the mutated value expected to reach upstream
						},
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"x-response-ext-processed": "true", // header added by ext-processor to client-bound response
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
