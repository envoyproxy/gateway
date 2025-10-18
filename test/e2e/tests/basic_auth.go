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
	ConformanceTests = append(ConformanceTests, BasicAuthTest)
}

var BasicAuthTest = suite.ConformanceTest{
	ShortName:   "BasicAuth",
	Description: "Resource with BasicAuth enabled",
	Manifests:   []string{"testdata/basic-auth.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("valid username password", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjE6dGVzdDE=", // user1:test1
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("without Authorization header", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
				},
				Response: http.Response{
					StatusCodes: []int{401},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("invalid username password", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjE6dGVzdDI=", // user1:test2
					},
				},
				Response: http.Response{
					StatusCodes: []int{401},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("per route configuration second route", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-2", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-2", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-2",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjQ6dGVzdDQ=", // user4:test4
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// https://github.com/envoyproxy/gateway/issues/2507
		t.Run("request without matching routes ", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/not-matching-route",
				},
				Response: http.Response{
					StatusCodes: []int{404},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// This test is to ensure that the BasicAuth filter is not applied to routes that under the same gateway but
		// do not have the BasicAuth policy applied.
		t.Run("request to an HTTP route without BasicAuth", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-no-basic-auth", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-1", Namespace: ns}, suite.ControllerName, ancestorRef)
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-2", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/no-basic-auth",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// This test verifies that when a user successfully authenticates using BasicAuth,
		// their username is forwarded to the backend as a request header.
		// The SecurityPolicy "basic-auth-3" is configured with the field "forwardUsernameHeader" set to "X-Authenticated-User".
		// The request uses valid credentials, and the test ensures that the response contains
		// the expected status code (200) and includes the "X-Authenticated-User" header with the correct username.
		t.Run("username forwarded as header", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-basic-auth-3", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "basic-auth-3", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-3",
					Headers: map[string]string{
						"Authorization": "Basic dXNlcjQ6dGVzdDQ=", // user4:test4
					},
				},
				// Verify that the http header X-Authenticated-User added before sending it to the backend
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/basic-auth-3",
						Headers: map[string]string{
							"X-Authenticated-User": "user4",
						},
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
