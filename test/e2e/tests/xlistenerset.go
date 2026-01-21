// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	gatewayapi "github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, XListenerSetTest, XListenerSetGRPCTest)
}

var XListenerSetTest = suite.ConformanceTest{
	ShortName:   "XListenerSet",
	Description: "HTTPRoute should attach to an XListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/xlistenerset.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "xlistener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "xlistener-httproute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		hostOnly := gwAddrWithPort
		if host, _, splitErr := net.SplitHostPort(gwAddrWithPort); splitErr == nil {
			hostOnly = host
		}
		listenerAddr := net.JoinHostPort(hostOnly, "18081")

		parents := []gwapiv1.RouteParentStatus{{
			ParentRef: gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindXListenerSet),
				Name:        gwapiv1.ObjectName("xlistener-set"),
				Namespace:   gatewayapi.NamespacePtr(ns),
				SectionName: gatewayapi.SectionNamePtr("extra-http"),
			},
			ControllerName: gwapiv1.GatewayController(suite.ControllerName),
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
		}}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.HTTPRoute{})

		expected := http.ExpectedResponse{
			Request: http.Request{
				Path: "/xlistener",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, listenerAddr, expected)
	},
}

var XListenerSetGRPCTest = suite.ConformanceTest{
	ShortName:   "XListenerSetGRPC",
	Description: "GRPCRoute should attach to an XListenerSet listener and serve traffic",
	Manifests: []string{
		"testdata/xlistenerset.yaml",
		"testdata/xlistenerset-grpc.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "xlistener-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "xlistener-grpcroute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		require.NoError(t, err)

		hostOnly := gwAddrWithPort
		if host, _, splitErr := net.SplitHostPort(gwAddrWithPort); splitErr == nil {
			hostOnly = host
		}
		listenerAddr := net.JoinHostPort(hostOnly, "18082")

		parents := []gwapiv1.RouteParentStatus{{
			ParentRef: gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindXListenerSet),
				Name:        gwapiv1.ObjectName("xlistener-set"),
				Namespace:   gatewayapi.NamespacePtr(ns),
				SectionName: gatewayapi.SectionNamePtr("extra-grpc"),
			},
			ControllerName: gwapiv1.GatewayController(suite.ControllerName),
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
		}}

		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, false, &gwapiv1.GRPCRoute{})

		expected := grpc.ExpectedResponse{
			EchoRequest: &grpcechoserver.EchoRequest{},
			Backend:     "grpc-xlistener-backend",
			Namespace:   ns,
		}

		grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, listenerAddr, expected)
	},
}
