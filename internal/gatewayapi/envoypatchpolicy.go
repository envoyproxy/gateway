// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/status"
)

func (t *Translator) ProcessEnvoyPatchPolicies(envoyPatchPolicies []*egv1a1.EnvoyPatchPolicy, xdsIR XdsIRMap) {
	// Sort based on priority
	sort.Slice(envoyPatchPolicies, func(i, j int) bool {
		return envoyPatchPolicies[i].Spec.Priority < envoyPatchPolicies[j].Spec.Priority
	})

	for _, policy := range envoyPatchPolicies {
		policy := policy.DeepCopy()
		targetNs := policy.Spec.TargetRef.Namespace
		if targetNs == nil {
			// This status condition will not get updated in the resource because
			// we dont have access to the IR yet, but it has been kept here in case we publish
			// the status from this layer instead of the xds layer.

			status.SetEnvoyPatchPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonInvalid,
				"TargetRef.Namespace must be set",
			)
			continue
		}

		// Get the IR
		// It must exist since the gateways have already been processed
		irKey := irStringKey(string(*targetNs), string(policy.Spec.TargetRef.Name))
		gwXdsIR, ok := xdsIR[irKey]
		if !ok {
			// This status condition will not get updated in the resource because
			// the IR is missing, but it has been kept here in case we publish
			// the status from this layer instead of the xds layer.
			message := fmt.Sprintf("Gateway:%s not found.", policy.Spec.TargetRef.Name)

			status.SetEnvoyPatchPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonTargetNotFound,
				message,
			)
			continue
		}

		// Create the IR with the context need to publish the status later
		policyIR := ir.EnvoyPatchPolicy{}
		policyIR.Name = policy.Name
		policyIR.Namespace = policy.Namespace
		policyIR.Status = &policy.Status

		// Append the IR
		gwXdsIR.EnvoyPatchPolicies = append(gwXdsIR.EnvoyPatchPolicies, &policyIR)

		// Ensure policy can only target a Gateway
		if policy.Spec.TargetRef.Group != gwv1b1.GroupName || policy.Spec.TargetRef.Kind != KindGateway {
			message := fmt.Sprintf("TargetRef.Group:%s TargetRef.Kind:%s, only TargetRef.Group:%s and TargetRef.Kind:%s is supported.",
				policy.Spec.TargetRef.Group, policy.Spec.TargetRef.Kind, gwv1b1.GroupName, KindGateway)

			status.SetEnvoyPatchPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonInvalid,
				message,
			)
			continue
		}

		// Ensure Policy and target Gateway are in the same namespace
		if policy.Namespace != string(*targetNs) {
			message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, EnvoyPatchPolicy can only target a Gateway in the same namespace.",
				policy.Namespace, *targetNs)

			status.SetEnvoyPatchPolicyCondition(policy,
				gwv1a2.PolicyConditionAccepted,
				metav1.ConditionFalse,
				gwv1a2.PolicyReasonInvalid,
				message,
			)
			continue
		}

		// Save the patch
		for _, patch := range policy.Spec.JSONPatches {
			irPatch := ir.JSONPatchConfig{}
			irPatch.Type = string(patch.Type)
			irPatch.Name = patch.Name
			irPatch.Operation.Op = string(patch.Operation.Op)
			irPatch.Operation.Path = patch.Operation.Path
			irPatch.Operation.Value = patch.Operation.Value

			policyIR.JSONPatches = append(policyIR.JSONPatches, &irPatch)
		}

		// Set Accepted=True
		status.SetEnvoyPatchPolicyCondition(policy,
			gwv1a2.PolicyConditionAccepted,
			metav1.ConditionTrue,
			gwv1a2.PolicyReasonAccepted,
			"EnvoyPatchPolicy has been accepted.",
		)
	}
}
