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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (t *Translator) ProcessExtensionServerPolicies(
	policies []unstructured.Unstructured,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) ([]unstructured.Unstructured, error) {
	res := []unstructured.Unstructured{}
	// ExtensionServerPolicies are already sorted by the provider layer

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	routeMap := map[policyTargetRouteKey]*policyRouteTargetContext{}
	for _, route := range routes {
		key := policyTargetRouteKey{
			Kind:      string(route.GetRouteType()),
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}
		routeMap[key] = &policyRouteTargetContext{RouteContext: route}
	}
	gatewayMap := map[types.NamespacedName]*policyGatewayTargetContext{}
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	policyCopies := extensionServerPolicyCopiesWithStatusDeepCopy(policies)

	handledPolicies := make(map[types.NamespacedName]*unstructured.Unstructured, len(policies))
	// handledPoliciesOrder tracks insertion order so we can build res deterministically.
	handledPoliciesOrder := make([]types.NamespacedName, 0, len(policies))

	var errs error

	// Translate
	// 1. First translate Policies targeting RouteRules
	// 2. Next translate Policies targeting xRoutes
	// 3. Then translate Policies targeting Listeners
	// 4. Finally, the policies targeting Gateways

	// Process the policies targeting RouteRules (HTTP + TCP)
	for i, currPolicy := range policies {
		policyName := utils.NamespacedName(&currPolicy)
		policyTargetRefs, err := extractTargetRefs(&currPolicy, gateways)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("error finding targetRefs for policy %s: %w", currPolicy.GetName(), err))
			continue
		}
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(policyTargetRefs, currPolicy.GetNamespace())
		for _, currTarget := range targetRefs {
			if isRouteRule(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = policyCopies[i]
					handledPolicies[policyName] = policy
					handledPoliciesOrder = append(handledPoliciesOrder, policyName)
				}
				t.processExtensionServerPolicyForRoute(xdsIR, routeMap, policy, currTarget)
			}
		}
	}

	// Process the policies targeting xRoutes (HTTP + TCP)
	for i, currPolicy := range policies {
		policyName := utils.NamespacedName(&currPolicy)
		policyTargetRefs, err := extractTargetRefs(&currPolicy, gateways)
		if err != nil {
			// First iteration already handles error propagation
			continue
		}
		gvk := currPolicy.GroupVersionKind()
		targetRefs := resolvePolicyTargets(
			policyTargetRefs,
			routes,
			resources.ReferenceGrants,
			gvk.Group,
			gvk.Kind,
			currPolicy.GetNamespace(),
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isRoute(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = policyCopies[i]
					handledPolicies[policyName] = policy
					handledPoliciesOrder = append(handledPoliciesOrder, policyName)
				}
				t.processExtensionServerPolicyForRoute(xdsIR, routeMap, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Listeners
	for i, currPolicy := range policies {
		policyName := utils.NamespacedName(&currPolicy)
		policyTargetRefs, err := extractTargetRefs(&currPolicy, gateways)
		if err != nil {
			// First iteration already handles error propagation
			continue
		}
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(policyTargetRefs, currPolicy.GetNamespace())
		for _, currTarget := range targetRefs {
			if isListener(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = policyCopies[i]
					handledPolicies[policyName] = policy
					handledPoliciesOrder = append(handledPoliciesOrder, policyName)
				}
				t.processExtensionServerPolicyForGateway(xdsIR, gatewayMap, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Gateways
	for i, currPolicy := range policies {
		policyName := utils.NamespacedName(&currPolicy)
		policyTargetRefs, err := extractTargetRefs(&currPolicy, gateways)
		if err != nil {
			// First iteration already handles error propagation
			continue
		}
		gvk := currPolicy.GroupVersionKind()
		targetRefs := resolvePolicyTargets(
			policyTargetRefs,
			gateways,
			resources.ReferenceGrants,
			gvk.Group,
			gvk.Kind,
			currPolicy.GetNamespace(),
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isGateway(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = policyCopies[i]
					handledPolicies[policyName] = policy
					handledPoliciesOrder = append(handledPoliciesOrder, policyName)
				}
				t.processExtensionServerPolicyForGateway(xdsIR, gatewayMap, policy, currTarget)
			}
		}
	}

	// Only include policies that were accepted (have at least one ancestor status set).
	for _, name := range handledPoliciesOrder {
		policy := handledPolicies[name]
		if len(ExtServerPolicyStatusAsPolicyStatus(policy).Ancestors) > 0 {
			res = append(res, *policy)
		}
	}

	return res, errs
}

func extractTargetRefs(policy *unstructured.Unstructured, gateways []*GatewayContext) (egv1a1.PolicyTargetReferences, error) {
	var targetRefs egv1a1.PolicyTargetReferences
	spec, found := policy.Object["spec"].(map[string]any)
	if !found {
		return targetRefs, fmt.Errorf("no targets found for the policy")
	}
	specAsJSON, err := json.Marshal(spec)
	if err != nil {
		return targetRefs, fmt.Errorf("no targets found for the policy")
	}
	if err := json.Unmarshal(specAsJSON, &targetRefs); err != nil {
		return targetRefs, fmt.Errorf("no targets found for the policy")
	}
	if
		(
			targetRefs.TargetRef == nil ||
			targetRefs.TargetRef.LocalPolicyTargetReference.Group == "" ||
			targetRefs.TargetRef.LocalPolicyTargetReference.Kind == "" ||
			targetRefs.TargetRef.LocalPolicyTargetReference.Name == "") &&
			len(targetRefs.TargetRefs) < 1 &&
			len(targetRefs.TargetSelectors) < 1 {
		return targetRefs, fmt.Errorf("no targets found for the policy")
	}
	return targetRefs, nil
}

func (t *Translator) processExtensionServerPolicyForRoute(
	xdsIR resource.XdsIRMap,
	routeMap map[policyTargetRouteKey]*policyRouteTargetContext,
	policy *unstructured.Unstructured,
	currTarget policyTargetReferenceWithSectionName,
) {
	targetedRoute := resolveExtServerPolicyRouteTargetRef(currTarget, routeMap)
	if targetedRoute == nil {
		return
	}

	parentRefs := GetParentReferences(targetedRoute)

	for _, p := range parentRefs {
		parentRefCtx := targetedRoute.GetRouteParentContext(p)
		if parentRefCtx == nil {
			continue
		}
		gtwCtx := parentRefCtx.GetGateway()
		if gtwCtx == nil {
			continue
		}

		irKey := t.getIRKey(gtwCtx.Gateway)
		gwXDS, ok := xdsIR[irKey]
		if !ok {
			continue
		}

		found := false
		switch targetedRoute.GetRouteType() {
		case resource.KindHTTPRoute, resource.KindGRPCRoute:
			for _, listener := range parentRefCtx.listeners {
				irListener := gwXDS.GetHTTPListener(irListenerName(listener))
				if irListener == nil {
					continue
				}
				// Only add policy to listener if it does not target a specific route
				if currTarget.SectionName == nil {
					irListener.ExtensionServerPolicies = append(irListener.ExtensionServerPolicies, &ir.UnstructuredRef{Object: policy})
				}
				for _, r := range irListener.Routes {
					if currTarget.SectionName != nil && string(*currTarget.SectionName) != r.Metadata.SectionName {
						continue
					}
					if strings.HasPrefix(r.Name, irRoutePrefix(targetedRoute)) {
						r.ExtensionServerPolicies = append(r.ExtensionServerPolicies, &ir.UnstructuredRef{Object: policy})
						found = true
					}
				}
			}
		}

		if found {
			policyStatus := ExtServerPolicyStatusAsPolicyStatus(policy)
			gatewayNN := utils.NamespacedName(gtwCtx)
			ancestorRef := getAncestorRefForPolicy(gatewayNN, p.SectionName)
			status.SetAcceptedForPolicyAncestor(&policyStatus, &ancestorRef, t.GatewayControllerName, policy.GetGeneration())
			policy.Object["status"] = PolicyStatusToUnstructured(policyStatus)
		}
	}
}

func (t *Translator) processExtensionServerPolicyForGateway(
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	policy *unstructured.Unstructured,
	currTarget policyTargetReferenceWithSectionName,
) {

	// Negative statuses have already been assigned so its safe to skip
	gateway := resolveExtServerPolicyGatewayTargetRef(currTarget, gatewayMap)
	if gateway == nil {
		// unable to find a matching Gateway for policy
		return
	}

	// Append policy extension server policy list for related gateway.
	gatewayKey := t.getIRKey(gateway.Gateway)
	xdsIR[gatewayKey].ExtensionServerPolicies = append(xdsIR[gatewayKey].ExtensionServerPolicies, &ir.UnstructuredRef{Object: policy})

	if t.translateExtServerPolicyForGateway(policy, gateway, currTarget, xdsIR) {
		policyStatus := ExtServerPolicyStatusAsPolicyStatus(policy)
		gatewayNN := utils.NamespacedName(gateway)
		ancestorRef := getAncestorRefForPolicy(gatewayNN, currTarget.SectionName)
		status.SetAcceptedForPolicyAncestor(&policyStatus, &ancestorRef, t.GatewayControllerName, policy.GetGeneration())
		policy.Object["status"] = PolicyStatusToUnstructured(policyStatus)
	}
}

func resolveExtServerPolicyGatewayTargetRef(
	target policyTargetReferenceWithSectionName,
	gateways map[types.NamespacedName]*policyGatewayTargetContext,
) *GatewayContext {
	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: string(target.Namespace),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil
	}

	return gateway.GatewayContext
}

func resolveExtServerPolicyRouteTargetRef(
	target policyTargetReferenceWithSectionName,
	routes map[policyTargetRouteKey]*policyRouteTargetContext,
) RouteContext {
	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: string(target.Namespace),
	}
	route, ok := routes[key]

	// Route not found
	if !ok {
		return nil
	}

	// If sectionName is set, make sure its valid
	if target.SectionName != nil {
		if err := validateRouteRuleSectionName(*target.SectionName, key, route); err != nil {
			// unable to find a matching section name for policy in route
			return nil
		}
	}

	return route.RouteContext
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
func extensionServerPolicyCopiesWithStatusDeepCopy(policies []unstructured.Unstructured) []*unstructured.Unstructured {
	copies := make([]*unstructured.Unstructured, len(policies))
	for i, p := range policies {
		p.Object = maps.Clone(p.Object) // shallow copy map - no shared ref for "status" key
		if statusObj, ok := policies[i].Object["status"].(map[string]any); ok {
			p.Object["status"] = runtime.DeepCopyJSON(statusObj)
		}
		copies[i] = &p
	}
	return copies
}
