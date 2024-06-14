// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"

	"github.com/envoyproxy/gateway/test/benchmark/suite"
)

func init() {
	BenchmarkTests = append(BenchmarkTests, ScaleHTTPRoutes)
}

var ScaleHTTPRoutes = suite.BenchmarkTest{
	ShortName:   "ScaleHTTPRoute",
	Description: "Fixed one Gateway and different number of HTTPRoutes on a scale of (10, 50, 100, 300, 500).",
	Test: func(t *testing.T, suite *suite.BenchmarkTestSuite) {
		var (
			ctx            = context.Background()
			ns             = "benchmark-test"
			err            error
			requestHeaders = []string{
				"Host: www.benchmark.com",
				"Host: x-nighthawk-test-server-config: {response_body_size:20, static_delay:\"0s\"}",
			}
		)

		// Setup and create a gateway.
		gatewayNN := types.NamespacedName{Name: "benchmark", Namespace: ns}
		gateway := suite.GatewayTemplate.DeepCopy()
		gateway.SetName(gatewayNN.Name)
		err = suite.CreateResource(ctx, gateway)
		require.NoError(t, err)

		// Setup and scale httproutes.
		httpRouteName := "benchmark-route-%d"
		httpRouteScales := []uint16{5, 10, 20} // TODO: for test, update later
		httpRouteNNs := make([]types.NamespacedName, 0, httpRouteScales[len(httpRouteScales)-1])

		suite.RegisterCleanup(t, ctx, gateway, &gwapiv1.HTTPRoute{})

		var start uint16 = 0
		for _, scale := range httpRouteScales {
			t.Logf("Scaling HTTPRoutes to %d", scale)

			err = suite.ScaleHTTPRoutes(t, ctx, [2]uint16{start, scale}, httpRouteName, gatewayNN.Name, func(route *gwapiv1.HTTPRoute) {
				routeNN := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}
				httpRouteNNs = append(httpRouteNNs, routeNN)
			})
			require.NoError(t, err)
			start = scale

			gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), httpRouteNNs...)

			// Run benchmark test at different scale.
			name := fmt.Sprintf("scale-httproutes-%d", scale)
			err = suite.Benchmark(t, ctx, name, gatewayAddr, requestHeaders...)
			require.NoError(t, err)

			// TODO: Scrape the benchmark results.
		}
	},
}
