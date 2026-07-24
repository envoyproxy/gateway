// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"math"
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
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	egv1a1validation "github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/ratelimit"
	"github.com/envoyproxy/gateway/internal/utils/regex"
)

const (
	MaxConsistentHashTableSize = 5000011 // https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-cluster-maglevlbconfig
	// ResponseBodyConfigMapKey is the key used in ConfigMaps to store custom response body data
	ResponseBodyConfigMapKey = "response.body"
)

// btpRoutingKey identifies a BTP routing type target
type btpRoutingKey struct {
	Kind, Namespace, Name, SectionName string
}

// BTPRoutingTypeIndex holds RoutingType values from BackendTrafficPolicies
// This avoids an O(BTPs) lookup for every iteration of processDestination.
type BTPRoutingTypeIndex struct {
	routeRuleLevel           map[btpRoutingKey]*egv1a1.RoutingType
	routeLevel               map[btpRoutingKey]*egv1a1.RoutingType
	listenerSetListenerLevel map[btpRoutingKey]*egv1a1.RoutingType
	listenerSetLevel         map[btpRoutingKey]*egv1a1.RoutingType
	listenerLevel            map[btpRoutingKey]*egv1a1.RoutingType
	gatewayLevel             map[btpRoutingKey]*egv1a1.RoutingType
}

// BuildBTPRoutingTypeIndex builds a pre-computed index of RoutingType values
// from BackendTrafficPolicies, organized by priority-level.
// BTPs are pre-sorted by the provider layer, so first-write-wins respects priority.
func hasBTPRoutingType(btps []*egv1a1.BackendTrafficPolicy) bool {
	for _, btp := range btps {
		if btp.Spec.RoutingType != nil {
			return true
		}
	}

	return false
}

func BuildBTPRoutingTypeIndex(
	btps []*egv1a1.BackendTrafficPolicy,
	routes []client.Object,
	gateways []*GatewayContext,
	listenerSets []*gwapiv1.ListenerSet,
	referenceGrants []*gwapiv1b1.ReferenceGrant,
	namespaceLookup func(string) *corev1.Namespace,
) *BTPRoutingTypeIndex {
	idx := &BTPRoutingTypeIndex{
		routeRuleLevel:           make(map[btpRoutingKey]*egv1a1.RoutingType),
		routeLevel:               make(map[btpRoutingKey]*egv1a1.RoutingType),
		listenerSetListenerLevel: make(map[btpRoutingKey]*egv1a1.RoutingType),
		listenerSetLevel:         make(map[btpRoutingKey]*egv1a1.RoutingType),
		listenerLevel:            make(map[btpRoutingKey]*egv1a1.RoutingType),
		gatewayLevel:             make(map[btpRoutingKey]*egv1a1.RoutingType),
	}

	// Combine supported targets into a single target slice for target resolution.
	allTargets := make([]client.Object, 0, len(routes)+len(gateways)+len(listenerSets))
	allTargets = append(allTargets, routes...)
	for _, gw := range gateways {
		allTargets = append(allTargets, gw)
	}
	for _, ls := range listenerSets {
		allTargets = append(allTargets, ls)
	}

	for _, btp := range btps {
		if btp.Spec.RoutingType == nil {
			continue
		}
		refs := resolvePolicyTargets(
			btp.Spec.PolicyTargetReferences,
			allTargets,
			referenceGrants,
			egv1a1.GroupName,
			egv1a1.KindBackendTrafficPolicy,
			btp.Namespace,
			namespaceLookup,
		)
		for _, ref := range refs {
			kind := string(ref.Kind)
			key := btpRoutingKey{
				Kind:        kind,
				Namespace:   string(ref.Namespace),
				Name:        string(ref.Name),
				SectionName: string(ptr.Deref(ref.SectionName, "")),
			}

			switch kind {
			case resource.KindGateway:
				if ref.SectionName != nil {
					if _, exists := idx.listenerLevel[key]; !exists {
						idx.listenerLevel[key] = btp.Spec.RoutingType
					}
				} else {
					if _, exists := idx.gatewayLevel[key]; !exists {
						idx.gatewayLevel[key] = btp.Spec.RoutingType
					}
				}
			case resource.KindListenerSet:
				if ref.SectionName != nil {
					if _, exists := idx.listenerSetListenerLevel[key]; !exists {
						idx.listenerSetListenerLevel[key] = btp.Spec.RoutingType
					}
				} else {
					if _, exists := idx.listenerSetLevel[key]; !exists {
						idx.listenerSetLevel[key] = btp.Spec.RoutingType
					}
				}
			default:
				if ref.SectionName != nil {
					if _, exists := idx.routeRuleLevel[key]; !exists {
						idx.routeRuleLevel[key] = btp.Spec.RoutingType
					}
				} else {
					if _, exists := idx.routeLevel[key]; !exists {
						idx.routeLevel[key] = btp.Spec.RoutingType
					}
				}
			}
		}
	}

	return idx
}

// LookupBTPRoutingType resolves the RoutingType for a specific route rule
// and gateway/listener combination by checking the index in
// priority order: routeRule > route > listener > gateway.
// Returns nil if no matching BTP RoutingType is found, or if the index is nil.
func (idx *BTPRoutingTypeIndex) LookupBTPRoutingType(
	routeKind gwapiv1.Kind,
	routeNN types.NamespacedName,
	gatewayNN types.NamespacedName,
	listenerName *gwapiv1.SectionName,
	listenerSetNN *types.NamespacedName,
	routeRuleName *gwapiv1.SectionName,
) *egv1a1.RoutingType {
	if idx == nil {
		return nil
	}

	// 1. Route-rule level (most specific)
	if routeRuleName != nil {
		key := btpRoutingKey{
			Kind:        string(routeKind),
			Namespace:   routeNN.Namespace,
			Name:        routeNN.Name,
			SectionName: string(*routeRuleName),
		}
		if rt, ok := idx.routeRuleLevel[key]; ok {
			return rt
		}
	}

	// 2. Route level
	routeKey := btpRoutingKey{
		Kind:      string(routeKind),
		Namespace: routeNN.Namespace,
		Name:      routeNN.Name,
	}
	if rt, ok := idx.routeLevel[routeKey]; ok {
		return rt
	}

	// 3. ListenerSet listener level, then ListenerSet level for routes attached through a ListenerSet.
	if listenerSetNN != nil {
		if listenerName != nil {
			listenerSetListenerKey := btpRoutingKey{
				Kind:        resource.KindListenerSet,
				Namespace:   listenerSetNN.Namespace,
				Name:        listenerSetNN.Name,
				SectionName: string(*listenerName),
			}
			if rt, ok := idx.listenerSetListenerLevel[listenerSetListenerKey]; ok {
				return rt
			}
		}

		listenerSetKey := btpRoutingKey{
			Kind:      resource.KindListenerSet,
			Namespace: listenerSetNN.Namespace,
			Name:      listenerSetNN.Name,
		}
		if rt, ok := idx.listenerSetLevel[listenerSetKey]; ok {
			return rt
		}
	}

	// 4. Gateway listener level. ListenerSet-attached routes intentionally skip
	// Gateway listener policy lookup because Gateway listeners and ListenerSet
	// listeners are sibling scopes.
	if listenerSetNN == nil && listenerName != nil {
		listenerKey := btpRoutingKey{
			Kind:        resource.KindGateway,
			Namespace:   gatewayNN.Namespace,
			Name:        gatewayNN.Name,
			SectionName: string(*listenerName),
		}
		if rt, ok := idx.listenerLevel[listenerKey]; ok {
			return rt
		}
	}

	// 5. Gateway level (least specific)
	gwKey := btpRoutingKey{
		Kind:      resource.KindGateway,
		Namespace: gatewayNN.Namespace,
		Name:      gatewayNN.Name,
	}
	if rt, ok := idx.gatewayLevel[gwKey]; ok {
		return rt
	}

	return nil
}

// deprecatedFieldsUsedInBackendTrafficPolicy returns a map of deprecated field paths to their alternatives.
func deprecatedFieldsUsedInBackendTrafficPolicy(policy *egv1a1.BackendTrafficPolicy) map[string]string {
	deprecatedFields := make(map[string]string)
	if policy.Spec.TargetRef != nil {
		deprecatedFields["spec.targetRef"] = "spec.targetRefs"
	}
	if len(policy.Spec.Compression) > 0 {
		deprecatedFields["spec.compression"] = "spec.compressor"
	}
	return deprecatedFields
}

