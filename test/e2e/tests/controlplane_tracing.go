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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	controlplanetracing "github.com/envoyproxy/gateway/test/utils/controlplane_tracing"
)

func init() {
	ConformanceTests = append(ConformanceTests, ControlPlaneTracingTest)
}

var ControlPlaneTracingTest = suite.ConformanceTest{
	ShortName:   "ControlPlaneTracing",
	Description: "Verify that control plane traces are being generated and exported to OpenTelemetry collector",
	Manifests:   []string{"testdata/controlplane-tracing.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("OpenTelemetry", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "controlplane-tracing-test", Namespace: ns}
			gwNN := types.NamespacedName{Name: "controlplane-tracing-test", Namespace: ns}

			// Wait for gateway and route to be accepted
			// This will trigger control plane operations that should generate traces
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				&gwapiv1.HTTPRoute{},
				false,
				routeNN,
			)

			// Make a test request to ensure the gateway is fully operational
			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/test",
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(
				t,
				suite.RoundTripper,
				suite.TimeoutConfig,
				gwAddr,
				expectedResponse,
			)

			// Wait for traces to be exported and verify they exist
			// Control plane traces should have service.name=envoy-gateway
			tlog.Logf(t, "waiting for control plane traces to be exported...")
			if err := wait.PollUntilContextTimeout(
				context.TODO(),
				2*time.Second,
				2*time.Minute,
				true,
				func(ctx context.Context) (bool, error) {
					// Query Tempo for control plane traces
					traceCount, err := controlplanetracing.QueryControlPlaneTraces(t, suite.Client, "envoy-gateway")
					if err != nil {
						tlog.Logf(t, "failed to query traces from tempo: %v", err)
						return false, nil
					}

					tlog.Logf(t, "found %d control plane traces", traceCount)

					// We expect at least some traces from the gateway operations
					if traceCount > 0 {
						return true, nil
					}

					return false, nil
				},
			); err != nil {
				t.Errorf("failed to find control plane traces in tempo: %v", err)
			}

			// Verify specific span names exist
			// These span names are created by the instrumented code in the gateway
			tlog.Logf(t, "verifying expected span names exist...")
			expectedSpanNames := []string{
				"GatewayApiRunner.subscribeAndTranslate",
				"XdsRunner.subscribeAndTranslate",
			}

			if err := wait.PollUntilContextTimeout(
				context.TODO(),
				2*time.Second,
				2*time.Minute,
				true,
				func(ctx context.Context) (bool, error) {
					hasExpectedSpans, err := controlplanetracing.VerifyExpectedSpans(
						t,
						suite.Client,
						"envoy-gateway",
						expectedSpanNames,
					)
					if err != nil {
						tlog.Logf(t, "failed to verify expected spans: %v", err)
						return false, nil
					}
					return hasExpectedSpans, nil
				},
			); err != nil {
				t.Errorf("failed to find expected span names: %v", err)
			}

			tlog.Logf(t, "control plane tracing test completed successfully")
		})
	},
}
