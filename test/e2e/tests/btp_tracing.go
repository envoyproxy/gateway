// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/utils/naming"
	"github.com/envoyproxy/gateway/test/utils/tracing"
)

func init() {
	ConformanceTests = append(ConformanceTests, BTPTracingTest)
}

var BTPTracingTest = suite.ConformanceTest{
	ShortName:   "BTPTracing",
	Description: "Test BackendTrafficPolicy tracing on HTTPRoute",
	Manifests:   []string{"testdata/btp-tracing.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "tracing-otel", Namespace: ns}

		// For non-override, the tracing provider this's simply same as OpenTelemetryTracingTest
		t.Run("NonOverride", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "no-override", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/otel",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			tags := map[string]string{
				"component":    "proxy",
				"provider":     "otel",
				"service.name": naming.ServiceName(gwNN),
			}
			tracing.ExpectedTraceCount(t, suite, gwAddr, expectedResponse, tags)
		})

		t.Run("Override", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "override", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/otel-override",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			tags := map[string]string{
				"component":    "proxy",
				"provider":     "otel-override",
				"service.name": naming.ServiceName(gwNN),
			}
			tracing.ExpectedTraceCount(t, suite, gwAddr, expectedResponse, tags)
		})
	},
}
