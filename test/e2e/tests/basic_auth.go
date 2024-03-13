// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests, BasicAuthTest)
}

var BasicAuthTest = suite.ConformanceTest{
	ShortName:   "BasicAuth",
	Description: "Resource with BasicAuth enabled",
	Manifests:   []string{"testdata/basic-auth.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("valid username password", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjE6dGVzdDE=", // user1:test1
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("without Authorization header", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
				},
				Response: http.Response{
					StatusCode: 401,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("invalid username password", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjE6dGVzdDI=", // user1:test2
					},
				},
				Response: http.Response{
					StatusCode: 401,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("per route configuration second route", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-2", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-2", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-2",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjQ6dGVzdDQ=", // user4:test4
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		// https://github.com/envoyproxy/gateway/issues/2507
		t.Run("request without matching routes ", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/not-matching-route",
				},
				Response: http.Response{
					StatusCode: 404,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})
	},
}
