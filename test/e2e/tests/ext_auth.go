// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwhttp "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, ExtAuthTest)
}

// ExtAuthTest tests ExtAuth authentication for an http route with ExtAuth configured.
// The http route points to an application to verify that ExtAuth authentication works on application/http path level.
var ExtAuthTest = suite.ConformanceTest{
	ShortName:   "ExtAuth",
	Description: "Test ExtAuth authentication",
	Manifests:   []string{"testdata/ext-auth-http-service.yaml", "testdata/ext-auth-securitypolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with ext auth authentication", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			securityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns})
			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the keycloak pod to be configured with the test user and client
			waitForPods(t, suite.Client, ns, map[string]string{"app": "http-ext-auth"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "dG9rZW4x", // token1 (base64 encoded)
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
			routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			securityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns})
			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the keycloak pod to be configured with the test user and client
			waitForPods(t, suite.Client, ns, map[string]string{"app": "http-ext-auth"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
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

		t.Run("invalid credential", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			securityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns})
			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the keycloak pod to be configured with the test user and client
			waitForPods(t, suite.Client, ns, map[string]string{"app": "http-ext-auth"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "dG9rZW40", // token4 (base64 encoded)
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

		t.Run("http route without ext auth authentication", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			securityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns})
			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the keycloak pod to be configured with the test user and client
			waitForPods(t, suite.Client, ns, map[string]string{"app": "http-ext-auth"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/public",
				},
				Response: gwhttp.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := gwhttp.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := gwhttp.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})
	},
}
