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
	Test: func(t *testing.T, bSuite *suite.BenchmarkTestSuite) {
		var (
			ctx            = context.Background()
			ns             = "benchmark-test"
			err            error
			requestHeaders = []string{
				"Host: www.benchmark.com",
			}
			routeNameFormat = "benchmark-route-%d"
		)

		gatewayNN := types.NamespacedName{Name: "benchmark", Namespace: ns}
		gateway := bSuite.GatewayTemplate.DeepCopy()
		gateway.SetName(gatewayNN.Name)
		err = bSuite.CreateResource(ctx, gateway)
		require.NoError(t, err)

		bSuite.RegisterCleanup(t, ctx, gateway, &gwapiv1.HTTPRoute{})

		t.Run("scaling up httproutes", func(t *testing.T) {
			routeScales := []uint16{10, 20}
			routeNNs := make([]types.NamespacedName, 0, routeScales[len(routeScales)-1])

			var start uint16 = 0
			for _, scale := range routeScales {
				t.Run(fmt.Sprintf("scaling up httproutes to %d", scale), func(t *testing.T) {
					err = bSuite.ScaleUpHTTPRoutes(ctx, [2]uint16{start, scale}, routeNameFormat, gatewayNN.Name, func(route *gwapiv1.HTTPRoute) {
						routeNN := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}
						routeNNs = append(routeNNs, routeNN)
					})
					require.NoError(t, err)
					start = scale

					gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, bSuite.Client, bSuite.TimeoutConfig,
						bSuite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), routeNNs...)

					// Run benchmark test at different scale.
					name := fmt.Sprintf("scale-up-httproutes-%d", scale)
					err = bSuite.Benchmark(t, ctx, name, gatewayAddr, requestHeaders...)
					require.NoError(t, err)
				})
			}
		})

		// TODO: implement scaling down httproutes

		return
	},
}
