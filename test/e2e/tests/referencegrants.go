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
	ConformanceTests = append(ConformanceTests, MultiReferenceGrantsSameNamespaceTest)
}

var MultiReferenceGrantsSameNamespaceTest = suite.ConformanceTest{
	ShortName:   "MultiReferenceGrantsSameNamespace",
	Description: "Test for multiple reference grants in the same namespace",
	Manifests:   []string{"testdata/multi-referencegrants-same-namespace-services.yaml", "testdata/multi-referencegrants-same-namespace.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		resourceNS := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "multi-referencegrant-same-namespace", Namespace: resourceNS}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: resourceNS}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		targetHost := "multireferencegrant.local"
		targetNS := "multireferencegrants-ns"
		testcases := []http.ExpectedResponse{
			{
				Request: http.Request{
					Host: targetHost,
					Path: "/v1/echo",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Backend:   "app-backend-v1",
				Namespace: targetNS,
			},
			{
				Request: http.Request{
					Host: targetHost,
					Path: "/v2/echo",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Backend:   "app-backend-v2",
				Namespace: targetNS,
			},
			{
				Request: http.Request{
					Host: targetHost,
					Path: "/v3/echo",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Backend:   "app-backend-v3",
				Namespace: targetNS,
			},
		}

		for i, tc := range testcases {
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
