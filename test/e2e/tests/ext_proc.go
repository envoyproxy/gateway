// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, ExtProcTest)
}

// ExtProcTest tests ExtProc for http routes and verifies %EG_EXT_PROC_FILTER_STATE% access log expansion.
var ExtProcTest = suite.ConformanceTest{
	ShortName:   "ExtProc",
	Description: "Test ExtProc service that adds request and response headers, and access log operator resolution",
	Manifests:   []string{"testdata/ext-proc-service.yaml", "testdata/ext-proc-envoyextensionpolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		gatewayNS := GetGatewayResourceNamespace()

		podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "grpc-ext-proc"}, corev1.PodRunning, &podReady)

		t.Run("http route with ext proc", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-with-ext-proc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "www.example.com",
					Path: "/processor",
					Headers: map[string]string{
						"x-request-client-header": "original",
					},
				},
				ExpectedRequest: &httputils.ExpectedRequest{
					Request: httputils.Request{
						Path: "/processor",
						Headers: map[string]string{
							"x-request-ext-processed":          "true",
							"x-request-client-header-received": "original",
							"x-request-client-header":          "mutated",
							"x-request-xds-route-name":         "httproute/gateway-conformance-infra/http-with-ext-proc/rule/0/match/0/www_example_com",
							"x-request-from-ext-proc-metadata": "received",
						},
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-response-ext-processed":        "true",
						"x-response-xds-route-name":       "httproute/gateway-conformance-infra/http-with-ext-proc/rule/0/match/0/www_example_com",
						"x-response-rbac-result-metadata": "allowed",
					},
				},
				Namespace: ns,
			}

			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-test-route-rule", Namespace: ns}, suite.ControllerName, ancestorRef)
			expectedResponse.Request.Path = "/processor-route-rule-name"
			expectedResponse.ExpectedRequest.Path = "/processor-route-rule-name"
			expectedResponse.ExpectedRequest.Headers["x-request-xds-route-name"] = "httproute/gateway-conformance-infra/http-with-ext-proc/rule/1/match/0/www_example_com"
			expectedResponse.Response.Headers["x-response-xds-route-name"] = "httproute/gateway-conformance-infra/http-with-ext-proc/rule/1/match/0/www_example_com"
			expectedResponse.Response.Headers["x-response-request-path"] = "/processor-route-rule-name"

			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("http route without proc mode", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-without-procmode", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-no-procmode-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "www.example.com",
					Path: "/no-processor",
					Headers: map[string]string{
						"x-request-client-header": "original",
					},
				},
				ExpectedRequest: &httputils.ExpectedRequest{
					Request: httputils.Request{
						Path: "/no-processor",
						Headers: map[string]string{
							"x-request-client-header": "original",
						},
					},
				},
				Response: httputils.Response{
					StatusCodes:   []int{200},
					AbsentHeaders: []string{"x-response-ext-processed"},
				},
				Namespace: ns,
			}

			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("http route with uds ext proc", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-with-ext-proc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-uds-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "www.example.com",
					Path: "/uds-processor",
					Headers: map[string]string{
						"x-request-client-header": "original",
					},
				},
				ExpectedRequest: &httputils.ExpectedRequest{
					Request: httputils.Request{
						Path: "/uds-processor",
						Headers: map[string]string{
							"x-request-ext-processed":          "true",
							"x-request-client-header-received": "original",
							"x-request-client-header":          "mutated",
						},
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-response-ext-processed": "true",
					},
				},
				Namespace: ns,
			}

			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// File sink logs land in the envoy container stdout which fluent-bit collects.
		fileLogLabels := map[string]string{
			"job":       fmt.Sprintf("%s/envoy", gatewayNS),
			"namespace": gatewayNS,
			"container": "envoy",
		}

		t.Run("access log: straightforward expansion", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "ext-proc-accesslog-route1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "ext-proc-accesslog-gtw", Namespace: ns}
			gwHost := kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
			gwAddr := net.JoinHostPort(gwHost, "8083")

			ancestorRef := gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindGateway),
				Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:        gwapiv1.ObjectName(gwNN.Name),
				SectionName: gatewayapi.SectionNamePtr("http-8083"),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep1", Namespace: ns},
				suite.ControllerName, ancestorRef)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "accesslog.example.com",
					Path: "/route1",
				},
				Response:  httputils.Response{StatusCodes: []int{200}},
				Namespace: ns,
			}
			runLogTest(t, suite, gwAddr, &expectedResponse, fileLogLabels, "ep_bs=[0-9]+ .*path=/route1", 1)
		})

		// route2 + eep2 are applied explicitly after eep1 is accepted to guarantee
		// eep1 holds the name claim and eep2 receives AmbiguousDefinition.
		t.Run("access log: first-match on shared name", func(t *testing.T) {
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig,
				"testdata/ext-proc-accesslog-ambiguous.yaml", true)

			route1NN := types.NamespacedName{Name: "ext-proc-accesslog-route1", Namespace: ns}
			route2NN := types.NamespacedName{Name: "ext-proc-accesslog-route2", Namespace: ns}
			gwNN := types.NamespacedName{Name: "ext-proc-accesslog-gtw", Namespace: ns}
			gwHost := kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false,
				route1NN, route2NN)
			gwAddr := net.JoinHostPort(gwHost, "8083")

			ancestorRef := gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindGateway),
				Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:        gwapiv1.ObjectName(gwNN.Name),
				SectionName: gatewayapi.SectionNamePtr("http-8083"),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep1", Namespace: ns},
				suite.ControllerName, ancestorRef)
			EnvoyExtensionPolicyMustHaveWarning(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep2", Namespace: ns},
				suite.ControllerName, ancestorRef, "auth-proc")

			for path, pattern := range map[string]string{
				"/route1": "ep_bs=[0-9]+ .*path=/route1",
				"/route2": "ep_bs=- .*path=/route2",
			} {
				resp := httputils.ExpectedResponse{
					Request:   httputils.Request{Host: "accesslog.example.com", Path: path},
					Response:  httputils.Response{StatusCodes: []int{200}},
					Namespace: ns,
				}
				runLogTest(t, suite, gwAddr, &resp, fileLogLabels, pattern, 1)
			}
		})

		t.Run("access log: filter-chain isolation", func(t *testing.T) {
			routeANN := types.NamespacedName{Name: "ext-proc-accesslog-route-a", Namespace: ns}
			routeBNN := types.NamespacedName{Name: "ext-proc-accesslog-route-b", Namespace: ns}
			gwNN := types.NamespacedName{Name: "ext-proc-accesslog-gtw", Namespace: ns}

			gwAddrA := net.JoinHostPort(kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeANN), "8084")
			gwAddrB := net.JoinHostPort(kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeBNN), "8085")

			ancestorRefA := gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindGateway),
				Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:        gwapiv1.ObjectName(gwNN.Name),
				SectionName: gatewayapi.SectionNamePtr("http-8084"),
			}
			ancestorRefB := gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindGateway),
				Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:        gwapiv1.ObjectName(gwNN.Name),
				SectionName: gatewayapi.SectionNamePtr("http-8085"),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep-a", Namespace: ns},
				suite.ControllerName, ancestorRefA)
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep-b", Namespace: ns},
				suite.ControllerName, ancestorRefB)

			respA := httputils.ExpectedResponse{
				Request:   httputils.Request{Host: "accesslog-iso.example.com", Path: "/isolated-a"},
				Response:  httputils.Response{StatusCodes: []int{200}},
				Namespace: ns,
			}
			runLogTest(t, suite, gwAddrA, &respA, fileLogLabels, "ep_bs_iso=[0-9]+ .*path=/isolated-a", 1)

			respB := httputils.ExpectedResponse{
				Request:   httputils.Request{Host: "accesslog-iso.example.com", Path: "/isolated-b"},
				Response:  httputils.Response{StatusCodes: []int{200}},
				Namespace: ns,
			}
			runLogTest(t, suite, gwAddrB, &respB, fileLogLabels, "ep_bs_iso=[0-9]+ .*path=/isolated-b", 1)
		})

		t.Run("metrics: shared statPrefix aggregates across listeners", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "ext-proc-accesslog-gtw", Namespace: ns}

			// Both eep-a and eep-b share statPrefix "isolated-proc" across separate listeners.
			// Each listener increments streams_started independently; combined total must be ≥ 2,
			// confirming the shared statPrefix aggregates across filter-chain-isolated deployments.
			promClient, err := prometheus.NewClient(suite.Client,
				types.NamespacedName{Name: "prometheus", Namespace: "monitoring"})
			if err != nil {
				t.Fatalf("failed to create prometheus client: %v", err)
			}
			pql := fmt.Sprintf(
				`sum(envoy_http_ext_proc_isolated_proc_streams_started{gateway_envoyproxy_io_owning_gateway_name=%q})`,
				gwNN.Name)
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					val, err := promClient.QuerySum(ctx, pql)
					if err != nil {
						tlog.Logf(t, "isolated-proc streams_started not yet available: %v", err)
						return false, nil
					}
					tlog.Logf(t, "isolated-proc streams_started=%.0f", val)
					return val >= 2, nil
				}); err != nil {
				t.Errorf("isolated-proc streams_started did not reach 2: %v", err)
			}
		})
	},
}
