// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyGatewayDynamicResolverBackendTest)
}

var EnvoyGatewayDynamicResolverBackendTest = suite.ConformanceTest{
	ShortName:   "EnvoyGatewayDynamicResolverBackend",
	Description: "Routes with a backend ref to a dynamic resolver backend",
	Manifests: []string{
		"testdata/httproute-with-dynamic-resolver-backend.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "httproute-with-dynamic-resolver-backend", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-dynamic-resolver", Namespace: ns})

		t.Run("route to service foo", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "test-service-foo.gateway-conformance-infra",
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
		t.Run("route to service bar", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "test-service-bar.gateway-conformance-infra",
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
