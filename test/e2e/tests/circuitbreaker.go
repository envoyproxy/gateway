// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, CircuitBreakerTest)
}

var CircuitBreakerTest = suite.ConformanceTest{
	ShortName:   "CircuitBreaker",
	Description: "Test circuit breaker functionality",
	Manifests:   []string{"testdata/circuitbreaker.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Deny All Requests", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-circuitbreaker", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
			// expect overflow since the policy applies a "closed" circuit breaker
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/circuitbreaker",
				},
				Response: http.Response{
					StatusCodes: []int{503},
					Headers: map[string]string{
						"x-envoy-overloaded": "true",
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("Retry budget worked", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-retry-budget", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			// make sure that the backend is healthy.
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/retry-budget",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			})

			promClient, err := prometheus.NewClient(suite.Client,
				types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
			)
			require.NoError(t, err)

			promQL := `envoy_cluster_upstream_rq_retry_overflow{app_kubernetes_io_name="envoy",app_kubernetes_io_managed_by="envoy-gateway"}`

			// expect 503 since the policy applies a retry budget with 0% success rate
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/status/503",
				},
				Response: http.Response{
					StatusCodes: []int{503},
				},
				Namespace: ns,
			})

			err = wait.PollUntilContextTimeout(t.Context(), time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
				v, err := promClient.QuerySum(ctx, promQL)
				if err != nil {
					tlog.Logf(t, "failed to query prometheus: %v", err)
					return false, nil
				}

				if v > 0 {
					tlog.Logf(t, "got expected retry overflow metric: %v", v)
					return true, nil
				}

				tlog.Logf(t, "retry overflow metric not updated yet: %v", v)
				return false, nil
			})
			require.NoError(t, err)
		})
	},
}
