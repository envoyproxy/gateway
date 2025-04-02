// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build resilience

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/resilience/suite"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ResilienceTests = append(ResilienceTests, ESResilience)
}

var ESResilience = suite.ResilienceTest{
	ShortName:   "ESResilience",
	Description: "Extension Server resilience test",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		const (
			namespace                              = "envoy-gateway-system"
			PrometheusXDSTranslatorErrors          = `watchable_subscribe_total{runner="xds-translator", status="failure"}`
			PrometheusEnvoyConnectedToControlPlane = `envoy_control_plane_connected_state`
		)

		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}

		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/base.yaml", true)

		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_for_extension_manager.yaml", true)

		t.Run("Verify Snapshot Preservation after XDS translation error when using failOpen mode", func(t *testing.T) {
			ctx := context.Background()

			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "valid-route-for-extension-server", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			t.Log("Route is translated")
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.pass.com",
					Path: "/pass",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "http", "http")
			http.AwaitConvergence(t, threshold, timeout, func(elapsed time.Duration) bool {
				cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
					return false
				}
				if err := http.CompareRequest(t, &req, cReq, cRes, expectedResponse); err != nil {
					tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
					return false
				}
				return true
			})

			t.Log("Check Route is modified by extension server")
			// ensure the extension server is connected and mutating resources
			expectedResponse = http.ExpectedResponse{
				Request: http.Request{
					Host: "www.pass.com.extserver", // this domain was added by extension server
					Path: "/pass",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req = http.MakeRequest(t, &expectedResponse, gwAddr, "http", "http")
			http.AwaitConvergence(t, threshold, timeout, func(elapsed time.Duration) bool {
				cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
					return false
				}
				if err := http.CompareRequest(t, &req, cReq, cRes, expectedResponse); err != nil {
					tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
					return false
				}
				return true
			})

			t.Log("Getting current error translation count")
			translatorErrors, err := waitForMetricValueVerification(t, suite, PrometheusXDSTranslatorErrors, func(actual float64) bool {
				return actual > -1
			})
			require.NoError(t, err, "Failed to get initial translator errors")

			t.Log("Creating a translation error with an additional route")
			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_fail_extension_manager.yaml", true)
			kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			t.Log("Waiting for translation error metric to increase from current state")
			translatorErrors, err = waitForMetricValueVerification(t, suite, PrometheusXDSTranslatorErrors, func(actual float64) bool {
				return actual > translatorErrors
			})
			require.NoError(t, err, "Failed to capture translator error increase")

			t.Log("Confirming cache preserved after translation error")
			// updating valid route, update should not be reflected
			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_for_extension_manager_update.yaml", true)

			t.Log("Restarting proxy and confirming control plane cache still preserved")
			err = suite.Kube().ScaleDeploymentAndWait(ctx, envoyproxy, namespace, 0, time.Minute, true)
			require.NoError(t, err, "Failed to scale deployment replicas")

			err = suite.Kube().ScaleDeploymentAndWait(ctx, envoyproxy, namespace, 1, time.Minute, true)
			require.NoError(t, err, "Failed to scale deployment replicas")

			// confirm that the old (pre-error) cache is preserved
			req = http.MakeRequest(t, &expectedResponse, gwAddr, "http", "http")
			http.AwaitConvergence(t, threshold, timeout, func(elapsed time.Duration) bool {
				cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
					return false
				}
				if err := http.CompareRequest(t, &req, cReq, cRes, expectedResponse); err != nil {
					tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
					return false
				}
				return true
			})

			t.Log("Deleting control plane cache and confirming proxy cache is still preserved")
			t.Log("Scaling down control plane")
			err = suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 0, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment replicas")

			t.Log("Waiting for proxy to be disconnected from stopped control plane")
			_, err = waitForMetricValueVerification(t, suite, PrometheusEnvoyConnectedToControlPlane, func(actual float64) bool {
				return actual == 0
			})
			require.NoError(t, err, "Failed to get proxy metrics")

			t.Log("Scaling up control plane")
			err = suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 1, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment replicas")

			t.Log("Waiting for proxy to be connected to restarted control plane")
			_, err = waitForMetricValueVerification(t, suite, PrometheusEnvoyConnectedToControlPlane, func(actual float64) bool {
				return actual > 0
			})
			require.NoError(t, err, "Failed to capture envoy connection to control plane")

			t.Log("Checking proxy cache is still preserved when connected to control plane with no snapshot")
			req = http.MakeRequest(t, &expectedResponse, gwAddr, "http", "http")
			http.AwaitConvergence(t, threshold, timeout, func(elapsed time.Duration) bool {
				cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
					return false
				}
				if err := http.CompareRequest(t, &req, cReq, cRes, expectedResponse); err != nil {
					tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
					return false
				}
				return true
			})

			t.Log("Fixing translation error by changing the failing route")
			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_fail_extension_manager_update.yaml", true)

			t.Log("Verifying configuration updates are now reflected by calling the updated endpoint")
			expectedResponse = http.ExpectedResponse{
				Request: http.Request{
					Host: "www.pass.com",
					Path: "/pass-updated",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req = http.MakeRequest(t, &expectedResponse, gwAddr, "http", "http")

			http.AwaitConvergence(t, threshold, timeout, func(elapsed time.Duration) bool {
				cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
					return false
				}
				if err := http.CompareRequest(t, &req, cReq, cRes, expectedResponse); err != nil {
					tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
					return false
				}
				return true
			})
		})
	},
}

func waitForMetricValueVerification(t *testing.T, suite *suite.ResilienceTestSuite, query string, verifier func(actual float64) bool) (float64, error) {
	var actual float64
	if err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, time.Minute, true, func(ctx context.Context) (done bool, err error) {
		v, err := prometheus.QueryPrometheus(suite.Client, query)
		if err != nil {
			tlog.Logf(t, "failed to query prometheus: %v", err)
			return false, err
		}
		tlog.Logf(t, "query response: %v", v)
		if v != nil {
			var fv float64
			switch {
			case v.Type() == model.ValScalar:
				fv = float64(v.(*model.Scalar).Value)
			case v.Type() == model.ValVector:
				vectorVal := v.(model.Vector)
				if len(vectorVal) == 0 {
					fv = 0
				} else {
					fv = float64(vectorVal[0].Value)
				}
			}

			if verifier(fv) {
				tlog.Logf(t, "got expected value: %v", fv)
				actual = fv
				return true, nil
			} else {
				tlog.Logf(t, "got unexpected value: %v", fv)
			}
		}
		return false, nil
	}); err != nil {
		t.Errorf("failed to get expected response for the last request for metrics: %v", err)
		return 0, err
	}

	return actual, nil
}
