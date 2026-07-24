// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	perr "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/luavalidator"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/wasm"
)

// oci URL prefix
const (
	ociURLPrefix = "oci://"
	// LuaConfigMapKey is the key used in ConfigMaps to store Lua scripts
	LuaConfigMapKey = "lua"
)

// deprecatedFieldsUsedInEnvoyExtensionPolicy returns a map of deprecated field paths to their alternatives.
func deprecatedFieldsUsedInEnvoyExtensionPolicy(policy *egv1a1.EnvoyExtensionPolicy) map[string]string {
	deprecatedFields := make(map[string]string)
	if policy.Spec.TargetRef != nil {
		deprecatedFields["spec.targetRef"] = "spec.targetRefs"
	}
	return deprecatedFields
}

func validateDynamicModuleRemoteURL(rawURL string) error {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return err
	}

	switch parsedURL.Scheme {
	case "http", "https":
	default:
		return fmt.Errorf("unsupported URL scheme %q", parsedURL.Scheme)
	}

	if parsedURL.Hostname() == "" {
		return errors.New("URL must include a hostname")
	}

	if port := parsedURL.Port(); port != "" {
		if _, err := strconv.Atoi(port); err != nil {
			return fmt.Errorf("invalid URL port %q: %w", port, err)
		}
	}

	return nil
}

func (t *Translator) ProcessEnvoyExtensionPolicies(
	envoyExtensionPolicies []*egv1a1.EnvoyExtensionPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) []*egv1a1.EnvoyExtensionPolicy {
	var res []*egv1a1.EnvoyExtensionPolicy
	// EnvoyExtensionPolicies are already sorted by the provider layer

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	routeMapSize := len(routes)
	gatewayMapSize := len(gateways)
	listenerSetMapSize := len(resources.ListenerSets)

	routeMap := make(map[policyTargetRouteKey]*policyRouteTargetContext, routeMapSize)
	for _, route := range routes {
		key := policyTargetRouteKey{
			Kind:      string(route.GetRouteType()),
			Name:      route.GetName(),
			Namespace: route.GetNamespace(),
		}
		routeMap[key] = &policyRouteTargetContext{RouteContext: route}
	}

	gatewayMap := make(map[types.NamespacedName]*policyGatewayTargetContext, gatewayMapSize)
	for _, gw := range gateways {
		key := utils.NamespacedName(gw)
		gatewayMap[key] = &policyGatewayTargetContext{GatewayContext: gw}
	}

	listenerSetMap := make(map[types.NamespacedName]*policyListenerSetTargetContext, listenerSetMapSize)
	for _, ls := range resources.ListenerSets {
		key := utils.NamespacedName(ls)
		listenerSetMap[key] = &policyListenerSetTargetContext{ListenerSet: ls}
	}

	handledPolicies := make(map[types.NamespacedName]*egv1a1.EnvoyExtensionPolicy)

	// overrides records child scopes whose policies displace policies attached
	// to their parent scopes.
	overrides := newPolicyScopeGraph()

	// Translate
	// 1. First translate Policies targeting RouteRules
	// 2. Next translate Policies targeting xRoutes
	// 3. Then translate Policies targeting ListenerSet Listeners
	// 4. Then translate Policies targeting ListenerSets
	// 5. Then translate Policies targeting Gateway Listeners
	// 6. Finally, the policies targeting Gateways

	// Process the policies targeting RouteRules
	for i, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isRouteRule(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = envoyExtensionPolicies[i]
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				t.processEnvoyExtensionPolicyForRoute(resources, xdsIR,
					routeMap, listenerSetMap, overrides, policy, currTarget)
			}
		}
	}

	// Process the policies targeting xRoutes
	for i, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			routes,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindEnvoyExtensionPolicy,
			currPolicy.Namespace,
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isRoute(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = envoyExtensionPolicies[i]
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				t.processEnvoyExtensionPolicyForRoute(resources, xdsIR,
					routeMap, listenerSetMap, overrides, policy, currTarget)
			}
		}
	}

	// Process the policies targeting ListenerSet Listeners
	for i, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isListenerSetListener(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = policyCopies[i]
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				t.processEnvoyExtensionPolicyForListenerSet(resources, xdsIR,
					gatewayMap, listenerSetMap, overrides, policy, currTarget)
			}
		}
	}

	// Process the policies targeting ListenerSets
	for i, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			resources.ListenerSets,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindEnvoyExtensionPolicy,
			currPolicy.Namespace,
			t.GetNamespace,
		)
		for _, currTarget := range targetRefs {
			if isListenerSet(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = policyCopies[i]
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				t.processEnvoyExtensionPolicyForListenerSet(resources, xdsIR,
					gatewayMap, listenerSetMap, overrides, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Gateway Listeners
	for i, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isListener(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = envoyExtensionPolicies[i]
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				t.processEnvoyExtensionPolicyForGateway(resources, xdsIR,
					gatewayMap, overrides, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Gateways
	for i, currPolicy := range envoyExtensionPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			gateways,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindEnvoyExtensionPolicy,
			currPolicy.Namespace,
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isGateway(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = envoyExtensionPolicies[i]
					res = append(res, policy)
					handledPolicies[policyName] = policy
				}

				t.processEnvoyExtensionPolicyForGateway(resources, xdsIR,
					gatewayMap, overrides, policy, currTarget)
			}
		}
	}

	for _, policy := range res {
		// Truncate Ancestor list of longer than 16
		if len(policy.Status.Ancestors) > 16 {
			status.TruncatePolicyAncestors(&policy.Status, t.GatewayControllerName, policy.Generation)
		}
	}
	return res
}

func (t *Translator) processEnvoyExtensionPolicyForRoute(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	routeMap map[policyTargetRouteKey]*policyRouteTargetContext,
	listenerSetMap map[types.NamespacedName]*policyListenerSetTargetContext,
	overrides policyScopeGraph,
	policy *egv1a1.EnvoyExtensionPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targetedRoute RouteContext
		ancestorRefs  []*gwapiv1.ParentReference
		resolveErr    *status.PolicyResolveError
	)

	targetedRoute, resolveErr = resolveEnvoyExtensionPolicyRouteTargetRef(currTarget, routeMap)
	// Skip if the route is not found
	// It's not necessarily an error because the EnvoyExtensionPolicy may be
	// reconciled by multiple controllers. And the other controller may
	// have the target route.
	if targetedRoute == nil {
		return
	}

	// Find the parent resource that the route belongs to and record its
	// ancestor status and override relationship.
	parentRefs := GetManagedParentReferences(targetedRoute)
	routeNN := utils.NamespacedName(targetedRoute)
	routeAsChildScope := routeScope(routeNN)
	for _, p := range parentRefs {
		parentNamespace := targetedRoute.GetNamespace()
		if p.Namespace != nil {
			parentNamespace = string(*p.Namespace)
		}
		parentNN := types.NamespacedName{Namespace: parentNamespace, Name: string(p.Name)}

		if p.Kind == nil || *p.Kind == resource.KindGateway {
			// Record the Route under the Gateway scope it attaches to:
			// Gateway listener when sectionName is set, otherwise Gateway.
			if p.SectionName != nil {
				overrides.Add(gatewayListenerScope(parentNN, *p.SectionName), routeAsChildScope)
			} else {
				overrides.Add(gatewayScope(parentNN), routeAsChildScope)
			}

			// Do need a section name since the policy is targeting to a route
			ancestorRef := getAncestorRefForPolicy(parentNN, p.SectionName)
			ancestorRefs = append(ancestorRefs, &ancestorRef)
		} else if *p.Kind == resource.KindListenerSet {
			// The Route attaches through a ListenerSet. Resolve the ListenerSet
			// so its parent Gateway can be registered as structural containment;
			// the Route relationship itself is recorded under the ListenerSet
			// scope below.
			lsCtx, ok := listenerSetMap[parentNN]
			if !ok {
				continue
			}
			parentGwNN := types.NamespacedName{
				Name:      string(lsCtx.Spec.ParentRef.Name),
				Namespace: NamespaceDerefOr(lsCtx.Spec.ParentRef.Namespace, lsCtx.Namespace),
			}
			overrides.RegisterListenerSet(parentNN, parentGwNN)

			if p.SectionName != nil {
				overrides.Add(listenerSetListenerScope(parentNN, *p.SectionName), routeAsChildScope)
			} else {
				overrides.Add(listenerSetScope(parentNN), routeAsChildScope)
			}

			// ListenerSet-attached Route policies report status against the
			// ListenerSet itself.
			ancestorRef := getAncestorRefForListenerSetPolicy(parentNN, p.SectionName)
			ancestorRefs = append(ancestorRefs, &ancestorRef)
		}
	}

	// Set conditions for resolve error, then skip current xroute
	if resolveErr != nil {
		status.SetResolveErrorForPolicyAncestors(&policy.Status,
			ancestorRefs,
			t.GatewayControllerName,
			policy.Generation,
			resolveErr,
		)
		return
	}

	// Set conditions for translation error if it got any
	if err := t.translateEnvoyExtensionPolicyForRoute(policy, targetedRoute, currTarget, xdsIR, resources); err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			ancestorRefs,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation)

	// Check for deprecated fields and set warning if any are found
	if deprecatedFields := deprecatedFieldsUsedInEnvoyExtensionPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}

	// Check if this policy is overridden by other policies targeting at route rule levels
	key := policyTargetRouteKey{
		Kind:      string(currTarget.Kind),
		Name:      string(currTarget.Name),
		Namespace: string(currTarget.Namespace),
	}
	overriddenTargetsMessage := getOverriddenTargetsMessageForRoute(routeMap[key], currTarget.SectionName)
	if overriddenTargetsMessage != "" {
		status.SetConditionForPolicyAncestors(&policy.Status,
			ancestorRefs,
			t.GatewayControllerName,
			egv1a1.PolicyConditionOverridden,
			metav1.ConditionTrue,
			egv1a1.PolicyReasonOverridden,
			"This policy is being overridden by other envoyExtensionPolicies for "+overriddenTargetsMessage,
			policy.Generation,
		)
	}
}

