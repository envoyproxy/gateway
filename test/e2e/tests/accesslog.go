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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, FileAccessLogTest, OpenTelemetryTestText, OpenTelemetryTestJSON, ALSTest, OpenTelemetryTestJSONAsDefault)
}

var FileAccessLogTest = suite.ConformanceTest{
	ShortName:   "FileAccessLog",
	Description: "Make sure file access log is working",
	Manifests:   []string{"testdata/accesslog-file.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		gatewayNS := GetGatewayResourceNamespace()
		labels := map[string]string{
			"job":       fmt.Sprintf("%s/envoy", gatewayNS),
			"namespace": gatewayNS,
			"container": "envoy",
		}
		match := "test-annotation-value"

		t.Run("Positive", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "accesslog-file", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/file",
					Headers: map[string]string{
						"x-envoy-logged": "1",
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, match, 1, nil)
		})

		t.Run("Negative", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "accesslog-file", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/file",
					// envoy will not log this request without the header x-envoy-logged
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, match, 0, nil)
		})

		t.Run("TLSRoute Listener Logs", func(t *testing.T) {
			// access log format: LISTENER ACCESS LOG %UPSTREAM_PROTOCOL% %RESPONSE_CODE% %METADATA(LISTENER_FILTER_CHAIN:envoy-gateway:resources)%
			// Ensure that Listener is emitting the log: protocol and response code should be
			// empty in listener logs as they are upstream L7 attributes
			// filter chain metadata is added to TLSRoute filter chain
			expectedMatch := "LISTENER ACCESS LOG - 0.*same-namespace.*"

			ns := "gateway-conformance-infra"
			certNN := types.NamespacedName{Name: "backend-tls-certificate", Namespace: ns}
			routeNN := types.NamespacedName{Name: "tlsroute-acccesslog", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls"), routeNN)

			// make sure listener is ready
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "example.com",
					Path: "/",
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			req := httputils.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "https")

			// This test uses the same key/cert pair as both a client cert and server cert
			// Both backend and client treat the self-signed cert as a trusted CA
			cPem, keyPem, _, err := GetTLSSecret(suite.Client, certNN)
			if err != nil {
				t.Fatalf("unexpected error finding TLS secret: %v", err)
			}

			req.KeyPem = keyPem
			req.CertPem = cPem
			req.Server = "example.com"

			// make sure listener is ready
			httputils.WaitForConsistentResponse(t, suite.RoundTripper, req, expectedResponse, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, expectedMatch, 0, &req)
		})

		t.Run("HTTPRoute Listener Logs", func(t *testing.T) {
			// access log format: LISTENER ACCESS LOG %UPSTREAM_PROTOCOL% %RESPONSE_CODE% %METADATA(LISTENER_FILTER_CHAIN:envoy-gateway:resources)%
			// Ensure that Listener is emitting the log: protocol and response code should be
			// empty in listener logs as they are upstream L7 attributes
			// filter chain metadata is not added to the default filter chain
			expectedMatch := "LISTENER ACCESS LOG - 0 -"
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "accesslog-file", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/file",
					Headers: map[string]string{
						"connection": "close",
					},
				},
				ExpectedRequest: &httputils.ExpectedRequest{
					Request: httputils.Request{
						Path: "/file",
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, expectedMatch, 0, nil)
		})
	},
}

var OpenTelemetryTestText = suite.ConformanceTest{
	ShortName:   "OpenTelemetryTextAccessLog",
	Description: "Make sure OpenTelemetry text access log is working",
	Manifests:   []string{"testdata/accesslog-otel.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		labels := getOTELLabels(ns)
		routeNN := types.NamespacedName{Name: "accesslog-otel", Namespace: ns}
		gwNN := types.NamespacedName{Name: "accesslog-gtw", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		t.Run("Positive", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/otel",
					Headers: map[string]string{
						"x-envoy-logged": "1",
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, "", 1, nil)
		})

		t.Run("Negative", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/otel",
					// envoy will not log this request without the header x-envoy-logged
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, "", 0, nil)
		})
	},
}

