// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, MergeBackendsTest)
}

var MergeBackendsTest = suite.ConformanceTest{
	ShortName: "MergeBackends",
	Description: "With EnvoyProxy.spec.mergeBackends enabled, two routes referencing the same " +
		"backend both route correctly, and Envoy generates a single deduplicated Cluster for that " +
		"backend instead of one Cluster per route.",
	Manifests: []string{"testdata/merge-backends.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{Name: "merge-backends-gtw", Namespace: ConformanceInfraNamespace}
		routeANN := types.NamespacedName{Name: "merge-backends-route-a", Namespace: ConformanceInfraNamespace}
		routeBNN := types.NamespacedName{Name: "merge-backends-route-b", Namespace: ConformanceInfraNamespace}
		routeCNN := types.NamespacedName{Name: "merge-backends-route-c", Namespace: ConformanceInfraNamespace}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeANN, routeBNN, routeCNN)

		t.Run("all three routes reach the backend", func(t *testing.T) {
			for _, path := range []string{"/merge-backends-a", "/merge-backends-b", "/merge-backends-c"} {
				expected := http.ExpectedResponse{
					Request:   http.Request{Path: path},
					Response:  http.Response{StatusCodes: []int{200}},
					Namespace: ConformanceInfraNamespace,
				}
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
			}
		})

		t.Run("routes A and B share one merged Cluster, route C keeps its own demerged Cluster", func(t *testing.T) {
			names, err := envoyClusterNames(t, suite, gwNN)
			if err != nil {
				t.Fatalf("failed to fetch Envoy cluster names: %v", err)
			}

			const (
				mergedCluster   = "service/gateway-conformance-infra/infra-backend-v1/8080/http"
				demergedCluster = "httproute/gateway-conformance-infra/merge-backends-route-c/rule/0"
			)
			var hasMerged, hasDemerged bool
			for _, name := range names {
				switch name {
				case mergedCluster:
					hasMerged = true
				case demergedCluster:
					hasDemerged = true
				}
			}
			if !hasMerged {
				t.Errorf("expected a merged Cluster %q for routes A and B, not found (all clusters: %v)", mergedCluster, names)
			}
			if !hasDemerged {
				t.Errorf("expected route C to keep its own demerged Cluster %q, not found (all clusters: %v)", demergedCluster, names)
			}
		})
	},
}

// envoyClusterNames returns the distinct Cluster names configured on gwNN's Envoy proxy.
func envoyClusterNames(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName) ([]string, error) {
	t.Helper()

	body, err := fetchEnvoyClustersOutput(t, suite,
		"app.kubernetes.io/name=envoy",
		"gateway.envoyproxy.io/owning-gateway-name="+gwNN.Name,
		"gateway.envoyproxy.io/owning-gateway-namespace="+gwNN.Namespace,
	)
	if err != nil {
		return nil, err
	}

	// /clusters returns one stat per line as "<cluster_name>::stat_path::<value>", e.g.:
	//   service/gateway-conformance-infra/infra-backend-v1/8080/http::observability_name::backend/...
	//   service/gateway-conformance-infra/infra-backend-v1/8080/http::default_priority::max_connections::1024
	// Many lines share the same cluster name, so keep only the first "::"-delimited field, deduped.
	seen := map[string]bool{}
	var names []string
	for _, line := range strings.Split(body, "\n") {
		name, _, ok := strings.Cut(line, "::")
		if !ok || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names, nil
}
