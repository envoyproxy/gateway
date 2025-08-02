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
	ShortName:   "ResponseOverrideSpecificUser",
	Description: "Response Override",
	Manifests:   []string{"testdata/response-override.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("response override", func(t *testing.T) {
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
			verifyCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/404", "text/plain", "404 Oops! Your request is not found.", 404)
			verifyCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/500", "application/json", `{"error": "Internal Server Error"}`, 500)
			verifyCustomResponse(t, suite.TimeoutConfig, gwAddr, "/status/403", "", "", 404)
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
