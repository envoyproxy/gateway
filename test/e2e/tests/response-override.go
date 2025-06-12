// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
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
	Description: "Response Override with Backend Traffic Policy and HTTPRoute Filters",
	Manifests:   []string{"testdata/response-override.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
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

		// Test backend traffic policy with dynamic variables
		t.Run("backend traffic policy with dynamic variables", func(t *testing.T) {
			// Test JSON response with variables
			verifyBackendTrafficPolicyResponse(t, suite.TimeoutConfig, gwAddr, "/backend/404", "user-404", 404)
			// Test inline response with variables
			verifyBackendTrafficPolicyResponse(t, suite.TimeoutConfig, gwAddr, "/backend/500", "user-500", 500)
			// Test valueref response with variables
			verifyBackendTrafficPolicyResponse(t, suite.TimeoutConfig, gwAddr, "/backend/503", "user-503", 503)
		})

		// Test HTTPRoute filters with direct responses
		t.Run("HTTPRoute filter responses", func(t *testing.T) {
			verifyDirectResponse(t, suite.TimeoutConfig, gwAddr, "/filter/json", "application/json",
				`{"response_type":"JSON","message":"Direct response from HTTPRoute filter","source":"filter"}`, 503)

			verifyDirectResponse(t, suite.TimeoutConfig, gwAddr, "/filter/inline", "text/html",
				"Inline response from filter - Service temporarily unavailable", 503)

			verifyDirectResponse(t, suite.TimeoutConfig, gwAddr, "/filter/valueref", "text/html",
				`{"message":"ValueRef response from ConfigMap","source":"configmap"}`, 503)
		})
	},
}

func verifyBackendTrafficPolicyResponse(t *testing.T, timeoutConfig config.TimeoutConfig, gwAddr,
	path, userID string, expectedStatusCode int,
) {
	reqURL := url.URL{
		Scheme: "http",
		Host:   httputils.CalculateHost(t, gwAddr, "http"),
		Path:   path,
	}

	httputils.AwaitConvergence(t, timeoutConfig.RequiredConsecutiveSuccesses, timeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
		req, err := http.NewRequest("GET", reqURL.String(), nil)
		if err != nil {
			tlog.Logf(t, "failed to create request: %v", err)
			return false
		}
		req.Header.Set("X-User-ID", userID)

		client := &http.Client{}
		rsp, err := client.Do(req)
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

		// Verify that the custom response header is added
		customHeader := rsp.Header.Get("Response-Override-Test")
		if customHeader != "True" {
			tlog.Logf(t, "expected Response-Override-Test header to be 'True' but got '%s'", customHeader)
			return false
		}

		// Verify response body contains expected values based on status code
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			tlog.Logf(t, "failed to read response body: %v", err)
			return false
		}

		bodyStr := string(body)
		var expectedBody string

		switch expectedStatusCode {
		case 404:
			// Verify JSON response for 404
			if !contains(bodyStr, fmt.Sprintf(`"user_id":"%s"`, userID)) ||
				!contains(bodyStr, `"error":"not_found"`) {
				tlog.Logf(t, "expected response body to contain user_id and error fields but got: %s", bodyStr)
				return false
			}
		case 500:
			// Verify inline response for 500
			expectedBody = fmt.Sprintf("Error for user %s: Internal Server Error", userID)
			if bodyStr != expectedBody {
				tlog.Logf(t, "expected response body to be '%s' but got '%s'", expectedBody, bodyStr)
				return false
			}
		case 503:
			// Verify valueref response for 503
			expectedBody = fmt.Sprintf("ConfigMap response with user: %s", userID)
			if bodyStr != expectedBody {
				tlog.Logf(t, "expected response body to be '%s' but got '%s'", expectedBody, bodyStr)
				return false
			}
		}

		return true
	})

	tlog.Logf(t, "Backend traffic policy response test passed for %s", path)
}

func verifyDirectResponse(t *testing.T, timeoutConfig config.TimeoutConfig, gwAddr,
	path, expectedContentType string, expectedBody string, expectedStatusCode int,
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

		// Verify content type
		contentType := rsp.Header.Get("Content-Type")
		if contentType != expectedContentType {
			tlog.Logf(t, "expected content type to be %s but got %s", expectedContentType, contentType)
			return false
		}

		// Verify response body
		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			tlog.Logf(t, "failed to read response body: %v", err)
			return false
		}

		bodyStr := string(body)
		if bodyStr != expectedBody {
			tlog.Logf(t, "expected response body to be '%s' but got '%s'", expectedBody, bodyStr)
			return false
		}

		return true
	})

	tlog.Logf(t, "Direct response test passed for %s", path)
}

func contains(text, substring string) bool {
	return len(text) >= len(substring) &&
		(text == substring ||
			contains(text[1:], substring) ||
			(len(text) > 0 && text[:len(substring)] == substring))
}