func (t *Translator) processEnvoyExtensionPolicyForListenerSet(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	listenerSetMap map[types.NamespacedName]*policyListenerSetTargetContext,
	overrides policyScopeGraph,
	policy *egv1a1.EnvoyExtensionPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targeted   *gwapiv1.ListenerSet
		resolveErr *status.PolicyResolveError
	)

	targeted, resolveErr = resolveEnvoyExtensionPolicyListenerSetTargetRef(currTarget, listenerSetMap)
	// Skip if the ListenerSet is not found. The EnvoyExtensionPolicy may be
	// reconciled by multiple controllers, and another controller may own it.
	if targeted == nil {
		return
	}

	parentGatewayNN := types.NamespacedName{
		Name:      string(targeted.Spec.ParentRef.Name),
		Namespace: NamespaceDerefOr(targeted.Spec.ParentRef.Namespace, targeted.Namespace),
	}
	gateway, ok := gatewayMap[parentGatewayNN]
	if !ok {
		return
	}

	// Use the ListenerSet itself as the policy ancestor (not the parent Gateway).
	listenerSetNN := utils.NamespacedName(targeted)
	ancestorRef := getAncestorRefForListenerSetPolicy(listenerSetNN, currTarget.SectionName)

	// Set conditions for resolve error, then skip current ListenerSet
	if resolveErr != nil {
		status.SetResolveErrorForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			policy.Generation,
			resolveErr,
		)
		return
	}

	// Record the ListenerSet policy under the scope it attaches to. Listener
	// policies are children of the ListenerSet scope; ListenerSet-wide policies
	// are children of the parent Gateway scope.
	if currTarget.SectionName != nil {
		overrides.RegisterListenerSet(listenerSetNN, parentGatewayNN)
		overrides.Add(listenerSetScope(listenerSetNN), listenerSetListenerScope(listenerSetNN, *currTarget.SectionName))
	} else {
		overrides.Add(gatewayScope(parentGatewayNN), listenerSetScope(listenerSetNN))
	}

	if err := t.translateEnvoyExtensionPolicyForListenerSet(policy, currTarget, gateway.GatewayContext, targeted, xdsIR, resources); err != nil {
		status.SetTranslationErrorForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation)

	// Check for deprecated fields and set warning if any are found
	if deprecatedFields := deprecatedFieldsUsedInEnvoyExtensionPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}

	// Determine this policy's own scope so we can look up overriding child scopes from the relation maps.
	var lsParentScope policyScope
	if currTarget.SectionName == nil {
		lsParentScope = listenerSetScope(listenerSetNN)
	} else {
		lsParentScope = listenerSetListenerScope(listenerSetNN, *currTarget.SectionName)
	}

	overriddenMessage := formatPolicyScopes(overrides.GetWithDescendants(lsParentScope))
	if overriddenMessage != "" {
		status.SetConditionForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			egv1a1.PolicyConditionOverridden,
			metav1.ConditionTrue,
			egv1a1.PolicyReasonOverridden,
			"This policy is being overridden by other envoyExtensionPolicies for "+overriddenMessage,
			policy.Generation,
		)
	}
}

