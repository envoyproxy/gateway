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
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyExtensionPolicyListenerSetTest)
}

var EnvoyExtensionPolicyListenerSetTest = suite.ConformanceTest{
	ShortName:   "EnvoyExtensionPolicyListenerSet",
	Description: "EnvoyExtensionPolicy targeting ListenerSets, ListenerSet listeners, and ListenerSet-attached routes",
	Manifests:   []string{"testdata/ext-proc-service.yaml", "testdata/envoyextensionpolicy-listenerset.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

		WaitForPods(t, suite.Client, ns, map[string]string{"app": "grpc-ext-proc"}, corev1.PodRunning, &podReady)

		t.Run("listenerset policy only applies to listenerset listeners", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "eep-listenerset-policy", Namespace: ns}
			lsNN := types.NamespacedName{Name: "eep-listenerset-policy-ls", Namespace: ns}

			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				types.NamespacedName{Name: "eep-listenerset-gateway-route", Namespace: ns},
			)

			routeParents := []gwapiv1.RouteParentStatus{
				createListenerSetParent(suite.ControllerName, lsNN.Name, "ls-http"),
			}
			kubernetes.RouteMustHaveParents(
				t,
				suite.Client,
				suite.TimeoutConfig,
				types.NamespacedName{Name: "eep-listenerset-policy-route", Namespace: ns},
				routeParents,
				false,
				&gwapiv1.HTTPRoute{},
			)

			EnvoyExtensionPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "eep-listenerset-policy", Namespace: ns},
				suite.ControllerName,
				listenerSetPolicyAncestor(ns, lsNN.Name, ""),
			)

			lsAddr := getListenerAddr(gwAddr, "18191")
			expectNoExtProcHeaders(t, suite, gwAddr, "/gateway-policy", ns)
			expectExtProcHeaders(t, suite, lsAddr, "/listenerset-policy", ns)
		})

		t.Run("listenerset listener policy only applies to the selected listener", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "eep-listener-policy-gateway", Namespace: ns}
			lsNN := types.NamespacedName{Name: "eep-listener-policy-ls", Namespace: ns}

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
			require.NoError(t, err)

			kubernetes.RouteMustHaveParents(
				t,
				suite.Client,
				suite.TimeoutConfig,
				types.NamespacedName{Name: "eep-listener-policy-route", Namespace: ns},
				[]gwapiv1.RouteParentStatus{
					createListenerSetParent(suite.ControllerName, lsNN.Name, "section-http"),
				},
				false,
				&gwapiv1.HTTPRoute{},
			)
			kubernetes.RouteMustHaveParents(
				t,
				suite.Client,
				suite.TimeoutConfig,
				types.NamespacedName{Name: "eep-listener-policy-other-route", Namespace: ns},
				[]gwapiv1.RouteParentStatus{
					createListenerSetParent(suite.ControllerName, lsNN.Name, "other-http"),
				},
				false,
				&gwapiv1.HTTPRoute{},
			)

			EnvoyExtensionPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "eep-listener-policy", Namespace: ns},
				suite.ControllerName,
				listenerSetPolicyAncestor(ns, lsNN.Name, "section-http"),
			)

			listenerAddr := getListenerAddr(gwAddr, "18193")
			otherListenerAddr := getListenerAddr(gwAddr, "18194")
			expectExtProcHeaders(t, suite, listenerAddr, "/listener-policy", ns)
			expectNoExtProcHeaders(t, suite, otherListenerAddr, "/listener-policy-other", ns)
		})

		t.Run("route policy attached through listenerset applies", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "eep-route-policy-gateway", Namespace: ns}
			lsNN := types.NamespacedName{Name: "eep-route-policy-ls", Namespace: ns}
			routeNN := types.NamespacedName{Name: "eep-route-policy-route", Namespace: ns}

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
			require.NoError(t, err)

			routeParents := []gwapiv1.RouteParentStatus{
				createListenerSetParent(suite.ControllerName, lsNN.Name, "route-http"),
			}
			kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, routeParents, false, &gwapiv1.HTTPRoute{})

			EnvoyExtensionPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "eep-route-policy", Namespace: ns},
				suite.ControllerName,
				listenerSetPolicyAncestor(ns, lsNN.Name, "route-http"),
			)

			routeAddr := getListenerAddr(gwAddr, "18196")
			expectExtProcHeaders(t, suite, routeAddr, "/route-policy", ns)
		})
	},
}

func expectExtProcHeaders(t *testing.T, suite *suite.ConformanceTestSuite, addr, path, namespace string) {
	t.Helper()

	expectedResponse := httputils.ExpectedResponse{
		Request: httputils.Request{
			Path: path,
		},
		ExpectedRequest: &httputils.ExpectedRequest{
			Request: httputils.Request{
				Path: path,
				Headers: map[string]string{
					"x-request-ext-processed": "true",
				},
			},
		},
		Response: httputils.Response{
			StatusCodes: []int{200},
			Headers: map[string]string{
				"x-response-ext-processed": "true",
			},
		},
		Namespace: namespace,
	}

	httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, addr, expectedResponse)
}

func expectNoExtProcHeaders(t *testing.T, suite *suite.ConformanceTestSuite, addr, path, namespace string) {
	t.Helper()

	expectedResponse := httputils.ExpectedResponse{
		Request: httputils.Request{
			Path: path,
		},
		ExpectedRequest: &httputils.ExpectedRequest{
			Request: httputils.Request{
				Path: path,
			},
		},
		Response: httputils.Response{
			StatusCodes:   []int{200},
			AbsentHeaders: []string{"x-response-ext-processed"},
		},
		Namespace: namespace,
	}

	httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, addr, expectedResponse)
}
