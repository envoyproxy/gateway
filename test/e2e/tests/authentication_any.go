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
	ConformanceTests = append(ConformanceTests, AuthenticationAnyTest)
}

var AuthenticationAnyTest = suite.ConformanceTest{
	ShortName:   "AuthenticationAny",
	Description: "SecurityPolicy with authentication mode Any accepts JWT or Basic credentials",
	Manifests:   []string{"testdata/authentication-any.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "authentication-any", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authentication-any", Namespace: ns}, suite.ControllerName, gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		})

		const (
			validBasic   = "Basic dXNlcjE6dGVzdDE=" // user1:test1
			invalidBasic = "Basic dXNlcjE6d3Jvbmc=" // user1:wrong
		)
		for _, test := range []struct {
			name          string
			path          string
			authorization string
			status        int
		}{
			{name: "valid JWT", path: "/authentication-any", authorization: "Bearer " + v1Token, status: 200},
			{name: "valid Basic", path: "/authentication-any", authorization: validBasic, status: 200},
			{name: "no credentials", path: "/authentication-any", status: 403},
			{name: "invalid Basic", path: "/authentication-any", authorization: invalidBasic, status: 401},
			{name: "invalid JWT", path: "/authentication-any", authorization: "Bearer " + invalidToken, status: 401},
			{name: "valid Basic with invalid JWT", path: "/authentication-any?access_token=" + invalidToken, authorization: validBasic, status: 401},
			{name: "valid JWT with invalid Basic", path: "/authentication-any?access_token=" + v1Token, authorization: invalidBasic, status: 401},
		} {
			t.Run(test.name, func(t *testing.T) {
				request := http.Request{Path: test.path}
				if test.authorization != "" {
					request.Headers = map[string]string{"Authorization": test.authorization}
				}
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
					Request:   request,
					Response:  http.Response{StatusCodes: []int{test.status}},
					Namespace: ns,
				})
			})
		}
	},
}