func (t *Translator) ProcessBackendTrafficPolicies(
	resources *resource.Resources,
	gateways []*GatewayContext,
	routes []RouteContext,
	xdsIR resource.XdsIRMap,
) []*egv1a1.BackendTrafficPolicy {
	backendTrafficPolicies := resources.BackendTrafficPolicies
	// BackendTrafficPolicies are already sorted by the provider layer

	routeMapSize := len(routes)
	gatewayMapSize := len(gateways)
	policyMapSize := len(backendTrafficPolicies)
	listenerSetMapSize := len(resources.ListenerSets)

	res := make([]*egv1a1.BackendTrafficPolicy, 0, policyMapSize)

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
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

	// Map of attached Policy to Gateway. It is used to merge policies process.
	gatewayPolicyMap := make(map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy, gatewayMapSize)

	// Map of attached Policy to ListenerSet. It is used for merge policy processing.
	listenerSetPolicyMap := make(map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy, listenerSetMapSize)

	// overrides records child scopes whose policies displace policies attached
	// to their parent scopes.
	overrides := newPolicyScopeGraph()

	// merged records Route scopes whose policies were merged into policies
	// attached to their parent scopes.
	merged := newPolicyScopeGraph()

	handledPolicies := make(map[types.NamespacedName]*egv1a1.BackendTrafficPolicy, policyMapSize)

	// Translate
	// 1. First translate Policies targeting RouteRules
	// 2. Next translate Policies targeting xRoutes
	// 3. Then translate Policies targeting ListenerSet Listeners
	// 4. Then translate Policies targeting ListenerSets
	// 5. Then translate Policies targeting Gateway Listeners
	// 6. Finally, the policies targeting Gateways

	// Build gateway policy maps, which are needed when processing the policies targeting xRoutes.
	t.buildGatewayPolicyMap(backendTrafficPolicies, gateways, gatewayMap, gatewayPolicyMap, resources.ReferenceGrants)
	// Build ListenerSet policy maps, which are needed when processing the policies targeting xRoutes.
	t.buildListenerSetBackendTrafficPolicyMap(backendTrafficPolicies, listenerSetMap, listenerSetPolicyMap, resources)

	// Process the policies targeting RouteRules
	for i, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isRouteRule(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = backendTrafficPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processBackendTrafficPolicyForRoute(xdsIR,
					routeMap, listenerSetMap, gatewayPolicyMap, listenerSetPolicyMap, overrides, merged, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Routes
	for i, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			routes,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindBackendTrafficPolicy,
			currPolicy.Namespace,
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isRoute(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = backendTrafficPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processBackendTrafficPolicyForRoute(xdsIR,
					routeMap, listenerSetMap, gatewayPolicyMap, listenerSetPolicyMap, overrides, merged, policy, currTarget)
			}
		}
	}

	// Process the policies targeting ListenerSet Listeners
	for i, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isListenerSetListener(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = policyCopies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}
				t.processBackendTrafficPolicyForListenerSet(xdsIR,
					gatewayMap, listenerSetMap, overrides, merged, policy, currTarget)
			}
		}
	}

	// Process the policies targeting ListenerSets
	for i, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			resources.ListenerSets,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindBackendTrafficPolicy,
			currPolicy.Namespace,
			t.GetNamespace,
		)
		for _, currTarget := range targetRefs {
			if isListenerSet(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = policyCopies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}
				t.processBackendTrafficPolicyForListenerSet(xdsIR,
					gatewayMap, listenerSetMap, overrides, merged, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Gateway Listeners
	for i, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isListener(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = backendTrafficPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}
				t.processBackendTrafficPolicyForGateway(xdsIR,
					gatewayMap, overrides, merged, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Gateways
	for i, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			gateways,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindBackendTrafficPolicy,
			currPolicy.Namespace,
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isGateway(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = backendTrafficPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}
				t.processBackendTrafficPolicyForGateway(xdsIR,
					gatewayMap, overrides, merged, policy, currTarget)
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

func (t *Translator) buildGatewayPolicyMap(
	backendTrafficPolicies []*egv1a1.BackendTrafficPolicy,
	gateways []*GatewayContext,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	gatewayPolicyMap map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy,
	referenceGrants []*gwapiv1b1.ReferenceGrant,
) {
	for _, currPolicy := range backendTrafficPolicies {
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			gateways,
			referenceGrants,
			egv1a1.GroupName,
			egv1a1.KindBackendTrafficPolicy,
			currPolicy.Namespace,
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if currTarget.Kind == resource.KindGateway {
				// Check if the gateway exists
				key := types.NamespacedName{
					Name:      string(currTarget.Name),
					Namespace: string(currTarget.Namespace),
				}
				gateway, ok := gatewayMap[key]
				if !ok {
					continue
				}

				// Check if the specified listener exists when sectionName is set
				if currTarget.SectionName != nil {
					if err := validateGatewayListenerSectionName(
						*currTarget.SectionName,
						key,
						gatewayDirectListeners(gateway.GatewayContext),
					); err != nil {
						continue
					}
				}

				mapKey := NamespacedNameWithSection{
					NamespacedName: key,
					SectionName:    ptr.Deref(currTarget.SectionName, ""),
				}
				if _, ok := gatewayPolicyMap[mapKey]; ok {
					continue
				}
				gatewayPolicyMap[mapKey] = currPolicy
			}
		}
	}
}

// buildListenerSetBackendTrafficPolicyMap populates listenerSetPolicyMap with the
// first BackendTrafficPolicy attached to each (ListenerSet, sectionName) pair.
// Subsequent conflicting attachments are reported elsewhere; this map is the
// source of truth used by the merge step to find the closest parent policy
// for routes attached via a ListenerSet.
func (t *Translator) buildListenerSetBackendTrafficPolicyMap(
	backendTrafficPolicies []*egv1a1.BackendTrafficPolicy,
	listenerSetMap map[types.NamespacedName]*policyListenerSetTargetContext,
	listenerSetPolicyMap map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy,
	resources *resource.Resources,
) {
	for _, currPolicy := range backendTrafficPolicies {
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			resources.ListenerSets,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindBackendTrafficPolicy,
			currPolicy.Namespace,
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if currTarget.Kind != resource.KindListenerSet {
				continue
			}

			key := types.NamespacedName{
				Name:      string(currTarget.Name),
				Namespace: string(currTarget.Namespace),
			}
			ls, ok := listenerSetMap[key]
			if !ok {
				continue
			}

			if currTarget.SectionName != nil {
				if err := validateListenerSetListenerSectionName(
					*currTarget.SectionName,
					key,
					ls.Spec.Listeners,
				); err != nil {
					continue
				}
			}

			mapKey := NamespacedNameWithSection{
				NamespacedName: key,
				SectionName:    ptr.Deref(currTarget.SectionName, ""),
			}
			if _, ok := listenerSetPolicyMap[mapKey]; ok {
				continue
			}
			listenerSetPolicyMap[mapKey] = currPolicy
		}
	}
}

func (t *Translator) processBackendTrafficPolicyForRoute(
	xdsIR resource.XdsIRMap,
	routeMap map[policyTargetRouteKey]*policyRouteTargetContext,
	listenerSetMap map[types.NamespacedName]*policyListenerSetTargetContext,
	gatewayPolicyMap map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy,
	listenerSetPolicyMap map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy,
	overrides policyScopeGraph,
	merged policyScopeGraph,
	policy *egv1a1.BackendTrafficPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targetedRoute RouteContext
		resolveErr    *status.PolicyResolveError
	)

	targetedRoute, resolveErr = resolveBackendTrafficPolicyRouteTargetRef(currTarget, routeMap)
	// Skip if the route is not found
	// It's not necessarily an error because the BackendTrafficPolicy may be
	// reconciled by multiple controllers. And the other controller may
	// have the target route.
	if targetedRoute == nil {
		return
	}

	// Collect the route's parent refs for policy status and merge handling.
	// At the same time, record this Route scope under each parent attachment
	// scope whose policy it can override.
	parentRefs := GetManagedParentReferences(targetedRoute)
	ancestorRefs := make([]*gwapiv1.ParentReference, 0, len(parentRefs))
	// parentRefCtxs holds parent gateway/listener contexts for using in policy merge logic.
	parentRefCtxs := make([]*RouteParentContext, 0, len(parentRefs))
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

			// Do need a section name since the policy is targeting to a route.
			ancestorRef := getAncestorRefForPolicy(parentNN, p.SectionName)
			ancestorRefs = append(ancestorRefs, &ancestorRef)
			if parentRefCtx := targetedRoute.GetRouteParentContext(p); parentRefCtx != nil {
				parentRefCtxs = append(parentRefCtxs, parentRefCtx)
			}
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

			// Record at the most-specific LS scope.
			if p.SectionName != nil {
				overrides.Add(listenerSetListenerScope(parentNN, *p.SectionName), routeAsChildScope)
			} else {
				overrides.Add(listenerSetScope(parentNN), routeAsChildScope)
			}

			// ListenerSet-attached Route policies report status against the
			// ListenerSet itself.
			ancestorRef := getAncestorRefForListenerSetPolicy(parentNN, p.SectionName)
			ancestorRefs = append(ancestorRefs, &ancestorRef)
			if parentRefCtx := targetedRoute.GetRouteParentContext(p); parentRefCtx != nil {
				parentRefCtxs = append(parentRefCtxs, parentRefCtx)
			}
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

	if policy.Spec.MergeType == nil {
		// Set conditions for translation error if it got any
		if err := t.translateBackendTrafficPolicyForRoute(policy, targetedRoute, currTarget, xdsIR, nil); err != nil {
			status.SetTranslationErrorForPolicyAncestors(&policy.Status,
				ancestorRefs,
				t.GatewayControllerName,
				policy.Generation,
				status.Error2ConditionMsg(err),
			)
		}
	} else {
		// Merge with the closest policy in the Route's attachment hierarchy.
		// Gateway listeners check the Gateway listener policy first, then the
		// Gateway policy. ListenerSet listeners check the ListenerSet listener
		// policy, then the ListenerSet policy, then the parent Gateway policy;
		// they intentionally skip Gateway listener policies because those are
		// sibling scopes.
		for _, parentRefCtx := range parentRefCtxs {
			for _, listener := range parentRefCtx.listeners {
				gwNN := utils.NamespacedName(listener.gateway.Gateway)
				var (
					ancestorRef  gwapiv1.ParentReference
					parentPolicy *egv1a1.BackendTrafficPolicy
					parentScope  policyScope
				)

				if listener.isFromListenerSet() {
					lsNN := types.NamespacedName{
						Name:      listener.listenerSet.Name,
						Namespace: listener.listenerSet.Namespace,
					}
					ancestorRef = getAncestorRefForListenerSetPolicy(lsNN, &listener.Name)

					lsListenerKey := NamespacedNameWithSection{NamespacedName: lsNN, SectionName: listener.Name}
					lsKey := NamespacedNameWithSection{NamespacedName: lsNN}
					gwKey := NamespacedNameWithSection{NamespacedName: gwNN}

					if p, ok := listenerSetPolicyMap[lsListenerKey]; ok {
						parentPolicy, parentScope = p, listenerSetListenerScope(lsNN, listener.Name)
					} else if p, ok := listenerSetPolicyMap[lsKey]; ok {
						parentPolicy, parentScope = p, listenerSetScope(lsNN)
					} else if p, ok := gatewayPolicyMap[gwKey]; ok {
						parentPolicy, parentScope = p, gatewayScope(gwNN)
					}
				} else {
					ancestorRef = getAncestorRefForPolicy(gwNN, &listener.Name)

					listenerMapKey := NamespacedNameWithSection{NamespacedName: gwNN, SectionName: listener.Name}
					gwMapKey := NamespacedNameWithSection{NamespacedName: gwNN}
					if p, ok := gatewayPolicyMap[listenerMapKey]; ok {
						parentPolicy, parentScope = p, gatewayListenerScope(gwNN, listener.Name)
					} else if p, ok := gatewayPolicyMap[gwMapKey]; ok {
						parentPolicy, parentScope = p, gatewayScope(gwNN)
					}
				}

				if parentPolicy == nil {
					// not found, fall back to the current policy
					if err := t.translateBackendTrafficPolicyForRoute(policy, targetedRoute, currTarget, xdsIR, listener); err != nil {
						status.SetConditionForPolicyAncestor(&policy.Status,
							&ancestorRef,
							t.GatewayControllerName,
							gwapiv1.PolicyConditionAccepted, metav1.ConditionFalse,
							egv1a1.PolicyReasonInvalid,
							status.Error2ConditionMsg(err),
							policy.Generation,
						)
					}
					continue
				}

				// merge with parent policy
				if err := t.translateBackendTrafficPolicyForRouteWithMerge(
					policy, parentPolicy, currTarget, listener, targetedRoute, xdsIR,
				); err != nil {
					status.SetConditionForPolicyAncestor(&policy.Status,
						&ancestorRef,
						t.GatewayControllerName,
						gwapiv1.PolicyConditionAccepted, metav1.ConditionFalse,
						egv1a1.PolicyReasonInvalid,
						status.Error2ConditionMsg(err),
						policy.Generation,
					)
					continue
				}

				// Record the merged route under the parent scope so the parent's
				// status can list the routes that were merged into it.
				merged.Add(parentScope, routeAsChildScope)

				status.SetConditionForPolicyAncestor(&policy.Status,
					&ancestorRef,
					t.GatewayControllerName,
					egv1a1.PolicyConditionMerged,
					metav1.ConditionTrue,
					egv1a1.PolicyReasonMerged,
					fmt.Sprintf("Merged with policy %s/%s", parentPolicy.Namespace, parentPolicy.Name),
					policy.Generation,
				)
			}
		}
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation)

	// Check for deprecated fields and set warning if any are found
	if deprecatedFields := deprecatedFieldsUsedInBackendTrafficPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}

	// Check if this policy is overridden by other policies targeting at route rule levels
	// If policy target is route rule, we can skip the check
	if currTarget.SectionName != nil {
		return
	}

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
			"This policy is being overridden by other backendTrafficPolicy for "+overriddenTargetsMessage,
			policy.Generation,
		)
	}
}

