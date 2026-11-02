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
	ConformanceTests = append(ConformanceTests, ExtensionServerVHostPolicyTest)
}

var ExtensionServerVHostPolicyTest = suite.ConformanceTest{
	ShortName:   "ExtensionServerVHostPolicy",
	Description: "Extension server injects a response header from a VHostMetadataPolicy attached to a listener",
	Manifests:   []string{"testdata/extension-server-vhost-policy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "extension-server-vhost-policy", Namespace: ns}
		gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
			http.ExpectedResponse{
				Request: http.Request{
					Host: "vhost-policy.example.io",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-ext-server-policy": "vhost-policy-applied",
					},
				},
				Namespace: ns,
			},
		)
	},
}
