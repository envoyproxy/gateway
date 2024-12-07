// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build resilience

package tests

import (
	"context"
	"fmt"
	"github.com/envoyproxy/gateway/test/resilience/suite"
	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"testing"
	"time"
)

func init() {
	ResilienceTests = append(ResilienceTests, EPResilience)
	zapLog, err := zap.NewDevelopment() // Use zap.NewProduction() for production
	if err != nil {
		panic(fmt.Sprintf("Failed to create logger: %v", err))
	}
	log.SetLogger(zapr.NewLogger(zapLog))
}

var EPResilience = suite.ResilienceTest{
	ShortName:   "EGResilience",
	Description: "Kube API server failure and EG gateway",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		var ()

		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}

		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/base.yaml", true)

		t.Run("envoy proxies continue to work even when eg is offline", func(t *testing.T) {
			ctx := context.Background()
			t.Log("ensure envoy proxy is running")
			err := suite.Kube().CheckDeploymentReplicas(ctx, envoygateway, namespace, 2, time.Minute)
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

			resultCh := make(chan error, 1)
			go func() {
				err := suite.Kube().CheckConnectivityJob(fmt.Sprintf("http://%s/route-change", gwAddr), 10)
				resultCh <- err // Send the error (or nil) to the channel
			}()
			err = <-resultCh
			require.NoError(t, err, "Failed during connectivity checkup")
		})
	},
}
