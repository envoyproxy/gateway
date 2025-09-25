// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPBackendExtAuthTest)
}

// HTTPBackendExtAuthTest tests ExtAuth authentication for a http route with ExtAuth configured.
// Almost like HTTPExtAuthTest, but the security policy reference to the backend service.
var HTTPBackendExtAuthTest = suite.ConformanceTest{
	ShortName:   "HTTPBackendExtAuth",
	Description: "Test HTTP ExtAuth authentication with backend",
	Manifests:   []string{"testdata/ext-auth-service.yaml", "testdata/ext-auth-http-backend-securitypolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-ext-auth-backend", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-backend", Namespace: ns}, suite.ControllerName, ancestorRef)
		// Wait for the http ext auth service pod to be ready
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "envoy-ext-auth"}, corev1.PodRunning, &PodReady)

		t.Run("http route with ext auth backend ref", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "Bearer token2",
					},
				},
				// Verify that the http headers returned by the ext auth service
				// are added to the original request before sending it to the backend
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "www.example.com",
						Path: "/myapp",
						Headers: map[string]string{
							"x-current-user": "user2",
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

		t.Run("route base on headersToBackend", func(t *testing.T) {
			v2ExpectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "Bearer token2",
					},
				},
				Backend: "infra-backend-v2",
				// Verify that the http headers returned by the ext auth service
				// are added to the original request before sending it to the backend
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "www.example.com",
						Path: "/myapp",
						Headers: map[string]string{
							"x-current-user": "user2",
						},
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, v2ExpectedResponse)

			v3ExpectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "Bearer token3",
					},
				},
				// Verify that the http headers returned by the ext auth service
				// are added to the original request before sending it to the backend
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "www.example.com",
						Path: "/myapp",
						Headers: map[string]string{
							"x-current-user": "user3",
						},
					},
				},
				Backend: "infra-backend-v3",
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, v3ExpectedResponse)
		})
	},
}
