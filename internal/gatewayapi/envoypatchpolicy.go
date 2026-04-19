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
		var (
			ancestorRef gwapiv1.ParentReference
			resolveErr  *status.PolicyResolveError
			targetKind  string
			irKey       string
		)

		refKind, refName := policy.Spec.TargetRef.Kind, policy.Spec.TargetRef.Name
		if t.MergeGateways {
			targetKind = resource.KindGatewayClass
			// if ref GatewayClass name is not same as t.GatewayClassName, it will be skipped below.
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

		// Ensure EnvoyPatchPolicy is targeting to a support type
		if policy.Spec.TargetRef.Group != gwapiv1.GroupName || string(refKind) != targetKind {
			message := fmt.Sprintf("TargetRef.Group:%s TargetRef.Kind:%s, only TargetRef.Group:%s and TargetRef.Kind:%s is supported.",
				policy.Spec.TargetRef.Group, policy.Spec.TargetRef.Kind, gwapiv1.GroupName, targetKind)

			resolveErr = &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonInvalid,
				Message: message,
			}

			// For invalid targetRef kind, we need to find the correct xdsIR key to attach status
			// In MergeGateways mode, use GatewayClass name; in default mode, use Gateway namespace/name
			var statusIRKey string
			var statusAncestorRef gwapiv1.ParentReference
			if t.MergeGateways {
				statusIRKey = string(t.GatewayClassName)
				statusAncestorRef = gwapiv1.ParentReference{
					Group: GroupPtr(gwapiv1.GroupName),
					Kind:  KindPtr(resource.KindGatewayClass),
					Name:  t.GatewayClassName,
				}
			} else {
				gatewayNN := types.NamespacedName{
					Namespace: policy.Namespace,
					Name:      string(refName),
				}
				statusIRKey = t.IRKey(gatewayNN)
				statusAncestorRef = getAncestorRefForPolicy(gatewayNN, nil)
			}

			// Try to attach status to the correct xdsIR if it exists
			if gwXdsIR, ok := xdsIR[statusIRKey]; ok {
				policyIR := ir.EnvoyPatchPolicy{}
				policyIR.Name = policy.Name
				policyIR.Namespace = policy.Namespace
				policyIR.Generation = policy.Generation
				policyIR.Status = &policy.Status
				gwXdsIR.EnvoyPatchPolicies = append(gwXdsIR.EnvoyPatchPolicies, &policyIR)
			}

			status.SetResolveErrorForPolicyAncestor(&policy.Status,
				&statusAncestorRef,
				t.GatewayControllerName,
				policy.Generation,
				resolveErr,
			)

			continue
		}

		gwXdsIR, ok := xdsIR[irKey]
		if !ok {
			// The TargetRef Gateway is not an accepted Gateway, then skip processing.
			continue
		}

		// Create the IR with the context need to publish the status later
		policyIR := ir.EnvoyPatchPolicy{}
		policyIR.Name = policy.Name
		policyIR.Namespace = policy.Namespace
		policyIR.Generation = policy.Generation
		policyIR.Status = &policy.Status

		// Append the IR
		gwXdsIR.EnvoyPatchPolicies = append(gwXdsIR.EnvoyPatchPolicies, &policyIR)

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
	}
}
