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
			irKey       string
		)

		refKind, refName := policy.Spec.TargetRef.Kind, policy.Spec.TargetRef.Name

		// Determine the expected target kind and IR key based on mode
		if t.MergeGateways {
			// In MergeGateways mode, only GatewayClass targetRef is supported
			if string(refKind) == resource.KindGateway {
				// Gateway targetRef is not supported in MergeGateways mode
				ancestorRef = gwapiv1.ParentReference{
					Group:     GroupPtr(gwapiv1.GroupName),
					Kind:      KindPtr(resource.KindGateway),
					Name:      refName,
					Namespace: NamespacePtr(policy.Namespace),
				}
				
				message := fmt.Sprintf("TargetRef.Kind:%s is not supported in MergeGateways mode, only TargetRef.Kind:%s is supported.",
					refKind, resource.KindGatewayClass)
				resolveErr = &status.PolicyResolveError{
					Reason:  gwapiv1.PolicyReasonInvalid,
					Message: message,
				}
				
				// Create a minimal IR entry to publish the error status
				policyIR := ir.EnvoyPatchPolicy{}
				policyIR.Name = policy.Name
				policyIR.Namespace = policy.Namespace
				policyIR.Generation = policy.Generation
				policyIR.Status = &policy.Status
				
				status.SetResolveErrorForPolicyAncestor(&policy.Status,
					&ancestorRef,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)
				
				// We still need to add this to an IR so the status gets published
				// Use the GatewayClassName as the key since that's the merged IR key
				irKey = string(t.GatewayClassName)
				if gwXdsIR, ok := xdsIR[irKey]; ok {
					gwXdsIR.EnvoyPatchPolicies = append(gwXdsIR.EnvoyPatchPolicies, &policyIR)
				}
				
				continue
			}
			
			// GatewayClass targetRef
			irKey = string(refName)
			ancestorRef = gwapiv1.ParentReference{
				Group: GroupPtr(gwapiv1.GroupName),
				Kind:  KindPtr(resource.KindGatewayClass),
				Name:  refName,
			}
		} else {
			// In default mode, only Gateway targetRef is supported
			if string(refKind) != resource.KindGateway {
				// Non-Gateway targetRef is not supported in default mode
				gatewayNN := types.NamespacedName{
					Namespace: policy.Namespace,
					Name:      string(refName),
				}
				// GatewayClass is cluster-scoped; other kinds are namespace-scoped
				var nsPtr *gwapiv1.Namespace
				if string(refKind) != resource.KindGatewayClass {
					nsPtr = NamespacePtr(policy.Namespace)
				}
				ancestorRef = gwapiv1.ParentReference{
					Group:     GroupPtr(gwapiv1.GroupName),
					Kind:      KindPtr(gwapiv1.Kind(refKind)),
					Name:      refName,
					Namespace: nsPtr,
				}

				message := fmt.Sprintf("TargetRef.Kind:%s is not supported in default mode, only TargetRef.Kind:%s is supported.",
					refKind, resource.KindGateway)
				resolveErr = &status.PolicyResolveError{
					Reason:  gwapiv1.PolicyReasonInvalid,
					Message: message,
				}

				// Create a minimal IR entry to publish the error status
				policyIR := ir.EnvoyPatchPolicy{}
				policyIR.Name = policy.Name
				policyIR.Namespace = policy.Namespace
				policyIR.Generation = policy.Generation
				policyIR.Status = &policy.Status

				status.SetResolveErrorForPolicyAncestor(&policy.Status,
					&ancestorRef,
					t.GatewayControllerName,
					policy.Generation,
					resolveErr,
				)

				// Try to find a matching Gateway IR to attach the status
				irKey = t.IRKey(gatewayNN)
				if gwXdsIR, ok := xdsIR[irKey]; ok {
					gwXdsIR.EnvoyPatchPolicies = append(gwXdsIR.EnvoyPatchPolicies, &policyIR)
				}

				continue
			}

			// Gateway targetRef
			gatewayNN := types.NamespacedName{
				Namespace: policy.Namespace,
				Name:      string(refName),
			}
			irKey = t.IRKey(gatewayNN)
			ancestorRef = getAncestorRefForPolicy(gatewayNN, nil)
		}

		gwXdsIR, ok := xdsIR[irKey]
		if !ok {
			// The TargetRef resource is not found or not accepted, skip processing without status.
			// Status will not be published for policies targeting non-existent or non-accepted resources.
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

		// Validate targetRef group
		if policy.Spec.TargetRef.Group != gwapiv1.GroupName {
			message := fmt.Sprintf("TargetRef.Group:%s is not supported, only TargetRef.Group:%s is supported.",
				policy.Spec.TargetRef.Group, gwapiv1.GroupName)

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
