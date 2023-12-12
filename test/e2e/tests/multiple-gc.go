// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from upstream gateway-api, it will be moved to upstream.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, MultipleGCTest)
}

var MultipleGCTest = suite.ConformanceTest{
	ShortName:   "MultipleGC",
	Description: "Testing multiple GatewayClass with the same controller",
	Manifests:   []string{"testdata/multiple-gc.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("gc-1", func(t *testing.T) {
			ns := "private"
			routeNN := types.NamespacedName{Name: "private-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "private-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
		})
		t.Run("gc-2", func(t *testing.T) {
			ns := "internet"
			routeNN := types.NamespacedName{Name: "internet-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "internet-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
		})

	},
}
