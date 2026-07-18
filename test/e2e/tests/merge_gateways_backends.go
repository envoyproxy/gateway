// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"strconv"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	MergeGatewaysTests = append(MergeGatewaysTests, MergeGatewaysBackendsTest)
}

var MergeGatewaysBackendsTest = suite.ConformanceTest{
	ShortName: "MergeGatewaysBackends",
	Description: "With both MergeGateways and MergeBackends enabled, two Gateways with different " +
		"gateway-scoped BackendTrafficPolicy CircuitBreaker settings both route to the same backend; " +
		"each Gateway must get its own Cluster with its own settings, not share one across Gateways.",
	Manifests: []string{"testdata/merge-gateways-merge-backends.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gw1NN := types.NamespacedName{Name: "merge-gateways-backends-1", Namespace: ns}
		gw2NN := types.NamespacedName{Name: "merge-gateways-backends-2", Namespace: ns}
		route1NN := types.NamespacedName{Name: "merge-gateways-backends-route-1", Namespace: ns}
		route2NN := types.NamespacedName{Name: "merge-gateways-backends-route-2", Namespace: ns}

		gw1Addr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gw1NN), route1NN)
		gw2Addr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gw2NN), route2NN)

		t.Run("both routes reach the shared backend", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gw1Addr, http.ExpectedResponse{
				Request:   http.Request{Path: "/merge-gateways-backends-1", Host: "www.merge-gateways-backends-1.com"},
				Response:  http.Response{StatusCodes: []int{200}},
				Namespace: ns,
			})
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gw2Addr, http.ExpectedResponse{
				Request:   http.Request{Path: "/merge-gateways-backends-2", Host: "www.merge-gateways-backends-2.com"},
				Response:  http.Response{StatusCodes: []int{200}},
				Namespace: ns,
			})
		})

		t.Run("each Gateway keeps its own Cluster and CircuitBreaker", func(t *testing.T) {
			maxConnByCluster, err := mergeGatewaysMaxConnectionsByCluster(t, suite)
			if err != nil {
				t.Fatalf("failed to fetch Envoy cluster stats: %v", err)
			}

			// The cluster name embeds its owning Gateway's "namespace/name" identity (see
			// irBackendClusterName), so match each Gateway to its own Cluster directly instead of
			// just checking the set of maxConnections values.
			for gwName, want := range map[string]int{
				"merge-gateways-backends-1": 1111,
				"merge-gateways-backends-2": 2222,
			} {
				var matched []string
				for name := range maxConnByCluster {
					if strings.Contains(name, "infra-backend-v1") && strings.Contains(name, gwName) {
						matched = append(matched, name)
					}
				}
				if len(matched) != 1 {
					t.Fatalf("expected exactly 1 Cluster for backend infra-backend-v1 owned by Gateway %s, got %d: %v", gwName, len(matched), matched)
				}
				if got := maxConnByCluster[matched[0]]; got != want {
					t.Errorf("Cluster %s: expected maxConnections=%d, got %d", matched[0], want, got)
				}
			}
		})
	},
}

// mergeGatewaysMaxConnectionsByCluster returns, for each Cluster on the merge-gateways-backends
// Envoy proxy, its circuit breaker max_connections value.
func mergeGatewaysMaxConnectionsByCluster(t *testing.T, suite *suite.ConformanceTestSuite) (map[string]int, error) {
	t.Helper()

	body, err := fetchEnvoyClustersOutput(t, suite,
		"app.kubernetes.io/name=envoy",
		"gateway.envoyproxy.io/owning-gatewayclass=merge-gateways-backends",
	)
	if err != nil {
		return nil, err
	}

	// /clusters returns one stat per line as "<cluster_name>::<stat_path>::<value>", e.g.:
	//   backend/service/.../infra-backend-v1/8080/http/...::default_priority::max_connections::1111
	result := make(map[string]int)
	for _, line := range strings.Split(body, "\n") {
		name, rest, ok := strings.Cut(line, "::default_priority::max_connections::")
		if !ok {
			continue
		}
		value, err := strconv.Atoi(rest)
		if err != nil {
			continue
		}
		result[name] = value
	}
	return result, nil
}
