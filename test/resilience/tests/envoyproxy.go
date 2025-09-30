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

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/resilience/suite"
)

func init() {
	ResilienceTests = append(ResilienceTests, EPResilience)
}

var EPResilience = suite.ResilienceTest{
	ShortName:   "EPResilience",
	Description: "Envoyproxy resilience test",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		var ()

		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}

		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/base.yaml", true)

		t.Run("Envoy proxies continue to work even when eg is offline", func(t *testing.T) {
			ctx := context.Background()

			t.Log("Scaling down the deployment to 2 replicas")
			err := suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 2, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment replicas")

			t.Log("ensure envoy proxy is running")
			err = suite.Kube().CheckDeploymentReplicas(ctx, envoygateway, namespace, 2, time.Minute)
			require.NoError(t, err, "Failed to check deployment replicas")

			t.Log("Scaling down the deployment to 0 replicas")
			err = suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 0, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment to replicas")

			t.Cleanup(func() {
				err := suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 1, time.Minute, false)
				require.NoError(t, err, "Failed to restore replica count.")
			})

			require.NoError(t, err, "failed to add cleanup")

			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "backend", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/welcome",
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
				if err := http.CompareRoundTrip(t, &req, cReq, cRes, expectedResponse); err != nil {
					tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
					return false
				}
				return true
			})
		})
	},
}
