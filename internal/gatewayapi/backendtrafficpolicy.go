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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

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
)

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

	// Map of Gateway to the routes attached to it.
	gatewayRouteMap := &GatewayPolicyRouteMap{
		Routes:       make(map[NamespacedNameWithSection]sets.Set[string], gatewayMapSize),
		SectionIndex: make(map[types.NamespacedName]sets.Set[string], gatewayMapSize),
	}

	// Map of attached Policy to Gateway. It is used to merge policies process.
	gatewayPolicyMap := make(map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy, gatewayMapSize)

	// Map of Gateway to the routes merged to it.
	gatewayPolicyMerged := &GatewayPolicyRouteMap{
		Routes:       make(map[NamespacedNameWithSection]sets.Set[string], gatewayMapSize),
		SectionIndex: make(map[types.NamespacedName]sets.Set[string], gatewayMapSize),
	}

	handledPolicies := make(map[types.NamespacedName]*egv1a1.BackendTrafficPolicy, policyMapSize)

	// Translate
	// 1. First translate Policies targeting RouteRules
	// 2. Next translate Policies targeting xRoutes
	// 3. Then translate Policies targeting Listeners
	// 4. Finally, the policies targeting Gateways

	// Build gateway policy maps, which are needed when processing the policies targeting xRoutes.
	t.buildGatewayPolicyMap(backendTrafficPolicies, gateways, gatewayMap, gatewayPolicyMap)

	// Process the policies targeting RouteRules
	for _, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			// If the target is not a gateway, then it's an xRoute. If the section name is defined, then it's a route rule.
			if currTarget.Kind != resource.KindGateway && currTarget.SectionName != nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processBackendTrafficPolicyForRoute(xdsIR,
					routeMap, gatewayRouteMap, gatewayPolicyMerged, gatewayPolicyMap, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Routes
	for _, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, routes, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			// If the target is not a gateway, then it's an xRoute. If the section name is not defined, then it's a route.
			if currTarget.Kind != resource.KindGateway && currTarget.SectionName == nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processBackendTrafficPolicyForRoute(xdsIR,
					routeMap, gatewayRouteMap, gatewayPolicyMerged, gatewayPolicyMap, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Listeners
	for _, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			// If the target is a gateway and the section name is defined, then it's a listener.
			if currTarget.Kind == resource.KindGateway && currTarget.SectionName != nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}
				t.processBackendTrafficPolicyForGateway(xdsIR,
					gatewayMap, gatewayRouteMap, gatewayPolicyMerged, policy, currTarget)
			}
		}
	}

	// Process the policies targeting Gateways
	for _, currPolicy := range backendTrafficPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			// If the target is a gateway and the section name is not defined, then it's a gateway.
			if currTarget.Kind == resource.KindGateway && currTarget.SectionName == nil {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = currPolicy
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}
				t.processBackendTrafficPolicyForGateway(xdsIR,
					gatewayMap, gatewayRouteMap, gatewayPolicyMerged, policy, currTarget)
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
) {
	for _, currPolicy := range backendTrafficPolicies {
		targetRefs := getPolicyTargetRefs(currPolicy.Spec.PolicyTargetReferences, gateways, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if currTarget.Kind == resource.KindGateway {
				// Check if the gateway exists
				key := types.NamespacedName{
					Name:      string(currTarget.Name),
					Namespace: currPolicy.Namespace,
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
						gateway.listeners,
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

func (t *Translator) processBackendTrafficPolicyForRoute(
	xdsIR resource.XdsIRMap,
	routeMap map[policyTargetRouteKey]*policyRouteTargetContext,
	gatewayRouteMap *GatewayPolicyRouteMap,
	gatewayPolicyMergedMap *GatewayPolicyRouteMap,
	gatewayPolicyMap map[NamespacedNameWithSection]*egv1a1.BackendTrafficPolicy,
	policy *egv1a1.BackendTrafficPolicy,
	currTarget gwapiv1.LocalPolicyTargetReferenceWithSectionName,
) {
	var (
		targetedRoute RouteContext
		resolveErr    *status.PolicyResolveError
	)

	targetedRoute, resolveErr = resolveBackendTrafficPolicyRouteTargetRef(policy, currTarget, routeMap)
	// Skip if the route is not found
	// It's not necessarily an error because the BackendTrafficPolicy may be
	// reconciled by multiple controllers. And the other controller may
	// have the target route.
	if targetedRoute == nil {
		return
	}

	// Find the Gateway that the route belongs to and add it to the
	// gatewayRouteMap and ancestor list, which will be used to check
	// policy overrides and populate its ancestor status.
	parentRefs := GetParentReferences(targetedRoute)
	ancestorRefs := make([]*gwapiv1.ParentReference, 0, len(parentRefs))
	// parentRefCtxs holds parent gateway/listener contexts for using in policy merge logic.
	parentRefCtxs := make([]*RouteParentContext, 0, len(parentRefs))
	for _, p := range parentRefs {
		if p.Kind == nil || *p.Kind == resource.KindGateway {
			namespace := targetedRoute.GetNamespace()
			if p.Namespace != nil {
				namespace = string(*p.Namespace)
			}

			mapKey := NamespacedNameWithSection{
				NamespacedName: types.NamespacedName{
					Name:      string(p.Name),
					Namespace: namespace,
				},
				SectionName: ptr.Deref(p.SectionName, ""),
			}
			if _, ok := gatewayRouteMap.Routes[mapKey]; !ok {
				gatewayRouteMap.Routes[mapKey] = make(sets.Set[string])
			}
			gatewayRouteMap.Routes[mapKey].Insert(utils.NamespacedName(targetedRoute).String())

			// Register section name to Gateway index for efficient lookup when retrieving overridden and merged targets
			if _, ok := gatewayRouteMap.SectionIndex[mapKey.NamespacedName]; !ok {
				gatewayRouteMap.SectionIndex[mapKey.NamespacedName] = make(sets.Set[string])
			}
			gatewayRouteMap.SectionIndex[mapKey.NamespacedName].Insert(string(mapKey.SectionName))

			// Do need a section name since the policy is targeting to a route.
			ancestorRef := getAncestorRefForPolicy(mapKey.NamespacedName, p.SectionName)
			ancestorRefs = append(ancestorRefs, &ancestorRef)

			// Only process parentRefs that were handled by this translator
			// (skip those referencing Gateways with different GatewayClasses)
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
		if err := t.translateBackendTrafficPolicyForRoute(policy, targetedRoute, currTarget, xdsIR, nil, nil); err != nil {
			status.SetTranslationErrorForPolicyAncestors(&policy.Status,
				ancestorRefs,
				t.GatewayControllerName,
				policy.Generation,
				status.Error2ConditionMsg(err),
			)
		}
	} else {
		for _, parentRefCtx := range parentRefCtxs {
			for _, listener := range parentRefCtx.listeners {
				gwNN := utils.NamespacedName(listener.gateway.Gateway)
				ancestorRef := getAncestorRefForPolicy(gwNN, &listener.Name)

				// Find Gateway listener level policy
				listenerMapKey := NamespacedNameWithSection{
					NamespacedName: gwNN,
					SectionName:    listener.Name,
				}
				listenerPolicy := gatewayPolicyMap[listenerMapKey]

				// Find Gateway level policy
				gwMapKey := NamespacedNameWithSection{
					NamespacedName: gwNN,
				}
				gwPolicy := gatewayPolicyMap[gwMapKey]
				if gwPolicy == nil && listenerPolicy == nil {
					// not found, fall back to the current policy
					if err := t.translateBackendTrafficPolicyForRoute(policy, targetedRoute, currTarget, xdsIR, &gwNN, &listener.Name); err != nil {
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

				parentPolicy := gwPolicy
				if listenerPolicy != nil {
					parentPolicy = listenerPolicy
				}
				// merge with parent policy
				if err := t.translateBackendTrafficPolicyForRouteWithMerge(
					policy, parentPolicy, currTarget, gwNN, &listener.Name,
					targetedRoute, xdsIR,
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

				// Record the merged routes for gateway
				if _, ok := gatewayPolicyMergedMap.Routes[listenerMapKey]; !ok {
					gatewayPolicyMergedMap.Routes[listenerMapKey] = make(sets.Set[string])
				}
				gatewayPolicyMergedMap.Routes[listenerMapKey].Insert(utils.NamespacedName(targetedRoute).String())

				// Register section name to Gateway index for efficient lookup when retrieving overridden and merged targets
				if _, ok := gatewayPolicyMergedMap.SectionIndex[listenerMapKey.NamespacedName]; !ok {
					gatewayPolicyMergedMap.SectionIndex[listenerMapKey.NamespacedName] = make(sets.Set[string])
				}
				gatewayPolicyMergedMap.SectionIndex[listenerMapKey.NamespacedName].Insert(string(listenerMapKey.SectionName))

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

	// Check if this policy is overridden by other policies targeting at route rule levels
	// If policy target is route rule, we can skip the check
	if currTarget.SectionName != nil {
		return
	}

	key := policyTargetRouteKey{
		Kind:      string(currTarget.Kind),
		Name:      string(currTarget.Name),
		Namespace: policy.Namespace,
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

func (t *Translator) processBackendTrafficPolicyForGateway(
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	gatewayRouteMap *GatewayPolicyRouteMap,
	gatewayPolicyMergedMap *GatewayPolicyRouteMap,
	policy *egv1a1.BackendTrafficPolicy,
	currTarget gwapiv1.LocalPolicyTargetReferenceWithSectionName,
) {
	var (
		targetedGateway *GatewayContext
		resolveErr      *status.PolicyResolveError
	)

	// Negative statuses have already been assigned so it's safe to skip
	targetedGateway, resolveErr = resolveBackendTrafficPolicyGatewayTargetRef(policy, currTarget, gatewayMap)
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

	// Set conditions for translation error if it got any
	if err := t.translateBackendTrafficPolicyForGateway(policy, currTarget, targetedGateway, xdsIR); err != nil {
		status.SetTranslationErrorForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation)

	overriddenMessage, mergedMessage := getOverriddenAndMergedTargetsMessageForGateway(
		gatewayMap[gatewayNN], gatewayRouteMap, gatewayPolicyMergedMap, currTarget.SectionName)

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
	policy *egv1a1.BackendTrafficPolicy,
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
	gateways map[types.NamespacedName]*policyGatewayTargetContext,
) (*GatewayContext, *status.PolicyResolveError) {
	// Check if the gateway exists
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: policy.Namespace,
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
			gateway.listeners,
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

func resolveBackendTrafficPolicyRouteTargetRef(
	policy *egv1a1.BackendTrafficPolicy,
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
	routes map[policyTargetRouteKey]*policyRouteTargetContext,
) (RouteContext, *status.PolicyResolveError) {
	// Check if the route exists
	key := policyTargetRouteKey{
		Kind:      string(target.Kind),
		Name:      string(target.Name),
		Namespace: policy.Namespace,
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
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
	xdsIR resource.XdsIRMap,
	policyTargetGatewayNN *types.NamespacedName,
	policyTargetListener *gwapiv1.SectionName,
) error {
	tf, errs := t.buildTrafficFeatures(policy)
	if tf == nil {
		// should not happen
		return nil
	}

	// Apply IR to all relevant routes
	for key, x := range xdsIR {
		// if gatewayNN is not nil, only apply to the specific gateway
		if policyTargetGatewayNN != nil && key != t.IRKey(*policyTargetGatewayNN) {
			// Skip if not the gateway wanted
			continue
		}
		t.applyTrafficFeatureToRoute(route, tf, errs, policy, target, x, policyTargetListener)
	}

	return errs
}

func (t *Translator) translateBackendTrafficPolicyForRouteWithMerge(
	policy, parentPolicy *egv1a1.BackendTrafficPolicy,
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
	policyTargetGatewayNN types.NamespacedName, policyTargetListener *gwapiv1.SectionName, route RouteContext,
	xdsIR resource.XdsIRMap,
) error {
	mergedPolicy, err := mergeBackendTrafficPolicy(policy, parentPolicy)
	if err != nil {
		return fmt.Errorf("error merging policies: %w", err)
	}

	// Build traffic features from the merged policy
	tf, errs := t.buildTrafficFeatures(mergedPolicy)
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
		tfGW, _ := t.buildTrafficFeatures(parentPolicy)
		tfRoute, _ := t.buildTrafficFeatures(policy)

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
		tfGW, _ := t.buildTrafficFeatures(parentPolicy)
		if tfGW != nil && tfGW.RateLimit != nil {
			// Use the gateway policy's rate limit with its original rule names
			tf.RateLimit = tfGW.RateLimit
		}
	}
	// Case 3: Only route policy has rate limits or neither has rate limits - use default behavior (tf already built from merged policy)

	x, ok := xdsIR[t.IRKey(policyTargetGatewayNN)]
	if !ok {
		// should not happen.
		return nil
	}
	t.applyTrafficFeatureToRoute(route, tf, errs, mergedPolicy, target, x, policyTargetListener)

	return nil
}

func (t *Translator) applyTrafficFeatureToRoute(route RouteContext,
	tf *ir.TrafficFeatures, errs error,
	policy *egv1a1.BackendTrafficPolicy,
	target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
	x *ir.Xds,
	policyTargetListener *gwapiv1.SectionName,
) {
	routeStatName := ""
	if tf.Telemetry != nil && tf.Telemetry.Metrics != nil {
		routeStatName = ptr.Deref(tf.Telemetry.Metrics.RouteStatName, "")
	}

	prefix := irRoutePrefix(route)
	for _, tcp := range x.TCP {
		// if listenerName is not nil, only apply to the specific listener
		if policyTargetListener != nil && string(*policyTargetListener) != tcp.Metadata.SectionName {
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
			}
		}
	}

	for _, udp := range x.UDP {
		// if listenerName is not nil, only apply to the specific listener
		if policyTargetListener != nil && string(*policyTargetListener) != udp.Metadata.SectionName {
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
		if policyTargetListener != nil && string(*policyTargetListener) != http.Metadata.SectionName {
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
						StatusCode: ptr.To(uint32(500)),
					}
					routesWithDirectResponse.Insert(r.Name)
					continue
				}

				r.Traffic = tf.DeepCopy()

				if localTo, err := buildClusterSettingsTimeout(&policy.Spec.ClusterSettings); err == nil {
					r.Traffic.Timeout = localTo
				}

				// Update the Host field in HealthCheck, now that we have access to the Route Hostname.
				r.Traffic.HealthCheck.SetHTTPHostIfAbsent(r.Hostname)

				if policy.Spec.UseClientProtocol != nil {
					r.UseClientProtocol = policy.Spec.UseClientProtocol
				}
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

func mergeBackendTrafficPolicy(routePolicy, gwPolicy *egv1a1.BackendTrafficPolicy) (*egv1a1.BackendTrafficPolicy, error) {
	if routePolicy.Spec.MergeType == nil || gwPolicy == nil {
		return routePolicy, nil
	}

	return utils.Merge[*egv1a1.BackendTrafficPolicy](gwPolicy, routePolicy, *routePolicy.Spec.MergeType)
}

func (t *Translator) buildTrafficFeatures(policy *egv1a1.BackendTrafficPolicy) (*ir.TrafficFeatures, error) {
	var (
		rl          *ir.RateLimit
		lb          *ir.LoadBalancer
		pp          *ir.ProxyProtocol
		hc          *ir.HealthCheck
		cb          *ir.CircuitBreaker
		fi          *ir.FaultInjection
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

	if ro, err = t.buildResponseOverride(policy); err != nil {
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

	ds = translateDNS(&policy.Spec.ClusterSettings, utils.NamespacedName(policy).String())

	return &ir.TrafficFeatures{
		RateLimit:         rl,
		LoadBalancer:      lb,
		ProxyProtocol:     pp,
		HealthCheck:       hc,
		CircuitBreaker:    cb,
		FaultInjection:    fi,
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
		Telemetry:         policy.Spec.Telemetry,
	}, errs
}

func (t *Translator) translateBackendTrafficPolicyForGateway(
	policy *egv1a1.BackendTrafficPolicy, target gwapiv1.LocalPolicyTargetReferenceWithSectionName,
	gateway *GatewayContext, xdsIR resource.XdsIRMap,
) error {
	tf, errs := t.buildTrafficFeatures(policy)
	if tf == nil {
		// should not happen
		return errs
	}

	routeStatName := ""
	if tf.Telemetry != nil && tf.Telemetry.Metrics != nil {
		routeStatName = ptr.Deref(tf.Telemetry.Metrics.RouteStatName, "")
	}

	// Apply IR to all the routes within the specific Gateway
	// If the feature is already set, then skip it, since it must have
	// set by a policy attaching to the route
	irKey := t.getIRKey(gateway.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]

	policyTarget := irStringKey(policy.Namespace, string(target.Name))

	for _, tcp := range x.TCP {
		gatewayName := tcp.Name[0:strings.LastIndex(tcp.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		// If specified the sectionName must match listenerName from ir listener metadata.
		if target.SectionName != nil && string(*target.SectionName) != tcp.Metadata.SectionName {
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
		}
	}

	for _, udp := range x.UDP {
		gatewayName := udp.Name[0:strings.LastIndex(udp.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		// If specified the sectionName must match listenerName from ir listener metadata.
		if target.SectionName != nil && string(*target.SectionName) != udp.Metadata.SectionName {
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
		gatewayName := http.Name[0:strings.LastIndex(http.Name, "/")]
		if t.MergeGateways && gatewayName != policyTarget {
			continue
		}

		// If specified the sectionName must match listenerName from ir listener metadata.
		if target.SectionName != nil && string(*target.SectionName) != http.Metadata.SectionName {
			continue
		}

		// A Policy targeting the most specific scope(xRoute) wins over a policy
		// targeting a lesser specific scope(Gateway).
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
					StatusCode: ptr.To(uint32(500)),
				}
				routesWithDirectResponse.Insert(r.Name)
				continue
			}

			r.Traffic = tf.DeepCopy()

			if localTo, err := buildClusterSettingsTimeout(&policy.Spec.ClusterSettings); err == nil {
				r.Traffic.Timeout = localTo
			}

			// Update the Host field in HealthCheck, now that we have access to the Route Hostname.
			r.Traffic.HealthCheck.SetHTTPHostIfAbsent(r.Hostname)

			if policy.Spec.UseClientProtocol != nil {
				r.UseClientProtocol = policy.Spec.UseClientProtocol
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

	return errs
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
	for _, rule := range local.Rules {
		if len(rule.ClientSelectors) == 0 {
			if defaultLimit != nil {
				return nil, fmt.Errorf("local rateLimit can not have more than one rule without clientSelectors")
			}
			defaultLimit = &ir.RateLimitValue{
				Requests: rule.Limit.Requests,
				Unit:     ir.RateLimitUnit(rule.Limit.Unit),
			}
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

		irRule, err = buildRateLimitRule(rule)
		if err != nil {
			return nil, err
		}
		// Set the Name field as <policy-ns>/<policy-name>/rule/<rule-index>
		irRule.Name = irRuleName(policy.Namespace, policy.Name, i)
		irRules = append(irRules, irRule)
	}

	rateLimit := &ir.RateLimit{
		Local: &ir.LocalRateLimit{
			Default: *defaultLimit,
			Rules:   irRules,
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
		irRules[i], err = buildRateLimitRule(rule)
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

func buildRateLimitRule(rule egv1a1.RateLimitRule) (*ir.RateLimitRule, error) {
	irRule := &ir.RateLimitRule{
		Limit: ir.RateLimitValue{
			Requests: rule.Limit.Requests,
			Unit:     ir.RateLimitUnit(rule.Limit.Unit),
		},
		HeaderMatches: make([]*ir.StringMatch, 0),
		MethodMatches: make([]*ir.StringMatch, 0),
		Shared:        rule.Shared,
		ShadowMode:    rule.ShadowMode,
	}

	for _, match := range rule.ClientSelectors {
		if len(match.Headers) == 0 && len(match.Methods) == 0 &&
			match.Path == nil && match.SourceCIDR == nil {
			return nil, fmt.Errorf(
				"unable to translate rateLimit. At least one of the" +
					" header or method or path or sourceCIDR must be specified")
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
				Exact:  ptr.To(string(method.Value)),
				Invert: method.Invert,
			})
		}

		if match.Path != nil {
			switch ptr.Deref(match.Path.Type, gwapiv1.PathMatchPathPrefix) {
			case gwapiv1.PathMatchPathPrefix:
				if match.Path.Value == "/" {
					irRule.PathMatch = &ir.StringMatch{
						Prefix: ptr.To(match.Path.Value),
						Invert: match.Path.Invert,
					}
				} else {
					// envoy ratelimit HeaderMatcher doesn't support PathSeparatedPrefix like route matching,
					// so we use regex to achieve the same path-separated prefix behavior.
					irRule.PathMatch = &ir.StringMatch{
						SafeRegex: ptr.To(regex.PathSeparatedPrefixRegex(match.Path.Value)),
						Invert:    match.Path.Invert,
					}
				}
			case gwapiv1.PathMatchExact:
				irRule.PathMatch = &ir.StringMatch{
					Exact:  ptr.To(match.Path.Value),
					Invert: match.Path.Invert,
				}
			case gwapiv1.PathMatchRegularExpression:
				if err := regex.Validate(match.Path.Value); err != nil {
					return nil, err
				}
				irRule.PathMatch = &ir.StringMatch{
					SafeRegex: ptr.To(match.Path.Value),
					Invert:    match.Path.Invert,
				}
			default:
				return nil, fmt.Errorf("unable to translate rateLimit: invalid path type.")
			}
		}

		if match.SourceCIDR != nil {
			// distinct means that each IP Address within the specified Source IP CIDR is treated as a
			// distinct client selector and uses a separate rate limit bucket/counter.
			distinct := false
			sourceCIDR := match.SourceCIDR.Value
			if match.SourceCIDR.Type != nil && *match.SourceCIDR.Type == egv1a1.SourceMatchDistinct {
				distinct = true
			}

			cidrMatch, err := parseCIDR(sourceCIDR)
			if err != nil {
				return nil, fmt.Errorf("unable to translate rateLimit: %w", err)
			}
			cidrMatch.Distinct = distinct
			irRule.CIDRMatch = cidrMatch
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
		ret.Format = ptr.To(fmt.Sprintf("%%DYNAMIC_METADATA(%s:%s)%%",
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
				fi.Abort.GrpcStatus = policy.Spec.FaultInjection.Abort.GrpcStatus
			}
			if policy.Spec.FaultInjection.Abort.HTTPStatus != nil {
				fi.Abort.HTTPStatus = policy.Spec.FaultInjection.Abort.HTTPStatus
			}
		}
	}
	return fi
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

func (t *Translator) buildResponseOverride(policy *egv1a1.BackendTrafficPolicy) (*ir.ResponseOverride, error) {
	if len(policy.Spec.ResponseOverride) == 0 {
		return nil, nil
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
				redirect.Hostname = ptr.To(string(*ro.Redirect.Hostname))
			}
			if ro.Redirect.Port != nil {
				redirect.Port = ptr.To(uint32(*ro.Redirect.Port))
			}
			if ro.Redirect.StatusCode != nil {
				redirect.StatusCode = ptr.To(int32(*ro.Redirect.StatusCode))
			}

			rules = append(rules, ir.ResponseOverrideRule{
				Name:     defaultResponseOverrideRuleName(policy, index),
				Match:    match,
				Redirect: redirect,
			})
		} else {
			response := &ir.CustomResponse{
				ContentType: ro.Response.ContentType,
			}

			if ro.Response.StatusCode != nil {
				response.StatusCode = ptr.To(uint32(*ro.Response.StatusCode))
			}

			var err error
			response.Body, err = t.getCustomResponseBody(ro.Response.Body, policy.Namespace)
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
			b, dataOk := cm.Data["response.body"]
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
				return nil, fmt.Errorf("can't find the key response.body in the referenced configmap %s", body.ValueRef.Name)
			}
		} else {
			return nil, fmt.Errorf("can't find the referenced configmap %s", body.ValueRef.Name)
		}
	} else if body.Inline != nil {
		inlineValue := []byte(*body.Inline)
		return inlineValue, nil
	}

	return nil, nil
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
		irCompression := make([]*ir.Compression, 0, len(compressor))
		for i, c := range compressor {
			// Only add compression if the corresponding compressor not null
			if (c.Type == egv1a1.GzipCompressorType && c.Gzip != nil) ||
				(c.Type == egv1a1.BrotliCompressorType && c.Brotli != nil) ||
				(c.Type == egv1a1.ZstdCompressorType && c.Zstd != nil) {
				irCompression = append(irCompression, &ir.Compression{
					Type:        c.Type,
					ChooseFirst: i == 0, // only the first compressor is marked as ChooseFirst
				})
			}
		}
		return irCompression
	}

	// Fallback to the deprecated Compression field
	if compression == nil {
		return nil
	}
	irCompression := make([]*ir.Compression, 0, len(compression))
	for i, c := range compression {
		irCompression = append(irCompression, &ir.Compression{
			Type:        c.Type,
			ChooseFirst: i == 0, // only the first compressor is marked as ChooseFirst
		})
	}

	return irCompression
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
