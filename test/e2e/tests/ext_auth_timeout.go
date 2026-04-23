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
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, ExtAuthTimeoutTest)
}

// ExtAuthTimeoutTest tests ExtAuth timeout behavior for gRPC/HTTP auth service.
// This test verifies that when a very short timeout is configured, the auth service
// times out and returns a 403 response.
var ExtAuthTimeoutTest = suite.ConformanceTest{
	ShortName:   "ExtAuthTimeout",
	Description: "Test ExtAuth timeout behavior with auth service",
	Manifests: []string{
		"testdata/ext-auth-service-with-delay.yaml",
		"testdata/ext-auth-grpc-timeout-securitypolicy.yaml",
		"testdata/ext-auth-http-timeout-securitypolicy.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		httpRouteNN := types.NamespacedName{Name: "http-with-http-ext-auth", Namespace: ns}
		grpcRouteNN := types.NamespacedName{Name: "http-with-grpc-ext-auth", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false,
			httpRouteNN, grpcRouteNN)
		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "grpc-ext-auth", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "http-ext-auth", Namespace: ns}, suite.ControllerName, ancestorRef)
		// Wait for the envoy ext auth service pod to be ready
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "envoy-ext-auth"}, corev1.PodRunning, &PodReady)

		for _, tc := range []struct {
			name string
			path string
		}{
			{
				name: "GRPC",
				path: "/grpc",
			},
			{
				name: "HTTP",
				path: "/http",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				// With timeout, the auth service should timeout immediately
				// This should result in a 403 response due to timeout
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Host: "www.example.com",
						Path: tc.path,
						Headers: map[string]string{
							"Authorization": "Bearer token1",
						},
					},
					Response: http.Response{
						StatusCodes: []int{403},
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})
		}
	},
}
