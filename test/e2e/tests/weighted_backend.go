// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"regexp"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

const sendRequests = 50

func init() {
	ConformanceTests = append(ConformanceTests, WeightedBackendTest)
}

var WeightedBackendTest = suite.ConformanceTest{
	ShortName:   "WeightedRoute",
	Description: "Resource with Weight Backend enabled, and worked as expected.",
	Manifests: []string{
		"testdata/weighted-backend-all-equal.yaml",
		"testdata/weighted-backend-bluegreen.yaml",
		"testdata/weighted-backend-completing-rollout.yaml",
		"testdata/weighted-backend-mixed-valid-and-invalid.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("SameWeight", func(t *testing.T) {
			// The received request is approximately 1:1
			expected := map[string]int{
				"infra-backend-v1": sendRequests * .5,
				"infra-backend-v2": sendRequests * .5,
			}
			runWeightedBackendTest(t, suite, nil, "weight-equal-http-route", "/same-weight", "infra-backend", expected)
		})
		t.Run("BlueGreen", func(t *testing.T) {
			// The received request is approximately 9:1
			expected := map[string]int{
				"infra-backend-v1": sendRequests * .9,
				"infra-backend-v2": sendRequests * .1,
			}
			runWeightedBackendTest(t, suite, nil, "weight-bluegreen-http-route", "/blue-green", "infra-backend", expected)
		})
		t.Run("CompleteRollout", func(t *testing.T) {
			// All the requests should be proxied to v1
			expected := map[string]int{
				"infra-backend-v1": sendRequests * 1,
				"infra-backend-v2": sendRequests * 0,
			}
			runWeightedBackendTest(t, suite, nil, "weight-complete-rollout-http-route", "/complete-rollout", "infra-backend", expected)
		})

		t.Run("MixedValidAndInvalid", func(t *testing.T) {
			// Requests should be distributed to valid and invalid backends according to their weights
			testMixedValidAndInvalid(t, suite)
		})
	},
}

func runWeightedBackendTest(t *testing.T, suite *suite.ConformanceTestSuite, gateway *types.NamespacedName, routeName, path, backendName string, expectedOutput map[string]int) {
	weightEqualRoute := types.NamespacedName{Name: routeName, Namespace: ConformanceInfraNamespace}

	gatewayRef := kubernetes.NewGatewayRef(SameNamespaceGateway)
	if gateway != nil {
		gatewayRef = kubernetes.NewGatewayRef(*gateway)
	}

	gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, gatewayRef, weightEqualRoute)

	// Make sure all test resources are ready
	kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ConformanceInfraNamespace})

	expectedResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
		},
		Response: http.Response{
			StatusCodes: []int{200},
		},
		Namespace: ConformanceInfraNamespace,
	}
	// Make sure the backend is ready
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

	req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

	weightMap := make(map[string]int)
	for range sendRequests {
		cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
		if err != nil {
			t.Errorf("failed to get expected response: %v", err)
		}

		if err := http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse); err != nil {
			t.Errorf("failed to compare request and response: %v", err)
		}

		podName := cReq.Pod
		if len(podName) == 0 {
			// it shouldn't be missing here
			t.Errorf("failed to get pod header in response: %v", err)
		} else {
			// all we need is the pod Name prefix
			podNamePrefix := extractPodNamePrefix(podName, backendName)
			weightMap[podNamePrefix]++
		}
	}

	// We iterate over the actual traffic Map with podNamePrefix as the key to get the actual traffic.
	// Given an offset of 3, we expect the expected traffic to be within the actual traffic [actual-3,actual+3] interval.
	for prefix, actual := range weightMap {
		expect := expectedOutput[prefix]
		if !AlmostEquals(actual, expect, 3) {
			t.Errorf("The actual traffic weights are not consistent with the expected routing weights, actual %s %d, expected %s %d", prefix, actual, prefix, expect)
		}
	}
}

func extractPodNamePrefix(podName, prefix string) string {
	pattern := regexp.MustCompile(prefix + `-(.+?)-`)
	match := pattern.FindStringSubmatch(podName)
	if len(match) > 1 {
		version := match[1]
		return fmt.Sprintf("%s-%s", prefix, version)
	}

	return podName
}

func testMixedValidAndInvalid(t *testing.T, suite *suite.ConformanceTestSuite) {
	weightEqualRoute := types.NamespacedName{Name: "weight-mixed-valid-and-invalid-http-route", Namespace: ConformanceInfraNamespace}
	gatewayRef := kubernetes.NewGatewayRef(SameNamespaceGateway)
	gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, gatewayRef, weightEqualRoute)

	// Make sure all test resources are ready
	kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ConformanceInfraNamespace})

	scenarios := []struct {
		name        string
		path        string
		failureCode int
	}{
		{
			name:        "MixedValidAndInvalid",
			path:        "/mixed-valid-and-invalid",
			failureCode: 500,
		},
		{
			name:        "MixedValidAndNoEndpoints",
			path:        "/mixed-valid-and-no-endpoints",
			failureCode: 503,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			runMixedValidAndInvalidScenario(t, suite, gwAddr, scenario.path, scenario.failureCode)
		})
	}
}

const (
	mixedValidAndInvalidRequests    = 100
	mixedValidAndInvalidSuccessLow  = 80
	mixedValidAndInvalidSuccessHigh = 99
)

func runMixedValidAndInvalidScenario(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr, path string, failureCode int) {
	t.Helper()

	// Make sure the valid(response 200) backend are ready.
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
		Request: http.Request{
			Path: path,
		},
		Namespace: ConformanceInfraNamespace,
		Response: http.Response{
			StatusCodes: []int{200},
		},
	})

	// Test if the requests are distributed to valid and invalid backends according to their weights
	expected := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
		},
		Response: http.Response{
			StatusCodes: []int{200, failureCode},
		},
		Namespace: ConformanceInfraNamespace,
	}
	req := http.MakeRequest(t, &expected, gwAddr, "HTTP", "http")

	successCount := 0
	failureCount := 0
	for i := 0; i < mixedValidAndInvalidRequests; i++ {
		_, response, err := suite.RoundTripper.CaptureRoundTrip(req)
		if err != nil {
			t.Errorf("failed to get expected response: %v", err)
			continue
		}
		switch response.StatusCode {
		case 200:
			successCount++
		case failureCode:
			failureCount++
		default:
			t.Errorf("unexpected status code %d for %s", response.StatusCode, path)
		}
	}

	if successCount < mixedValidAndInvalidSuccessLow || successCount > mixedValidAndInvalidSuccessHigh {
		t.Errorf("actual success count for %s is not within the expected range, success %d", path, successCount)
	}

	if successCount+failureCount != mixedValidAndInvalidRequests {
		t.Errorf("the sum of success and failure count for %s is not equal to the total requests, success %d, failure %d, total %d", path, successCount, failureCount, mixedValidAndInvalidRequests)
	}

	t.Logf("success count for %s is %d, failure count for %s is %d", path, successCount, path, failureCount)
}
