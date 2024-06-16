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
	Test: func(t *testing.T, bSuite *suite.BenchmarkTestSuite) (reports []suite.BenchmarkReport) {
		var (
			ctx            = context.Background()
			ns             = "benchmark-test"
			err            error
			requestHeaders = []string{
				"Host: www.benchmark.com",
			}
		)

		// Setup and create a gateway.
		gatewayNN := types.NamespacedName{Name: "benchmark", Namespace: ns}
		gateway := bSuite.GatewayTemplate.DeepCopy()
		gateway.SetName(gatewayNN.Name)
		err = bSuite.CreateResource(ctx, gateway)
		require.NoError(t, err)

		// Setup and scale httproutes.
		httpRouteName := "benchmark-route-%d"
		httpRouteScales := []uint16{10, 50, 100, 300, 500}
		httpRouteNNs := make([]types.NamespacedName, 0, httpRouteScales[len(httpRouteScales)-1])

		bSuite.RegisterCleanup(t, ctx, gateway, &gwapiv1.HTTPRoute{})

		var start uint16 = 0
		for _, scale := range httpRouteScales {
			err = bSuite.ScaleHTTPRoutes(t, ctx, [2]uint16{start, scale}, httpRouteName, gatewayNN.Name, func(route *gwapiv1.HTTPRoute) {
				routeNN := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}
				httpRouteNNs = append(httpRouteNNs, routeNN)
			})
			require.NoError(t, err)
			start = scale

			gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, bSuite.Client, bSuite.TimeoutConfig, bSuite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), httpRouteNNs...)

			// Run benchmark test at different scale.
			name := fmt.Sprintf("scale-httproutes-%d", scale)
			report, err := bSuite.Benchmark(t, ctx, name, gatewayAddr, requestHeaders...)
			require.NoError(t, err)

			report.Print(t, name)
			reports = append(reports, *report)
		}

		return
	},
}
