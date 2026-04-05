// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"math"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, WeightedZonesTest)
}

var WeightedZonesTest = suite.ConformanceTest{
	ShortName:   "WeightedZones",
	Description: "Test weighted zone traffic distribution via BackendTrafficPolicy",
	Manifests:   []string{"testdata/weighted-zones.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		WaitForPods(t, suite.Client, ns, map[string]string{"app": "weighted-zones-backend"}, corev1.PodRunning, &PodReady)

		BackendTrafficPolicyMustBeAccepted(t,
			suite.Client,
			types.NamespacedName{Name: "weighted-zones-policy", Namespace: ns},
			suite.ControllerName,
			gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(ns),
				Name:      gwapiv1.ObjectName("same-namespace"),
			},
		)

		t.Run("traffic should be distributed according to zone weights", func(t *testing.T) {
			// BTP configures zone "1" with weight 80 and zone "2" with weight 20.
			// Deployments pin one pod per zone via nodeSelector.
			expected := map[string]int{
				"weighted-zones-backend-zone1": int(math.Round(sendRequests * 0.80)),
				"weighted-zones-backend-zone2": int(math.Round(sendRequests * 0.20)),
			}
			runWeightedBackendTest(t, suite, nil, "weighted-zones-route", "/weighted-zones", "weighted-zones-backend", expected)
		})
	},
}
