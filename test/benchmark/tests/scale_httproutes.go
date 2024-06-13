// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package tests

import (
	"context"
	"testing"

	"github.com/envoyproxy/gateway/test/benchmark/suite"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
)

func init() {
	BenchmarkTests = append(BenchmarkTests, ScaleHTTPRoutes)
}

var ScaleHTTPRoutes = suite.BenchmarkTest{
	ShortName:   "ScaleHTTPRoute",
	Description: "Fixed one Gateway with different scale of HTTPRoutes: [10, 50, 100, 300, 500].",
	Test: func(t *testing.T, suite *suite.BenchmarkTestSuite) {
		var (
			ctx = context.Background()
			ns  = "benchmark-test"
			err error
		)

		// Setup and create a gateway.
		gatewayNN := types.NamespacedName{Name: "benchmark", Namespace: ns}
		gateway := suite.GatewayTemplate.DeepCopy()
		gateway.SetName(gatewayNN.Name)
		err = suite.CreateResource(ctx, gateway)
		require.NoError(t, err)

		// Setup and scale httproutes.
		httpRouteName := "benchmark-route-%d"
		httpRouteScales := []uint16{5, 10, 20}
		httpRouteNNs := make([]types.NamespacedName, 0, httpRouteScales[len(httpRouteScales)-1])

		var start uint16 = 0
		for _, scale := range httpRouteScales {
			t.Logf("Scaling HTTPRoutes to %d", scale)

			err = suite.ScaleHTTPRoutes(ctx, [2]uint16{start, scale}, httpRouteName, gatewayNN.Name, func(route *gwapiv1.HTTPRoute) {
				routeNN := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}
				httpRouteNNs = append(httpRouteNNs, routeNN)
			})
			require.NoError(t, err)
			start = scale

			kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), httpRouteNNs...)

			// TODO: run benchmark and report
		}

		// Cleanup
		err = suite.CleanupResource(ctx, gateway)
		require.NoError(t, err)

		err = suite.CleanupScaledResources(ctx, &gwapiv1.HTTPRoute{})
		require.NoError(t, err)
	},
}
