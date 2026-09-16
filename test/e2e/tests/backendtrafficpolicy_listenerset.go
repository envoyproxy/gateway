// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTrafficPolicyListenerSetTest)
}

var BackendTrafficPolicyListenerSetTest = suite.ConformanceTest{
	ShortName:   "BackendTrafficPolicyListenerSet",
	Description: "BackendTrafficPolicy targeting ListenerSets, ListenerSet-attached routes, and ListenerSet merge parents",
	Manifests:   []string{"testdata/backendtrafficpolicy-listenerset.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		t.Run("listenerset policy only applies to listenerset listeners", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "btp-listenerset-policy", Namespace: ns}
			lsNN := types.NamespacedName{Name: "btp-listenerset-policy-ls", Namespace: ns}

			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				types.NamespacedName{Name: "btp-listenerset-gateway-route", Namespace: ns},
			)

			routeParents := []gwapiv1.RouteParentStatus{
				createListenerSetParent(suite.ControllerName, lsNN.Name, "ls-http"),
			}
			kubernetes.RouteMustHaveParents(
				t,
				suite.Client,
				suite.TimeoutConfig,
				types.NamespacedName{Name: "btp-listenerset-policy-route", Namespace: ns},
				routeParents,
				false,
				&gwapiv1.HTTPRoute{},
			)

			gwAncestorRef := gatewayPolicyAncestor(ns, gwNN.Name)
			lsAncestorRef := listenerSetPolicyAncestor(ns, lsNN.Name, "")
			BackendTrafficPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "btp-listenerset-gateway", Namespace: ns},
				suite.ControllerName,
				gwAncestorRef,
			)
			BackendTrafficPolicyMustBeOverridden(
				t,
				suite.Client,
				types.NamespacedName{Name: "btp-listenerset-gateway", Namespace: ns},
				suite.ControllerName,
				gwAncestorRef,
			)
			BackendTrafficPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "btp-listenerset-policy", Namespace: ns},
				suite.ControllerName,
				lsAncestorRef,
			)

			lsAddr := getListenerAddr(gwAddr, "18187")
			expectBackendTrafficPolicyStatus(t, suite, gwAddr, "/gateway-policy", 418)
			expectBackendTrafficPolicyStatus(t, suite, lsAddr, "/listenerset-policy", 419)
		})

		t.Run("route policy attached through listenerset applies without other policies", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "btp-route-policy-gateway", Namespace: ns}
			lsNN := types.NamespacedName{Name: "btp-route-policy-ls", Namespace: ns}
			routeNN := types.NamespacedName{Name: "btp-route-policy-route", Namespace: ns}

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
			if err != nil {
				t.Fatalf("failed to get gateway address: %v", err)
			}

			routeParents := []gwapiv1.RouteParentStatus{
				createListenerSetParent(suite.ControllerName, lsNN.Name, "route-http"),
			}
			kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, routeParents, false, &gwapiv1.HTTPRoute{})

			ancestorRef := listenerSetPolicyAncestor(ns, lsNN.Name, "route-http")
			BackendTrafficPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "btp-route-policy", Namespace: ns},
				suite.ControllerName,
				ancestorRef,
			)

			routeAddr := getListenerAddr(gwAddr, "18189")
			expectBackendTrafficPolicyStatus(t, suite, routeAddr, "/route-policy", 420)
		})

		t.Run("route policy merges with listenerset policy", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "btp-merge-policy-gateway", Namespace: ns}
			lsNN := types.NamespacedName{Name: "btp-merge-policy-ls", Namespace: ns}
			routeNN := types.NamespacedName{Name: "btp-merge-policy-route", Namespace: ns}

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
			if err != nil {
				t.Fatalf("failed to get gateway address: %v", err)
			}

			routeParents := []gwapiv1.RouteParentStatus{
				createListenerSetParent(suite.ControllerName, lsNN.Name, "merge-http"),
			}
			kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, routeParents, false, &gwapiv1.HTTPRoute{})

			lsAncestorRef := listenerSetPolicyAncestor(ns, lsNN.Name, "")
			routeAncestorRef := listenerSetPolicyAncestor(ns, lsNN.Name, "merge-http")
			BackendTrafficPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "btp-merge-listenerset", Namespace: ns},
				suite.ControllerName,
				lsAncestorRef,
			)
			BackendTrafficPolicyMustBeMerged(
				t,
				suite.Client,
				types.NamespacedName{Name: "btp-merge-listenerset", Namespace: ns},
				suite.ControllerName,
				lsAncestorRef,
			)
			BackendTrafficPolicyMustBeMerged(
				t,
				suite.Client,
				types.NamespacedName{Name: "btp-merge-route", Namespace: ns},
				suite.ControllerName,
				routeAncestorRef,
			)

			mergeAddr := getListenerAddr(gwAddr, "18191")
			expectBackendTrafficPolicyStatus(t, suite, mergeAddr, "/merge-policy", 500)
		})
	},
}

func expectBackendTrafficPolicyStatus(t *testing.T, suite *suite.ConformanceTestSuite, addr, path string, statusCode int) {
	t.Helper()

	expectedResponse := httputils.ExpectedResponse{
		Request: httputils.Request{
			Path: path,
		},
		Response: httputils.Response{
			StatusCode: statusCode,
		},
	}
	httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, addr, expectedResponse)
}
