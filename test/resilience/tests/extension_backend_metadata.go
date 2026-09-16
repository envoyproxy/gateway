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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"

	"github.com/envoyproxy/gateway/test/resilience/suite"
)

const PrometheusXDSSnapshotSuccess = `xds_snapshot_create_total{status="success"}`

func init() {
	ResilienceTests = append(ResilienceTests, ExtensionBackendMetadata)
}

var ExtensionBackendMetadata = suite.ResilienceTest{
	ShortName:   "ExtensionBackendMetadata",
	Description: "Extension backend metadata updates trigger reconciliation",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		ctx := t.Context()
		ns := "gateway-resilience"
		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}

		// Install the extension CRDs (including FooBackend) before the control plane
		// restarts so its provider can establish a watch on it; otherwise
		// controller-runtime times out on cache sync and the pod enters CrashLoopBackOff.
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/extension_crds.yaml", true)

		// Capture the original envoy-gateway-config data so we can restore it on cleanup.
		// The conformance Applier's cleanup deletes resources it Updated, which would
		// remove the helm-owned ConfigMap and wedge the cluster for subsequent test runs.
		cmKey := client.ObjectKey{Name: "envoy-gateway-config", Namespace: namespace}
		originalCM := &corev1.ConfigMap{}
		require.NoError(t, suite.Client.Get(ctx, cmKey, originalCM), "Failed to read original envoy-gateway-config")
		originalData := originalCM.Data
		t.Cleanup(func() {
			restoreCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()
			cm := &corev1.ConfigMap{}
			if err := suite.Client.Get(restoreCtx, cmKey, cm); err != nil {
				t.Logf("could not fetch envoy-gateway-config for restore: %v", err)
				return
			}
			cm.Data = originalData
			if err := suite.Client.Update(restoreCtx, cm); err != nil {
				t.Logf("failed to restore envoy-gateway-config: %v", err)
				return
			}
			if err := suite.Kube().ScaleDeploymentAndWait(restoreCtx, envoygateway, namespace, 0, time.Minute, false); err != nil {
				t.Logf("failed to scale down envoy-gateway during restore: %v", err)
			}
			if err := suite.Kube().ScaleDeploymentAndWait(restoreCtx, envoygateway, namespace, 1, time.Minute, false); err != nil {
				t.Logf("failed to scale up envoy-gateway during restore: %v", err)
			}
		})

		// Apply the extensionManager ConfigMap (cleanup=false so the Applier does not
		// Delete the pre-existing helm-owned resource; the restore Cleanup above handles
		// teardown) and restart the control plane to pick it up. This reuses the
		// gateway-simple-extension-server already deployed by `make enable-simple-extension-server`.
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/config_extension_backend_metadata.yaml", false)
		err := suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 0, time.Minute, false)
		require.NoError(t, err, "Failed to scale down envoy-gateway")
		err = suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 1, time.Minute, false)
		require.NoError(t, err, "Failed to scale up envoy-gateway")

		routeNN := types.NamespacedName{Name: "extension-backend-metadata", Namespace: ns}
		gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_for_extension_backend_metadata.yaml", true)
		kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		backend := &unstructured.Unstructured{}
		backend.SetAPIVersion("ext-a.example.io/v1")
		backend.SetKind("FooBackend")
		require.NoError(t, suite.Client.Get(ctx, types.NamespacedName{Namespace: ns, Name: "metadata-backend"}, backend))

		before, err := waitForMetricValueVerification(t, suite, PrometheusXDSSnapshotSuccess, func(actual float64) bool {
			return actual >= 0
		})
		require.NoError(t, err, "Failed to get initial xDS snapshot count")

		backend.SetAnnotations(map[string]string{"example.io/revision": "2"})
		require.NoError(t, suite.Client.Update(ctx, backend))

		_, err = waitForMetricValueVerification(t, suite, PrometheusXDSSnapshotSuccess, func(actual float64) bool {
			return actual > before
		})
		require.NoError(t, err, "backend annotation update did not regenerate xDS")
	},
}
