// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/utils/naming"
)

func init() {
	ConformanceTests = append(ConformanceTests, OpenTelemetryTracingTest, ZipkinTracingTest, DatadogTracingTest)
}

var OpenTelemetryTracingTest = suite.ConformanceTest{
	ShortName:   "OpenTelemetryTracing",
	Description: "Make sure OpenTelemetry tracing is working",
	Manifests:   []string{"testdata/tracing-otel.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("tempo", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "tracing-otel", Namespace: ns}
			gwNN := types.NamespacedName{Name: "tracing-otel", Namespace: ns}
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
			// let's wait for the log to be sent to stdout
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					count, err := QueryTraceFromTempo(t, suite.Client, tags)
					if err != nil {
						tlog.Logf(t, "failed to get trace count from tempo: %v", err)
						return false, nil
					}

					if count > 0 {
						return true, nil
					}
					return false, nil
				}); err != nil {
				t.Errorf("failed to get trace from tempo: %v", err)
			}
		})
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
					StatusCode: 200,
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
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					preCount, err := QueryTraceFromTempo(t, suite.Client, tags)
					if err != nil {
						tlog.Logf(t, "failed to get trace count from tempo: %v", err)
						return false, nil
					}

					httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

					// looks like we need almost 15 seconds to get the trace from Tempo?
					err = wait.PollUntilContextTimeout(context.TODO(), time.Second, 15*time.Second, true, func(ctx context.Context) (done bool, err error) {
						curCount, err := QueryTraceFromTempo(t, suite.Client, tags)
						if err != nil {
							tlog.Logf(t, "failed to get curCount count from tempo: %v", err)
							return false, nil
						}

						if curCount > preCount {
							return true, nil
						}

						return false, nil
					})
					if err != nil {
						tlog.Logf(t, "failed to get current count from tempo: %v", err)
						return false, nil
					}

					return true, nil
				}); err != nil {
				t.Errorf("failed to get trace from tempo: %v", err)
			}
		})
	},
}

var DatadogTracingTest = suite.ConformanceTest{
	ShortName:   "DatadogTracing",
	Description: "Make sure Datadog tracing is working",
	Manifests:   []string{"testdata/tracing-datadog.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("tempo", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "tracing-datadog", Namespace: ns}
			gwNN := types.NamespacedName{Name: "eg-special-case-datadog", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/datadog",
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
				"provider":     "datadog",
				"service.name": fmt.Sprintf("%s.%s", gwNN.Name, gwNN.Namespace),
			}
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(ctx context.Context) (bool, error) {
					preCount, err := QueryTraceFromTempo(t, suite.Client, tags)
					if err != nil {
						tlog.Logf(t, "failed to get trace count from tempo: %v", err)
						return false, nil
					}

					httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

					// looks like we need almost 15 seconds to get the trace from Tempo?
					err = wait.PollUntilContextTimeout(context.TODO(), time.Second, 60*time.Second, true, func(ctx context.Context) (done bool, err error) {
						curCount, err := QueryTraceFromTempo(t, suite.Client, tags)
						if err != nil {
							tlog.Logf(t, "failed to get curCount count from tempo: %v", err)
							return false, nil
						}

						if curCount > preCount {
							return true, nil
						}

						return false, nil
					})
					if err != nil {
						tlog.Logf(t, "failed to get current count from tempo: %v", err)
						return false, nil
					}

					return true, nil
				}); err != nil {
				t.Errorf("failed to get trace from tempo: %v", err)
			}
		})
	},
}
