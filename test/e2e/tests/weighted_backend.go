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
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("SameWeight", func(t *testing.T) {
			// The received request is approximately 1:1
			expected := map[string]int{
				"infra-backend-v1": sendRequests * .5,
				"infra-backend-v2": sendRequests * .5,
			}
			runWeightedBackendTest(t, suite, "weight-equal-http-route", "/same-weight", "infra-backend", expected)
		})
		t.Run("BlueGreen", func(t *testing.T) {
			// The received request is approximately 9:1
			expected := map[string]int{
				"infra-backend-v1": sendRequests * .9,
				"infra-backend-v2": sendRequests * .1,
			}
			runWeightedBackendTest(t, suite, "weight-bluegreen-http-route", "/blue-green", "infra-backend", expected)
		})
		t.Run("CompleteRollout", func(t *testing.T) {
			// All the requests should be proxied to v1
			expected := map[string]int{
				"infra-backend-v1": sendRequests * 1,
				"infra-backend-v2": sendRequests * 0,
			}
			runWeightedBackendTest(t, suite, "weight-complete-rollout-http-route", "/complete-rollout", "infra-backend", expected)
		})
	},
}

func runWeightedBackendTest(t *testing.T, suite *suite.ConformanceTestSuite, routeName, path, backendName string, expectedOutput map[string]int) {
	weightEqualRoute := types.NamespacedName{Name: routeName, Namespace: ConformanceInfraNamespace}
	gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
		suite.ControllerName,
		kubernetes.NewGatewayRef(SameNamespaceGateway), weightEqualRoute)

	// make sure all backend resources are ready
	kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ConformanceInfraNamespace})

	expectedResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
		},
		Response: http.Response{
			StatusCode: 200,
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

		if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
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
