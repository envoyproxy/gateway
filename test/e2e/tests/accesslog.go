// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, FileAccessLogTest, OpenTelemetryTest)
}

var FileAccessLogTest = suite.ConformanceTest{
	ShortName:   "FileAccessLog",
	Description: "Make sure file access log is working",
	Manifests:   []string{"testdata/accesslog-file.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Stdout", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "accesslog-file", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/file",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			labels := map[string]string{
				"job":                "fluentbit",
				"k8s_namespace_name": "envoy-gateway-system",
				"k8s_container_name": "envoy",
			}
			// let's wait for the log to be sent to stdout
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					// query log count from loki
					count, err := QueryLogCountFromLoki(t, suite.Client, types.NamespacedName{
						Namespace: "envoy-gateway-system",
					}, labels)
					if err != nil {
						t.Logf("failed to get log count from loki: %v", err)
						return false, nil
					}

					if count > 0 {
						return true, nil
					}
					return false, nil
				}); err != nil {
				t.Errorf("failed to wait log flush to loki: %v", err)
			}

			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					// query log count from loki
					preCount, err := QueryLogCountFromLoki(t, suite.Client, types.NamespacedName{
						Namespace: "envoy-gateway-system",
					}, labels)
					if err != nil {
						t.Logf("failed to get log count from loki: %v", err)
						return false, nil
					}

					httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

					// it will take some time for fluent-bit to collect the log and send to loki
					// let's wait for a while
					if err := wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 15*time.Second, true, func(_ context.Context) (bool, error) {
						count, err := QueryLogCountFromLoki(t, suite.Client, types.NamespacedName{
							Namespace: "envoy-gateway-system",
						}, labels)
						if err != nil {
							t.Logf("failed to get log count from loki: %v", err)
							return false, nil
						}

						delta := count - preCount
						if delta == 1 {
							return true, nil
						}

						t.Logf("preCount=%d, count=%d", preCount, count)
						return false, nil
					}); err != nil {
						return false, nil
					}

					return true, nil
				}); err != nil {
				t.Errorf("failed to get log count from loki: %v", err)
			}
		})
	},
}

var OpenTelemetryTest = suite.ConformanceTest{
	ShortName:   "OpenTelemetryAccessLog",
	Description: "Make sure OpenTelemetry access log is working",
	Manifests:   []string{"testdata/accesslog-otel.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("OTel", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "accesslog-otel", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
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

			labels := map[string]string{
				"k8s_namespace_name": "envoy-gateway-system",
				"exporter":           "OTLP",
			}
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					// query log count from loki
					preCount, err := QueryLogCountFromLoki(t, suite.Client, types.NamespacedName{
						Namespace: "envoy-gateway-system",
					}, labels)
					if err != nil {
						t.Logf("failed to get log count from loki: %v", err)
						return false, nil
					}

					httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

					if err := wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 10*time.Second, true, func(_ context.Context) (bool, error) {
						count, err := QueryLogCountFromLoki(t, suite.Client, types.NamespacedName{
							Namespace: "envoy-gateway-system",
						}, labels)
						if err != nil {
							t.Logf("failed to get log count from loki: %v", err)
							return false, nil
						}

						delta := count - preCount
						if delta == 1 {
							return true, nil
						}

						t.Logf("preCount=%d, count=%d", preCount, count)
						return false, nil
					}); err != nil {
						return false, nil
					}

					return true, nil
				}); err != nil {
				t.Errorf("failed to get log count from loki: %v", err)
			}
		})
	},
}

// QueryLogCountFromLoki queries log count from loki
// TODO: move to utils package if needed
func QueryLogCountFromLoki(t *testing.T, c client.Client, nn types.NamespacedName, keyValues map[string]string) (int, error) {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "monitoring",
		Name:      "loki",
	}, &svc); err != nil {
		return -1, err
	}
	lokiHost := ""
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			lokiHost = ing.IP
			break
		}
	}

	qParams := make([]string, 0, len(keyValues))
	for k, v := range keyValues {
		qParams = append(qParams, fmt.Sprintf("%s=\"%s\"", k, v))
	}

	q := "{" + strings.Join(qParams, ",") + "}"
	params := url.Values{}
	params.Add("query", q)
	params.Add("start", fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())) // query logs from last 10 minutes
	lokiQueryURL := fmt.Sprintf("http://%s:3100/loki/api/v1/query_range?%s", lokiHost, params.Encode())
	res, err := http.DefaultClient.Get(lokiQueryURL)
	if err != nil {
		return -1, err
	}
	t.Logf("get response from loki, query=%s, status=%s", q, res.Status)

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}

	lokiResponse := &LokiQueryResponse{}
	if err := json.Unmarshal(b, lokiResponse); err != nil {
		return -1, err
	}

	if len(lokiResponse.Data.Result) == 0 {
		return 0, nil
	}

	total := 0
	for _, res := range lokiResponse.Data.Result {
		total += len(res.Values)
	}
	t.Logf("get response from loki, query=%s, total=%d", q, total)
	return total, nil
}

type LokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric interface{}
			Values []interface{} `json:"values"`
		}
	}
}
