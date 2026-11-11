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
	ConformanceTests = append(ConformanceTests, EnvoyGatewayBackendShortDNSTest)
}

var EnvoyGatewayBackendShortDNSTest = suite.ConformanceTest{
	ShortName:   "EnvoyGatewayBackendShortDNS",
	Description: "Route to a Backend with a single-label DNS endpoint",
	Manifests:   []string{"testdata/httproute-to-backend-short-dns.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		gwNN := types.NamespacedName{Name: "gateway-short-dns", Namespace: ns}
		routeNN := types.NamespacedName{Name: "httproute-backend-short-dns", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-short-dns", Namespace: ns})
		WaitForGatewayPodsReady(t, suite.Client, gwNN)

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path: "/backend-short-dns",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Backend:   "infra-backend-v1",
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
	},
}
