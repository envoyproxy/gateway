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
	ConformanceTests = append(ConformanceTests, AuthorizationPathTest)
}

var AuthorizationPathTest = suite.ConformanceTest{
	ShortName:   "AuthzWithPath",
	Description: "Authorization with HTTP path match (PathPrefix and RegularExpression with invert)",
	Manifests:   []string{"testdata/authorization-path.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		prefixRouteNN := types.NamespacedName{Name: "http-with-authorization-path-prefix", Namespace: ns}
		invertRouteNN := types.NamespacedName{Name: "http-with-authorization-path-regex-invert", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), prefixRouteNN, invertRouteNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-path-prefix", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-path-regex-invert", Namespace: ns}, suite.ControllerName, ancestorRef)

		// PathPrefix tests
		t.Run("path prefix match with correct header should be allowed", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/auth-path-prefix/public/resource",
					Headers: map[string]string{
						"x-user-id": "john",
					},
				},
				Response:  http.Response{StatusCodes: []int{200}},
				Namespace: ns,
			})
		})

		t.Run("path prefix match with another allowed header value should be allowed", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/auth-path-prefix/public/resource",
					Headers: map[string]string{
						"x-user-id": "alice",
					},
				},
				Response:  http.Response{StatusCodes: []int{200}},
				Namespace: ns,
			})
		})

		t.Run("path prefix match with wrong header value should be denied", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/auth-path-prefix/public/resource",
					Headers: map[string]string{
						"x-user-id": "eve",
					},
				},
				Response:  http.Response{StatusCodes: []int{403}},
				Namespace: ns,
			})
		})

		t.Run("path prefix match with no header should be denied", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Path: "/auth-path-prefix/public/resource"},
				Response:  http.Response{StatusCodes: []int{403}},
				Namespace: ns,
			})
		})

		t.Run("path not matching prefix with correct header should be denied", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/auth-path-prefix/private/resource",
					Headers: map[string]string{
						"x-user-id": "john",
					},
				},
				Response:  http.Response{StatusCodes: []int{403}},
				Namespace: ns,
			})
		})

		// Regex + invert tests
		t.Run("path not matching regex (invert) with correct header should be allowed", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/auth-path-invert/public/resource",
					Headers: map[string]string{
						"x-role": "admin",
					},
				},
				Response:  http.Response{StatusCodes: []int{200}},
				Namespace: ns,
			})
		})

		t.Run("path matching regex (invert) with correct header should be denied", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/auth-path-invert/secret/data",
					Headers: map[string]string{
						"x-role": "admin",
					},
				},
				Response:  http.Response{StatusCodes: []int{403}},
				Namespace: ns,
			})
		})

		t.Run("path not matching regex (invert) with wrong header should be denied", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/auth-path-invert/public/resource",
					Headers: map[string]string{
						"x-role": "user",
					},
				},
				Response:  http.Response{StatusCodes: []int{403}},
				Namespace: ns,
			})
		})
	},
}
