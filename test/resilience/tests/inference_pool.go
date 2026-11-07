// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build resilience

package tests

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"

	"github.com/envoyproxy/gateway/test/resilience/suite"
)

func init() {
	ResilienceTests = append(ResilienceTests, InferencePoolWeightedClusters)
}

// InferencePoolWeightedClusters checks how Envoy Gateway builds clusters for custom backends,
// using InferencePool and the simple extension server started with --inference-pool.
//
// Each InferencePool backendRef gets its own cluster, which the extension turns into an
// ORIGINAL_DST cluster that reads the endpoint from a header specific to that pool. The test sends
// one header per pool, each naming a pod of that pool, so the pod that answers shows which cluster
// Envoy picked. A Service backendRef in the same rule keeps its EDS cluster and is not modified.
//
// If Envoy Gateway passed a cluster more than one InferencePool, the extension would return an
// error and the routes would not be programmed.
var InferencePoolWeightedClusters = suite.ResilienceTest{
	ShortName:   "InferencePoolWeightedClusters",
	Description: "InferencePool backendRefs get one cluster each and are routed with the backendRef weights",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		const (
			ns       = "gateway-resilience"
			host     = "inference-pool.example.com"
			requests = 100
		)
		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}

		// Install the InferencePool CRD before the control plane restarts so its provider can
		// watch it; otherwise the cache sync times out and the pod enters CrashLoopBackOff.
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/inferencepool_crd.yaml", true)
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/inferencepool_rbac.yaml", true)
		useEnvoyGatewayConfig(t, suite, &ap, "testdata/config_inference_pool.yaml")

		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_for_inference_pool.yaml", true)
		gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN),
			types.NamespacedName{Name: "inference-pools-weighted", Namespace: ns},
			types.NamespacedName{Name: "inference-pool-and-service", Namespace: ns},
			types.NamespacedName{Name: "inference-pool-zero-weight", Namespace: ns},
		)

		podA := readyPod(t, suite, ns, "pool-a")
		podB := readyPod(t, suite, ns, "pool-b")
		headers := map[string]string{
			"x-pool-a-endpoint": net.JoinHostPort(podA.Status.PodIP, "3000"),
			"x-pool-b-endpoint": net.JoinHostPort(podB.Status.PodIP, "3000"),
		}

		localTimeout := suite.TimeoutConfig
		localTimeout.RequiredConsecutiveSuccesses = 2
		localTimeout.MaxTimeToConsistency = time.Minute

		cases := []struct {
			name     string
			path     string
			expected map[string]int // backend app label -> expected requests out of `requests`
		}{
			{name: "two pools weighted 3:1", path: "/pools", expected: map[string]int{"pool-a": 75, "pool-b": 25}},
			{name: "service and pool weighted 1:1", path: "/mixed", expected: map[string]int{"pool-svc": 50, "pool-a": 50}},
			{name: "pool with weight 0 gets no traffic", path: "/zero", expected: map[string]int{"pool-a": requests}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				expected := http.ExpectedResponse{
					Request:   http.Request{Host: host, Path: tc.path, Headers: headers},
					Response:  http.Response{StatusCodes: []int{200}},
					Namespace: ns,
				}
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, localTimeout, gwAddr, expected)

				req := http.MakeRequest(t, &expected, gwAddr, "HTTP", "http")
				got := map[string]int{}
				for range requests {
					cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
					require.NoError(t, err)
					require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cResp, expected))
					got[appForPod(cReq.Pod)]++
				}

				// Allow some slack around the expected split, as the weighted pick is random.
				const tolerance = 15
				for app, want := range tc.expected {
					require.InDelta(t, want, got[app], tolerance, "requests to %s, got %v", app, got)
				}
				for app, n := range got {
					_, ok := tc.expected[app]
					require.True(t, ok, "unexpected %d requests to %s, got %v", n, app, got)
				}
			})
		}
	},
}

// useEnvoyGatewayConfig replaces the envoy-gateway-config ConfigMap with the one in file and
// restarts the control plane, restoring the original config when the test ends.
func useEnvoyGatewayConfig(t *testing.T, suite *suite.ResilienceTestSuite, ap *kubernetes.Applier, file string) {
	ctx := t.Context()
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

	// cleanup=false: the Applier would otherwise delete the helm-owned ConfigMap; the restore
	// Cleanup above handles teardown.
	ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, file, false)
	require.NoError(t, suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 0, time.Minute, false),
		"Failed to scale down envoy-gateway")
	require.NoError(t, suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 1, time.Minute, false),
		"Failed to scale up envoy-gateway")
}

// readyPod waits for a running pod with the given app label and an IP, and returns it.
func readyPod(t *testing.T, suite *suite.ResilienceTestSuite, ns, app string) *corev1.Pod {
	var pod *corev1.Pod
	require.Eventually(t, func() bool {
		pods := &corev1.PodList{}
		if err := suite.Client.List(t.Context(), pods, client.InNamespace(ns), client.MatchingLabels{"app": app}); err != nil {
			return false
		}
		for i := range pods.Items {
			p := &pods.Items[i]
			if p.Status.Phase == corev1.PodRunning && p.Status.PodIP != "" && p.DeletionTimestamp == nil {
				pod = p
				return true
			}
		}
		return false
	}, suite.TimeoutConfig.MaxTimeToConsistency, time.Second, "no running pod for app %s", app)
	return pod
}

// appForPod returns the Deployment name for a pod name like "<deployment>-<hash>-<id>".
func appForPod(pod string) string {
	parts := strings.Split(pod, "-")
	if len(parts) < 3 {
		return pod
	}
	return strings.Join(parts[:len(parts)-2], "-")
}