func (t *Translator) processEnvoyExtensionPolicyForGateway(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	overrides policyScopeGraph,
	policy *egv1a1.EnvoyExtensionPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targetedGateway *GatewayContext
		resolveErr      *status.PolicyResolveError
	)

	targetedGateway, resolveErr = resolveEnvoyExtensionPolicyGatewayTargetRef(currTarget, gatewayMap)
	// Skip if the gateway is not found
	// It's not necessarily an error because the EnvoyExtensionPolicy may be
	// reconciled by multiple controllers. And the other controller may
	// have the target gateway.
	if targetedGateway == nil {
		return
	}

	// Find its ancestor reference by resolved gateway, even with resolve error
	gatewayNN := utils.NamespacedName(targetedGateway)
	ancestorRef := getAncestorRefForPolicy(gatewayNN, currTarget.SectionName)

	// Set conditions for resolve error, then skip current gateway
	if resolveErr != nil {
		status.SetResolveErrorForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			policy.Generation,
			resolveErr,
		)

		return
	}

	// Record this policy as an override of the parent Gateway scope when the
	// target is a Gateway listener (sectionName set).
	if currTarget.SectionName != nil {
		overrides.Add(gatewayScope(gatewayNN), gatewayListenerScope(gatewayNN, *currTarget.SectionName))
	}

	// Set conditions for translation error if it got any
	if err := t.translateEnvoyExtensionPolicyForGateway(policy, currTarget, targetedGateway, xdsIR, resources); err != nil {
		status.SetTranslationErrorForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation)

	// Check for deprecated fields and set warning if any are found
	if deprecatedFields := deprecatedFieldsUsedInEnvoyExtensionPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}

	// Determine this policy's own scope so we can look up overriding child scopes from the relation maps.
	var parentScope policyScope
	if currTarget.SectionName == nil {
		parentScope = gatewayScope(gatewayNN)
	} else {
		parentScope = gatewayListenerScope(gatewayNN, *currTarget.SectionName)
	}
	overriddenTargetsMessage := formatPolicyScopes(overrides.GetWithDescendants(parentScope))
	if overriddenTargetsMessage != "" {
		status.SetConditionForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			egv1a1.PolicyConditionOverridden,
			metav1.ConditionTrue,
			egv1a1.PolicyReasonOverridden,
			"This policy is being overridden by other envoyExtensionPolicies for "+overriddenTargetsMessage,
			policy.Generation,
		)
	}
}

