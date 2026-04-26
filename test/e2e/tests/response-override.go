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
			verifyCustomResponse(t, nil, &suite.TimeoutConfig, gwAddr, "/status/404", "text/plain", "404 Oops! Your request is not found.", 404, map[string]string{
				"X-Add-Header":  "added-404",
				"X-Set-Header":  "set-404",
				"X-Error-Type":  "not-found",
				"Cache-Control": "no-cache",
			})

			// Test 500 response override with add and set headers
			verifyCustomResponse(t, nil, &suite.TimeoutConfig, gwAddr, "/status/500", "application/json", `{"error": "Internal Server Error"}`, 500, map[string]string{
				"X-Add-Header": "added-500",
				"X-Set-Header": "set-500",
			})

			// Test 403 response override with add and set headers (status override to 404)
			verifyCustomResponse(t, nil, &suite.TimeoutConfig, gwAddr, "/status/403", "", "", 404, map[string]string{
				"X-Add-Header": "added-403",
				"X-Set-Header": "set-403",
			})
			verifyCustomResponse(t, nil, &suite.TimeoutConfig, gwAddr, "/status/401", "", "", 301)

			// Test 418 response override with request header match.
			// Clients sending Accept: application/json get a JSON body; others get HTML.
			verifyCustomResponse(t, map[string]string{"Accept": "application/json"},
				&suite.TimeoutConfig, gwAddr, "/status/418",
				"application/json", `{"error":"I am a teapot"}`, 418)
			verifyCustomResponse(t, map[string]string{"Accept": "text/html"},
				&suite.TimeoutConfig, gwAddr, "/status/418",
				"text/html", "<html><body><h1>I'm a teapot</h1></body></html>", 418)
		})
	},
}

func verifyCustomResponse(t *testing.T, withHeaders map[string]string, timeoutConfig *config.TimeoutConfig, gwAddr,
	path, expectedContentType, expectedBody string, expectedStatusCode int, expectedHeaders ...map[string]string,
) {
	if timeoutConfig == nil {
		t.Fatalf("timeoutConfig cannot be nil")
	}

	reqURL := url.URL{
		Scheme: "http",
		Host:   httputils.CalculateHost(t, gwAddr, "http"),
		Path:   path,
	}

	httputils.AwaitConvergence(t, timeoutConfig.RequiredConsecutiveSuccesses, timeoutConfig.MaxTimeToConsistency, func(_ time.Duration) bool {
		req, err := http.NewRequest(http.MethodGet, reqURL.String(), nil)
		if err != nil {
			tlog.Logf(t, "failed to build request: %v", err)
			return false
		}
		for k, v := range withHeaders {
			req.Header.Set(k, v)
		}

		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			tlog.Logf(t, "failed to get response: %v", err)
			return false
		}

		// Verify that the response body is overridden
		defer rsp.Body.Close()
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			tlog.Logf(t, "failed to read response body: %v", err)
			return false
		}
		if string(body) != expectedBody {
			tlog.Logf(t, "expected response body to be %s but got %s", expectedBody, string(body))
			return false
		}

		// Verify that the content type is overridden
		contentType := rsp.Header.Get("Content-Type")
		if contentType != expectedContentType {
			tlog.Logf(t, "expected content type to be %s but got %s", expectedContentType, contentType)
			return false
		}

		if expectedStatusCode != rsp.StatusCode {
			tlog.Logf(t, "expected status code to be %d but got %d", expectedStatusCode, rsp.StatusCode)
			return false
		}

		// Verify expected headers if provided
		if len(expectedHeaders) > 0 {
			for headerName, expectedValue := range expectedHeaders[0] {
				actualValue := rsp.Header.Get(headerName)
				if actualValue != expectedValue {
					tlog.Logf(t, "expected header %s to be %s but got %s", headerName, expectedValue, actualValue)
					return false
				}
			}
		}

		return true
	})

	tlog.Logf(t, "Request passed")
}
