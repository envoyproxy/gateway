// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, CorsTest)
}

var CorsTest = suite.ConformanceTest{
	ShortName:   "Cors",
	Description: "Resource with CORS enabled",
	Manifests:   []string{"testdata/cors.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("should enable cors with Allow Origin Exact", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-cors-exact", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cors-exact", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cors-exact",
					Headers: map[string]string{
						"Origin": "https://www.foo.com",
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"access-control-allow-origin":   "https://www.foo.com",
						"access-control-expose-headers": "x-header-3, x-header-4",
					},
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("should enable cors with Allow Origin Regex", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-cors-exact", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cors-exact", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path:   "/cors-exact",
					Method: "OPTIONS",
					Headers: map[string]string{
						"Origin": "https://anydomain.foobar.com",
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"access-control-allow-origin":   "https://anydomain.foobar.com",
						"access-control-expose-headers": "x-header-3, x-header-4",
					},
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("should not contain cors headers when Origin not registered", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-cors-exact", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cors-exact", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cors-exact",
					Headers: map[string]string{
						"Origin": "https://unknown.foo.com",
					},
				},
				Response: http.Response{
					AbsentHeaders: []string{"access-control-allow-origin"},
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})

		t.Run("should enable cors with wildcard matching", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-cors-wildcard", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cors-wildcard", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cors-exact",
					Headers: map[string]string{
						"Origin": "https://foo.bar.com",
						"Access-control-request-method": "GET",
						"Access-control-request-headers": "x-header-1, x-header-2",
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"access-control-allow-origin":   "https://foo.bar.com",
						"Access-control-allow-methods": "GET",
						"Access-control-allow-headers": "x-header-1, x-header-2",
					},
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})
	},
}