func (t *Translator) processBackendTrafficPolicyForListenerSet(
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	listenerSetMap map[types.NamespacedName]*policyListenerSetTargetContext,
	overrides policyScopeGraph,
	merged policyScopeGraph,
	policy *egv1a1.BackendTrafficPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targeted   *gwapiv1.ListenerSet
		resolveErr *status.PolicyResolveError
	)

	targeted, resolveErr = resolveBackendTrafficPolicyListenerSetTargetRef(currTarget, listenerSetMap)
	// Skip if the ListenerSet is not found. It may be reconciled by another controller.
	if targeted == nil {
		return
	}

	parentGatewayNN := types.NamespacedName{
		Name:      string(targeted.Spec.ParentRef.Name),
		Namespace: NamespaceDerefOr(targeted.Spec.ParentRef.Namespace, targeted.Namespace),
	}
	gateway, ok := gatewayMap[parentGatewayNN]
	// The ListenerSet may exist while its parent Gateway is not in the accepted
	// Gateway set for this translation run.
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

	if err := t.translateBackendTrafficPolicyForListenerSet(policy, gateway.GatewayContext, targeted, currTarget, xdsIR); err != nil {
		status.SetTranslationErrorForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation)

	// Determine this policy's own scope so we can look up routes merged into it
	// and child scopes overriding it.
	var lsParentScope policyScope
	if currTarget.SectionName == nil {
		lsParentScope = listenerSetScope(listenerSetNN)
	} else {
		lsParentScope = listenerSetListenerScope(listenerSetNN, *currTarget.SectionName)
	}

	mergedScopes := merged.GetDirectChildren(lsParentScope)
	mergedMessage := formatPolicyScopes(mergedScopes)
	// Merged routes are excluded from the override message so a route doesn't
	// appear in both sections.
	overriddenMessage := formatPolicyScopes(overrides.GetWithDescendants(lsParentScope).Difference(mergedScopes))
	if mergedMessage != "" {
		status.SetConditionForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			egv1a1.PolicyConditionMerged,
			metav1.ConditionTrue,
			egv1a1.PolicyReasonMerged,
			"This policy is being merged by other backendTrafficPolicies for "+mergedMessage,
			policy.Generation,
		)
	}
	if overriddenMessage != "" {
		status.SetConditionForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			egv1a1.PolicyConditionOverridden,
			metav1.ConditionTrue,
			egv1a1.PolicyReasonOverridden,
			"This policy is being overridden by other backendTrafficPolicies for "+overriddenMessage,
			policy.Generation,
		)
	}

	// Check for deprecated fields and set warning if any are found
	if deprecatedFields := deprecatedFieldsUsedInBackendTrafficPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}
}

func (t *Translator) processBackendTrafficPolicyForGateway(
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	overrides policyScopeGraph,
	merged policyScopeGraph,
	policy *egv1a1.BackendTrafficPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targetedGateway *GatewayContext
		resolveErr      *status.PolicyResolveError
	)

	// Negative statuses have already been assigned so it's safe to skip
	targetedGateway, resolveErr = resolveBackendTrafficPolicyGatewayTargetRef(currTarget, gatewayMap)
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
	if err := t.translateBackendTrafficPolicyForGateway(policy, targetedGateway, currTarget, xdsIR); err != nil {
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
	if deprecatedFields := deprecatedFieldsUsedInBackendTrafficPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}

	// Determine this policy's own scope so we can look up merged and overriding
	// child scopes from the relation maps.
	var parentScope policyScope
	if currTarget.SectionName == nil {
		parentScope = gatewayScope(gatewayNN)
	} else {
		parentScope = gatewayListenerScope(gatewayNN, *currTarget.SectionName)
	}

	mergedScopes := merged.GetDirectChildren(parentScope)
	mergedMessage := formatPolicyScopes(mergedScopes)
	// Merged routes are excluded from the override message so a route doesn't
	// appear in both sections.
	overriddenMessage := formatPolicyScopes(overrides.GetWithDescendants(parentScope).Difference(mergedScopes))

	if mergedMessage != "" {
		status.SetConditionForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			egv1a1.PolicyConditionMerged,
			metav1.ConditionTrue,
			egv1a1.PolicyReasonMerged,
			"This policy is being merged by other backendTrafficPolicies for "+mergedMessage,
			policy.Generation,
		)
	}
	if overriddenMessage != "" {
		status.SetConditionForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			egv1a1.PolicyConditionOverridden,
			metav1.ConditionTrue,
			egv1a1.PolicyReasonOverridden,
			"This policy is being overridden by other backendTrafficPolicies for "+overriddenMessage,
			policy.Generation,
		)
	}
}

