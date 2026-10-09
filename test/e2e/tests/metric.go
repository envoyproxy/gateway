// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		MetricTest,
		MetricWorkqueueAndRestclientTest,
		MetricCompressorGzipTest,
		MetricCompressorBrotliTest,
		MetricCompressorZstdTest,
		OtelMetricPrefixTest,
	)
}

var MetricTest = suite.ConformanceTest{
	ShortName:   "ProxyMetrics",
	Description: "Make sure metric is working",
	Manifests:   []string{"testdata/metric.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "metric-prometheus", Namespace: ns}
		gwNN := types.NamespacedName{Name: "metric-prometheus", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		t.Run("prometheus", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/prom",
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// let's check the metric
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					pql := fmt.Sprintf(`envoy_cluster_default_total_match_count{app_kubernetes_io_component="proxy", app_kubernetes_io_managed_by="envoy-gateway", app_kubernetes_io_name="envoy", envoy_cluster_name="xds_cluster", gateway_envoyproxy_io_owning_gateway_name="%s"}`, "same-namespace")
					v, err := prometheus.QueryPrometheus(suite.Client, pql)
					if err != nil {
						tlog.Logf(t, "failed to get metric: %v", err)
						return false, nil
					}
					if v != nil {
						tlog.Logf(t, "got expected value: %v", v)
						return true, nil
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
					StatusCodes: []int{200},
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

var MetricWorkqueueAndRestclientTest = suite.ConformanceTest{
	ShortName:   "MetricWorkqueueAndRestclientTest",
	Description: "Ensure workqueue and restclient metrics are exposed",
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ctx := context.Background()
		promClient, err := prometheus.NewClient(suite.Client,
			types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
		)
		require.NoError(t, err)

		verifyMetrics := func(t *testing.T, metricQuery, metricName string) {
			httputils.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					v, err := promClient.QuerySum(ctx, metricQuery)
					if err != nil {
						tlog.Logf(t, "failed to get %s metrics: %v", metricName, err)
						return false
					}
					tlog.Logf(t, "%s metrics query count: %v", metricName, v)
					return true
				},
			)
		}

		t.Run("verify workqueue metrics", func(t *testing.T) {
			verifyMetrics(t, `workqueue_adds_total{namespace="envoy-gateway-system"}`, "workqueue")
		})

		t.Run("verify restclient metrics", func(t *testing.T) {
			verifyMetrics(t, `rest_client_request_duration_seconds_sum{namespace="envoy-gateway-system"}`, "restclient")
		})
	},
}

// MetricCompressorGzipTest, MetricCompressorBrotliTest and MetricCompressorZstdTest are split into
// separate ConformanceTests (each with its own manifest) rather than one test applying all three
// Gateways/Services at once, so MetalLB isn't asked to assign three LoadBalancer IPs at the same time.
var MetricCompressorGzipTest = suite.ConformanceTest{
	ShortName:   "MetricCompressorGzip",
	Description: "Make sure metric is working with gzip compressor",
	Manifests:   []string{"testdata/metric-compressor-gzip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		runMetricCompressorTest(t, suite, "gateway-conformance-infra", egv1a1.GzipCompressorType)
	},
}

var MetricCompressorBrotliTest = suite.ConformanceTest{
	ShortName:   "MetricCompressorBrotli",
	Description: "Make sure metric is working with brotli compressor",
	Manifests:   []string{"testdata/metric-compressor-brotli.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		runMetricCompressorTest(t, suite, "gateway-conformance-infra", egv1a1.BrotliCompressorType)
	},
}

var MetricCompressorZstdTest = suite.ConformanceTest{
	ShortName:   "MetricCompressorZstd",
	Description: "Make sure metric is working with zstd compressor",
	Manifests:   []string{"testdata/metric-compressor-zstd.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		runMetricCompressorTest(t, suite, "gateway-conformance-infra", egv1a1.ZstdCompressorType)
	},
}

var OtelMetricPrefixTest = suite.ConformanceTest{
	ShortName:   "OtelMetricPrefix",
	Description: "Make sure metrics arrive with configured prefix",
	Manifests:   []string{"testdata/metrics-with-prefix.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "metric-otel-prefix", Namespace: ns}
		gwNN := types.NamespacedName{Name: "metric-otel-prefix", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		expectedResponse := httputils.ExpectedResponse{
			Request: httputils.Request{
				Path: "/prom-prefix",
			},
			Response: httputils.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

		err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, time.Minute, true, func(_ context.Context) (done bool, err error) {
			ok, err := scrapeMetricsHasPrefix(suite.Client, types.NamespacedName{
				Namespace: "monitoring",
				Name:      "prefix-metrix-gtw-metrics",
			}, 19001, "/metrics", "eg_e2e_prefix")
			if err != nil {
				tlog.Logf(t, "failed to scrape metrics: %v", err)
				return false, nil
			}
			return ok, nil
		})
		require.NoError(t, err, "failed to retrieve metrics with expected prefix within timeout.")
	},
}

func scrapeMetricsHasPrefix(c client.Client, nn types.NamespacedName, port int32, path, prefix string) (bool, error) {
	url, err := RetrieveURL(c, nn, port, path)
	if err != nil {
		return false, err
	}
	mfs, err := RetrieveMetrics(url, 5*time.Second)
	if err != nil {
		return false, err
	}
	for name := range mfs {
		if strings.HasPrefix(name, prefix) {
			return true, nil
		}
	}
	return false, nil
}

func runMetricCompressorTest(t *testing.T, suite *suite.ConformanceTestSuite, ns string, compressorType egv1a1.CompressorType) {
	compressor := strings.ToLower(string(compressorType)) // Gzip -> gzip
	routeName := fmt.Sprintf("%s-route", compressor)
	gtwName := fmt.Sprintf("%s-gtw", compressor)
	checkPath := fmt.Sprintf("/%s", compressor)

	routeNN := types.NamespacedName{Name: routeName, Namespace: ns}
	gwNN := types.NamespacedName{Name: gtwName, Namespace: ns}
	gwHost := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
	gwAddr := net.JoinHostPort(gwHost, "80")

	// make sure listener is ready
	expectedResponse := httputils.ExpectedResponse{
		Request: httputils.Request{
			Path: checkPath,
		},
		Response: httputils.Response{
			StatusCodes: []int{200},
		},
		Namespace: ns,
	}
	httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

	// Scrape the stats endpoint through a port-forward to the proxy pod rather than through the
	// Gateway's LoadBalancer address: MetalLB recycles addresses from its pool as tests create and
	// delete Services, so a probe of <gateway address>:19001 can be answered by whatever owned that
	// address a moment earlier, which shows up as a confusing HTTP error instead of a connect error.
	fwd := proxyStatsForwarder(t, suite, gwNN)
	defer fwd.Stop()
	statsAddr := fmt.Sprintf("http://%s/stats/prometheus", fwd.Address())
	tlog.Logf(t, "check stats from %s", statsAddr)

	err := wait.PollUntilContextTimeout(t.Context(), time.Second, time.Minute, true, func(_ context.Context) (done bool, err error) {
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

// proxyStatsForwarder starts a port-forward to the stats port of the Envoy proxy pod backing the
// given Gateway. The caller is responsible for stopping the returned forwarder.
func proxyStatsForwarder(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName) kube.PortForwarder {
	t.Helper()

	cli, err := kube.NewForRestConfig(suite.RestConfig)
	require.NoError(t, err)

	pods, err := cli.PodsForSelector(GetGatewayResourceNamespace(),
		"app.kubernetes.io/name=envoy",
		fmt.Sprintf("gateway.envoyproxy.io/owning-gateway-name=%s", gwNN.Name),
		fmt.Sprintf("gateway.envoyproxy.io/owning-gateway-namespace=%s", gwNN.Namespace),
	)
	require.NoError(t, err)
	require.NotEmpty(t, pods.Items, "no Envoy proxy pod found for Gateway %s", gwNN.String())

	// stats are exposed at port 19001
	fwd, err := kube.NewLocalPortForwarder(cli, types.NamespacedName{
		Namespace: pods.Items[0].Namespace,
		Name:      pods.Items[0].Name,
	}, 0, 19001)
	require.NoError(t, err)
	require.NoError(t, fwd.Start())

	return fwd
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