func resolveEnvoyExtensionPolicyGatewayTargetRef(
	target policyTargetReferenceWithSectionName,
	gateways map[types.NamespacedName]*policyGatewayTargetContext,
) (*GatewayContext, *status.PolicyResolveError) {
	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: string(target.Namespace),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	if !ok {
		return nil, nil
	}

	// If sectionName is set, make sure its valid
	if target.SectionName != nil {
		if err := validateGatewayListenerSectionName(
			*target.SectionName,
			key,
			gatewayDirectListeners(gateway.GatewayContext),
		); err != nil {
			return gateway.GatewayContext, err
		}
	}

	if target.SectionName == nil {
		// Check if another policy targeting the same Gateway exists
		if gateway.attached {
			message := fmt.Sprintf("Unable to target Gateway %s, another EnvoyExtensionPolicy has already attached to it",
				string(target.Name))

			return gateway.GatewayContext, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		gateway.attached = true
	} else {
		listenerName := string(*target.SectionName)
		if gateway.attachedToListeners != nil && gateway.attachedToListeners.Has(listenerName) {
			message := fmt.Sprintf("Unable to target Listener %s/%s, another EnvoyExtensionPolicy has already attached to it",
				string(target.Name), listenerName)

			return gateway.GatewayContext, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		if gateway.attachedToListeners == nil {
			gateway.attachedToListeners = make(sets.Set[string])
		}
		gateway.attachedToListeners.Insert(listenerName)
	}

	gateways[key] = gateway

	return gateway.GatewayContext, nil
}

func resolveEnvoyExtensionPolicyListenerSetTargetRef(
	target policyTargetReferenceWithSectionName,
	listenerSets map[types.NamespacedName]*policyListenerSetTargetContext,
) (*gwapiv1.ListenerSet, *status.PolicyResolveError) {
	// Find the ListenerSet
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: string(target.Namespace),
	}
	ls, ok := listenerSets[key]
	if !ok {
		return nil, nil
	}

	// If sectionName is set, make sure its valid
	if target.SectionName != nil {
		if err := validateListenerSetListenerSectionName(
			*target.SectionName,
			key,
			ls.Spec.Listeners,
		); err != nil {
			return ls.ListenerSet, err
		}
	}

	if target.SectionName == nil {
		// Check if another policy targeting the same ListenerSet exists
		if ls.attached {
			message := fmt.Sprintf("Unable to target ListenerSet %s, another EnvoyExtensionPolicy has already attached to it",
				string(target.Name))

			return ls.ListenerSet, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		ls.attached = true
	} else {
		listenerName := string(*target.SectionName)
		if ls.attachedToListeners != nil && ls.attachedToListeners.Has(listenerName) {
			message := fmt.Sprintf("Unable to target Listener %s/%s, another EnvoyExtensionPolicy has already attached to it",
				string(target.Name), listenerName)

			return ls.ListenerSet, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		if ls.attachedToListeners == nil {
			ls.attachedToListeners = make(sets.Set[string])
		}
		ls.attachedToListeners.Insert(listenerName)
	}

	listenerSets[key] = ls

	return ls.ListenerSet, nil
}

func resolveEnvoyExtensionPolicyRouteTargetRef(
	target policyTargetReferenceWithSectionName,
	routes map[policyTargetRouteKey]*policyRouteTargetContext,
) (RouteContext, *status.PolicyResolveError) {
	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: string(target.Namespace),
	}

	route, ok := routes[key]
	// Route not found
	if !ok {
		return nil, nil
	}

	// If sectionName is set, make sure its valid
	if target.SectionName != nil {
		if err := validateRouteRuleSectionName(*target.SectionName, key, route); err != nil {
			return route.RouteContext, err
		}
	}

	if target.SectionName == nil {
		// Check if another policy targeting the same xRoute exists
		if route.attached {
			message := fmt.Sprintf("Unable to target %s %s, another EnvoyExtensionPolicy has already attached to it",
				string(target.Kind), string(target.Name))

			return route.RouteContext, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		route.attached = true
	} else {
		routeRuleName := string(*target.SectionName)
		if route.attachedToRouteRules != nil && route.attachedToRouteRules.Has(routeRuleName) {
			message := fmt.Sprintf("Unable to target RouteRule %s/%s, another EnvoyExtensionPolicy has already attached to it",
				string(target.Name), routeRuleName)

			return route.RouteContext, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}
		if route.attachedToRouteRules == nil {
			route.attachedToRouteRules = make(sets.Set[string])
		}
		route.attachedToRouteRules.Insert(routeRuleName)
	}

	routes[key] = route

	return route.RouteContext, nil
}

func (t *Translator) translateEnvoyExtensionPolicyForRoute(
	policy *egv1a1.EnvoyExtensionPolicy,
	route RouteContext,
	target policyTargetReferenceWithSectionName,
	xdsIR resource.XdsIRMap,
	resources *resource.Resources,
) error {
	var (
		wasms                                                 []ir.Wasm
		luas                                                  []ir.Lua
		wasmFailOpen, extProcFailOpen                         bool
		wasmError, luaError, extProcError, dynamicModuleError error
		errs                                                  error
	)

	if wasms, wasmError, wasmFailOpen = t.buildWasms(policy, resources); wasmError != nil {
		wasmError = perr.WithMessage(wasmError, "Wasm")
		errs = errors.Join(errs, wasmError)
	}

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	parentRefs := GetParentReferences(route)
	routesWithDirectResponse := sets.New[string]()
	for _, p := range parentRefs {
		// Skip if this parentRef was not processed by this translator
		// (e.g., references a Gateway with a different GatewayClass)
		parentRefCtx := route.GetRouteParentContext(p)
		if parentRefCtx == nil {
			continue
		}
		gtwCtx := parentRefCtx.GetGateway()
		if gtwCtx == nil {
			continue
		}

		if luas, luaError = t.buildLuas(policy, gtwCtx.envoyProxy); luaError != nil {
			luaError = perr.WithMessage(luaError, "Lua")
			errs = errors.Join(errs, luaError)
		}

		var extProcs []ir.ExtProc
		if extProcs, extProcError, extProcFailOpen = t.buildExtProcs(policy, resources, gtwCtx); extProcError != nil {
			extProcError = perr.WithMessage(extProcError, "ExtProc")
			errs = errors.Join(errs, extProcError)
		}

		var dynamicModules []ir.DynamicModule
		if dynamicModules, dynamicModuleError = t.buildDynamicModules(policy, gtwCtx.envoyProxy); dynamicModuleError != nil {
			dynamicModuleError = perr.WithMessage(dynamicModuleError, "DynamicModule")
			errs = errors.Join(errs, dynamicModuleError)
		}

		irKey := t.getIRKey(gtwCtx.Gateway)
		for _, listener := range parentRefCtx.listeners {
			irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))
			if irListener != nil {
				for _, r := range irListener.Routes {
					// If specified the sectionName must match route rule from ir route metadata.
					if target.SectionName != nil && string(*target.SectionName) != r.Metadata.SectionName {
						continue
					}

					// A Policy targeting the most specific scope(xRoute rule) wins over a policy
					// targeting a lesser specific scope(xRoute).
					if strings.HasPrefix(r.Name, prefix) {
						// if already set - there's a specific level policy, so skip
						if r.EnvoyExtensions != nil {
							continue
						}

						failRoute := false
						// Lua extension doesn't have a fail open option, so fail the route if there is a lua error
						// TODO: we may also add fail open option for Lua extension to align with other extensions
						if luaError != nil {
							failRoute = true
						}
						if wasmError != nil {
							failRoute = failRoute || !wasmFailOpen
						}
						if extProcError != nil {
							failRoute = failRoute || !extProcFailOpen
						}
						if dynamicModuleError != nil {
							failRoute = true
						}
						if failRoute {
							r.DirectResponse = &ir.CustomResponse{
								StatusCode: new(uint32(500)),
							}
							routesWithDirectResponse.Insert(r.Name)
						} else {
							r.EnvoyExtensions = &ir.EnvoyExtensionFeatures{
								ExtProcs:       extProcs,
								Wasms:          wasms,
								Luas:           luas,
								DynamicModules: dynamicModules,
							}
						}
					}
				}
			}
		}
	}
	if len(routesWithDirectResponse) > 0 {
		t.Logger.Info("setting 500 direct response in routes due to errors in EnvoyExtensionPolicy",
			"policy", fmt.Sprintf("%s/%s", policy.Namespace, policy.Name),
			"routes", sets.List(routesWithDirectResponse),
			"error", errs,
		)
	}

	return errs
}

func (t *Translator) translateEnvoyExtensionPolicyForGateway(
	policy *egv1a1.EnvoyExtensionPolicy,
	target policyTargetReferenceWithSectionName,
	gateway *GatewayContext,
	xdsIR resource.XdsIRMap,
	resources *resource.Resources,
) error {
	return t.translateEnvoyExtensionPolicyForListeners(
		policy,
		gateway,
		xdsIR,
		resources,
		gatewayPolicyTargetListeners(gateway, target),
	)
}

func (t *Translator) translateEnvoyExtensionPolicyForListenerSet(
	policy *egv1a1.EnvoyExtensionPolicy,
	target policyTargetReferenceWithSectionName,
	gateway *GatewayContext,
	listenerSet *gwapiv1.ListenerSet,
	xdsIR resource.XdsIRMap,
	resources *resource.Resources,
) error {
	return t.translateEnvoyExtensionPolicyForListeners(
		policy,
		gateway,
		xdsIR,
		resources,
		listenerSetPolicyTargetListeners(gateway, listenerSet, target),
	)
}

func (t *Translator) translateEnvoyExtensionPolicyForListeners(
	policy *egv1a1.EnvoyExtensionPolicy,
	gateway *GatewayContext,
	xdsIR resource.XdsIRMap,
	resources *resource.Resources,
	targetListeners []*ListenerContext,
) error {
	var (
		extProcs                                              []ir.ExtProc
		wasms                                                 []ir.Wasm
		luas                                                  []ir.Lua
		dynamicModules                                        []ir.DynamicModule
		wasmFailOpen, extProcFailOpen                         bool
		wasmError, luaError, extProcError, dynamicModuleError error
		errs                                                  error
	)

	if extProcs, extProcError, extProcFailOpen = t.buildExtProcs(policy, resources, gateway); extProcError != nil {
		extProcError = perr.WithMessage(extProcError, "ExtProc")
		errs = errors.Join(errs, extProcError)
	}
	if wasms, wasmError, wasmFailOpen = t.buildWasms(policy, resources); wasmError != nil {
		wasmError = perr.WithMessage(wasmError, "Wasm")
		errs = errors.Join(errs, wasmError)
	}
	if luas, luaError = t.buildLuas(policy, gateway.envoyProxy); luaError != nil {
		luaError = perr.WithMessage(luaError, "Lua")
		errs = errors.Join(errs, luaError)
	}
	if dynamicModules, dynamicModuleError = t.buildDynamicModules(policy, gateway.envoyProxy); dynamicModuleError != nil {
		dynamicModuleError = perr.WithMessage(dynamicModuleError, "DynamicModule")
		errs = errors.Join(errs, dynamicModuleError)
	}

	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]
	listenerNames := sets.New[string]()
	for _, listener := range targetListeners {
		listenerNames.Insert(irListenerName(listener))
	}

	routesWithDirectResponse := sets.New[string]()
	for _, http := range x.HTTP {
		if !listenerNames.Has(http.Name) {
			continue
		}

		// A Policy targeting the specific scope(xRoute rule, xRoute, Gateway
		// listener, ListenerSet listener) wins over a policy targeting a lesser
		// specific scope(Gateway/ListenerSet).
		for _, r := range http.Routes {
			// if already set - there's a specific level policy, so skip
			if r.EnvoyExtensions != nil {
				continue
			}

			failRoute := false
			// Lua extension doesn't have a fail open option, so fail the route if there is a lua error
			// TODO: we may also add fail open option for Lua extension to align with other extensions
			if luaError != nil {
				failRoute = true
			}
			if wasmError != nil {
				failRoute = failRoute || !wasmFailOpen
			}
			if extProcError != nil {
				failRoute = failRoute || !extProcFailOpen
			}
			if dynamicModuleError != nil {
				failRoute = true
			}
			if failRoute {
				r.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(r.Name)
			} else {
				r.EnvoyExtensions = &ir.EnvoyExtensionFeatures{
					ExtProcs:       extProcs,
					Wasms:          wasms,
					Luas:           luas,
					DynamicModules: dynamicModules,
				}
			}
		}
	}
	if len(routesWithDirectResponse) > 0 {
		t.Logger.Info("setting 500 direct response in routes due to errors in EnvoyExtensionPolicy",
			"policy", fmt.Sprintf("%s/%s", policy.Namespace, policy.Name),
			"routes", sets.List(routesWithDirectResponse),
			"error", errs,
		)
	}

	return errs
}

