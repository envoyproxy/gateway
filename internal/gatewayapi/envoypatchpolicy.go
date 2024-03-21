// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

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
		var (
			policy       = policy.DeepCopy()
			ancestorRefs []gwv1a2.ParentReference
			resolveErr   *status.PolicyResolveError
			targetKind   string
			irKey        string
		)

		targetNs := policy.Spec.TargetRef.Namespace
		// If empty, default to namespace of policy
		if targetNs == nil {
			targetNs = ptr.To(gwv1.Namespace(policy.Namespace))
		}

		if t.MergeGateways {
			targetKind = KindGatewayClass
			irKey = string(t.GatewayClassName)

			ancestorRefs = []gwv1a2.ParentReference{
				{
					Group: GroupPtr(gwv1.GroupName),
					Kind:  KindPtr(targetKind),
					Name:  policy.Spec.TargetRef.Name,
				},
			}
		} else {
			targetKind = KindGateway
			gatewayNN := types.NamespacedName{
				Namespace: string(*targetNs),
				Name:      string(policy.Spec.TargetRef.Name),
			}
			// It must exist since the gateways have already been processed
			irKey = irStringKey(gatewayNN.Namespace, gatewayNN.Name)

			ancestorRefs = []gwv1a2.ParentReference{
				getAncestorRefForPolicy(gatewayNN, nil),
			}
		}

		gwXdsIR, ok := xdsIR[irKey]
		if !ok {
			continue
		}

		// Create the IR with the context need to publish the status later
		policyIR := ir.EnvoyPatchPolicy{}
		policyIR.Name = policy.Name
		policyIR.Namespace = policy.Namespace
		policyIR.Status = &policy.Status

		// Append the IR
		gwXdsIR.EnvoyPatchPolicies = append(gwXdsIR.EnvoyPatchPolicies, &policyIR)

		// Ensure EnvoyPatchPolicy is enabled
		if !t.EnvoyPatchPolicyEnabled {
			resolveErr = &status.PolicyResolveError{
				Reason:  egv1a1.PolicyReasonDisabled,
				Message: "EnvoyPatchPolicy is disabled in the EnvoyGateway configuration",
			}
			status.SetResolveErrorForPolicyAncestors(&policy.Status,
				ancestorRefs,
				t.GatewayControllerName,
				policy.Generation,
				resolveErr,
			)

			continue
		}

		// Ensure EnvoyPatchPolicy is targeting to a support type
		if policy.Spec.TargetRef.Group != gwv1.GroupName || string(policy.Spec.TargetRef.Kind) != targetKind {
			message := fmt.Sprintf("TargetRef.Group:%s TargetRef.Kind:%s, only TargetRef.Group:%s and TargetRef.Kind:%s is supported.",
				policy.Spec.TargetRef.Group, policy.Spec.TargetRef.Kind, gwv1.GroupName, targetKind)

			resolveErr = &status.PolicyResolveError{
				Reason:  gwv1a2.PolicyReasonInvalid,
				Message: message,
			}
			status.SetResolveErrorForPolicyAncestors(&policy.Status,
				ancestorRefs,
				t.GatewayControllerName,
				policy.Generation,
				resolveErr,
			)

			continue
		}

		// Ensure EnvoyPatchPolicy and target Gateway are in the same namespace
		if policy.Namespace != string(*targetNs) {
			message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, EnvoyPatchPolicy can only target a %s in the same namespace.",
				policy.Namespace, *targetNs, targetKind)

			resolveErr = &status.PolicyResolveError{
				Reason:  gwv1a2.PolicyReasonInvalid,
				Message: message,
			}
			status.SetResolveErrorForPolicyAncestors(&policy.Status,
				ancestorRefs,
				t.GatewayControllerName,
				policy.Generation,
				resolveErr,
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
			irPatch.Operation.From = patch.Operation.From
			irPatch.Operation.Value = patch.Operation.Value

			policyIR.JSONPatches = append(policyIR.JSONPatches, &irPatch)
		}

		// Set Accepted=True
		status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName)
	}
}