func resolveBackendTrafficPolicyGatewayTargetRef(
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

	// If sectionName is set, make sure it's valid
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
			message := fmt.Sprintf("Unable to target Gateway %s, another BackendTrafficPolicy has already attached to it",
				string(target.Name))

			return gateway.GatewayContext, &status.PolicyResolveError{
				Reason:  gwapiv1.PolicyReasonConflicted,
				Message: message,
			}
		}

		// Set context and save
		gateway.attached = true
	} else {
		listenerName := string(*target.SectionName)
		if gateway.attachedToListeners != nil && gateway.attachedToListeners.Has(listenerName) {
			message := fmt.Sprintf("Unable to target Listener %s/%s, another BackendTrafficPolicy has already attached to it",
				key, listenerName)

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

func resolveBackendTrafficPolicyListenerSetTargetRef(
	target policyTargetReferenceWithSectionName,
	listenerSets map[types.NamespacedName]*policyListenerSetTargetContext,
) (*gwapiv1.ListenerSet, *status.PolicyResolveError) {
	// Find the ListenerSet
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: string(target.Namespace),
	}
	ls, ok := listenerSets[key]
	// ListenerSet not found
	// It's not an error if the ListenerSet is not found because the BackendTrafficPolicy
	// may be reconciled by multiple controllers, and the ListenerSet may not be managed
	// by this controller.
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
			message := fmt.Sprintf("Unable to target ListenerSet %s, another BackendTrafficPolicy has already attached to it",
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
			message := fmt.Sprintf("Unable to target Listener %s/%s, another BackendTrafficPolicy has already attached to it",
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

func resolveBackendTrafficPolicyRouteTargetRef(
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

	// If sectionName is set, make sure it's valid
	if target.SectionName != nil {
		if err := validateRouteRuleSectionName(*target.SectionName, key, route); err != nil {
			return route.RouteContext, err
		}
	}

	if target.SectionName == nil {
		// Check if another policy targeting the same xRoute exists
		if route.attached {
			message := fmt.Sprintf("Unable to target %s %s, another BackendTrafficPolicy has already attached to it",
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
			message := fmt.Sprintf("Unable to target RouteRule %s/%s, another BackendTrafficPolicy has already attached to it",
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

func (t *Translator) translateBackendTrafficPolicyForRoute(
	policy *egv1a1.BackendTrafficPolicy,
	route RouteContext,
	target policyTargetReferenceWithSectionName,
	xdsIR resource.XdsIRMap,
	policyTargetListener *ListenerContext,
) error {
	tf, errs := t.buildTrafficFeatures(policy, nil)
	if tf == nil {
		// should not happen
		return nil
	}

	var targetListenerName string
	if policyTargetListener != nil {
		targetListenerName = irListenerName(policyTargetListener)
	}

	// Apply IR to all relevant routes
	for key, x := range xdsIR {
		// if policyTargetListener is not nil, only apply within its parent Gateway
		if policyTargetListener != nil && key != t.getIRKey(policyTargetListener.gateway.Gateway) {
			// Skip if not the gateway wanted
			continue
		}
		t.applyTrafficFeatureToRoute(route, tf, errs, policy, target, x, targetListenerName)
	}

	return errs
}

func (t *Translator) translateBackendTrafficPolicyForRouteWithMerge(
	policy, parentPolicy *egv1a1.BackendTrafficPolicy,
	target policyTargetReferenceWithSectionName,
	policyTargetListener *ListenerContext, route RouteContext,
	xdsIR resource.XdsIRMap,
) error {
	mergedPolicy, owners, err := t.mergeBackendTrafficPolicy(policy, parentPolicy)
	if err != nil {
		return fmt.Errorf("error merging policies: %w", err)
	}

	// Build traffic features from the merged policy
	tf, errs := t.buildTrafficFeatures(mergedPolicy, owners)
	if tf == nil {
		// should not happen
		return nil
	}

	// Since GlobalRateLimit merge relies on IR auto-generated key: (<policy-ns>/<policy-name>/rule/<rule-index>)
	// We can't simply merge the BTP's using utils.Merge() we need to specifically merge the GlobalRateLimit.Rules using IR fields.
	// Since ir.TrafficFeatures is not a built-in Kubernetes API object with defined merging strategies and it does not support a deep merge (for lists/maps).

	// Handle rate limit merging cases:
	// 1. Both policies have rate limits - merge them
	// 2. Only gateway policy has rate limits - preserve gateway policy's rule names
	// 3. Only route policy has rate limits - use route policy's rule names (default behavior)
	if policy.Spec.RateLimit != nil && parentPolicy.Spec.RateLimit != nil {
		tfGW, _ := t.buildTrafficFeatures(parentPolicy, nil)
		tfRoute, _ := t.buildTrafficFeatures(policy, nil)

		if tfGW != nil && tfRoute != nil &&
			tfGW.RateLimit != nil && tfRoute.RateLimit != nil {

			mergedRL, err := utils.Merge(tfGW.RateLimit, tfRoute.RateLimit, *policy.Spec.MergeType)
			if err != nil {
				return fmt.Errorf("error merging rate limits: %w", err)
			}
			// Replace the rate limit in the merged features if successful
			tf.RateLimit = mergedRL
		}
	} else if policy.Spec.RateLimit == nil && parentPolicy.Spec.RateLimit != nil {
		// Case 2: Only gateway policy has rate limits - preserve gateway policy's rule names
		tfGW, _ := t.buildTrafficFeatures(parentPolicy, nil)
		if tfGW != nil && tfGW.RateLimit != nil {
			// Use the gateway policy's rate limit with its original rule names
			tf.RateLimit = tfGW.RateLimit
		}
	}
	// Case 3: Only route policy has rate limits or neither has rate limits - use default behavior (tf already built from merged policy)

	x, ok := xdsIR[t.getIRKey(policyTargetListener.gateway.Gateway)]
	if !ok {
		// should not happen.
		return nil
	}
	t.applyTrafficFeatureToRoute(route, tf, errs, mergedPolicy, target, x, irListenerName(policyTargetListener))

	return errs
}

func (t *Translator) applyTrafficFeatureToRoute(route RouteContext,
	tf *ir.TrafficFeatures, errs error,
	policy *egv1a1.BackendTrafficPolicy,
	target policyTargetReferenceWithSectionName,
	x *ir.Xds,
	policyTargetListenerName string,
) {
	routeStatName := ""
	if tf.Telemetry != nil && tf.Telemetry.Metrics != nil {
		routeStatName = ptr.Deref(tf.Telemetry.Metrics.RouteStatName, "")
	}

	prefix := irRoutePrefix(route)
	for _, tcp := range x.TCP {
		// if listenerName is not nil, only apply to the specific listener
		if policyTargetListenerName != "" && policyTargetListenerName != tcp.Name {
			// Skip if not the listener wanted
			continue
		}
		for _, r := range tcp.Routes {
			// If specified the sectionName in policy target, must match route rule from ir route metadata.
			if target.SectionName != nil && string(*target.SectionName) != r.Destination.Metadata.SectionName {
				continue
			}
			if strings.HasPrefix(r.Destination.Name, prefix) {
				// only set attributes which weren't already set by a more
				// specific policy
				setIfNil(&r.LoadBalancer, tf.LoadBalancer)
				setIfNil(&r.ProxyProtocol, tf.ProxyProtocol)
				setIfNil(&r.HealthCheck, tf.HealthCheck)
				setIfNil(&r.CircuitBreaker, tf.CircuitBreaker)
				setIfNil(&r.TCPKeepalive, tf.TCPKeepalive)
				setIfNil(&r.Timeout, tf.Timeout)
				setIfNil(&r.BackendConnection, tf.BackendConnection)
				setIfNil(&r.DNS, tf.DNS)
				setIfNil(&r.StatName, buildRouteStatName(routeStatName, r.Metadata))
				appendTrafficPolicyMetadata(r.Metadata, policy)
			}
		}
	}

	for _, udp := range x.UDP {
		// if listenerName is not nil, only apply to the specific listener
		if policyTargetListenerName != "" && policyTargetListenerName != udp.Name {
			// Skip if not the listener wanted
			continue
		}
		if udp.Route != nil {
			r := udp.Route
			// If specified the sectionName in policy target, must match route rule from ir route metadata.
			if target.SectionName != nil && string(*target.SectionName) != r.Destination.Metadata.SectionName {
				continue
			}
			if strings.HasPrefix(r.Destination.Name, prefix) {
				// only set attributes which weren't already set by a more
				// specific policy
				setIfNil(&r.LoadBalancer, tf.LoadBalancer)
				setIfNil(&r.DNS, tf.DNS)
			}
		}
	}

	routesWithDirectResponse := sets.New[string]()
	for _, http := range x.HTTP {
		// if listenerName is not nil, only apply to the specific listener
		if policyTargetListenerName != "" && policyTargetListenerName != http.Name {
			// Skip if not the listener wanted
			continue
		}
		for _, r := range http.Routes {
			// If specified the sectionName in policy target, must match route rule from ir route metadata.
			if target.SectionName != nil && string(*target.SectionName) != r.Metadata.SectionName {
				continue
			}
			// Apply if there is a match
			if strings.HasPrefix(r.Name, prefix) {
				// If any of the features are already set, it means that a more specific
				// policy (targeting xRoute rule) has already set it, so we skip it.
				if r.Traffic != nil || r.UseClientProtocol != nil {
					continue
				}

				r.StatName = buildRouteStatName(routeStatName, r.Metadata)
				if errs != nil {
					// Return a 500 direct response
					r.DirectResponse = &ir.CustomResponse{
						StatusCode: new(uint32(500)),
					}
					routesWithDirectResponse.Insert(r.Name)
					continue
				}

				r.Traffic = tf.DeepCopy()

				if r.Traffic != nil && r.Traffic.LoadBalancer != nil &&
					r.Traffic.LoadBalancer.BackendUtilization != nil &&
					!ptr.Deref(r.Traffic.LoadBalancer.BackendUtilization.KeepResponseHeaders, false) {
					headersToRemove := []string{"endpoint-load-metrics", "endpoint-load-metrics-bin"}
					for _, h := range headersToRemove {
						found := false
						for _, existing := range r.RemoveResponseHeaders {
							if existing == h {
								found = true
								break
							}
						}
						if !found {
							r.RemoveResponseHeaders = append(r.RemoveResponseHeaders, h)
						}
					}
				}

				if localTo, err := buildClusterSettingsTimeout(&policy.Spec.ClusterSettings); err == nil {
					r.Traffic.Timeout = localTo
				}

				if policy.Spec.UseClientProtocol != nil {
					r.UseClientProtocol = policy.Spec.UseClientProtocol
				}
				appendTrafficPolicyMetadata(r.Metadata, policy)
			}
		}
	}
	if len(routesWithDirectResponse) > 0 {
		t.Logger.Info("setting 500 direct response in routes due to errors in BackendTrafficPolicy",
			"policy", utils.NamespacedName(policy),
			"routes", sets.List(routesWithDirectResponse),
			"error", errs,
		)
	}
}

// mergeBackendTrafficPolicy merges route policy into gateway policy, returning the merged
// policy and the per-field owners used to resolve references against the contributing
// policy's namespace.
func (t *Translator) mergeBackendTrafficPolicy(routePolicy, gwPolicy *egv1a1.BackendTrafficPolicy) (*egv1a1.BackendTrafficPolicy, *backendTrafficPolicyOwners, error) {
	if routePolicy.Spec.MergeType == nil || gwPolicy == nil {
		return routePolicy, nil, nil
	}

	mergedPolicy, err := utils.Merge(gwPolicy, routePolicy, *routePolicy.Spec.MergeType)
	if err != nil {
		return nil, nil, err
	}
	return mergedPolicy, buildBackendTrafficPolicyOwners(routePolicy, gwPolicy), nil
}

// buildTrafficFeatures builds IR traffic features from a BackendTrafficPolicy. owners is
// the per-field owners for a merged policy, or nil to resolve references against the
// policy's own namespace.
func (t *Translator) buildTrafficFeatures(policy *egv1a1.BackendTrafficPolicy, owners *backendTrafficPolicyOwners) (*ir.TrafficFeatures, error) {
	var (
		rl          *ir.RateLimit
		bl          *ir.BandwidthLimit
		lb          *ir.LoadBalancer
		pp          *ir.ProxyProtocol
		hc          *ir.HealthCheck
		cb          *ir.CircuitBreaker
		fi          *ir.FaultInjection
		ac          *ir.AdmissionControl
		to          *ir.Timeout
		ka          *ir.TCPKeepalive
		rt          *ir.Retry
		bc          *ir.BackendConnection
		ds          *ir.DNS
		h2          *ir.HTTP2Settings
		ro          *ir.ResponseOverride
		rb          *ir.RequestBuffer
		cp          []*ir.Compression
		httpUpgrade []ir.HTTPUpgradeConfig
		err, errs   error
	)

	if policy.Spec.RateLimit != nil {
		if rl, err = t.buildRateLimit(policy); err != nil {
			err = perr.WithMessage(err, "RateLimit")
			errs = errors.Join(errs, err)
		}
	}
	if policy.Spec.BandwidthLimit != nil {
		if bl, err = buildBandwidthLimit(policy.Spec.BandwidthLimit); err != nil {
			err = perr.WithMessage(err, "BandwidthLimit")
			errs = errors.Join(errs, err)
		}
	}
	if lb, err = buildLoadBalancer(&policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "LoadBalancer")
		errs = errors.Join(errs, err)
	}
	pp = buildProxyProtocol(&policy.Spec.ClusterSettings)
	hc = buildHealthCheck(&policy.Spec.ClusterSettings)
	if cb, err = buildCircuitBreaker(&policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "CircuitBreaker")
		errs = errors.Join(errs, err)
	}
	if policy.Spec.FaultInjection != nil {
		fi = t.buildFaultInjection(policy)
	}
	if policy.Spec.AdmissionControl != nil {
		ac = t.buildAdmissionControl(policy)
	}
	if ka, err = buildTCPKeepAlive(&policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "TCPKeepalive")
		errs = errors.Join(errs, err)
	}

	if rt, err = buildRetry(policy.Spec.Retry); err != nil {
		err = perr.WithMessage(err, "Retry")
		errs = errors.Join(errs, err)
	}

	if to, err = buildClusterSettingsTimeout(&policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "Timeout")
		errs = errors.Join(errs, err)
	}

	if bc, err = buildBackendConnection(&policy.Spec.ClusterSettings); err != nil {
		err = perr.WithMessage(err, "BackendConnection")
		errs = errors.Join(errs, err)
	}

	if h2, err = buildIRHTTP2Settings(policy.Spec.HTTP2); err != nil {
		err = perr.WithMessage(err, "HTTP2")
		errs = errors.Join(errs, err)
	}

	if ro, err = t.buildResponseOverride(policy, owners); err != nil {
		err = perr.WithMessage(err, "ResponseOverride")
		errs = errors.Join(errs, err)
	}

	if rb, err = buildRequestBuffer(policy.Spec.RequestBuffer); err != nil {
		err = perr.WithMessage(err, "RequestBuffer")
		errs = errors.Join(errs, err)
	}

	if err = validateTelemetry(policy.Spec.Telemetry); err != nil {
		err = perr.WithMessage(err, "Telemetry")
		errs = errors.Join(errs, err)
	}

	cp = buildCompression(policy.Spec.Compression, policy.Spec.Compressor)
	httpUpgrade = buildHTTPProtocolUpgradeConfig(policy.Spec.HTTPUpgrade)
	if rb != nil && len(httpUpgrade) > 0 {
		err = errors.New("requestBuffer cannot be used together with httpUpgrade")
		err = perr.WithMessage(err, "RequestBuffer")
		errs = errors.Join(errs, err)
	}

	ds = translateDNS(&policy.Spec.ClusterSettings, utils.NamespacedName(policy).String())

	return &ir.TrafficFeatures{
		RateLimit:         rl,
		BandwidthLimit:    bl,
		LoadBalancer:      lb,
		ProxyProtocol:     pp,
		HealthCheck:       hc,
		CircuitBreaker:    cb,
		FaultInjection:    fi,
		AdmissionControl:  ac,
		TCPKeepalive:      ka,
		Retry:             rt,
		BackendConnection: bc,
		HTTP2:             h2,
		DNS:               ds,
		Timeout:           to,
		ResponseOverride:  ro,
		RequestBuffer:     rb,
		Compression:       cp,
		HTTPUpgrade:       httpUpgrade,
		Telemetry:         buildBackendTelemetry(policy.Spec.Telemetry),
	}, errs
}

func buildBackendTelemetry(telemetry *egv1a1.BackendTelemetry) *ir.BackendTelemetry {
	if telemetry == nil {
		return nil
	}
	return &ir.BackendTelemetry{
		Tracing: buildBackendTracing(telemetry.Tracing),
		Metrics: buildBackendMetrics(telemetry.Metrics),
	}
}

func buildBackendTracing(tracing *egv1a1.Tracing) *ir.BackendTracing {
	if tracing == nil {
		return nil
	}
	return &ir.BackendTracing{
		SamplingFraction:        tracing.SamplingFraction,
		ClientSamplingFraction:  tracing.ClientSamplingFraction,
		OverallSamplingFraction: tracing.OverallSamplingFraction,
		CustomTags:              ir.CustomTagMapToSlice(tracing.CustomTags),
		Tags:                    ir.MapToSlice(tracing.Tags),
		SpanName:                tracing.SpanName,
	}
}

func buildBackendMetrics(metrics *egv1a1.BackendMetrics) *ir.BackendMetrics {
	if metrics == nil {
		return nil
	}
	return &ir.BackendMetrics{
		RouteStatName: metrics.RouteStatName,
	}
}

func (t *Translator) translateBackendTrafficPolicyForGateway(
	policy *egv1a1.BackendTrafficPolicy,
	gtwCtx *GatewayContext,
	target policyTargetReferenceWithSectionName,
	xdsIR resource.XdsIRMap,
) error {
	return t.translateBackendTrafficPolicyForListeners(
		policy,
		gtwCtx,
		gatewayBackendTrafficPolicyTargetListeners(gtwCtx, target),
		xdsIR,
	)
}

func (t *Translator) translateBackendTrafficPolicyForListenerSet(
	policy *egv1a1.BackendTrafficPolicy,
	gtwCtx *GatewayContext,
	listenerSet *gwapiv1.ListenerSet,
	target policyTargetReferenceWithSectionName,
	xdsIR resource.XdsIRMap,
) error {
	return t.translateBackendTrafficPolicyForListeners(
		policy,
		gtwCtx,
		listenerSetBackendTrafficPolicyTargetListeners(gtwCtx, listenerSet, target),
		xdsIR,
	)
}

func gatewayBackendTrafficPolicyTargetListeners(
	gtwCtx *GatewayContext,
	target policyTargetReferenceWithSectionName,
) []*ListenerContext {
	listeners := make([]*ListenerContext, 0, len(gtwCtx.listeners))
	for _, listener := range gtwCtx.listeners {
		if target.SectionName != nil {
			if listener.isFromListenerSet() || listener.Name != *target.SectionName {
				continue
			}
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

func listenerSetBackendTrafficPolicyTargetListeners(
	gtwCtx *GatewayContext,
	listenerSet *gwapiv1.ListenerSet,
	target policyTargetReferenceWithSectionName,
) []*ListenerContext {
	listeners := make([]*ListenerContext, 0, len(gtwCtx.listeners))
	for _, listener := range gtwCtx.listeners {
		if !listener.isFromListenerSet() {
			continue
		}
		if listener.listenerSet.Namespace != listenerSet.Namespace || listener.listenerSet.Name != listenerSet.Name {
			continue
		}
		if target.SectionName != nil && listener.Name != *target.SectionName {
			continue
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

func (t *Translator) translateBackendTrafficPolicyForListeners(
	policy *egv1a1.BackendTrafficPolicy,
	gtwCtx *GatewayContext,
	targetListeners []*ListenerContext,
	xdsIR resource.XdsIRMap,
) error {
	tf, errs := t.buildTrafficFeatures(policy, nil)
	if tf == nil {
		// should not happen
		return errs
	}

	routeStatName := ""
	if tf.Telemetry != nil && tf.Telemetry.Metrics != nil {
		routeStatName = ptr.Deref(tf.Telemetry.Metrics.RouteStatName, "")
	}

	irKey := t.getIRKey(gtwCtx.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]

	listenerNames := sets.New[string]()
	for _, listener := range targetListeners {
		listenerNames.Insert(irListenerName(listener))
	}

	for _, tcp := range x.TCP {
		if !listenerNames.Has(tcp.Name) {
			continue
		}

		for _, r := range tcp.Routes {
			// only set attributes which weren't already set by a more
			// specific policy
			setIfNil(&r.LoadBalancer, tf.LoadBalancer)
			setIfNil(&r.ProxyProtocol, tf.ProxyProtocol)
			setIfNil(&r.HealthCheck, tf.HealthCheck)
			setIfNil(&r.CircuitBreaker, tf.CircuitBreaker)
			setIfNil(&r.TCPKeepalive, tf.TCPKeepalive)
			setIfNil(&r.Timeout, tf.Timeout)
			setIfNil(&r.DNS, tf.DNS)
			setIfNil(&r.StatName, buildRouteStatName(routeStatName, r.Metadata))
			appendTrafficPolicyMetadata(r.Metadata, policy)
		}
	}

	for _, udp := range x.UDP {
		if !listenerNames.Has(udp.Name) {
			continue
		}

		if udp.Route == nil {
			continue
		}

		route := udp.Route

		// only set attributes which weren't already set by a more
		// specific policy
		setIfNil(&route.LoadBalancer, tf.LoadBalancer)
		setIfNil(&route.DNS, tf.DNS)
	}

	routesWithDirectResponse := sets.New[string]()
	for _, http := range x.HTTP {
		if !listenerNames.Has(http.Name) {
			continue
		}

		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway/ListenerSet).
		for _, r := range http.Routes {
			// If any of the features are already set, it means that a more specific
			// policy (targeting xRoute rule, xRoute, listener) has already set it, so we skip it.
			if r.Traffic != nil || r.UseClientProtocol != nil {
				continue
			}

			setIfNil(&r.StatName, buildRouteStatName(routeStatName, r.Metadata))
			if errs != nil {
				// Return a 500 direct response
				r.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(r.Name)
				continue
			}

			r.Traffic = tf.DeepCopy()
			if localTo, err := buildClusterSettingsTimeout(&policy.Spec.ClusterSettings); err == nil {
				r.Traffic.Timeout = localTo
			}

			if policy.Spec.UseClientProtocol != nil {
				r.UseClientProtocol = policy.Spec.UseClientProtocol
			}

			appendTrafficPolicyMetadata(r.Metadata, policy)
		}
	}
	if len(routesWithDirectResponse) > 0 {
		t.Logger.Info("setting 500 direct response in routes due to errors in BackendTrafficPolicy",
			"policy", utils.NamespacedName(policy),
			"routes", sets.List(routesWithDirectResponse),
			"error", errs,
		)
	}

	return errs
}

func appendTrafficPolicyMetadata(md *ir.ResourceMetadata, policy *egv1a1.BackendTrafficPolicy) {
	if md == nil || policy == nil {
		return
	}

	md.Policies = append(md.Policies, &ir.PolicyMetadata{
		Kind:      egv1a1.KindBackendTrafficPolicy,
		Name:      policy.Name,
		Namespace: policy.Namespace,
	})
}

func (t *Translator) buildRateLimit(policy *egv1a1.BackendTrafficPolicy) (*ir.RateLimit, error) {
	// For backward compatibility, process the deprecated Type field if specified.
	if policy.Spec.RateLimit.Type != nil {
		switch *policy.Spec.RateLimit.Type {
		case egv1a1.GlobalRateLimitType:
			return t.buildGlobalRateLimit(policy)
		case egv1a1.LocalRateLimitType:
			return t.buildLocalRateLimit(policy)
		}
		return nil, fmt.Errorf("invalid rateLimit type: %s", *policy.Spec.RateLimit.Type)
	}

	return t.buildBothRateLimit(policy)
}

func (t *Translator) buildLocalRateLimit(policy *egv1a1.BackendTrafficPolicy) (*ir.RateLimit, error) {
	if policy.Spec.RateLimit.Local == nil {
		return nil, fmt.Errorf("local configuration empty for rateLimit")
	}

	local := policy.Spec.RateLimit.Local

	// Envoy local rateLimit requires a default limit to be set for a route.
	// EG uses the first rule without clientSelectors as the default route-level
	// limit. If no such rule is found, EG uses a default limit of uint32 max.
	var defaultLimit *ir.RateLimitValue
	var defaultXRateLimitOption *egv1a1.XRateLimitHeadersOption
	for _, rule := range local.Rules {
		if len(rule.ClientSelectors) == 0 {
			if defaultLimit != nil {
				return nil, fmt.Errorf("local rateLimit can not have more than one rule without clientSelectors")
			}
			defaultLimit = &ir.RateLimitValue{
				Requests: rule.Limit.Requests,
				Unit:     ir.RateLimitUnit(rule.Limit.Unit),
			}
			// Capture the xRateLimit setting for the default bucket
			defaultXRateLimitOption = rule.XRateLimitHeaders
		}
	}
	// If no rule without clientSelectors is found, use uint32 max as the default
	// limit, which effectively make the default limit unlimited.
	if defaultLimit == nil {
		defaultLimit = &ir.RateLimitValue{
			Requests: math.MaxUint32,
			Unit:     ir.RateLimitUnit(egv1a1.RateLimitUnitSecond),
		}
	}

	// Validate that the rule limit unit is a multiple of the default limit unit.
	// This is required by Envoy local rateLimit implementation.
	// see https://github.com/envoyproxy/envoy/blob/6d9a6e995f472526de2b75233abca69aa00021ed/source/extensions/filters/common/local_ratelimit/local_ratelimit_impl.cc#L49
	defaultLimitUnit := ratelimit.UnitToSeconds(egv1a1.RateLimitUnit(defaultLimit.Unit))
	for _, rule := range local.Rules {
		ruleLimitUint := ratelimit.UnitToSeconds(rule.Limit.Unit)
		if defaultLimitUnit == 0 || ruleLimitUint%defaultLimitUnit != 0 {
			return nil, fmt.Errorf("local rateLimit rule limit unit must be a multiple of the default limit unit")
		}
	}

	var err error
	var irRule *ir.RateLimitRule
	irRules := make([]*ir.RateLimitRule, 0)
	for i, rule := range local.Rules {
		// We don't process the rule without clientSelectors here because it's
		// previously used as the default route-level limit.
		if len(rule.ClientSelectors) == 0 {
			continue
		}

		irRule, err = buildRateLimitRule(&rule)
		if err != nil {
			return nil, err
		}
		// Set the Name field as <policy-ns>/<policy-name>/rule/<rule-index>
		irRule.Name = irRuleName(policy.Namespace, policy.Name, i)
		irRules = append(irRules, irRule)
	}

	rateLimit := &ir.RateLimit{
		Local: &ir.LocalRateLimit{
			Default:                 *defaultLimit,
			Rules:                   irRules,
			DefaultXRateLimitOption: defaultXRateLimitOption,
		},
	}

	return rateLimit, nil
}

func (t *Translator) buildGlobalRateLimit(policy *egv1a1.BackendTrafficPolicy) (*ir.RateLimit, error) {
	if policy.Spec.RateLimit.Global == nil {
		return nil, fmt.Errorf("global configuration empty for rateLimit")
	}
	if !t.GlobalRateLimitEnabled {
		return nil, fmt.Errorf("enable Ratelimit in the EnvoyGateway config to configure global rateLimit")
	}

	global := policy.Spec.RateLimit.Global
	rateLimit := &ir.RateLimit{
		Global: &ir.GlobalRateLimit{
			Rules: make([]*ir.RateLimitRule, len(global.Rules)),
		},
	}

	irRules := rateLimit.Global.Rules
	var err error
	for i, rule := range global.Rules {
		irRules[i], err = buildRateLimitRule(&rule)
		if err != nil {
			return nil, err
		}
		// Set the Name field as <policy-ns>/<policy-name>/rule/<rule-index>
		irRules[i].Name = irRuleName(policy.Namespace, policy.Name, i)
	}

	return rateLimit, nil
}

func (t *Translator) buildBothRateLimit(policy *egv1a1.BackendTrafficPolicy) (*ir.RateLimit, error) {
	var (
		localRateLimit  *ir.RateLimit
		globalRateLimit *ir.RateLimit
		err             error
	)

	if policy.Spec.RateLimit.Local != nil {
		localRateLimit, err = t.buildLocalRateLimit(policy)
		if err != nil {
			return nil, err
		}
	}
	if policy.Spec.RateLimit.Global != nil {
		globalRateLimit, err = t.buildGlobalRateLimit(policy)
		if err != nil {
			return nil, err
		}
	}
	rl := &ir.RateLimit{}
	if localRateLimit != nil && localRateLimit.Local != nil {
		rl.Local = localRateLimit.Local
	}
	if globalRateLimit != nil && globalRateLimit.Global != nil {
		rl.Global = globalRateLimit.Global
	}
	return rl, nil
}

func buildRateLimitRule(rule *egv1a1.RateLimitRule) (*ir.RateLimitRule, error) {
	irRule := &ir.RateLimitRule{
		Limit: ir.RateLimitValue{
			Requests: rule.Limit.Requests,
			Unit:     ir.RateLimitUnit(rule.Limit.Unit),
		},
		HeaderMatches:    make([]*ir.StringMatch, 0),
		MethodMatches:    make([]*ir.StringMatch, 0),
		Shared:           rule.Shared,
		ShadowMode:       rule.ShadowMode,
		XRateLimitOption: rule.XRateLimitHeaders,
	}

	if md := rule.Limit.FromMetadata; md != nil {
		irRule.Limit.FromMetadata = &ir.RateLimitValueMetadata{
			Namespace: md.Namespace,
			Key:       md.Key,
		}
	}

	for _, match := range rule.ClientSelectors {
		if len(match.Headers) == 0 && len(match.Methods) == 0 &&
			match.Path == nil && match.SourceCIDR == nil && len(match.QueryParams) == 0 {
			return nil, fmt.Errorf(
				"unable to translate rateLimit. At least one of the" +
					" header or method or path or sourceCIDR or queryParameters must be specified")
		}
		for _, header := range match.Headers {
			switch {
			case header.Type == nil && header.Value != nil:
				fallthrough
			case *header.Type == egv1a1.HeaderMatchExact && header.Value != nil:
				m := &ir.StringMatch{
					Name:   header.Name,
					Exact:  header.Value,
					Invert: header.Invert,
				}
				irRule.HeaderMatches = append(irRule.HeaderMatches, m)
			case *header.Type == egv1a1.HeaderMatchRegularExpression && header.Value != nil:
				if err := regex.Validate(*header.Value); err != nil {
					return nil, err
				}
				m := &ir.StringMatch{
					Name:      header.Name,
					SafeRegex: header.Value,
					Invert:    header.Invert,
				}
				irRule.HeaderMatches = append(irRule.HeaderMatches, m)
			case *header.Type == egv1a1.HeaderMatchDistinct && header.Value == nil:
				if header.Invert != nil && *header.Invert {
					return nil, fmt.Errorf("unable to translate rateLimit." +
						"Invert is not applicable for distinct header match type")
				}
				m := &ir.StringMatch{
					Name:     header.Name,
					Distinct: true,
				}
				irRule.HeaderMatches = append(irRule.HeaderMatches, m)
			default:
				return nil, fmt.Errorf(
					"unable to translate rateLimit. Either the header." +
						"Type is not valid or the header is missing a value")
			}
		}

		for _, method := range match.Methods {
			irRule.MethodMatches = append(irRule.MethodMatches, &ir.StringMatch{
				Exact:  new(string(method.Value)),
				Invert: method.Invert,
			})
		}

		if match.Path != nil {
			switch ptr.Deref(match.Path.Type, gwapiv1.PathMatchPathPrefix) {
			case gwapiv1.PathMatchPathPrefix:
				if match.Path.Value == "/" {
					irRule.PathMatch = &ir.StringMatch{
						Prefix: new(match.Path.Value),
						Invert: match.Path.Invert,
					}
				} else {
					// envoy ratelimit HeaderMatcher doesn't support PathSeparatedPrefix like route matching,
					// so we use regex to achieve the same path-separated prefix behavior.
					irRule.PathMatch = &ir.StringMatch{
						SafeRegex: new(regex.PathSeparatedPrefixRegex(match.Path.Value)),
						Invert:    match.Path.Invert,
					}
				}
			case gwapiv1.PathMatchExact:
				irRule.PathMatch = &ir.StringMatch{
					Exact:  new(match.Path.Value),
					Invert: match.Path.Invert,
				}
			case gwapiv1.PathMatchRegularExpression:
				if err := regex.Validate(match.Path.Value); err != nil {
					return nil, err
				}
				irRule.PathMatch = &ir.StringMatch{
					SafeRegex: new(match.Path.Value),
					Invert:    match.Path.Invert,
				}
			default:
				return nil, fmt.Errorf("unable to translate rateLimit: invalid path type.")
			}
		}

		if match.SourceCIDR != nil {
			distinct := false
			if match.SourceCIDR.Type != nil && *match.SourceCIDR.Type == egv1a1.SourceMatchDistinct {
				distinct = true
			}
			invert := false
			if match.SourceCIDR.Invert != nil {
				invert = *match.SourceCIDR.Invert
			}

			cidrMatch, err := parseCIDR(match.SourceCIDR.Value)
			if err != nil {
				return nil, fmt.Errorf("unable to translate rateLimit: %w", err)
			}
			cidrMatch.Distinct = distinct
			cidrMatch.Invert = invert
			irRule.CIDRMatch = cidrMatch
		}

		for _, queryParam := range match.QueryParams {
			// Validate QueryParamMatch
			if queryParam.Name == "" {
				return nil, fmt.Errorf("name is required when QueryParamMatch is specified")
			}

			var stringMatch ir.StringMatch

			// Default to Exact match if Type is not specified
			matchType := egv1a1.QueryParamMatchExact
			if queryParam.Type != nil {
				matchType = *queryParam.Type
			}

			switch matchType {
			case egv1a1.QueryParamMatchExact:
				if queryParam.Value == nil || *queryParam.Value == "" {
					return nil, fmt.Errorf("value is required for Exact query parameter match")
				}
				stringMatch = ir.StringMatch{
					Name:   queryParam.Name,
					Exact:  queryParam.Value,
					Invert: queryParam.Invert,
				}
			case egv1a1.QueryParamMatchRegularExpression:
				if queryParam.Value == nil || *queryParam.Value == "" {
					return nil, fmt.Errorf("value is required for RegularExpression query parameter match")
				}
				if err := regex.Validate(*queryParam.Value); err != nil {
					return nil, err
				}
				stringMatch = ir.StringMatch{
					Name:      queryParam.Name,
					SafeRegex: queryParam.Value,
					Invert:    queryParam.Invert,
				}
			case egv1a1.QueryParamMatchDistinct:
				if queryParam.Invert != nil && *queryParam.Invert {
					return nil, fmt.Errorf("unable to translate rateLimit." +
						"Invert is not applicable for distinct query parameter match type")
				}
				if queryParam.Value != nil {
					return nil, fmt.Errorf("unable to translate rateLimit." +
						"Value is not applicable for distinct query parameter match type")
				}
				stringMatch = ir.StringMatch{
					Name:     queryParam.Name,
					Distinct: true,
				}
			default:
				return nil, fmt.Errorf("invalid query parameter match type: %s", matchType)
			}

			m := &ir.QueryParamMatch{
				StringMatch: stringMatch,
			}
			irRule.QueryParamMatches = append(irRule.QueryParamMatches, m)
		}
	}

	if cost := rule.Cost; cost != nil {
		if cost.Request != nil {
			irRule.RequestCost = translateRateLimitCost(cost.Request)
		}
		if cost.Response != nil {
			irRule.ResponseCost = translateRateLimitCost(cost.Response)
		}
	}
	return irRule, nil
}

func translateRateLimitCost(cost *egv1a1.RateLimitCostSpecifier) *ir.RateLimitCost {
	ret := &ir.RateLimitCost{}
	if cost.Number != nil {
		ret.Number = cost.Number
	}
	if cost.Metadata != nil {
		ret.Format = new(fmt.Sprintf("%%DYNAMIC_METADATA(%s:%s)%%",
			cost.Metadata.Namespace, cost.Metadata.Key))
	}
	return ret
}

func int64ToUint32(in int64) (uint32, bool) {
	if in >= 0 && in <= math.MaxUint32 {
		return uint32(in), true
	}
	return 0, false
}

func buildBandwidthLimit(bandwidth *egv1a1.BandwidthLimitSpec) (*ir.BandwidthLimit, error) {
	if bandwidth == nil {
		return nil, nil
	}

	bl := &ir.BandwidthLimit{}

	if bandwidth.Request != nil {
		bytes, ok := bandwidth.Request.Limit.Value.AsInt64()
		if !ok {
			return nil, fmt.Errorf("request limit value must be convertible to an int64")
		}
		if bytes < 0 {
			return nil, fmt.Errorf("request limit value must be positive")
		}
		kibps, err := bandwidthToKibps(uint64(bytes), bandwidth.Request.Limit.Unit)
		if err != nil {
			return nil, fmt.Errorf("request: %w", err)
		}
		bl.Request = &ir.BandwidthLimitConfig{
			LimitKibps: kibps,
		}
	}
	if bandwidth.Response != nil {
		bytes, ok := bandwidth.Response.Limit.Value.AsInt64()
		if !ok {
			return nil, fmt.Errorf("response limit value must be convertible to an int64")
		}
		if bytes < 0 {
			return nil, fmt.Errorf("response limit value must be positive")
		}
		kibps, err := bandwidthToKibps(uint64(bytes), bandwidth.Response.Limit.Unit)
		if err != nil {
			return nil, fmt.Errorf("response: %w", err)
		}
		bl.Response = &ir.BandwidthLimitConfig{
			LimitKibps: kibps,
		}

		if bandwidth.Response.ResponseTrailers != nil {
			bl.Response.ResponseTrailers = &ir.BandwidthLimitResponseTrailers{
				Prefix: bandwidth.Response.ResponseTrailers.Prefix,
			}
		}
	}
	return bl, nil
}

// bandwidthToKibps converts bytes-per-unit to kibibytes-per-second (KiB/s).
// Returns an error if the result is below Envoy's minimum of 1 KiB/s.
func bandwidthToKibps(limit uint64, unit egv1a1.BandwidthLimitUnit) (uint64, error) {
	var secondsPerUnit uint64
	switch unit {
	case egv1a1.BandwidthLimitUnitMinute:
		secondsPerUnit = 60
	case egv1a1.BandwidthLimitUnitHour:
		secondsPerUnit = 3600
	default: // Second
		secondsPerUnit = 1
	}
	kibps := limit / (secondsPerUnit * 1024)
	if kibps == 0 {
		return 0, fmt.Errorf("bandwidth limit of %d bytes per %s is below the minimum of 1 KiB/s", limit, unit)
	}
	return kibps, nil
}

func (t *Translator) buildFaultInjection(policy *egv1a1.BackendTrafficPolicy) *ir.FaultInjection {
	var (
		fi  *ir.FaultInjection
		d   time.Duration
		err error
	)
	if policy.Spec.FaultInjection != nil {
		fi = &ir.FaultInjection{}

		if policy.Spec.FaultInjection.Delay != nil {
			if policy.Spec.FaultInjection.Delay.FixedDelay != nil {
				d, err = time.ParseDuration(string(*policy.Spec.FaultInjection.Delay.FixedDelay))
				if err != nil {
					return nil
				}
			}
			fi.Delay = &ir.FaultInjectionDelay{
				Percentage: policy.Spec.FaultInjection.Delay.Percentage,
				FixedDelay: ir.MetaV1DurationPtr(d),
			}
		}
		if policy.Spec.FaultInjection.Abort != nil {
			fi.Abort = &ir.FaultInjectionAbort{
				Percentage: policy.Spec.FaultInjection.Abort.Percentage,
			}

			if policy.Spec.FaultInjection.Abort.GrpcStatus != nil {
				fi.Abort.GrpcStatus = new(uint32(*policy.Spec.FaultInjection.Abort.GrpcStatus))
			}
			if policy.Spec.FaultInjection.Abort.HTTPStatus != nil {
				fi.Abort.HTTPStatus = policy.Spec.FaultInjection.Abort.HTTPStatus
			}
		}
	}
	return fi
}

func (t *Translator) buildAdmissionControl(policy *egv1a1.BackendTrafficPolicy) *ir.AdmissionControl {
	if policy.Spec.AdmissionControl == nil {
		return nil
	}

	ac := &ir.AdmissionControl{
		MinSuccessRate:      policy.Spec.AdmissionControl.MinSuccessRate,
		RejectionAggression: policy.Spec.AdmissionControl.RejectionAggression,
		MinRequestRate:      policy.Spec.AdmissionControl.MinRequestRate,
		MaxRejectionPercent: policy.Spec.AdmissionControl.MaxRejectionPercent,
	}

	if policy.Spec.AdmissionControl.SamplingWindow != nil {
		if d, err := time.ParseDuration(string(*policy.Spec.AdmissionControl.SamplingWindow)); err == nil {
			ac.SamplingWindow = &metav1.Duration{Duration: d}
		}
	}

	if policy.Spec.AdmissionControl.SuccessCriteria != nil {
		ac.SuccessCriteria = &ir.AdmissionControlSuccessCriteria{}

		if policy.Spec.AdmissionControl.SuccessCriteria.HTTP != nil {
			httpStatuses := make([]int32, len(policy.Spec.AdmissionControl.SuccessCriteria.HTTP.StatusCodes))
			for i, s := range policy.Spec.AdmissionControl.SuccessCriteria.HTTP.StatusCodes {
				httpStatuses[i] = int32(s)
			}
			ac.SuccessCriteria.HTTP = &ir.HTTPSuccessCriteria{
				StatusCodes: httpStatuses,
			}
		}

		if policy.Spec.AdmissionControl.SuccessCriteria.GRPC != nil {
			grpcStatuses := make([]string, len(policy.Spec.AdmissionControl.SuccessCriteria.GRPC.StatusCodes))
			for i, s := range policy.Spec.AdmissionControl.SuccessCriteria.GRPC.StatusCodes {
				grpcStatuses[i] = string(s)
			}
			ac.SuccessCriteria.GRPC = &ir.GRPCSuccessCriteria{
				StatusCodes: grpcStatuses,
			}
		}
	}

	return ac
}

func makeIrStatusSet(in []egv1a1.HTTPStatus) []ir.HTTPStatus {
	statusSet := sets.NewInt()
	for _, r := range in {
		statusSet.Insert(int(r))
	}
	irStatuses := make([]ir.HTTPStatus, 0, statusSet.Len())

	for _, r := range statusSet.List() {
		irStatuses = append(irStatuses, ir.HTTPStatus(r))
	}
	return irStatuses
}

func makeIrTriggerSet(in []egv1a1.TriggerEnum) []ir.TriggerEnum {
	triggerSet := sets.NewString()
	for _, r := range in {
		triggerSet.Insert(string(r))
	}
	irTriggers := make([]ir.TriggerEnum, 0, triggerSet.Len())

	for _, r := range triggerSet.List() {
		irTriggers = append(irTriggers, ir.TriggerEnum(r))
	}
	return irTriggers
}

func buildRequestBuffer(spec *egv1a1.RequestBuffer) (*ir.RequestBuffer, error) {
	if spec == nil {
		return nil, nil
	}

	maxBytes, ok := spec.Limit.AsInt64()
	if !ok {
		return nil, fmt.Errorf("limit must be convertible to an int64")
	}

	if maxBytes < 0 || maxBytes > math.MaxUint32 {
		return nil, fmt.Errorf("limit value %s is out of range, must be between 0 and %d",
			spec.Limit.String(), math.MaxUint32)
	}

	return &ir.RequestBuffer{
		Limit: spec.Limit,
	}, nil
}

func (t *Translator) buildResponseOverride(policy *egv1a1.BackendTrafficPolicy, owners *backendTrafficPolicyOwners) (*ir.ResponseOverride, error) {
	if len(policy.Spec.ResponseOverride) == 0 {
		return nil, nil
	}

	// Resolve body ValueRefs against the owner's namespace, falling back to the policy's own.
	responseOverrideNs := policy.Namespace
	if owners != nil && owners.responseOverride != nil {
		responseOverrideNs = owners.responseOverride.Namespace
	}

	rules := make([]ir.ResponseOverrideRule, 0, len(policy.Spec.ResponseOverride))
	for index, ro := range policy.Spec.ResponseOverride {
		match := ir.CustomResponseMatch{
			StatusCodes: make([]ir.StatusCodeMatch, 0, len(ro.Match.StatusCodes)),
		}

		for _, code := range ro.Match.StatusCodes {
			if code.Type != nil && *code.Type == egv1a1.StatusCodeValueTypeRange {
				match.StatusCodes = append(match.StatusCodes, ir.StatusCodeMatch{
					Range: &ir.StatusCodeRange{
						Start: code.Range.Start,
						End:   code.Range.End,
					},
				})
			} else {
				match.StatusCodes = append(match.StatusCodes, ir.StatusCodeMatch{
					Value: code.Value,
				})
			}
		}

		if ro.Redirect != nil {
			redirect := &ir.Redirect{
				Scheme: ro.Redirect.Scheme,
			}
			if ro.Redirect.Path != nil {
				redirect.Path = &ir.HTTPPathModifier{
					FullReplace:        ro.Redirect.Path.ReplaceFullPath,
					PrefixMatchReplace: ro.Redirect.Path.ReplacePrefixMatch,
				}
			}
			if ro.Redirect.Hostname != nil {
				redirect.Hostname = new(string(*ro.Redirect.Hostname))
			}
			if ro.Redirect.Port != nil {
				redirect.Port = new(uint32(*ro.Redirect.Port))
			}
			if ro.Redirect.StatusCode != nil {
				redirect.StatusCode = new(int32(*ro.Redirect.StatusCode))
			}

			rules = append(rules, ir.ResponseOverrideRule{
				Name:     defaultResponseOverrideRuleName(policy, index),
				Match:    match,
				Redirect: redirect,
				Source:   sourceFromAPI(ro.Source),
			})
		} else {
			response := &ir.CustomResponse{
				ContentType: ro.Response.ContentType,
			}

			if ro.Response.StatusCode != nil {
				response.StatusCode = new(uint32(*ro.Response.StatusCode))
			}

			var err error
			response.Body, err = t.getCustomResponseBody(ro.Response.Body, responseOverrideNs)
			if err != nil {
				return nil, err
			}

			rhm := ro.Response.Header
			if rhm != nil {
				for h := range rhm.Add {
					response.AddResponseHeaders = append(response.AddResponseHeaders, ir.AddHeader{
						Name:   string(rhm.Add[h].Name),
						Append: true,
						Value:  []string{rhm.Add[h].Value},
					})
				}
				for h := range rhm.Set {
					response.AddResponseHeaders = append(response.AddResponseHeaders, ir.AddHeader{
						Name:   string(rhm.Set[h].Name),
						Append: false,
						Value:  []string{rhm.Set[h].Value},
					})
				}
			}

			rules = append(rules, ir.ResponseOverrideRule{
				Name:     defaultResponseOverrideRuleName(policy, index),
				Match:    match,
				Response: response,
				Source:   sourceFromAPI(ro.Source),
			})
		}
	}
	return &ir.ResponseOverride{
		Name:  irConfigName(policy),
		Rules: rules,
	}, nil
}

func (t *Translator) getCustomResponseBody(
	body *egv1a1.CustomResponseBody,
	policyNs string,
) ([]byte, error) {
	if body == nil {
		return nil, nil
	}

	if body.Type != nil && *body.Type == egv1a1.ResponseValueTypeValueRef {
		cm := t.GetConfigMap(policyNs, string(body.ValueRef.Name))
		if cm != nil {
			b, dataOk := cm.Data[ResponseBodyConfigMapKey]
			switch {
			case dataOk:
				data := []byte(b)
				return data, nil
			case len(cm.Data) > 0: // Fallback to the first key if response.body is not found
				for _, value := range cm.Data {
					data := []byte(value)
					return data, nil
				}
			case len(cm.BinaryData) > 0:
				for _, binData := range cm.BinaryData {
					return binData, nil
				}
			default:
				return nil, fmt.Errorf("can't find the key %s in the referenced configmap %s/%s", ResponseBodyConfigMapKey, policyNs, body.ValueRef.Name)
			}
		} else {
			return nil, fmt.Errorf("can't find the referenced configmap %s/%s", policyNs, body.ValueRef.Name)
		}
	} else if body.Inline != nil {
		inlineValue := []byte(*body.Inline)
		return inlineValue, nil
	}

	return nil, nil
}

// backendTrafficPolicyOwners records which policy (route or parent) contributed each
// merged field that references other objects, so references resolve against the owner's
// namespace. Mirrors the field-owner pattern used for SecurityPolicy.
type backendTrafficPolicyOwners struct {
	responseOverride *egv1a1.BackendTrafficPolicy
}

// buildBackendTrafficPolicyOwners picks the owner of each merged field: the route policy
// when it sets the field, otherwise the parent.
func buildBackendTrafficPolicyOwners(route, parent *egv1a1.BackendTrafficPolicy) *backendTrafficPolicyOwners {
	responseOverrideOwner := parent
	if len(route.Spec.ResponseOverride) > 0 {
		responseOverrideOwner = route
	}
	return &backendTrafficPolicyOwners{
		responseOverride: responseOverrideOwner,
	}
}

func sourceFromAPI(s *egv1a1.ResponseOverrideSource) egv1a1.ResponseOverrideSource {
	if s == nil {
		return ""
	}
	return *s
}

func defaultResponseOverrideRuleName(policy *egv1a1.BackendTrafficPolicy, index int) string {
	return fmt.Sprintf(
		"%s/responseoverride/rule/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

func buildCompression(compression, compressor []*egv1a1.Compression) []*ir.Compression {
	// Handle the Compressor field first (higher priority)
	if len(compressor) > 0 {
		result := make([]*ir.Compression, 0, len(compressor))
		for i, c := range compressor {
			// Only add compression if the corresponding compressor not null
			if (c.Type == egv1a1.GzipCompressorType && c.Gzip != nil) ||
				(c.Type == egv1a1.BrotliCompressorType && c.Brotli != nil) ||
				(c.Type == egv1a1.ZstdCompressorType && c.Zstd != nil) {
				irCompression := ir.Compression{
					Type:        c.Type,
					ChooseFirst: i == 0, // only the first compressor is marked as ChooseFirst
				}
				if c.MinContentLength != nil {
					minContentLength, ok := c.MinContentLength.AsInt64()
					if ok {
						irCompression.MinContentLength = new(uint32(minContentLength))
					}
				}
				result = append(result, &irCompression)
			}
		}
		return result
	}

	// Fallback to the deprecated Compression field
	if compression == nil {
		return nil
	}
	result := make([]*ir.Compression, 0, len(compression))
	for i, c := range compression {
		irCompression := ir.Compression{
			Type:        c.Type,
			ChooseFirst: i == 0, // only the first compressor is marked as ChooseFirst
		}
		if c.MinContentLength != nil {
			minContentLength, ok := c.MinContentLength.AsInt64()
			if ok {
				irCompression.MinContentLength = new(uint32(minContentLength))
			}
		}
		result = append(result, &irCompression)
	}

	return result
}

func buildHTTPProtocolUpgradeConfig(cfgs []*egv1a1.ProtocolUpgradeConfig) []ir.HTTPUpgradeConfig {
	if len(cfgs) == 0 {
		return nil
	}

	result := make([]ir.HTTPUpgradeConfig, 0, len(cfgs))
	for _, cfg := range cfgs {
		upgrade := ir.HTTPUpgradeConfig{
			Type: cfg.Type,
		}
		if cfg.Connect != nil {
			upgrade.Connect = &ir.ConnectConfig{
				Terminate: ptr.Deref(cfg.Connect.Terminate, false),
			}
		}
		result = append(result, upgrade)
	}

	return result
}

func validateTelemetry(telemetry *egv1a1.BackendTelemetry) error {
	if telemetry == nil {
		return nil
	}

	if telemetry.Metrics != nil && ptr.Deref(telemetry.Metrics.RouteStatName, "") != "" {
		return egv1a1validation.ValidateRouteStatName(*telemetry.Metrics.RouteStatName)
	}

	return nil
}

func buildRouteStatName(routeStatName string, metadata *ir.ResourceMetadata) *string {
	if routeStatName == "" || metadata == nil {
		return nil
	}

	statName := strings.ReplaceAll(routeStatName, egv1a1.StatFormatterRouteName, metadata.Name)
	statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteNamespace, metadata.Namespace)
	statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteKind, metadata.Kind)

	if metadata.SectionName == "" {
		statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteRuleName, "-")
	} else {
		statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteRuleName, metadata.SectionName)
	}

	return &statName
}
