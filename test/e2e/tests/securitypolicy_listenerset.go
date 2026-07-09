// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, SecurityPolicyListenerSetTest)
}

var SecurityPolicyListenerSetTest = suite.ConformanceTest{
	ShortName:   "SecurityPolicyListenerSet",
	Description: "SecurityPolicy targeting ListenerSets, ListenerSet-attached routes, and ListenerSet merge parents",
	Manifests:   []string{"testdata/securitypolicy-listenerset.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		t.Run("listenerset policy only applies to listenerset listeners", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "sp-listenerset-policy", Namespace: ns}
			lsNN := types.NamespacedName{Name: "sp-listenerset-policy-ls", Namespace: ns}

			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				types.NamespacedName{Name: "sp-listenerset-gateway-route", Namespace: ns},
			)

			routeParents := []gwapiv1.RouteParentStatus{
				createListenerSetParent(suite.ControllerName, lsNN.Name, "ls-http"),
			}
			kubernetes.RouteMustHaveParents(
				t,
				suite.Client,
				suite.TimeoutConfig,
				types.NamespacedName{Name: "sp-listenerset-policy-route", Namespace: ns},
				routeParents,
				false,
				&gwapiv1.HTTPRoute{},
			)

			gwAncestorRef := gatewayPolicyAncestor(ns, gwNN.Name)
			lsAncestorRef := listenerSetPolicyAncestor(ns, lsNN.Name, "")
			SecurityPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "sp-listenerset-gateway", Namespace: ns},
				suite.ControllerName,
				gwAncestorRef,
			)
			SecurityPolicyMustBeOverridden(
				t,
				suite.Client,
				types.NamespacedName{Name: "sp-listenerset-gateway", Namespace: ns},
				suite.ControllerName,
				gwAncestorRef,
			)
			SecurityPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "sp-listenerset-policy", Namespace: ns},
				suite.ControllerName,
				lsAncestorRef,
			)

			lsAddr := getListenerAddr(gwAddr, "18181")
			expectCORSPreflight(t, suite, gwAddr, "/gateway-policy", "https://gateway.example.com", "GET", "x-gateway-header")
			expectCORSPreflight(t, suite, lsAddr, "/listenerset-policy", "https://listenerset.example.com", "POST", "x-listenerset-header")
			expectNoCORSPreflight(t, suite, gwAddr, "/gateway-policy", "https://listenerset.example.com", "POST", "x-listenerset-header")
		})

		t.Run("route policy attached through listenerset applies without other policies", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "sp-route-policy-gateway", Namespace: ns}
			lsNN := types.NamespacedName{Name: "sp-route-policy-ls", Namespace: ns}
			routeNN := types.NamespacedName{Name: "sp-route-policy-route", Namespace: ns}

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
			require.NoError(t, err)

			routeParents := []gwapiv1.RouteParentStatus{
				createListenerSetParent(suite.ControllerName, lsNN.Name, "route-http"),
			}
			kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, routeParents, false, &gwapiv1.HTTPRoute{})

			ancestorRef := listenerSetPolicyAncestor(ns, lsNN.Name, "route-http")
			SecurityPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "sp-route-policy", Namespace: ns},
				suite.ControllerName,
				ancestorRef,
			)

			routeAddr := getListenerAddr(gwAddr, "18183")
			expectCORSPreflight(t, suite, routeAddr, "/route-policy", "https://route.example.com", "PUT", "x-route-header")
		})

		t.Run("route policy merges with listenerset policy", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "sp-merge-policy-gateway", Namespace: ns}
			lsNN := types.NamespacedName{Name: "sp-merge-policy-ls", Namespace: ns}
			routeNN := types.NamespacedName{Name: "sp-merge-policy-route", Namespace: ns}

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
			require.NoError(t, err)

			routeParents := []gwapiv1.RouteParentStatus{
				createListenerSetParent(suite.ControllerName, lsNN.Name, "merge-http"),
			}
			kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, routeParents, false, &gwapiv1.HTTPRoute{})

			lsAncestorRef := listenerSetPolicyAncestor(ns, lsNN.Name, "")
			routeAncestorRef := listenerSetPolicyAncestor(ns, lsNN.Name, "merge-http")
			SecurityPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "sp-merge-listenerset", Namespace: ns},
				suite.ControllerName,
				lsAncestorRef,
			)
			SecurityPolicyMustBeMerged(
				t,
				suite.Client,
				types.NamespacedName{Name: "sp-merge-listenerset", Namespace: ns},
				suite.ControllerName,
				lsAncestorRef,
			)
			SecurityPolicyMustBeMerged(
				t,
				suite.Client,
				types.NamespacedName{Name: "sp-merge-route", Namespace: ns},
				suite.ControllerName,
				routeAncestorRef,
			)

			mergeAddr := getListenerAddr(gwAddr, "18185")
			expectCORSPreflight(t, suite, mergeAddr, "/merge-policy", "https://merge-listenerset.example.com", "PATCH", "x-merge-header")
		})
	},
}

func gatewayPolicyAncestor(namespace, name string) gwapiv1.ParentReference {
	return gwapiv1.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(namespace),
		Name:      gwapiv1.ObjectName(name),
	}
}

func listenerSetPolicyAncestor(namespace, name, sectionName string) gwapiv1.ParentReference {
	ancestorRef := gwapiv1.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindListenerSet),
		Namespace: gatewayapi.NamespacePtr(namespace),
		Name:      gwapiv1.ObjectName(name),
	}
	if sectionName != "" {
		ancestorRef.SectionName = gatewayapi.SectionNamePtr(sectionName)
	}
	return ancestorRef
}

func expectCORSPreflight(
	t *testing.T,
	suite *suite.ConformanceTestSuite,
	addr, path, origin, method, requestHeaders string,
) {
	t.Helper()

	expectedResponse := httputils.ExpectedResponse{
		Request: httputils.Request{
			Path:   path,
			Method: "OPTIONS",
			Headers: map[string]string{
				"Origin":                         origin,
				"access-control-request-method":  method,
				"access-control-request-headers": requestHeaders,
			},
		},
		ExpectedRequest: &httputils.ExpectedRequest{
			Request: httputils.Request{
				Host:    "",
				Method:  "OPTIONS",
				Path:    "",
				Headers: nil,
			},
		},
		Response: httputils.Response{
			StatusCodes: []int{200},
			Headers: map[string]string{
				"access-control-allow-origin":  origin,
				"access-control-allow-methods": method,
				"access-control-allow-headers": requestHeaders,
			},
		},
		Namespace: "",
	}
	httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, addr, expectedResponse)
}

func expectNoCORSPreflight(
	t *testing.T,
	suite *suite.ConformanceTestSuite,
	addr, path, origin, method, requestHeaders string,
) {
	t.Helper()

	expectedResponse := httputils.ExpectedResponse{
		Request: httputils.Request{
			Path:   path,
			Method: "OPTIONS",
			Headers: map[string]string{
				"Origin":                         origin,
				"access-control-request-method":  method,
				"access-control-request-headers": requestHeaders,
			},
		},
		ExpectedRequest: &httputils.ExpectedRequest{
			Request: httputils.Request{
				Host:    "",
				Method:  "OPTIONS",
				Path:    "",
				Headers: nil,
			},
		},
		Response: httputils.Response{
			StatusCodes:   []int{200},
			AbsentHeaders: []string{"access-control-allow-origin"},
		},
		Namespace: "",
	}
	httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, addr, expectedResponse)
}