func (t *Translator) buildLuas(
	policy *egv1a1.EnvoyExtensionPolicy,
	envoyProxy *egv1a1.EnvoyProxy,
) ([]ir.Lua, error) {
	if policy == nil {
		return nil, nil
	}

	// If Lua EnvoyExtensionPolicies are disabled, skip building Lua filters.
	if len(policy.Spec.Lua) > 0 && t.LuaEnvoyExtensionPolicyDisabled {
		return nil, fmt.Errorf("Skipping Lua EnvoyExtensionPolicy as feature is disabled in the Gateway")
	}

	luaIRList := make([]ir.Lua, 0, len(policy.Spec.Lua))

	for idx, ep := range policy.Spec.Lua {
		name := irConfigNameForLua(policy, idx)
		luaIR, err := t.buildLua(name, policy, ep, envoyProxy)
		if err != nil {
			return nil, err
		}
		luaIRList = append(luaIRList, *luaIR)
	}
	return luaIRList, nil
}

func (t *Translator) buildLua(
	name string,
	policy *egv1a1.EnvoyExtensionPolicy,
	lua egv1a1.Lua,
	envoyProxy *egv1a1.EnvoyProxy,
) (*ir.Lua, error) {
	var luaCode *string
	var err error
	if lua.Type == egv1a1.LuaValueTypeValueRef {
		luaCode, err = t.getLuaBodyFromLocalObjectReference(lua.ValueRef, policy.Namespace)
	} else {
		luaCode = lua.Inline
	}
	if err != nil {
		return nil, err
	}

	if err = luavalidator.NewLuaValidator(*luaCode, envoyProxy).Validate(); err != nil {
		return nil, fmt.Errorf("validation failed for lua body in policy with name %v: %w", name, err)
	}
	return &ir.Lua{
		Name:          name,
		Code:          luaCode,
		FilterContext: lua.FilterContext,
	}, nil
}

