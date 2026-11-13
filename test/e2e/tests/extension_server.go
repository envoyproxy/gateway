// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, ExtensionServerVHostPolicyTest, ExtensionServerInferencePoolTest)
}

var ExtensionServerVHostPolicyTest = suite.ConformanceTest{
	ShortName:   "ExtensionServerVHostPolicy",
	Description: "Extension server injects a response header from a VHostMetadataPolicy attached to a listener",
	Manifests:   []string{"testdata/extension-server-vhost-policy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "extension-server-vhost-policy", Namespace: ns}
		gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
			http.ExpectedResponse{
				Request: http.Request{
					Host: "vhost-policy.example.io",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-ext-server-policy": "vhost-policy-applied",
					},
				},
				Namespace: ns,
			},
		)
	},
}

// ExtensionServerInferencePoolTest checks how Envoy Gateway builds clusters for custom backends,
// using InferencePool and the simple extension server started with --inference-pool.
//
// Each InferencePool backendRef gets its own cluster, which the extension turns into an
// ORIGINAL_DST cluster that reads the endpoint from a header specific to that pool. The test sends
// one header per pool, each naming a pod of that pool, so the pod that answers shows which cluster
// Envoy picked. A Service backendRef in the same rule keeps its EDS cluster and is not modified.
//
// If Envoy Gateway passed a cluster more than one InferencePool, the extension would return an
// error and the routes would not be programmed.
var ExtensionServerInferencePoolTest = suite.ConformanceTest{
	ShortName:   "ExtensionServerInferencePool",
	Description: "InferencePool backendRefs get one cluster each and are routed with the backendRef weights",
	Manifests:   []string{"testdata/extension-server-inference-pool.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const (
			host     = "inference-pool.example.com"
			requests = 100
		)
		ns := ConformanceInfraNamespace
		gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN),
			types.NamespacedName{Name: "inference-pools-weighted", Namespace: ns},
			types.NamespacedName{Name: "inference-pool-and-service", Namespace: ns},
			types.NamespacedName{Name: "inference-pool-zero-weight", Namespace: ns},
		)

		podA := readyPodWithIP(t, suite, ns, "pool-a")
		podB := readyPodWithIP(t, suite, ns, "pool-b")
		// pool-svc is reached through its Service (EDS), so make sure its pod is ready as well.
		readyPodWithIP(t, suite, ns, "pool-svc")
		headers := map[string]string{
			"x-pool-a-endpoint": net.JoinHostPort(podA.Status.PodIP, "3000"),
			"x-pool-b-endpoint": net.JoinHostPort(podB.Status.PodIP, "3000"),
		}

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
				req := http.MakeRequest(t, &expected, gwAddr, "HTTP", "http")

				// A single successful response is not enough here: the rule splits traffic across
				// clusters that converge independently (e.g. the EDS cluster of pool-svc may not
				// have endpoints yet while the ORIGINAL_DST cluster of pool-a already works).
				// Wait until every expected backend has answered successfully before measuring.
				http.AwaitConvergence(t, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency,
					func(elapsed time.Duration) bool {
						seen := map[string]bool{}
						for range 20 {
							cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
							if err != nil {
								tlog.Logf(t, "request failed after %v: %v", elapsed, err)
								return false
							}
							if err := http.CompareRoundTrip(t, &req, cReq, cResp, expected); err != nil {
								tlog.Logf(t, "response not as expected after %v: %v", elapsed, err)
								return false
							}
							seen[deploymentForPod(cReq.Pod)] = true
						}
						for app := range tc.expected {
							if !seen[app] {
								tlog.Logf(t, "backend %s has not answered yet after %v, seen %v", app, elapsed, seen)
								return false
							}
						}
						return true
					})

				got := map[string]int{}
				for range requests {
					cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
					require.NoError(t, err)
					require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cResp, expected))
					got[deploymentForPod(cReq.Pod)]++
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

// readyPodWithIP waits for a running and ready pod with the given app label and an IP, and returns it.
func readyPodWithIP(t *testing.T, suite *suite.ConformanceTestSuite, ns, app string) *corev1.Pod {
	var pod *corev1.Pod
	require.Eventually(t, func() bool {
		pods := &corev1.PodList{}
		if err := suite.Client.List(t.Context(), pods, client.InNamespace(ns), client.MatchingLabels{"app": app}); err != nil {
			return false
		}
		for i := range pods.Items {
			p := &pods.Items[i]
			if p.Status.Phase == corev1.PodRunning && p.Status.PodIP != "" && p.DeletionTimestamp == nil && isPodReady(p) {
				pod = p
				return true
			}
		}
		return false
	}, suite.TimeoutConfig.MaxTimeToConsistency, time.Second, "no ready pod for app %s", app)
	return pod
}

func isPodReady(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
}

// deploymentForPod returns the Deployment name for a pod name like "<deployment>-<hash>-<id>".
func deploymentForPod(pod string) string {
	parts := strings.Split(pod, "-")
	if len(parts) < 3 {
		return pod
	}
	return strings.Join(parts[:len(parts)-2], "-")
}
