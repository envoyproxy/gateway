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
	ConformanceTests = append(ConformanceTests, EnvoyPatchPolicyTest)
}

var EnvoyPatchPolicyTest = suite.ConformanceTest{
	ShortName:   "EnvoyPatchPolicy",
	Description: "update xds using EnvoyPatchPolicy",
	Manifests:   []string{"testdata/envoy-patch-policy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("envoy patch policy", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-envoy-patch-policy", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/foo",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

			customResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/bar",
				},
				Response: http.Response{
					StatusCode: 406,
				},
				Namespace: ns,
			}

			// Send a request to an invalid path and expect a custom response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, customResp)
		})
	},
}