// getLuaBodyFromLocalObjectReference assumes the local object reference points to a Kubernetes ConfigMap
func (t *Translator) getLuaBodyFromLocalObjectReference(
	valueRef *gwapiv1.LocalObjectReference,
	policyNs string,
) (*string, error) {
	cm := t.GetConfigMap(policyNs, string(valueRef.Name))
	if cm != nil {
		b, dataOk := cm.Data[LuaConfigMapKey]
		switch {
		case dataOk:
			return &b, nil
		case len(cm.Data) > 0: // Fallback to the first key if lua is not found
			for _, value := range cm.Data {
				b = value
				break
			}
			return &b, nil
		default:
			return nil, fmt.Errorf("can't find the key %s in the referenced configmap %s", LuaConfigMapKey, valueRef.Name)
		}

	} else {
		return nil, fmt.Errorf("can't find the referenced configmap %s in namespace %s", valueRef.Name, policyNs)
	}
}

func (t *Translator) buildExtProcs(policy *egv1a1.EnvoyExtensionPolicy, resources *resource.Resources, gtwCtx *GatewayContext) ([]ir.ExtProc, error, bool) {
	var (
		failOpen bool
		errs     error
	)

	if policy == nil {
		return nil, nil, failOpen
	}

	extProcIRList := make([]ir.ExtProc, 0, len(policy.Spec.ExtProc))

	hasFailClose := false
	for idx, ep := range policy.Spec.ExtProc {
		name := irConfigNameForExtProc(policy, idx)
		extProcIR, err := t.buildExtProc(name, policy, &ep, idx, resources, gtwCtx)
		if err != nil {
			errs = errors.Join(errs, err)
			if ep.FailOpen == nil || !*ep.FailOpen {
				hasFailClose = true
			}
			continue
		}
		extProcIRList = append(extProcIRList, *extProcIR)
	}

	// If any failed ExtProcs are not fail open, the whole policy is not fail open.
	if errs != nil && !hasFailClose {
		failOpen = true
	}
	return extProcIRList, errs, failOpen
}

func (t *Translator) buildExtProc(
	name string,
	policy *egv1a1.EnvoyExtensionPolicy,
	extProc *egv1a1.ExtProc,
	extProcIdx int,
	resources *resource.Resources,
	gtwCtx *GatewayContext,
) (*ir.ExtProc, error) {
	var (
		rd        *ir.RouteDestination
		authority string
		err       error
	)

	if rd, err = t.translateExtServiceBackendRefs(policy, extProc.BackendRefs, ir.GRPC, resources, gtwCtx, "extproc", extProcIdx); err != nil {
		return nil, err
	}

	if extProc.BackendRefs[0].Port != nil {
		authority = fmt.Sprintf(
			"%s.%s:%d",
			extProc.BackendRefs[0].Name,
			NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policy.Namespace),
			*extProc.BackendRefs[0].Port)
	} else {
		authority = fmt.Sprintf(
			"%s.%s",
			extProc.BackendRefs[0].Name,
			NamespaceDerefOr(extProc.BackendRefs[0].Namespace, policy.Namespace))
	}

	traffic, err := translateTrafficFeatures(extProc.BackendSettings)
	if err != nil {
		return nil, err
	}

	extProcIR := &ir.ExtProc{
		Name:        name,
		Destination: *rd,
		Traffic:     traffic,
		Authority:   authority,
	}

	if extProc.MessageTimeout != nil {
		d, err := time.ParseDuration(string(*extProc.MessageTimeout))
		if err != nil {
			return nil, fmt.Errorf("invalid ExtProc MessageTimeout value %v", extProc.MessageTimeout)
		}
		extProcIR.MessageTimeout = ir.MetaV1DurationPtr(d)
	}

	if extProc.ShadowMode != nil {
		extProcIR.ShadowMode = extProc.ShadowMode
	}

	if extProc.FailOpen != nil {
		extProcIR.FailOpen = extProc.FailOpen
	}

	if extProc.StatusOnError != nil {
		extProcIR.StatusOnError = extProc.StatusOnError
	}

	if extProc.ProcessingMode != nil {
		if extProc.ProcessingMode.Request != nil {
			extProcIR.RequestHeaderProcessing = true
			if extProc.ProcessingMode.Request.Body != nil {
				extProcIR.RequestBodyProcessingMode = new(ir.ExtProcBodyProcessingMode(*extProc.ProcessingMode.Request.Body))
			}

			if extProc.ProcessingMode.Request.Attributes != nil {
				extProcIR.RequestAttributes = append(extProcIR.RequestAttributes, extProc.ProcessingMode.Request.Attributes...)
			}
		}

		if extProc.ProcessingMode.Response != nil {
			extProcIR.ResponseHeaderProcessing = true
			if extProc.ProcessingMode.Response.Body != nil {
				extProcIR.ResponseBodyProcessingMode = new(ir.ExtProcBodyProcessingMode(*extProc.ProcessingMode.Response.Body))
			}

			if extProc.ProcessingMode.Response.Attributes != nil {
				extProcIR.ResponseAttributes = append(extProcIR.ResponseAttributes, extProc.ProcessingMode.Response.Attributes...)
			}
		}
		extProcIR.AllowModeOverride = extProc.ProcessingMode.AllowModeOverride
	}

	if extProc.Metadata != nil {
		if extProc.Metadata.AccessibleNamespaces != nil {
			extProcIR.ForwardingMetadataNamespaces = append(extProcIR.ForwardingMetadataNamespaces,
				extProc.Metadata.AccessibleNamespaces...)
		}

		if extProc.Metadata.WritableNamespaces != nil {
			extProcIR.ReceivingMetadataNamespaces = append(extProcIR.ReceivingMetadataNamespaces,
				extProc.Metadata.WritableNamespaces...)
		}
	}

	return extProcIR, err
}

