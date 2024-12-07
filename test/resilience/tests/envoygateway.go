// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build resilience

package tests

import (
	"context"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/envoyproxy/gateway/test/resilience/suite"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"testing"
	"time"
)

const (
	namespace    = "envoy-gateway-system"
	envoygateway = "envoy-gateway"
	targetString = "successfully acquired lease"
	apiServerIP  = "10.96.0.1"
	timeout      = 5 * time.Minute
	policyName   = "egress-rules"
)

func init() {
	ResilienceTests = append(ResilienceTests, EGResilience)
	// Create a Zap logger
	zapLog, err := zap.NewDevelopment() // Use zap.NewProduction() for production
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}

	// Set the logger for controller-runtime
	log.SetLogger(zapr.NewLogger(zapLog))
}

var EGResilience = suite.ResilienceTest{
	ShortName:   "EGResilience",
	Description: "Kube API server failure and EG gateway",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/base.yaml", true)

		t.Run("envoyproxy reconciles missed resources and sync xds after api server connectivity is restored", func(t *testing.T) {
			err := suite.Kube().ScaleDeploymentAndWait(context.Background(), envoygateway, namespace, 0, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment")
			err = suite.Kube().ScaleDeploymentAndWait(context.Background(), envoygateway, namespace, 1, time.Minute, false)
			require.NoError(t, err, "Failed to scale deployment")

			t.Log("Monitoring logs to identify the leader pod")
			name, err := suite.Kube().MonitorDeploymentLogs(context.Background(), time.Now(), namespace, envoygateway, targetString, timeout, false)
			require.NoError(t, err, "Failed to monitor logs for leader pod")
			require.NotEmpty(t, name, "Leader pod name should not be empty")

			t.Log("Simulating API server down for all pods")
			err = suite.WithResCleanUp(context.Background(), t, func() (client.Object, error) {
				return suite.Kube().ManageEgress(context.Background(), apiServerIP, namespace, policyName, true, map[string]string{})
			})
			require.NoError(t, err, "unable to block api server connectivity")

			err = suite.Kube().WaitForDeploymentReplicaCount(context.Background(), "envoy-gateway", namespace, 0, time.Minute, false)

			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_changes.yaml", true)
			t.Log("backend routes changed")

			t.Log("restore API server connectivity")
			_, err = suite.Kube().ManageEgress(context.Background(), apiServerIP, namespace, policyName, false, map[string]string{})
			require.NoError(t, err, "unable to unblock api server connectivity")

			err = suite.Kube().WaitForDeploymentReplicaCount(context.Background(), "envoy-gateway", namespace, 1, time.Minute, false)
			require.NoError(t, err, "Failed to ensure that pod is online")
			t.Log("eg is online")

			t.Log("Monitoring logs to identify the leader pod")
			name, err = suite.Kube().MonitorDeploymentLogs(context.Background(), time.Now(), namespace, envoygateway, targetString, timeout, false)
			require.NoError(t, err, "Failed to monitor logs")
			require.NotEmpty(t, name, "Leader pod name should not be empty")

			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "backend", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			resultCh := make(chan error, 1)
			go func() {
				t.Log("Connecting to:", fmt.Sprintf("http://%s/route-change", gwAddr))
				err := retry.Do(
					func() error {
						return suite.Kube().CheckConnectivityJob(fmt.Sprintf("http://%s/route-change", gwAddr), 10)
					},
					retry.Attempts(3),                 // Number of retry attempts
					retry.DelayType(retry.FixedDelay), // Use a fixed delay between retries
					retry.Delay(2*time.Second),        // Delay duration between retries
				)
				resultCh <- err
			}()
			err = <-resultCh
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

			t.Log("Monitoring logs to identify the leader pod")
			name, err := suite.Kube().MonitorDeploymentLogs(ctx, time.Now(), namespace, envoygateway, targetString, timeout, false)
			require.NoError(t, err, "Failed to monitor logs for leader pod")
			require.NotEmpty(t, name, "Leader pod name should not be empty")

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
			t.Log("Verifying deployment scales down to 1 replica")
			err = suite.Kube().CheckDeploymentReplicas(ctx, envoygateway, namespace, 1, time.Minute)
			require.NoError(t, err, "Deployment did not scale down")

			t.Log("Monitoring logs for a new leader pod")
			name, err = suite.Kube().MonitorDeploymentLogs(ctx, time.Now().Add(-time.Minute), namespace, envoygateway, targetString, timeout, false)
			require.NoError(t, err, "Failed to monitor logs for new leader pod")
			require.NotEmpty(t, name, "New leader pod name should not be empty")
		})
	},
}
