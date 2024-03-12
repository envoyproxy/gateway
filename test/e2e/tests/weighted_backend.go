// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

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

func init() {
	ConformanceTests = append(ConformanceTests, WeightEqualTest, WeightBlueGreenTest, WeightCompleteRolloutTest)
}

const sendRequest = 50

var WeightEqualTest = suite.ConformanceTest{
	ShortName:   "WeightEqualBackend",
	Description: "Resource with Weight Backend enabled, and use the all backend weight is equal",
	Manifests:   []string{"testdata/weighted-backend-all-equal.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("all backends receive the same weight of traffic", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			weightEqualRoute := types.NamespacedName{Name: "weight-equal-http-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), weightEqualRoute)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			// Since we only route to pods with "infra-backend-v 1" and "infra-backend-v 2" prefixes
			// So here we use fixed weight values
			expected := map[string]int{
				"infra-backend-v1": sendRequest * .5,
				"infra-backend-v2": sendRequest * .5,
			}
			weightMap := make(map[string]int)
			for i := 0; i < sendRequest; i++ {
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
					podNamePrefix := ExtractPodNamePrefix(podName)
					weightMap[podNamePrefix]++
				}
			}

			// We iterate over the actual traffic Map with podNamePrefix as the key to get the actual traffic.
			// Given an offset of 3, we expect the expected traffic to be within the actual traffic [actual-3,actual+3] interval.
			for prefix, actual := range weightMap {
				expect := expected[prefix]
				if !AlmostEquals(actual, expect, 3) {
					t.Errorf("The actual traffic weights are not consistent with the expected routing weights")
				}
			}
		})
	},
}

var WeightBlueGreenTest = suite.ConformanceTest{
	ShortName:   "WeightBlueGreenBackend",
	Description: "Resource with Weight Backend enabled, and use the blue-green policy weight setting",
	Manifests:   []string{"testdata/weighted-backend-bluegreen.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("all backends receive the blue green weight of traffic", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			weightEqualRoute := types.NamespacedName{Name: "weight-bluegreen-http-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), weightEqualRoute)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			// Since we only route to pods with "infra-backend-v 1" and "infra-backend-v 2" prefixes
			// So here we use fixed weight values
			expected := map[string]int{
				"infra-backend-v1": sendRequest * .9,
				"infra-backend-v2": sendRequest * .1,
			}
			weightMap := make(map[string]int)
			for i := 0; i < sendRequest; i++ {
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
					podNamePrefix := ExtractPodNamePrefix(podName)
					weightMap[podNamePrefix]++
				}
			}

			// We iterate over the actual traffic Map with podNamePrefix as the key to get the actual traffic.
			// Given an offset of 3, we expect the expected traffic to be within the actual traffic [actual-3,actual+3] interval.
			for prefix, actual := range weightMap {
				expect := expected[prefix]
				if !AlmostEquals(actual, expect, 3) {
					t.Errorf("The actual traffic weights are not consistent with the expected routing weights")
				}
			}
		})
	},
}

var WeightCompleteRolloutTest = suite.ConformanceTest{
	ShortName:   "WeightCompleteRollout",
	Description: "Resource with Weight Backend enabled, and use the completing rollout policy weight setting",
	Manifests:   []string{"testdata/weight-backend-completing-rollout.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("all backends receive the complete rollout weight of traffic", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			weightEqualRoute := types.NamespacedName{Name: "weight-complete-rollout-http-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), weightEqualRoute)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			// Since we only route to pods with "infra-backend-v 1" and "infra-backend-v 2" prefixes
			// So here we use fixed weight values
			expected := map[string]int{
				"infra-backend-v1": sendRequest * 1,
				"infra-backend-v2": sendRequest * 0,
			}
			weightMap := make(map[string]int)
			for i := 0; i < sendRequest; i++ {
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
					podNamePrefix := ExtractPodNamePrefix(podName)
					weightMap[podNamePrefix]++
				}
			}

			// We iterate over the actual traffic Map with podNamePrefix as the key to get the actual traffic.
			// Given an offset of 3, we expect the expected traffic to be within the actual traffic [actual-3,actual+3] interval.
			for prefix, actual := range weightMap {
				expect := expected[prefix]
				if !AlmostEquals(actual, expect, 3) {
					t.Errorf("The actual traffic weights are not consistent with the expected routing weights")
				}
			}
		})
	},
}

// ExtractPodNamePrefix Extract the Pod Name prefix
func ExtractPodNamePrefix(podName string) string {

	pattern := regexp.MustCompile(`infra-backend-(.+?)-`)
	match := pattern.FindStringSubmatch(podName)
	if len(match) > 1 {
		version := match[1]
		return fmt.Sprintf("infra-backend-%s", version)
	}

	return podName
}
