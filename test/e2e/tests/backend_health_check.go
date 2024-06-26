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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendHealthCheckActiveHTTPTest)
}

var BackendHealthCheckActiveHTTPTest = suite.ConformanceTest{
	ShortName:   "BackendHealthCheckActiveHTTP",
	Description: "Resource with BackendHealthCheckActiveHTTP enabled",
	Manifests:   []string{"testdata/backend-health-check-active-http.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("active http", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			route1NN := types.NamespacedName{Name: "http-with-health-check-active-http-pass", Namespace: ns}
			route2NN := types.NamespacedName{Name: "http-with-health-check-active-http-fail", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), route1NN, route2NN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "health-check-active-http-pass-btp", Namespace: ns}, suite.ControllerName, ancestorRef)
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "health-check-active-http-fail-btp", Namespace: ns}, suite.ControllerName, ancestorRef)

			t.Run("health check pass", func(t *testing.T) {
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path: "/health-check-active-http-pass",
					},
					Response: http.Response{
						StatusCode: 200,
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})

			t.Run("health check fail", func(t *testing.T) {
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path: "/health-check-active-http-fail",
					},
					Response: http.Response{
						StatusCode: 503,
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})
		})
	},
}
