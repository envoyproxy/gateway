// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
			gwNN := types.NamespacedName{Name: "zone-aware-gtw", Namespace: ns}
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

			// The EnvoyProxy pods have a nodeSelector targeting zone "0" so all requests
			// should be routed to upstreams there.
			expected := map[string]int{
				"1": sendRequests,
			}
			endpointslice, err := suite.Clientset.DiscoveryV1().EndpointSlices(ns).List(context.Background(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(map[string]string{discoveryv1.LabelServiceName: "zone-aware-backend"}).String()})
			require.NoError(t, err)
			require.Len(t, endpointslice.Items, 1)
			podZoneMap := make(map[string]string)
			for _, sl := range endpointslice.Items {
				for _, ep := range sl.Endpoints {
					if ep.Zone == nil {
						t.Fatalf("endpoint %s/%s has no zone", sl.Namespace, sl.Name)
					}
					podZoneMap[ep.TargetRef.Name] = *ep.Zone
				}
			}
			reqMap := make(map[string]int)
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
					podZone := podZoneMap[podName]
					reqMap[podZone]++
				}
			}

			// We iterate over the actual traffic Map with podNamePrefix as the key to get the actual traffic.
			// Given an offset of 3, we expect the expected traffic to be within the actual traffic [actual-3,actual+3] interval.
			for prefix, actual := range reqMap {
				expect := expected[prefix]
				if !AlmostEquals(actual, expect, 3) {
					t.Errorf("The actual traffic distribution between zones is not consistent with the expected: %v", cmp.Diff(actual, expect))
				}
			}
		})
	},
}
