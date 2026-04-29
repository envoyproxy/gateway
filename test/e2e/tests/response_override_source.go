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
	ConformanceTests = append(ConformanceTests, ResponseOverrideSourceTest)
}

var ResponseOverrideSourceTest = suite.ConformanceTest{
	ShortName:   "ResponseOverrideSource",
	Description: "Response Override with source: Local and source: Backend",
	Manifests:   []string{"testdata/response-override-source.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		localRouteNN := types.NamespacedName{Name: "response-override-source-local", Namespace: ns}
		backendRouteNN := types.NamespacedName{Name: "response-override-source-backend", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), localRouteNN, backendRouteNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "response-override-source-local", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "response-override-source-backend", Namespace: ns}, suite.ControllerName, ancestorRef)
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "response-override-source-local", Namespace: ns}, suite.ControllerName, ancestorRef)
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "response-override-source-backend", Namespace: ns}, suite.ControllerName, ancestorRef)

		// source: Local — should fire for Envoy-generated 401s but not upstream 401s.
		t.Run("source Local fires on Envoy-generated 401", func(t *testing.T) {
			// No JWT token → Envoy rejects the request with a locally-generated 401.
			// The source: Local rule should override it.
			checkResponseOverrideSource(t, &suite.TimeoutConfig, gwAddr, "/local/status/200", "", `{"error": "Unauthorized"}`, "")
		})

		t.Run("source Local does not fire on upstream 401", func(t *testing.T) {
			// Valid JWT + backend returns 401 → upstream response, not Envoy-generated.
			// The source: Local rule should not match, so the raw upstream 401 passes through.
			// No exact passthrough body: echo-basic returns a request-echo JSON we don't control.
			checkResponseOverrideSource(t, &suite.TimeoutConfig, gwAddr, "/local/status/401", jwtToken, "", "")
		})

		// source: Backend — should fire for upstream 401s but not Envoy-generated 401s.
		t.Run("source Backend fires on upstream 401", func(t *testing.T) {
			// Valid JWT + backend returns 401 → upstream response.
			// The source: Backend rule should override it.
			checkResponseOverrideSource(t, &suite.TimeoutConfig, gwAddr, "/backend/status/401", jwtToken, `{"error": "Upstream Error"}`, "")
		})

		t.Run("source Backend does not fire on Envoy-generated 401", func(t *testing.T) {
			// No JWT token → Envoy rejects the request with a locally-generated 401.
			// The source: Backend rule should not match, so Envoy's default JWT error passes through.
			checkResponseOverrideSource(t, &suite.TimeoutConfig, gwAddr, "/backend/status/200", "", "", "Jwt is missing")
		})
	},
}

// checkResponseOverrideSource makes a GET request with an optional JWT Bearer token, verifies the
// status code, and — when expectedBody is non-empty — verifies the response body matches exactly.
// An empty expectedBody means the response must NOT have been overridden by a custom-response rule
// (i.e. the body should not equal either of the two custom bodies used in this test).
func checkResponseOverrideSource(t *testing.T, timeoutConfig *config.TimeoutConfig, gwAddr, path, token, expectedBody, passthroughBody string) {
	t.Helper()

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
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		rsp, err := http.DefaultClient.Do(req)
		if err != nil {
			tlog.Logf(t, "failed to get response: %v", err)
			return false
		}
		defer rsp.Body.Close()

		if rsp.StatusCode != http.StatusUnauthorized {
			tlog.Logf(t, "expected status 401 but got %d", rsp.StatusCode)
			return false
		}

		body, err := io.ReadAll(rsp.Body)
		if err != nil {
			tlog.Logf(t, "failed to read body: %v", err)
			return false
		}

		if expectedBody != "" {
			if string(body) != expectedBody {
				tlog.Logf(t, "expected body %q but got %q", expectedBody, string(body))
				return false
			}
		} else {
			// No override expected: verify neither custom body was applied.
			if string(body) == `{"error": "Unauthorized"}` || string(body) == `{"error": "Upstream Error"}` {
				tlog.Logf(t, "response was unexpectedly overridden, body: %q", string(body))
				return false
			}
			// When we know the exact passthrough body, assert it too.
			if passthroughBody != "" && string(body) != passthroughBody {
				tlog.Logf(t, "expected passthrough body %q but got %q", passthroughBody, string(body))
				return false
			}
		}

		return true
	})

	tlog.Logf(t, "Request passed")
}
