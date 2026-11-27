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
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, RequestBodyLimitTest)
}

var RequestBodyLimitTest = suite.ConformanceTest{
	ShortName:   "RequestBodyLimit",
	Description: "Verify that BackendTrafficPolicy requestBodyLimit rejects request bodies larger than maxSize with HTTP 413 without buffering them",
	Manifests:   []string{"testdata/request-body-limit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		limitedRouteNN := types.NamespacedName{Name: "request-body-limit", Namespace: ns}
		defaultRouteNN := types.NamespacedName{Name: "request-body-limit-default", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
			t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), limitedRouteNN, defaultRouteNN,
		)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client,
			types.NamespacedName{Name: "request-body-limit", Namespace: ns},
			suite.ControllerName, ancestorRef)

		// 1 KiB is below the 64Ki maxSize of the policy, 100 KiB is above it. 100 KiB is also larger
		// than the default per-connection buffer limit, which proves that the route without a limit
		// streams the body instead of rejecting it.
		small := strings.Repeat("x", 1024)
		large := strings.Repeat("x", 100*1024)

		t.Run("body within maxSize succeeds", func(t *testing.T) {
			expectRequestBodyStatus(t, suite, gwAddr, "/request-body-limit", small, http.StatusOK)
		})

		t.Run("body above maxSize announced by Content-Length is rejected with 413", func(t *testing.T) {
			expectRequestBodyStatus(t, suite, gwAddr, "/request-body-limit", large, http.StatusRequestEntityTooLarge)
		})

		t.Run("streamed body above maxSize is rejected with 413", func(t *testing.T) {
			expectChunkedRequestBodyStatus(t, suite, gwAddr, "/request-body-limit", large, http.StatusRequestEntityTooLarge)
		})

		t.Run("route without requestBodyLimit accepts the same body", func(t *testing.T) {
			expectRequestBodyStatus(t, suite, gwAddr, "/request-body-limit-default", large, http.StatusOK)
		})
	},
}

// expectChunkedRequestBodyStatus POSTs the given body to gwAddr+path with chunked transfer encoding,
// so that no Content-Length header is sent and the body size is only known once it has been streamed,
// and waits until the response status code converges to expectedStatusCode.
func expectChunkedRequestBodyStatus(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr, path, body string, expectedStatusCode int) {
	t.Helper()

	url := fmt.Sprintf("http://%s%s", gwAddr, path)
	httpClient := &http.Client{Timeout: 30 * time.Second}

	httputils.AwaitConvergence(t, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency,
		func(elapsed time.Duration) bool {
			// net/http only derives a Content-Length from a few well-known reader types; wrapping the
			// reader hides its type, so the request is sent with Transfer-Encoding: chunked.
			req, err := http.NewRequest(http.MethodPost, url, io.NopCloser(strings.NewReader(body)))
			if err != nil {
				tlog.Logf(t, "building POST %s failed after %v: %v", url, elapsed, err)
				return false
			}
			req.Header.Set("Content-Type", "application/octet-stream")

			resp, err := httpClient.Do(req)
			if err != nil {
				tlog.Logf(t, "chunked POST %s error after %v: %v", url, elapsed, err)
				return false
			}
			defer resp.Body.Close()

			if resp.StatusCode == expectedStatusCode {
				return true
			}

			tlog.Logf(t, "chunked POST %s got status %d, want %d after %v", url, resp.StatusCode, expectedStatusCode, elapsed)
			return false
		})
}
