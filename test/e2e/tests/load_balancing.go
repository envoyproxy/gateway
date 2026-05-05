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
	"k8s.io/apimachinery/pkg/util/sets"
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
		RoundRobinLoadBalancing,
		SourceIPBasedConsistentHashLoadBalancing,
		HeaderBasedConsistentHashLoadBalancing,
		CookieBasedConsistentHashLoadBalancing,
		EndpointOverrideLoadBalancing,
		MultiHeaderConsistentHashHeaderLoadBalancing,
		QueryParamsBasedConsistentHashLoadBalancing,
		BackendUtilizationLoadBalancingTest,
		BackendUtilizationWeightedZonesLoadBalancingTest,
	)
}

var BackendUtilizationLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "BackendUtilizationLoadBalancing",
	Description: "Test that BackendUtilization shifts traffic toward lower-utilization backends using ORCA metrics",
	Manifests: []string{
		"testdata/load_balancing_backend_utilization.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const (
			warmupRequests = 20
			sendRequests   = 100
			lowUtilMinPct  = 80 // ORCA utilization split is 90:10 but gives a 10% buffer
		)

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "backend-utilization-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-utilization-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-utilization"}, corev1.PodRunning, &PodReady)

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path: "/backend-utilization",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}
		req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

		// Identify which pod is low-util vs high-util by deployment name prefix.
		isLowUtil := func(podName string) bool {
			return strings.Contains(podName, "lb-backend-low-util")
		}

		t.Run("warmup until both backends are hit", func(t *testing.T) {
			if err := wait.PollUntilContextTimeout(t.Context(), time.Second, 60*time.Second, true, func(_ context.Context) (bool, error) {
				return runTrafficTest(t, suite, &req, &expectedResponse, warmupRequests, func(trafficMap map[string]int) bool {
					return len(trafficMap) >= 2
				}), nil
			}); err != nil {
				tlog.Errorf(t, "failed to hit both backends during warmup: %v", err)
			}
		})

		// Pause to allow envoy to compute weights. Should be longer than WeightUpdatePeriod duration.
		time.Sleep(200 * time.Millisecond)

		t.Run("traffic should skew toward low-utilization backend", func(t *testing.T) {
			if err := wait.PollUntilContextTimeout(t.Context(), time.Second, 60*time.Second, true, func(_ context.Context) (bool, error) {
				return runTrafficTest(t, suite, &req, &expectedResponse, sendRequests, func(trafficMap map[string]int) bool {
					lowCount := 0
					total := 0
					for podName, count := range trafficMap {
						total += count
						if isLowUtil(podName) {
							lowCount += count
						}
					}
					if total == 0 {
						return false
					}
					lowPct := (lowCount * 100) / total
					tlog.Logf(t, "traffic distribution: low-util=%d/%d (%d%%), high-util=%d/%d",
						lowCount, total, lowPct, total-lowCount, total)
					return lowPct >= lowUtilMinPct
				}), nil
			}); err != nil {
				tlog.Errorf(t, "failed to run backend utilization load balancing test: %v", err)
			}
		})
	},
}

