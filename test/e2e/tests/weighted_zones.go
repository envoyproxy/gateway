// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, WeightedZonesTest)
}

var WeightedZonesTest = suite.ConformanceTest{
	ShortName:   "WeightedZones",
	Description: "Test weighted zone traffic distribution via BackendTrafficPolicy",
	Manifests:   []string{"testdata/weighted-zones.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		WaitForPods(t, suite.Client, ns, map[string]string{"app": "weighted-zones-backend"}, corev1.PodRunning, &PodReady)

		BackendTrafficPolicyMustBeAccepted(t,
			suite.Client,
			types.NamespacedName{Name: "weighted-zones-policy", Namespace: ns},
			suite.ControllerName,
			gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(ns),
				Name:      gwapiv1.ObjectName("same-namespace"),
			},
		)

		// Pods being Ready doesn't mean Envoy has received both endpoints yet,
		// wait until both zones have served traffic before checking the distribution.
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(SameNamespaceGateway), types.NamespacedName{Name: "weighted-zones-route", Namespace: ns})
		expectedResponse := http.ExpectedResponse{
			Request:   http.Request{Path: "/weighted-zones"},
			Response:  http.Response{StatusCodes: []int{200}},
			Namespace: ns,
		}
		req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
		hitZone1, hitZone2 := false, false
		if err := wait.PollUntilContextTimeout(t.Context(), 100*time.Millisecond, suite.TimeoutConfig.MaxTimeToConsistency, true,
			func(_ context.Context) (bool, error) {
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					tlog.Logf(t, "warmup request failed: %v", err)
					return false, nil
				}
				if err := http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse); err != nil {
					tlog.Logf(t, "warmup unexpected response: %v", err)
					return false, nil
				}
				hitZone1 = hitZone1 || strings.Contains(cReq.Pod, "weighted-zones-backend-zone1")
				hitZone2 = hitZone2 || strings.Contains(cReq.Pod, "weighted-zones-backend-zone2")
				return hitZone1 && hitZone2, nil
			}); err != nil {
			t.Fatalf("failed waiting for both zones to receive traffic: zone1=%v zone2=%v: %v", hitZone1, hitZone2, err)
		}

		t.Run("traffic should be distributed according to zone weights", func(t *testing.T) {
			// BTP configures zone "1" with weight 80 and zone "2" with weight 20.
			// Deployments pin one pod per zone via nodeSelector.
			expected := map[string]int{
				"weighted-zones-backend-zone1": int(math.Round(sendRequests * 0.80)),
				"weighted-zones-backend-zone2": int(math.Round(sendRequests * 0.20)),
			}
			runWeightedBackendTest(t, suite, nil, "weighted-zones-route", "/weighted-zones", "weighted-zones-backend", expected)
		})
	},
}
