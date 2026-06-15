// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"bytes"
	//nolint:gosec // SHA1 is required to validate htpasswd {SHA} format.
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/mail"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	perr "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/regex"
)

const (
	defaultRedirectURL           = "%REQ(x-forwarded-proto)%://%REQ(:authority)%/oauth2/callback"
	defaultRedirectPath          = "/oauth2/callback"
	defaultLogoutPath            = "/logout"
	defaultForwardAccessToken    = false
	defaultRefreshToken          = true
	defaultPassThroughAuthHeader = false
	defaultOIDCHTTPTimeout       = 5 * time.Second

	// nolint: gosec
	oidcHMACSecretName = "envoy-oidc-hmac"
	oidcHMACSecretKey  = "hmac-secret"
	// JWKSConfigMapKey is the key used in ConfigMaps to store JWKS data
	JWKSConfigMapKey = "jwks"
)

// deprecatedFieldsUsedInSecurityPolicy returns a map of deprecated field paths to their alternatives.
func deprecatedFieldsUsedInSecurityPolicy(policy *egv1a1.SecurityPolicy) map[string]string {
	deprecatedFields := make(map[string]string)
	if policy.Spec.TargetRef != nil {
		deprecatedFields["spec.targetRef"] = "spec.targetRefs"
	}
	if policy.Spec.ExtAuth != nil {
		if policy.Spec.ExtAuth.GRPC != nil && policy.Spec.ExtAuth.GRPC.BackendRef != nil {
			deprecatedFields["spec.extAuth.grpc.backendRef"] = "spec.extAuth.grpc.backendRefs"
		}
		if policy.Spec.ExtAuth.HTTP != nil && policy.Spec.ExtAuth.HTTP.BackendRef != nil {
			deprecatedFields["spec.extAuth.http.backendRef"] = "spec.extAuth.http.backendRefs"
		}
	}
	return deprecatedFields
}

func (t *Translator) ProcessSecurityPolicies(
	securityPolicies []*egv1a1.SecurityPolicy,
	gateways []*GatewayContext,
	routes []RouteContext,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) []*egv1a1.SecurityPolicy {
	// Cache is only reused during one translation across multiple routes and gateways.
	// The failed fetches will be retried in the next translation when the provider resources are reconciled again.
	t.oidcDiscoveryCache = newOIDCDiscoveryCache()

	// SecurityPolicies are already sorted by the provider layer

	// First build a map out of the routes and gateways for faster lookup since users might have thousands of routes or more.
	// For gateways this probably isn't quite as necessary.
	routeMapSize := len(routes)
	gatewayMapSize := len(gateways)
	policyMapSize := len(securityPolicies)
	listenerSetMapSize := len(resources.ListenerSets)

	// Pre-allocate result slice and maps with estimated capacity to reduce memory allocations
	res := make([]*egv1a1.SecurityPolicy, 0, len(securityPolicies))
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

	handledPolicies := make(map[types.NamespacedName]*egv1a1.SecurityPolicy, policyMapSize)

	// Map of attached Policy to Gateway. Used for policy merge process.
	gatewayPolicyMap := make(map[NamespacedNameWithSection]*egv1a1.SecurityPolicy, gatewayMapSize)

	// Map of attached Policy to ListenerSet. Used for policy merge process.
	listenerSetPolicyMap := make(map[NamespacedNameWithSection]*egv1a1.SecurityPolicy, listenerSetMapSize)

	// overrides records child scopes whose policies displace policies attached
	// to their parent scopes.
	overrides := newPolicyScopeGraph()

	// merged records Route scopes whose policies were merged into policies
	// attached to their parent scopes.
	merged := newPolicyScopeGraph()

	// Translate
	// 1. First translate Policies targeting RouteRules
	// 2. Next translate Policies targeting xRoutes
	// 3. Then translate Policies targeting ListenerSet Listeners
	// 4. Then translate Policies targeting ListenerSets
	// 5. Then translate Policies targeting Gateway Listeners
	// 6. Finally, the policies targeting Gateways

	// Build gateway policy maps, which are needed when processing the policies targeting xRoutes.
	t.buildGatewayPolicyMapForSecurity(securityPolicies, gateways, gatewayMap, gatewayPolicyMap, resources.ReferenceGrants)
	// Build ListenerSet policy maps, which are needed when processing the policies targeting xRoutes.
	t.buildListenerSetSecurityPolicyMap(securityPolicies, listenerSetMap, listenerSetPolicyMap, resources)

	// Process the policies targeting RouteRules (HTTP + TCP)
	for i, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isRouteRule(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = securityPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForRoute(resources, xdsIR, routeMap, listenerSetMap, gatewayPolicyMap, listenerSetPolicyMap, overrides, merged, policy, currTarget)
			}
		}
	}
	// Process the policies targeting xRoutes (HTTP + TCP)
	for i, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			routes,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindSecurityPolicy,
			currPolicy.Namespace,
			t.GetNamespace)
		for _, currTarget := range targetRefs {
			if isRoute(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = securityPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForRoute(resources, xdsIR, routeMap, listenerSetMap, gatewayPolicyMap, listenerSetPolicyMap, overrides, merged, policy, currTarget)
			}
		}
	}
	// Process the policies targeting ListenerSets Listeners
	for i, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isListenerSetListener(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = securityPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForListenerSet(resources, xdsIR, gatewayMap, listenerSetMap, overrides, merged, policy, currTarget)
			}
		}
	}
	// Process the policies targeting ListenerSets
	for i, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			resources.ListenerSets,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindSecurityPolicy,
			currPolicy.Namespace,
			t.GetNamespace,
		)
		for _, currTarget := range targetRefs {
			if isListenerSet(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = securityPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForListenerSet(resources, xdsIR, gatewayMap, listenerSetMap, overrides, merged, policy, currTarget)
			}
		}
	}
	// Process the policies targeting Gateway Listeners
	for i, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		// Only resolve TargetRefs from targetRefs field since TargetSelectors can't specify sectionName.
		targetRefs := resolvePolicyTargetsFromReferences(currPolicy.Spec.PolicyTargetReferences, currPolicy.Namespace)
		for _, currTarget := range targetRefs {
			if isListener(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = securityPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForGateway(resources, xdsIR, gatewayMap, overrides, merged, policy, currTarget)
			}
		}
	}
	// Process the policies targeting Gateways
	for i, currPolicy := range securityPolicies {
		policyName := utils.NamespacedName(currPolicy)
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			gateways,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindSecurityPolicy,
			currPolicy.Namespace,
			t.GetNamespace)

		for _, currTarget := range targetRefs {
			if isGateway(currTarget) {
				policy, found := handledPolicies[policyName]
				if !found {
					policy = securityPolicies[i]
					handledPolicies[policyName] = policy
					res = append(res, policy)
				}

				t.processSecurityPolicyForGateway(resources, xdsIR, gatewayMap, overrides, merged, policy, currTarget)
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

func (t *Translator) buildGatewayPolicyMapForSecurity(
	securityPolicies []*egv1a1.SecurityPolicy,
	gateways []*GatewayContext,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	gatewayPolicyMap map[NamespacedNameWithSection]*egv1a1.SecurityPolicy,
	referenceGrants []*gwapiv1b1.ReferenceGrant,
) {
	for _, currPolicy := range securityPolicies {
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			gateways,
			referenceGrants,
			egv1a1.GroupName,
			egv1a1.KindSecurityPolicy,
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

				// Only store the first policy for this Gateway/Listener - conflicts are handled elsewhere
				if _, ok := gatewayPolicyMap[mapKey]; ok {
					continue
				}
				gatewayPolicyMap[mapKey] = currPolicy
			}
		}
	}
}

