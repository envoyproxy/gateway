// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

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
)

func init() {
	ConformanceTests = append(ConformanceTests, FileAccessLogTest, OpenTelemetryTest, ALSTest)
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
					Headers: map[string]string{
						"x-envoy-logged": "1",
					},
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
					count, err := QueryLogCountFromLoki(t, suite.Client, labels, "test-annotation-value")
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
					preCount, err := QueryLogCountFromLoki(t, suite.Client, labels, "test-annotation-value")
					if err != nil {
						t.Logf("failed to get log count from loki: %v", err)
						return false, nil
					}

					httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

					// it will take some time for fluent-bit to collect the log and send to loki
					// let's wait for a while
					if err := wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 15*time.Second, true, func(_ context.Context) (bool, error) {
						count, err := QueryLogCountFromLoki(t, suite.Client, labels, "test-annotation-value")
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
					Headers: map[string]string{
						"x-envoy-logged": "1",
					},
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
					preCount, err := QueryLogCountFromLoki(t, suite.Client, labels, "")
					if err != nil {
						t.Logf("failed to get log count from loki: %v", err)
						return false, nil
					}

					httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

					if err := wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 10*time.Second, true, func(_ context.Context) (bool, error) {
						count, err := QueryLogCountFromLoki(t, suite.Client, labels, "")
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

var ALSTest = suite.ConformanceTest{
	ShortName:   "ALS",
	Description: "Make sure ALS access log is working",
	Manifests:   []string{"testdata/accesslog-als.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("HTTP", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "accesslog-als", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			preCount := ALSLogCount(t, suite)
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/als",
					Headers: map[string]string{
						"x-envoy-logged": "1",
					},
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					curCount := ALSLogCount(t, suite)
					return preCount < curCount, nil
				}); err != nil {
				t.Errorf("failed to get log count from loki: %v", err)
			}
		})
	},
}
