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
	ResilienceTests = append(ResilienceTests, MultipleExtManagers)
}

var MultipleExtManagers = suite.ResilienceTest{
	ShortName:   "MultipleExtManagers",
	Description: "Multiple extension managers chaining and resource isolation",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		ctx := t.Context()
		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}

		// Install the extension CRDs before the control plane restarts so its provider
		// can establish watches on ext-a/ext-b resources; otherwise controller-runtime
		// times out on cache sync and the pod enters CrashLoopBackOff.
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

		// Apply the multi-manager ConfigMap (cleanup=false so the Applier does not Delete
		// the pre-existing helm-owned resource; the restore Cleanup above handles teardown)
		// and restart the control plane to pick it up.
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/config_multiple_ext_managers.yaml", false)
		err := suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 0, time.Minute, false)
		require.NoError(t, err, "Failed to scale down envoy-gateway")
		err = suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 1, time.Minute, false)
		require.NoError(t, err, "Failed to scale up envoy-gateway")

		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/base.yaml", true)

		localTimeout := suite.TimeoutConfig
		localTimeout.RequiredConsecutiveSuccesses = 2
		localTimeout.MaxTimeToConsistency = time.Minute

		t.Run("Chaining applies both VirtualHost mutations in order", func(t *testing.T) {
			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "chaining-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}

			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_for_multiple_ext_managers.yaml", true)
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			t.Log("Verify base route works")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, localTimeout, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Host: "www.chain.com",
					Path: "/chain",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			})

			t.Log("Verify chained VirtualHost mutations: ext-a appends .ext-a, ext-b appends .ext-b")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, localTimeout, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Host: "www.chain.com.ext-a.ext-b",
					Path: "/chain",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			})
		})

		t.Run("Resource isolation: only owning extension receives resources", func(t *testing.T) {
			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "isolation-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}

			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_for_resource_isolation.yaml", true)
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			t.Log("Verify ext-a processed the route with its extension resource (FooFilter)")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, localTimeout, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Host: "www.isolation.com",
					Path: "/isolate",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-ext-server": "ext-a",
					},
				},
				Namespace: ns,
			})
		})
	},
}
