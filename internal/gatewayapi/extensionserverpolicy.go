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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (t *Translator) ProcessExtensionServerPolicies(policies []unstructured.Unstructured,
	gateways []*GatewayContext,
	xdsIR resource.XdsIRMap,
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
	policyIndex := -1
	// Process the policies targeting Gateways. Only update the policy status if it was accepted.
	// A policy is considered accepted if at least one targetRef contained inside matched a listener.
	for _, policy := range policies {
		policyIndex++
		policy := policy.DeepCopy()
		var policyStatus gwapiv1a2.PolicyStatus
		accepted := false
		targetRefs, err := extractTargetRefs(policy, gateways)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("error finding targetRefs for policy %s: %w", policy.GetName(), err))
			continue
		}
		for _, currTarget := range targetRefs {
			if currTarget.Kind != resource.KindGateway {
				errs = errors.Join(errs, fmt.Errorf("extension policy %s doesn't target a Gateway", policy.GetName()))
				continue
			}

			// Negative statuses have already been assigned so its safe to skip
			gateway := resolveExtServerPolicyGatewayTargetRef(policy, currTarget, gatewayMap)
			if gateway == nil {
				// unable to find a matching Gateway for policy
				continue
			}

			// Append policy extension server policy list for related gateway.
			gatewayKey := t.getIRKey(gateway.Gateway)
			unstructuredPolicy := &ir.UnstructuredRef{
				Object: &policies[policyIndex],
			}
			xdsIR[gatewayKey].ExtensionServerPolicies = append(xdsIR[gatewayKey].ExtensionServerPolicies, unstructuredPolicy)

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

func extractTargetRefs(policy *unstructured.Unstructured, gateways []*GatewayContext) ([]gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, error) {
	spec, found := policy.Object["spec"].(map[string]any)
	if !found {
		return nil, fmt.Errorf("no targets found for the policy")
	}
	specAsJSON, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("no targets found for the policy")
	}
	var targetRefs egv1a1.PolicyTargetReferences
	if err := json.Unmarshal(specAsJSON, &targetRefs); err != nil {
		return nil, fmt.Errorf("no targets found for the policy")
	}
	ret := getPolicyTargetRefs(targetRefs, gateways)
	if len(ret) == 0 {
		return nil, fmt.Errorf("no targets found for the policy")
	}
	return ret, nil
}

func policyStatusToUnstructured(policyStatus gwapiv1a2.PolicyStatus) map[string]any {
	ret := map[string]any{}
	// No need to check the marshal/unmarshal error here
	d, _ := json.Marshal(policyStatus)
	_ = json.Unmarshal(d, &ret)
	return ret
}

func resolveExtServerPolicyGatewayTargetRef(policy *unstructured.Unstructured, target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName, gateways map[types.NamespacedName]*policyGatewayTargetContext) *GatewayContext {
	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: policy.GetNamespace(),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil
	}

	return gateway.GatewayContext
}

func (t *Translator) translateExtServerPolicyForGateway(
	policy *unstructured.Unstructured,
	gateway *GatewayContext,
	target gwapiv1a2.LocalPolicyTargetReferenceWithSectionName,
	xdsIR resource.XdsIRMap,
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