// buildListenerSetSecurityPolicyMap populates listenerSetPolicyMap with the
// first SecurityPolicy attached to each (ListenerSet, sectionName) pair.
// Subsequent conflicting attachments are reported elsewhere; this map is the
// source of truth used by the merge step to find the closest parent policy
// for routes attached via a ListenerSet.
func (t *Translator) buildListenerSetSecurityPolicyMap(
	securityPolicies []*egv1a1.SecurityPolicy,
	listenerSetMap map[types.NamespacedName]*policyListenerSetTargetContext,
	listenerSetPolicyMap map[NamespacedNameWithSection]*egv1a1.SecurityPolicy,
	resources *resource.Resources,
) {
	for _, currPolicy := range securityPolicies {
		targetRefs := resolvePolicyTargets(
			currPolicy.Spec.PolicyTargetReferences,
			resources.ListenerSets,
			resources.ReferenceGrants,
			egv1a1.GroupName,
			egv1a1.KindSecurityPolicy,
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

func (t *Translator) processSecurityPolicyForRoute(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	routeMap map[policyTargetRouteKey]*policyRouteTargetContext,
	listenerSetMap map[types.NamespacedName]*policyListenerSetTargetContext,
	gatewayPolicyMap map[NamespacedNameWithSection]*egv1a1.SecurityPolicy,
	listenerSetPolicyMap map[NamespacedNameWithSection]*egv1a1.SecurityPolicy,
	overrides policyScopeGraph,
	merged policyScopeGraph,
	policy *egv1a1.SecurityPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targetedRoute RouteContext
		resolveErr    *status.PolicyResolveError
	)

	targetedRoute, resolveErr = resolveSecurityPolicyRouteTargetRef(currTarget, routeMap)
	// Skip if the route is not found
	// It's not necessarily an error because the SecurityPolicy may be
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

			// Only process parentRefs that were handled by this translator
			// (skip those referencing Gateways with different GatewayClasses)
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

	// Protocol-specific validation: pick the appropriate validator and message,
	// then run it once to keep the flow linear and easier to read.
	validator := validateSecurityPolicy
	errMsg := "invalid SecurityPolicy"
	if currTarget.Kind == resource.KindTCPRoute {
		validator = validateSecurityPolicyForTCP
		errMsg = "invalid SecurityPolicy for TCP route"
	}
	if err := validator(policy); err != nil {
		status.SetTranslationErrorForPolicyAncestors(&policy.Status,
			ancestorRefs,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(fmt.Errorf("%s: %w", errMsg, err)),
		)

		return
	}

	// Check if merging is enabled
	if policy.Spec.MergeType == nil {
		// No merging - use existing translation logic
		if err := t.translateSecurityPolicyForRoute(policy, &securityPolicyOwners{}, targetedRoute, currTarget, resources, xdsIR, nil); err != nil {
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
					parentPolicy *egv1a1.SecurityPolicy
					parentScope  policyScope
					ancestorRef  gwapiv1.ParentReference
				)
				if listener.isFromListenerSet() {
					lsNN := types.NamespacedName{
						Namespace: listener.listenerSet.Namespace,
						Name:      listener.listenerSet.Name,
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

					listenerKey := NamespacedNameWithSection{NamespacedName: gwNN, SectionName: listener.Name}
					gwKey := NamespacedNameWithSection{NamespacedName: gwNN}

					if p, ok := gatewayPolicyMap[listenerKey]; ok {
						parentPolicy, parentScope = p, gatewayListenerScope(gwNN, listener.Name)
					} else if p, ok := gatewayPolicyMap[gwKey]; ok {
						parentPolicy, parentScope = p, gatewayScope(gwNN)
					}
				}

				if parentPolicy == nil {
					// No parent policy found, fall back to current policy
					if err := t.translateSecurityPolicyForRoute(policy, &securityPolicyOwners{}, targetedRoute, currTarget, resources, xdsIR, listener); err != nil {
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

				// Merge with parent policy
				mergedPolicy, owners, err := mergeSecurityPolicy(policy, parentPolicy)
				if err != nil {
					status.SetConditionForPolicyAncestor(&policy.Status,
						&ancestorRef,
						t.GatewayControllerName,
						gwapiv1.PolicyConditionAccepted, metav1.ConditionFalse,
						egv1a1.PolicyReasonInvalid,
						fmt.Sprintf("error merging policies: %v", err),
						policy.Generation,
					)
					continue
				}

				if err := validator(mergedPolicy); err != nil {
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

				// Apply merged policy
				if err := t.translateSecurityPolicyForRoute(mergedPolicy, owners, targetedRoute, currTarget, resources, xdsIR, listener); err != nil {
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
	if deprecatedFields := deprecatedFieldsUsedInSecurityPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestors(&policy.Status, ancestorRefs, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}

	// Check if this policy is overridden by other policies targeting at route rule levels
	// If policy target is route rule, we can skip the check
	if currTarget.SectionName != nil {
		return
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
			"This policy is being overridden by other securityPolicies for "+overriddenTargetsMessage,
			policy.Generation,
		)
	}
}

func (t *Translator) processSecurityPolicyForListenerSet(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	listenerSetMap map[types.NamespacedName]*policyListenerSetTargetContext,
	overrides policyScopeGraph,
	merged policyScopeGraph,
	policy *egv1a1.SecurityPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targeted   *gwapiv1.ListenerSet
		resolveErr *status.PolicyResolveError
	)

	targeted, resolveErr = resolveSecurityPolicyListenerSetTargetRef(currTarget, listenerSetMap)
	// Skip if the ListenerSet is not found
	// It's not necessarily an error because the SecurityPolicy may be
	// reconciled by multiple controllers. And the other controller may
	// have the target ListenerSet.
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

	if err := t.translateSecurityPolicyForListenerSet(policy, gateway.GatewayContext, targeted, currTarget, resources, xdsIR); err != nil {
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
			"This policy is being merged by other securityPolicies for "+mergedMessage,
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
			"This policy is being overridden by other securityPolicies for "+overriddenMessage,
			policy.Generation,
		)
	}

	// Check for deprecated fields and set warning if any are found
	if deprecatedFields := deprecatedFieldsUsedInSecurityPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}
}

func (t *Translator) processSecurityPolicyForGateway(
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	gatewayMap map[types.NamespacedName]*policyGatewayTargetContext,
	overrides policyScopeGraph,
	merged policyScopeGraph,
	policy *egv1a1.SecurityPolicy,
	currTarget policyTargetReferenceWithSectionName,
) {
	var (
		targetedGateway *GatewayContext
		resolveErr      *status.PolicyResolveError
	)

	targetedGateway, resolveErr = resolveSecurityPolicyGatewayTargetRef(currTarget, gatewayMap)
	// Skip if the gateway is not found
	// It's not necessarily an error because the SecurityPolicy may be
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
	if err := t.translateSecurityPolicyForGateway(policy, targetedGateway, currTarget, resources, xdsIR); err != nil {
		status.SetTranslationErrorForPolicyAncestor(&policy.Status,
			&ancestorRef,
			t.GatewayControllerName,
			policy.Generation,
			status.Error2ConditionMsg(err),
		)
	}

	// Set Accepted condition if it is unset
	status.SetAcceptedForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation)

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
			"This policy is being merged by other securityPolicies for "+mergedMessage,
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
			"This policy is being overridden by other securityPolicies for "+overriddenMessage,
			policy.Generation,
		)
	}

	// Check for deprecated fields and set warning if any are found
	if deprecatedFields := deprecatedFieldsUsedInSecurityPolicy(policy); len(deprecatedFields) > 0 {
		status.SetDeprecatedFieldsWarningForPolicyAncestor(&policy.Status, &ancestorRef, t.GatewayControllerName, policy.Generation, deprecatedFields)
	}
}

// validateSecurityPolicy validates the SecurityPolicy.
// It checks some constraints that are not covered by the CRD schema validation.
func validateSecurityPolicy(p *egv1a1.SecurityPolicy) error {
	apiKeyAuth := p.Spec.APIKeyAuth
	if apiKeyAuth != nil {
		if err := validateAPIKeyAuth(apiKeyAuth); err != nil {
			return err
		}
	}

	oidc := p.Spec.OIDC
	jwt := p.Spec.JWT
	if oidc != nil && oidc.PassThroughAuthHeader != nil && *oidc.PassThroughAuthHeader {
		if jwt == nil {
			return errors.New("the OIDC.PassThroughAuthHeader setting must be used in conjunction with JWT settings")
		}

		hasValidJwtExtractor := false
		for _, provider := range jwt.Providers {
			// When ExtractFrom is not specified it falls back to looking at the "Authorization: Bearer ..." header
			if provider.ExtractFrom == nil || len(provider.ExtractFrom.Headers) > 0 {
				hasValidJwtExtractor = true
				break
			}
		}
		if !hasValidJwtExtractor {
			return errors.New("the OIDC.PassThroughAuthHeader setting must be used in conjunction with a JWT provider that is configured to read from a header")
		}
	}

	basicAuth := p.Spec.BasicAuth
	if basicAuth != nil {
		if err := validateBasicAuth(basicAuth); err != nil {
			return err
		}
	}
	return nil
}

// validateSecurityPolicyForTCP ensures SecurityPolicy usage on TCP is compatible.
//
// TCP supports Authorization with ClientCIDRs ONLY.
// - Principals.JWT      => invalid (HTTP-only)
// - Principals.Headers  => invalid (HTTP-only)
// - Empty/no Authorization is allowed and results in no-op on TCP.
// Returns an error when any HTTP-only field is present or CIDRs are invalid.
func validateSecurityPolicyForTCP(p *egv1a1.SecurityPolicy) error {
	if p.Spec.CORS != nil || p.Spec.JWT != nil || p.Spec.OIDC != nil || p.Spec.APIKeyAuth != nil || p.Spec.BasicAuth != nil || p.Spec.ExtAuth != nil {
		return fmt.Errorf("only authorization is supported for TCP (routes/listeners)")
	}
	if p.Spec.Authorization == nil || len(p.Spec.Authorization.Rules) == 0 {
		return nil
	}
	for i := range p.Spec.Authorization.Rules {
		rule := &p.Spec.Authorization.Rules[i]
		if rule.CEL != nil {
			return fmt.Errorf("rule %d: CEL not supported for TCP", i)
		}
		if rule.Principal == nil {
			continue
		}
		if rule.Principal.JWT != nil {
			return fmt.Errorf("rule %d: JWT not supported for TCP", i)
		}
		if len(rule.Principal.Headers) > 0 {
			return fmt.Errorf("rule %d: headers not supported for TCP", i)
		}
		if len(rule.Principal.ClientIPGeoLocations) > 0 {
			return fmt.Errorf("rule %d: clientIPGeoLocations not supported for TCP", i)
		}
		if err := validateCIDRs(rule.Principal.ClientCIDRs); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}
	return nil
}

// validateCIDRs validates CIDR strings for TCP authorization rules.
func validateCIDRs(cidrs []egv1a1.CIDR) error {
	for _, c := range cidrs {
		if _, _, err := net.ParseCIDR(string(c)); err != nil {
			return fmt.Errorf("invalid ClientCIDR %q: %w", c, err)
		}
	}
	return nil
}

func validateAPIKeyAuth(apiKeyAuth *egv1a1.APIKeyAuth) error {
	for _, keySource := range apiKeyAuth.ExtractFrom {
		// only one of headers, params or cookies is supposed to be specified.
		if len(keySource.Headers) > 0 && len(keySource.Params) > 0 ||
			len(keySource.Headers) > 0 && len(keySource.Cookies) > 0 ||
			len(keySource.Params) > 0 && len(keySource.Cookies) > 0 {
			return errors.New("only one of headers, params or cookies must be specified")
		}
	}
	return nil
}

// validateBasicAuth validates the BasicAuth configuration.
// Currently, we only validate that the secret exists, but we don't validate
// the content of the secret. This function will be called when the security policy
// is being processed, but before the secret is actually read.
func validateBasicAuth(_ *egv1a1.BasicAuth) error {
	// The actual validation of the htpasswd format will happen when the secret is read
	// in the buildBasicAuth function.
	return nil
}

func resolveSecurityPolicyGatewayTargetRef(
	target policyTargetReferenceWithSectionName,
	gateways map[types.NamespacedName]*policyGatewayTargetContext,
) (*GatewayContext, *status.PolicyResolveError) {
	// Find the Gateway
	key := types.NamespacedName{
		Name:      string(target.Name),
		Namespace: string(target.Namespace),
	}
	gateway, ok := gateways[key]

	// Gateway not found
	// It's not an error if the gateway is not found because the SecurityPolicy
	// may be reconciled by multiple controllers, and the gateway may not be managed
	// by this controller.
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
			message := fmt.Sprintf("Unable to target Gateway %s, another SecurityPolicy has already attached to it",
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
			message := fmt.Sprintf("Unable to target Listener %s/%s, another SecurityPolicy has already attached to it",
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

func resolveSecurityPolicyListenerSetTargetRef(
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
	// It's not an error if the ListenerSet is not found because the SecurityPolicy
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
			message := fmt.Sprintf("Unable to target ListenerSet %s, another SecurityPolicy has already attached to it",
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
			message := fmt.Sprintf("Unable to target Listener %s/%s, another SecurityPolicy has already attached to it",
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

func resolveSecurityPolicyRouteTargetRef(
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
	// It's not an error if the gateway is not found because the SecurityPolicy
	// may be reconciled by multiple controllers, and the gateway may not be managed
	// by this controller.
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
			message := fmt.Sprintf("Unable to target %s %s, another SecurityPolicy has already attached to it",
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
			message := fmt.Sprintf("Unable to target RouteRule %s/%s, another SecurityPolicy has already attached to it",
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

func (t *Translator) translateSecurityPolicyForRoute(
	policy *egv1a1.SecurityPolicy,
	owners *securityPolicyOwners,
	route RouteContext,
	target policyTargetReferenceWithSectionName,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	targetListener *ListenerContext,
) error {
	// Build IR
	var (
		cors               *ir.CORS
		apiKeyAuth         *ir.APIKeyAuth
		basicAuth          *ir.BasicAuth
		authorization      *ir.Authorization
		err, errs          error
		hasNonExtAuthError bool
	)

	if policy.Spec.CORS != nil {
		cors = t.buildCORS(policy.Spec.CORS)
	}

	if policy.Spec.BasicAuth != nil {
		if basicAuth, err = t.buildBasicAuth(
			policy,
			owners,
			resources,
		); err != nil {
			err = perr.WithMessage(err, "BasicAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.APIKeyAuth != nil {
		if apiKeyAuth, err = t.buildAPIKeyAuth(
			policy,
			owners,
			resources,
		); err != nil {
			err = perr.WithMessage(err, "APIKeyAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.Authorization != nil {
		if authorization, err = t.buildAuthorization(policy, owners); err != nil {
			err = perr.WithMessage(err, "Authorization")
			errs = errors.Join(errs, err)
		}
	}

	hasNonExtAuthError = errs != nil

	// Apply IR to all relevant routes
	prefix := irRoutePrefix(route)
	parentRefs := GetParentReferences(route)
	routesWithDirectResponse := sets.New[string]()

	var targetListenerName string
	var targetGatewayNN types.NamespacedName
	if targetListener != nil {
		targetListenerName = irListenerName(targetListener)
		targetGatewayNN = types.NamespacedName{
			Namespace: targetListener.gateway.Namespace,
			Name:      targetListener.gateway.Name,
		}
	}
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

		// If targetListener is set, only apply within its parent Gateway.
		if targetListener != nil {
			gtwNN := types.NamespacedName{
				Namespace: gtwCtx.Namespace,
				Name:      gtwCtx.Name,
			}
			if gtwNN != targetGatewayNN {
				continue
			}
		}

		var extAuth *ir.ExtAuth
		var extAuthErr error
		if policy.Spec.ExtAuth != nil {
			if extAuth, extAuthErr = t.buildExtAuth(
				policy,
				owners,
				resources,
				gtwCtx,
			); extAuthErr != nil {
				extAuthErr = perr.WithMessage(extAuthErr, "ExtAuth")
				errs = errors.Join(errs, extAuthErr)
			}
		}

		var oidc *ir.OIDC
		if policy.Spec.OIDC != nil {
			if oidc, err = t.buildOIDC(
				policy,
				owners,
				resources,
				gtwCtx,
			); err != nil {
				err = perr.WithMessage(err, "OIDC")
				errs = errors.Join(errs, err)
				hasNonExtAuthError = true
			}
		}

		var jwt *ir.JWT
		if policy.Spec.JWT != nil {
			if jwt, err = t.buildJWT(
				policy,
				owners,
				resources,
				gtwCtx,
			); err != nil {
				err = perr.WithMessage(err, "JWT")
				errs = errors.Join(errs, err)
				hasNonExtAuthError = true
			}
		}

		// Pre-create security features to avoid repeated allocations
		securityFeatures := &ir.SecurityFeatures{
			CORS:          cors,
			JWT:           jwt,
			OIDC:          oidc,
			APIKeyAuth:    apiKeyAuth,
			BasicAuth:     basicAuth,
			ExtAuth:       extAuth,
			Authorization: authorization,
		}

		irKey := t.getIRKey(gtwCtx.Gateway)
		switch route.GetRouteType() {
		case resource.KindTCPRoute:
			for _, listener := range parentRefCtx.listeners {
				// If targetListener is set, only apply to that exact listener.
				if targetListener != nil && targetListenerName != irListenerName(listener) {
					continue
				}
				tl := xdsIR[irKey].GetTCPListener(irListenerName(listener))
				for _, r := range tl.Routes {
					// If target.SectionName is specified it must match the route-rule section name
					// in the IR. For HTTP/GRPC routes this is r.Metadata.SectionName; for TCP
					// routes the section name is currently stored on r.Destination.Metadata.SectionName.
					if target.SectionName != nil && string(*target.SectionName) != r.Destination.Metadata.SectionName {
						continue
					}

					if r.Authorization != nil {
						continue
					}
					// Only authorization for TCP
					if authorization != nil {
						authCopy := *authorization
						r.Authorization = &authCopy
					}
				}
			}
		case resource.KindHTTPRoute, resource.KindGRPCRoute:
			var (
				hasBaseErrs    = errs != nil
				directResponse = &ir.CustomResponse{StatusCode: new(uint32(500))}
			)
			for _, listener := range parentRefCtx.listeners {
				// If targetListener is set, only apply to that exact listener.
				if targetListener != nil && targetListenerName != irListenerName(listener) {
					continue
				}
				irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))
				if irListener != nil {
					var (
						geoIPProvider              *ir.GeoIPProvider
						geoIPErr                   error
						listenerHasNonExtAuthError = hasNonExtAuthError
						geoIPValidated             bool
					)

					for _, r := range irListener.Routes {
						// If specified the sectionName must match route rule from ir route metadata.
						if target.SectionName != nil && string(*target.SectionName) != r.Metadata.SectionName {
							continue
						}

						// A Policy targeting the most specific scope(xRoute rule) wins over a policy
						// targeting a lesser specific scope(xRoute).
						if strings.HasPrefix(r.Name, prefix) {
							// if already set - there's a specific level policy, so skip.
							if r.Security != nil {
								continue
							}

							r.Security = securityFeatures

							// Validate GeoIP if clientIPGeoLocations is used in the Authorization.
							// We have to validate GeoIP here because it reuses the listener-level ClientIPDetection configuration from CTP.
							if r.Security.Authorization.UsesClientIPGeoLocations() && !geoIPValidated {
								geoIPProvider, geoIPErr = validateAuthorizationGeoIP(authorization, gtwCtx.envoyProxy, irListener.ClientIPDetection)
								if geoIPErr != nil {
									geoIPErr = perr.WithMessage(geoIPErr, "Authorization")
									errs = errors.Join(errs, geoIPErr)
									listenerHasNonExtAuthError = true
								} else if geoIPProvider != nil {
									irListener.GeoIPProvider = geoIPProvider
								}
								// We only need to validate GeoIP once per listener
								geoIPValidated = true
							}

							if geoIPErr != nil || hasBaseErrs {
								// If there is only error for ext auth and ext auth is set to fail open, then skip the ext auth
								// and allow the request to go through.
								// Otherwise, return a 500 direct response to avoid unauthorized access.
								shouldFailOpen := extAuthErr != nil && !listenerHasNonExtAuthError && ptr.Deref(policy.Spec.ExtAuth.FailOpen, false)
								if !shouldFailOpen {
									// Return a 500 direct response to avoid unauthorized access
									r.DirectResponse = directResponse
									routesWithDirectResponse.Insert(r.Name)
								}
							}
						}
					}
				}
			}
		}
	}
	if len(routesWithDirectResponse) > 0 {
		t.Logger.Info("setting 500 direct response in routes due to errors in SecurityPolicy",
			"policy", fmt.Sprintf("%s/%s", policy.Namespace, policy.Name),
			"routes", sets.List(routesWithDirectResponse),
			"error", errs,
		)
	}
	return errs
}

func gatewaySecurityPolicyTargetListeners(
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

func gatewayDirectListeners(gtwCtx *GatewayContext) []*ListenerContext {
	listeners := make([]*ListenerContext, 0, len(gtwCtx.listeners))
	for _, listener := range gtwCtx.listeners {
		if listener.isFromListenerSet() {
			continue
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

func listenerSetSecurityPolicyTargetListeners(
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

func (t *Translator) translateSecurityPolicyForListenerSet(
	policy *egv1a1.SecurityPolicy,
	gtwCtx *GatewayContext,
	listenerSet *gwapiv1.ListenerSet,
	target policyTargetReferenceWithSectionName,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) error {
	return t.translateSecurityPolicyForListeners(
		policy,
		gtwCtx,
		resources,
		xdsIR,
		listenerSetSecurityPolicyTargetListeners(gtwCtx, listenerSet, target),
	)
}

func (t *Translator) translateSecurityPolicyForGateway(
	policy *egv1a1.SecurityPolicy,
	gtwCtx *GatewayContext,
	target policyTargetReferenceWithSectionName,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) error {
	return t.translateSecurityPolicyForListeners(
		policy,
		gtwCtx,
		resources,
		xdsIR,
		gatewaySecurityPolicyTargetListeners(gtwCtx, target),
	)
}

func (t *Translator) translateSecurityPolicyForListeners(
	policy *egv1a1.SecurityPolicy,
	gtwCtx *GatewayContext,
	resources *resource.Resources,
	xdsIR resource.XdsIRMap,
	targetListeners []*ListenerContext,
) error {
	// Build IR
	noOwners := &securityPolicyOwners{}
	var (
		cors                  *ir.CORS
		jwt                   *ir.JWT
		oidc                  *ir.OIDC
		apiKeyAuth            *ir.APIKeyAuth
		basicAuth             *ir.BasicAuth
		extAuth               *ir.ExtAuth
		authorization         *ir.Authorization
		extAuthErr, err, errs error
		hasNonExtAuthError    bool
	)

	if policy.Spec.CORS != nil {
		cors = t.buildCORS(policy.Spec.CORS)
	}

	if policy.Spec.JWT != nil {
		if jwt, err = t.buildJWT(
			policy,
			noOwners,
			resources,
			gtwCtx,
		); err != nil {
			err = perr.WithMessage(err, "JWT")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.OIDC != nil {
		if oidc, err = t.buildOIDC(
			policy,
			noOwners,
			resources,
			gtwCtx,
		); err != nil {
			err = perr.WithMessage(err, "OIDC")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.BasicAuth != nil {
		if basicAuth, err = t.buildBasicAuth(
			policy,
			noOwners,
			resources,
		); err != nil {
			err = perr.WithMessage(err, "BasicAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.APIKeyAuth != nil {
		if apiKeyAuth, err = t.buildAPIKeyAuth(
			policy,
			noOwners,
			resources,
		); err != nil {
			err = perr.WithMessage(err, "APIKeyAuth")
			errs = errors.Join(errs, err)
		}
	}

	if policy.Spec.Authorization != nil {
		if authorization, err = t.buildAuthorization(policy, noOwners); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	hasNonExtAuthError = errs != nil

	if policy.Spec.ExtAuth != nil {
		if extAuth, extAuthErr = t.buildExtAuth(
			policy,
			noOwners,
			resources,
			gtwCtx,
		); extAuthErr != nil {
			extAuthErr = perr.WithMessage(extAuthErr, "ExtAuth")
			errs = errors.Join(errs, extAuthErr)
		}
	}

	// Apply IR to all the routes within the specific Gateway that originated
	// from the gateway to which this security policy was attached.
	// If the feature is already set, then skip it, since it must have be
	// set by a policy attaching to the route
	//
	// Note: there are multiple features in a security policy, even if some of them
	// are invalid, we still want to apply the valid ones.
	irKey := t.getIRKey(gtwCtx.Gateway)
	// Should exist since we've validated this
	x := xdsIR[irKey]
	listenerNames := sets.New[string]()
	for _, listener := range targetListeners {
		listenerNames.Insert(irListenerName(listener))
	}

	// Pre-create security features and error response to avoid repeated allocations
	securityFeatures := &ir.SecurityFeatures{
		CORS:          cors,
		JWT:           jwt,
		OIDC:          oidc,
		APIKeyAuth:    apiKeyAuth,
		BasicAuth:     basicAuth,
		ExtAuth:       extAuth,
		Authorization: authorization,
	}

	routesWithDirectResponse := sets.New[string]()
	hasBaseErrs := errs != nil
	directResponse := &ir.CustomResponse{StatusCode: new(uint32(500))}
	for _, h := range x.HTTP {
		if !listenerNames.Has(h.Name) {
			continue
		}

		var (
			geoIPProvider              *ir.GeoIPProvider
			geoIPErr                   error
			listenerHasNonExtAuthError = hasNonExtAuthError
		)

		if authorization.UsesClientIPGeoLocations() {
			// We have to validate GeoIP here because it requires the listener-level ClientIPDetection configuration
			geoIPProvider, geoIPErr = validateAuthorizationGeoIP(authorization, gtwCtx.envoyProxy, h.ClientIPDetection)
			if geoIPErr != nil {
				geoIPErr = perr.WithMessage(geoIPErr, "Authorization")
				errs = errors.Join(errs, geoIPErr)
				listenerHasNonExtAuthError = true
			} else if geoIPProvider != nil {
				h.GeoIPProvider = geoIPProvider
			}
		}

		var errorResponse *ir.CustomResponse
		if geoIPErr != nil || hasBaseErrs {
			// If there is only error for ext auth and ext auth is set to fail open, then skip the ext auth
			// and allow the request to go through.
			// Otherwise, return a 500 direct response to avoid unauthorized access.
			shouldFailOpen := extAuthErr != nil && !listenerHasNonExtAuthError && ptr.Deref(policy.Spec.ExtAuth.FailOpen, false)
			if !shouldFailOpen {
				errorResponse = directResponse
			}
		}

		// A Policy targeting the specific scope(xRoute rule, xRoute, Gateway listener) wins over a policy
		// targeting a lesser specific scope(Gateway).
		for _, r := range h.Routes {
			// if already set - there's a specific level policy, so skip.
			if r.Security != nil {
				continue
			}
			r.Security = securityFeatures
			if errorResponse != nil {
				r.DirectResponse = errorResponse
				routesWithDirectResponse.Insert(r.Name)
			}
		}
	}
	if len(routesWithDirectResponse) > 0 {
		t.Logger.Info("setting 500 direct response in routes due to errors in SecurityPolicy",
			"policy", fmt.Sprintf("%s/%s", policy.Namespace, policy.Name),
			"routes", sets.List(routesWithDirectResponse),
			"error", errs,
		)
	}

	// Pre-create a TCP-only authorization object to avoid re-allocation
	var tcpAuthorization *ir.Authorization
	if authorization != nil {
		authCopy := *authorization
		tcpAuthorization = &authCopy
	}

	// Apply to TCP listeners (Authorization only).
	if tcpAuthorization != nil {
		for _, tl := range x.TCP {
			if tl == nil || len(tl.Routes) == 0 {
				continue
			}
			if !listenerNames.Has(tl.Name) {
				continue
			}
			// A Policy targeting the specific scope(xRoute rule, xRoute, Gateway listener) wins over a policy
			// targeting a lesser specific scope(Gateway).
			for _, r := range tl.Routes {
				// if already set - there's a specific level policy, so skip.
				if r.Authorization != nil {
					continue
				}
				r.Authorization = tcpAuthorization
			}
		}
	}

	return errs
}

func (t *Translator) buildCORS(cors *egv1a1.CORS) *ir.CORS {
	var allowOrigins []*ir.StringMatch

	for _, origin := range cors.AllowOrigins {
		if containsWildcard(string(origin)) {
			regexStr := wildcard2regex(string(origin))
			allowOrigins = append(allowOrigins, &ir.StringMatch{
				SafeRegex: &regexStr,
			})
		} else {
			allowOrigins = append(allowOrigins, &ir.StringMatch{
				Exact: (*string)(&origin),
			})
		}
	}

	irCORS := &ir.CORS{
		AllowOrigins:     allowOrigins,
		AllowMethods:     cors.AllowMethods,
		AllowHeaders:     cors.AllowHeaders,
		ExposeHeaders:    cors.ExposeHeaders,
		AllowCredentials: cors.AllowCredentials != nil && *cors.AllowCredentials,
	}

	if cors.MaxAge != nil {
		if d, err := time.ParseDuration(string(*cors.MaxAge)); err == nil {
			irCORS.MaxAge = ir.MetaV1DurationPtr(d)
		}
	}

	return irCORS
}

func containsWildcard(s string) bool {
	return strings.ContainsAny(s, "*")
}

func wildcard2regex(wildcard string) string {
	regexStr := strings.ReplaceAll(wildcard, ".", "\\.")
	regexStr = strings.ReplaceAll(regexStr, "*", ".*")
	return regexStr
}

func (t *Translator) buildJWT(
	policy *egv1a1.SecurityPolicy,
	owners *securityPolicyOwners,
	resources *resource.Resources,
	gtwCtx *GatewayContext,
) (*ir.JWT, error) {
	if err := validateJWTProvider(policy.Spec.JWT.Providers); err != nil {
		return nil, err
	}

	jwtOwnerPolicy := policyOwnerOr(owners.jwtProviders, policy)
	providers := make([]ir.JWTProvider, 0, len(policy.Spec.JWT.Providers))
	for i, p := range policy.Spec.JWT.Providers {
		provider := ir.JWTProvider{
			Name:           p.Name,
			Issuer:         p.Issuer,
			Audiences:      p.Audiences,
			ClaimToHeaders: p.ClaimToHeaders,
			RecomputeRoute: p.RecomputeRoute,
			ExtractFrom:    p.ExtractFrom,
		}
		if p.RemoteJWKS != nil {
			remoteJWKS, err := t.buildRemoteJWKS(jwtOwnerPolicy, p.RemoteJWKS, i, resources, gtwCtx)
			if err != nil {
				return nil, err
			}
			provider.RemoteJWKS = remoteJWKS
		} else {
			localJWKS, err := t.buildLocalJWKS(jwtOwnerPolicy, p.LocalJWKS)
			if err != nil {
				return nil, err
			}
			provider.LocalJWKS = &localJWKS
		}
		providers = append(providers, provider)
	}

	return &ir.JWT{
		AllowMissing: ptr.Deref(policy.Spec.JWT.Optional, false),
		Providers:    providers,
	}, nil
}

func validateJWTProvider(providers []egv1a1.JWTProvider) error {
	var errs []error

	var names []string
	for _, provider := range providers {
		if len(provider.Name) == 0 {
			errs = append(errs, errors.New("jwt provider cannot be an empty string"))
		}

		if len(provider.Issuer) != 0 {
			switch {
			// Issuer follows StringOrURI format based on https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.1.
			// Hence, when it contains ':', it MUST be a valid URI.
			case strings.Contains(provider.Issuer, ":"):
				if _, err := url.ParseRequestURI(provider.Issuer); err != nil {
					errs = append(errs, fmt.Errorf("invalid issuer; when issuer contains ':' character, it MUST be a valid URI"))
				}
			// Adding reserved character for '@', to represent an email address.
			// Hence, when it contains '@', it MUST be a valid Email Address.
			case strings.Contains(provider.Issuer, "@"):
				if _, err := mail.ParseAddress(provider.Issuer); err != nil {
					errs = append(errs, fmt.Errorf("invalid issuer; when issuer contains '@' character, it MUST be a valid Email Address format: %w", err))
				}
			}
		}

		if (provider.RemoteJWKS == nil && provider.LocalJWKS == nil) ||
			(provider.RemoteJWKS != nil && provider.LocalJWKS != nil) {
			errs = append(errs, fmt.Errorf(
				"either remoteJWKS or localJWKS must be specified for jwt provider: %s", provider.Name))
		}

		if provider.RemoteJWKS != nil {
			if len(provider.RemoteJWKS.URI) == 0 {
				errs = append(errs, fmt.Errorf("uri must be set for remote JWKS provider: %s", provider.Name))
			} else if _, err := url.ParseRequestURI(provider.RemoteJWKS.URI); err != nil {
				errs = append(errs, fmt.Errorf("invalid remote JWKS URI: %w", err))
			}
		}

		if provider.LocalJWKS != nil {
			localJWKS := provider.LocalJWKS
			if localJWKS.Type == nil || *localJWKS.Type == egv1a1.LocalJWKSTypeInline {
				if localJWKS.Inline == nil {
					errs = append(errs, fmt.Errorf("inline JWKS must be set for local JWKS provider: %s if type is Inline", provider.Name))
				}
			} else if localJWKS.ValueRef == nil {
				errs = append(errs, fmt.Errorf("valueRef must be set for local JWKS provider: %s if type is ValueRef", provider.Name))
			}
		}

		if len(errs) == 0 {
			if strErrs := validation.IsQualifiedName(provider.Name); len(strErrs) != 0 {
				for _, strErr := range strErrs {
					errs = append(errs, errors.New(strErr))
				}
			}
			// Ensure uniqueness among provider names.
			if names == nil {
				names = append(names, provider.Name)
			} else {
				for _, name := range names {
					if name == provider.Name {
						errs = append(errs, fmt.Errorf("provider name %s must be unique", provider.Name))
					} else {
						names = append(names, provider.Name)
					}
				}
			}
		}

		for _, claimToHeader := range provider.ClaimToHeaders {
			switch {
			case len(claimToHeader.Header) == 0:
				errs = append(errs, fmt.Errorf("header must be set for claimToHeader provider: %s", claimToHeader.Header))
			case len(claimToHeader.Claim) == 0:
				errs = append(errs, fmt.Errorf("claim must be set for claimToHeader provider: %s", claimToHeader.Claim))
			}
		}
	}

	return errors.Join(errs...)
}

func (t *Translator) buildRemoteJWKS(
	policy *egv1a1.SecurityPolicy,
	remoteJWKS *egv1a1.RemoteJWKS,
	index int,
	resources *resource.Resources,
	gtwCtx *GatewayContext,
) (*ir.RemoteJWKS, error) {
	var (
		protocol              ir.AppProtocol
		rd                    *ir.RouteDestination
		traffic               *ir.TrafficFeatures
		err                   error
		cacheDuration         *metav1.Duration
		failedRefetchDuration *metav1.Duration
	)

	u, err := url.Parse(remoteJWKS.URI)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "https" {
		protocol = ir.HTTPS
	} else {
		protocol = ir.HTTP
	}

	if len(remoteJWKS.BackendRefs) > 0 {
		if rd, err = t.translateExtServiceBackendRefs(
			policy, remoteJWKS.BackendRefs, protocol, resources, gtwCtx, "jwt", index); err != nil {
			return nil, err
		}
	}

	if remoteJWKS.BackendSettings != nil {
		if traffic, err = translateTrafficFeatures(remoteJWKS.BackendSettings); err != nil {
			return nil, err
		}
	}

	if remoteJWKS.CacheDuration != nil {
		d, err := time.ParseDuration(string(*remoteJWKS.CacheDuration))
		if err != nil {
			return nil, err
		}
		cacheDuration = ir.MetaV1DurationPtr(d)
	}

	if remoteJWKS.FailedRefetchDuration != nil {
		d, err := time.ParseDuration(string(*remoteJWKS.FailedRefetchDuration))
		if err != nil {
			return nil, err
		}
		failedRefetchDuration = ir.MetaV1DurationPtr(d)
	}

	return &ir.RemoteJWKS{
		Destination:           rd,
		Traffic:               traffic,
		URI:                   remoteJWKS.URI,
		CacheDuration:         cacheDuration,
		FailedRefetchDuration: failedRefetchDuration,
	}, nil
}

func (t *Translator) buildLocalJWKS(
	policy *egv1a1.SecurityPolicy,
	localJWKS *egv1a1.LocalJWKS,
) (string, error) {
	jwksType := egv1a1.LocalJWKSTypeInline
	if localJWKS.Type != nil {
		jwksType = *localJWKS.Type
	}

	if jwksType == egv1a1.LocalJWKSTypeValueRef {
		cm := t.GetConfigMap(policy.Namespace, string(localJWKS.ValueRef.Name))
		if cm == nil {
			return "", fmt.Errorf("local JWKS ConfigMap %s/%s not found", policy.Namespace, localJWKS.ValueRef.Name)
		}

		jwksBytes, ok := cm.Data[JWKSConfigMapKey]
		if ok {
			return jwksBytes, nil
		}
		if len(cm.Data) > 0 {
			// Fallback to the first entry in the ConfigMap data if "jwks" key is not present
			for _, v := range cm.Data {
				return v, nil
			}
		}

		return "", fmt.Errorf(
			"JWKS data not found in ConfigMap %s/%s, no %q key and no other data found",
			JWKSConfigMapKey,
			cm.Namespace, cm.Name)
	}

	return *localJWKS.Inline, nil
}

func (t *Translator) buildOIDC(
	policy *egv1a1.SecurityPolicy,
	owners *securityPolicyOwners,
	resources *resource.Resources,
	gtwCtx *GatewayContext,
) (*ir.OIDC, error) {
	var (
		oidc                   = policy.Spec.OIDC
		provider               *ir.OIDCProvider
		clientID               string
		clientSecret           *corev1.Secret
		redirectURL            = defaultRedirectURL
		redirectPath           = defaultRedirectPath
		logoutPath             = defaultLogoutPath
		forwardAccessToken     = defaultForwardAccessToken
		refreshToken           = defaultRefreshToken
		passThroughAuthHeader  = defaultPassThroughAuthHeader
		disableTokenEncryption = false
		err                    error
	)

	if provider, err = t.buildOIDCProvider(policy, owners, resources, gtwCtx); err != nil {
		return nil, err
	}

	// Client ID can be specified either as a string or as a reference to a secret.
	switch {
	case oidc.ClientID != nil:
		clientID = *oidc.ClientID
	case oidc.ClientIDRef != nil:
		ownerPolicy := policyOwnerOr(owners.oidcClientIDRef, policy)
		from := crossNamespaceFrom{
			group:     egv1a1.GroupName,
			kind:      resource.KindSecurityPolicy,
			namespace: ownerPolicy.Namespace,
		}

		var clientIDSecret *corev1.Secret
		if clientIDSecret, err = t.validateSecretRef(true, from, *oidc.ClientIDRef, resources); err != nil {
			return nil, err
		}
		clientIDBytes, ok := clientIDSecret.Data[egv1a1.OIDCClientIDKey]
		if !ok || len(clientIDBytes) == 0 {
			return nil, fmt.Errorf("client ID not found in secret %s/%s", clientIDSecret.Namespace, clientIDSecret.Name)
		}
		clientID = string(clientIDBytes)
	default:
		// This is just a sanity check - the CRD validation should have caught this.
		return nil, fmt.Errorf("client ID must be specified in OIDC policy %s/%s", policy.Namespace, policy.Name)
	}

	clientSecretOwner := policyOwnerOr(owners.oidcClientSecret, policy)
	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      resource.KindSecurityPolicy,
		namespace: clientSecretOwner.Namespace,
	}
	if clientSecret, err = t.validateSecretRef(true, from, oidc.ClientSecret, resources); err != nil {
		return nil, err
	}

	clientSecretBytes, ok := clientSecret.Data[egv1a1.OIDCClientSecretKey]
	if !ok || len(clientSecretBytes) == 0 {
		return nil, fmt.Errorf(
			"client secret not found in secret %s/%s",
			clientSecret.Namespace, clientSecret.Name)
	}

	scopes := appendOpenidScopeIfNotExist(oidc.Scopes)

	if oidc.RedirectURL != nil {
		path, err := extractRedirectPath(*oidc.RedirectURL)
		if err != nil {
			return nil, err
		}
		redirectURL = *oidc.RedirectURL
		redirectPath = path
	}
	if oidc.LogoutPath != nil {
		logoutPath = *oidc.LogoutPath
	}
	if oidc.ForwardAccessToken != nil {
		forwardAccessToken = *oidc.ForwardAccessToken
	}
	if oidc.RefreshToken != nil {
		refreshToken = *oidc.RefreshToken
	}

	if oidc.PassThroughAuthHeader != nil {
		passThroughAuthHeader = *oidc.PassThroughAuthHeader
	}
	if oidc.DisableTokenEncryption != nil {
		disableTokenEncryption = *oidc.DisableTokenEncryption
	}

	oidcOwner := policyOwnerOr(owners.oidc, policy)

	// Generate a unique cookie suffix for oauth filters.
	// This is to avoid cookie name collision when multiple security policies are applied
	// to the same route.
	suffix := utils.Digest32(string(oidcOwner.UID))

	// Get the HMAC secret.
	// HMAC secret is generated by the CertGen job and stored in a secret
	// We need to rotate the HMAC secret in the future, probably the same
	// way we rotate the certs generated by the CertGen job.
	hmacSecret := t.GetSecret(t.ControllerNamespace, oidcHMACSecretName)
	if hmacSecret == nil {
		return nil, fmt.Errorf("HMAC secret %s/%s not found", t.ControllerNamespace, oidcHMACSecretName)
	}
	hmacData, ok := hmacSecret.Data[oidcHMACSecretKey]
	if !ok || len(hmacData) == 0 {
		return nil, fmt.Errorf(
			"HMAC secret not found in secret %s/%s", t.ControllerNamespace, oidcHMACSecretName)
	}

	irOIDC := &ir.OIDC{
		Name:                   irConfigName(oidcOwner),
		Provider:               *provider,
		ClientID:               clientID,
		ClientSecret:           clientSecretBytes,
		Scopes:                 scopes,
		Resources:              oidc.Resources,
		RedirectURL:            redirectURL,
		RedirectPath:           redirectPath,
		LogoutPath:             logoutPath,
		ForwardAccessToken:     forwardAccessToken,
		RefreshToken:           refreshToken,
		CookieSuffix:           suffix,
		CookieNameOverrides:    policy.Spec.OIDC.CookieNames,
		CookieDomain:           policy.Spec.OIDC.CookieDomain,
		CookieConfig:           policy.Spec.OIDC.CookieConfig,
		HMACSecret:             hmacData,
		PassThroughAuthHeader:  passThroughAuthHeader,
		DisableTokenEncryption: disableTokenEncryption,
		DenyRedirect:           oidc.DenyRedirect,
	}

	if oidc.DefaultTokenTTL != nil {
		if d, err := time.ParseDuration(string(*oidc.DefaultTokenTTL)); err == nil {
			irOIDC.DefaultTokenTTL = ir.MetaV1DurationPtr(d)
		} else {
			return nil, fmt.Errorf("invalid defaultTokenTTL: %w", err)
		}
	}

	if oidc.DefaultRefreshTokenTTL != nil {
		if d, err := time.ParseDuration(string(*oidc.DefaultRefreshTokenTTL)); err == nil {
			irOIDC.DefaultRefreshTokenTTL = ir.MetaV1DurationPtr(d)
		} else {
			return nil, fmt.Errorf("invalid defaultRefreshTokenTTL: %w", err)
		}
	}

	if oidc.CSRFTokenTTL != nil {
		if d, err := time.ParseDuration(string(*oidc.CSRFTokenTTL)); err == nil {
			irOIDC.CSRFTokenTTL = ir.MetaV1DurationPtr(d)
		} else {
			return nil, fmt.Errorf("invalid csrfTokenTTL: %w", err)
		}
	}

	return irOIDC, nil
}

func (t *Translator) buildOIDCProvider(
	policy *egv1a1.SecurityPolicy,
	owners *securityPolicyOwners,
	resources *resource.Resources,
	gtwCtx *GatewayContext,
) (*ir.OIDCProvider, error) {
	var (
		provider              = policy.Spec.OIDC.Provider
		tokenEndpoint         string
		authorizationEndpoint string
		endSessionEndpoint    *string
		protocol              ir.AppProtocol
		rd                    *ir.RouteDestination
		traffic               *ir.TrafficFeatures
		providerTLS           *ir.TLSUpstreamConfig
		err                   error
	)

	var u *url.URL
	if provider.TokenEndpoint != nil {
		u, err = url.Parse(*provider.TokenEndpoint)
	} else {
		u, err = url.Parse(provider.Issuer)
	}

	if err != nil {
		return nil, err
	}

	if u.Scheme == "https" {
		protocol = ir.HTTPS
	} else {
		protocol = ir.HTTP
	}

	oidcProviderOwner := policyOwnerOr(owners.oidcProviderBackendRefs, policy)
	if len(provider.BackendRefs) > 0 {
		if rd, err = t.translateExtServiceBackendRefs(
			oidcProviderOwner, provider.BackendRefs, protocol, resources, gtwCtx, "oidc", 0); err != nil {
			return nil, err
		}
	}

	if rd != nil {
		for _, bc := range rd.GetBackendClusters() {
			for _, st := range bc.Settings {
				if st.TLS != nil {
					providerTLS = st.TLS
					break
				}
			}
			if providerTLS != nil {
				break
			}
		}
	}

	// Discover the token and authorization endpoints from the issuer's well-known url if not explicitly specified.
	// EG assumes that the issuer url uses the same protocol and CA as the token endpoint.
	// If we need to support different protocols or CAs, we need to add more fields to the OIDCProvider CRD.
	var (
		userProvidedAuthorizationEndpoint = ptr.Deref(provider.AuthorizationEndpoint, "")
		userProvidedTokenEndpoint         = ptr.Deref(provider.TokenEndpoint, "")
		userProvidedEndSessionEndpoint    = ptr.Deref(provider.EndSessionEndpoint, "")
	)

	// Authorization endpoint and token endpoint are required fields.
	// If either of them is not provided, we need to fetch them from the issuer's well-known url.
	if userProvidedAuthorizationEndpoint == "" || userProvidedTokenEndpoint == "" {
		// Fetch the endpoints from the issuer's well-known url.
		discoveredConfig, err := t.fetchEndpointsFromIssuer(provider.Issuer, providerTLS)
		if err != nil {
			return nil, err
		}

		// Prioritize using the explicitly provided authorization endpoints if available.
		// This allows users to add extra parameters to the authorization endpoint if needed.
		if userProvidedAuthorizationEndpoint != "" {
			authorizationEndpoint = userProvidedAuthorizationEndpoint
		} else {
			authorizationEndpoint = discoveredConfig.AuthorizationEndpoint
		}

		// Prioritize using the explicitly provided token endpoints if available.
		// This may not be necessary, but we do it for consistency with authorization endpoint.
		if userProvidedTokenEndpoint != "" {
			tokenEndpoint = userProvidedTokenEndpoint
		} else {
			tokenEndpoint = discoveredConfig.TokenEndpoint
		}

		// Prioritize using the explicitly provided end session endpoints if available.
		// This may not be necessary, but we do it for consistency with other endpoints.
		if userProvidedEndSessionEndpoint != "" {
			endSessionEndpoint = &userProvidedEndSessionEndpoint
		} else {
			endSessionEndpoint = discoveredConfig.EndSessionEndpoint
		}
	} else {
		tokenEndpoint = *provider.TokenEndpoint
		authorizationEndpoint = *provider.AuthorizationEndpoint
		endSessionEndpoint = provider.EndSessionEndpoint
	}

	if err = validateTokenEndpoint(tokenEndpoint); err != nil {
		return nil, err
	}

	if traffic, err = translateTrafficFeatures(provider.BackendSettings); err != nil {
		return nil, err
	}

	return &ir.OIDCProvider{
		Destination:           rd,
		Traffic:               traffic,
		AuthorizationEndpoint: authorizationEndpoint,
		TokenEndpoint:         tokenEndpoint,
		EndSessionEndpoint:    endSessionEndpoint,
	}, nil
}

func extractRedirectPath(redirectURL string) (string, error) {
	schemeDelimiter := strings.Index(redirectURL, "://")
	if schemeDelimiter <= 0 {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	scheme := redirectURL[:schemeDelimiter]
	if scheme != "http" && scheme != "https" && scheme != "%REQ(x-forwarded-proto)%" {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	hostDelimiter := strings.Index(redirectURL[schemeDelimiter+3:], "/")
	if hostDelimiter <= 0 {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	path := redirectURL[schemeDelimiter+3+hostDelimiter:]
	if path == "/" {
		return "", fmt.Errorf("invalid redirect URL %s", redirectURL)
	}
	return path, nil
}

// appendOpenidScopeIfNotExist appends the openid scope to the provided scopes
// if it is not already present.
// `openid` is a required scope for OIDC.
// see https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
func appendOpenidScopeIfNotExist(scopes []string) []string {
	const authScopeOpenID = "openid"

	hasOpenIDScope := false
	for _, scope := range scopes {
		if scope == authScopeOpenID {
			hasOpenIDScope = true
		}
	}
	if !hasOpenIDScope {
		scopes = append(scopes, authScopeOpenID)
	}
	return scopes
}

type OpenIDConfig struct {
	TokenEndpoint         string  `json:"token_endpoint"`
	AuthorizationEndpoint string  `json:"authorization_endpoint"`
	EndSessionEndpoint    *string `json:"end_session_endpoint,omitempty"`
}

func (o *OpenIDConfig) validate() error {
	if o.TokenEndpoint == "" {
		return errors.New("token_endpoint not found in OpenID configuration")
	}
	if o.AuthorizationEndpoint == "" {
		return errors.New("authorization_endpoint not found in OpenID configuration")
	}
	return nil
}

func (t *Translator) fetchEndpointsFromIssuer(issuerURL string, providerTLS *ir.TLSUpstreamConfig) (*OpenIDConfig, error) {
	if config, cachedErr, ok := t.oidcDiscoveryCache.Get(issuerURL); ok {
		if cachedErr != nil {
			return nil, cachedErr
		}
		return config, nil
	}

	config, err := discoverEndpointsFromIssuer(issuerURL, providerTLS)
	if err != nil {
		t.oidcDiscoveryCache.Set(issuerURL, nil, err)
		return nil, err
	}

	t.oidcDiscoveryCache.Set(issuerURL, config, nil)
	return config, nil
}

func discoverEndpointsFromIssuer(issuerURL string, providerTLS *ir.TLSUpstreamConfig) (*OpenIDConfig, error) {
	var (
		tlsConfig *tls.Config
		err       error
	)

	if providerTLS != nil {
		if tlsConfig, err = providerTLS.ToTLSConfig(); err != nil {
			return nil, err
		}
	}

	client := &http.Client{Timeout: defaultOIDCHTTPTimeout}
	if tlsConfig != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	// Parse the OpenID configuration response
	var config OpenIDConfig
	if err = backoff.Retry(func() error {
		resp, err := client.Get(fmt.Sprintf("%s/.well-known/openid-configuration", issuerURL))
		// Retry on transport errors
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		switch {
		// Retry on transient errors
		case retryable(resp.StatusCode):
			return fmt.Errorf("transient error fetching openid-configuration from issuer URL: %s, status code: %d", issuerURL, resp.StatusCode)
		// Do not retry on client errors
		case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusBadRequest:
			return &backoff.PermanentError{Err: fmt.Errorf("failed fetching openid-configuration from issuer URL: %s, status code: %d", issuerURL, resp.StatusCode)}
		case resp.StatusCode == http.StatusOK:
			// Do not retry if decoding fails
			if err = json.NewDecoder(resp.Body).Decode(&config); err != nil {
				return &backoff.PermanentError{Err: fmt.Errorf("error decoding openid-configuration response: %w", err)}
			}
		default:
			// Do not retry on other status codes
			return &backoff.PermanentError{Err: fmt.Errorf("unexpected status code %d when fetching openid-configuration from issuer URL: %s", resp.StatusCode, issuerURL)}
		}
		return nil
	}, backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(5*time.Second))); err != nil {
		return nil, err
	}

	if err = config.validate(); err != nil {
		return nil, fmt.Errorf("invalid openid-configuration from issuer URL %s: %w", issuerURL, err)
	}

	return &config, nil
}

// oidcDiscoveryCache is a cache for auto-discovered OIDC configurations from the issuer's well-known URL.
// The cache is only used within the current translation, so no need to lock it or expire entries.
type oidcDiscoveryCache struct {
	entries map[string]cachedOIDCEntry
}

type cachedOIDCEntry struct {
	config *OpenIDConfig
	err    error
}

func newOIDCDiscoveryCache() *oidcDiscoveryCache {
	return &oidcDiscoveryCache{
		entries: make(map[string]cachedOIDCEntry),
	}
}

func (c *oidcDiscoveryCache) Get(issuer string) (*OpenIDConfig, error, bool) {
	if c == nil {
		return nil, nil, false
	}

	entry, ok := c.entries[issuer]
	if !ok {
		return nil, nil, false
	}

	return entry.config, entry.err, true
}

func (c *oidcDiscoveryCache) Set(issuer string, cfg *OpenIDConfig, err error) {
	if c == nil {
		return
	}

	c.entries[issuer] = cachedOIDCEntry{
		config: cfg,
		err:    err,
	}
}

func retryable(code int) bool {
	return code >= 500 &&
		(code != http.StatusNotImplemented &&
			code != http.StatusHTTPVersionNotSupported &&
			code != http.StatusNetworkAuthenticationRequired)
}

// validateTokenEndpoint validates the token endpoint URL
func validateTokenEndpoint(tokenEndpoint string) error {
	parsedURL, err := url.Parse(tokenEndpoint)
	if err != nil {
		return fmt.Errorf("error parsing token endpoint URL: %w", err)
	}

	if _, err := netip.ParseAddr(parsedURL.Hostname()); err == nil {
		return fmt.Errorf("token endpoint URL must be a domain name: %s", tokenEndpoint)
	}

	if parsedURL.Port() != "" {
		_, err = strconv.Atoi(parsedURL.Port())
		if err != nil {
			return fmt.Errorf("error parsing token endpoint URL port: %w", err)
		}
	}
	return nil
}

func (t *Translator) buildAPIKeyAuth(
	policy *egv1a1.SecurityPolicy,
	owners *securityPolicyOwners,
	resources *resource.Resources,
) (*ir.APIKeyAuth, error) {
	ownerPolicy := policyOwnerOr(owners.apiKeyAuthCredentialRefs, policy)
	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      resource.KindSecurityPolicy,
		namespace: ownerPolicy.Namespace,
	}

	expected := len(policy.Spec.APIKeyAuth.CredentialRefs)
	apiKeyCredentials := make([]ir.APIKeyCredential, 0, expected)
	seenKeys := make(sets.Set[string])
	seenClients := make(sets.Set[string])

	for _, ref := range policy.Spec.APIKeyAuth.CredentialRefs {
		credentialsSecret, err := t.validateSecretRef(true, from, ref, resources)
		if err != nil {
			return nil, err
		}
		clientIDs := make([]string, 0, len(credentialsSecret.Data))
		for clientID := range credentialsSecret.Data {
			clientIDs = append(clientIDs, clientID)
		}
		sort.Strings(clientIDs)
		for _, clientid := range clientIDs {
			key := credentialsSecret.Data[clientid]
			if seenClients.Has(clientid) {
				continue
			}

			keyString := string(key)
			if seenKeys.Has(keyString) {
				return nil, errors.New("duplicated API key")
			}

			seenKeys.Insert(keyString)
			seenClients.Insert(clientid)
			apiKeyCredentials = append(apiKeyCredentials, ir.APIKeyCredential{
				Client: []byte(clientid),
				Key:    key,
			})
		}
	}

	extractFrom := make([]*ir.ExtractFrom, 0, len(policy.Spec.APIKeyAuth.ExtractFrom))
	for _, e := range policy.Spec.APIKeyAuth.ExtractFrom {
		extractFrom = append(extractFrom, &ir.ExtractFrom{
			Headers: e.Headers,
			Cookies: e.Cookies,
			Params:  e.Params,
		})
	}

	return &ir.APIKeyAuth{
		Credentials:           apiKeyCredentials,
		ExtractFrom:           extractFrom,
		ForwardClientIDHeader: policy.Spec.APIKeyAuth.ForwardClientIDHeader,
		Sanitize:              policy.Spec.APIKeyAuth.Sanitize,
	}, nil
}

func (t *Translator) buildBasicAuth(
	policy *egv1a1.SecurityPolicy,
	owners *securityPolicyOwners,
	resources *resource.Resources,
) (*ir.BasicAuth, error) {
	var (
		basicAuth   = policy.Spec.BasicAuth
		usersSecret *corev1.Secret
		err         error
	)

	ownerPolicy := policyOwnerOr(owners.basicAuth, policy)
	from := crossNamespaceFrom{
		group:     egv1a1.GroupName,
		kind:      resource.KindSecurityPolicy,
		namespace: ownerPolicy.Namespace,
	}
	if usersSecret, err = t.validateSecretRef(true, from, basicAuth.Users, resources); err != nil {
		return nil, err
	}

	usersSecretBytes, ok := usersSecret.Data[egv1a1.BasicAuthUsersSecretKey]
	if !ok || len(usersSecretBytes) == 0 {
		return nil, fmt.Errorf(
			"secret %s/%s must contain a non-empty \"%s\" key",
			usersSecret.Namespace, usersSecret.Name, egv1a1.BasicAuthUsersSecretKey)
	}

	// Normalize CRLF to LF so the \r is not included in the hash,
	// which would cause Envoy to reject it as an invalid SHA hash length.
	usersSecretBytes = bytes.ReplaceAll(usersSecretBytes, []byte("\r\n"), []byte("\n"))

	// Validate the htpasswd format
	if err := validateHtpasswdFormat(usersSecretBytes); err != nil {
		return nil, err
	}

	return &ir.BasicAuth{
		Name:                  irConfigName(ownerPolicy),
		Users:                 usersSecretBytes,
		ForwardUsernameHeader: basicAuth.ForwardUsernameHeader,
	}, nil
}

// validateHtpasswdFormat validates that the htpasswd data is in the correct format.
// Currently, only the SHA format is supported by Envoy.
func validateHtpasswdFormat(data []byte) error {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid htpasswd format: each line must be in the format 'username:password'")
		}

		password := parts[1]
		if !strings.HasPrefix(password, "{SHA}") {
			return fmt.Errorf("unsupported htpasswd format: please use {SHA}")
		}
		// Envoy BasicAuth only supports unsalted SHA1 {SHA}<base64> generated by htpasswd.
		shaBytes, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(password, "{SHA}"))
		if err != nil {
			return fmt.Errorf("invalid htpasswd format: {SHA} must be base64-encoded SHA1")
		}
		if len(shaBytes) != sha1.Size {
			return fmt.Errorf("invalid htpasswd format: {SHA} must be SHA1 (%d bytes)", sha1.Size)
		}
	}
	return nil
}

func (t *Translator) buildExtAuth(
	policy *egv1a1.SecurityPolicy,
	owners *securityPolicyOwners,
	resources *resource.Resources,
	gtwCtx *GatewayContext,
) (*ir.ExtAuth, error) {
	var (
		http              = policy.Spec.ExtAuth.HTTP
		grpc              = policy.Spec.ExtAuth.GRPC
		backendRefs       []egv1a1.BackendRef
		backendSettings   *egv1a1.ClusterSettings
		protocol          ir.AppProtocol
		rd                *ir.RouteDestination
		authority         string
		err               error
		traffic           *ir.TrafficFeatures
		contextExtensions []*ir.ContextExtention
	)

	backendRefsOwnerPolicy := policyOwnerOr(owners.extAuthBackendRefs, policy)

	// These are sanity checks, they should never happen because the API server
	// should have caught them
	if http == nil && grpc == nil {
		return nil, errors.New("one of grpc or http must be specified")
	} else if http != nil && grpc != nil {
		return nil, errors.New("only one of grpc or http can be specified")
	}

	switch {
	case http != nil:
		protocol = ir.HTTP
		backendSettings = http.BackendSettings
		switch {
		case len(http.BackendRefs) > 0:
			backendRefs = http.BackendRefs
		case http.BackendRef != nil:
			backendRefs = []egv1a1.BackendRef{
				{
					BackendObjectReference: *http.BackendRef,
				},
			}
		default:
			// This is a sanity check, it should never happen because the API server should have caught it
			return nil, errors.New("http backend refs must be specified")
		}
	case grpc != nil:
		protocol = ir.GRPC
		backendSettings = grpc.BackendSettings
		switch {
		case len(grpc.BackendRefs) > 0:
			backendRefs = grpc.BackendRefs
		case grpc.BackendRef != nil:
			backendRefs = []egv1a1.BackendRef{
				{
					BackendObjectReference: *grpc.BackendRef,
				},
			}
		default:
			// This is a sanity check, it should never happen because the API server should have caught it
			return nil, errors.New("grpc backend refs must be specified")
		}
	}

	if rd, err = t.translateExtServiceBackendRefs(
		backendRefsOwnerPolicy, backendRefs, protocol, resources, gtwCtx, "extauth", 0); err != nil {
		return nil, err
	}

	for _, backendRef := range backendRefs {
		// Authority is the calculated hostname that will be used as the Authority header.
		// If there are multiple backend referenced, simply use the first one - there are no good answers here.
		// When translated to XDS, the authority is used on the filter level not on the cluster level.
		// There's no way to translate to XDS and use a different authority for each backendref
		if authority == "" {
			authority = t.backendRefAuthority(&backendRef.BackendObjectReference, backendRefsOwnerPolicy)
		}
	}

	if traffic, err = translateTrafficFeatures(backendSettings); err != nil {
		return nil, err
	}

	if contextExtensions, err = t.buildContextExtensions(policy.Spec.ExtAuth.ContextExtensions, owners, policy); err != nil {
		return nil, err
	}

	extAuthOwner := policyOwnerOr(owners.extAuth, policy)
	extAuth := &ir.ExtAuth{
		Name:                 irConfigName(extAuthOwner),
		HeadersToExtAuth:     policy.Spec.ExtAuth.HeadersToExtAuth,
		ContextExtensions:    contextExtensions,
		FailOpen:             policy.Spec.ExtAuth.FailOpen,
		Traffic:              traffic,
		RecomputeRoute:       policy.Spec.ExtAuth.RecomputeRoute,
		IncludeRouteMetadata: policy.Spec.ExtAuth.IncludeRouteMetadata,
		Timeout:              parseExtAuthTimeout(policy.Spec.ExtAuth.Timeout),
		StatusOnError:        policy.Spec.ExtAuth.StatusOnError,
	}

	if http != nil {
		extAuth.HTTP = &ir.HTTPExtAuthService{
			Destination:      *rd,
			Authority:        authority,
			Path:             ptr.Deref(http.Path, ""),
			PathOverride:     ptr.Deref(http.PathOverride, ""),
			HeadersToBackend: http.HeadersToBackend,
		}
	} else {
		extAuth.GRPC = &ir.GRPCExtAuthService{
			Destination: *rd,
			Authority:   authority,
		}
	}

	if policy.Spec.ExtAuth.BodyToExtAuth != nil {
		extAuth.BodyToExtAuth = &ir.BodyToExtAuth{
			MaxRequestBytes: policy.Spec.ExtAuth.BodyToExtAuth.MaxRequestBytes,
		}
	}

	return extAuth, nil
}

// parseExtAuthTimeout parses the timeout from gwapiv1.Duration to metav1.Duration.
func parseExtAuthTimeout(timeout *gwapiv1.Duration) *metav1.Duration {
	if timeout == nil {
		return nil
	}
	d, err := time.ParseDuration(string(*timeout))
	if err != nil {
		return nil
	}
	return &metav1.Duration{
		Duration: d,
	}
}

func (t *Translator) buildContextExtensions(
	contextExtensions []*egv1a1.ContextExtension,
	owners *securityPolicyOwners,
	defaultOwner *egv1a1.SecurityPolicy,
) ([]*ir.ContextExtention, error) {
	if len(contextExtensions) == 0 {
		return nil, nil
	}

	ctxExts := make([]*ir.ContextExtention, 0, len(contextExtensions))
	for _, ext := range contextExtensions {
		var value ir.PrivateBytes
		if ext.Type == egv1a1.ContextExtensionValueTypeValueRef {
			ownerPolicy := policyOwnerOr(owners.extAuthContextExtensions[ext.Name], defaultOwner)
			var err error
			if value, err = t.getContextExtensionValueFromRef(ext.ValueRef, ownerPolicy.Namespace); err != nil {
				return nil, err
			}
		} else if ext.Value != nil {
			value = ir.PrivateBytes(*ext.Value)
		}

		ctxExts = append(ctxExts, &ir.ContextExtention{Name: ext.Name, Value: value})
	}

	return ctxExts, nil
}

// getContextExtensionValueFromRef assumes the local object reference points to
// a Kubernetes ConfigMap or Secret.
func (t *Translator) getContextExtensionValueFromRef(
	valueRef *egv1a1.LocalObjectKeyReference,
	policyNs string,
) (ir.PrivateBytes, error) {
	if valueRef == nil {
		return nil, errors.New("unexpected nil reference")
	}

	switch valueRef.Kind {
	case resource.KindConfigMap:
		cm := t.GetConfigMap(policyNs, string(valueRef.Name))
		if cm != nil {
			s, dataOk := cm.Data[valueRef.Key]
			if !dataOk {
				return nil, fmt.Errorf("can't find the key %q in the referenced configmap %q", valueRef.Key, valueRef.Name)
			}
			return ir.PrivateBytes(s), nil
		}
		return nil, fmt.Errorf("can't find the referenced configmap %q in namespace %q", valueRef.Name, policyNs)
	case resource.KindSecret:
		sec := t.GetSecret(policyNs, string(valueRef.Name))
		if sec != nil {
			b, dataOk := sec.Data[valueRef.Key]
			if !dataOk {
				return nil, fmt.Errorf("can't find the key %q in the referenced secret %q", valueRef.Key, valueRef.Name)
			}
			dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(b)))
			n, err := base64.StdEncoding.Decode(dbuf, b)
			if err != nil {
				return nil, fmt.Errorf("error decoding the base64 value of the key %q in the referenced secret %q: %w", valueRef.Key, valueRef.Name, err)
			}
			return ir.PrivateBytes(dbuf[:n]), nil
		}
		return nil, fmt.Errorf("can't find the referenced secret %q in namespace %q", valueRef.Name, policyNs)
	}
	return nil, fmt.Errorf("unexpected reference to kind %q", valueRef.Kind)
}

func (t *Translator) backendRefAuthority(
	backendRef *gwapiv1.BackendObjectReference,
	policy *egv1a1.SecurityPolicy,
) string {
	if backendRef == nil {
		return ""
	}

	backendNamespace := NamespaceDerefOr(backendRef.Namespace, policy.Namespace)
	backendKind := KindDerefOr(backendRef.Kind, resource.KindService)
	if backendKind == resource.KindBackend {
		backend := t.GetBackend(backendNamespace, string(backendRef.Name))
		if backend != nil {
			// TODO: exists multi FQDN endpoints?
			for _, ep := range backend.Spec.Endpoints {
				if ep.FQDN != nil {
					return net.JoinHostPort(ep.FQDN.Hostname, strconv.Itoa(int(ep.FQDN.Port)))
				}
			}
		}
	}

	// Port is mandatory for Kubernetes services
	if backendKind == resource.KindService || backendKind == resource.KindServiceImport {
		return net.JoinHostPort(
			fmt.Sprintf("%s.%s", backendRef.Name, backendNamespace),
			strconv.Itoa(int(*backendRef.Port)),
		)
	}

	// Fallback to the backendRef name, normally it's a unix domain socket in this case
	return fmt.Sprintf("%s.%s", backendRef.Name, backendNamespace)
}

func (t *Translator) buildAuthorization(
	policy *egv1a1.SecurityPolicy,
	owners *securityPolicyOwners,
) (*ir.Authorization, error) {
	var (
		authorization = policy.Spec.Authorization
		irAuth        = &ir.Authorization{}
		// The default action is Deny if not specified
		defaultAction = egv1a1.AuthorizationActionDeny
	)

	ownerPolicy := policyOwnerOr(owners.authorizationRules, policy)

	if authorization.DefaultAction != nil {
		defaultAction = *authorization.DefaultAction
	}
	irAuth.DefaultAction = defaultAction

	for i := range authorization.Rules {
		rule := &authorization.Rules[i]
		irPrincipal := ir.Principal{}

		if rule.Principal != nil {
			for _, cidr := range rule.Principal.ClientCIDRs {
				cidrMatch, err := parseCIDR(string(cidr))
				if err != nil {
					return nil, fmt.Errorf("unable to translate authorization rule: %w", err)
				}

				irPrincipal.ClientCIDRs = append(irPrincipal.ClientCIDRs, cidrMatch)
			}

			irPrincipal.JWT = rule.Principal.JWT
			irPrincipal.Headers = rule.Principal.Headers
			irPrincipal.ClientIPGeoLocations = rule.Principal.ClientIPGeoLocations
		}

		if err := validateAuthorizationOperation(rule.Operation); err != nil {
			return nil, fmt.Errorf("unable to translate authorization rule: %w", err)
		}

		var name string
		if rule.Name != nil && *rule.Name != "" {
			name = *rule.Name
		} else {
			name = defaultAuthorizationRuleName(ownerPolicy, i)
		}

		var celExpression *string
		if rule.CEL != nil {
			if !validCELExpression(string(*rule.CEL)) {
				return nil, fmt.Errorf("invalid CEL expression: %s", *rule.CEL)
			}
			celExpression = new(string(*rule.CEL))
		}

		irAuth.Rules = append(irAuth.Rules, &ir.AuthorizationRule{
			Name:      name,
			Action:    rule.Action,
			Operation: rule.Operation,
			Principal: irPrincipal,
			CEL:       celExpression,
		})
	}

	return irAuth, nil
}

func validateAuthorizationOperation(operation *egv1a1.Operation) error {
	if operation == nil || operation.Path == nil {
		return nil
	}

	switch ptr.Deref(operation.Path.Type, gwapiv1.PathMatchPathPrefix) {
	case gwapiv1.PathMatchPathPrefix, gwapiv1.PathMatchExact:
		return nil
	case gwapiv1.PathMatchRegularExpression:
		return regex.Validate(operation.Path.Value)
	default:
		return fmt.Errorf("invalid path type")
	}
}

func validateAuthorizationGeoIP(
	authorization *ir.Authorization,
	envoyProxy *egv1a1.EnvoyProxy,
	clientIPDetection *ir.ClientIPDetectionSettings,
) (*ir.GeoIPProvider, error) {
	if clientIPDetection == nil {
		return nil, errors.New("authorization clientIPGeoLocations requires ClientTrafficPolicy.spec.clientIPDetection to be configured")
	}

	modeCount := 0
	if clientIPDetection.XForwardedFor != nil {
		modeCount++
	}
	if clientIPDetection.CustomHeader != nil {
		modeCount++
	}
	if clientIPDetection.DirectSourceIP != nil {
		modeCount++
	}
	if modeCount != 1 {
		return nil, errors.New("authorization clientIPGeoLocations requires exactly one of ClientTrafficPolicy.spec.clientIPDetection.{xForwardedFor,customHeader,directSourceIP}")
	}

	if clientIPDetection.XForwardedFor != nil &&
		len(clientIPDetection.XForwardedFor.TrustedCIDRs) > 0 {
		return nil, errors.New("authorization clientIPGeoLocations does not support ClientIPDetection.XForwardedFor.TrustedCIDRs")
	}

	geoIPProvider, err := buildGeoIPProvider(envoyProxy)
	if err != nil {
		return nil, err
	}
	if geoIPProvider == nil || geoIPProvider.MaxMind == nil {
		return nil, errors.New("authorization clientIPGeoLocations requires EnvoyProxy.spec.geoIP.provider to be configured")
	}

	country, region, city, asn, isp, anonymous := authorization.GeoIPRequirements()
	maxMind := geoIPProvider.MaxMind

	if country && maxMind.CountryDBPath == nil && maxMind.CityDBPath == nil {
		return nil, errors.New("authorization clientIPGeoLocations.country requires EnvoyProxy.spec.geoIP.provider.maxMind.countryDbSource or cityDbSource")
	}
	if region && maxMind.CityDBPath == nil {
		return nil, errors.New("authorization clientIPGeoLocations.region requires EnvoyProxy.spec.geoIP.provider.maxMind.cityDbSource")
	}
	if city && maxMind.CityDBPath == nil {
		return nil, errors.New("authorization clientIPGeoLocations.city requires EnvoyProxy.spec.geoIP.provider.maxMind.cityDbSource")
	}
	if asn && maxMind.ASNDBPath == nil {
		return nil, errors.New("authorization clientIPGeoLocations.asn requires EnvoyProxy.spec.geoIP.provider.maxMind.asnDbSource")
	}
	if isp && maxMind.ISPDBPath == nil {
		return nil, errors.New("authorization clientIPGeoLocations.isp requires EnvoyProxy.spec.geoIP.provider.maxMind.ispDbSource")
	}
	if anonymous && maxMind.AnonymousIPDBPath == nil {
		return nil, errors.New("authorization clientIPGeoLocations.anonymous requires EnvoyProxy.spec.geoIP.provider.maxMind.anonymousIpDbSource")
	}

	return geoIPProvider, nil
}

func buildGeoIPProvider(envoyProxy *egv1a1.EnvoyProxy) (*ir.GeoIPProvider, error) {
	if envoyProxy == nil || envoyProxy.Spec.GeoIP == nil {
		return nil, nil
	}

	provider := envoyProxy.Spec.GeoIP.Provider
	switch provider.Type {
	case egv1a1.GeoIPProviderTypeMaxMind:
		if provider.MaxMind == nil {
			return nil, fmt.Errorf("geoIP provider MaxMind is missing maxMind configuration")
		}

		return &ir.GeoIPProvider{
			MaxMind: &ir.GeoIPMaxMindProvider{
				CityDBPath:        localGeoIPDBPath(provider.MaxMind.CityDBSource),
				CountryDBPath:     localGeoIPDBPath(provider.MaxMind.CountryDBSource),
				ASNDBPath:         localGeoIPDBPath(provider.MaxMind.ASNDBSource),
				ISPDBPath:         localGeoIPDBPath(provider.MaxMind.ISPDBSource),
				AnonymousIPDBPath: localGeoIPDBPath(provider.MaxMind.AnonymousIPDBSource),
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported geoIP provider type %q", provider.Type)
	}
}

func localGeoIPDBPath(source *egv1a1.GeoIPDBSource) *string {
	if source == nil {
		return nil
	}
	return &source.Local.Path
}

func defaultAuthorizationRuleName(policy *egv1a1.SecurityPolicy, index int) string {
	return fmt.Sprintf(
		"%s/authorization/rule/%s",
		irConfigName(policy),
		strconv.Itoa(index))
}

type securityPolicyOwners struct {
	basicAuth                *egv1a1.SecurityPolicy
	apiKeyAuthCredentialRefs *egv1a1.SecurityPolicy
	authorizationRules       *egv1a1.SecurityPolicy
	extAuth                  *egv1a1.SecurityPolicy
	extAuthBackendRefs       *egv1a1.SecurityPolicy
	extAuthContextExtensions map[string]*egv1a1.SecurityPolicy
	oidc                     *egv1a1.SecurityPolicy
	oidcProviderBackendRefs  *egv1a1.SecurityPolicy
	oidcClientIDRef          *egv1a1.SecurityPolicy
	oidcClientSecret         *egv1a1.SecurityPolicy
	jwtProviders             *egv1a1.SecurityPolicy
}

// policyOwnerOr returns owner if non-nil, otherwise fallback.
// Used to resolve per-field owners from securityPolicyOwners: the owner is the policy
// that contributed the field (route overrides parent), falling back to the active policy
// when no merge occurred or the field was not set by either side.
func policyOwnerOr(owner, fallback *egv1a1.SecurityPolicy) *egv1a1.SecurityPolicy {
	if owner != nil {
		return owner
	}
	return fallback
}

// mergeSecurityPolicy merges a route-level SecurityPolicy with a parent (Gateway/Listener) SecurityPolicy.
func mergeSecurityPolicy(routePolicy, parentPolicy *egv1a1.SecurityPolicy) (*egv1a1.SecurityPolicy, *securityPolicyOwners, error) {
	if routePolicy.Spec.MergeType == nil || parentPolicy == nil {
		return routePolicy, nil, nil
	}
	mergedPolicy, err := utils.Merge[*egv1a1.SecurityPolicy](parentPolicy, routePolicy, *routePolicy.Spec.MergeType)
	if err != nil {
		return nil, nil, err
	}
	return mergedPolicy, buildSecurityPolicyOwners(routePolicy, parentPolicy), nil
}

// ownerOf returns route if routeOwns(route) is true, otherwise parent.
// Use this when ownership of a merged field is determined by a single predicate.
func ownerOf(
	route, parent *egv1a1.SecurityPolicy,
	routeOwns func(*egv1a1.SecurityPolicy) bool,
) *egv1a1.SecurityPolicy {
	if routeOwns(route) {
		return route
	}
	return parent
}

// buildSecurityPolicyOwners determines, for each merged field, which policy
// (route or parent) is considered the owner. The owner is used later to resolve
// references (e.g. Secrets, BackendRefs) scoped to the owning policy's namespace,
// and to derive IR resource names tied to the owning policy.
func buildSecurityPolicyOwners(route, parent *egv1a1.SecurityPolicy) *securityPolicyOwners {
	return &securityPolicyOwners{
		basicAuth: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.BasicAuth != nil
		}),
		apiKeyAuthCredentialRefs: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.APIKeyAuth != nil && len(p.Spec.APIKeyAuth.CredentialRefs) > 0
		}),
		authorizationRules: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.Authorization != nil && len(p.Spec.Authorization.Rules) > 0
		}),
		extAuth: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.ExtAuth != nil
		}),
		extAuthBackendRefs: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			ea := p.Spec.ExtAuth
			if ea == nil {
				return false
			}
			if ea.HTTP != nil && (len(ea.HTTP.BackendRefs) > 0 || ea.HTTP.BackendRef != nil) {
				return true
			}
			if ea.GRPC != nil && (len(ea.GRPC.BackendRefs) > 0 || ea.GRPC.BackendRef != nil) {
				return true
			}
			return false
		}),
		extAuthContextExtensions: buildExtAuthContextExtensionOwners(route, parent),
		oidc: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.OIDC != nil
		}),
		oidcProviderBackendRefs: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.OIDC != nil && len(p.Spec.OIDC.Provider.BackendRefs) > 0
		}),
		oidcClientIDRef: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.OIDC != nil && p.Spec.OIDC.ClientIDRef != nil
		}),
		oidcClientSecret: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.OIDC != nil
		}),
		jwtProviders: ownerOf(route, parent, func(p *egv1a1.SecurityPolicy) bool {
			return p.Spec.JWT != nil && len(p.Spec.JWT.Providers) > 0
		}),
	}
}

// buildExtAuthContextExtensionOwners returns a per-key owner map for ExtAuth ContextExtensions.
// Parent keys are added first so that route-level extensions take precedence on conflict.
func buildExtAuthContextExtensionOwners(route, parent *egv1a1.SecurityPolicy) map[string]*egv1a1.SecurityPolicy {
	owners := make(map[string]*egv1a1.SecurityPolicy)
	if parent.Spec.ExtAuth != nil {
		for _, ext := range parent.Spec.ExtAuth.ContextExtensions {
			owners[ext.Name] = parent
		}
	}
	if route.Spec.ExtAuth != nil {
		for _, ext := range route.Spec.ExtAuth.ContextExtensions {
			owners[ext.Name] = route
		}
	}
	return owners
}
