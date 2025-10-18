// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	nethttp "net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		RoundRobinLoadBalancingTest,
		ConsistentHashSourceIPLoadBalancingTest,
		ConsistentHashHeaderLoadBalancingTest,
		ConsistentHashCookieLoadBalancingTest,
		EndpointOverrideLoadBalancingTest,
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

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "round-robin-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-roundrobin"}, corev1.PodRunning, &PodReady)

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		t.Run("traffic should be split evenly", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/round",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			compareFunc := func(trafficMap map[string]int) bool {
				even := sendRequests / replicas
				for _, count := range trafficMap {
					if !AlmostEquals(count, even, offset) {
						return false
					}
				}

				return true
			}

			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, 30*time.Second, true, func(_ context.Context) (bool, error) {
				return runTrafficTest(t, suite, &req, &expectedResponse, sendRequests, compareFunc), nil
			}); err != nil {
				tlog.Errorf(t, "failed to run round robin load balancing test: %v", err)
			}
		})
	},
}

type TrafficCompareFunc func(trafficMap map[string]int) bool

func runTrafficTest(t *testing.T, suite *suite.ConformanceTestSuite,
	req *roundtripper.Request, expectedResponse *http.ExpectedResponse,
	totalRequestCount int, compareFunc TrafficCompareFunc,
) bool {
	if req == nil {
		t.Fatalf("request cannot be nil")
	}
	if expectedResponse == nil {
		t.Fatalf("expected response cannot be nil")
	}
	trafficMap := make(map[string]int)
	for i := 0; i < totalRequestCount; i++ {
		request := *req
		cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(request)
		if err != nil {
			t.Errorf("failed to get expected response: %v", err)
		}

		if err := http.CompareRoundTrip(t, &request, cReq, cResp, *expectedResponse); err != nil {
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

	ret := compareFunc(trafficMap)
	if !ret {
		tlog.Logf(t, "traffic map: %v", trafficMap)
	}

	return ret
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

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "source-ip-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-sourceip"}, corev1.PodRunning, &PodReady)

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		t.Run("all traffics route to the same backend with same source ip", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/source",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			compareFunc := func(trafficMap map[string]int) bool {
				// All traffic should be routed to the same pod.
				return len(trafficMap) == 1
			}

			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, 30*time.Second, true, func(_ context.Context) (bool, error) {
				return runTrafficTest(t, suite, &req, &expectedResponse, sendRequests, compareFunc), nil
			}); err != nil {
				tlog.Errorf(t, "failed to run source ip based consistent hash load balancing test: %v", err)
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

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "header-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-header"}, corev1.PodRunning, &PodReady)

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path: "/header",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}
		headers := []string{"0.0.0.0", "1.2.3.4", "4.5.6.7", "7.8.9.10", "10.11.12.13"}

		for _, header := range headers {
			t.Run(header, func(t *testing.T) {
				req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
				req.Headers["Lb-Test-Header"] = []string{header}
				got := runTrafficTest(t, suite, &req, &expectedResponse, sendRequests, func(trafficMap map[string]int) bool {
					// All traffic should be routed to the same pod.
					return len(trafficMap) == 1
				})
				require.True(t, got)
			})
		}
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

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cookie-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-cookie"}, corev1.PodRunning, &PodReady)

		gwHost := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		gwAddr := net.JoinHostPort(gwHost, "80")

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

				for range sendRequests {
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

var EndpointOverrideLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "EndpointOverrideLoadBalancing",
	Description: "Test for endpoint override load balancing functionality",
	Manifests:   []string{"testdata/load_balancing_endpoint_override.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const sendRequests = 10

		ns := "gateway-conformance-infra"
		headerRouteNN := types.NamespacedName{Name: "endpoint-override-header-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "endpoint-override-header-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-endpointoverride"}, corev1.PodRunning, &PodReady)

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), headerRouteNN)

		// Get pods associated with the service to find valid pod IPs and names
		ctx, cancel := context.WithTimeout(context.Background(), suite.TimeoutConfig.GetTimeout)
		defer cancel()

		// Get pods by label selector
		podList := &corev1.PodList{}
		err := suite.Client.List(ctx, podList, &client.ListOptions{
			Namespace:     ns,
			LabelSelector: labels.SelectorFromSet(map[string]string{"app": "lb-backend-endpointoverride"}),
		})
		require.NoError(t, err, "failed to list pods")
		require.NotEmpty(t, podList.Items, "should have pods")

		// Create mapping of pod IP to pod name
		podIPToName := make(map[string]string)
		var validPodIPs []string
		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" {
				podIPToName[pod.Status.PodIP] = pod.Name
				validPodIPs = append(validPodIPs, pod.Status.PodIP)
			}
		}
		require.NotEmpty(t, validPodIPs, "should have valid pod IPs")

		// Get service port
		svc := &corev1.Service{}
		err = suite.Client.Get(ctx, types.NamespacedName{Name: "lb-backend-endpointoverride", Namespace: ns}, svc)
		require.NoError(t, err, "failed to get service")
		require.NotEmpty(t, svc.Spec.Ports, "service should have ports")
		servicePort := svc.Spec.Ports[0].TargetPort.IntValue()

		t.Logf("Found %d valid pods: %v, service port: %d", len(validPodIPs), podIPToName, servicePort)

		t.Run("header-based endpoint override with valid pod IP should route to specific pod", func(t *testing.T) {
			// Use the first valid pod IP as override host
			targetPodIP := validPodIPs[0]
			format := "%s:%d"
			if IPFamily == "ipv6" {
				format = "[%s]:%d"
			}
			overrideHost := fmt.Sprintf(format, targetPodIP, servicePort)

			// Get the expected pod name from our mapping
			expectPodName := podIPToName[targetPodIP]
			require.NotEmpty(t, expectPodName, "failed to get expected pod name for IP %s", targetPodIP)

			t.Logf("Testing endpoint override with valid pod IP: %s, expecting pod: %s", overrideHost, expectPodName)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/endpoint-override-header",
					Headers: map[string]string{
						"x-custom-host": overrideHost,
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			// Test that all requests go to the expected pod
			for i := 0; i < sendRequests; i++ {
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				require.NoError(t, err, "failed to get expected response")
				require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse), "failed to compare request and response")

				actualPodName := cReq.Pod
				require.Equal(t, expectPodName, actualPodName, "request %d: expected pod %s but got %s", i+1, expectPodName, actualPodName)
			}

			t.Logf("All %d requests with valid override host %s routed to expected pod: %s", sendRequests, overrideHost, expectPodName)
		})
		t.Run("header-based endpoint override with invalid pod IP should fallback to load balancer policy", func(t *testing.T) {
			// Use an invalid pod IP that's not in the service endpoints
			invalidOverrideHost := "192.168.99.99:8080"

			t.Logf("Testing endpoint override with invalid pod IP: %s (should fallback to load balancer policy)", invalidOverrideHost)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/endpoint-override-header",
					Headers: map[string]string{
						"x-custom-host": invalidOverrideHost,
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			// Make multiple requests and verify fallback behavior (just check 200 response)
			for i := 0; i < sendRequests; i++ {
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				require.NoError(t, err, "failed to get expected response")
				require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse), "failed to compare request and response")

				// For invalid override host, we just verify the response is 200 (fallback works)
				// No need to check specific pod routing since it should use fallback policy
			}

			t.Logf("All %d requests with invalid override host %s got 200 response (fallback working)", sendRequests, invalidOverrideHost)
		})

		t.Run("header-based endpoint override without header should fallback to load balancer policy", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/endpoint-override-header",
					// No x-custom-host header
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			// Make multiple requests and verify fallback behavior (just check 200 response)
			for i := 0; i < sendRequests; i++ {
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				require.NoError(t, err, "failed to get expected response")
				require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse), "failed to compare request and response")

				// For missing header, we just verify the response is 200 (fallback works)
				// No need to check specific pod routing since it should use fallback policy
			}

			t.Logf("All %d requests without override header got 200 response (fallback working)", sendRequests)
		})
	},
}
