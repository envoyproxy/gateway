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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"

	"github.com/envoyproxy/gateway/test/resilience/suite"
)

func init() {
	ResilienceTests = append(ResilienceTests, EnvoyProxy)
}

var EnvoyProxy = suite.ResilienceTest{
	ShortName:   "EnvoyProxy",
	Description: "Envoy proxy resilience test",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		// Preserve original convergence semantics for resilience tests
		localTimeout := suite.TimeoutConfig
		localTimeout.RequiredConsecutiveSuccesses = 2
		localTimeout.MaxTimeToConsistency = time.Minute

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
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/welcome",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, localTimeout, gwAddr, expectedResponse)
		})
		t.Run("Envoy proxy receives latest config after reconnecting", func(t *testing.T) {
			ctx := context.Background()

			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "backend", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
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

			t.Log("Verify existing configuration")
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/welcome",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(
				t,
				suite.RoundTripper,
				localTimeout,
				gwAddr,
				expectedResponse,
			)

			t.Log("Disconnecting Envoy proxy from Envoy Gateway xDS")

			xdsService := &corev1.Service{}
			err := suite.Client.Get(
				ctx,
				client.ObjectKey{Name: envoygateway, Namespace: namespace},
				xdsService,
			)
			require.NoError(t, err, "Failed to get Envoy Gateway service")
			require.NotEmpty(t, xdsService.Spec.ClusterIP, "Envoy Gateway service has no ClusterIP")

			proxyScope := map[string]string{
				"gateway.envoyproxy.io/owning-gateway-name":      "all-namespaces",
				"gateway.envoyproxy.io/owning-gateway-namespace": "gateway-resilience",
			}

			const proxyXDSPolicyName = "proxy-xds-egress-rules"
			policyRemoved := false

			t.Cleanup(func() {
				if !policyRemoved {
					_, _ = suite.Kube().ManageEgress(
						context.Background(),
						xdsService.Spec.ClusterIP,
						namespace,
						proxyXDSPolicyName,
						false,
						proxyScope,
					)
				}
			})
			t.Log("Waiting for Envoy proxy to be connected to control plane")
			_, err = waitForMetricValueVerification(
				t,
				suite,
				"envoy_control_plane_connected_state",
				func(actual float64) bool {
					return actual > 0
				},
			)
			require.NoError(t, err, "Failed to confirm Envoy proxy connected to control plane")

			_, err = suite.Kube().ManageEgress(
				ctx,
				xdsService.Spec.ClusterIP,
				namespace,
				proxyXDSPolicyName,
				true,
				proxyScope,
			)
			require.NoError(t, err, "Failed to block Envoy proxy xDS connectivity")

			t.Log("Waiting for Envoy proxy to disconnect from control plane")
			_, err = waitForMetricValueVerification(
				t,
				suite,
				"envoy_control_plane_connected_state",
				func(actual float64) bool {
					return actual == 0
				},
			)
			require.NoError(t, err, "Failed to confirm Envoy proxy disconnected from control plane")

			ap := kubernetes.Applier{
				ManifestFS:     suite.ManifestFS,
				GatewayClass:   suite.GatewayClassName,
				ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
			}

			var envoyGatewayPods corev1.PodList
			err = suite.Client.List(
				t.Context(),
				&envoyGatewayPods,
				client.InNamespace(namespace),
				client.MatchingLabels{"control-plane": "envoy-gateway"},
			)
			require.NoError(t, err, "Failed to list Envoy Gateway pods")

			activePods := make([]string, 0, 1)
			for i := range envoyGatewayPods.Items {
				pod := &envoyGatewayPods.Items[i]
				if pod.Status.Phase == corev1.PodRunning && pod.DeletionTimestamp == nil {
					activePods = append(activePods, pod.Name)
				}
			}
			require.Len(t, activePods, 1, "Expected exactly one running Envoy Gateway pod")

			snapshotQuery := `sum(xds_snapshot_create_total{status="success",namespace="` + namespace + `",pod="` + activePods[0] + `"})`

			snapshotCount, err := waitForMetricValueVerification(
				t,
				suite,
				snapshotQuery,
				func(actual float64) bool {
					return actual > 0
				},
			)
			require.NoError(t, err, "Failed to get current xDS snapshot count")

			t.Log("Updating configuration while Envoy proxy is disconnected")
			ap.MustApplyWithCleanup(
				t,
				suite.Client,
				suite.TimeoutConfig,
				"testdata/route_changes.yaml",
				true,
			)
			t.Log("Waiting for updated route to be accepted")
			gwAddr = kubernetes.GatewayAndRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				&gwapiv1.HTTPRoute{},
				false,
				routeNN,
			)

			t.Log("Waiting for updated configuration to generate a new xDS snapshot")
			_, err = waitForMetricValueVerification(
				t,
				suite,
				snapshotQuery,
				func(actual float64) bool {
					return actual > snapshotCount
				},
			)
			require.NoError(t, err, "Updated route did not generate a new xDS snapshot")

			t.Log("Restoring Envoy proxy xDS connectivity")
			_, err = suite.Kube().ManageEgress(
				ctx,
				xdsService.Spec.ClusterIP,
				namespace,
				proxyXDSPolicyName,
				false,
				proxyScope,
			)
			require.NoError(t, err, "Failed to restore Envoy proxy xDS connectivity")
			policyRemoved = true

			t.Log("Waiting for Envoy proxy to reconnect to control plane")
			_, err = waitForMetricValueVerification(
				t,
				suite,
				"envoy_control_plane_connected_state",
				func(actual float64) bool {
					return actual > 0
				},
			)
			require.NoError(t, err, "Failed to confirm Envoy proxy reconnected to control plane")
			t.Log("Verify Envoy proxy receives latest configuration")
			expectedResponse = http.ExpectedResponse{
				Request: http.Request{
					Path: "/route-change",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(
				t,
				suite.RoundTripper,
				localTimeout,
				gwAddr,
				expectedResponse,
			)
		})
	},
}
