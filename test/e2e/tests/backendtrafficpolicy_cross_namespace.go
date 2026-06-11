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
	ConformanceTests = append(ConformanceTests, BackendTrafficPolicyCrossNamespaceTest)
}

var BackendTrafficPolicyCrossNamespaceTest = suite.ConformanceTest{
	ShortName:   "BackendTrafficPolicyCrossNamespace",
	Description: "Test cross-namespace BackendTrafficPolicy attachment to HTTPRoute with and without ReferenceGrant",
	Manifests:   []string{"testdata/backendtrafficpolicy-cross-namespace.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("BackendTrafficPolicyCrossNamespace", func(t *testing.T) {
			policyNS := "gateway-conformance-infra"
			grantedNS := "btp-cross-ns-granted"
			deniedNS := "btp-cross-ns-denied"

			grantedRouteNN := types.NamespacedName{Name: "cross-namespace-btp", Namespace: grantedNS}
			deniedRouteNN := types.NamespacedName{Name: "cross-namespace-btp", Namespace: deniedNS}
			grantedGatewayNN := types.NamespacedName{Name: "cross-namespace-btp", Namespace: grantedNS}
			deniedGatewayNN := types.NamespacedName{Name: "cross-namespace-btp", Namespace: deniedNS}

			grantedAddr := kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(grantedGatewayNN), &gwapiv1.HTTPRoute{}, false, grantedRouteNN,
			)
			deniedAddr := kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(deniedGatewayNN), &gwapiv1.HTTPRoute{}, false, deniedRouteNN,
			)

			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, grantedRouteNN, grantedGatewayNN)
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, deniedRouteNN, deniedGatewayNN)

			grantedAncestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(grantedGatewayNN.Namespace),
				Name:      gwapiv1.ObjectName(grantedGatewayNN.Name),
			}

			BackendTrafficPolicyMustBeAccepted(
				t,
				suite.Client,
				types.NamespacedName{Name: "cross-namespace-btp-granted", Namespace: policyNS},
				suite.ControllerName,
				grantedAncestorRef,
			)

			grantedResponse := http.ExpectedResponse{
				Namespace: grantedNS,
				Request: http.Request{
					Host: "granted.cross-namespace-btp.example.com",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{418},
				},
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, grantedAddr, grantedResponse)

			deniedResponse := http.ExpectedResponse{
				// Set the expected namespace to empty string since the test uses a direct response HTTP filter.
				// The CapturedRequest is empty since the response is directly returned by the HTTP filter without forwarding to the backend.
				Namespace: "",
				Request: http.Request{
					Host: "denied.cross-namespace-btp.example.com",
					Path: "/",
				},
				// Set the expected request properties to empty strings since the test uses a direct response HTTP filter.
				// The CapturedRequest is empty since the response is directly returned by the HTTP filter without forwarding to the backend.
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host:    "",
						Method:  "",
						Path:    "",
						Headers: nil,
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, deniedAddr, deniedResponse)
		})
	},
}
