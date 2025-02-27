// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, MetricTest, MetricCompressorTest)
}

var MetricTest = suite.ConformanceTest{
	ShortName:   "ProxyMetrics",
	Description: "Make sure metric is working",
	Manifests:   []string{"testdata/metric.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "metric-prometheus", Namespace: ns}
		gwNN := types.NamespacedName{Name: "metric-prometheus", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("prometheus", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/prom",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// let's check the metric
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					if err := ScrapeMetrics(t, suite.Client, types.NamespacedName{
						Namespace: "envoy-gateway-system",
						Name:      "same-namespace-gw-metrics",
					}, 19001, "/stats/prometheus"); err != nil {
						tlog.Logf(t, "failed to get metric: %v", err)
						return false, nil
					}
					return true, nil
				}); err != nil {
				t.Errorf("failed to scrape metrics: %v", err)
			}
		})

		t.Run("otel", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/prom",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// let's check the metric
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					if err := ScrapeMetrics(t, suite.Client, types.NamespacedName{
						Namespace: "monitoring",
						Name:      "otel-collecot-prometheus",
					}, 19001, "/metrics"); err != nil {
						tlog.Logf(t, "failed to get metric: %v", err)
						return false, nil
					}
					return true, nil
				}); err != nil {
				t.Errorf("failed to scrape metrics: %v", err)
			}
		})
	},
}

var MetricCompressorTest = suite.ConformanceTest{
	ShortName:   "MetricCompressor",
	Description: "Make sure metric is working with compressor",
	Manifests:   []string{"testdata/metric-compressor.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		t.Run("gzip", func(t *testing.T) {
			runMetricCompressorTest(t, suite, ns, "gzip-route", "gzip-gtw", "/gzip")
		})
		t.Run("brotli", func(t *testing.T) {
			runMetricCompressorTest(t, suite, ns, "brotli-route", "brotli-gtw", "/brotli")
		})
	},
}

func runMetricCompressorTest(t *testing.T, suite *suite.ConformanceTestSuite, ns string, routeName, gtwName string, checkPath string) {
	routeNN := types.NamespacedName{Name: routeName, Namespace: ns}
	gwNN := types.NamespacedName{Name: gtwName, Namespace: ns}
	gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

	expectedResponse := httputils.ExpectedResponse{
		Request: httputils.Request{
			Path: checkPath,
		},
		Response: httputils.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	// make sure listener is ready
	httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
}
