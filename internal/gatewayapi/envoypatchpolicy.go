// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) ProcessEnvoyPatchPolicies(envoyPatchPolicies []*egv1a1.EnvoyPatchPolicy, xdsIR resource.XdsIRMap) {
	// EnvoyPatchPolicies are already sorted by the provider layer (priority, then timestamp, then name)

	for _, policy := range envoyPatchPolicies {
		var (
			policy       = policy.DeepCopy()
			ancestorRefs []gwapiv1a2.ParentReference
			resolveErr   *status.PolicyResolveError
			targetKind   string
			irKey        string
		)

		if t.MergeGateways {
			targetKind = resource.KindGatewayClass
			irKey = t.GatewayClass.Name

			ancestorRefs = []gwapiv1a2.ParentReference{
				{
					Group: GroupPtr(gwapiv1.GroupName),
					Kind:  KindPtr(targetKind),
					Name:  policy.Spec.TargetRef.Name,
				},
			}
		} else {
			targetKind = resource.KindGateway
			gatewayNN := types.NamespacedName{
				Namespace: policy.Namespace,
				Name:      string(policy.Spec.TargetRef.Name),
			}
			// It must exist since the gateways have already been processed
			irKey = irStringKey(gatewayNN.Namespace, gatewayNN.Name)

			ancestorRefs = []gwapiv1a2.ParentReference{
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
		if policy.Spec.TargetRef.Group != gwapiv1.GroupName || string(policy.Spec.TargetRef.Kind) != targetKind {
			message := fmt.Sprintf("TargetRef.Group:%s TargetRef.Kind:%s, only TargetRef.Group:%s and TargetRef.Kind:%s is supported.",
				policy.Spec.TargetRef.Group, policy.Spec.TargetRef.Kind, gwapiv1.GroupName, targetKind)

			resolveErr = &status.PolicyResolveError{
				Reason:  gwapiv1a2.PolicyReasonInvalid,
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
			irPatch.Operation.Op = ir.JSONPatchOp(patch.Operation.Op)
			irPatch.Operation.Path = patch.Operation.Path
			irPatch.Operation.JSONPath = patch.Operation.JSONPath
			irPatch.Operation.From = patch.Operation.From
			irPatch.Operation.Value = patch.Operation.Value

			policyIR.JSONPatches = append(policyIR.JSONPatches, &irPatch)
		}

		// Set Accepted=True
		status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation)
	}
}
