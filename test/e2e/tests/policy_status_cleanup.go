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
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, PolicyStatusCleanupSameGatewayClass, PolicyStatusCleanupMultipleGatewayClasses)
}

var PolicyStatusCleanupSameGatewayClass = suite.ConformanceTest{
	ShortName:   "PolicyStatusCleanupSameGatewayClass",
	Description: "Testing Policy Status Cleanup With Ancestors of The Same GatewayClass",
	Manifests:   []string{"testdata/policy-status-cleanup-multiple-ancestors-same-gwc.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("PolicyStatusCleanup", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			gw1NN, gw2NN := types.NamespacedName{Name: "same-namespace", Namespace: ns}, types.NamespacedName{Name: "all-namespaces", Namespace: ns}

			// Check the policy has two ancestors in its status.
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "backendtrafficpolicy-multiple-ancestors-same-gwc", Namespace: ns},
				suite.ControllerName,
				true,
				gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gw1NN.Namespace),
					Name:      gwapiv1.ObjectName(gw1NN.Name),
				},
				gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gw2NN.Namespace),
					Name:      gwapiv1.ObjectName(gw2NN.Name),
				},
			)

			// Change the policy to have a single ancestor, and check its status.
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/policy-status-cleanup-single-ancestor.yaml", false)
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "backendtrafficpolicy-multiple-ancestors-same-gwc", Namespace: ns},
				suite.ControllerName,
				true,
				gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gw1NN.Namespace),
					Name:      gwapiv1.ObjectName(gw1NN.Name),
				},
			)

			// Change the policy to have a single ancestor (that doesn't exist/belong to us), and check its status.
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/policy-status-cleanup-no-ancestor.yaml", false)
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "backendtrafficpolicy-multiple-ancestors-same-gwc", Namespace: ns},
				suite.ControllerName,
				true,
			)
		})
	},
}

var PolicyStatusCleanupMultipleGatewayClasses = suite.ConformanceTest{
	ShortName:   "PolicyStatusCleanupMultipleGatewayClasses",
	Description: "Testing Policy Status Cleanup With Ancestors of Multiple GatewayClasses",
	Manifests:   []string{"testdata/policy-status-cleanup-multiple-ancestors-multiple-gwcs.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("PolicyStatusCleanup", func(t *testing.T) {
			// Create the second gateway of a different gatewayclass, which the backendtrafficpolicy is already attached to.
			prevGwc := suite.Applier.GatewayClass
			suite.Applier.GatewayClass = "status-cleanup"
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/status-cleanup-gateway-different-gwc.yaml", true)
			suite.Applier.GatewayClass = prevGwc

			ns := "gateway-conformance-infra"
			gw1NN, gw2NN := types.NamespacedName{Name: "same-namespace", Namespace: ns}, types.NamespacedName{Name: "gateway-2", Namespace: ns}

			// Check the policy has two ancestors in its status.
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "backendtrafficpolicy-multiple-ancestors-same-gwc", Namespace: ns},
				suite.ControllerName,
				false,
				gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gw1NN.Namespace),
					Name:      gwapiv1.ObjectName(gw1NN.Name),
				},
				gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gw2NN.Namespace),
					Name:      gwapiv1.ObjectName(gw2NN.Name),
				},
			)

			// Change the policy to have a single ancestor, and check its status.
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/policy-status-cleanup-single-ancestor.yaml", false)
			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "backendtrafficpolicy-multiple-ancestors-same-gwc", Namespace: ns},
				suite.ControllerName,
				false,
				gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gw1NN.Namespace),
					Name:      gwapiv1.ObjectName(gw1NN.Name),
				},
			)
		})
	},
}
