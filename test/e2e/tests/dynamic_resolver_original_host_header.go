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
	ConformanceTests = append(ConformanceTests, DynamicResolverOriginalHostHeaderTest)
}

var DynamicResolverOriginalHostHeaderTest = suite.ConformanceTest{
	ShortName:   "DynamicResolverOriginalHostHeader",
	Description: "Verify dynamic resolver host rewrites preserve the X-ENVOY-ORIGINAL-HOST request header when Envoy headers are enabled",
	Manifests:   []string{"testdata/dynamic-resolver-original-host-header.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		routeNN := types.NamespacedName{Name: "dynamic-resolver-original-host-header", Namespace: ns}
		backendNN := types.NamespacedName{Name: "backend-dynamic-resolver-original-host", Namespace: ns}

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		BackendMustBeAccepted(t, suite.Client, backendNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "dynamic-resolver-original-host-ctp", Namespace: ns}, suite.ControllerName, ancestorRef)

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Host: "www.example.com",
				Path: "/original-host",
			},
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Host: "test-service-foo.gateway-conformance-infra.svc.cluster.local",
					Path: "/original-host",
					Headers: map[string]string{
						"x-envoy-original-host": "www.example.com",
					},
				},
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
	},
}
