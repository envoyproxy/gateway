// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/utils/naming"
	"github.com/envoyproxy/gateway/test/utils/tracing"
)

func init() {
	ConformanceTests = append(ConformanceTests, OpenTelemetryTracingTest, ZipkinTracingTest, DatadogTracingTest)
}

var OpenTelemetryTracingTest = suite.ConformanceTest{
	ShortName:   "OpenTelemetryTracing",
	Description: "Make sure OpenTelemetry tracing is working (default and custom service name)",
	Manifests:   []string{"testdata/tracing-otel.yaml", "testdata/tracing-otel-custom-service-name.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		cases := []struct {
			name        string
			routeName   string
			gwName      string
			path        string
			expectedSvc string
		}{
			{
				name:        "default-service-name",
				routeName:   "tracing-otel",
				gwName:      "tracing-otel",
				path:        "/otel",
				expectedSvc: naming.ServiceName(types.NamespacedName{Name: "tracing-otel", Namespace: "gateway-conformance-infra"}),
			},
			{
				name:        "custom-service-name",
				routeName:   "tracing-otel-custom-service-name",
				gwName:      "tracing-otel-custom-service-name",
				path:        "/custom-service",
				expectedSvc: "my-custom-service",
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				ns := "gateway-conformance-infra"
				routeNN := types.NamespacedName{Name: tc.routeName, Namespace: ns}
				gwNN := types.NamespacedName{Name: tc.gwName, Namespace: ns}
				gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
				expectedResponse := httputils.ExpectedResponse{
					Request: httputils.Request{
						Path: tc.path,
					},
					Response: httputils.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				}
				httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
				tags := map[string]string{
					"component":    "proxy",
					"provider":     "otel",
					"service.name": tc.expectedSvc,
				}
				tracing.ExpectedTraceCount(t, suite, gwAddr, &expectedResponse, tags)
			})
		}
	},
}

var ZipkinTracingTest = suite.ConformanceTest{
	ShortName:   "ZipkinTracing",
	Description: "Make sure Zipkin tracing is working",
	Manifests:   []string{"testdata/tracing-zipkin.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("tempo", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "tracing-zipkin", Namespace: ns}
			gwNN := types.NamespacedName{Name: "tracing-zipkin", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/zipkin",
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			tags := map[string]string{
				"component": "proxy",
				"provider":  "zipkin",
				// TODO: this came from --service-cluster, which is different from OTel,
				// should make them kept consistent
				"service.name": fmt.Sprintf("%s/%s", gwNN.Namespace, gwNN.Name),
			}
			tracing.ExpectedTraceCount(t, suite, gwAddr, &expectedResponse, tags)
		})
	},
}

var DatadogTracingTest = suite.ConformanceTest{
	ShortName:   "DatadogTracing",
	Description: "Make sure Datadog tracing is working (default and custom service name)",
	Manifests:   []string{"testdata/tracing-datadog.yaml", "testdata/tracing-datadog-custom-service-name.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		cases := []struct {
			name        string
			routeName   string
			gwName      string
			path        string
			expectedSvc string
		}{
			{
				name:        "default-service-name",
				routeName:   "tracing-datadog",
				gwName:      "eg-special-case-datadog",
				path:        "/datadog",
				expectedSvc: fmt.Sprintf("%s.%s", "eg-special-case-datadog", "gateway-conformance-infra"),
			},
			{
				name:        "custom-service-name",
				routeName:   "tracing-datadog-custom-service-name",
				gwName:      "tracing-datadog-custom-service-name",
				path:        "/datadog-custom-service",
				expectedSvc: "my-custom-service",
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				ns := "gateway-conformance-infra"
				routeNN := types.NamespacedName{Name: tc.routeName, Namespace: ns}
				gwNN := types.NamespacedName{Name: tc.gwName, Namespace: ns}
				gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
				expectedResponse := httputils.ExpectedResponse{
					Request: httputils.Request{
						Path: tc.path,
					},
					Response: httputils.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				}
				httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
				tags := map[string]string{
					"component":    "proxy",
					"provider":     "datadog",
					"service.name": tc.expectedSvc,
				}
				tracing.ExpectedTraceCount(t, suite, gwAddr, &expectedResponse, tags)
			})
		}
	},
}
