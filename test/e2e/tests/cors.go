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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, CORSFromSecurityPolicyTest, CORSFromHTTPCORSFilterTest)
}

var CORSFromSecurityPolicyTest = suite.ConformanceTest{
	ShortName:   "CORSFromSecurityPolicy",
	Description: "Test CORS from SecurityPolicy",
	Manifests:   []string{"testdata/cors-security-policy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		testCORS(t, suite, true)
	},
}

var CORSFromHTTPCORSFilterTest = suite.ConformanceTest{
	ShortName:   "CORSFromHTTPCORSFilter",
	Description: "Test CORS from HTTP CORS Filter",
	Manifests:   []string{"testdata/cors-http-cors-filter.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		testCORS(t, suite, false)
	},
}

func testCORS(t *testing.T, suite *suite.ConformanceTestSuite, isSP bool) {
	t.Run("should enable cors with Allow Origin Exact", func(t *testing.T) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-with-cors-exact", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		if isSP {
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cors-exact", Namespace: ns}, suite.ControllerName, ancestorRef)
		}

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path:   "/cors-exact",
				Method: "OPTIONS",
				Headers: map[string]string{
					"Origin":                         "https://www.foo.com",
					"access-control-request-method":  "GET",
					"access-control-request-headers": "x-header-1, x-header-2",
				},
			},
			// Set the expected request properties to empty strings.
			// This is a workaround to avoid the test failure.
			// The response body is empty because the request is a preflight request.
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Host:    "",
					Method:  "OPTIONS",
					Path:    "",
					Headers: nil,
				},
			},
			Response: http.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"access-control-allow-origin":   "https://www.foo.com",
					"access-control-allow-methods":  "GET, POST, PUT, PATCH, DELETE, OPTIONS",
					"access-control-allow-headers":  "x-header-1, x-header-2",
					"access-control-expose-headers": "x-header-3, x-header-4",
				},
			},
			Namespace: "",
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
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}

		if isSP {
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cors-exact", Namespace: ns}, suite.ControllerName, ancestorRef)
		}

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path:   "/cors-exact",
				Method: "OPTIONS",
				Headers: map[string]string{
					"Origin":                         "https://anydomain.foobar.com",
					"access-control-request-method":  "GET",
					"access-control-request-headers": "x-header-1, x-header-2",
				},
			},
			// Set the expected request properties to empty strings.
			// This is a workaround to avoid the test failure.
			// The response body is empty because the request is a preflight request.
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Host:    "",
					Method:  "OPTIONS",
					Path:    "",
					Headers: nil,
				},
			},
			Response: http.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"access-control-allow-origin":   "https://anydomain.foobar.com",
					"access-control-allow-methods":  "GET, POST, PUT, PATCH, DELETE, OPTIONS",
					"access-control-allow-headers":  "x-header-1, x-header-2",
					"access-control-expose-headers": "x-header-3, x-header-4",
				},
			},
			Namespace: "",
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
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}

		if isSP {
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cors-exact", Namespace: ns}, suite.ControllerName, ancestorRef)
		}

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path:   "/cors-exact",
				Method: "OPTIONS",
				Headers: map[string]string{
					"Origin":                         "https://unknown.foo.com",
					"access-control-request-method":  "GET",
					"access-control-request-headers": "x-header-1, x-header-2",
				},
			},
			// Set the expected request properties to empty strings.
			// This is a workaround to avoid the test failure.
			// The response body is empty because the request is a preflight request.
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Host:    "",
					Method:  "OPTIONS",
					Path:    "",
					Headers: nil,
				},
			},
			Response: http.Response{
				AbsentHeaders: []string{"access-control-allow-origin"},
			},
			Namespace: "",
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
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}

		if isSP {
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "cors-wildcard", Namespace: ns}, suite.ControllerName, ancestorRef)
		}

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path:   "/cors-wildcard",
				Method: "OPTIONS",
				Headers: map[string]string{
					"Origin":                         "https://foo.bar.com",
					"access-control-request-method":  "GET",
					"access-control-request-headers": "x-header-1, x-header-2",
				},
			},
			// Set the expected request properties to empty strings.
			// This is a workaround to avoid the test failure.
			// The response body is empty because the request is a preflight request.
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Host:    "",
					Method:  "OPTIONS",
					Path:    "",
					Headers: nil,
				},
			},
			Response: http.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"access-control-allow-origin":   "https://foo.bar.com",
					"access-control-allow-methods":  "GET",
					"access-control-allow-headers":  "x-header-1, x-header-2",
					"access-control-expose-headers": "*",
				},
			},
			Namespace: "",
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
}
