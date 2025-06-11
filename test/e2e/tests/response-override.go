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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
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
	ShortName:   "ResponseOverride",
	Description: "Response Override with Basic and Enhanced Features",
	Manifests:   []string{"testdata/response-override.yaml", "testdata/response-override-enhanced.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("basic response override", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "response-override", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "response-override", Namespace: ns}, suite.ControllerName, ancestorRef)
			verifyCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/404", "text/plain", "Oops! Your request is not found.", 404)
			verifyCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/500", "application/json", `{"error": "Internal Server Error"}`, 500)
			verifyCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/403", "", "", 404)
		})

		t.Run("enhanced response override", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "response-override-enhanced", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "response-override-enhanced", Namespace: ns}, suite.ControllerName, ancestorRef)

			// Test 404 response with JSON body format and custom headers
			t.Run("404 with JSON format and headers", func(t *testing.T) {
				verifyEnhancedCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/404",
					map[string]string{
						"Content-Type":     "application/json",
						"X-Custom-Error":   "Not Found",
						"X-Gateway-Source": "Envoy Gateway",
					},
					func(body string) bool {
						return containsAll(body, []string{`"status":"404"`, `"message":`, `"timestamp":`})
					},
					404)
			})

			// Test 500 response with text body format and custom headers
			t.Run("500 with text format and headers", func(t *testing.T) {
				verifyEnhancedCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/500",
					map[string]string{
						"Content-Type": "text/plain",
						"X-Error-Type": "Internal",
					},
					func(body string) bool {
						return containsAll(body, []string{"Error 500:", " at "})
					},
					500)
			})

			// Test 429 rate limit response with comprehensive headers and JSON format
			t.Run("429 with rate limit headers and JSON format", func(t *testing.T) {
				verifyEnhancedCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/429",
					map[string]string{
						"Content-Type":          "application/json; charset=utf-8",
						"X-RateLimit-Limit":     "100",
						"X-RateLimit-Remaining": "0",
						"Retry-After":           "60",
					},
					func(body string) bool {
						return containsAll(body, []string{
							`"error":"rate_limit_exceeded"`,
							`"message":"Too many requests"`,
							`"status_code":"429"`,
							`"limit":"100 requests per minute"`,
							`"reset_time":`,
						})
					},
					429)
			})

			// Test 503 with ConfigMap body and body format override
			t.Run("503 with ConfigMap body and format override", func(t *testing.T) {
				verifyEnhancedCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/503",
					map[string]string{
						"Content-Type":      "application/json",
						"X-Custom-Response": "true",
						"X-Service-Version": "v1.2.3",
					},
					func(body string) bool {
						return containsAll(body, []string{
							`"original_response":`,
							`"enhanced_status":"503"`,
							`"server_info":"Gateway Enhanced Response"`,
						})
					},
					503)
			})

			// Test header append functionality
			t.Run("header append functionality", func(t *testing.T) {
				verifyHeaderAppendBehavior(t, suite.TimeoutConfig, gwAddr, "/status/502")
			})
		})
	},
}

func verifyCustomResponse(t *testing.T, timeoutConfig config.TimeoutConfig, gwAddr,
	path, expectedContentType, expectedBody string, expectedStatusCode int,
) {
	reqURL := url.URL{
		Scheme: "http",
		Host:   httputils.CalculateHost(t, gwAddr, "http"),
		Path:   path,
	}

	httputils.AwaitConvergence(t, timeoutConfig.RequiredConsecutiveSuccesses, timeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
		rsp, err := http.Get(reqURL.String())
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

		return true
	})

	tlog.Logf(t, "Request passed")
}

func verifyEnhancedCustomResponse(t *testing.T, timeoutConfig config.TimeoutConfig, gwAddr,
	path string, expectedHeaders map[string]string, bodyValidator func(string) bool, expectedStatusCode int,
) {
	reqURL := url.URL{
		Scheme: "http",
		Host:   httputils.CalculateHost(t, gwAddr, "http"),
		Path:   path,
	}

	httputils.AwaitConvergence(t, timeoutConfig.RequiredConsecutiveSuccesses, timeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
		rsp, err := http.Get(reqURL.String())
		if err != nil {
			tlog.Logf(t, "failed to get response: %v", err)
			return false
		}
		defer rsp.Body.Close()

		// Verify status code
		if expectedStatusCode != rsp.StatusCode {
			tlog.Logf(t, "expected status code to be %d but got %d", expectedStatusCode, rsp.StatusCode)
			return false
		}

		// Verify headers
		for expectedHeader, expectedValue := range expectedHeaders {
			actualValue := rsp.Header.Get(expectedHeader)
			if actualValue != expectedValue {
				tlog.Logf(t, "expected header %s to be %s but got %s", expectedHeader, expectedValue, actualValue)
				return false
			}
		}

		// Verify response body
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			tlog.Logf(t, "failed to read response body: %v", err)
			return false
		}

		if !bodyValidator(string(body)) {
			tlog.Logf(t, "body validation failed for: %s", string(body))
			return false
		}

		return true
	})

	tlog.Logf(t, "Enhanced custom response test passed for %s", path)
}

func verifyHeaderAppendBehavior(t *testing.T, timeoutConfig config.TimeoutConfig, gwAddr, path string) {
	reqURL := url.URL{
		Scheme: "http",
		Host:   httputils.CalculateHost(t, gwAddr, "http"),
		Path:   path,
	}

	// Add a request with existing Cache-Control header to test append behavior
	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Cache-Control", "max-age=3600")

	httputils.AwaitConvergence(t, timeoutConfig.RequiredConsecutiveSuccesses, timeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
		client := &http.Client{}
		rsp, err := client.Do(req)
		if err != nil {
			tlog.Logf(t, "failed to get response: %v", err)
			return false
		}
		defer rsp.Body.Close()

		// Verify that Cache-Control header was appended (not overwritten)
		cacheControlValues := rsp.Header.Values("Cache-Control")
		if len(cacheControlValues) < 2 {
			tlog.Logf(t, "expected multiple Cache-Control headers due to append behavior, got: %v", cacheControlValues)
			return false
		}

		// Verify that X-Error-Source header was overwritten (not appended)
		errorSourceValues := rsp.Header.Values("X-Error-Source")
		if len(errorSourceValues) != 1 || errorSourceValues[0] != "backend-service" {
			tlog.Logf(t, "expected single X-Error-Source header with value 'backend-service', got: %v", errorSourceValues)
			return false
		}

		return true
	})

	tlog.Logf(t, "Header append behavior test passed")
}

func containsAll(text string, substrings []string) bool {
	for _, substring := range substrings {
		if !contains(text, substring) {
			return false
		}
	}
	return true
}

func contains(text, substring string) bool {
	return len(text) >= len(substring) &&
		(text == substring ||
			contains(text[1:], substring) ||
			(len(text) > 0 && text[:len(substring)] == substring))
}
