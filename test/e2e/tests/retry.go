// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, RetryTest)
}

var RetryTest = suite.ConformanceTest{
	ShortName:   "Retry",
	Description: "Test that the BackendTrafficPolicy API implementation supports retry",
	Manifests:   []string{"testdata/retry.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ctx := context.Background()

		promClient, err := prometheus.NewClient(suite.Client, types.NamespacedName{Name: "prometheus", Namespace: "monitoring"})
		require.NoError(t, err)

		t.Run("retry-on-500", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "retry-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/status/500",
				},
				Response: http.Response{
					StatusCodes: []int{500},
				},
				Namespace: ns,
			}

			promQL := fmt.Sprintf(`envoy_cluster_upstream_rq_retry{envoy_cluster_name="httproute/%s/%s/rule/0"}`, routeNN.Namespace, routeNN.Name)

			before := float64(0)
			v, err := promClient.QuerySum(ctx, promQL)
			if err == nil {
				before = v
			}
			tlog.Logf(t, "query count %s before: %v", promQL, before)

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}

			http.AwaitConvergence(t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					// check retry stats from Prometheus
					v, err := promClient.QuerySum(ctx, promQL)
					if err != nil {
						return false
					}
					tlog.Logf(t, "query count %s after: %v", promQL, v)

					delta := int64(v - before)
					// numRetries is 5, so delta mod 5 equals 0
					return delta > 0 && delta%5 == 0
				})
		})
	},
}
