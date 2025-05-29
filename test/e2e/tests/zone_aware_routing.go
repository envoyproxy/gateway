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
	Description: "Resource with Zone Aware Routing enabled",
	Manifests:   []string{"testdata/zone-aware-routing.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("only local zone should get requests", func(t *testing.T) {
			// Pods from the backend-local deployment have affinity
			// for the Envoy Proxy pods so should receive all requests.
			expected := map[string]int{
				"zone-aware-backend-local":    sendRequests,
				"zone-aware-backend-nonlocal": 0,
			}
			runWeightedBackendTest(t, suite, "zone-aware-http-route", "/", "zone-aware-backend", expected)
		})
	},
}
