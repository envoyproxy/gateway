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

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
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
						"x-request-ext-processed": "true", // header added by ext-processor to request
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"x-response-ext-processed": "true", // header added by ext-processor to request
					},
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

		t.Run("http route without ext proc", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-ext-proc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-proc-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}

			// Wait for the grpc ext auth service pod to be ready
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "grpc-ext-proc"}, corev1.PodRunning, podReady)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/no-processor",
				},
				Response: http.Response{
					StatusCode:    200,
					AbsentHeaders: []string{"x-response-ext-processed"},
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
