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
	Description: "Fixed one Gateway and different scales of HTTPRoutes.",
	Test: func(t *testing.T, bSuite *suite.BenchmarkTestSuite) (reports []*suite.BenchmarkReport) {
		var (
			ctx            = context.Background()
			ns             = "benchmark-test"
			err            error
			requestHeaders = []string{
				"Host: www.benchmark.com",
			}
		)

		gatewayNN := types.NamespacedName{Name: "benchmark", Namespace: ns}
		gateway := bSuite.GatewayTemplate.DeepCopy()
		gateway.SetName(gatewayNN.Name)
		err = bSuite.CreateResource(ctx, gateway)
		require.NoError(t, err)

		routeNameFormat := "benchmark-route-%d"
		routeScales := []uint16{10, 50, 100, 300, 500}
		routeScalesN := len(routeScales)
		routeNNs := make([]types.NamespacedName, 0, routeScales[routeScalesN-1])

		bSuite.RegisterCleanup(t, ctx, gateway, &gwapiv1.HTTPRoute{})

		t.Run("scaling up httproutes", func(t *testing.T) {
			var start uint16 = 0
			for _, scale := range routeScales {
				t.Run(fmt.Sprintf("scaling up httproutes to %d", scale), func(t *testing.T) {
					err = bSuite.ScaleUpHTTPRoutes(ctx, [2]uint16{start, scale}, routeNameFormat, gatewayNN.Name, func(route *gwapiv1.HTTPRoute) {
						routeNN := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}
						routeNNs = append(routeNNs, routeNN)

						t.Logf("Create HTTPRoute: %s", routeNN.String())
					})
					require.NoError(t, err)
					start = scale

					gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, bSuite.Client, bSuite.TimeoutConfig,
						bSuite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), routeNNs...)

					// Run benchmark test at different scale.
					name := fmt.Sprintf("scale-up-httproutes-%d", scale)
					report, err := bSuite.Benchmark(t, ctx, name, gatewayAddr, requestHeaders...)
					require.NoError(t, err)

					reports = append(reports, report)
				})
			}
		})

		t.Run("scaling down httproutes", func(t *testing.T) {
			start := routeScales[routeScalesN-1]

			for i := routeScalesN - 2; i >= 0; i-- {
				scale := routeScales[i]

				t.Run(fmt.Sprintf("scaling down httproutes to %d", scale), func(t *testing.T) {
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
					name := fmt.Sprintf("scale-down-httproutes-%d", scale)
					report, err := bSuite.Benchmark(t, ctx, name, gatewayAddr, requestHeaders...)
					require.NoError(t, err)

					reports = append(reports, report)
				})
			}
		})

		return
	},
}