var BackendUtilizationWeightedZonesLoadBalancingTest = suite.ConformanceTest{
	ShortName:   "BackendUtilizationWeightedZones",
	Description: "Test that weighted zones take precedence over backend utilization across localities",
	Manifests: []string{
		"testdata/load_balancing_backend_utilization_weighted_zones.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const (
			maxWarmupRequests = 200
			warmupTimeout     = 20 * time.Second
			sampleRequests    = 100
			zone1MinPct       = 80 // WeightedZones split is 90:10 but this gives a buffer of ~10%
		)

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "backend-utilization-weighted-zones-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-utilization-weighted-zones-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "lb-backend-utilization-weighted-zones"}, corev1.PodRunning, &PodReady)

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path: "/backend-utilization-weighted-zones",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}
		req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

		isZone1 := func(podName string) bool {
			return strings.Contains(podName, "zone1")
		}
		isZone2 := func(podName string) bool {
			return strings.Contains(podName, "zone2")
		}

		t.Run("warmup until both zones are hit", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), warmupTimeout)
			defer cancel()

			hitZone1, hitZone2 := false, false
			sent := 0
			for sent < maxWarmupRequests {
				if ctx.Err() != nil {
					break
				}
				request := req
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(request)
				sent++
				if err != nil {
					tlog.Logf(t, "warmup request failed: %v", err)
					continue
				}
				if err := http.CompareRoundTrip(t, &request, cReq, cResp, expectedResponse); err != nil {
					tlog.Logf(t, "warmup unexpected response: %v", err)
					continue
				}
				if isZone1(cReq.Pod) {
					hitZone1 = true
				}
				if isZone2(cReq.Pod) {
					hitZone2 = true
				}
				if hitZone1 && hitZone2 {
					tlog.Logf(t, "both zones hit after %d warmup requests", sent)
					return
				}
			}
			tlog.Errorf(t, "failed to hit both zones during warmup after %d requests: zone1=%v zone2=%v", sent, hitZone1, hitZone2)
		})

		// Pause to allow envoy to compute weights. Should be longer than WeightUpdatePeriod duration.
		time.Sleep(200 * time.Millisecond)

		t.Run("traffic should favor the higher weighted zone", func(t *testing.T) {
			if err := wait.PollUntilContextTimeout(t.Context(), time.Second, 60*time.Second, true, func(_ context.Context) (bool, error) {
				return runTrafficTest(t, suite, &req, &expectedResponse, sampleRequests, func(trafficMap map[string]int) bool {
					zone1Count := 0
					zone2Count := 0
					total := 0
					for podName, count := range trafficMap {
						total += count
						if isZone1(podName) {
							zone1Count += count
						}
						if isZone2(podName) {
							zone2Count += count
						}
					}
					if total == 0 {
						return false
					}
					zone1Pct := (zone1Count * 100) / total
					tlog.Logf(t, "traffic distribution: zone1=%d/%d (%d%%), zone2=%d/%d", zone1Count, total, zone1Pct, zone2Count, total)
					return zone1Pct >= zone1MinPct && zone2Count > 0
				}), nil
			}); err != nil {
				tlog.Errorf(t, "failed to run backend utilization weighted zones test: %v", err)
			}
		})
	},
}

var RoundRobinLoadBalancing = suite.ConformanceTest{
	ShortName:   "RoundRobinLoadBalancing",
	Description: "Test for round robin load balancing type",
	Manifests: []string{
		"testdata/load_balancing_round_robin.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const (
			// The replicas of the deployment gateway-conformance-infra/web-backend.
			replicas     = 4
			sendRequests = replicas * 50
			offset       = 5
		)

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "round-robin-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "round-robin-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
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

			if err := wait.PollUntilContextTimeout(t.Context(), time.Second, 30*time.Second, true, func(_ context.Context) (bool, error) {
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
		// this should never happen, just a sanity check for the caller
		t.Fatalf("request cannot be nil")
		return false
	}
	if expectedResponse == nil {
		// this should never happen, just a sanity check for the caller
		t.Fatalf("expected response cannot be nil")
		return false
	}
	trafficMap := make(map[string]int)
	for range totalRequestCount {
		request := *req
		cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(request)
		if err != nil {
			tlog.Logf(t, "failed to get expected response: %v", err)
			continue
		}

		if err := http.CompareRoundTrip(t, &request, cReq, cResp, *expectedResponse); err != nil {
			tlog.Logf(t, "failed to get expected response: %v", err)
			continue
		}

		podName := cReq.Pod
		if len(podName) == 0 {
			// it shouldn't be missing here
			tlog.Logf(t, "failed to get pod header in response: %v", err)
		} else {
			trafficMap[podName]++
		}
	}

	ret := compareFunc(trafficMap)
	if !ret {
		tlog.Logf(t, "traffic map: %v", trafficMap)
		// wait for a while to let envoy flush all the logs.
		time.Sleep(6 * time.Second)
		consistentHashDump(t, suite.RestConfig)
		t.FailNow()
	}

	return ret
}

var SourceIPBasedConsistentHashLoadBalancing = suite.ConformanceTest{
	ShortName:   "SourceIPBasedConsistentHashLoadBalancing",
	Description: "Test for source IP based consistent hash load balancing type",
	Manifests: []string{
		"testdata/load_balancing_consistent_hash_source_ip.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "source-ip-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "source-ip-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		t.Run("all traffics route to the same backend with same source ip", func(t *testing.T) {
			expectedResponse := &http.ExpectedResponse{
				Request: http.Request{
					Path: "/source",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, expectedResponse, gwAddr, "HTTP", "http")

			runConsistentHashLoadBalancingTest(t, suite, &req, expectedResponse)
		})
	},
}

