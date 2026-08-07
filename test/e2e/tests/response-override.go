// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, ResponseOverrideTest)
}

var ResponseOverrideTest = suite.ConformanceTest{
	ShortName:   "ResponseOverrideSpecificUser",
	Description: "Response Override",
	Manifests:   []string{"testdata/response-override.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("response override", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "response-override", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "response-override", Namespace: ns}, suite.ControllerName, ancestorRef)

			// Test 404 response override with add and set headers
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, &expectedResponse{
				path:        "/status/404",
				contentType: "text/plain",
				body:        "404 Oops! Your request is not found.",
				statusCode:  404,
				headers: map[string]string{
					"X-Add-Header":  "added-404",
					"X-Set-Header":  "set-404",
					"X-Error-Type":  "not-found",
					"Cache-Control": "no-cache",
				},
			})

			// Test 500 response override with add and set headers
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, &expectedResponse{
				path:        "/status/500",
				contentType: "application/json",
				body:        `{"error": "Internal Server Error"}`,
				statusCode:  500,
				headers: map[string]string{
					"X-Add-Header": "added-500",
					"X-Set-Header": "set-500",
				},
			})

			// Test 403 response override with add and set headers (status override to 404)
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, &expectedResponse{
				path:       "/status/403",
				statusCode: 404,
				headers: map[string]string{
					"X-Add-Header": "added-403",
					"X-Set-Header": "set-403",
				},
			})

			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, &expectedResponse{
				path:       "/status/401",
				statusCode: 301,
			})

			// Test header match response override and add X-Override-Matched header (body is also overridden)
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, &expectedResponse{
				path:           "/response-override-header-match",
				requestHeaders: map[string]string{"X-Echo-Set-Header": "X-Custom-Header: custom-value"},
				contentType:    "text/plain",
				body:           "matched on response header",
				statusCode:     200,
				headers:        map[string]string{"X-Override-Matched": "true"},
			})

			// Test header match response override NOT doing anything because the header does not match
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, &expectedResponse{
				path:           "/response-override-header-match",
				requestHeaders: map[string]string{"X-Echo-Set-Header": "X-Custom-Header: other-value"},
				contentType:    "application/json",
				body:           "matched on response header",
				bodyNotEqual:   true,
				statusCode:     200,
			})

			// Test header match response override NOT doing anything because the header is never set
			verifyCustomResponse(t, &suite.TimeoutConfig, gwAddr, &expectedResponse{
				path:         "/response-override-header-match",
				contentType:  "application/json",
				body:         "matched on response header",
				bodyNotEqual: true,
				statusCode:   200,
			})
		})
	},
}

type expectedResponse struct {
	path           string
	requestHeaders map[string]string
	contentType    string
	body           string
	bodyNotEqual   bool
	statusCode     int
	headers        map[string]string
}

func verifyCustomResponse(t *testing.T, timeoutConfig *config.TimeoutConfig, gwAddr string, expected *expectedResponse) {
	if timeoutConfig == nil {
		t.Fatalf("timeoutConfig cannot be nil")
	}

	reqURL := url.URL{
		Scheme: "http",
		Host:   httputils.CalculateHost(t, gwAddr, "http"),
		Path:   expected.path,
	}

	httputils.AwaitConvergence(t, timeoutConfig.RequiredConsecutiveSuccesses, timeoutConfig.MaxTimeToConsistency, func(_ time.Duration) bool {
		req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
		if err != nil {
			tlog.Logf(t, "failed to create request: %v", err)
			return false
		}
		for k, v := range expected.requestHeaders {
			req.Header.Set(k, v)
		}

		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			tlog.Logf(t, "failed to get response: %v", err)
			return false
		}

		if expected.statusCode != rsp.StatusCode {
			tlog.Logf(t, "expected status code to be %d but got %d", expected.statusCode, rsp.StatusCode)
			return false
		}

		// Verify that the content type is overridden
		if contentType := rsp.Header.Get("Content-Type"); contentType != expected.contentType {
			tlog.Logf(t, "expected content type to be %s but got %s", expected.contentType, contentType)
			return false
		}

		// Verify that the response body is (or, when bodyNotEqual is set, is not) overridden
		defer rsp.Body.Close()
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			tlog.Logf(t, "failed to read response body: %v", err)
			return false
		}

		if expected.bodyNotEqual {
			if string(body) == expected.body {
				tlog.Logf(t, "expected response body NOT to be %s", expected.body)
				return false
			}
		} else if string(body) != expected.body {
			tlog.Logf(t, "expected response body to be %s but got %s", expected.body, string(body))
			return false
		}

		// Verify expected headers if provided
		for headerName, expectedValue := range expected.headers {
			if actualValue := rsp.Header.Get(headerName); actualValue != expectedValue {
				tlog.Logf(t, "expected header %s to be %s but got %s", headerName, expectedValue, actualValue)
				return false
			}
		}

		return true
	})

	tlog.Logf(t, "Request passed")
}
