// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, EndpointFastPathTest)
}

var EndpointFastPathTest = suite.ConformanceTest{
	ShortName:   "EndpointFastPath",
	Description: "Endpoint updates and empty-backend fallback converge with fast publication enabled",
	Manifests:   []string{"testdata/endpoint-fast-path.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		if DeployProfile != "endpoint-fast-path" {
			t.Skip("requires the endpoint-fast-path deployment profile")
		}
		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, []string{ConformanceInfraNamespace})
		// Use a selectorless Service so only this test controls its endpoint inputs.
		endpoints := make([]discoveryv1.Endpoint, 0, 2)
		for _, app := range []string{"infra-backend-v1", "infra-backend-v2"} {
			var pods corev1.PodList
			require.NoError(t, s.Client.List(t.Context(), &pods, client.InNamespace(ConformanceInfraNamespace), client.MatchingLabels{"app": app}))
			require.NotEmpty(t, pods.Items)
			require.NotEmpty(t, pods.Items[0].Status.PodIP)
			endpoints = append(endpoints, discoveryv1.Endpoint{
				Addresses:  []string{pods.Items[0].Status.PodIP},
				Conditions: discoveryv1.EndpointConditions{Ready: new(true)},
			})
		}
		slice := &discoveryv1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Name: "endpoint-fast-path", Namespace: ConformanceInfraNamespace,
				Labels: map[string]string{discoveryv1.LabelServiceName: "endpoint-fast-path", discoveryv1.LabelManagedBy: "endpoint-fast-path-test"},
			},
			AddressType: discoveryv1.AddressTypeIPv4,
			Ports:       []discoveryv1.EndpointPort{{Name: new("http"), Protocol: new(corev1.ProtocolTCP), Port: new(int32(3000))}},
			Endpoints:   []discoveryv1.Endpoint{endpoints[0]},
		}
		require.NoError(t, s.Client.Create(t.Context(), slice))
		t.Cleanup(func() { require.NoError(t, client.IgnoreNotFound(s.Client.Delete(context.Background(), slice))) })
		route := types.NamespacedName{Namespace: ConformanceInfraNamespace, Name: "endpoint-fast-path"}
		address := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName,
			kubernetes.NewGatewayRef(SameNamespaceGateway), route)
		expect := func(backend string, code int) {
			t.Helper()
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, address, http.ExpectedResponse{
				Request: http.Request{Path: "/endpoint-fast-path"}, Response: http.Response{StatusCode: code},
				Backend: backend, Namespace: ConformanceInfraNamespace,
			})
		}
		expect("infra-backend-v1", 200)
		for _, step := range []struct {
			index, code int
			backend     string
			fast, full  bool
		}{
			// Swap the v1 endpoint for v2. The endpoint count stays at one, so the
			// fast path handles it; also check the fast path metrics.
			{index: 1, code: 200, backend: "infra-backend-v2", fast: true},
			// Remove all endpoints. Going to zero always takes the full path, which
			// turns the route into a direct 503 response. The fast path either skips it
			// as an empty transition or finds no cluster to update, depending on which
			// path runs first, so only the full publication counter is checked.
			{index: -1, code: 503, full: true},
			// Restore the v1 endpoint. Coming back from zero also takes the full path:
			// the cluster was removed with the 503, so the fast path has nothing to update.
			{index: 0, code: 200, backend: "infra-backend-v1", full: true},
		} {
			var before, fullBefore float64
			if step.fast {
				before = fastPathHandled(t, s)
			}
			if step.full {
				fullBefore = fullPublications(t, s)
			}
			require.NoError(t, s.Client.Get(t.Context(), client.ObjectKeyFromObject(slice), slice))
			slice.Endpoints = nil
			if step.index >= 0 {
				slice.Endpoints = []discoveryv1.Endpoint{endpoints[step.index]}
			}
			require.NoError(t, s.Client.Update(t.Context(), slice))
			expect(step.backend, step.code)
			if step.fast {
				// Both paths react to the same EndpointSlice update, and with a small
				// configuration full translation is quick. If the full result is
				// published first, the fast update is skipped as already applied
				// instead of published. Either outcome shows the fast path handled
				// the update, so count both. Other tests can also move these counters.
				require.NoError(t, wait.PollUntilContextTimeout(t.Context(), 3*time.Second, 2*time.Minute, true,
					func(context.Context) (bool, error) { return fastPathHandled(t, s) > before, nil }),
					"the fast path did not publish or skip the endpoint update")
			}
			if step.full {
				require.NoError(t, wait.PollUntilContextTimeout(t.Context(), 3*time.Second, 2*time.Minute, true,
					func(context.Context) (bool, error) { return fullPublications(t, s) > fullBefore, nil }),
					"the full path did not publish the endpoint update")
			}
		}
	},
}

// fastPathHandled returns how many endpoint updates the fast path has handled, as
// scraped by Prometheus: snapshots it published, plus updates it skipped because
// the full path had already published them. Missing counters count as 0.
func fastPathHandled(t *testing.T, s *suite.ConformanceTestSuite) float64 {
	return queryCount(t, s, `(sum(xds_snapshot_publications_total{path="fast"}) or vector(0)) + `+
		`(sum(xds_endpoint_refresh_skipped_total{path="fast",reason="already_applied"}) or vector(0))`)
}

// fullPublications returns how many snapshots the full path has published.
func fullPublications(t *testing.T, s *suite.ConformanceTestSuite) float64 {
	return queryCount(t, s, `sum(xds_snapshot_publications_total{path="full"}) or vector(0)`)
}

// queryCount runs a Prometheus query that returns one number. It returns 0 if
// the query fails or has no result.
func queryCount(t *testing.T, s *suite.ConformanceTestSuite, query string) float64 {
	t.Helper()
	v, err := prometheus.QueryPrometheus(s.Client, query)
	if err != nil {
		tlog.Logf(t, "failed to query Prometheus: %v", err)
		return 0
	}
	if vec, ok := v.(model.Vector); ok && len(vec) > 0 {
		return float64(vec[0].Value)
	}
	return 0
}