var HeaderBasedConsistentHashLoadBalancing = suite.ConformanceTest{
	ShortName:   "HeaderBasedConsistentHashLoadBalancing",
	Description: "Test for header based consistent hash load balancing type",
	Manifests: []string{
		"testdata/load_balancing_consistent_hash_header.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "header-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "header-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		expectedResponse := &http.ExpectedResponse{
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
				req := http.MakeRequest(t, expectedResponse, gwAddr, "HTTP", "http")
				req.Headers["Lb-Test-Header"] = []string{header}
				runConsistentHashLoadBalancingTest(t, suite, &req, expectedResponse)
			})
		}
	},
}

var MultiHeaderConsistentHashHeaderLoadBalancing = suite.ConformanceTest{
	ShortName:   "MultiHeaderBasedConsistentHashLoadBalancing",
	Description: "Test for multiple header based consistent hash load balancing type",
	Manifests: []string{
		"testdata/load_balancing_consistent_hash_multi_header.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "header-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "header-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		expectedResponse := &http.ExpectedResponse{
			Request: http.Request{
				Path: "/header",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		// Test with different combinations of multiple headers
		headerCombinations := []struct {
			name    string
			header1 string
			header2 string
		}{
			{"combo1", "value-a", "value-b"},
			{"combo2", "value-x", "value-y"},
			{"combo3", "test1", "test2"},
			{"combo4", "foo", "bar"},
			{"combo5", "alpha", "beta"},
		}

		for _, combo := range headerCombinations {
			t.Run(combo.name, func(t *testing.T) {
				req := http.MakeRequest(t, expectedResponse, gwAddr, "HTTP", "http")
				// Set both headers for the consistent hash
				req.Headers["Lb-Test-1"] = []string{combo.header1}
				req.Headers["Lb-Test-2"] = []string{combo.header2}

				runConsistentHashLoadBalancingTest(t, suite, &req, expectedResponse)
			})
		}
	},
}

func runConsistentHashLoadBalancingTest(t *testing.T, suite *suite.ConformanceTestSuite, req *roundtripper.Request, expectedResponse *http.ExpectedResponse) {
	if err := wait.PollUntilContextTimeout(t.Context(), time.Second, time.Minute, true, func(_ context.Context) (bool, error) {
		got := runTrafficTest(t, suite, req, expectedResponse, sendRequests, func(trafficMap map[string]int) bool {
			// All traffic with the same hash combination should route to the same pod
			return len(trafficMap) == 1
		})
		return got, nil
	}); err != nil {
		tlog.Errorf(t, "failed to run consistent hash load balancing test: %v", err)
	}
}

var CookieBasedConsistentHashLoadBalancing = suite.ConformanceTest{
	ShortName:   "CookieBasedConsistentHashLoadBalancing",
	Description: "Test for cookie based consistent hash load balancing type",
	Manifests: []string{
		"testdata/load_balancing_consistent_hash_cookie.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const sendRequests = 10

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "cookie-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cookie-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)

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
				Timeout:   suite.TimeoutConfig.RequestTimeout,
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
						t.Logf("failed to get response: %v", err)
						continue
					}

					body, err := io.ReadAll(resp.Body)
					require.NoError(t, err)
					if resp.StatusCode != nethttp.StatusOK {
						t.Logf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
						continue
					}

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
						t.Logf("failed to get pod header in response: %v", err)
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
				Timeout:   suite.TimeoutConfig.RequestTimeout,
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

			waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(_ context.Context) (bool, error) {
				resp, err := client.Do(req)
				if err != nil {
					tlog.Logf(t, "failed to get response: %v", err)
					return false, err
				}

				if resp.StatusCode != nethttp.StatusOK {
					tlog.Logf(t, "unexpected status code: %d", resp.StatusCode)
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

var QueryParamsBasedConsistentHashLoadBalancing = suite.ConformanceTest{
	ShortName:   "QueryParamsBasedConsistentHashLoadBalancing",
	Description: "Test for multiple query parameter based consistent hash load balancing type",
	Manifests: []string{
		"testdata/load_balancing_consistent_hash_query_parameter.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "query-parameter-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "query-parameter-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		// Test with different combinations of multiple queries
		queryCombinations := []struct {
			name   string
			query1 string
			query2 string
		}{
			{"combo1", "value-a", "value-b"},
			{"combo2", "value-x", "value-y"},
			{"combo3", "test1", "test2"},
			{"combo4", "foo", "bar"},
			{"combo5", "alpha", "beta"},
		}
		for _, queryCombination := range queryCombinations {
			t.Run(queryCombination.name, func(t *testing.T) {
				expectedResponse := &http.ExpectedResponse{
					Request: http.Request{
						Path: fmt.Sprintf("/query-parameter?lb-query-parameter-first=%s&lb-query-parameter-second=%s", queryCombination.query1, queryCombination.query2),
					},
					Response: http.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				}
				req := http.MakeRequest(t, expectedResponse, gwAddr, "HTTP", "http")
				runConsistentHashLoadBalancingTest(t, suite, &req, expectedResponse)
			})
		}
	},
}

var EndpointOverrideLoadBalancing = suite.ConformanceTest{
	ShortName:   "EndpointOverrideLoadBalancing",
	Description: "Test for endpoint override load balancing functionality",
	Manifests: []string{
		"testdata/load_balancing_endpoint_override.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const sendRequests = 10

		ns := "gateway-conformance-infra"
		headerRouteNN := types.NamespacedName{Name: "endpoint-override-header-lb-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "endpoint-override-header-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), headerRouteNN)
		// Get pods associated with the service to find valid pod IPs and names
		ctx, cancel := context.WithTimeout(t.Context(), suite.TimeoutConfig.GetTimeout)
		defer cancel()

		// Get pods by label selector
		podList := &corev1.PodList{}
		err := suite.Client.List(ctx, podList, &client.ListOptions{
			Namespace:     ns,
			LabelSelector: labels.SelectorFromSet(map[string]string{"app": "web-backend"}),
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
		err = suite.Client.Get(ctx, types.NamespacedName{Name: "web-backend", Namespace: ns}, svc)
		require.NoError(t, err, "failed to get service")
		require.NotEmpty(t, svc.Spec.Ports, "service should have ports")
		servicePort := svc.Spec.Ports[0].TargetPort.IntValue()

		t.Logf("Found %d valid pods: %v, service port: %d", len(validPodIPs), podIPToName, servicePort)

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

			runEndpointOverrideTest(t, suite, gwAddr, podIPToName, &expectedResponse)

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
			runEndpointOverrideTest(t, suite, gwAddr, podIPToName, &expectedResponse)

			t.Logf("All %d requests without override header got 200 response (fallback working)", sendRequests)
		})

		t.Run("header-based endpoint override with valid pod IP should route to specific pod", func(t *testing.T) {
			// Use the first valid pod IP as override host
			targetPodIP := validPodIPs[0]
			overrideHost := net.JoinHostPort(targetPodIP, fmt.Sprintf("%d", servicePort))

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
			for i := range sendRequests {
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				require.NoError(t, err, "failed to get expected response")
				require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse), "failed to compare request and response")

				actualPodName := cReq.Pod
				require.Equal(t, expectPodName, actualPodName, "request %d: expected pod %s but got %s", i+1, expectPodName, actualPodName)
			}

			t.Logf("All %d requests with valid override host %s routed to expected pod: %s", sendRequests, overrideHost, expectPodName)
		})
	},
}

func runEndpointOverrideTest(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr string, podIPToName map[string]string, expectedResponse *http.ExpectedResponse) {
	req := http.MakeRequest(t, expectedResponse, gwAddr, "HTTP", "http")

	allPodNames := sets.NewString()
	for _, podName := range podIPToName {
		allPodNames.Insert(podName)
	}

	tlog.Logf(t, "all pods: %v", allPodNames.List())

	// Make multiple requests and verify fallback behavior (just check 200 response)
	for range sendRequests {
		cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
		require.NoError(t, err, "failed to get expected response")
		require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cResp, *expectedResponse), "failed to compare request and response")

		// For invalid override host, we just verify the response is 200 (fallback works)
		// and the request can hit any of the valid pods since it should use fallback policy
		tlog.Logf(t, "received response from pod: %s", cReq.Pod)
		allPodNames.Delete(cReq.Pod)
	}

	require.Emptyf(t, allPodNames, "all the valid pods should be removed.")
}
