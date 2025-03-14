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
	ConformanceTests = append(ConformanceTests, CredentialInjectionTest)
}

var CredentialInjectionTest = suite.ConformanceTest{
	ShortName:   "CredentialInjection",
	Description: "Resource with CredentialInjection enabled",
	Manifests:   []string{"testdata/credential-injection.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("inject credential to the default Authorization header", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "credential-injection", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/foo",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Headers: map[string]string{
							"Authorization": "Basic dXNlcjE6dGVzdDE=",
						},
						Path: "/foo",
					},
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
