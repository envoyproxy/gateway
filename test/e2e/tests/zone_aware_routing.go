// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	// ConformanceTests = append(ConformanceTests, ZoneAwareRoutingTest)
}

var ZoneAwareRoutingTest = suite.ConformanceTest{
	ShortName:   "ZoneAwareRouting",
	Description: "Test Zone Aware Routing is working",
	Manifests: []string{
		"testdata/zone-aware-routing-backendref-enabled.yaml",
		"testdata/zone-aware-routing-btp-enabled.yaml",
		"testdata/zone-aware-routing-deployments.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("topology aware routing - only local zone should get requests", func(t *testing.T) {
			// Pods from the backend-local deployment have affinity
			// for the Envoy Proxy pods so should receive all requests.
			expected := map[string]int{
				"zone-aware-backend-local":    sendRequests,
				"zone-aware-backend-nonlocal": 0,
			}
			runWeightedBackendTest(t, suite, "topology-aware-routing", "/topology-aware-routing", "zone-aware-backend", expected)
		})
		t.Run("BackendTrafficPolicy - only local zone should get requests", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "btp-zone-aware", Namespace: "gateway-conformance-infra"},
				suite.ControllerName,
				gwapiv1a2.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr("gateway-conformance-infra"),
					Name:      gwapiv1.ObjectName("same-namespace"),
				},
			)

			// Pods from the backend-local deployment have affinity
			// for the Envoy Proxy pods so should receive all requests.
			expected := map[string]int{
				"zone-aware-backend-local":    sendRequests,
				"zone-aware-backend-nonlocal": 0,
			}
			runWeightedBackendTest(t, suite, "btp-zone-aware", "/btp-zone-aware", "zone-aware-backend", expected)
		})
	},
}