var OpenTelemetryTestJSONAsDefault = suite.ConformanceTest{
	ShortName:   "OpenTelemetryAccessLogJSONAsDefault",
	Description: "Make sure OpenTelemetry JSON access log is working as default when no format or type is specified",
	Manifests:   []string{"testdata/accesslog-otel-default.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		labels := getOTELLabels(ns)
		routeNN := types.NamespacedName{Name: "accesslog-otel", Namespace: ns}
		gwNN := types.NamespacedName{Name: "accesslog-gtw", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		t.Run("Positive", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/otel",
					Headers: map[string]string{
						"x-envoy-logged": "1",
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, "", 1, nil)
		})

		t.Run("Negative", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/otel",
					// envoy will not log this request without the header x-envoy-logged
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, "", 0, nil)
		})
	},
}

var OpenTelemetryTestJSON = suite.ConformanceTest{
	ShortName:   "OpenTelemetryAccessLogJSON",
	Description: "Make sure OpenTelemetry JSON access log is working with custom JSON attributes",
	Manifests:   []string{"testdata/accesslog-otel-json.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		labels := getOTELLabels(ns)
		routeNN := types.NamespacedName{Name: "accesslog-otel", Namespace: ns}
		gwNN := types.NamespacedName{Name: "accesslog-gtw", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		t.Run("Positive", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/otel",
					Headers: map[string]string{
						"x-envoy-logged": "1",
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, "", 1, nil)
		})

		t.Run("Negative", func(t *testing.T) {
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/otel",
					// envoy will not log this request without the header x-envoy-logged
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, "", 0, nil)
		})
	},
}

var ALSTest = suite.ConformanceTest{
	ShortName:   "ALS",
	Description: "Make sure ALS access log is working",
	Manifests:   []string{"testdata/accesslog-als.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		labels := map[string]string{
			"exporter": "OTLP",
		}
		match := "common_properties"

		t.Run("HTTP", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "accesslog-als", Namespace: ns}
			gwNN := types.NamespacedName{Name: "accesslog-gtw", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/als",
					Headers: map[string]string{
						"x-envoy-logged": "1",
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			runLogTest(t, suite, gwAddr, &expectedResponse, labels, match, 0, nil)
		})
	},
}

// getOTELLabels returns the appropriate OpenTelemetry labels based on gateway namespace mode
func getOTELLabels(testNamespace string) map[string]string {
	if IsGatewayNamespaceMode() {
		return map[string]string{
			"k8s_namespace_name": testNamespace,
			"exporter":           "OTLP",
		}
	}

	return map[string]string{
		"k8s_namespace_name": "envoy-gateway-system",
		"exporter":           "OTLP",
	}
}

func runLogTest(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr string, expectedResponse *httputils.ExpectedResponse, expectedLabels map[string]string, expectedMatch string, expectedDelta int, customRequest *roundtripper.Request) {
	if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, 3*time.Minute, true,
		func(ctx context.Context) (bool, error) {
			// query log count from loki
			preCount, err := QueryLogCountFromLoki(t, suite.Client, expectedLabels, expectedMatch)
			if err != nil {
				tlog.Logf(t, "failed to get log count from loki: %v", err)
				return false, nil
			}

			if customRequest != nil {
				httputils.WaitForConsistentResponse(t, suite.RoundTripper, *customRequest, *expectedResponse, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency)
			} else {
				httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, *expectedResponse)
			}

			// it will take some time for fluent-bit to collect the log and send to loki
			// let's wait for a while
			if err := wait.PollUntilContextTimeout(ctx, time.Second, 1*time.Minute, true, func(_ context.Context) (bool, error) {
				count, err := QueryLogCountFromLoki(t, suite.Client, expectedLabels, expectedMatch)
				if err != nil {
					tlog.Logf(t, "failed to get log count from loki: %v", err)
					return false, nil
				}

				delta := count - preCount
				if delta >= expectedDelta {
					return true, nil
				}

				tlog.Logf(t, "preCount=%d, count=%d", preCount, count)
				return false, nil
			}); err != nil {
				return false, nil
			}

			return true, nil
		}); err != nil {
		t.Errorf("failed to get log count from loki: %v", err)
	}
}
