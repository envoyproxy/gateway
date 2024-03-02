// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/test/e2e/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, RetryTest)

}

var RetryTest = suite.ConformanceTest{
	ShortName:   "Retry",
	Description: "Test that the BackendTrafficPolicy API implementation supports retry",
	Manifests:   []string{"testdata/retry.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("retry-on-500", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "retry-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/status/500",
				},
				Response: http.Response{
					StatusCode: 500,
				},
				Namespace: ns,
			}

			promAddr, err := prometheus.Address(suite.Client, types.NamespacedName{Name: "prometheus", Namespace: "monitoring"})
			require.NoError(t, err)
			promQL := fmt.Sprintf(`envoy_cluster_upstream_rq_retry{envoy_cluster_name="httproute/%s/%s/rule/0"}`, routeNN.Namespace, routeNN.Name)

			before := float64(0)
			v, err := prometheus.QuerySum(promAddr, promQL)
			if err == nil {
				before = v
			}
			t.Logf("query count %s before: %v", promQL, before)

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}

			http.AwaitConvergence(t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					// check retry stats from Prometheus
					v, err := prometheus.QuerySum(promAddr, promQL)
					if err != nil {
						return false
					}
					t.Logf("query count %s after: %v", promQL, v)

					delta := int64(v - before)
					// numRetries is 5, so delta mod 5 equals 0
					return delta > 0 && delta%5 == 0
				})
		})
	},
}
