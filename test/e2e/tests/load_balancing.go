// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	nethttp "net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		RoundRobinLoadBalancingTest,
		ConsistentHashSourceIPLoadBalancingTest,
		ConsistentHashHeaderLoadBalancingTest,
		ConsistentHashCookieLoadBalancingTest,
	)
}

var RoundRobinLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "RoundRobinLoadBalancing",
	Description: "Test for round robin load balancing type",
	Manifests:   []string{"testdata/load_balancing_round_robin.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const (
			sendRequests = 100
			replicas     = 3
			offset       = 5
		)

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
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-roundrobin"}, corev1.PodRunning, PodReady)

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

			if err := wait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, 15*time.Second, true, func(_ context.Context) (bool, error) {
				return runTrafficTest(t, suite, req, expectedResponse, replicas, sendRequests, offset), nil
			}); err != nil {
			}
		})
	},
}

func runTrafficTest(t *testing.T, suite *suite.ConformanceTestSuite,
	req roundtripper.Request, expectedResponse http.ExpectedResponse,
	replicas, totalRequestCount, offset int,
) bool {
	trafficMap := make(map[string]int)
	for i := 0; i < totalRequestCount; i++ {
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
			tlog.Errorf(t, "failed to get pod header in response: %v", err)
		} else {
			trafficMap[podName]++
		}
	}

	// Expect traffic number for each endpoint.
	even := totalRequestCount / replicas
	for podName, traffic := range trafficMap {
		if !AlmostEquals(traffic, even, offset) {
			tlog.Logf(t, "The traffic [Total:%d Replicas: %d Offset: %d] are not be split evenly for pod %s: %d",
				totalRequestCount, replicas, offset, podName, traffic)
			return false
		}
	}

	return true
}

var ConsistentHashSourceIPLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "SourceIPBasedConsistentHashLoadBalancing",
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
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-sourceip"}, corev1.PodRunning, PodReady)

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
	ShortName:   "HeaderBasedConsistentHashLoadBalancing",
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
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-header"}, corev1.PodRunning, PodReady)

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

var ConsistentHashCookieLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "CookieBasedConsistentHashLoadBalancing",
	Description: "Test for cookie based consistent hash load balancing type",
	Manifests:   []string{"testdata/load_balancing_consistent_hash_cookie.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const sendRequests = 10

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "cookie-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cookie-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-cookie"}, corev1.PodRunning, PodReady)

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("all traffics route to the same backend with same test cookie", func(t *testing.T) {
			cookieJar, err := cookiejar.New(nil)
			require.NoError(t, err)

			// Making request on our own since the gateway-api conformance suite does not support
			// setting cookies for one request.
			client := &nethttp.Client{
				Jar:       cookieJar,
				Transport: nethttp.DefaultTransport,
			}
			req, err := nethttp.NewRequest(nethttp.MethodGet, fmt.Sprintf("http://%s/cookie", gwAddr), nil)
			require.NoError(t, err)

			cookieValues := []string{"abc", "def", "ghi", "jkl", "mno"}
			for _, cookieValue := range cookieValues {
				// Same test cookie will always hit the same endpoint.
				var expectPodName string

				client.Jar.SetCookies(req.URL, []*nethttp.Cookie{
					{
						Name:  "Lb-Test-Cookie",
						Value: cookieValue,
					},
				})

				for i := 0; i < sendRequests; i++ {
					resp, err := client.Do(req)
					if err != nil {
						t.Errorf("failed to get response: %v", err)
					}

					body, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					require.Equal(t, nethttp.StatusOK, resp.StatusCode)

					// Parse response body.
					cReq := &roundtripper.CapturedRequest{}
					if resp.Header.Get("Content-Type") == "application/json" {
						err = json.Unmarshal(body, cReq)
						require.NoError(t, err)
					} else {
						t.Fatalf("unsupported response content type")
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

					require.NoError(t, resp.Body.Close())
				}
			}
		})

		t.Run("a cookie will be generated if the require cookie does not exist", func(t *testing.T) {
			cookieJar, err := cookiejar.New(nil)
			require.NoError(t, err)

			// Making request on our own since the gateway-api conformance suite does not support
			// setting cookies for one request.
			client := &nethttp.Client{
				Jar:       cookieJar,
				Transport: nethttp.DefaultTransport,
			}
			req, err := nethttp.NewRequest(nethttp.MethodGet, fmt.Sprintf("http://%s/cookie", gwAddr), nil)
			require.NoError(t, err)

			// A not desired cookie has been set.
			client.Jar.SetCookies(req.URL, []*nethttp.Cookie{
				{
					Name:  "foo",
					Value: "bar",
				},
			})

			waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
				resp, err := client.Do(req)
				if err != nil {
					t.Errorf("failed to get response: %v", err)
					return false, err
				}

				if resp.StatusCode != nethttp.StatusOK {
					return false, nil
				}

				if h := resp.Header.Get("set-cookie"); len(h) > 0 &&
					strings.Contains(h, "Lb-Test-Cookie") &&
					strings.Contains(h, "Max-Age=60; SameSite=Strict") {
					return true, nil
				}

				tlog.Logf(t, "Cookie have not been generated yet")
				return false, nil
			})
			require.NoError(t, waitErr)
		})
	},
}
