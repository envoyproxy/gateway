// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, ZoneAwareRoutingTest)
}

var ZoneAwareRoutingTest = suite.ConformanceTest{
	ShortName:   "ZoneAwareRouting",
	Description: "Test Zone Aware Routing is working",
	Manifests:   []string{"testdata/zone-aware-routing-backendref-enabled.yaml", "testdata/zone-aware-routing-btp-enabled.yaml", "testdata/zone-aware-routing-deployments.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("topology aware routing - only local zone should get requests", func(t *testing.T) {
			// Pods from the backend-local deployment have affinity
			// for the Envoy Proxy pods so should receive all requests.
			expected := map[string]int{
				"zone-aware-backend-local":    sendRequests,
				"zone-aware-backend-nonlocal": 0,
			}
			runWeightedBackendTest(t, suite, "topology-aware-routing", "/topology-aware-routing", "topology-aware-routing", expected)
		})
		t.Run("BackendTrafficPolicy - only local zone should get requests", func(t *testing.T) {
			// Pods from the backend-local deployment have affinity
			// for the Envoy Proxy pods so should receive all requests.
			expected := map[string]int{
				"zone-aware-backend-local":    sendRequests,
				"zone-aware-backend-nonlocal": 0,
			}
			runWeightedBackendTest(t, suite, "btp-zone-aware", "/btp-zone-aware", "btp-zone-aware", expected)
		})
	},
}
