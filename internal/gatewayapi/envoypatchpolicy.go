// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) ProcessEnvoyPatchPolicies(envoyPatchPolicies []*egv1a1.EnvoyPatchPolicy, xdsIR resource.XdsIRMap) {
	// EnvoyPatchPolicies are already sorted by the provider layer (priority, then timestamp, then name)
	for _, policy := range envoyPatchPolicies {
		targetRefs := getTargetRefsForEPP(policy)
		createdAnyPolicyIR := false

		for _, targetRef := range targetRefs {
			var (
				ancestorRef gwapiv1.ParentReference
				resolveErr  *status.PolicyResolveError
				targetKind  string
				irKey       string
				gwXdsIR     *ir.Xds
				ok          bool
			)

			refKind, refName := targetRef.Kind, targetRef.Name
			if t.MergeGateways {
				targetKind = resource.KindGatewayClass
				// if ref GatewayClass name is not same as t.GatewayClassName, it will be skipped when checking XDS IR.
				irKey = string(refName)
				ancestorRef = gwapiv1.ParentReference{
					Group: GroupPtr(gwapiv1.GroupName),
					Kind:  KindPtr(targetKind),
					Name:  refName,
				}
			} else {
				targetKind = resource.KindGateway
				gatewayNN := types.NamespacedName{
					Namespace: policy.Namespace,
					Name:      string(refName),
				}
				irKey = t.IRKey(gatewayNN)
				ancestorRef = getAncestorRefForPolicy(gatewayNN, nil)
			}

			// Ensure EnvoyPatchPolicy is enabled
			if !t.EnvoyPatchPolicyEnabled {
				resolveErr = &status.PolicyResolveError{
					Reason:  egv1a1.PolicyReasonDisabled,
					Message: "EnvoyPatchPolicy is disabled in the EnvoyGateway configuration",
				}
				status.SetResolveErrorForPolicyAncestor(&policy.Status,
					&ancestorRef,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			// Ensure EnvoyPatchPolicy is targeting to a support type
			if targetRef.Group != gwapiv1.GroupName || string(refKind) != targetKind {
				message := fmt.Sprintf("Target to %s/%s, only %s/%s is supported.",
					targetRef.Group, targetRef.Kind, gwapiv1.GroupName, targetKind)

				resolveErr = &status.PolicyResolveError{
					Reason:  gwapiv1.PolicyReasonInvalid,
					Message: message,
				}
				status.SetResolveErrorForPolicyAncestor(&policy.Status,
					&ancestorRef,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				continue
			}

			gwXdsIR, ok = xdsIR[irKey]
			if !ok {
				var message string
				message = fmt.Sprintf("Target to %s %s/%s does not exist.", targetKind, policy.Namespace, refName)
				// if mergeGateways is enabled, the TargetRef should be GatewayClass, otherwise it should be Gateway.
				if string(refKind) != targetKind {
					message = fmt.Sprintf("Target to %s, only %s is supported when MergeGateways is %t.", refKind, targetKind, t.MergeGateways)
				}

				resolveErr = &status.PolicyResolveError{
					Reason:  egv1a1.PolicyReasonInvalid,
					Message: message,
				}
				status.SetResolveErrorForPolicyAncestor(&policy.Status,
					&ancestorRef,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)
				continue
			}

			// Create the IR with the context need to publish the status later
			policyIR := ir.EnvoyPatchPolicy{}
			policyIR.Name = policy.Name
			policyIR.Namespace = policy.Namespace
			policyIR.Generation = policy.Generation
			policyIR.Status = &policy.Status
			policyIR.AncestorRef = &ancestorRef

			// Append the IR
			gwXdsIR.EnvoyPatchPolicies = append(gwXdsIR.EnvoyPatchPolicies, &policyIR)
			createdAnyPolicyIR = true

			// Save the patch
			for _, patch := range policy.Spec.JSONPatches {
				irPatch := ir.JSONPatchConfig{}
				irPatch.Type = string(patch.Type)
				irPatch.Name = patch.Name
				irPatch.Operation.Op = ir.JSONPatchOp(patch.Operation.Op)
				irPatch.Operation.Path = patch.Operation.Path
				irPatch.Operation.JSONPath = patch.Operation.JSONPath
				irPatch.Operation.From = patch.Operation.From
				irPatch.Operation.Value = patch.Operation.Value

				policyIR.JSONPatches = append(policyIR.JSONPatches, &irPatch)
			}

			// Set Accepted=True
			status.SetAcceptedForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation)
			// If there are no valid TargetRefs, add a warning condition to the policy status.
			if len(policy.Spec.TargetRefs) == 0 {
				status.SetDeprecatedFieldsWarningForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation,
					map[string]string{
						"spec.targetRef": "spec.targetRefs",
					})
			}
		}

		// If no policyIR was created (all targets were rejected), but the policy has status
		// (from rejected targets), create a status-only policyIR to ensure status gets published.
		// Attach it to the first available XdsIR.
		if !createdAnyPolicyIR && len(policy.Status.Ancestors) > 0 {
			for _, gwXdsIR := range xdsIR {
				policyIR := ir.EnvoyPatchPolicy{}
				policyIR.Name = policy.Name
				policyIR.Namespace = policy.Namespace
				policyIR.Generation = policy.Generation
				policyIR.Status = &policy.Status
				// No AncestorRef since this is a status-only entry for all rejected targets
				policyIR.AncestorRef = nil

				gwXdsIR.EnvoyPatchPolicies = append(gwXdsIR.EnvoyPatchPolicies, &policyIR)
				break
			}
		}
	}
}

// getTargetRefsForEPP returns the target refs for the given EnvoyPatchPolicy, handling both the deprecated TargetRef and the new TargetRefs fields.
// There's CEL validation to ensure that only one of TargetRef or TargetRefs is set, so we can safely return the non-nil field.
func getTargetRefsForEPP(policy *egv1a1.EnvoyPatchPolicy) []gwapiv1.LocalPolicyTargetReference {
	if policy.Spec.TargetRef != nil {
		return []gwapiv1.LocalPolicyTargetReference{*policy.Spec.TargetRef}
	}
	return policy.Spec.TargetRefs
}
