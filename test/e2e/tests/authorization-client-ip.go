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
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests, AuthorizationClientIPTest)
}

var AuthorizationClientIPTest = suite.ConformanceTest{
	ShortName:   "Authorization with client IP",
	Description: "Authorization with client IP Allow/Deny list",
	Manifests:   []string{"testdata/authorization-client-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		route1NN := types.NamespacedName{Name: "http-with-authorization-client-ip-1", Namespace: ns}
		route2NN := types.NamespacedName{Name: "http-with-authorization-client-ip-2", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), route1NN, route2NN)

		ancestorRef := gwv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwv1.GroupName),
			Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwv1.ObjectName(gwNN.Name),
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
					StatusCode: 403,
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
					StatusCode: 200,
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
					StatusCode: 200,
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
					StatusCode: 200,
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

		t.Run("second route-default action: deny", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected2",
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.3.1", // not in the allowed list
					},
				},
				Response: http.Response{
					StatusCode: 403,
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
