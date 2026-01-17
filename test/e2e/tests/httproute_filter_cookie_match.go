// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteFilterCookieMatchTest)
}

var HTTPRouteFilterCookieMatchTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteFilterCookieMatch",
	Description: "HTTPRouteFilter cookie matches are ANDed with HTTPRoute rule matches",
	Manifests:   []string{"testdata/httproute-filter-cookie-match.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		t.Run("rule and filter matches", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "httproute-cookie-rule-filter", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			okReq := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cookie-rule",
					Headers: map[string]string{
						"Cookie": "choco=chip",
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okReq)

			missReq := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cookie-rule",
				},
				Response: http.Response{
					StatusCodes: []int{404},
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, missReq)
		})

		t.Run("filter matches only", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "httproute-cookie-filter-only", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			okReq := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cookie-filter-only",
					Headers: map[string]string{
						"Cookie": "choco=chip",
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okReq)

			missReq := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cookie-filter-only",
				},
				Response: http.Response{
					StatusCodes: []int{404},
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, missReq)
		})
	},
}
