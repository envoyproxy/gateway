// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (t *Translator) ProcessExtensionServerPolicies(policies []unstructured.Unstructured,
	gateways []*GatewayContext,
	referenceGrants []*gwapiv1b1.ReferenceGrant,
	xdsIR resource.XdsIRMap,
) ([]unstructured.Unstructured, error) {
	res := []unstructured.Unstructured{}
	// ExtensionServerPolicies are already sorted by the provider layer

	// First build a map out of the gateways for faster lookup
	gatewayMap := map[types.NamespacedName]*policyGatewayTargetContext{}
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	policyCopies := extensionServerPolicyCopiesWithStatusDeepCopy(policies)

	var errs error
	// Process the policies targeting Gateways. Only update the policy status if it was accepted.
	// A policy is considered accepted if at least one targetRef contained inside matched a listener.
	for i := range policies {
		policy := policyCopies[i]
		var policyStatus gwapiv1.PolicyStatus
		accepted := false

		targetRefs, err := extractTargetRefs(&policy, gateways, referenceGrants, t.GetNamespace)
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
			gateway := resolveExtServerPolicyGatewayTargetRef(&policy, currTarget, gatewayMap)
			if gateway == nil {
				// unable to find a matching Gateway for policy
				continue
			}

			// Append policy extension server policy list for related gateway.
			gatewayKey := t.getIRKey(gateway.Gateway)
			unstructuredPolicy := &ir.UnstructuredRef{
				Object: &policy,
			}
			xdsIR[gatewayKey].ExtensionServerPolicies = append(xdsIR[gatewayKey].ExtensionServerPolicies, unstructuredPolicy)

			// Set conditions for translation if it got any
			if t.translateExtServerPolicyForGateway(&policy, gateway, currTarget, xdsIR) {
				// Set Accepted condition if it is unset
				// Only add a status condition if the policy was added into the IR
				// Find its ancestor reference by resolved gateway, even with resolve error
				gatewayNN := utils.NamespacedName(gateway)
				ancestorRef := getAncestorRefForPolicy(gatewayNN, currTarget.SectionName)
				status.SetAcceptedForPolicyAncestor(&policyStatus, &ancestorRef, t.GatewayControllerName, policy.GetGeneration())
				accepted = true
			}
		}
		if accepted {
			policy.Object["status"] = PolicyStatusToUnstructured(policyStatus)
			res = append(res, policy)
		}
	}

	return res, errs
}

func extractTargetRefs(
	policy *unstructured.Unstructured,
	gateways []*GatewayContext,
	referenceGrants []*gwapiv1b1.ReferenceGrant,
	namespaceLookup func(string) *corev1.Namespace,
) ([]policyTargetReferenceWithSectionName, error) {
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
	ret := getPolicyTargetRefs(
		targetRefs,
		gateways,
		crossNamespaceFrom{
			group:     policy.GroupVersionKind().Group,
			kind:      policy.GroupVersionKind().Kind,
			namespace: policy.GetNamespace(),
		},
		referenceGrants,
		policy.GetNamespace(),
		namespaceLookup,
	)
	if len(ret) == 0 {
		return nil, fmt.Errorf("no targets found for the policy")
	}
	return ret, nil
}

func resolveExtServerPolicyGatewayTargetRef(policy *unstructured.Unstructured, target policyTargetReferenceWithSectionName, gateways map[types.NamespacedName]*policyGatewayTargetContext) *GatewayContext {
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

func PolicyStatusToUnstructured(policyStatus gwapiv1.PolicyStatus) map[string]any {
	ret := map[string]any{}
	// No need to check the marshal/unmarshal error here
	d, _ := json.Marshal(policyStatus)
	_ = json.Unmarshal(d, &ret)
	return ret
}

func ExtServerPolicyStatusAsPolicyStatus(policy *unstructured.Unstructured) gwapiv1.PolicyStatus {
	statusObj := policy.Object["status"]
	status := gwapiv1.PolicyStatus{}
	if _, ok := statusObj.(map[string]any); ok {
		// No need to check the json marshal/unmarshal error, the policyStatus was
		// created via a typed object so the marshalling/unmarshalling will always
		// work
		d, _ := json.Marshal(statusObj)
		_ = json.Unmarshal(d, &status)
	} else if _, ok := statusObj.(gwapiv1.PolicyStatus); ok {
		status = statusObj.(gwapiv1.PolicyStatus)
	}
	return status
}

func (t *Translator) translateExtServerPolicyForGateway(
	policy *unstructured.Unstructured,
	gateway *GatewayContext,
	target policyTargetReferenceWithSectionName,
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

// extensionServerPolicyCopiesWithStatusDeepCopy returns shallow copies with deep-copied status entries.
// Status is mutated during translation and shares a pointer with the watchable coalesce goroutine.
func extensionServerPolicyCopiesWithStatusDeepCopy(policies []unstructured.Unstructured) []unstructured.Unstructured {
	copies := make([]unstructured.Unstructured, len(policies))
	for i, p := range policies {
		p.Object = maps.Clone(p.Object) // shallow copy map - no shared ref for "status" key
		if statusObj, ok := policies[i].Object["status"].(map[string]any); ok {
			p.Object["status"] = runtime.DeepCopyJSON(statusObj)
		}
		copies[i] = p
	}
	return copies
}
