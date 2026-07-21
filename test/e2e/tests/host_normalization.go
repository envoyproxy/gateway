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

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, HostNormalizationTest)
}

var HostNormalizationTest = suite.ConformanceTest{
	ShortName:   "HostNormalization",
	Description: "Strip the trailing dot from the Host header before route matching",
	Manifests:   []string{"testdata/host-normalization.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-host-normalization", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "strip-trailing-host-dot-ctp", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("host without trailing dot should match the route", func(t *testing.T) {
			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "host-normalization.example.com",
					Path: "/host-normalization",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})

		t.Run("host with trailing dot should be normalized and match the route", func(t *testing.T) {
			expected := http.ExpectedResponse{
				Request: http.Request{
					Host: "host-normalization.example.com.",
					Path: "/host-normalization",
				},
				// The trailing dot is stripped before routing, so the backend
				// receives the normalized Host header.
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "host-normalization.example.com",
						Path: "/host-normalization",
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})
	},
}
