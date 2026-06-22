// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
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
	ConformanceTests = append(ConformanceTests, RequestBodyBufferLimitTest)
}

var RequestBodyBufferLimitTest = suite.ConformanceTest{
	ShortName:   "RequestBodyBufferLimit",
	Description: "Verify that BackendTrafficPolicy requestBodyBufferLimit bounds the request body size Envoy will buffer",
	Manifests:   []string{"testdata/request-body-buffer-limit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		defaultRouteNN := types.NamespacedName{Name: "request-body-buffer-limit-default", Namespace: ns}
		highRouteNN := types.NamespacedName{Name: "request-body-buffer-limit-high", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
			t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), defaultRouteNN, highRouteNN,
		)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		// The inline Lua filter forces the request body to be buffered on both routes.
		EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
			types.NamespacedName{Name: "request-body-buffer-limit-buffer-body", Namespace: ns},
			suite.ControllerName, ancestorRef)
		// The requestBodyBufferLimit policy raises the buffer limit on the "high" route.
		BackendTrafficPolicyMustBeAccepted(t, suite.Client,
			types.NamespacedName{Name: "request-body-buffer-limit-high", Namespace: ns},
			suite.ControllerName, ancestorRef)

		// 100 KiB body: larger than the default per-connection buffer limit (32768 bytes)
		// but smaller than the 1Mi requestBodyBufferLimit configured on the "high" route.
		body := strings.Repeat("x", 100*1024)

		t.Run("request exceeding the default buffer limit is rejected with 413", func(t *testing.T) {
			expectRequestBodyStatus(t, suite, gwAddr, "/request-body-buffer-limit-default", body, http.StatusRequestEntityTooLarge)
		})

		t.Run("request within the configured buffer limit succeeds", func(t *testing.T) {
			expectRequestBodyStatus(t, suite, gwAddr, "/request-body-buffer-limit-high", body, http.StatusOK)
		})
	},
}

// expectRequestBodyStatus POSTs the given body to gwAddr+path and waits until the
// response status code converges to expectedStatusCode.
func expectRequestBodyStatus(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr, path, body string, expectedStatusCode int) {
	t.Helper()

	url := fmt.Sprintf("http://%s%s", gwAddr, path)
	httpClient := &http.Client{Timeout: 30 * time.Second}

	httputils.AwaitConvergence(t, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency,
		func(elapsed time.Duration) bool {
			resp, err := httpClient.Post(url, "application/octet-stream", strings.NewReader(body))
			if err != nil {
				tlog.Logf(t, "POST %s error after %v: %v", url, elapsed, err)
				return false
			}
			defer resp.Body.Close()

			if resp.StatusCode == expectedStatusCode {
				return true
			}

			tlog.Logf(t, "POST %s got status %d, want %d after %v", url, resp.StatusCode, expectedStatusCode, elapsed)
			return false
		})
}
