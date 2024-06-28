// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (t *Translator) ProcessExtensionServerPolicies(policies []unstructured.Unstructured,
	gateways []*GatewayContext,
	xdsIR XdsIRMap,
) ([]unstructured.Unstructured, error) {
	res := []unstructured.Unstructured{}

	// Sort based on timestamp
	sort.Slice(policies, func(i, j int) bool {
		iTime := policies[i].GetCreationTimestamp()
		jTime := policies[j].GetCreationTimestamp()
		return iTime.Before(&jTime)
	})

	// First build a map out of the gateways for faster lookup
	gatewayMap := map[types.NamespacedName]*policyGatewayTargetContext{}
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	var errs error
	// Process the policies targeting Gateways. Only update the policy status if it was accepted.
	// A policy is considered accepted if at least one targetRef contained inside matched a listener.
	for _, policy := range policies {
		policy := policy.DeepCopy()
		var policyStatus gwapiv1a2.PolicyStatus
		accepted := false
		targetRefs, err := extractTargetRefs(policy)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("error finding targetRefs for policy %s: %w", policy.GetName(), err))
			continue
		}
		for _, currTarget := range targetRefs {
			if currTarget.Kind != KindGateway {
				errs = errors.Join(errs, fmt.Errorf("extension policy %s doesn't target a Gateway", policy.GetName()))
				continue
			}

			// Negative statuses have already been assigned so its safe to skip
			gateway, resolveErr := resolveExtServerPolicyGatewayTargetRef(policy, currTarget, gatewayMap)
			if gateway == nil {
				// unable to find a matching Gateway for policy
				continue
			}

			// Skip the gateway. Don't add anything to the policy status.
			if resolveErr != nil {
				// The targetRef part is somehow wrong, this policy can't be attached.
				continue
			}

			// Set conditions for translation if it got any
			if t.translateExtServerPolicyForGateway(policy, gateway, currTarget, xdsIR) {
				// Set Accepted condition if it is unset
				// Only add a status condition if the policy was added into the IR
				// Find its ancestor reference by resolved gateway, even with resolve error
				gatewayNN := utils.NamespacedName(gateway)
				ancestorRefs := []gwapiv1a2.ParentReference{
					getAncestorRefForPolicy(gatewayNN, currTarget.SectionName),
				}
				status.SetAcceptedForPolicyAncestors(&policyStatus, ancestorRefs, t.GatewayControllerName)
				accepted = true
			}
		}
		if accepted {
			res = append(res, *policy)
			policy.Object["status"] = policyStatusToUnstructured(policyStatus)
		}
	}

	return res, errs
}

func extractTargetRefs(policy *unstructured.Unstructured) ([]gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, error) {
	ret := []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{}
	spec, found := policy.Object["spec"].(map[string]any)
	if !found {
		return nil, fmt.Errorf("no targets found for the policy")
	}
	targetRefs, found := spec["targetRefs"]
	if found {
		if refArr, ok := targetRefs.([]any); ok {
			for i := range refArr {
				ref, err := extractSingleTargetRef(refArr[i])
				if err != nil {
					return nil, err
				}
				ret = append(ret, ref)
			}
		} else {
			return nil, fmt.Errorf("targetRefs is not an array")
		}
	}
	targetRef, found := spec["targetRef"]
	if found {
		ref, err := extractSingleTargetRef(targetRef)
		if err != nil {
			return nil, err
		}
		ret = append(ret, ref)
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("no targets found for the policy")
	}
	return ret, nil
}

func extractSingleTargetRef(data any) (gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, error) {
	var currRef gwapiv1a2.LocalPolicyTargetReferenceWithSectionName
	d, err := json.Marshal(data)
	if err != nil {
		return currRef, err
	}
	if err := json.Unmarshal(d, &currRef); err != nil {
		return currRef, err
	}
	if currRef.Group == "" || currRef.Name == "" || currRef.Kind == "" {
		return currRef, fmt.Errorf("invalid targetRef found: %s", string(d))
	}
	return currRef, nil
}

func policyStatusToUnstructured(policyStatus gwapiv1a2.PolicyStatus) map[string]any {
	ret := map[string]any{}
	// No need to check the marshal/unmarshal error here
	d, _ := json.Marshal(policyStatus)
	_ = json.Unmarshal(d, &ret)
	return ret
}

func resolveExtServerPolicyGatewayTargetRef(policy *unstructured.Unstructured, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, gateways map[types.NamespacedName]*policyGatewayTargetContext) (*GatewayContext, *status.PolicyResolveError) {
	targetNs := ptr.To(gwapiv1b1.Namespace(policy.GetNamespace()))

	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: string(*targetNs),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil, nil
	}

	// Ensure Policy and target are in the same namespace
	if policy.GetNamespace() != string(*targetNs) {
		message := fmt.Sprintf("Namespace:%s TargetRef.Namespace:%s, extension server policies can only target a resource in the same namespace.",
			policy.GetNamespace(), *targetNs)

		return gateway.GatewayContext, &status.PolicyResolveError{
			Reason:  gwapiv1a2.PolicyReasonInvalid,
			Message: message,
		}
	}

	return gateway.GatewayContext, nil
}

func (t *Translator) translateExtServerPolicyForGateway(
	policy *unstructured.Unstructured,
	gateway *GatewayContext,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	xdsIR XdsIRMap,
) bool {
	irKey := t.getIRKey(gateway.Gateway)
	gwIR := xdsIR[irKey]
	found := false
	for _, currListener := range gwIR.HTTP {
		listenerName := currListener.Name[strings.LastIndex(currListener.Name, "/")+1:]
		if target.SectionName != nil && string(*target.SectionName) != listenerName {
			continue
		}
		currListener.ExtensionRefs = append(currListener.ExtensionRefs, &ir.UnstructuredRef{
			Object: policy,
		})
		found = true
	}
	for _, currListener := range gwIR.TCP {
		listenerName := currListener.Name[strings.LastIndex(currListener.Name, "/")+1:]
		if target.SectionName != nil && string(*target.SectionName) != listenerName {
			continue
		}
		currListener.ExtensionRefs = append(currListener.ExtensionRefs, &ir.UnstructuredRef{
			Object: policy,
		})
		found = true
	}
	for _, currListener := range gwIR.UDP {
		listenerName := currListener.Name[strings.LastIndex(currListener.Name, "/")+1:]
		if target.SectionName != nil && string(*target.SectionName) != listenerName {
			continue
		}
		currListener.ExtensionRefs = append(currListener.ExtensionRefs, &ir.UnstructuredRef{
			Object: policy,
		})
		found = true
	}
	return found
}
