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

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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

			policyNamespacedName := types.NamespacedName{Name: "backendtrafficpolicy-multiple-ancestors-same-gwc", Namespace: ns}

			// Check the policy has two ancestors in its status.
			BackendTrafficPolicyMustBeAcceptedByAllAncestors(t,
				suite.Client,
				policyNamespacedName,
				suite.ControllerName,
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
			BackendTrafficPolicyMustBeAcceptedByAllAncestors(t,
				suite.Client,
				policyNamespacedName,
				suite.ControllerName,
				gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gw1NN.Namespace),
					Name:      gwapiv1.ObjectName(gw1NN.Name),
				},
			)

			// Update the policy status to include a status ancestor of another controller.
			policy := &egv1a1.BackendTrafficPolicy{}
			err := suite.Client.Get(context.Background(), policyNamespacedName, policy)
			require.NoErrorf(t, err, "error getting BackendTrafficPolicy %s", policyNamespacedName.String())
			otherControllerAncestor := gwapiv1.PolicyAncestorStatus{
				AncestorRef: gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Name:      "other-controller-ancestor",
					Namespace: gatewayapi.NamespacePtr(policy.Namespace),
				},
				ControllerName: "gateway.envoyproxy.io/other-gatewayclass-controller",
				Conditions: []metav1.Condition{
					{
						Type:               string(gwapiv1.PolicyConditionAccepted),
						Status:             metav1.ConditionTrue,
						LastTransitionTime: metav1.NewTime(time.Now()),
						Reason:             string(gwapiv1.PolicyConditionAccepted),
					},
				},
			}
			policy.Status.Ancestors = append(policy.Status.Ancestors, otherControllerAncestor)
			err = suite.Client.Status().Update(context.Background(), policy)
			require.NoErrorf(t, err, "error updating BackendTrafficPolicy status %s", policyNamespacedName.String())

			// Change the policy spec to have a corresponding ancestor and trigger reconciliation.
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/policy-status-cleanup-no-ancestor.yaml", false)

			// Check its status to only have the ancestor from the other controller.
			BackendTrafficPolicyMustBeAcceptedByAllAncestors(t,
				suite.Client,
				policyNamespacedName,
				"gateway.envoyproxy.io/other-gatewayclass-controller",
				[]gwapiv1.ParentReference{
					{
						Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
						Kind:      gatewayapi.KindPtr(resource.KindGateway),
						Name:      "other-controller-ancestor",
						Namespace: gatewayapi.NamespacePtr(policy.Namespace),
					},
				}...,
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
			BackendTrafficPolicyMustBeAcceptedByAllAncestors(t,
				suite.Client,
				types.NamespacedName{Name: "backendtrafficpolicy-multiple-ancestors-same-gwc", Namespace: ns},
				suite.ControllerName,
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
			BackendTrafficPolicyMustBeAcceptedByAllAncestors(t,
				suite.Client,
				types.NamespacedName{Name: "backendtrafficpolicy-multiple-ancestors-same-gwc", Namespace: ns},
				suite.ControllerName,
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
