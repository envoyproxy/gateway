// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		RoundRobinLoadBalancingTest,
		ConsistentHashSourceIPLoadBalancingTest,
		ConsistentHashHeaderLoadBalancingTest,
	)
}

var RoundRobinLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "Round Robin Load Balancing",
	Description: "Test for round robin load balancing type",
	Manifests:   []string{"testdata/load_balancing_round_robin.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const sendRequests = 100

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "round-robin-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "round-robin-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("traffic should be split evenly", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/round",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			trafficMap := make(map[string]int)
			for i := 0; i < sendRequests; i++ {
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Errorf("failed to get expected response: %v", err)
				}

				if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
					t.Errorf("failed to compare request and response: %v", err)
				}

				podName := cReq.Pod
				if len(podName) == 0 {
					// it shouldn't be missing here
					t.Errorf("failed to get pod header in response: %v", err)
				} else {
					trafficMap[podName]++
				}
			}

			// Expect traffic number for each endpoint.
			even := sendRequests / 4

			for podName, traffic := range trafficMap {
				if !AlmostEquals(traffic, even, 3) {
					t.Errorf("The traffic are not be split evenly for pod %s: %d", podName, traffic)
				}
			}
		})
	},
}

var ConsistentHashSourceIPLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "Source IP based Consistent Hash Load Balancing",
	Description: "Test for source IP based consistent hash load balancing type",
	Manifests:   []string{"testdata/load_balancing_consistent_hash_source_ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const sendRequests = 10

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "source-ip-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "source-ip-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("all traffics route to the same backend with same source ip", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/source",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			// Same source IP will always hit the same endpoint.
			var expectPodName string

			for i := 0; i < sendRequests; i++ {
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Errorf("failed to get expected response: %v", err)
				}

				if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
					t.Errorf("failed to compare request and response: %v", err)
				}

				podName := cReq.Pod
				if len(podName) == 0 {
					// it shouldn't be missing here
					t.Errorf("failed to get pod header in response: %v", err)
				} else {
					if len(expectPodName) == 0 {
						expectPodName = podName
					} else {
						require.Equal(t, expectPodName, podName)
					}
				}
			}
		})
	},
}

var ConsistentHashHeaderLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "Header based Consistent Hash Load Balancing",
	Description: "Test for header based consistent hash load balancing type",
	Manifests:   []string{"testdata/load_balancing_consistent_hash_header.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const sendRequests = 10

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "header-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "header-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("all traffics route to the same backend with same test header", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/header",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			headers := []string{"0.0.0.0", "1.2.3.4", "4.5.6.7", "7.8.9.10", "10.11.12.13"}

			for _, header := range headers {
				// Same test header will always hit the same endpoint.
				var expectPodName string

				for i := 0; i < sendRequests; i++ {
					req.Headers["Lb-Test-Header"] = []string{header}
					cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
					if err != nil {
						t.Errorf("failed to get expected response: %v", err)
					}

					if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
						t.Errorf("failed to compare request and response: %v", err)
					}

					podName := cReq.Pod
					if len(podName) == 0 {
						// it shouldn't be missing here
						t.Errorf("failed to get pod header in response: %v", err)
					} else {
						if len(expectPodName) == 0 {
							expectPodName = podName
						} else {
							require.Equal(t, expectPodName, podName)
						}
					}
				}
			}
		})
	},
}
