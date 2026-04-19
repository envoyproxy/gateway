// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestProcessEnvoyPatchPolicies_NonMergedInvalidAndValidSameNameDoNotCollide(t *testing.T) {
	translator := &Translator{
		MergeGateways:           false,
		EnvoyPatchPolicyEnabled: true,
		GatewayControllerName:   "envoy-gateway-test-controller",
	}

	gatewayNN := types.NamespacedName{Namespace: "default", Name: "shared-gw"}
	xdsIR := resource.XdsIRMap{
		translator.IRKey(gatewayNN): &ir.Xds{},
	}

	policy := &egv1a1.EnvoyPatchPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "epp",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: egv1a1.EnvoyPatchPolicySpec{
			Type: egv1a1.JSONPatchEnvoyPatchType,
			TargetRefs: []gwapiv1.LocalPolicyTargetReference{
				// Invalid kind with the same name as the valid Gateway target.
				{Group: gwapiv1.GroupName, Kind: gwapiv1.Kind(resource.KindGatewayClass), Name: gwapiv1.ObjectName("shared-gw")},
				{Group: gwapiv1.GroupName, Kind: gwapiv1.Kind(resource.KindGateway), Name: gwapiv1.ObjectName("shared-gw")},
			},
		},
	}

	translator.ProcessEnvoyPatchPolicies([]*egv1a1.EnvoyPatchPolicy{policy}, xdsIR)

	require.Len(t, policy.Status.Ancestors, 2)
	// Only the valid gateway target should produce an attachable IR entry.
	require.Len(t, xdsIR[translator.IRKey(gatewayNN)].EnvoyPatchPolicies, 1)

	var invalidAncestor, validAncestor *gwapiv1.PolicyAncestorStatus
	for i := range policy.Status.Ancestors {
		ancestor := &policy.Status.Ancestors[i]
		switch {
		case ancestor.AncestorRef.Kind != nil && *ancestor.AncestorRef.Kind == resource.KindGatewayClass:
			invalidAncestor = ancestor
		case ancestor.AncestorRef.Kind != nil && *ancestor.AncestorRef.Kind == resource.KindGateway:
			validAncestor = ancestor
		}
	}

	require.NotNil(t, invalidAncestor)
	require.NotNil(t, validAncestor)

	invalidAccepted := meta.FindStatusCondition(invalidAncestor.Conditions, string(gwapiv1.PolicyConditionAccepted))
	validAccepted := meta.FindStatusCondition(validAncestor.Conditions, string(gwapiv1.PolicyConditionAccepted))
	require.NotNil(t, invalidAccepted)
	require.NotNil(t, validAccepted)

	assert.Equal(t, metav1.ConditionFalse, invalidAccepted.Status)
	assert.Equal(t, metav1.ConditionTrue, validAccepted.Status)
	// Valid Gateway ancestor should remain normalized with namespace in non-merged mode.
	require.NotNil(t, validAncestor.AncestorRef.Namespace)
	assert.Equal(t, gwapiv1.Namespace("default"), *validAncestor.AncestorRef.Namespace)
}

func TestGetAncestorRefForEnvoyPatchPolicyTargetRef_UsesTargetFields(t *testing.T) {
	ref := gwapiv1.LocalPolicyTargetReference{
		Group: gwapiv1.Group("example.io"),
		Kind:  gwapiv1.Kind("ExampleKind"),
		Name:  gwapiv1.ObjectName("example-name"),
	}

	ancestor := getAncestorRefForEnvoyPatchPolicyTargetRef(ref)
	require.NotNil(t, ancestor.Group)
	require.NotNil(t, ancestor.Kind)
	assert.Equal(t, gwapiv1.Group("example.io"), *ancestor.Group)
	assert.Equal(t, gwapiv1.Kind("ExampleKind"), *ancestor.Kind)
	assert.Equal(t, gwapiv1.ObjectName("example-name"), ancestor.Name)
}
