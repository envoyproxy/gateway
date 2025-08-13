// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"math"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, ZoneAwareRoutingTest)
}

var ZoneAwareRoutingTest = suite.ConformanceTest{
	ShortName:   "ZoneAwareRouting",
	Description: "Test Zone Aware Routing is working",
	Manifests: []string{
		"testdata/zone-aware-routing-backendref-enabled.yaml",
		"testdata/zone-aware-routing-btp-force-local-zone.yaml",
		"testdata/zone-aware-routing-btp-no-force-local-zone.yaml",
		"testdata/zone-aware-routing-deployments.yaml",
		"testdata/zone-aware-routing-gateways.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("topology aware routing - only local zone should get requests", func(t *testing.T) {
			// Pods from the backend-local deployment have affinity
			// for the Envoy Proxy pods so should receive all requests.
			expected := map[string]int{
				"zone-aware-backend-local":    sendRequests,
				"zone-aware-backend-nonlocal": 0,
			}
			runWeightedBackendTest(t, suite, nil, "topology-aware-routing", "/topology-aware-routing", "zone-aware-backend", expected)
		})
		t.Run("BackendTrafficPolicy - ForceLocalZone - only local zone should get requests", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "btp-force-local-zone", Namespace: "gateway-conformance-infra"},
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
			runWeightedBackendTest(t, suite, nil, "btp-force-local-zone", "/btp-force-local-zone", "zone-aware-backend", expected)
		})
		t.Run("BackendTrafficPolicy - No ForceLocalZone - local zone should get around 75% of requests", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "btp-no-force-local-zone", Namespace: "gateway-conformance-infra"},
				suite.ControllerName,
				gwapiv1a2.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr("gateway-conformance-infra"),
					Name:      gwapiv1.ObjectName("same-namespace"),
				},
			)

			// Pods from the backend-local deployment have affinity for the Envoy Proxy pods.
			// ForceLocal is not used, and overall we have 4 backend pods, 3 local and 1 non-local.
			// Distribution of upstream is 75% local, 25% non-local. Distribution of envoy is 100% local.
			// Expect local upstream to get around 75% of traffic.
			expected := map[string]int{
				"zone-aware-backend-local":    int(math.Round(sendRequests * .75)),
				"zone-aware-backend-nonlocal": int(math.Round(sendRequests * .25)),
			}
			runWeightedBackendTest(t, suite, nil, "btp-no-force-local-zone", "/btp-no-force-local-zone", "zone-aware-backend", expected)
		})
		t.Run("BackendTrafficPolicy - No ForceLocalZone - Hardcoded service name in EnvoyProxy - local zone should get around 75% of requests", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "btp-no-force-local-zone-hardcoded-svc-name", Namespace: "gateway-conformance-infra"},
				suite.ControllerName,
				gwapiv1a2.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr("gateway-conformance-infra"),
					Name:      gwapiv1.ObjectName("zone-aware-routing-gtw"),
				},
			)

			// Pods from the backend-local deployment have affinity for the Envoy Proxy pods.
			// ForceLocal is not used, and overall we have 4 backend pods, 3 local and 1 non-local.
			// Distribution of upstream is 75% local, 25% non-local. Distribution of envoy is 100% local.
			// Expect local upstream to get around 75% of traffic.
			expected := map[string]int{
				"zone-aware-backend-local":    int(math.Round(sendRequests * .75)),
				"zone-aware-backend-nonlocal": int(math.Round(sendRequests * .25)),
			}
			runWeightedBackendTest(t, suite, &types.NamespacedName{Name: "zone-aware-routing-gtw", Namespace: "gateway-conformance-infra"}, "btp-no-force-local-zone-hardcoded-svc-name", "/btp-no-force-local-zone-hardcoded-svc-name", "zone-aware-backend", expected)
		})
	},
}
