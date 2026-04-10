// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
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
			gwXdsIR     *ir.Xds
			ok          bool
		)

		refGroup, refKind, refName := policy.Spec.TargetRef.Group, policy.Spec.TargetRef.Kind, policy.Spec.TargetRef.Name
		if t.MergeGateways {
			targetKind = resource.KindGatewayClass
			// if ref GatewayClass name is not same as t.GatewayClassName, it will be skipped in L74.
			irKey = string(refName)
			ancestorRef = gwapiv1.ParentReference{
				Group: GroupPtr(gwapiv1.GroupName),
				Kind:  KindPtr(targetKind),
				Name:  refName,
			}

			gwXdsIR, ok = xdsIR[irKey]
			if !ok {
				// The TargetRef GatewayClass is not an accepted GatewayClass, then skip processing.
				message := fmt.Sprintf(
					"TargetRef.Group:%s TargetRef.Kind:%s TargetRef.Namespace:%s TargetRef.Name:%s not found or not accepted (MergeGateways=%t).",
					refGroup, refKind, policy.Namespace, string(refName), t.MergeGateways,
				)
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

		} else {
			targetKind = resource.KindGateway
			gatewayNN := types.NamespacedName{
				Namespace: policy.Namespace,
				Name:      string(refName),
			}
			irKey = t.IRKey(gatewayNN)
			ancestorRef = getAncestorRefForPolicy(gatewayNN, nil)

			gwXdsIR, ok = xdsIR[irKey]
			if !ok {
				// The TargetRef Gateway is not an accepted Gateway, then skip processing.
				continue
			}
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

		// Ensure EnvoyPatchPolicy is targeting to a support type
		if refGroup != gwapiv1.GroupName || string(refKind) != targetKind {
			message := fmt.Sprintf("TargetRef.Group:%s TargetRef.Kind:%s, only TargetRef.Group:%s and TargetRef.Kind:%s is supported.",
				refGroup, policy.Spec.TargetRef.Kind, gwapiv1.GroupName, targetKind)

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
			irPatch.Name = ir.StringMatch{
				Exact: patch.Name,
			}
			if patch.NameSelector != nil {
				irPatch.Name = *toIRStringMatch(patch.NameSelector)
			}
			// Validate that regex is valid if it's a regex match
			if r := irPatch.Name.SafeRegex; r != nil {
				_, err := regexp.Compile(*r)
				if err != nil {
					message := fmt.Sprintf("invalid regex in NameSelector: %v", err)
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
			}
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

func toIRStringMatch(sm *egv1a1.StringMatch) *ir.StringMatch {
	if sm == nil {
		return nil
	}

	switch ptr.Deref(sm.Type, egv1a1.StringMatchExact) {
	case egv1a1.StringMatchExact:
		return &ir.StringMatch{
			Exact: &sm.Value,
		}
	case egv1a1.StringMatchPrefix:
		return &ir.StringMatch{
			Prefix: &sm.Value,
		}
	case egv1a1.StringMatchSuffix:
		return &ir.StringMatch{
			Suffix: &sm.Value,
		}
	case egv1a1.StringMatchRegularExpression:
		return &ir.StringMatch{
			SafeRegex: &sm.Value,
		}
	default:
		return nil
	}
}
