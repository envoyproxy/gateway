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
	ConformanceTests = append(ConformanceTests, GRPCExtAuthTimeoutTest)
}

// GRPCExtAuthTimeoutTest tests ExtAuth timeout behavior for gRPC auth service.
// This test verifies that when a very short timeout is configured, the auth service
// times out and returns a 403 response.
var GRPCExtAuthTimeoutTest = suite.ConformanceTest{
	ShortName:   "GRPCExtAuthTimeout",
	Description: "Test ExtAuth timeout behavior with GRPC auth service",
	Manifests:   []string{"testdata/ext-auth-service.yaml", "testdata/ext-auth-grpc-timeout-securitypolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns}, suite.ControllerName, ancestorRef)
		// Wait for the grpc ext auth service pod to be ready
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "envoy-ext-auth"}, corev1.PodRunning, &PodReady)

		t.Run("ext auth with timeout behavior", func(t *testing.T) {
			// With 0ms timeout, the auth service should timeout immediately
			// This should result in a 403 response due to timeout
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "Bearer token1",
					},
				},
				Response: http.Response{
					StatusCode: 403,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