func irConfigNameForExtProc(policy *egv1a1.EnvoyExtensionPolicy, index int) string {
	return fmt.Sprintf(
		"%s/extproc/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

func irConfigNameForLua(policy *egv1a1.EnvoyExtensionPolicy, index int) string {
	return fmt.Sprintf(
		"%s/lua/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

func (t *Translator) buildWasms(
	policy *egv1a1.EnvoyExtensionPolicy,
	resources *resource.Resources,
) ([]ir.Wasm, error, bool) {
	var (
		failOpen bool
		errs     error
	)

	if len(policy.Spec.Wasm) == 0 {
		return nil, nil, failOpen
	}

	wasmIRList := make([]ir.Wasm, 0, len(policy.Spec.Wasm))

	if t.WasmCache == nil {
		return nil, fmt.Errorf("wasm cache is not initialized"), failOpen
	}

	if policy == nil {
		return nil, nil, failOpen
	}

	hasFailClose := false
	for idx, wasm := range policy.Spec.Wasm {
		name := irConfigNameForWasm(policy, idx)
		wasmIR, err := t.buildWasm(name, &wasm, policy, idx, resources)
		if err != nil {
			errs = errors.Join(errs, err)
			if wasm.FailOpen == nil || !*wasm.FailOpen {
				hasFailClose = true
			}
			continue
		}
		wasmIRList = append(wasmIRList, *wasmIR)
	}

	// If any failed ExtProcs are not fail open, the whole policy is not fail open.
	if errs != nil && !hasFailClose {
		failOpen = true
	}

	return wasmIRList, errs, failOpen
}

func (t *Translator) buildWasm(
	name string,
	config *egv1a1.Wasm,
	policy *egv1a1.EnvoyExtensionPolicy,
	idx int,
	resources *resource.Resources,
) (*ir.Wasm, error) {
	var (
		failOpen   = false
		code       *ir.HTTPWasmCode
		pullPolicy wasm.PullPolicy
		// the checksum provided by the user, it's used to validate the wasm module
		// downloaded from the original HTTP server or the OCI registry
		originalChecksum string
		servingURL       string // the wasm module download URL from the EG HTTP server
		caCert           []byte
		err              error
	)

	if config.FailOpen != nil {
		failOpen = *config.FailOpen
	}

	if config.Code.PullPolicy != nil {
		switch *config.Code.PullPolicy {
		case egv1a1.ImagePullPolicyAlways:
			pullPolicy = wasm.Always
		case egv1a1.ImagePullPolicyIfNotPresent:
			pullPolicy = wasm.IfNotPresent
		default:
			pullPolicy = wasm.Unspecified
		}
	}

	switch config.Code.Type {
	case egv1a1.HTTPWasmCodeSourceType:
		var checksum string

		// This is a sanity check, the validation should have caught this
		if config.Code.HTTP == nil {
			return nil, fmt.Errorf("missing HTTP field in Wasm code source")
		}

		if config.Code.HTTP.SHA256 != nil {
			originalChecksum = *config.Code.HTTP.SHA256
		}

		http := config.Code.HTTP

		if http.TLS != nil {
			from := crossNamespaceFrom{
				group:     egv1a1.GroupName,
				kind:      resource.KindEnvoyExtensionPolicy,
				namespace: policy.Namespace,
			}
			if caCert, err = t.validateAndGetDataAtKeyInRef(http.TLS.CACertificateRef, "ca.crt", resources, from); err != nil {
				return nil, err
			}
		}

		if servingURL, checksum, err = t.WasmCache.Get(http.URL, &wasm.GetOptions{
			Checksum:        originalChecksum,
			PullPolicy:      pullPolicy,
			ResourceName:    irConfigNameForWasm(policy, idx),
			ResourceVersion: policy.ResourceVersion,
			CACert:          caCert,
		}); err != nil {
			return nil, err
		}

		code = &ir.HTTPWasmCode{
			ServingURL:  servingURL,
			OriginalURL: http.URL,
			SHA256:      checksum,
		}

	case egv1a1.ImageWasmCodeSourceType:
		var (
			image      = config.Code.Image
			secret     *corev1.Secret
			pullSecret []byte
			// the checksum of the wasm module extracted from the OCI image
			// it's different from the checksum for the OCI image
			checksum string
		)

		// This is a sanity check, the validation should have caught this
		if image == nil {
			return nil, fmt.Errorf("missing Image field in Wasm code source")
		}

		if image.TLS != nil {
			from := crossNamespaceFrom{
				group:     egv1a1.GroupName,
				kind:      resource.KindEnvoyExtensionPolicy,
				namespace: policy.Namespace,
			}
			if caCert, err = t.validateAndGetDataAtKeyInRef(image.TLS.CACertificateRef, "ca.crt", resources, from); err != nil {
				return nil, err
			}
		}

		if image.PullSecretRef != nil {
			from := crossNamespaceFrom{
				group:     egv1a1.GroupName,
				kind:      resource.KindEnvoyExtensionPolicy,
				namespace: policy.Namespace,
			}

			if secret, err = t.validateSecretRef(
				true, from, *image.PullSecretRef, resources); err != nil {
				return nil, err
			}

			if data, ok := secret.Data[corev1.DockerConfigJsonKey]; ok {
				pullSecret = data
			} else {
				return nil, fmt.Errorf("missing %s key in secret %s/%s", corev1.DockerConfigJsonKey, secret.Namespace, secret.Name)
			}
		}

		// Wasm Cache requires the URL to be in the format "scheme://<URL>"
		imageURL := image.URL
		if !strings.HasPrefix(image.URL, ociURLPrefix) {
			imageURL = fmt.Sprintf("%s%s", ociURLPrefix, image.URL)
		}

		// If the url is an OCI image, and neither digest nor tag is provided, use the latest tag.
		if !hasDigest(imageURL) && !hasTag(imageURL) {
			imageURL += ":latest"
		}

		if config.Code.Image.SHA256 != nil {
			originalChecksum = *config.Code.Image.SHA256
		}

		// The wasm checksum is different from the OCI image digest.
		// The original checksum in the EEP is used to match the digest of OCI image.
		// The returned checksum from the cache is the checksum of the wasm file
		// extracted from the OCI image, which is used by the envoy to verify the wasm file.
		if servingURL, checksum, err = t.WasmCache.Get(imageURL, &wasm.GetOptions{
			Checksum:        originalChecksum,
			PullSecret:      pullSecret,
			PullPolicy:      pullPolicy,
			ResourceName:    irConfigNameForWasm(policy, idx),
			ResourceVersion: policy.ResourceVersion,
			CACert:          caCert,
		}); err != nil {
			return nil, err
		}

		code = &ir.HTTPWasmCode{
			ServingURL:  servingURL,
			SHA256:      checksum,
			OriginalURL: imageURL,
		}
	default:
		// should never happen because of kubebuilder validation, just a sanity check
		return nil, fmt.Errorf("unsupported Wasm code source type %q", config.Code.Type)
	}

	wasmName := name
	if config.Name != nil {
		wasmName = *config.Name
	}
	wasmIR := &ir.Wasm{
		Name:     name,
		RootID:   config.RootID,
		WasmName: wasmName,
		Config:   config.Config,
		FailOpen: failOpen,
		Code:     code,
	}

	if config.Env != nil && len(config.Env.HostKeys) > 0 {
		wasmIR.HostKeys = config.Env.HostKeys
	}

	return wasmIR, nil
}

func hasDigest(imageURL string) bool {
	return strings.Contains(imageURL, "@")
}

func hasTag(imageURL string) bool {
	parts := strings.Split(imageURL[len(ociURLPrefix):], ":")
	// Verify that we aren't confusing a tag for a hostname with port.
	return len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/")
}

func irConfigNameForWasm(policy client.Object, index int) string {
	return fmt.Sprintf(
		"%s/wasm/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

func irConfigNameForDynamicModule(policy *egv1a1.EnvoyExtensionPolicy, index int) string {
	return fmt.Sprintf(
		"%s/dynamic-module/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

func (t *Translator) buildDynamicModules(
	policy *egv1a1.EnvoyExtensionPolicy,
	envoyProxy *egv1a1.EnvoyProxy,
) ([]ir.DynamicModule, error) {
	var errs error

	if policy == nil || len(policy.Spec.DynamicModule) == 0 {
		return nil, nil
	}

	// Build registry lookup map from EnvoyProxy
	registry := make(map[string]*egv1a1.DynamicModuleEntry)
	if envoyProxy != nil {
		for i := range envoyProxy.Spec.DynamicModules {
			entry := &envoyProxy.Spec.DynamicModules[i]
			registry[entry.Name] = entry
		}
	}

	dmIRList := make([]ir.DynamicModule, 0, len(policy.Spec.DynamicModule))

	for idx, dm := range policy.Spec.DynamicModule {
		name := irConfigNameForDynamicModule(policy, idx)

		// Validate module exists in registry
		entry, ok := registry[dm.Name]
		if !ok {
			errs = errors.Join(errs, fmt.Errorf("dynamic module %q is not registered in the EnvoyProxy dynamicModules allowlist", dm.Name))
			continue
		}

		filterName := ptr.Deref(dm.FilterName, "")

		dmIR := ir.DynamicModule{
			Name:           name,
			FilterName:     filterName,
			Config:         dm.Config,
			DoNotClose:     ptr.Deref(entry.DoNotClose, false),
			LoadGlobally:   ptr.Deref(entry.LoadGlobally, false),
			TerminalFilter: ptr.Deref(dm.TerminalFilter, false),
		}

		switch sourceType := ptr.Deref(entry.Source.Type, egv1a1.LocalDynamicModuleSourceType); sourceType {
		case egv1a1.RemoteDynamicModuleSourceType:
			if entry.Source.Remote == nil {
				errs = errors.Join(errs, fmt.Errorf("dynamic module %q has no remote source configured", dm.Name))
				continue
			}
			if entry.Source.Remote.URL == "" {
				errs = errors.Join(errs, fmt.Errorf("dynamic module %q has no remote source URL configured", dm.Name))
				continue
			}
			if entry.Source.Remote.SHA256 == "" {
				errs = errors.Join(errs, fmt.Errorf("dynamic module %q has no remote source SHA256 configured", dm.Name))
				continue
			}
			if err := validateDynamicModuleRemoteURL(entry.Source.Remote.URL); err != nil {
				errs = errors.Join(errs, fmt.Errorf("dynamic module %q has invalid remote source URL %q: %w", dm.Name, entry.Source.Remote.URL, err))
				continue
			}
			dmIR.Remote = &ir.RemoteDynamicModuleSource{
				URL:    entry.Source.Remote.URL,
				SHA256: entry.Source.Remote.SHA256,
			}
		case egv1a1.LocalDynamicModuleSourceType:
			if entry.Source.Local == nil {
				errs = errors.Join(errs, fmt.Errorf("dynamic module %q has no local source configured", dm.Name))
				continue
			}
			dmIR.Path = entry.Source.Local.Path
		default:
			errs = errors.Join(errs, fmt.Errorf("dynamic module %q has unsupported source type %q", dm.Name, sourceType))
			continue
		}

		dmIRList = append(dmIRList, dmIR)
	}

	return dmIRList, errs
}
