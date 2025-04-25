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

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, ZoneAwareRoutingTest)
}

var ZoneAwareRoutingTest = suite.ConformanceTest{
	ShortName:   "ZoneAwareRouting",
	Description: "Resource with Zone Aware Routing enabled",
	Manifests:   []string{"testdata/zone-aware-routing.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("only local zone should get requests", func(t *testing.T) {
			const sendRequests = 50

			ns := "gateway-conformance-infra"
			zoneAwareRoute := types.NamespacedName{Name: "zone-aware-http-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			_, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.GatewayRef{
				NamespacedName: gwNN,
			})
			require.NoError(t, err)
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), zoneAwareRoute)

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

			expected := map[string]int{
				"zone-aware-backend-nonlocal": 0,
				"zone-aware-backend-local":    sendRequests,
			}

			weightMap := make(map[string]int)
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
					// all we need is the pod Name prefix
					podNamePrefix := ZoneRoutingExtractPodNamePrefix(podName)
					weightMap[podNamePrefix]++
				}
			}
			// We iterate over the actual traffic Map with podNamePrefix as the key to get the actual traffic.
			// Given an offset of 3, we expect the expected traffic to be within the actual traffic [actual-3,actual+3] interval.
			for prefix, actual := range weightMap {
				expect := expected[prefix]
				if !AlmostEquals(actual, expect, 3) {
					t.Errorf("The actual traffic distribution between zones is not consistent with the expected: %v", weightMap)
				}
			}
		})
	},
}

// ExtractPodNamePrefix Extract the Pod Name prefix
func ZoneRoutingExtractPodNamePrefix(podName string) string {
	pattern := regexp.MustCompile(`zone-aware-backend-(.+?)-`)
	match := pattern.FindStringSubmatch(podName)
	if len(match) > 1 {
		version := match[1]
		return fmt.Sprintf("zone-aware-backend-%s", version)
	}

	return podName
}
