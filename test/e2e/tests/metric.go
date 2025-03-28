// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
			runMetricCompressorTest(t, suite, ns, egv1a1.GzipCompressorType)
		})
		t.Run("brotli", func(t *testing.T) {
			runMetricCompressorTest(t, suite, ns, egv1a1.BrotliCompressorType)
		})
	},
}

func runMetricCompressorTest(t *testing.T, suite *suite.ConformanceTestSuite, ns string, compressorType egv1a1.CompressorType) {
	compressor := strings.ToLower(string(compressorType)) // Gzip -> gzip
	routeName := fmt.Sprintf("%s-route", compressor)
	gtwName := fmt.Sprintf("%s-gtw", compressor)
	checkPath := fmt.Sprintf("/%s", compressor)

	routeNN := types.NamespacedName{Name: routeName, Namespace: ns}
	gwNN := types.NamespacedName{Name: gtwName, Namespace: ns}
	gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

	// make sure listener is ready
	expectedResponse := httputils.ExpectedResponse{
		Request: httputils.Request{
			Path: checkPath,
		},
		Response: httputils.Response{
			StatusCode: 200,
		},
		Namespace: ns,
	}
	httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

	// make sure compression work as expected
	statsNN := types.NamespacedName{Namespace: "envoy-gateway-system", Name: fmt.Sprintf("%s-gtw-metrics", compressor)}
	var statsHost string
	if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true, func(_ context.Context) (done bool, err error) {
		addr, err := ServiceHost(suite.Client, statsNN, 19001)
		if err != nil {
			tlog.Logf(t, "failed to get service host %s: %v", statsNN, err)
			return false, nil
		}
		if addr != "" {
			statsHost = addr
			return true, nil
		}
		return false, nil
	}); err != nil {
		t.Errorf("failed to get service host %s: %v", statsNN, err)
		return
	}

	statsAddr := fmt.Sprintf("http://%s/stats/prometheus", statsHost)
	tlog.Logf(t, "check stats from %s", statsAddr)

	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true, func(_ context.Context) (done bool, err error) {
		if err := checkStatsEncoding(suite, statsAddr, compressorType); err != nil {
			tlog.Logf(t, "failed to check stats encoding: %v", err)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		tlog.Errorf(t, "failed to check stats encoding: %v", err)
	}
}

func checkStatsEncoding(suite *suite.ConformanceTestSuite, statsAddr string, compressorType egv1a1.CompressorType) error {
	req, err := http.NewRequest("GET", statsAddr, nil)
	if err != nil {
		return err
	}
	encoding := ContentEncoding(compressorType)
	req.Header.Set("Accept-Encoding", encoding)

	client := http.Client{
		Timeout: suite.TimeoutConfig.GetTimeout,
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get response from %s: %w", statsAddr, err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get response from %s, status code %d", statsAddr, res.StatusCode)
	}

	got := res.Header.Get("content-encoding")
	if got != encoding {
		return fmt.Errorf("Content-Encoding is not %s, got %s", encoding, got)
	}

	return nil
}
