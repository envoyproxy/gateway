// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from upstream gateway-api, it will be moved to upstream.

//go:build e2e

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, RouteStatusCleanupSameGatewayClass, RouteStatusCleanupMultipleGatewayClasses)
}

var RouteStatusCleanupSameGatewayClass = suite.ConformanceTest{
	ShortName:   "RouteStatusCleanupSameGatewayClass",
	Description: "Testing Route Status Cleanup With Parents of The Same GatewayClass",
	Manifests:   []string{"testdata/route-status-cleanup-multiple-parents-same-gwc.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("RouteStatusCleanup", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "tcp-route-status-cleanup", Namespace: ns}
			gw1NN, gw2NN := types.NamespacedName{Name: "gateway-1", Namespace: ns}, types.NamespacedName{Name: "gateway-2", Namespace: ns}
			gwRefs := []GatewayRef{NewGatewayRef(gw1NN), NewGatewayRef(gw2NN)}
			gwAddrs := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, gwRefs, routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			require.Len(t, gwAddrs, 2)
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddrs[0], OkResp)
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddrs[1], OkResp)

			// Change the route to have a single parent, and check its status
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route-status-cleanup-single-parent.yaml", false)
			gwRefs = []GatewayRef{NewGatewayRef(gw1NN)}
			gwAddrs = GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, gwRefs, routeNN)

			// Send a request to an valid path and expect a successful response
			require.Len(t, gwAddrs, 1)
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddrs[0], OkResp)
		})
	},
}

var RouteStatusCleanupMultipleGatewayClasses = suite.ConformanceTest{
	ShortName:   "RouteStatusCleanupMultipleGatewayClasses",
	Description: "Testing Route Status Cleanup With Parents of Multiple GatewayClasses",
	Manifests:   []string{"testdata/route-status-cleanup-multiple-parents-multiple-gwcs.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("RouteStatusCleanup", func(t *testing.T) {
			// Create the second gateway of a different gatewayclass, which the route is already attached to.
			prevGwc := suite.Applier.GatewayClass
			suite.Applier.GatewayClass = "status-cleanup"
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/status-cleanup-gateway-different-gwc.yaml", true)
			suite.Applier.GatewayClass = prevGwc

			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "tcp-route-status-cleanup", Namespace: ns}
			gw1NN, gw2NN := types.NamespacedName{Name: "gateway-1", Namespace: ns}, types.NamespacedName{Name: "gateway-2", Namespace: ns}
			gwRefs := []GatewayRef{NewGatewayRef(gw1NN), NewGatewayRef(gw2NN)}
			gwAddrs := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, gwRefs, routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			require.Len(t, gwAddrs, 2)
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddrs[0], OkResp)
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddrs[1], OkResp)

			// Change the route to have a single parent, and check its status
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route-status-cleanup-single-parent.yaml", false)
			gwRefs = []GatewayRef{NewGatewayRef(gw1NN)}
			gwAddrs = GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, gwRefs, routeNN)

			// Send a request to an valid path and expect a successful response
			require.Len(t, gwAddrs, 1)
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddrs[0], OkResp)
		})
	},
}
