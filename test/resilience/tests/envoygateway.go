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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/resilience/suite"
)

const (
	namespace    = "envoy-gateway-system"
	envoygateway = "envoy-gateway"
	targetString = "successfully acquired lease"
	apiServerIP  = "10.96.0.1"
	timeout      = 2 * time.Minute
	policyName   = "egress-rules"
	leaseName    = "5b9825d2.gateway.envoyproxy.io"
	trashHold    = 2
)

func init() {
	ResilienceTests = append(ResilienceTests, EGResilience)
}

var EGResilience = suite.ResilienceTest{
	ShortName:   "EGResilience",
	Description: "Envoygateway resilience test",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/base.yaml", true)

		// this test will fail until https://github.com/envoyproxy/gateway/pull/4767/files is merged
		t.Run("Secondary EnvoyGateway instances can serve an up to date xDS", func(t *testing.T) {
			ctx := context.Background()
			t.Log("Scaling down the deployment to 0 replicas")
			err := suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 0, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment replicas")

			t.Log("Scaling up the deployment to 3 replicas")
			err = suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 3, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment replicas")

			t.Log("Waiting for leader election")
			// Ensure leadership was taken
			name, err := suite.Kube().GetElectedLeader(context.Background(), namespace, leaseName, metav1.Now(), time.Minute*2)
			require.NoError(t, err, "unable to detect leader election")

			t.Log("Marking the identified pod as leader")
			suite.Kube().MarkAsLeader(namespace, name)

			// Pods rely on connectivity to the API server to participate in leader election processes.
			// Without this connectivity, they cannot become leaders, in this test we won't bring it back.
			// The secondary pods will continue to operate using their last known good configuration (xDS)
			// and share it with envoy proxies accordingly.
			t.Log("Simulating API server connection failure for all pods")
			err = suite.WithResCleanUp(ctx, t, func() (client.Object, error) {
				return suite.Kube().ManageEgress(ctx, apiServerIP, namespace, policyName, true, map[string]string{"app.kubernetes.io/name": "gateway-helm"})
			})
			require.NoError(t, err, "Failed to simulate API server connection failure")

			// The leader pod should go down, the standby pods remain
			// When a leader pod loses connectivity to the API server, Kubernetes does not immediately terminate or stop the pod.
			// Instead, the pod itself detects the loss of connectivity, initiates a graceful teardown process, and restarts to attempt
			// to reconnect to the API server.
			// The replica count for the deployment remains at 3 throughout the process.
			// Kubernetes does not schedule a new pod to replace the one that lost connectivity because the existing pod is not
			// considered failed from Kubernetes’ perspective. It’s the responsibility of the application running inside the
			// pod (e.g., the leader election logic) to handle reconnection attempts or restart itself.
			t.Log("Verifying deployment scales down to 2 replica")
			err = suite.Kube().CheckDeploymentReplicas(ctx, envoygateway, namespace, 2, time.Minute)
			require.NoError(t, err, "Deployment did not scale down")

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
			http.AwaitConvergence(t, trashHold, timeout, func(elapsed time.Duration) bool {
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

		t.Run("EnvoyGateway reconciles missed resources and sync xDS after api server connectivity is restored", func(t *testing.T) {
			err := suite.Kube().ScaleDeploymentAndWait(context.Background(), envoygateway, namespace, 0, timeout, false)
			require.NoError(t, err, "Failed to scale deployment")
			err = suite.Kube().ScaleDeploymentAndWait(context.Background(), envoygateway, namespace, 1, timeout, false)
			require.NoError(t, err, "Failed to scale deployment")

			// Ensure leadership was taken
			_, err = suite.Kube().GetElectedLeader(context.Background(), namespace, leaseName, metav1.Now(), timeout)
			require.NoError(t, err, "unable to detect leader election")

			t.Log("Simulating API server down for all pods")
			err = suite.WithResCleanUp(context.Background(), t, func() (client.Object, error) {
				return suite.Kube().ManageEgress(context.Background(), apiServerIP, namespace, policyName, true, map[string]string{})
			})
			require.NoError(t, err, "unable to block api server connectivity")

			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_changes.yaml", true)
			t.Log("backend routes changed")

			t.Log("restore API server connectivity")
			_, err = suite.Kube().ManageEgress(context.Background(), apiServerIP, namespace, policyName, false, map[string]string{})
			require.NoError(t, err, "unable to unblock api server connectivity")

			err = suite.Kube().WaitForDeploymentReplicaCount(context.Background(), envoygateway, namespace, 1, time.Minute, false)
			require.NoError(t, err, "Failed to ensure that pod is online")
			_, err = suite.Kube().GetElectedLeader(context.Background(), namespace, leaseName, metav1.Now(), time.Minute*2)
			require.NoError(t, err, "unable to detect leader election")
			t.Log("eg is online")
			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "backend", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/route-change",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "http", "http")
			http.AwaitConvergence(t, trashHold, time.Minute, func(elapsed time.Duration) bool {
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

			require.NoError(t, err, "Failed during connectivity checkup")
		})

		t.Run("Leader election transitions when leader loses API server connection", func(t *testing.T) {
			ctx := context.Background()
			t.Log("Scaling down the deployment to 0 replicas")
			err := suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 0, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment replicas")

			t.Log("Scaling up the deployment to 2 replicas")
			err = suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 2, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment replicas")

			t.Log("Waiting for leader election")
			// Ensure leadership was taken
			name, err := suite.Kube().GetElectedLeader(context.Background(), namespace, leaseName, metav1.Now(), time.Minute*2)
			require.NoError(t, err, "unable to detect leader election")

			t.Log("Marking the identified pod as leader")
			suite.Kube().MarkAsLeader(namespace, name)

			t.Log("Simulating API server connection failure for the leader")
			err = suite.WithResCleanUp(ctx, t, func() (client.Object, error) {
				return suite.Kube().ManageEgress(ctx, apiServerIP, namespace, policyName, true, map[string]string{
					"leader": "true",
				})
			})
			require.NoError(t, err, "Failed to simulate API server connection failure")

			// leader pod should go down, the standby remain
			t.Log("Verifying deployment scales down to 1 replicas")
			err = suite.Kube().CheckDeploymentReplicas(ctx, envoygateway, namespace, 1, time.Minute)
			require.NoError(t, err, "Deployment did not scale down")

			// Ensure leadership was taken
			newLeader, err := suite.Kube().GetElectedLeader(context.Background(), namespace, leaseName, metav1.Now(), time.Minute*2)
			require.NoError(t, err, "unable to detect leader election")
			require.NotEqual(t, newLeader, name, "new leader name should not be equal to the first leader")
			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_changes.yaml", true)
			t.Log("backend routes changed")

			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "backend", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/route-change",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "http", "http")

			http.AwaitConvergence(t, trashHold, timeout, func(elapsed time.Duration) bool {
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
