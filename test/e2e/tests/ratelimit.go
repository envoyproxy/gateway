// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

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
	ConformanceTests = append(ConformanceTests, RateLimitBlockAllIPs)
}

var RateLimitBlockAllIPs = suite.ConformanceTest{
	ShortName:   "RateLimitBlockAllIPs",
	Description: "Limit all requests",
	Manifests:   []string{"testdata/ratelimit-block-all-ips.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("RateLimitBlockAllIPs", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-ratelimit", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 429,
				},
				Namespace: ns,
			})
		})
	},
}
