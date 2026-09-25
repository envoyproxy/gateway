// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

// policyKey uniquely identifies an extension server policy by gvk and namespaced name.
// See https://github.com/envoyproxy/gateway/pull/9249#discussion_r3437536064
type policyKey struct {
	gvk schema.GroupVersionKind
	types.NamespacedName
}

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

	handledPolicies := make(map[policyKey]*unstructured.Unstructured, len(policies))
	// handledPoliciesOrder tracks insertion order so we can build res deterministically.
	handledPoliciesOrder := make([]policyKey, 0, len(policies))

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
		key := policyKey{
			gvk:            policies[i].GroupVersionKind(),
			NamespacedName: utils.NamespacedName(&policies[i]),
		}
		policy, found := handledPolicies[key]
		if !found {
			policy = &policies[i]
			handledPolicies[key] = policy
			handledPoliciesOrder = append(handledPoliciesOrder, key)
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
	for _, key := range handledPoliciesOrder {
		policy := handledPolicies[key]
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
	supportedRouteKind := routeType == resource.KindHTTPRoute || routeType == resource.KindGRPCRoute

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

		if !supportedRouteKind {
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

		// The targetRef specified a sectionName (rule) that does not exist on the route.
		if resolveErr != nil {
			status.SetResolveErrorForPolicyAncestor(&policyStatus, &ancestorRef, t.GatewayControllerName, policy.GetGeneration(), resolveErr)
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
				r.ExtensionServerPolicies = t.appendUnstructuredRefIfAbsent(gwXDS, gtwCtx, r.ExtensionServerPolicies, policy)
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
	gwIR := xdsIR[gatewayKey]
	gwIR.ExtensionServerPolicies = t.appendUnstructuredRefIfAbsent(gwIR, gateway, gwIR.ExtensionServerPolicies, policy)

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

// translateExtServerPolicyForGateway attaches the policy to the IR listeners of
// the targeted Gateway: all of them, or only the one named by the target's
// sectionName. The listeners are taken from the Gateway rather than from the IR,
// because with mergeGateways enabled the IR holds the listeners of every Gateway
// in the GatewayClass, including listeners of other Gateways with the same name.
func (t *Translator) translateExtServerPolicyForGateway(
	policy *unstructured.Unstructured,
	gateway *GatewayContext,
	target policyTargetReferenceWithSectionName,
	xdsIR resource.XdsIRMap,
) bool {
	gwIR := xdsIR[t.getIRKey(gateway.Gateway)]
	found := false
	for _, listener := range gateway.listeners {
		if target.SectionName != nil && *target.SectionName != listener.Name {
			continue
		}
		irListener := getIRCoreListener(gwIR, irListenerName(listener))
		if irListener == nil {
			continue
		}
		irListener.ExtensionRefs = append(irListener.ExtensionRefs, t.getOrCreateExtensionResource(gwIR, gateway, policy))
		found = true
	}
	return found
}

// getIRCoreListener returns the common details of the HTTP, TCP or UDP IR
// listener with the given name, or nil if the IR has no such listener.
func getIRCoreListener(xdsIR *ir.Xds, name string) *ir.CoreListenerDetails {
	if l := xdsIR.GetHTTPListener(name); l != nil {
		return &l.CoreListenerDetails
	}
	if l := xdsIR.GetTCPListener(name); l != nil {
		return &l.CoreListenerDetails
	}
	if l := xdsIR.GetUDPListener(name); l != nil {
		return &l.CoreListenerDetails
	}
	return nil
}

// appendUnstructuredRefIfAbsent appends a ref to obj to refs, unless refs already carries a ref
// to the same resource. Identity is compared by the stable Name getOrCreateExtensionResource
// derives for obj, or by Object pointer when no gateway context is available to scope a name.
func (t *Translator) appendUnstructuredRefIfAbsent(gwIR *ir.Xds, gatewayCtx *GatewayContext, refs []*ir.UnstructuredRef, obj *unstructured.Unstructured) []*ir.UnstructuredRef {
	ref := t.getOrCreateExtensionResource(gwIR, gatewayCtx, obj)
	for _, existing := range refs {
		if ref.Object != nil {
			if existing.Object == ref.Object {
				return refs
			}
		} else if existing.Name == ref.Name {
			return refs
		}
	}
	return append(refs, ref)
}

// getOrCreateExtensionResource returns a ref to obj. It registers obj once per distinct identity
// into gwIR.ExtensionResources and returns a lightweight name-only ref, using
// t.ExtensionResourceMap as a find-or-create cache; a later call for an already-registered
// identity refreshes the canonical entry to the new obj rather than keeping the first one seen.
// When no gateway context is available to scope a name, it falls back to a self-contained ref
// with obj embedded directly.
func (t *Translator) getOrCreateExtensionResource(
	gwIR *ir.Xds,
	gatewayCtx *GatewayContext,
	obj *unstructured.Unstructured,
) *ir.UnstructuredRef {
	if gatewayCtx == nil {
		return &ir.UnstructuredRef{Object: obj}
	}

	gvk := obj.GroupVersionKind()
	key := ExtensionResourceKey{
		GatewayIRKey: t.getIRKey(gatewayCtx.Gateway),
		Group:        gvk.Group,
		Kind:         gvk.Kind,
		Namespace:    obj.GetNamespace(),
		Name:         obj.GetName(),
	}

	if t.ExtensionResourceMap == nil {
		t.ExtensionResourceMap = make(map[ExtensionResourceKey]*ir.UnstructuredRef)
	}

	if canonical, ok := t.ExtensionResourceMap[key]; ok {
		// Refresh to the object instance being registered now. The same identity can be
		// resolved from more than one source (e.g. an extensionRef/custom backend as well as
		// an ExtensionServerPolicy target), each holding its own *unstructured.Unstructured
		// copy, and later phases (e.g. ExtensionServerPolicy status) mutate their copy in
		// place after registering it here. Keeping the most recently registered copy
		// canonical ensures those mutations remain visible to anything resolving by Name,
		// instead of pinning the registry to a stale, earlier copy.
		canonical.Object = obj
		return &ir.UnstructuredRef{Name: canonical.Name}
	}

	name := irExtensionResourceName(&key)
	canonical := &ir.UnstructuredRef{Name: name, Object: obj}
	t.ExtensionResourceMap[key] = canonical
	if gwIR != nil {
		gwIR.ExtensionResources = append(gwIR.ExtensionResources, canonical)
	}
	return &ir.UnstructuredRef{Name: name}
}
