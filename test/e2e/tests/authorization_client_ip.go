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
	ConformanceTests = append(ConformanceTests, AuthorizationClientIPTest, AuthorizationClientIPTrustedCidrsTest)
}

var AuthorizationClientIPTest = suite.ConformanceTest{
	ShortName:   "AuthzWithClientIP",
	Description: "Authorization with client IP Allow/Deny list",
	Manifests:   []string{"testdata/authorization-client-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		route1NN := types.NamespacedName{Name: "http-with-authorization-client-ip-1", Namespace: ns}
		route2NN := types.NamespacedName{Name: "http-with-authorization-client-ip-2", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), route1NN, route2NN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-client-ip-1", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-client-ip-2", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("first route-denied IP", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected1",
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.1.1", // in the denied list
					},
				},
				Response: http.Response{
					StatusCodes: []int{403},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("first route-allowed IP", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected1",
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.2.1", // in the allowed list
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    "/protected1",
						Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("first route-default action: allow", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected1",
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.3.1", // not in the denied list
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    "/protected1",
						Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// Test the second route
		t.Run("second route-allowed IP", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected2",
					Headers: map[string]string{
						"X-Forwarded-For": "10.0.1.1", // in the allowed list
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    "/protected2",
						Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("second route-default action: deny", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected2",
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.3.1", // not in the allowed list
					},
				},
				Response: http.Response{
					StatusCodes: []int{403},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}

var AuthorizationClientIPTrustedCidrsTest = suite.ConformanceTest{
	ShortName:   "AuthzWithClientIPTrustedCIDRs",
	Description: "Authorization with client IP Allow/Deny list using trusted CIDRs",
	Manifests:   []string{"testdata/authorization-client-ip-trusted-cidrs.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		route1NN := types.NamespacedName{Name: "http-with-authorization-client-ip-trusted-cidr", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), route1NN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-client-ip-trusted-cidr", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("first route-allowed IP no proxy", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected1",
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.2.1", // not in deny list
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    "/protected1",
						Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("first route-allowed IP with trusted proxies", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected1",
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.2.1,172.16.1.17,10.0.1.12", // first untrusted ip 192.168.2.1 not in deny list
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    "/protected1",
						Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("first route-not allowed IP with an untrusted proxy", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected1",
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.2.1,192.168.1.1,10.0.1.12", // first untrusted ip 192.168.1.1 in deny list
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    "/protected1",
						Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
					},
				},
				Response: http.Response{
					StatusCodes: []int{403},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
