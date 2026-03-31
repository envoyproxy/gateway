// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from upstream gateway-api, it will be moved to upstream.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	gatewayapi "github.com/envoyproxy/gateway/internal/gatewayapi"
)

var (
	InternetGCTests []suite.ConformanceTest
	PrivateGCTests  []suite.ConformanceTest
)

func init() {
	MultipleGCTests = make(map[string][]suite.ConformanceTest)
	InternetGCTests = append(InternetGCTests, InternetGCTest, HTTPRouteStatusAggregatesAcrossGatewayClassesTest, PolicyStatusAggregatesAcrossGatewayClassesTest)
	PrivateGCTests = append(PrivateGCTests, PrivateGCTest)
	MultipleGCTests["internet"] = InternetGCTests
	MultipleGCTests["private"] = PrivateGCTests
}

var InternetGCTest = suite.ConformanceTest{
	ShortName:   "InternetGC",
	Description: "Testing multiple GatewayClass with the same controller",
	Manifests:   []string{"testdata/internet-gc.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("internet gc", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "internet-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "internet-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
		})
	},
}

var PrivateGCTest = suite.ConformanceTest{
	ShortName:   "PrivateGC",
	Description: "Testing multiple GatewayClass with the same controller",
	Manifests:   []string{"testdata/private-gc.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("private gc", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "private-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "private-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
		})
	},
}

var HTTPRouteStatusAggregatesAcrossGatewayClassesTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteStatusAggregatesAcrossGatewayClasses",
	Description: "HTTPRoute status should aggregate parents across multiple GatewayClasses managed by the same controller",
	Manifests:   []string{"testdata/httproute-status-multiple-gc.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("httproute status aggregates across gateway classes", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "multiple-gc-route", Namespace: ns}
			internetGatewayNN := types.NamespacedName{Name: "internet-gateway", Namespace: ns}
			privateGatewayNN := types.NamespacedName{Name: "private-gateway-for-hr-status", Namespace: ns}

			_, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(internetGatewayNN))
			if err != nil {
				t.Fatalf("failed to get %s Gateway address: %v", internetGatewayNN.Name, err)
			}

			_, err = kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(privateGatewayNN))
			if err != nil {
				t.Fatalf("failed to get %s Gateway address: %v", privateGatewayNN.Name, err)
			}

			parents := []gwapiv1.RouteParentStatus{
				createGatewayParent(suite.ControllerName, internetGatewayNN.Name, "http"),
				createGatewayParent(suite.ControllerName, privateGatewayNN.Name, "http"),
			}

			kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.HTTPRoute{})
		})
	},
}

var PolicyStatusAggregatesAcrossGatewayClassesTest = suite.ConformanceTest{
	ShortName:   "PolicyStatusAggregatesAcrossGatewayClasses",
	Description: "BackendTrafficPolicy status should aggregate ancestors across multiple GatewayClasses managed by the same controller",
	Manifests:   []string{"testdata/policy-status-multiple-gc.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("backendtrafficpolicy status aggregates across gateway classes", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			policyNN := types.NamespacedName{Name: "multiple-gc-btp", Namespace: ns}
			internetGatewayNN := types.NamespacedName{Name: "internet-gateway", Namespace: ns}
			privateGatewayNN := types.NamespacedName{Name: "private-gateway-for-pol-status", Namespace: ns}

			_, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(internetGatewayNN))
			if err != nil {
				t.Fatalf("failed to get %s Gateway address: %v", internetGatewayNN.Name, err)
			}

			_, err = kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(privateGatewayNN))
			if err != nil {
				t.Fatalf("failed to get %s Gateway address: %v", privateGatewayNN.Name, err)
			}

			internetAncestorRef := createGatewayPolicyAncestorRef(internetGatewayNN.Name, "http")
			privateAncestorRef := createGatewayPolicyAncestorRef(privateGatewayNN.Name, "http")

			waitErr := wait.PollUntilContextTimeout(context.Background(), time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
				policy := &egv1a1.BackendTrafficPolicy{}
				if err := suite.Client.Get(ctx, policyNN, policy); err != nil {
					return false, err
				}

				return policyAcceptedByAncestor(policy.Status.Ancestors, suite.ControllerName, internetAncestorRef) &&
					policyAcceptedByAncestor(policy.Status.Ancestors, suite.ControllerName, privateAncestorRef), nil
			})

			require.NoErrorf(t, waitErr, "error waiting for BackendTrafficPolicy status ancestors to aggregate across gateway classes")
		})
	},
}

func createGatewayPolicyAncestorRef(gatewayName, sectionName string) gwapiv1.ParentReference {
	return gwapiv1.ParentReference{
		Group:       gatewayapi.GroupPtr(gwapiv1.GroupVersion.Group),
		Kind:        gatewayapi.KindPtr("Gateway"),
		Name:        gwapiv1.ObjectName(gatewayName),
		Namespace:   gatewayapi.NamespacePtr("gateway-conformance-infra"),
		SectionName: gatewayapi.SectionNamePtr(sectionName),
	}
}

func createGatewayParent(controllerName, gatewayName, sectionName string) gwapiv1.RouteParentStatus {
	return gwapiv1.RouteParentStatus{
		ParentRef:      createGatewayPolicyAncestorRef(gatewayName, sectionName),
		ControllerName: gwapiv1.GatewayController(controllerName),
		Conditions: []metav1.Condition{
			{
				Type:   string(gwapiv1.RouteConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.RouteReasonAccepted),
			},
			{
				Type:   string(gwapiv1.RouteConditionResolvedRefs),
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.RouteReasonResolvedRefs),
			},
		},
	}
}
