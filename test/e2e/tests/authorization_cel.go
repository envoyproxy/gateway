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
	ConformanceTests = append(ConformanceTests, AuthorizationCELTest)
}

var AuthorizationCELTest = suite.ConformanceTest{
	ShortName:   "AuthzWithCEL",
	Description: "Authorization with CEL expressions",
	Manifests:   []string{"testdata/authorization-cel.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		celOnlyRouteNN := types.NamespacedName{Name: "http-with-authorization-cel", Namespace: ns}
		celAndPrincipalRouteNN := types.NamespacedName{Name: "http-with-authorization-cel-and-principal", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), celOnlyRouteNN, celAndPrincipalRouteNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-cel", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-cel-and-principal", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("cel-only matching path should be allowed", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Path: "/authz-cel/public/resource"},
				Response:  http.Response{StatusCodes: []int{200}},
				Namespace: ns,
			})
		})

		t.Run("cel-only non-matching path should be denied", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Path: "/authz-cel/private/resource"},
				Response:  http.Response{StatusCodes: []int{403}},
				Namespace: ns,
			})
		})

		t.Run("cel and principal should allow when both match", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path:   "/authz-cel-principal",
					Method: "POST",
					Headers: map[string]string{
						"x-team": "team-a",
					},
				},
				Response:  http.Response{StatusCodes: []int{200}},
				Namespace: ns,
			})
		})

		t.Run("cel and principal should deny when cel does not match", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path:   "/authz-cel-principal",
					Method: "GET",
					Headers: map[string]string{
						"x-team": "team-a",
					},
				},
				Response:  http.Response{StatusCodes: []int{403}},
				Namespace: ns,
			})
		})

		t.Run("cel and principal should deny when principal does not match", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path:   "/authz-cel-principal",
					Method: "POST",
					Headers: map[string]string{
						"x-team": "team-b",
					},
				},
				Response:  http.Response{StatusCodes: []int{403}},
				Namespace: ns,
			})
		})
	},
}
