// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"

	"github.com/envoyproxy/gateway/test/benchmark/suite"
)

func init() {
	BenchmarkTests = append(BenchmarkTests, ScaleDeployments)
}

var ScaleDeployments = suite.BenchmarkTest{
	ShortName: "ScaleDeployments",
	Description: `Fixed 1 Gateway, 500 HTTPRoutes, and different scales of Deployments (range from 10 to 250) with 2 replicas, 
each Deployment only associated with 1 hostname and Service, Deployments and hostnames are evenly distributed by routes.`,
	Test: func(t *testing.T, bSuite *suite.BenchmarkTestSuite) (reports []*suite.BenchmarkReport) {
		var (
			ctx                       = context.Background()
			ns                        = "benchmark-test"
			totalRoutes        uint16 = 500
			deploymentReplicas int32  = 2
			err                error
		)

		routeNameFormat := "benchmark-route-%d"
		routeHostnameFormat := "www.benchmark-%d.com"
		gatewayNN := types.NamespacedName{Name: "benchmark", Namespace: ns}
		gateway := bSuite.GatewayTemplate.DeepCopy()
		gateway.SetName(gatewayNN.Name)
		err = bSuite.CreateResource(ctx, gateway)
		require.NoError(t, err)

		deploymentsScales := []uint16{10, 50, 100, 250}
		deploymentsScalesN := len(deploymentsScales)

		bSuite.RegisterCleanup(t, ctx, gateway, &appsv1.Deployment{}, &corev1.Service{})

		t.Run("scaling up deployments", func(t *testing.T) {
			var start uint16 = 0
			for _, scale := range deploymentsScales {
				routePerHost := totalRoutes / scale
				testName := fmt.Sprintf("scaling up deployments from %d to %d with %d routes per hostname and service", start, scale, routePerHost)

				t.Run(testName, func(t *testing.T) {
					err = bSuite.ScaleUpDeployments(ctx, [2]uint16{start, scale}, deploymentReplicas, nil)
					require.NoError(t, err)

					// Setup httproutes to 1k
					gatewayAddr := waitAndScaleHTTPRoutesToN(t, bSuite, ctx, routeNameFormat, routeHostnameFormat, gatewayNN, totalRoutes, routePerHost)

					// Run benchmark test at different scale.
					start = scale
					jobName := fmt.Sprintf("scale-up-deployments-%d", scale)
					report, err := bSuite.Benchmark(t, ctx, jobName, testName, gatewayAddr, routeHostnameFormat, int(routePerHost))
					require.NoError(t, err)
					reports = append(reports, report)

					// Clean httproutes for next scale case
					err = bSuite.ScaleDownHTTPRoutes(ctx, [2]uint16{totalRoutes, 0}, routeNameFormat, gatewayNN.Name, nil)
					require.NoError(t, err)
				})
			}
		})

		t.Run("scaling down deployments", func(t *testing.T) {
			start := deploymentsScales[deploymentsScalesN-1]
			for i := deploymentsScalesN - 2; i >= 0; i-- {
				scale := deploymentsScales[i]
				routePerHost := totalRoutes / scale
				testName := fmt.Sprintf("scaling down deployments from %d to %d with %d routes per hostname and service", start, scale, routePerHost)

				t.Run(testName, func(t *testing.T) {
					err = bSuite.ScaleDownDeployments(ctx, [2]uint16{start, scale}, nil)
					require.NoError(t, err)

					// Setup httproutes to 1k
					gatewayAddr := waitAndScaleHTTPRoutesToN(t, bSuite, ctx, routeNameFormat, routeHostnameFormat, gatewayNN, totalRoutes, routePerHost)

					// Run benchmark test at different scale.
					jobName := fmt.Sprintf("scale-down-deployments-%d", scale)
					report, err := bSuite.Benchmark(t, ctx, jobName, testName, gatewayAddr, routeHostnameFormat, int(routePerHost))
					require.NoError(t, err)
					reports = append(reports, report)

					// Clean httproutes for next scale case
					err = bSuite.ScaleDownHTTPRoutes(ctx, [2]uint16{totalRoutes, 0}, routeNameFormat, gatewayNN.Name, nil)
					require.NoError(t, err)
				})
			}
		})

		return
	},
}

func waitAndScaleHTTPRoutesToN(t *testing.T, bSuite *suite.BenchmarkTestSuite, ctx context.Context, nameFormat, hostnameFormat string, gatewayNN types.NamespacedName, n, routePerHost uint16) string {
	routeNNs := make([]types.NamespacedName, 0, n)
	err := bSuite.ScaleUpHTTPRoutes(ctx, [2]uint16{0, n}, nameFormat, hostnameFormat, gatewayNN.Name, routePerHost, func(route *gwapiv1.HTTPRoute) {
		routeNN := types.NamespacedName{Name: route.Name, Namespace: route.Namespace}
		routeNNs = append(routeNNs, routeNN)

		t.Logf("Create HTTPRoute: %s with hostname %s and backendRef %s",
			routeNN.String(), route.Spec.Hostnames[0], route.Spec.Rules[0].BackendRefs[0].Name)
	})
	require.NoError(t, err)

	gatewayAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, bSuite.Client, bSuite.TimeoutConfig,
		bSuite.ControllerName, kubernetes.NewGatewayRef(gatewayNN), routeNNs...)
	return gatewayAddr
}
