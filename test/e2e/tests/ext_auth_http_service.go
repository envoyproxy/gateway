// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPExtAuthTest)
}

// HTTPExtAuthTest tests ExtAuth authentication for an http route with ExtAuth configured.
// The http route points to an application to verify that ExtAuth authentication works on application/http path level.
// The ExtAuth service is an HTTP service.
var HTTPExtAuthTest = suite.ConformanceTest{
	ShortName:   "HTTPExtAuth",
	Description: "Test ExtAuth authentication",
	Manifests:   []string{"testdata/ext-auth-http-service.yaml", "testdata/ext-auth-http-securitypolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with ext auth authentication", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the http ext auth service pod to be ready
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "http-ext-auth"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "Bearer token1",
					},
				},
				// Verify that the http headers returned by the ext auth service
				// are added to the original request before sending it to the backend
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "www.example.com",
						Path: "/myapp",
						Headers: map[string]string{
							"x-current-user": "user1",
						},
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("without Authorization header", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the http ext auth service pod to be ready
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "http-ext-auth"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
				},
				Response: http.Response{
					StatusCode: 403,
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

		t.Run("invalid credential", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the http ext auth service pod to be ready
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "http-ext-auth"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "Bearer invalid-token",
					},
				},
				Response: http.Response{
					StatusCode: 403,
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

		t.Run("http route without ext auth authentication", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the http ext auth service pod to be ready
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "http-ext-auth"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/public",
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
	},
}
