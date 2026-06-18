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

	// Extract and validate the target references once per policy. The same
	// references are reused across the translation phases below.
	targetRefsList := make([]egv1a1.PolicyTargetReferences, len(policies))
	validPolicy := make([]bool, len(policies))
	for i := range policies {
		refs, err := extractTargetRefs(&policies[i])
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("error finding targetRefs for policy %s: %w", policies[i].GetName(), err))
			continue
		}
		targetRefsList[i] = refs
		validPolicy[i] = true
	}

	getOrInitPolicy := func(i int) *unstructured.Unstructured {
		policyName := utils.NamespacedName(&policies[i])
		policy, found := handledPolicies[policyName]
		if !found {
			policy = policyCopies[i]
			handledPolicies[policyName] = policy
			handledPoliciesOrder = append(handledPoliciesOrder, policyName)
		}
		return policy
	}

	// Translate, in order:
	// 1. Policies targeting route rules (HTTPRoute/GRPCRoute rules)
	// 2. Policies targeting routes (HTTPRoute/GRPCRoute)
	// 3. Policies targeting Listeners
	// 4. Policies targeting Gateways

	// Process the policies targeting route rules.
	for i := range policies {
		if !validPolicy[i] {
			continue
		}
		targetRefs := resolvePolicyTargetsFromReferences(targetRefsList[i], policies[i].GetNamespace())
		for _, currTarget := range targetRefs {
			if isRouteRule(currTarget) {
				t.processExtensionServerPolicyForRoute(xdsIR, routeMap, getOrInitPolicy(i), currTarget)
			}
		}
	}

	// Process the policies targeting routes.
	for i := range policies {
		if !validPolicy[i] {
			continue
		}
		gvk := policies[i].GroupVersionKind()
		targetRefs := resolvePolicyTargets(
			targetRefsList[i],
			routes,
			resources.ReferenceGrants,
			gvk.Group,
			gvk.Kind,
			policies[i].GetNamespace(),
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isRoute(currTarget) {
				t.processExtensionServerPolicyForRoute(xdsIR, routeMap, getOrInitPolicy(i), currTarget)
			}
		}
	}

	// Process the policies targeting Listeners
	for i := range policies {
		if !validPolicy[i] {
			continue
		}
		targetRefs := resolvePolicyTargetsFromReferences(targetRefsList[i], policies[i].GetNamespace())
		for _, currTarget := range targetRefs {
			if isListener(currTarget) {
				t.processExtensionServerPolicyForGateway(xdsIR, gatewayMap, getOrInitPolicy(i), currTarget)
			}
		}
	}

	// Process the policies targeting Gateways
	for i := range policies {
		if !validPolicy[i] {
			continue
		}
		gvk := policies[i].GroupVersionKind()
		targetRefs := resolvePolicyTargets(
			targetRefsList[i],
			gateways,
			resources.ReferenceGrants,
			gvk.Group,
			gvk.Kind,
			policies[i].GetNamespace(),
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isGateway(currTarget) {
				t.processExtensionServerPolicyForGateway(xdsIR, gatewayMap, getOrInitPolicy(i), currTarget)
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

func extractTargetRefs(policy *unstructured.Unstructured) (egv1a1.PolicyTargetReferences, error) {
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
	if (targetRefs.TargetRef == nil ||
		targetRefs.TargetRef.Group == "" ||
		targetRefs.TargetRef.Kind == "" ||
		targetRefs.TargetRef.Name == "") &&
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
	targetedRoute, resolveErr := resolveExtServerPolicyRouteTargetRef(currTarget, routeMap)
	if targetedRoute == nil {
		// Route not found
		return
	}

	// We only handle HTTPRoute/GRPCRoute for now
	routeType := targetedRoute.GetRouteType()
	supportedRouteType := routeType == resource.KindHTTPRoute || routeType == resource.KindGRPCRoute

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

		policyStatus := ExtServerPolicyStatusAsPolicyStatus(policy)
		gatewayNN := utils.NamespacedName(gtwCtx)
		ancestorRef := getAncestorRefForPolicy(gatewayNN, p.SectionName)

		// The targetRef specified a sectionName (rule) that does not exist on the route.
		if resolveErr != nil {
			status.SetResolveErrorForPolicyAncestor(&policyStatus, &ancestorRef, t.GatewayControllerName, policy.GetGeneration(), resolveErr)
			policy.Object["status"] = PolicyStatusToUnstructured(policyStatus)
			continue
		}

		if !supportedRouteType {
			status.SetTranslationErrorForPolicyAncestor(
				&policyStatus,
				&ancestorRef,
				t.GatewayControllerName,
				policy.GetGeneration(),
				fmt.Sprintf("ExtensionServerPolicy does not support targeting %s", routeType),
			)
			policy.Object["status"] = PolicyStatusToUnstructured(policyStatus)
			continue
		}

		found := false
		for _, listener := range parentRefCtx.listeners {
			irListener := gwXDS.GetHTTPListener(irListenerName(listener))
			if irListener == nil {
				continue
			}
			for _, r := range irListener.Routes {
				if currTarget.SectionName != nil && string(*currTarget.SectionName) != r.Metadata.SectionName {
					// Section name is specified but does not match the current route
					continue
				}
				if !strings.HasPrefix(r.Name, irRoutePrefix(targetedRoute)) {
					// target does not match the current route
					continue
				}
				r.ExtensionServerPolicies = appendUnstructuredRefIfAbsent(r.ExtensionServerPolicies, policy)
				found = true
			}
		}

		if found {
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
) (RouteContext, *status.PolicyResolveError) {
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: string(target.Namespace),
	}
	route, ok := routes[key]

	// Route not found. The policy may legitimately resolve once the route is
	// created, so we stay silent and do not set a negative status.
	if !ok {
		return nil, nil
	}

	if target.SectionName != nil {
		if err := validateRouteRuleSectionName(*target.SectionName, key, route); err != nil {
			return route.RouteContext, err
		}
	}

	return route.RouteContext, nil
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

func appendUnstructuredRefIfAbsent(refs []*ir.UnstructuredRef, policy *unstructured.Unstructured) []*ir.UnstructuredRef {
	for _, ref := range refs {
		if ref.Object == policy {
			return refs
		}
	}
	return append(refs, &ir.UnstructuredRef{Object: policy})
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
