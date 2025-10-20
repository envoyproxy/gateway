// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package tests

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/stat"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/benchmark/suite"
)

func init() {
	BenchmarkTests = append(BenchmarkTests, ScaleHTTPRoutes)
}

var ScaleHTTPRoutes = suite.BenchmarkTest{
	ShortName:   "ScaleHTTPRoute",
	Description: "Fixed one Gateway and different scales of HTTPRoutes with different portion of hostnames.",
	Test: func(t *testing.T, bSuite *suite.BenchmarkTestSuite) (reports []*suite.BenchmarkReport) {
		var (
			ctx               = context.Background()
			ns                = "benchmark-test"
			totalHosts uint16 = 5
			err        error
		)

		gatewayNN := types.NamespacedName{Name: "benchmark", Namespace: ns}
		gateway := bSuite.GatewayTemplate.DeepCopy()
		gateway.SetName(gatewayNN.Name)
		err = bSuite.CreateResource(ctx, gateway)
		require.NoError(t, err)

		routeNameFormat := "benchmark-route-%d"
		routeHostnameFormat := "www.benchmark-%d.com"
		routeScales := []uint16{10, 50, 100, 300, 500, 1000}
		routeScalesN := len(routeScales)
		routeNNs := make([]types.NamespacedName, 0, routeScales[routeScalesN-1])

		bSuite.RegisterCleanup(t, ctx, gateway, &gwapiv1.HTTPRoute{})

		t.Run("scaling up httproutes", func(t *testing.T) {
			var start, batch uint16 = 0, 0
			for _, scale := range routeScales {
				routePerHost := scale / totalHosts
				testName := fmt.Sprintf("scaling up httproutes to %d with %d routes per hostname", scale, routePerHost)

				r := roundtripper.DefaultRoundTripper{
					Debug:         true,
					TimeoutConfig: bSuite.TimeoutConfig,
				}
				t.Run(testName, func(t *testing.T) {
					tlog.Logf(t, "Start scaling up HTTPRoutes to %d with %d routes per hostname", scale, routePerHost)
					startTime := time.Now()
					convergenceTimeByHost := map[string]time.Duration{}
					err = bSuite.ScaleUpHTTPRoutes(ctx, [2]uint16{start, scale}, routeNameFormat, routeHostnameFormat, gatewayNN.Name, routePerHost-batch,
						func(route *gwapiv1.HTTPRoute, applyAt time.Time) {
							routeNN := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}
							routeNNs = append(routeNNs, routeNN)
							host := string(route.Spec.Hostnames[0])
							gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, bSuite.Client, bSuite.TimeoutConfig,
								bSuite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), routeNN)

							req := http.MakeRequest(t, &http.ExpectedResponse{
								Request: http.Request{
									Host: host,
									Path: "/",
								},
								Response: http.Response{
									StatusCodes: []int{200},
								},
							}, gwAddr, "HTTP", "HTTP")
							http.WaitForConsistentResponse(t, &r, req, http.ExpectedResponse{
								Response: http.Response{
									StatusCodes: []int{200},
								},
							}, bSuite.TimeoutConfig.RequiredConsecutiveSuccesses, bSuite.TimeoutConfig.MaxTimeToConsistency)

							d := time.Since(applyAt)
							convergenceTimeByHost[host] = d

							t.Logf("Create HTTPRoute: %s with hostname %s, and became ready after %s", routeNN.String(), host, d)
						})
					require.NoError(t, err)
					start = scale
					batch = routePerHost

					// Check if we have convergence time for all hosts.
					if len(convergenceTimeByHost) != int(totalHosts) {
						t.Fatalf("Expected convergence time for %d hosts, but got %d", totalHosts, len(convergenceTimeByHost))
					}

					gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, bSuite.Client, bSuite.TimeoutConfig,
						bSuite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), routeNNs...)

					// Run benchmark test at different scale.
					jobName := fmt.Sprintf("scale-up-httproutes-%d", scale)
					report, err := bSuite.Benchmark(t, ctx, jobName, testName, gatewayAddr, routeHostnameFormat, int(totalHosts), startTime)
					require.NoError(t, err)
					report.RouteConvergence = getRouteConvergenceDuration(convergenceTimeByHost)

					reports = append(reports, report)
				})
			}
		})

		t.Run("scaling down httproutes", func(t *testing.T) {
			start := routeScales[routeScalesN-1]
			for i := routeScalesN - 2; i >= 0; i-- {
				scale := routeScales[i]
				routePerHost := scale / totalHosts
				testName := fmt.Sprintf("scaling down httproutes to %d with %d routes per hostname", scale, routePerHost)

				t.Run(testName, func(t *testing.T) {
					startTime := time.Now()
					err = bSuite.ScaleDownHTTPRoutes(ctx, [2]uint16{start, scale}, routeNameFormat, gatewayNN.Name, func(route *gwapiv1.HTTPRoute) {
						routeNN := routeNNs[len(routeNNs)-1]
						routeNNs = routeNNs[:len(routeNNs)-1]

						// Making sure we are deleting the right one route.
						require.Equal(t, types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, routeNN)

						t.Logf("Delete HTTPRoute: %s", routeNN.String())
					})
					require.NoError(t, err)
					start = scale

					gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, bSuite.Client, bSuite.TimeoutConfig,
						bSuite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), routeNNs...)

					// Run benchmark test at different scale.
					jobName := fmt.Sprintf("scale-down-httproutes-%d", scale)
					report, err := bSuite.Benchmark(t, ctx, jobName, testName, gatewayAddr, routeHostnameFormat, int(totalHosts), startTime)
					require.NoError(t, err)

					reports = append(reports, report)
				})
			}
		})

		return
	},
}

func getRouteConvergenceDuration(durations map[string]time.Duration) *suite.PerfDuration {
	weights := make([]float64, 0, len(durations))
	for _, d := range durations {
		weights = append(weights, float64(d.Microseconds()))
	}
	sort.Float64s(weights)

	return &suite.PerfDuration{
		P99: convertFloat64ToDuration(stat.Quantile(0.99, stat.Empirical, weights, nil)),
		P90: convertFloat64ToDuration(stat.Quantile(0.9, stat.Empirical, weights, nil)),
		P50: convertFloat64ToDuration(stat.Quantile(0.5, stat.Empirical, weights, nil)),
	}
}

func convertFloat64ToDuration(f float64) time.Duration {
	return time.Duration(f) * time.Microsecond
}
