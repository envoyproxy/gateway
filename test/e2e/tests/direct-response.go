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

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, DirectResponseTest)
}

var DirectResponseTest = suite.ConformanceTest{
	ShortName:   "DirectResponse",
	Description: "Direct",
	Manifests:   []string{"testdata/direct-response.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("direct response", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "direct-response", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
			verifyCustomResponse(t, gwAddr, "/status/404", "text/plain", "Oops! Your request is not found.")
			verifyCustomResponse(t, gwAddr, "/status/500", "application/json", `{"error": "Internal Server Error"}`)
		})
	},
}

func verifyCustomResponse(t *testing.T, gwAddr, path, expectedContentType, expectedBody string) {
	reqURL := url.URL{
		Scheme: "http",
		Host:   httputils.CalculateHost(t, gwAddr, "http"),
		Path:   path,
	}

	rsp, err := http.Get(reqURL.String())
	if err != nil {
		t.Fatalf("failed to get response: %v", err)
	}

	// Verify that the response body is overridden
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != expectedBody {
		t.Errorf("expected response body to be %s but got %s", expectedBody, string(body))
	}

	// Verify that the content type is overridden
	contentType := rsp.Header.Get("Content-Type")
	if contentType != expectedContentType {
		t.Errorf("expected content type to be %s but got %s", expectedContentType, contentType)
	}
}
