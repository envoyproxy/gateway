// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	labelsutil "github.com/envoyproxy/gateway/internal/utils/labels"
	"github.com/envoyproxy/gateway/internal/utils/regex"
)

const (
	// Following the description in `timeout` section of https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto
	// Request timeout, which is defined as Duration, specifies the upstream timeout for the route
	// If not specified, the default is 15s
	HTTPRequestTimeout = "15s"
	// egPrefix is a prefix of annotation keys that are processed by Envoy Gateway
	egPrefix = "gateway.envoyproxy.io/"
)

var (
	_                RoutesTranslator = (*Translator)(nil)
	validServiceName                  = `(?i)\.?[a-z_][a-z_0-9]*(\.[a-z_][a-z_0-9]*)*`
	validMethodName                   = `[A-Za-z_][A-Za-z_0-9]*`
)

type RoutesTranslator interface {
	ProcessHTTPRoutes(httpRoutes []*gwapiv1.HTTPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*HTTPRouteContext
	ProcessGRPCRoutes(grpcRoutes []*gwapiv1.GRPCRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*GRPCRouteContext
	ProcessTLSRoutes(tlsRoutes []*gwapiv1.TLSRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*TLSRouteContext
	ProcessTCPRoutes(tcpRoutes []*gwapiv1.TCPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*TCPRouteContext
	ProcessUDPRoutes(udpRoutes []*gwapiv1.UDPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*UDPRouteContext
}

func (t *Translator) ProcessHTTPRoutes(httpRoutes []*gwapiv1.HTTPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*HTTPRouteContext {
	relevantHTTPRoutes := make([]*HTTPRouteContext, 0, len(httpRoutes))

	// HTTPRoutes are already sorted by the provider layer

	for _, h := range httpRoutes {
		if h == nil {
			panic("received nil httproute")
		}
		httpRoute := &HTTPRouteContext{HTTPRoute: h}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(httpRoute, gateways)
		if !relevantRoute {
			continue
		}

		relevantHTTPRoutes = append(relevantHTTPRoutes, httpRoute)

		t.processHTTPRouteParentRefs(httpRoute, resources, xdsIR)
	}

	return relevantHTTPRoutes
}

func (t *Translator) ProcessGRPCRoutes(grpcRoutes []*gwapiv1.GRPCRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*GRPCRouteContext {
	relevantGRPCRoutes := make([]*GRPCRouteContext, 0, len(grpcRoutes))

	// GRPCRoutes are already sorted by the provider layer

	for _, g := range grpcRoutes {
		if g == nil {
			panic("received nil grpcroute")
		}
		grpcRoute := &GRPCRouteContext{GRPCRoute: g}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(grpcRoute, gateways)
		if !relevantRoute {
			continue
		}

		relevantGRPCRoutes = append(relevantGRPCRoutes, grpcRoute)

		t.processGRPCRouteParentRefs(grpcRoute, resources, xdsIR)
	}

	return relevantGRPCRoutes
}

func (t *Translator) processHTTPRouteParentRefs(httpRoute *HTTPRouteContext, resources *resource.Resources, xdsIR resource.XdsIRMap) {
	for _, parentRef := range httpRoute.ParentRefs {
		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		routesWithBackends, errs, unacceptedRules := t.processHTTPRouteRules(httpRoute, parentRef, resources, xdsIR)
		if len(errs) > 0 {
			routeStatus := GetRouteStatus(httpRoute)
			// errs are already grouped by condition type in TypedErrorCollector
			for _, err := range errs {
				// According to the Gateway API spec:
				// * RouteConditionAccepted=False should be set when all rules have failed to be accepted.'
				// * When an HTTPRoute contains a combination of both valid and invalid rules, the RouteConditionAccepted
				//   should be set to True and a RouteConditionPartiallyInvalid condition should be added with status=True.
				// Ref: https://gateway-api.sigs.k8s.io/geps/gep-1364
				if err.Type() == gwapiv1.RouteConditionAccepted {
					// Set RouteConditionAccepted=False only when all rules have failed to be accepted.
					if allRulesFailedAccepted := len(unacceptedRules) == len(httpRoute.Spec.Rules); allRulesFailedAccepted {
						status.SetRouteStatusCondition(routeStatus,
							parentRef.routeParentStatusIdx,
							httpRoute.GetGeneration(),
							gwapiv1.RouteConditionAccepted,
							metav1.ConditionFalse,
							err.Reason(),
							status.Error2ConditionMsg(err),
						)
					} else {
						// Set RouteConditionPartiallyInvalid=True when some rules have failed.
						status.SetRouteStatusCondition(routeStatus,
							parentRef.routeParentStatusIdx,
							httpRoute.GetGeneration(),
							gwapiv1.RouteConditionPartiallyInvalid,
							metav1.ConditionTrue,
							err.Reason(),
							formatDroppedRuleMessage(unacceptedRules, err),
						)
						// Set RouteConditionAccepted=True when some rules have succeeded.
						status.SetRouteStatusCondition(routeStatus,
							parentRef.routeParentStatusIdx,
							httpRoute.GetGeneration(),
							gwapiv1.RouteConditionAccepted,
							metav1.ConditionTrue,
							gwapiv1.RouteReasonAccepted,
							"Route is accepted",
						)
					}
				} else {
					status.SetRouteStatusCondition(routeStatus,
						parentRef.routeParentStatusIdx,
						httpRoute.GetGeneration(),
						err.Type(),
						metav1.ConditionFalse,
						err.Reason(),
						status.Error2ConditionMsg(err),
					)
				}
			}
		}

		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(httpRoute, gwapiv1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			routeStatus := GetRouteStatus(httpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				httpRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}
		hasHostnameIntersection := t.processHTTPRouteParentRefListener(httpRoute, routesWithBackends, parentRef, xdsIR)
		if !hasHostnameIntersection {
			routeStatus := GetRouteStatus(httpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				httpRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonNoMatchingListenerHostname,
				"There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(httpRoute, gwapiv1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if parentRef.HTTPRoute != nil &&
			len(parentRef.HTTPRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			routeStatus := GetRouteStatus(httpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				httpRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonAccepted,
				"Route is accepted",
			)
		}
	}
}

func formatDroppedRuleMessage(unacceptedRules []int, err status.Error) string {
	return fmt.Sprintf("Dropped Rule(s) %v: %s", unacceptedRules, status.Error2ConditionMsg(err))
}

// routeBackendRefDestination pairs one backendRef's already-computed DestinationSetting with the
// merge-cluster key it's eligible to use, if any. backendClusterKey is nil when this backendRef
// was never merge-eligible in the first place - ds is kept either way, so its content is available
// whichever way routeDestinationForListener ultimately decides. Every route kind (HTTP, GRPC, TLS,
// TCP, UDP) builds one of these per backendRef and hands the list to routeDestinationForListener.
type routeBackendRefDestination struct {
	ds                *ir.DestinationSetting
	backendClusterKey *BackendClusterKey
}

// httpRouteWithBackendDestinations pairs one rule-match's ir.HTTPRoute with its rule's not-yet-resolved
// routeBackendDestinations, deferring the final Settings/BackendClusterRefs split - and thus whether a
// merge-eligible backend actually gets a shared cluster - to routeDestinationForListener, once a
// specific listener is known. routeBackendDestinations is nil for a rule-match that already got a
// DirectResponse/Redirect while processHTTPRouteRules/processGRPCRouteRules were building it;
// route.Destination stays nil in that case too, exactly as it does today. routeRuleMetadata is
// carried separately from route.Metadata (a distinct object, deliberately not the same pointer) so
// that a later pass tagging route.Metadata.Policies with an applied policy doesn't also show up on
// Destination.Metadata.
type httpRouteWithBackendDestinations struct {
	route                    *ir.HTTPRoute
	routeBackendDestinations []routeBackendRefDestination
	destName                 string
	routeRuleMetadata        *ir.ResourceMetadata
	routeRuleName            *gwapiv1.SectionName
	statName                 *string
}

func (t *Translator) processHTTPRouteRules(httpRoute *HTTPRouteContext, parentRef *RouteParentContext, resources *resource.Resources, xdsIR resource.XdsIRMap) ([]*httpRouteWithBackendDestinations, []status.Error, []int) {
	var (
		routesWithBackends []*httpRouteWithBackendDestinations
		errorCollector     = &status.TypedErrorCollector{}
	)
	pattern := getStatPattern(httpRoute, parentRef, t.GatewayControllerName)

	// process each HTTPRouteRule, generate a unique Xds IR HTTPRoute per match of the rule
	unacceptedRules := sets.NewInt()
	for ruleIdx, rule := range httpRoute.Spec.Rules {
		// process HTTP Route filters first, so that the filters can be applied to the IR route later
		var processFilterError error
		httpFiltersContext, errs := t.ProcessHTTPFilters(parentRef, httpRoute, rule.Filters, ruleIdx, resources, xdsIR)
		if len(errs) > 0 {
			for _, err := range errs {
				errorCollector.Add(err)
				// Gateway API conformance: When a filter's backend Service exists but has no
				// endpoints (e.g. RequestMirror), do not treat it as a fatal filter error.
				// The route should continue to work; only BackendsAvailable is set to False.
				if err.Type() != status.RouteConditionBackendsAvailable {
					processFilterError = errors.Join(processFilterError, err)
				}
				if err.Type() == gwapiv1.RouteConditionAccepted {
					unacceptedRules.Insert(ruleIdx)
				}
			}
		}

		// build the metadata for this route rule
		routeRuleMetadata := buildResourceMetadata(httpRoute, rule.Name)

		// process HTTP Route Rules
		// the HTTPRouteRule matches are ORed, a rule is matched if any one of its matches is satisfied,
		// so generate a unique Xds IR HTTPRoute per match.
		ruleRoutes, err := t.processHTTPRouteRule(httpRoute, ruleIdx, httpFiltersContext, &rule, routeRuleMetadata)
		if err != nil {
			errorCollector.Add(status.NewRouteStatusError(
				fmt.Errorf("failed to process route rule %d: %w", ruleIdx, err),
				status.ConvertToAcceptedReason(err.Reason()),
			).WithType(gwapiv1.RouteConditionAccepted))
			unacceptedRules.Insert(ruleIdx)
			continue
		}

		var (
			destName                 = irRouteDestinationName(httpRoute, ruleIdx)
			routeBackendDestinations = make([]routeBackendRefDestination, 0, len(rule.BackendRefs))
			backendWeights           = &ir.BackendWeights{}
			backendRefNames          = make([]string, len(rule.BackendRefs))
			backendCustomRefs        = make([]*ir.UnstructuredRef, 0, len(rule.BackendRefs))
			processDestinationError  error
			failedNoReadyEndpoints   bool
			hasDynamicResolver       bool
		)

		gatewayCtx := GetRouteParentContext(httpRoute, *parentRef.ParentReference, t.GatewayControllerName).GetGateway()
		btpRoutingType := t.resolveBTPRoutingType(gatewayCtx, httpRoute, parentRef, rule.Name)
		btpEndpointHostname := t.resolveBTPEndpointHostname(gatewayCtx, httpRoute, parentRef, rule.Name)

		var mergeUnsafeForRule bool
		if t.isMergeBackendsEnabledForGateway(gatewayCtx) {
			backendRefs := toBackendObjectReferences(rule.BackendRefs, func(r gwapiv1.HTTPBackendRef) gwapiv1.BackendObjectReference { return r.BackendObjectReference })
			mergeUnsafeForRule = t.mergeIncompatibleForWeightedRule(gatewayCtx, httpRoute, backendRefs, rule.SessionPersistence != nil)
		}

		// process each backendRef, and calculate the destination settings for this rule
		for i := range rule.BackendRefs {
			backendNamespace := NamespaceDerefOr(rule.BackendRefs[i].Namespace, httpRoute.GetNamespace())
			backendRefCtx := BackendRefWithFilters{
				BackendRef: &rule.BackendRefs[i].BackendRef,
				Filters:    rule.BackendRefs[i].Filters,
			}

			// backendDest.ds will never be nil here because processDestination returns an empty DestinationSetting for invalid backendRefs.
			backendDest, unstructuredRef, err := t.processBackendRef(destName, i, backendRefCtx, parentRef, httpRoute, resources, gatewayCtx, btpRoutingType, btpEndpointHostname, xdsIR, mergeUnsafeForRule)
			if err != nil {
				// Gateway API conformance: When backendRef Service exists but has no endpoints,
				// the ResolvedRefs condition should NOT be set to False.
				if err.Reason() == status.RouteReasonEndpointsNotFound {
					errorCollector.Add(status.NewRouteStatusError(
						fmt.Errorf("failed to find endpoints: %w", err),
						err.Reason(),
					).WithType(status.RouteConditionBackendsAvailable))
					failedNoReadyEndpoints = true
				} else {
					errorCollector.Add(status.NewRouteStatusError(
						fmt.Errorf("failed to process route rule %d backendRef %d: %w", ruleIdx, i, err),
						err.Reason(),
					))
					backendDest.ds.Invalid = true
					processDestinationError = err
				}
			}

			if unstructuredRef != nil {
				backendCustomRefs = append(backendCustomRefs, unstructuredRef)
			}

			// skip backendRefs with weight 0 as they do not affect the traffic distribution
			if backendDest.ds.Weight != nil && *backendDest.ds.Weight == 0 {
				continue
			}

			// check if there is a dynamic resolver in the backendRefs
			if backendDest.ds.IsDynamicResolver {
				hasDynamicResolver = true
			}

			backendRefNames[i] = fmt.Sprintf("%s/%s", backendNamespace, rule.BackendRefs[i].Name)
			backendWeights.AddWeighted(backendDest.ds, backendDest.ds.Weight)

			routeBackendDestinations = append(routeBackendDestinations, backendDest)
		}

		switch {
		// return 500 if any filter processing error occurred
		case processFilterError != nil:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info(
					"setting 500 direct response in routes due to errors in processing filters",
					"routes", sets.List(routesWithDirectResponse),
					"error", processFilterError,
				)
			}
		// return 500 if no valid destination settings exist
		// the error is already added to the error list when processing the destination
		case processDestinationError != nil && backendWeights.Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info(
					"setting 500 direct response in routes due to errors in processing destinations",
					"routes", sets.List(routesWithDirectResponse),
					"error", processDestinationError,
				)
			}
		// return 503 if no ready endpoints exist
		// the error is already added to the error list when processing the destination
		case failedNoReadyEndpoints && backendWeights.Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(503)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 503 direct response in routes due to no ready endpoints",
					"routes", sets.List(routesWithDirectResponse))
			}
		// return 500 if the weight of all the valid destination settings(endpoints list is not empty) is 0
		case backendWeights.Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 500 direct response in routes due to all valid destinations having 0 weight",
					"routes", sets.List(routesWithDirectResponse))
			}
		// Host rewrite from path (PathRegex) is rejected for dynamic resolver routes: the upstream host is
		// derived from request-controlled path text, which is not validated by the dynamic forward proxy
		// loopback protection (that guard only inspects the rewrite header or :authority). Allowing it would
		// let a crafted path resolve to a loopback address and bypass the SSRF protection.
		case hasDynamicResolver && hasPathRegexHostRewrite(ruleRoutes):
			t.rejectPathRegexHostRewriteWithDynamicResolver(ruleRoutes, ruleIdx, errorCollector)
		// A route can only have one destination if this destination is a dynamic resolver, because the behavior of
		// multiple destinations with one being a dynamic resolver just doesn't make sense.
		case hasDynamicResolver && len(rule.BackendRefs) > 1:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			errorCollector.Add(status.NewRouteStatusError(
				fmt.Errorf(
					"failed to process route rule %d: dynamic resolver is not supported for multiple backendRefs",
					ruleIdx),
				status.RouteReasonInvalidBackendRef,
			))
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 500 direct response in routes due to dynamic resolver with multiple backendRefs",
					"routes", sets.List(routesWithDirectResponse))
			}
		}

		// finalize the IR routes for this rule, deferring the Settings/BackendClusterRefs split to
		// routeDestinationForListener - it needs a specific listener to decide, which isn't known
		// until processHTTPRouteParentRefListener's fan-out.
		for _, irRoute := range ruleRoutes {
			// add custom backend refs if any
			if len(backendCustomRefs) > 0 {
				irRoute.ExtensionRefs = append(irRoute.ExtensionRefs, backendCustomRefs...)
			}

			routeWithBackends := &httpRouteWithBackendDestinations{route: irRoute}
			if irRoute.DirectResponse == nil && irRoute.Redirect == nil {
				routeWithBackends.routeBackendDestinations = routeBackendDestinations
				routeWithBackends.destName = destName
				routeWithBackends.routeRuleMetadata = routeRuleMetadata
				routeWithBackends.routeRuleName = rule.Name
				if pattern != "" {
					routeWithBackends.statName = new(buildStatName(pattern, httpRoute, rule.Name, ruleIdx, backendRefNames))
				}
			}
			routesWithBackends = append(routesWithBackends, routeWithBackends)
		}
	}

	if errorCollector.Empty() {
		return routesWithBackends, nil, nil
	}

	return routesWithBackends, errorCollector.GetAllErrors(), unacceptedRules.List()
}

// hasPathRegexHostRewrite reports whether any of the given IR routes rewrites the upstream host
// from the request path via a regex substitution (urlRewrite.hostname.type: PathRegex).
func hasPathRegexHostRewrite(routes []*ir.HTTPRoute) bool {
	for _, irRoute := range routes {
		if irRoute.URLRewrite != nil && irRoute.URLRewrite.Host != nil && irRoute.URLRewrite.Host.PathRegex != nil {
			return true
		}
	}
	return false
}

// rejectPathRegexHostRewriteWithDynamicResolver fails the rule with a 500 direct response and an
// UnsupportedValue route error. Both HTTPRoute and GRPCRoute can attach an HTTPRouteFilter that
// rewrites the host from the path, so both need this guard.
func (t *Translator) rejectPathRegexHostRewriteWithDynamicResolver(
	ruleRoutes []*ir.HTTPRoute,
	ruleIdx int,
	errorCollector *status.TypedErrorCollector,
) {
	routesWithDirectResponse := sets.New[string]()
	for _, irRoute := range ruleRoutes {
		// If the route already has a direct response or redirect configured, then it was from a filter so skip
		// the direct response from errors.
		if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
			continue
		}
		irRoute.DirectResponse = &ir.CustomResponse{
			StatusCode: new(uint32(500)),
		}
		routesWithDirectResponse.Insert(irRoute.Name)
	}
	errorCollector.Add(status.NewRouteStatusError(
		fmt.Errorf(
			"failed to process route rule %d: host rewrite from path (PathRegex) is not supported with a dynamic resolver backend",
			ruleIdx),
		gwapiv1.RouteReasonUnsupportedValue,
	))
	if len(routesWithDirectResponse) > 0 {
		t.Logger.Info("setting 500 direct response in routes due to dynamic resolver with host rewrite from path",
			"routes", sets.List(routesWithDirectResponse))
	}
}

// parentListenerSetNN returns the ListenerSet parentRef attaches through, or nil when the route
// attaches directly to a Gateway. BTP index lookups need it to tell the two sibling listener
// scopes apart.
func parentListenerSetNN(routeCtx RouteContext, parentRef *RouteParentContext) *types.NamespacedName {
	if parentRef.Kind == nil || *parentRef.Kind != resource.KindListenerSet {
		return nil
	}
	parentNamespace := routeCtx.GetNamespace()
	if parentRef.Namespace != nil {
		parentNamespace = string(*parentRef.Namespace)
	}
	return &types.NamespacedName{
		Namespace: parentNamespace,
		Name:      string(parentRef.Name),
	}
}

// resolveBTPRoutingType resolves the effective BTP RoutingType override (if any) for a route
// rule, given gatewayCtx already resolved.
func (t *Translator) resolveBTPRoutingType(
	gatewayCtx *GatewayContext,
	routeCtx RouteContext,
	parentRef *RouteParentContext,
	routeRuleName *gwapiv1.SectionName,
) *egv1a1.RoutingType {
	if gatewayCtx == nil {
		return nil
	}
	return t.BTPRoutingTypeIndex.LookupBTPRoutingType(
		routeCtx.GetRouteType(),
		types.NamespacedName{Namespace: routeCtx.GetNamespace(), Name: routeCtx.GetName()},
		types.NamespacedName{Namespace: gatewayCtx.GetNamespace(), Name: gatewayCtx.GetName()},
		parentRef.SectionName,
		parentListenerSetNN(routeCtx, parentRef),
		routeRuleName,
	)
}

// resolveBTPEndpointHostname resolves the effective BTP EndpointHostname (if any) for a route
// rule, given gatewayCtx already resolved. It's resolveBTPRoutingType's counterpart.
func (t *Translator) resolveBTPEndpointHostname(
	gatewayCtx *GatewayContext,
	routeCtx RouteContext,
	parentRef *RouteParentContext,
	routeRuleName *gwapiv1.SectionName,
) *egv1a1.BackendEndpointHostname {
	if gatewayCtx == nil {
		return nil
	}
	return t.BTPEndpointHostnameIndex.LookupBTPEndpointHostname(
		routeCtx.GetRouteType(),
		types.NamespacedName{Namespace: routeCtx.GetNamespace(), Name: routeCtx.GetName()},
		types.NamespacedName{Namespace: gatewayCtx.GetNamespace(), Name: gatewayCtx.GetName()},
		parentRef.SectionName,
		parentListenerSetNN(routeCtx, parentRef),
		routeRuleName,
	)
}

// hasClusterSettingsBelowGateway reports whether listener - and, transitively, the
// route-rule/route it's serving - has a BTP/CTP-sourced cluster-scoped setting defined below
// Gateway scope, in either of two ways:
//
//   - BTP: a route-rule/route/listener-level BackendTrafficPolicy sets a cluster-scoped field, or
//     a route-rule/route-level one targets the rule with MergeType unset - replacing the
//     Gateway's settings instead of merging with them, even when it sets none of its own.
//   - CTP: a ClientTrafficPolicy sets an HTTP1 override on listener itself (or the ListenerSet it
//     belongs to).
//
// Both make cluster deduplication unsafe for this listener, though not for the same reason: the
// BTP settings would wrongly apply to the other routes/listeners sharing the cluster, while the
// CTP ones would be lost entirely, since a merged BackendCluster carries no HTTP1 settings.
func (t *Translator) hasClusterSettingsBelowGateway(
	gatewayCtx *GatewayContext,
	routeCtx RouteContext,
	listener *ListenerContext,
	routeRuleName *gwapiv1.SectionName,
) bool {
	if gatewayCtx == nil {
		return false
	}
	gatewayNN := types.NamespacedName{Namespace: gatewayCtx.GetNamespace(), Name: gatewayCtx.GetName()}
	var listenerSetNN *types.NamespacedName
	if listener.isFromListenerSet() {
		listenerSetNN = &types.NamespacedName{
			Namespace: listener.listenerSet.Namespace,
			Name:      listener.listenerSet.Name,
		}
	}
	if t.BTPClusterSettingsIndex.HasClusterSettingsBelowGateway(
		routeCtx.GetRouteType(),
		types.NamespacedName{Namespace: routeCtx.GetNamespace(), Name: routeCtx.GetName()},
		gatewayNN,
		&listener.Name,
		listenerSetNN,
		routeRuleName,
	) {
		return true
	}
	return t.CTPClusterSettingsIndex.HasClusterSettingsBelowGateway(gatewayNN, listener)
}

// gatewayXdsIR resolves the *ir.Xds for gatewayCtx's gateway from xdsIR. Returns nil if
// gatewayCtx is nil or the gateway has no corresponding entry (e.g. a failed gateway).
func (t *Translator) gatewayXdsIR(gatewayCtx *GatewayContext, xdsIR resource.XdsIRMap) *ir.Xds {
	if gatewayCtx == nil {
		return nil
	}
	return xdsIR[t.getIRKey(gatewayCtx.Gateway)]
}

// shouldMergeBackend decides whether a specific backend participates in cluster deduplication.
// Returns false when gatewayCtx is nil (e.g. an unresolved parentRef).
func (t *Translator) shouldMergeBackend(
	gatewayCtx *GatewayContext,
	btpRoutingType *egv1a1.RoutingType,
	mergeUnsafeForRule bool,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	ds *ir.DestinationSetting,
) bool {
	if gatewayCtx == nil {
		return false
	}
	// Cheapest check first: skip all the more expensive eligibility work below when merging is
	// off for this Gateway.
	cfg := t.mergeBackendsConfigForGateway(gatewayCtx)
	if cfg == nil {
		return false
	}
	// Custom/extension-provided and dynamic-resolver backends can never safely share a cluster.
	if !t.isMergeableBackendKind(backendRef, backendNamespace) {
		return false
	}
	// A rule-wide reason (route-level cluster settings, session persistence, fallback backend,
	// ConsistentHash) already makes merging unsafe for every backendRef in this rule.
	if mergeUnsafeForRule {
		return false
	}
	// A cluster keeps only the first-registered backendRef's Filters, so any backendRef carrying
	// filters that could legitimately differ per-backendRef (header modification, URL rewrite,
	// CredentialInjection, etc.) must never share one with a differently-configured backendRef.
	if ds.Filters != nil {
		return false
	}
	// The backend's target object must match the configured Selector, if any.
	if cfg.Selector != nil && !t.mergeBackendsSelectorMatches(cfg.Selector, backendRef, backendNamespace) {
		return false
	}
	// A rule whose effective RoutingType diverges from the gateway's baseline would leak that
	// divergence into a cluster shared with rules that don't diverge.
	if t.routingTypeDivergesForRule(gatewayCtx, btpRoutingType) {
		return false
	}

	return true
}

// mergeBackendsConfigForGateway resolves the effective MergeBackendsConfig for gatewayCtx,
// preferring a Gateway-level override over the GatewayClass/default value. Returns nil when
// disabled.
func (t *Translator) mergeBackendsConfigForGateway(gatewayCtx *GatewayContext) *MergeBackendsConfig {
	if gatewayCtx != nil && gatewayCtx.envoyProxy != nil && gatewayCtx.envoyProxy.Spec.MergeBackends != nil {
		cfg := gatewayCtx.envoyProxy.Spec.MergeBackends
		return &MergeBackendsConfig{Selector: cfg.Selector}
	}
	return t.MergeBackends
}

// isMergeBackendsEnabledForGateway resolves whether MergeBackends is enabled for gatewayCtx.
func (t *Translator) isMergeBackendsEnabledForGateway(gatewayCtx *GatewayContext) bool {
	return t.mergeBackendsConfigForGateway(gatewayCtx) != nil
}

// mergeBackendsSelectorMatches reports whether backendRef's target object matches selector. An
// unresolvable target or an unparsable selector does not match.
func (t *Translator) mergeBackendsSelectorMatches(selector *metav1.LabelSelector, backendRef gwapiv1.BackendObjectReference, backendNamespace string) bool {
	backendLabels, found := t.backendLabelsFor(backendRef, backendNamespace)
	if !found {
		return false
	}
	matches, err := labelsutil.SelectorMatch(selector, backendLabels)
	if err != nil {
		t.Logger.Error(err, "invalid mergeBackends selector, excluding backend from deduplication",
			"backendRef", backendRef.Name, "namespace", backendNamespace)
		return false
	}
	return matches
}

// backendLabelsFor returns the labels of the Service, ServiceImport, or Backend object backendRef
// resolves to, and whether it was found.
func (t *Translator) backendLabelsFor(backendRef gwapiv1.BackendObjectReference, backendNamespace string) (map[string]string, bool) {
	switch KindDerefOr(backendRef.Kind, resource.KindService) {
	case resource.KindServiceImport:
		svcImport := t.GetServiceImport(backendNamespace, string(backendRef.Name))
		if svcImport == nil {
			return nil, false
		}
		return svcImport.Labels, true
	case resource.KindService:
		svc := t.GetService(backendNamespace, string(backendRef.Name))
		if svc == nil {
			return nil, false
		}
		return svc.Labels, true
	case egv1a1.KindBackend:
		backend := t.GetBackend(backendNamespace, string(backendRef.Name))
		if backend == nil {
			return nil, false
		}
		return backend.Labels, true
	default:
		return nil, false
	}
}

// anyGatewayHasMergeBackendsEnabled reports whether MergeBackends is enabled for at least one of
// gateways, so callers can skip merge-only precomputation entirely when none of them merge.
func (t *Translator) anyGatewayHasMergeBackendsEnabled(gateways []*GatewayContext) bool {
	for _, gw := range gateways {
		if t.isMergeBackendsEnabledForGateway(gw) {
			return true
		}
	}
	return false
}

// routingTypeDivergesForRule reports whether this rule's effective RoutingType differs from the
// gateway's baseline, meaning merging would leak the baseline's routing behavior into a shared
// cluster whose rule resolved it differently.
func (t *Translator) routingTypeDivergesForRule(gatewayCtx *GatewayContext, btpRoutingType *egv1a1.RoutingType) bool {
	gwNN := types.NamespacedName{Namespace: gatewayCtx.GetNamespace(), Name: gatewayCtx.GetName()}
	gatewayBaseline := t.BTPRoutingTypeIndex.LookupGatewayBTRoutingType(gwNN)
	baseline := t.IsServiceRouting(gatewayCtx.envoyProxy, gatewayBaseline)
	effective := t.IsServiceRouting(gatewayCtx.envoyProxy, btpRoutingType)
	return baseline != effective
}

// processBackendRef processes backendRefContext into a routeBackendRefDestination: its ds is
// always set, and its backendClusterKey is set too when it's eligible for cluster deduplication -
// nil otherwise. The key is not registered into any BackendCluster yet - that's deferred to
// routeDestinationForListener, once a specific listener is known to actually need it. Callers that
// must never merge (e.g. mirror backends) pass mergeUnsafeForRule: true unconditionally.
func (t *Translator) processBackendRef(
	destName string,
	backendIdx int,
	backendRefContext BackendRefContext,
	parentRef *RouteParentContext,
	route RouteContext,
	resources *resource.Resources,
	gatewayCtx *GatewayContext,
	btpRoutingType *egv1a1.RoutingType,
	btpEndpointHostname *egv1a1.BackendEndpointHostname,
	xdsIR resource.XdsIRMap,
	mergeUnsafeForRule bool,
) (backendDest routeBackendRefDestination, unstructuredRef *ir.UnstructuredRef, err status.Error) {
	ds, unstructuredRef, err := t.processDestination(irDestinationSettingName(destName, backendIdx), backendRefContext, parentRef, route, resources, gatewayCtx, btpRoutingType, btpEndpointHostname, xdsIR)

	backendRef := backendRefContext.GetBackendRef().BackendObjectReference
	backendNamespace := NamespaceDerefOr(backendRef.Namespace, route.GetNamespace())

	// An invalid backendRef never contributes traffic.
	if err != nil {
		return routeBackendRefDestination{ds: ds}, unstructuredRef, err
	}
	// An explicitly zero-weight backendRef never contributes traffic either.
	if ds.Weight != nil && *ds.Weight == 0 {
		return routeBackendRefDestination{ds: ds}, unstructuredRef, err
	}
	// Without a resolved gateway IR there's nowhere to eventually register a shared BackendCluster.
	if t.gatewayXdsIR(gatewayCtx, xdsIR) == nil {
		return routeBackendRefDestination{ds: ds}, unstructuredRef, err
	}
	// None of the above is safe to share a merged BackendCluster for.
	if !t.shouldMergeBackend(gatewayCtx, btpRoutingType, mergeUnsafeForRule, backendRef, backendNamespace, ds) {
		return routeBackendRefDestination{ds: ds}, unstructuredRef, err
	}

	key := newBackendClusterKey(backendRef, backendNamespace)
	key = t.backendClusterKeyForGateway(&key, gatewayCtx, ds.Protocol)

	return routeBackendRefDestination{ds: ds, backendClusterKey: &key}, unstructuredRef, err
}

// toBackendObjectReferences projects a rule's route-kind-specific backendRefs (HTTPBackendRef,
// GRPCBackendRef) down to their common BackendObjectReference.
func toBackendObjectReferences[T any](refs []T, get func(T) gwapiv1.BackendObjectReference) []gwapiv1.BackendObjectReference {
	out := make([]gwapiv1.BackendObjectReference, len(refs))
	for i, ref := range refs {
		out[i] = get(ref)
	}
	return out
}

// newBackendClusterKey builds the BackendClusterKey identifying a backendRef's target backend.
func newBackendClusterKey(backendRef gwapiv1.BackendObjectReference, backendNamespace string) BackendClusterKey {
	return BackendClusterKey{
		Kind:      KindDerefOr(backendRef.Kind, resource.KindService),
		Namespace: backendNamespace,
		Name:      string(backendRef.Name),
		Port:      ptr.Deref(backendRef.Port, 0),
	}
}

// backendClusterKeyForGateway extends key with the gateway and protocol scoping a merged
// BackendCluster's key needs on top of the backend's own identity.
func (t *Translator) backendClusterKeyForGateway(key *BackendClusterKey, gatewayCtx *GatewayContext, protocol ir.AppProtocol) BackendClusterKey {
	out := *key
	out.GatewayIRKey = t.getIRKey(gatewayCtx.Gateway)
	out.Protocol = protocol
	return out
}

// distinctBackendObjectReferences deduplicates refs by resolved backend identity, so multiple
// refs targeting the same backend only count once.
func distinctBackendObjectReferences(routeCtx RouteContext, refs []gwapiv1.BackendObjectReference) []gwapiv1.BackendObjectReference {
	seen := make(map[BackendClusterKey]struct{}, len(refs))
	out := make([]gwapiv1.BackendObjectReference, 0, len(refs))
	for _, ref := range refs {
		backendNamespace := NamespaceDerefOr(ref.Namespace, routeCtx.GetNamespace())
		key := newBackendClusterKey(ref, backendNamespace)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, ref)
	}
	return out
}

// isMergeableBackendKind reports whether backendRef could ever safely share a BackendCluster with
// another route. Dynamic-resolver and custom (extension-provided) backends are excluded.
func (t *Translator) isMergeableBackendKind(backendRef gwapiv1.BackendObjectReference, backendNamespace string) bool {
	kind := KindDerefOr(backendRef.Kind, resource.KindService)
	if t.isCustomBackendResource(backendRef.Group, kind) {
		return false
	}
	if kind == egv1a1.KindBackend {
		if backend := t.GetBackend(backendNamespace, string(backendRef.Name)); backend != nil &&
			backend.Spec.Type != nil && *backend.Spec.Type == egv1a1.BackendTypeDynamicResolver {
			return false
		}
	}
	return true
}

// isFallbackBackend reports whether backendRef targets a Backend CR with Spec.Fallback set,
// Envoy Gateway's fallback-backend mechanism, which relies on Envoy's priority-based locality
// failover within a single cluster.
func (t *Translator) isFallbackBackend(backendRef gwapiv1.BackendObjectReference, backendNamespace string) bool {
	if KindDerefOr(backendRef.Kind, resource.KindService) != egv1a1.KindBackend {
		return false
	}
	backend := t.GetBackend(backendNamespace, string(backendRef.Name))
	return backend != nil && ptr.Deref(backend.Spec.Fallback, false)
}

// mergeIncompatibleForWeightedRule reports whether a rule-level condition makes cluster
// deduplication unsafe for any of this rule's backendRefs: session persistence, a fallback
// backend, or ConsistentHash load balancing. For HTTP/GRPC, whose weighted-clusters route action
// can represent multiple distinct clusters in one rule.
func (t *Translator) mergeIncompatibleForWeightedRule(
	gatewayCtx *GatewayContext,
	routeCtx RouteContext,
	backendRefs []gwapiv1.BackendObjectReference,
	sessionPersistent bool,
) bool {
	// A single backendRef has no multi-backend pool for the checks below to protect —
	// nothing to fragment, so it's always mergeable at this point.
	if len(backendRefs) <= 1 {
		return false
	}
	// Delegate to the remaining multi-backendRef checks.
	return t.weightedRuleBackendsMustBeInOneCluster(routeCtx, backendRefs, sessionPersistent, gatewayCtx)
}

// mergeIncompatibleForSingleClusterRule reports whether a rule-level condition makes cluster
// deduplication unsafe for a TCP/UDP/TLS rule. Unlike HTTP/GRPC, these route types have no
// weighted-cluster mechanism at the listener layer, so a rule's backendRefs must always resolve
// to a single cluster.
func (t *Translator) mergeIncompatibleForSingleClusterRule(
	backendRefs []gwapiv1.BackendObjectReference,
) bool {
	// This route type has no weighted-cluster mechanism at the listener layer, so a rule's
	// backendRefs must always resolve to a single cluster — letting them merge independently
	// could split the rule across clusters the listener can't reference together.
	return len(backendRefs) > 1
}

// weightedRuleBackendsMustBeInOneCluster reports whether a feature on this multi-backendRef
// HTTP/GRPC rule needs all its backends kept in one Envoy cluster, so they can't be split into
// per-identity clusters.
//
// Every feature whose behavior depends on the backends sharing one cluster — hash ring, priority
// failover, session affinity — MUST be listed here.
func (t *Translator) weightedRuleBackendsMustBeInOneCluster(
	routeCtx RouteContext,
	backendRefs []gwapiv1.BackendObjectReference,
	sessionPersistent bool,
	gatewayCtx *GatewayContext,
) bool {
	// Session persistence needs all of a rule's backends in one cluster to track affinity.
	if sessionPersistent {
		return true
	}
	// A fallback backend relies on Envoy's priority-based failover within a single cluster.
	for _, ref := range backendRefs {
		if t.isFallbackBackend(ref, NamespaceDerefOr(ref.Namespace, routeCtx.GetNamespace())) {
			return true
		}
	}
	// ConsistentHash needs the full combined backend pool, not per-identity split clusters.
	if gatewayCtx != nil {
		if t.BTPLoadBalancerIndex.IsConsistentHash(utils.NamespacedName(gatewayCtx.Gateway)) {
			return true
		}
	}
	return false
}

// getOrCreateBackendCluster finds or creates the BackendCluster for key, using
// t.BackendClusterMap as a find-or-create cache.
func (t *Translator) getOrCreateBackendCluster(
	gwIR *ir.Xds,
	key *BackendClusterKey,
	ds *ir.DestinationSetting,
) *ir.BackendCluster {
	if backendCluster, ok := t.BackendClusterMap[*key]; ok {
		return backendCluster
	}
	if t.BackendClusterMap == nil {
		t.BackendClusterMap = make(map[BackendClusterKey]*ir.BackendCluster)
	}

	clusterName := irBackendClusterName(key)

	// A deduped cluster's real weight lives on its BackendClusterRef, one per referencing route,
	// not on the shared Setting - clear it here to avoid a stale, misleading value. Its Name must
	// match the BackendCluster's own, not whichever route-scoped name ds carried in with.
	copied := *ds
	copied.Weight = nil
	copied.Name = clusterName

	backendCluster := &ir.BackendCluster{
		Name:     clusterName,
		Setting:  &copied,
		Metadata: ds.Metadata,
	}

	t.BackendClusterMap[*key] = backendCluster
	if gwIR != nil {
		gwIR.BackendClusters = append(gwIR.BackendClusters, backendCluster)
	}

	return backendCluster
}

// routeDestinationForListener builds the RouteDestination for one listener a route attaches to.
// A merge-eligible routeBackendRefDestination uses its backendClusterKey's shared cluster, unless this specific
// listener has its own ClusterSettings divergence, in which case it falls back to its own inline
// ds instead. getOrCreateBackendCluster's registration is deferred to here, lazily, so a backend
// that turns out divergent on every listener it attaches to never gets a merged cluster
// registered at all.
func (t *Translator) routeDestinationForListener(
	gwIR *ir.Xds,
	gatewayCtx *GatewayContext,
	routeCtx RouteContext,
	listener *ListenerContext,
	routeRuleName *gwapiv1.SectionName,
	destName string,
	routeRuleMetadata *ir.ResourceMetadata,
	statName *string,
	routeBackendDestinations []routeBackendRefDestination,
) *ir.RouteDestination {
	hasClusterSettings := t.hasClusterSettingsBelowGateway(gatewayCtx, routeCtx, listener, routeRuleName)

	destination := &ir.RouteDestination{
		Name:     destName,
		Metadata: routeRuleMetadata,
		StatName: statName,
	}
	for _, bd := range routeBackendDestinations {
		if bd.backendClusterKey == nil || hasClusterSettings {
			destination.Settings = append(destination.Settings, bd.ds)
			continue
		}
		backendCluster := t.getOrCreateBackendCluster(gwIR, bd.backendClusterKey, bd.ds)
		destination.BackendClusterRefs = append(destination.BackendClusterRefs, &ir.BackendClusterRef{Name: backendCluster.Name, Weight: bd.ds.Weight})
	}
	return destination
}

type routeMatchCombination struct {
	gwapiv1.HTTPRouteMatch
	cookies []egv1a1.HTTPCookieMatch
}

// buildRouteMatchCombinations builds a list of route match combinations from the given rule matches and filter matches.
// The rule matches are ANDed with the filter matches. The result is a list of X*Y combinations where X is the number of
// rule matches and Y is the number of filter matches.
func buildRouteMatchCombinations(ruleMatches []gwapiv1.HTTPRouteMatch, filterMatches []egv1a1.HTTPRouteMatchFilter) []routeMatchCombination {
	if len(ruleMatches) == 0 && len(filterMatches) == 0 {
		return nil
	}

	// If there are no filter matches, return the base matches directly.
	if len(filterMatches) == 0 {
		results := make([]routeMatchCombination, len(ruleMatches))
		for i, match := range ruleMatches {
			results[i] = routeMatchCombination{HTTPRouteMatch: match}
		}
		return results
	}

	// Cross product of base matches and filter matches.
	baseMatches := ruleMatches
	if len(baseMatches) == 0 {
		baseMatches = []gwapiv1.HTTPRouteMatch{{}}
	}
	total := len(baseMatches) * len(filterMatches)
	results := make([]routeMatchCombination, total)
	idx := 0
	for _, match := range baseMatches {
		for _, filterMatch := range filterMatches {
			results[idx] = routeMatchCombination{
				HTTPRouteMatch: match,
				cookies:        filterMatch.Cookies,
			}
			idx++
		}
	}

	return results
}

// buildCookieMatches converts the cookie matches from an HTTPRouteFilter into IR
// string matches. It is shared by the HTTPRoute and GRPCRoute translation paths.
func buildCookieMatches(cookies []egv1a1.HTTPCookieMatch) ([]*ir.StringMatch, error) {
	var matches []*ir.StringMatch
	for _, cookieMatch := range cookies {
		sm := &ir.StringMatch{
			Name: cookieMatch.Name,
		}
		matchType := egv1a1.CookieMatchExact
		if cookieMatch.Type != nil {
			matchType = *cookieMatch.Type
		}
		switch matchType {
		case egv1a1.CookieMatchExact:
			sm.Exact = new(cookieMatch.Value)
		case egv1a1.CookieMatchRegularExpression:
			if err := regex.Validate(cookieMatch.Value); err != nil {
				return nil, err
			}
			sm.SafeRegex = new(cookieMatch.Value)
		default:
			return nil, fmt.Errorf("unsupported cookie match type %q", matchType)
		}
		matches = append(matches, sm)
	}
	return matches, nil
}

func processRouteTrafficFeatures(irRoute *ir.HTTPRoute, rule *gwapiv1.HTTPRouteRule) {
	processRouteTimeout(irRoute, rule)
	processRouteRetry(irRoute, rule)
}

func processRouteTimeout(irRoute *ir.HTTPRoute, rule *gwapiv1.HTTPRouteRule) {
	if rule.Timeouts != nil {
		if rule.Timeouts.Request != nil {
			d, err := time.ParseDuration(string(*rule.Timeouts.Request))
			if err != nil {
				d, _ = time.ParseDuration(HTTPRequestTimeout)
			}
			irRoute.Timeout = ir.MetaV1DurationPtr(d)
		}

		// Only set the IR Route Timeout to the backend request timeout
		// when retries are not configured. When retries are configured,
		// the backend request timeout should set for per-retry timeout.
		if rule.Timeouts.BackendRequest != nil && rule.Retry == nil {
			d, err := time.ParseDuration(string(*rule.Timeouts.BackendRequest))
			if err != nil {
				d, _ = time.ParseDuration(HTTPRequestTimeout)
			}
			irRoute.Timeout = ir.MetaV1DurationPtr(d)
		}
	}
}

func processRouteRetry(irRoute *ir.HTTPRoute, rule *gwapiv1.HTTPRouteRule) {
	if rule.Retry == nil {
		return
	}

	retry := rule.Retry
	res := &ir.Retry{}
	if retry.Attempts != nil {
		res.NumRetries = new(uint32(*retry.Attempts))
	}
	if retry.Backoff != nil {
		backoff, err := time.ParseDuration(string(*retry.Backoff))
		if err == nil {
			if res.PerRetry == nil {
				res.PerRetry = &ir.PerRetryPolicy{}
			}
			res.PerRetry.BackOff = &ir.BackOffPolicy{
				BaseInterval: ir.MetaV1DurationPtr(backoff),
			}
		}
	}
	// xref: https://gateway-api.sigs.k8s.io/geps/gep-1742/#timeout-values
	if rule.Timeouts != nil && rule.Timeouts.BackendRequest != nil {
		backendRequestTimeout, err := time.ParseDuration(string(*rule.Timeouts.BackendRequest))
		if err == nil {
			if res.PerRetry == nil {
				res.PerRetry = &ir.PerRetryPolicy{}
			}
			res.PerRetry.Timeout = ir.MetaV1DurationPtr(backendRequestTimeout)
		}
	}
	if len(retry.Codes) > 0 {
		codes := make([]ir.HTTPStatus, 0, len(retry.Codes))
		for _, code := range retry.Codes {
			codes = append(codes, ir.HTTPStatus(code))
		}
		res.RetryOn = &ir.RetryOn{
			HTTPStatusCodes: codes,
		}
	}
	irRoute.Retry = res
}

func (t *Translator) processHTTPRouteRule(
	httpRoute *HTTPRouteContext,
	ruleIdx int,
	httpFiltersContext *HTTPFiltersContext,
	rule *gwapiv1.HTTPRouteRule,
	routeRuleMetadata *ir.ResourceMetadata,
) ([]*ir.HTTPRoute, status.Error) {
	var sessionPersistence *ir.SessionPersistence
	if rule.SessionPersistence != nil {
		var sessionName string
		if rule.SessionPersistence.SessionName == nil {
			// SessionName is optional on the gateway-api, but envoy requires it
			// so we generate the one here.

			// We generate a unique session name per route.
			// `/` isn't allowed in the header key, so we just replace it with `-`.
			sessionName = strings.ReplaceAll(irRouteDestinationName(httpRoute, ruleIdx), "/", "-")
		} else {
			sessionName = *rule.SessionPersistence.SessionName
		}

		switch {
		case rule.SessionPersistence.Type == nil || // Cookie-based session persistence is default.
			*rule.SessionPersistence.Type == gwapiv1.CookieBasedSessionPersistence:
			sessionPersistence = &ir.SessionPersistence{
				Cookie: &ir.CookieBasedSessionPersistence{
					Name: sessionName,
				},
			}
			if rule.SessionPersistence.AbsoluteTimeout != nil &&
				rule.SessionPersistence.CookieConfig != nil && rule.SessionPersistence.CookieConfig.LifetimeType != nil &&
				*rule.SessionPersistence.CookieConfig.LifetimeType == gwapiv1.PermanentCookieLifetimeType {
				ttl, err := time.ParseDuration(string(*rule.SessionPersistence.AbsoluteTimeout))
				if err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				sessionPersistence.Cookie.TTL = ir.MetaV1DurationPtr(ttl)
			}
		case *rule.SessionPersistence.Type == gwapiv1.HeaderBasedSessionPersistence:
			sessionPersistence = &ir.SessionPersistence{
				Header: &ir.HeaderBasedSessionPersistence{
					Name: sessionName,
				},
			}
		default:
			// Unknown session persistence type is specified.
			return nil, status.NewRouteStatusError(
				fmt.Errorf("unknown session persistence type %s", *rule.SessionPersistence.Type),
				gwapiv1.RouteReasonUnsupportedValue,
			)
		}
	}

	filterMatches := []egv1a1.HTTPRouteMatchFilter(nil)
	if httpFiltersContext != nil {
		filterMatches = httpFiltersContext.Matches
	}
	matches := buildRouteMatchCombinations(rule.Matches, filterMatches)

	capacity := len(matches)
	if capacity == 0 {
		capacity = 1
	}
	ruleRoutes := make([]*ir.HTTPRoute, 0, capacity)
	// If no matches are specified, the implementation MUST match every HTTP request.
	if len(matches) == 0 {
		irRoute := &ir.HTTPRoute{
			Name:     irRouteName(httpRoute, ruleIdx, -1),
			Metadata: routeRuleMetadata,
		}
		processRouteTrafficFeatures(irRoute, rule)
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
		ruleRoutes = append(ruleRoutes, irRoute)
	}

	// A rule is matched if any one of its matches
	// is satisfied (i.e. a logical "OR"), so generate
	// a unique Xds IR HTTPRoute per match.
	for matchIdx, match := range matches {
		irRoute := &ir.HTTPRoute{
			Name:               irRouteName(httpRoute, ruleIdx, matchIdx),
			SessionPersistence: sessionPersistence,
			Metadata:           routeRuleMetadata,
		}
		processRouteTrafficFeatures(irRoute, rule)

		if match.Path != nil {
			switch PathMatchTypeDerefOr(match.Path.Type, gwapiv1.PathMatchPathPrefix) {
			case gwapiv1.PathMatchPathPrefix:
				irRoute.PathMatch = &ir.StringMatch{
					Prefix: match.Path.Value,
				}
			case gwapiv1.PathMatchExact:
				irRoute.PathMatch = &ir.StringMatch{
					Exact: match.Path.Value,
				}
			case gwapiv1.PathMatchRegularExpression:
				if err := regex.Validate(*match.Path.Value); err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				irRoute.PathMatch = &ir.StringMatch{
					SafeRegex: match.Path.Value,
				}
			}
		}
		for _, headerMatch := range match.Headers {
			switch HeaderMatchTypeDerefOr(headerMatch.Type, gwapiv1.HeaderMatchExact) {
			case gwapiv1.HeaderMatchExact:
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:  string(headerMatch.Name),
					Exact: new(headerMatch.Value),
				})
			case gwapiv1.HeaderMatchRegularExpression:
				if err := regex.Validate(headerMatch.Value); err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:      string(headerMatch.Name),
					SafeRegex: new(headerMatch.Value),
				})
			}
		}
		for _, queryParamMatch := range match.QueryParams {
			switch QueryParamMatchTypeDerefOr(queryParamMatch.Type, gwapiv1.QueryParamMatchExact) {
			case gwapiv1.QueryParamMatchExact:
				irRoute.QueryParamMatches = append(irRoute.QueryParamMatches, &ir.StringMatch{
					Name:  string(queryParamMatch.Name),
					Exact: new(queryParamMatch.Value),
				})
			case gwapiv1.QueryParamMatchRegularExpression:
				if err := regex.Validate(queryParamMatch.Value); err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				irRoute.QueryParamMatches = append(irRoute.QueryParamMatches, &ir.StringMatch{
					Name:      string(queryParamMatch.Name),
					SafeRegex: new(queryParamMatch.Value),
				})
			}
		}

		if match.Method != nil {
			irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
				Name:  ":method",
				Exact: new(string(*match.Method)),
			})
		}
		cookieMatches, err := buildCookieMatches(match.cookies)
		if err != nil {
			return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
		}
		irRoute.CookieMatches = append(irRoute.CookieMatches, cookieMatches...)
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
		ruleRoutes = append(ruleRoutes, irRoute)

		// When using a CORS filter with method matching that excludes OPTIONS,
		// users must explicitly specify OPTIONS method match to handle CORS preflight requests.
		// - https://github.com/kubernetes-sigs/gateway-api/issues/3857
		//
		// Envoy Gateway improves user experience by implicitly creating the envoy route for CORS preflight.
		if (httpFiltersContext != nil && httpFiltersContext.CORS != nil) &&
			(match.Method != nil && string(*match.Method) != "OPTIONS") &&
			// Browsers will not send cookies for CORS preflight requests, so there's no need to create a CORS preflight
			// route if there are cookie matches.
			len(irRoute.CookieMatches) == 0 {
			corsRoute := &ir.HTTPRoute{
				Name:              irRouteName(httpRoute, ruleIdx, matchIdx) + "/cors-preflight",
				Metadata:          routeRuleMetadata,
				PathMatch:         irRoute.PathMatch,
				QueryParamMatches: irRoute.QueryParamMatches,
				CORS:              httpFiltersContext.CORS,
			}

			// Create header matches:
			// copy original headers (excluding :method) + add CORS headers (:method=OPTIONS, origin, access-control-request-method)
			headerMatches := make([]*ir.StringMatch, 0, len(irRoute.HeaderMatches)+2)
			for _, headerMatch := range irRoute.HeaderMatches {
				// Skip the original method match for CORS preflight route to avoid conflicting method requirements.
				if headerMatch.Name == ":method" {
					continue
				}
				headerMatches = append(headerMatches, headerMatch)
			}

			corsHeaders := []*ir.StringMatch{
				{
					Name:  ":method",
					Exact: new("OPTIONS"),
				},
				{
					Name:      "origin",
					SafeRegex: new(".*"),
				},
				{
					Name:      "access-control-request-method",
					SafeRegex: new(".*"),
				},
			}
			headerMatches = append(headerMatches, corsHeaders...)

			corsRoute.HeaderMatches = headerMatches
			ruleRoutes = append(ruleRoutes, corsRoute)
		}
	}

	return ruleRoutes, nil
}

func applyHTTPFiltersContextToIRRoute(httpFiltersContext *HTTPFiltersContext, irRoute *ir.HTTPRoute) {
	// Add the redirect filter or direct response that were created earlier to all the irRoutes
	if httpFiltersContext.RedirectResponse != nil {
		irRoute.Redirect = httpFiltersContext.RedirectResponse
	}
	if httpFiltersContext.DirectResponse != nil {
		irRoute.DirectResponse = httpFiltersContext.DirectResponse
	}
	if httpFiltersContext.URLRewrite != nil {
		irRoute.URLRewrite = httpFiltersContext.URLRewrite
	}
	if httpFiltersContext.CredentialInjection != nil {
		irRoute.CredentialInjection = httpFiltersContext.CredentialInjection
	}
	if len(httpFiltersContext.AddRequestHeaders) > 0 {
		irRoute.AddRequestHeaders = httpFiltersContext.AddRequestHeaders
	}
	if len(httpFiltersContext.RemoveRequestHeaders) > 0 {
		irRoute.RemoveRequestHeaders = httpFiltersContext.RemoveRequestHeaders
	}
	if len(httpFiltersContext.AddResponseHeaders) > 0 {
		irRoute.AddResponseHeaders = httpFiltersContext.AddResponseHeaders
	}
	if len(httpFiltersContext.RemoveResponseHeaders) > 0 {
		irRoute.RemoveResponseHeaders = httpFiltersContext.RemoveResponseHeaders
	}
	if httpFiltersContext.Mirrors != nil {
		irRoute.Mirrors = httpFiltersContext.Mirrors
	}
	if httpFiltersContext.CORS != nil {
		irRoute.CORS = httpFiltersContext.CORS
	}

	if len(httpFiltersContext.ExtensionRefs) > 0 {
		irRoute.ExtensionRefs = httpFiltersContext.ExtensionRefs
	}
}

func (t *Translator) processGRPCRouteParentRefs(grpcRoute *GRPCRouteContext, resources *resource.Resources, xdsIR resource.XdsIRMap) {
	for _, parentRef := range grpcRoute.ParentRefs {

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		routesWithBackends, errs, unacceptedRules := t.processGRPCRouteRules(grpcRoute, parentRef, resources, xdsIR)
		if len(errs) > 0 {
			routeStatus := GetRouteStatus(grpcRoute)
			// errs are already grouped by condition type in TypedErrorCollector
			for _, err := range errs {
				// According to the Gateway API spec:
				// * RouteConditionAccepted=False should be set when all rules have failed to be accepted.'
				// * When an HTTPRoute contains a combination of both valid and invalid rules, the RouteConditionAccepted
				//   should be set to True and a RouteConditionPartiallyInvalid condition should be added with status=True.
				// Ref: https://gateway-api.sigs.k8s.io/geps/gep-1364
				if err.Type() == gwapiv1.RouteConditionAccepted {
					// Set RouteConditionAccepted=False only when all rules have failed.
					if allRulesFailedAccepted := len(unacceptedRules) == len(grpcRoute.Spec.Rules); allRulesFailedAccepted {
						status.SetRouteStatusCondition(routeStatus,
							parentRef.routeParentStatusIdx,
							grpcRoute.GetGeneration(),
							gwapiv1.RouteConditionAccepted,
							metav1.ConditionFalse,
							err.Reason(),
							status.Error2ConditionMsg(err),
						)
					} else {
						// Set RouteConditionPartiallyInvalid=True when some rules have failed.
						status.SetRouteStatusCondition(routeStatus,
							parentRef.routeParentStatusIdx,
							grpcRoute.GetGeneration(),
							gwapiv1.RouteConditionPartiallyInvalid,
							metav1.ConditionTrue,
							err.Reason(),
							formatDroppedRuleMessage(unacceptedRules, err),
						)
						// Set RouteConditionAccepted=True when some rules have succeeded.
						status.SetRouteStatusCondition(routeStatus,
							parentRef.routeParentStatusIdx,
							grpcRoute.GetGeneration(),
							gwapiv1.RouteConditionAccepted,
							metav1.ConditionTrue,
							gwapiv1.RouteReasonAccepted,
							"Route is accepted",
						)
					}
				} else {
					status.SetRouteStatusCondition(routeStatus,
						parentRef.routeParentStatusIdx,
						grpcRoute.GetGeneration(),
						err.Type(),
						metav1.ConditionFalse,
						err.Reason(),
						status.Error2ConditionMsg(err),
					)
				}
			}
		}

		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(grpcRoute, gwapiv1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			routeStatus := GetRouteStatus(grpcRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				grpcRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		if parentRef.HasCondition(grpcRoute, gwapiv1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}
		hasHostnameIntersection := t.processHTTPRouteParentRefListener(grpcRoute, routesWithBackends, parentRef, xdsIR)
		if !hasHostnameIntersection {
			routeStatus := GetRouteStatus(grpcRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				grpcRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonNoMatchingListenerHostname,
				"There were no hostname intersections between the GRPCRoute and this parent ref's Listener(s).",
			)
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if parentRef.GRPCRoute != nil &&
			len(parentRef.GRPCRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			routeStatus := GetRouteStatus(grpcRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				grpcRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonAccepted,
				"Route is accepted",
			)
		}

	}
}

func (t *Translator) processGRPCRouteRules(grpcRoute *GRPCRouteContext, parentRef *RouteParentContext, resources *resource.Resources, xdsIR resource.XdsIRMap) ([]*httpRouteWithBackendDestinations, []status.Error, []int) {
	var (
		routesWithBackends []*httpRouteWithBackendDestinations
		errorCollector     = &status.TypedErrorCollector{}
	)
	pattern := getStatPattern(grpcRoute, parentRef, t.GatewayControllerName)

	// compute matches, filters, backends
	unacceptedRules := sets.NewInt()
	for ruleIdx := range grpcRoute.Spec.Rules {
		rule := &grpcRoute.Spec.Rules[ruleIdx]

		// process GRPC route filters first, so that the filters can be applied to the IR route later
		var processFilterError error
		httpFiltersContext, errs := t.ProcessGRPCFilters(parentRef, grpcRoute, rule.Filters, resources, xdsIR)
		if len(errs) > 0 {
			for _, err := range errs {
				errorCollector.Add(err)
				// Gateway API conformance: When a filter's backend Service exists but has no
				// endpoints (e.g. RequestMirror), do not treat it as a fatal filter error.
				// The route should continue to work; only BackendsAvailable is set to False.
				if err.Type() != status.RouteConditionBackendsAvailable {
					processFilterError = errors.Join(processFilterError, err)
				}
				if err.Type() == gwapiv1.RouteConditionAccepted {
					unacceptedRules.Insert(ruleIdx)
				}
			}
		}

		// process GRPC Route Rules
		// a rule is matched if any one of its matches
		// is satisfied (i.e. a logical "OR"), so generate
		// a unique Xds IR HTTPRoute per match.
		ruleRoutes, err := t.processGRPCRouteRule(grpcRoute, ruleIdx, httpFiltersContext, rule)
		if err != nil {
			unacceptedRules.Insert(ruleIdx)
			errorCollector.Add(status.NewRouteStatusError(
				fmt.Errorf("failed to process route rule %d: %w", ruleIdx, err),
				status.ConvertToAcceptedReason(err.Reason()),
			).WithType(gwapiv1.RouteConditionAccepted))
			continue
		}

		var (
			destName                 = irRouteDestinationName(grpcRoute, ruleIdx)
			routeBackendDestinations = make([]routeBackendRefDestination, 0, len(rule.BackendRefs))
			backendWeights           = &ir.BackendWeights{}
			backendRefNames          = make([]string, len(rule.BackendRefs))
			processDestinationError  error
			failedNoReadyEndpoints   bool
			hasDynamicResolver       bool
			routeRuleMetadata        = buildResourceMetadata(grpcRoute, rule.Name)
		)

		gatewayCtx := GetRouteParentContext(grpcRoute, *parentRef.ParentReference, t.GatewayControllerName).GetGateway()
		btpRoutingType := t.resolveBTPRoutingType(gatewayCtx, grpcRoute, parentRef, rule.Name)
		btpEndpointHostname := t.resolveBTPEndpointHostname(gatewayCtx, grpcRoute, parentRef, rule.Name)

		var mergeIncompatible bool
		if t.isMergeBackendsEnabledForGateway(gatewayCtx) {
			backendRefs := toBackendObjectReferences(rule.BackendRefs, func(r gwapiv1.GRPCBackendRef) gwapiv1.BackendObjectReference { return r.BackendObjectReference })
			mergeIncompatible = t.mergeIncompatibleForWeightedRule(gatewayCtx, grpcRoute, backendRefs, false)
		}

		// process each backendRef, and calculate the destination settings for this rule
		for i := range rule.BackendRefs {
			backendNamespace := NamespaceDerefOr(rule.BackendRefs[i].Namespace, grpcRoute.GetNamespace())
			backendRefCtx := BackendRefWithFilters{
				BackendRef: &rule.BackendRefs[i].BackendRef,
				Filters:    rule.BackendRefs[i].Filters,
			}
			// backendDest.ds will never be nil here because processDestination returns an empty DestinationSetting for invalid backendRefs.
			backendDest, _, err := t.processBackendRef(destName, i, backendRefCtx, parentRef, grpcRoute, resources, gatewayCtx, btpRoutingType, btpEndpointHostname, xdsIR, mergeIncompatible)
			if err != nil {
				// Gateway API conformance: When backendRef Service exists but has no endpoints,
				// the ResolvedRefs condition should NOT be set to False.
				if err.Reason() == status.RouteReasonEndpointsNotFound {
					errorCollector.Add(status.NewRouteStatusError(
						fmt.Errorf("failed to find endpoints: %w", err),
						err.Reason(),
					).WithType(status.RouteConditionBackendsAvailable))
					failedNoReadyEndpoints = true
				} else {
					errorCollector.Add(status.NewRouteStatusError(
						fmt.Errorf("failed to process route rule %d backendRef %d: %w", ruleIdx, i, err),
						err.Reason(),
					))
					processDestinationError = err
				}
			}

			// skip backendRefs with weight 0 as they do not affect the traffic distribution
			if backendDest.ds.Weight != nil && *backendDest.ds.Weight == 0 {
				continue
			}

			// check if there is a dynamic resolver in the backendRefs
			if backendDest.ds.IsDynamicResolver {
				hasDynamicResolver = true
			}

			backendRefNames[i] = fmt.Sprintf("%s/%s", backendNamespace, rule.BackendRefs[i].Name)
			backendWeights.AddWeighted(backendDest.ds, backendDest.ds.Weight)

			routeBackendDestinations = append(routeBackendDestinations, backendDest)
		}

		switch {
		// return 500 if any filter processing error occurred
		case processFilterError != nil:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 500 direct response in routes due to errors in processing filters",
					"routes", sets.List(routesWithDirectResponse),
					"error", processFilterError,
				)
			}
		// return 500 if any destination setting is invalid
		// the error is already added to the error list when processing the destination
		case processDestinationError != nil && backendWeights.Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 500 direct response in routes due to errors in processing destinations",
					"routes", sets.List(routesWithDirectResponse),
					"error", processDestinationError,
				)
			}
		// return 503 if endpoints does not exist
		// the error is already added to the error list when processing the destination
		case failedNoReadyEndpoints && backendWeights.Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(503)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 503 direct response in routes due to no ready endpoints",
					"routes", sets.List(routesWithDirectResponse))
			}
		// return 500 if the weight of all the valid destination settings(endpoints list is not empty) is 0
		case backendWeights.Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: new(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Error(errors.New("all valid destinations have 0 weight"), "setting 500 direct response in routes due to all valid destinations having 0 weight",
					"routes", sets.List(routesWithDirectResponse))
			}
		// Host rewrite from path (PathRegex) is rejected for dynamic resolver routes: the upstream host is
		// derived from request-controlled path text, which is not validated by the dynamic forward proxy
		// loopback protection (that guard only inspects the rewrite header or :authority). Allowing it would
		// let a crafted path resolve to a loopback address and bypass the SSRF protection.
		case hasDynamicResolver && hasPathRegexHostRewrite(ruleRoutes):
			t.rejectPathRegexHostRewriteWithDynamicResolver(ruleRoutes, ruleIdx, errorCollector)
		}

		// finalize the IR routes for this rule, deferring the Settings/BackendClusterRefs split to
		// routeDestinationForListener - it needs a specific listener to decide, which isn't known
		// until processHTTPRouteParentRefListener's fan-out.
		for _, irRoute := range ruleRoutes {
			irRoute.IsHTTP2 = true

			routeWithBackends := &httpRouteWithBackendDestinations{route: irRoute}
			if irRoute.DirectResponse == nil && irRoute.Redirect == nil {
				routeWithBackends.routeBackendDestinations = routeBackendDestinations
				routeWithBackends.destName = destName
				routeWithBackends.routeRuleMetadata = routeRuleMetadata
				routeWithBackends.routeRuleName = rule.Name
				if pattern != "" {
					routeWithBackends.statName = new(buildStatName(pattern, grpcRoute, rule.Name, ruleIdx, backendRefNames))
				}
			}
			routesWithBackends = append(routesWithBackends, routeWithBackends)
		}
	}

	if errorCollector.Empty() {
		return routesWithBackends, nil, nil
	}

	return routesWithBackends, errorCollector.GetAllErrors(), unacceptedRules.List()
}

// grpcRouteMatchCombination is a single gRPC route match ANDed with the cookie
// matches contributed by an HTTPRouteFilter referenced from the GRPCRoute.
type grpcRouteMatchCombination struct {
	gwapiv1.GRPCRouteMatch
	cookies []egv1a1.HTTPCookieMatch
}

// buildGRPCRouteMatchCombinations builds the list of gRPC route match combinations
// from the rule matches and the filter (HTTPRouteFilter) matches. The rule matches
// are ANDed with the filter matches, producing an X*Y cross product. It mirrors
// buildRouteMatchCombinations used for HTTPRoute.
func buildGRPCRouteMatchCombinations(ruleMatches []gwapiv1.GRPCRouteMatch, filterMatches []egv1a1.HTTPRouteMatchFilter) []grpcRouteMatchCombination {
	if len(ruleMatches) == 0 && len(filterMatches) == 0 {
		return nil
	}

	// If there are no filter matches, return the base matches directly.
	if len(filterMatches) == 0 {
		results := make([]grpcRouteMatchCombination, len(ruleMatches))
		for i, match := range ruleMatches {
			results[i] = grpcRouteMatchCombination{GRPCRouteMatch: match}
		}
		return results
	}

	// Cross product of base matches and filter matches.
	baseMatches := ruleMatches
	if len(baseMatches) == 0 {
		baseMatches = []gwapiv1.GRPCRouteMatch{{}}
	}
	total := len(baseMatches) * len(filterMatches)
	results := make([]grpcRouteMatchCombination, total)
	idx := 0
	for _, match := range baseMatches {
		for _, filterMatch := range filterMatches {
			results[idx] = grpcRouteMatchCombination{
				GRPCRouteMatch: match,
				cookies:        filterMatch.Cookies,
			}
			idx++
		}
	}

	return results
}

func (t *Translator) processGRPCRouteRule(grpcRoute *GRPCRouteContext, ruleIdx int, httpFiltersContext *HTTPFiltersContext, rule *gwapiv1.GRPCRouteRule) ([]*ir.HTTPRoute, status.Error) {
	filterMatches := []egv1a1.HTTPRouteMatchFilter(nil)
	if httpFiltersContext != nil {
		filterMatches = httpFiltersContext.Matches
	}
	matches := buildGRPCRouteMatchCombinations(rule.Matches, filterMatches)

	capacity := len(matches)
	if capacity == 0 {
		capacity = 1
	}
	ruleRoutes := make([]*ir.HTTPRoute, 0, capacity)

	// If no matches are specified, the implementation MUST match every gRPC request.
	if len(matches) == 0 {
		irRoute := &ir.HTTPRoute{
			Name: irRouteName(grpcRoute, ruleIdx, -1),
		}
		irRoute.Metadata = buildResourceMetadata(grpcRoute, rule.Name)
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
		ruleRoutes = append(ruleRoutes, irRoute)
	}

	// A rule is matched if any one of its matches
	// is satisfied (i.e. a logical "OR"), so generate
	// a unique Xds IR HTTPRoute per match.
	for matchIdx, match := range matches {
		irRoute := &ir.HTTPRoute{
			Name: irRouteName(grpcRoute, ruleIdx, matchIdx),
		}
		irRoute.Metadata = buildResourceMetadata(grpcRoute, rule.Name)
		for _, headerMatch := range match.Headers {
			switch GRPCHeaderMatchTypeDerefOr(headerMatch.Type, gwapiv1.GRPCHeaderMatchExact) {
			case gwapiv1.GRPCHeaderMatchExact:
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:  string(headerMatch.Name),
					Exact: new(headerMatch.Value),
				})
			case gwapiv1.GRPCHeaderMatchRegularExpression:
				if err := regex.Validate(headerMatch.Value); err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:      string(headerMatch.Name),
					SafeRegex: new(headerMatch.Value),
				})
			}
		}

		if match.Method != nil {
			// GRPC's path is in the form of "/<service>/<method>"
			switch GRPCMethodMatchTypeDerefOr(match.Method.Type, gwapiv1.GRPCMethodMatchExact) {
			case gwapiv1.GRPCMethodMatchExact:
				t.processGRPCRouteMethodExact(match.Method, irRoute)
			case gwapiv1.GRPCMethodMatchRegularExpression:
				if match.Method.Service != nil {
					if err := regex.Validate(*match.Method.Service); err != nil {
						return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
					}
				}
				if match.Method.Method != nil {
					if err := regex.Validate(*match.Method.Method); err != nil {
						return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
					}
				}
				t.processGRPCRouteMethodRegularExpression(match.Method, irRoute)
			}
		}

		// Additional cookie matches contributed by a referenced HTTPRouteFilter.
		cookieMatches, err := buildCookieMatches(match.cookies)
		if err != nil {
			return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
		}
		irRoute.CookieMatches = append(irRoute.CookieMatches, cookieMatches...)

		ruleRoutes = append(ruleRoutes, irRoute)
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
	}
	return ruleRoutes, nil
}

func (t *Translator) processGRPCRouteMethodExact(method *gwapiv1.GRPCMethodMatch, irRoute *ir.HTTPRoute) {
	switch {
	case method.Service != nil && method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			Exact: new(fmt.Sprintf("/%s/%s", *method.Service, *method.Method)),
		}
	case method.Method != nil:
		// Use a header match since the PathMatch doesn't support Suffix matching
		irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
			Name:   ":path",
			Suffix: new(fmt.Sprintf("/%s", *method.Method)),
		})
	case method.Service != nil:
		irRoute.PathMatch = &ir.StringMatch{
			Prefix: new(fmt.Sprintf("/%s", *method.Service)),
		}
	}
}

func (t *Translator) processGRPCRouteMethodRegularExpression(method *gwapiv1.GRPCMethodMatch, irRoute *ir.HTTPRoute) {
	switch {
	case method.Service != nil && method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: new(fmt.Sprintf("/%s/%s", *method.Service, *method.Method)),
		}
	case method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: new(fmt.Sprintf("/%s/%s", validServiceName, *method.Method)),
		}
	case method.Service != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: new(fmt.Sprintf("/%s/%s", *method.Service, validMethodName)),
		}
	}
}

func (t *Translator) processHTTPRouteParentRefListener(route RouteContext, routesWithBackends []*httpRouteWithBackendDestinations, parentRef *RouteParentContext, xdsIR resource.XdsIRMap) bool {
	// need to check hostname intersection if there are listeners
	hasHostnameIntersection := len(parentRef.listeners) == 0

	for _, listener := range parentRef.listeners {
		hosts := computeHosts(GetHostnames(route), listener)
		if len(hosts) == 0 {
			continue
		}
		hasHostnameIntersection = true
		listener.IncrementAttachedRoutes()
		if !listener.IsReady() {
			continue
		}

		gwIR := xdsIR[t.getIRKey(listener.gateway.Gateway)]

		perHostRoutes := make([]*ir.HTTPRoute, 0, len(hosts)*len(routesWithBackends))
		for _, host := range hosts {
			for _, routeWithBackends := range routesWithBackends {
				// Deep copy the route first to avoid modifying the original and
				// affecting other listeners that may be attached to the same route.
				// This is important when a route has multiple parent refs (listeners)
				// with different ports, as the redirect port needs to be derived
				// independently for each listener.
				routeRoute := routeWithBackends.route.DeepCopy()
				// If the redirect port is not set, the final redirect port must be derived.
				if routeRoute.Redirect != nil && routeRoute.Redirect.Port == nil {
					redirectPort := uint32(listener.Port)
					// If redirect scheme is not-empty, the redirect post must be the
					// well-known port associated with the redirect scheme.
					if scheme := routeRoute.Redirect.Scheme; scheme != nil {
						switch strings.ToLower(*scheme) {
						case "http":
							redirectPort = 80
						case "https":
							redirectPort = 443
						}
					}
					// If the redirect scheme does not have a well-known port, or
					// if the redirect scheme is empty, the redirect port must be the Gateway Listener port.
					routeRoute.Redirect.Port = &redirectPort
				}
				// Remove dots from the hostname before appending it to the IR Route name
				// since dots are special chars used in stats tag extraction in Envoy
				underscoredHost := strings.ReplaceAll(host, ".", "_")
				routeRoute.Name = fmt.Sprintf("%s/%s", routeRoute.Name, underscoredHost)
				routeRoute.Hostname = host

				if routeWithBackends.routeBackendDestinations != nil {
					routeRoute.Destination = t.routeDestinationForListener(
						gwIR,
						listener.gateway,
						route,
						listener,
						routeWithBackends.routeRuleName,
						routeWithBackends.destName,
						routeWithBackends.routeRuleMetadata,
						routeWithBackends.statName,
						routeWithBackends.routeBackendDestinations,
					)
				}

				perHostRoutes = append(perHostRoutes, routeRoute)
			}
		}

		irListener := gwIR.GetHTTPListener(irListenerName(listener))

		if irListener != nil {
			if route.GetRouteType() == resource.KindGRPCRoute {
				if irListener.GRPC == nil {
					irListener.GRPC = &ir.GRPCSettings{}
				}

				// Backward-compatible defaulting for listeners with an attached GRPCRoute.
				// gRPC stats may already be populated from EnvoyProxy at this point, so
				// only default it when unset. gRPC-Web is also enabled by default here,
				// but any explicit ClientTrafficPolicy setting is translated later and
				// will override this value.
				irListener.GRPC.EnableGRPCWeb = new(true)
				if irListener.GRPC.EnableGRPCStats == nil {
					irListener.GRPC.EnableGRPCStats = new(true)
				}
			}

			irListener.Routes = append(irListener.Routes, perHostRoutes...)
		}
	}

	return hasHostnameIntersection
}

// routeKey returns a "kind/namespace/name" key for a route resource.
// Kind is included so HTTPRoute and GRPCRoute with the same namespace/name
// do not collide in the route lookup.
func routeKey(kind, namespace, name string) string {
	return kind + "/" + namespace + "/" + name
}

// routeDisplayNameFromKey converts a routeKey ("Kind/namespace/name") into a
// human-readable form ("Kind namespace/name") for user-facing status messages.
func routeDisplayNameFromKey(key string) string {
	kind, nsName, ok := strings.Cut(key, "/")
	if !ok {
		return key
	}
	return kind + " " + nsName
}

// overlapKey is a canonical representation of a route's match conditions.
// Routes sharing the same overlapKey within a listener match the exact same
// set of requests and are therefore considered overlapping.
type overlapKey struct {
	hostname string
	path     string
	headers  string
	query    string
	cookies  string
}

// checkRouteOverlaps detects overlapping route matches across all IR listeners
// and sets a warning Overlap condition on the affected HTTPRoutes and GRPCRoutes.
func (t *Translator) checkRouteOverlaps(httpRoutes []*HTTPRouteContext, grpcRoutes []*GRPCRouteContext, xdsIR resource.XdsIRMap) {
	// overlaps tracks per IR listener the overlapping buckets each route
	// belongs to. Key: IR listener name -> route key -> buckets (sets of route
	// keys, including the route itself). Bucket sets are shared across their
	// members rather than expanded into per-route conflict pairs, keeping
	// storage linear in the number of routes.
	type listenerOverlaps map[string][]map[string]struct{}
	overlaps := make(map[string]listenerOverlaps)

	for _, xds := range xdsIR {
		for _, httpListener := range xds.HTTP {
			// Bucket routes by their canonical overlap key. Any bucket with
			// more than one distinct route contains overlapping routes.
			buckets := make(map[overlapKey]map[string]struct{})
			for _, r := range httpListener.Routes {
				if r.Metadata == nil {
					continue
				}
				rKey := routeKey(r.Metadata.Kind, r.Metadata.Namespace, r.Metadata.Name)
				k := buildOverlapKey(r)
				if buckets[k] == nil {
					buckets[k] = make(map[string]struct{})
				}
				buckets[k][rKey] = struct{}{}
			}
			for _, routeKeys := range buckets {
				if len(routeKeys) < 2 {
					continue
				}
				if overlaps[httpListener.Name] == nil {
					overlaps[httpListener.Name] = make(listenerOverlaps)
				}
				lo := overlaps[httpListener.Name]
				for k := range routeKeys {
					lo[k] = append(lo[k], routeKeys)
				}
			}
		}
	}

	if len(overlaps) == 0 {
		return
	}

	// Build a combined lookup from "kind/namespace/name" to RouteContext and its ParentRefs.
	type routeInfo struct {
		route      RouteContext
		parentRefs map[gwapiv1.ParentReference]*RouteParentContext
	}
	routeByKey := make(map[string]*routeInfo, len(httpRoutes)+len(grpcRoutes))
	for _, hr := range httpRoutes {
		routeByKey[routeKey(string(hr.GetRouteType()), hr.GetNamespace(), hr.GetName())] = &routeInfo{route: hr, parentRefs: hr.ParentRefs}
	}
	for _, gr := range grpcRoutes {
		routeByKey[routeKey(string(gr.GetRouteType()), gr.GetNamespace(), gr.GetName())] = &routeInfo{route: gr, parentRefs: gr.ParentRefs}
	}

	// Set the Overlap warning condition only on parentRefs whose listeners
	// match an IR listener where the overlap was detected.
	for rKey, info := range routeByKey {
		routeStatus := GetRouteStatus(info.route)
		// Collect all conflicts for this route across the parentRefs that have overlaps.
		for _, parentRef := range info.parentRefs {
			var conflicts map[string]struct{}
			for _, listener := range parentRef.listeners {
				lo, ok := overlaps[irListenerName(listener)]
				if !ok {
					continue
				}
				for _, bucket := range lo[rKey] {
					if conflicts == nil {
						conflicts = make(map[string]struct{}, len(bucket))
					}
					for c := range bucket {
						if c == rKey {
							continue
						}
						conflicts[c] = struct{}{}
					}
				}
			}
			if len(conflicts) == 0 {
				continue
			}

			conflictNames := make([]string, 0, len(conflicts))
			for name := range conflicts {
				conflictNames = append(conflictNames, routeDisplayNameFromKey(name))
			}
			sort.Strings(conflictNames)

			msg := fmt.Sprintf("Overlapping match conditions with route(s): %s", strings.Join(conflictNames, ", "))

			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				info.route.GetGeneration(),
				status.RouteConditionRouteRulesOverlap,
				metav1.ConditionTrue,
				status.RouteReasonRouteRulesOverlap,
				msg,
			)
		}
	}
}

// buildOverlapKey returns a canonical key capturing a route's match conditions.
// Two routes with the same overlapKey match the exact same set of requests.
// Header names are normalized to lowercase since HTTP header names are
// case-insensitive, and slice-valued matches (headers, query params, cookies)
// are sorted so that ordering does not affect equality.
func buildOverlapKey(r *ir.HTTPRoute) overlapKey {
	k := overlapKey{
		hostname: r.Hostname,
		headers:  stringMatchSliceKey(r.HeaderMatches, true),
		query:    stringMatchSliceKey(r.QueryParamMatches, false),
		cookies:  stringMatchSliceKey(r.CookieMatches, false),
	}
	if r.Traffic.HasConnectUpgrade() {
		// A CONNECT upgrade replaces the route's path matcher with Envoy's
		// CONNECT matcher, so CONNECT routes match all CONNECT requests
		// regardless of path and only ever overlap other CONNECT routes.
		// The sentinel cannot collide with pathMatchKey output, which always
		// contains NUL separators.
		k.path = "CONNECT"
	} else {
		k.path = pathMatchKey(r.PathMatch)
	}
	return k
}

// pathMatchKey serializes route path matches the same way they are interpreted
// by the xDS translator: no path match is equivalent to prefix "/", and
// non-root prefixes have one trailing slash trimmed before translation.
func pathMatchKey(s *ir.StringMatch) string {
	if s == nil {
		return stringMatchKey(&ir.StringMatch{Prefix: new("/")}, false)
	}
	if s.Prefix == nil || *s.Prefix == "/" {
		return stringMatchKey(s, false)
	}

	normalized := s.DeepCopy()
	normalized.Prefix = new(strings.TrimSuffix(*s.Prefix, "/"))
	return stringMatchKey(normalized, false)
}

// stringMatchKey serializes a StringMatch into a canonical string.
// When lowercaseName is true, the Name field is normalized to lowercase.
func stringMatchKey(s *ir.StringMatch, lowercaseName bool) string {
	if s == nil {
		return ""
	}
	name := s.Name
	if lowercaseName {
		name = strings.ToLower(name)
	}
	var b strings.Builder
	b.WriteString(name)
	b.WriteByte('\x00')
	if s.Exact != nil {
		b.WriteByte('e')
		b.WriteString(*s.Exact)
	}
	b.WriteByte('\x00')
	if s.Prefix != nil {
		b.WriteByte('p')
		b.WriteString(*s.Prefix)
	}
	b.WriteByte('\x00')
	if s.Suffix != nil {
		b.WriteByte('s')
		b.WriteString(*s.Suffix)
	}
	b.WriteByte('\x00')
	if s.SafeRegex != nil {
		b.WriteByte('r')
		b.WriteString(*s.SafeRegex)
	}
	b.WriteByte('\x00')
	if s.Distinct {
		b.WriteByte('d')
	}
	b.WriteByte('\x00')
	if s.Invert != nil && *s.Invert {
		b.WriteByte('i')
	}
	return b.String()
}

// stringMatchSliceKey serializes a slice of StringMatch into a canonical string
// that is independent of element order.
func stringMatchSliceKey(s []*ir.StringMatch, lowercaseName bool) string {
	if len(s) == 0 {
		return ""
	}
	keys := make([]string, len(s))
	for i, m := range s {
		keys[i] = stringMatchKey(m, lowercaseName)
	}
	sort.Strings(keys)
	return strings.Join(keys, "\x01")
}

func buildResourceMetadata(obj client.Object, sectionName *gwapiv1.SectionName) *ir.ResourceMetadata {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	if kind == "" {
		// Typed objects fetched via controller-runtime clients have an empty
		// TypeMeta; fall back to a type-based lookup so Kind stays reliable.
		kind = kindForObject(obj)
	}
	metadata := &ir.ResourceMetadata{
		Kind:        kind,
		Name:        obj.GetName(),
		Namespace:   obj.GetNamespace(),
		Annotations: ir.MapToSlice(filterEGPrefix(obj.GetAnnotations())),
	}
	if sectionName != nil {
		metadata.SectionName = string(*sectionName)
	}
	return metadata
}

// kindForObject returns the Kind string for a known Gateway API or Kubernetes
// object type. Returns an empty string for unknown types.
func kindForObject(obj client.Object) string {
	// Route wrapper types (HTTPRouteContext, GRPCRouteContext, etc.) report
	// their Kind via the RouteContext interface; the switch below matches the
	// raw types only.
	if r, ok := obj.(RouteContext); ok {
		return string(r.GetRouteType())
	}
	switch obj.(type) {
	case *gwapiv1.Gateway:
		return resource.KindGateway
	case *corev1.Service:
		return resource.KindService
	case *mcsapiv1a1.ServiceImport:
		return resource.KindServiceImport
	case *egv1a1.Backend:
		return resource.KindBackend
	}
	return ""
}

func filterEGPrefix(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}

	var out map[string]string
	for k, v := range in {
		if strings.HasPrefix(k, egPrefix) {
			if out == nil {
				out = make(map[string]string, len(in))
			}
			out[strings.TrimPrefix(k, egPrefix)] = v
		}
	}
	return out
}

func (t *Translator) ProcessTLSRoutes(tlsRoutes []*gwapiv1.TLSRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*TLSRouteContext {
	relevantTLSRoutes := make([]*TLSRouteContext, 0, len(tlsRoutes))
	// TLSRoutes are already sorted by the provider layer

	for _, tls := range tlsRoutes {
		if tls == nil {
			panic("received nil tlsroute")
		}
		tlsRoute := &TLSRouteContext{TLSRoute: tls}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(tlsRoute, gateways)
		if !relevantRoute {
			continue
		}

		relevantTLSRoutes = append(relevantTLSRoutes, tlsRoute)

		t.processTLSRouteParentRefs(tlsRoute, resources, xdsIR)
	}

	return relevantTLSRoutes
}

func (t *Translator) processTLSRouteParentRefs(tlsRoute *TLSRouteContext, resources *resource.Resources, xdsIR resource.XdsIRMap) {
	for _, parentRef := range tlsRoute.ParentRefs {

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		var (
			routeBackendDestinations []routeBackendRefDestination
			resolveErrs              = &status.MultiStatusError{}
			destName                 = irRouteDestinationName(tlsRoute, -1 /*rule index*/)
			routeRuleMetadata        = buildResourceMetadata(tlsRoute, nil)
		)

		gatewayCtx := GetRouteParentContext(tlsRoute, *parentRef.ParentReference, t.GatewayControllerName).GetGateway()
		mergeBackendsEnabled := t.isMergeBackendsEnabledForGateway(gatewayCtx)

		// TLSRouteRule has no match criteria, so every rule's backends pool into one destination -
		// merge eligibility must account for the distinct backends across all rules, not just this one.
		var allRuleBackendRefs []gwapiv1.BackendObjectReference
		if mergeBackendsEnabled {
			for _, rule := range tlsRoute.Spec.Rules {
				allRuleBackendRefs = append(allRuleBackendRefs, toBackendObjectReferences(rule.BackendRefs, func(r gwapiv1.BackendRef) gwapiv1.BackendObjectReference { return r.BackendObjectReference })...)
			}
			allRuleBackendRefs = distinctBackendObjectReferences(tlsRoute, allRuleBackendRefs)
		}

		// compute backends
		for _, rule := range tlsRoute.Spec.Rules {
			btpRoutingType := t.resolveBTPRoutingType(gatewayCtx, tlsRoute, parentRef, rule.Name)
			btpEndpointHostname := t.resolveBTPEndpointHostname(gatewayCtx, tlsRoute, parentRef, rule.Name)

			var mergeIncompatible bool
			if mergeBackendsEnabled {
				mergeIncompatible = t.mergeIncompatibleForSingleClusterRule(allRuleBackendRefs)
			}

			for i := range rule.BackendRefs {
				backendRefCtx := DirectBackendRef{BackendRef: &rule.BackendRefs[i]}

				// backendDest.ds will never be nil here because processDestination returns an empty DestinationSetting for invalid backendRefs.
				backendDest, _, err := t.processBackendRef(destName, i, backendRefCtx, parentRef, tlsRoute, resources, gatewayCtx, btpRoutingType, btpEndpointHostname, xdsIR, mergeIncompatible)
				if err != nil {
					resolveErrs.Add(err)
					continue
				}

				// skip backendRefs with weight 0 as they do not affect the traffic distribution
				if backendDest.ds.Weight != nil && *backendDest.ds.Weight > 0 {
					routeBackendDestinations = append(routeBackendDestinations, backendDest)
				}
			}

			// TODO handle:
			//	- no valid backend refs
			//	- sum of weights for valid backend refs is 0
			//	- returning 500's for invalid backend refs
			//	- etc.
		}

		// A route can only have a single destination if that destination is a dynamic resolver,
		// because combining a dynamic resolver with other backends doesn't make sense. A dynamic
		// resolver is never merge-eligible, so it can only ever appear with a nil backendClusterKey - but
		// the count check must still cover every routeBackendRefDestination regardless of backendClusterKey.
		hasDynamicResolver := false
		for _, bd := range routeBackendDestinations {
			if bd.ds.IsDynamicResolver {
				hasDynamicResolver = true
				break
			}
		}
		if hasDynamicResolver && len(routeBackendDestinations) > 1 {
			resolveErrs.Add(status.NewRouteStatusError(
				errors.New("dynamic resolver is not supported for multiple backendRefs"),
				status.RouteReasonInvalidBackendRef,
			))
			// Drop the destinations so neither a dynamic forward proxy cluster nor a regular
			// cluster is produced from an invalid combination of backends.
			routeBackendDestinations = nil
		}

		routeStatus := GetRouteStatus(tlsRoute)
		if !resolveErrs.Empty() {
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tlsRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				resolveErrs.Reason(),
				resolveErrs.Error())
		} else {
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tlsRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		// need to check hostname intersection if there are listeners
		hasHostnameIntersection := len(parentRef.listeners) == 0
		for _, listener := range parentRef.listeners {
			hosts := computeHosts(GetHostnames(tlsRoute), listener)
			if len(hosts) == 0 {
				continue
			}
			hasHostnameIntersection = true
			listener.IncrementAttachedRoutes()
			if !listener.IsReady() {
				continue
			}

			irKey := t.getIRKey(listener.gateway.Gateway)
			gwXdsIR := xdsIR[irKey]
			irListener := gwXdsIR.GetTCPListener(irListenerName(listener))
			if irListener != nil {
				var tlsConfig *ir.TLS
				if irListener.TLS != nil {
					// Listener is in terminate mode. A dynamic resolver backend forwards the connection
					// based on the SNI and requires TLS passthrough, so it cannot be used with a listener
					// that terminates TLS (Envoy would forward the decrypted stream instead).
					if hasDynamicResolver {
						routeStatus := GetRouteStatus(tlsRoute)
						status.SetRouteStatusCondition(routeStatus,
							parentRef.routeParentStatusIdx,
							tlsRoute.GetGeneration(),
							gwapiv1.RouteConditionResolvedRefs,
							metav1.ConditionFalse,
							gwapiv1.RouteReasonUnsupportedValue,
							"Dynamic resolver backend is only supported with TLS passthrough listeners",
						)
						continue
					}
					tlsConfig = &ir.TLS{
						Terminate: irListener.TLS,
					}
					// If hostnames specified, add SNI config for routing
					if len(hosts) > 0 {
						tlsConfig.TLSInspectorConfig = &ir.TLSInspectorConfig{
							SNIs: hosts,
						}
					}
				} else {
					// Passthrough mode - only SNI inspection
					tlsConfig = &ir.TLS{
						TLSInspectorConfig: &ir.TLSInspectorConfig{
							SNIs: hosts,
						},
					}
				}

				irRoute := &ir.TCPRoute{
					Name: irTCPRouteName(tlsRoute),
					TLS:  tlsConfig,
					// routeRuleName is always nil for TLS: its route metadata never carries a
					// rule-scoped section name, so a rule-scoped BTP/CTP setting can never apply
					// to it - there's no rule-scope divergence to protect against. Route- and
					// listener-scope divergence still resolve correctly with nil.
					Destination: t.routeDestinationForListener(
						gwXdsIR,
						gatewayCtx,
						tlsRoute,
						listener,
						nil,
						destName,
						routeRuleMetadata,
						nil,
						routeBackendDestinations,
					),
					Metadata: routeRuleMetadata,
				}
				irListener.Routes = append(irListener.Routes, irRoute)
			}
		}

		if !hasHostnameIntersection {
			routeStatus := GetRouteStatus(tlsRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tlsRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonNoMatchingListenerHostname,
				"There were no hostname intersections between the TLSRoute and this parent ref's Listener(s).",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(tlsRoute, gwapiv1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if parentRef.TLSRoute != nil &&
			len(parentRef.TLSRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			routeStatus := GetRouteStatus(tlsRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tlsRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonAccepted,
				"Route is accepted",
			)
		}
	}
}

func (t *Translator) ProcessUDPRoutes(udpRoutes []*gwapiv1.UDPRoute, gateways []*GatewayContext, resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) []*UDPRouteContext {
	relevantUDPRoutes := make([]*UDPRouteContext, 0, len(udpRoutes))
	// UDPRoutes are already sorted by the provider layer

	for _, u := range udpRoutes {
		if u == nil {
			panic("received nil udproute")
		}
		udpRoute := &UDPRouteContext{UDPRoute: u}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(udpRoute, gateways)
		if !relevantRoute {
			continue
		}

		relevantUDPRoutes = append(relevantUDPRoutes, udpRoute)

		t.processUDPRouteParentRefs(udpRoute, resources, xdsIR)
	}

	return relevantUDPRoutes
}

func (t *Translator) processUDPRouteParentRefs(udpRoute *UDPRouteContext, resources *resource.Resources, xdsIR resource.XdsIRMap) {
	for _, parentRef := range udpRoute.ParentRefs {
		// compute backends
		if len(udpRoute.Spec.Rules) != 1 {
			routeStatus := GetRouteStatus(udpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				udpRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				"InvalidRule",
				"One and only one rule is supported",
			)
			continue
		}

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		// udpRoute must have a single rule, so Spec.Rules[0] is always safe to index below.
		var (
			routeBackendDestinations []routeBackendRefDestination
			resolveErrs              = &status.MultiStatusError{}
			destName                 = irRouteDestinationName(udpRoute, -1 /*rule index*/)
			routeRuleMetadata        = buildResourceMetadata(udpRoute, udpRoute.Spec.Rules[0].Name)
		)

		gatewayCtx := GetRouteParentContext(udpRoute, *parentRef.ParentReference, t.GatewayControllerName).GetGateway()
		btpRoutingType := t.resolveBTPRoutingType(gatewayCtx, udpRoute, parentRef, udpRoute.Spec.Rules[0].Name)
		btpEndpointHostname := t.resolveBTPEndpointHostname(gatewayCtx, udpRoute, parentRef, udpRoute.Spec.Rules[0].Name)

		var mergeIncompatible bool
		if t.isMergeBackendsEnabledForGateway(gatewayCtx) {
			backendRefs := toBackendObjectReferences(udpRoute.Spec.Rules[0].BackendRefs, func(r gwapiv1.BackendRef) gwapiv1.BackendObjectReference { return r.BackendObjectReference })
			mergeIncompatible = t.mergeIncompatibleForSingleClusterRule(backendRefs)
		}

		for i := range udpRoute.Spec.Rules[0].BackendRefs {
			backendRefCtx := DirectBackendRef{BackendRef: &udpRoute.Spec.Rules[0].BackendRefs[i]}

			// backendDest.ds will never be nil here because processDestination returns an empty DestinationSetting for invalid backendRefs.
			backendDest, _, err := t.processBackendRef(destName, i, backendRefCtx, parentRef, udpRoute, resources, gatewayCtx, btpRoutingType, btpEndpointHostname, xdsIR, mergeIncompatible)
			if err != nil {
				resolveErrs.Add(err)
				continue
			}

			// skip backendRefs with weight 0 as they do not affect the traffic distribution
			if backendDest.ds.Weight != nil && *backendDest.ds.Weight > 0 {
				routeBackendDestinations = append(routeBackendDestinations, backendDest)
			}
		}

		routeStatus := GetRouteStatus(udpRoute)
		if !resolveErrs.Empty() {
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				udpRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				resolveErrs.Reason(),
				resolveErrs.Error())
		} else {
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				udpRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(udpRoute, gwapiv1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		accepted := false
		for _, listener := range parentRef.listeners {
			accepted = true
			listener.IncrementAttachedRoutes()
			if !listener.IsReady() {
				continue
			}

			irKey := t.getIRKey(listener.gateway.Gateway)

			gwXdsIR := xdsIR[irKey]
			irListener := gwXdsIR.GetUDPListener(irListenerName(listener))
			// When multiple UDPRoutes target the same Gateway listener, all of must report Accepted=True.
			// Only the oldest route is attached to the listener, and the listener's AttachedRoutes count must reflect this.
			// https://github.com/kubernetes-sigs/gateway-api/blob/cf34ac933d068c6008598cce945819ce9cee16be/conformance/tests/udproute-multiple-routes-attachment.go#L107
			if irListener != nil && irListener.Route == nil {
				irListener.Route = &ir.UDPRoute{
					Name: irUDPRouteName(udpRoute),
					Destination: t.routeDestinationForListener(
						gwXdsIR,
						gatewayCtx,
						udpRoute,
						listener,
						udpRoute.Spec.Rules[0].Name,
						destName,
						routeRuleMetadata,
						nil,
						routeBackendDestinations,
					),
				}
			}
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if accepted && parentRef.UDPRoute != nil &&
			len(parentRef.UDPRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			routeStatus := GetRouteStatus(udpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				udpRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonAccepted,
				"Route is accepted",
			)
		}

		if !accepted {
			routeStatus := GetRouteStatus(udpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				udpRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
				"Multiple routes on the same UDP listener",
			)
		}
	}
}

func (t *Translator) ProcessTCPRoutes(tcpRoutes []*gwapiv1.TCPRoute, gateways []*GatewayContext, resources *resource.Resources,
	xdsIR resource.XdsIRMap,
) []*TCPRouteContext {
	relevantTCPRoutes := make([]*TCPRouteContext, 0, len(tcpRoutes))
	// TCPRoutes are already sorted by the provider layer

	for _, tcp := range tcpRoutes {
		if tcp == nil {
			panic("received nil tcproute")
		}
		tcpRoute := &TCPRouteContext{TCPRoute: tcp}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(tcpRoute, gateways)
		if !relevantRoute {
			continue
		}

		relevantTCPRoutes = append(relevantTCPRoutes, tcpRoute)

		t.processTCPRouteParentRefs(tcpRoute, resources, xdsIR)
	}

	return relevantTCPRoutes
}

func (t *Translator) processTCPRouteParentRefs(tcpRoute *TCPRouteContext, resources *resource.Resources, xdsIR resource.XdsIRMap) {
	for _, parentRef := range tcpRoute.ParentRefs {
		// compute backends
		if len(tcpRoute.Spec.Rules) != 1 {
			routeStatus := GetRouteStatus(tcpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tcpRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				"InvalidRule",
				"One and only one rule is supported",
			)
			continue
		}

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		var (
			routeBackendDestinations []routeBackendRefDestination
			resolveErrs              = &status.MultiStatusError{}
			destName                 = irRouteDestinationName(tcpRoute, -1 /*rule index*/)
			routeRuleMetadata        = buildResourceMetadata(tcpRoute, tcpRoute.Spec.Rules[0].Name)
		)

		gatewayCtx := GetRouteParentContext(tcpRoute, *parentRef.ParentReference, t.GatewayControllerName).GetGateway()
		btpRoutingType := t.resolveBTPRoutingType(gatewayCtx, tcpRoute, parentRef, tcpRoute.Spec.Rules[0].Name)
		btpEndpointHostname := t.resolveBTPEndpointHostname(gatewayCtx, tcpRoute, parentRef, tcpRoute.Spec.Rules[0].Name)

		var mergeIncompatible bool
		if t.isMergeBackendsEnabledForGateway(gatewayCtx) {
			backendRefs := toBackendObjectReferences(tcpRoute.Spec.Rules[0].BackendRefs, func(r gwapiv1.BackendRef) gwapiv1.BackendObjectReference { return r.BackendObjectReference })
			mergeIncompatible = t.mergeIncompatibleForSingleClusterRule(backendRefs)
		}

		for i := range tcpRoute.Spec.Rules[0].BackendRefs {
			backendRefCtx := DirectBackendRef{BackendRef: &tcpRoute.Spec.Rules[0].BackendRefs[i]}
			backendDest, _, err := t.processBackendRef(destName, i, backendRefCtx, parentRef, tcpRoute, resources, gatewayCtx, btpRoutingType, btpEndpointHostname, xdsIR, mergeIncompatible)
			// skip adding the route and provide the reason via route status.
			if err != nil {
				resolveErrs.Add(err)
				continue
			}

			// skip backendRefs with weight 0 as they do not affect the traffic distribution
			if backendDest.ds.Weight != nil && *backendDest.ds.Weight > 0 {
				routeBackendDestinations = append(routeBackendDestinations, backendDest)
			}
		}

		routeStatus := GetRouteStatus(tcpRoute)
		if !resolveErrs.Empty() {
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tcpRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				resolveErrs.Reason(),
				resolveErrs.Error())
		} else {
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tcpRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(tcpRoute, gwapiv1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		accepted := false
		for _, listener := range parentRef.listeners {
			accepted = true
			listener.IncrementAttachedRoutes()
			if !listener.IsReady() {
				continue
			}
			irKey := t.getIRKey(listener.gateway.Gateway)

			gwXdsIR := xdsIR[irKey]
			irListener := gwXdsIR.GetTCPListener(irListenerName(listener))
			// When multiple TCPRoutes target the same Gateway listener, all of must report Accepted=True.
			// Only the oldest route is attached to the listener, and the listener's AttachedRoutes count must reflect this.
			// https://github.com/kubernetes-sigs/gateway-api/blob/cf34ac933d068c6008598cce945819ce9cee16be/conformance/tests/tcproute-multiple-routes-attachment.go#L104
			if irListener != nil && len(irListener.Routes) == 0 {
				irRoute := &ir.TCPRoute{
					Name: irTCPRouteName(tcpRoute),
					Destination: t.routeDestinationForListener(
						gwXdsIR,
						gatewayCtx,
						tcpRoute,
						listener,
						tcpRoute.Spec.Rules[0].Name,
						destName,
						routeRuleMetadata,
						nil,
						routeBackendDestinations,
					),
					Metadata: buildResourceMetadata(tcpRoute, nil),
				}

				if irListener.TLS != nil {
					irRoute.TLS = &ir.TLS{Terminate: irListener.TLS}

					if listener.Hostname != nil {
						irRoute.TLS.TLSInspectorConfig = &ir.TLSInspectorConfig{
							SNIs: []string{string(*listener.Hostname)},
						}
					}
				}

				irListener.Routes = append(irListener.Routes, irRoute)
			}
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if accepted && parentRef.TCPRoute != nil &&
			len(parentRef.TCPRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			routeStatus := GetRouteStatus(tcpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tcpRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonAccepted,
				"Route is accepted",
			)
		}
		if !accepted {
			routeStatus := GetRouteStatus(tcpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tcpRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue,
				"Multiple routes on the same TCP listener",
			)
		}

	}
}

// processDestination translates a backendRef into a destination settings.
// If an error occurs during this conversion, an error is returned, and the associated routes are expected to become inactive.
// This will result in a direct 500 response for HTTP-based requests.
func (t *Translator) processDestination(name string, backendRefContext BackendRefContext,
	parentRef *RouteParentContext, route RouteContext, resources *resource.Resources,
	gatewayCtx *GatewayContext, btpRoutingType *egv1a1.RoutingType,
	btpEndpointHostname *egv1a1.BackendEndpointHostname, xdsIR resource.XdsIRMap,
) (ds *ir.DestinationSetting, unstructuredRef *ir.UnstructuredRef, err status.Error) {
	var (
		routeType  = route.GetRouteType()
		weight     = uint32(ptr.Deref(backendRefContext.GetBackendRef().Weight, int32(1)))
		backendRef = backendRefContext.GetBackendRef()
	)

	// Create an empty DS without endpoints
	// This represents an invalid DS.
	emptyDS := &ir.DestinationSetting{
		Name:   name,
		Weight: &weight,
	}

	backendNamespace := NamespaceDerefOr(backendRef.Namespace, route.GetNamespace())
	if !t.isCustomBackendResource(backendRef.Group, KindDerefOr(backendRef.Kind, resource.KindService)) {
		err = t.validateBackendRef(backendRefContext, route, resources, backendNamespace, routeType)
		{
			// Empty DS means the backend is invalid and an error to fail the associated route.
			if err != nil {
				return emptyDS, nil, err
			}
		}
	} else {
		// Custom backend resources still require ReferenceGrant for cross-namespace references.
		if err = t.validateBackendNamespace(backendRef, route, resources, routeType); err != nil {
			return emptyDS, nil, err
		}
	}

	// Skip processing backends with 0 weight
	if weight == 0 {
		return emptyDS, nil, nil
	}

	var envoyProxy *egv1a1.EnvoyProxy
	if gatewayCtx != nil {
		envoyProxy = gatewayCtx.envoyProxy
	}

	protocol := inspectAppProtocolByRouteKind(routeType)

	// Process BackendTLSPolicy first to ensure status is set.
	tls, tlsErr := t.applyBackendTLSSetting(
		backendRef.BackendObjectReference,
		backendNamespace,
		gwapiv1.ParentReference{
			Group:       parentRef.Group,
			Kind:        parentRef.Kind,
			Namespace:   parentRef.Namespace,
			Name:        parentRef.Name,
			SectionName: parentRef.SectionName,
			Port:        parentRef.Port,
		},
		resources,
		gatewayCtx,
	)
	if tlsErr != nil {
		return emptyDS, nil, status.NewRouteStatusError(tlsErr, status.RouteReasonInvalidBackendTLS)
	}

	switch KindDerefOr(backendRef.Kind, resource.KindService) {
	case resource.KindServiceImport:
		ds, err = t.processServiceImportDestinationSetting(name, backendRef.BackendObjectReference, backendNamespace, protocol, envoyProxy, btpRoutingType)
		if err != nil {
			return emptyDS, nil, err
		}
	case resource.KindService:
		ds, err = t.processServiceDestinationSetting(name, backendRef.BackendObjectReference, backendNamespace, protocol, envoyProxy, btpRoutingType, btpEndpointHostname)
		if err != nil {
			return emptyDS, nil, err
		}
		svc := t.GetService(backendNamespace, string(backendRef.Name))
		ds.IPFamily = getServiceIPFamily(svc)
		ds.PreferLocal = processPreferLocalZone(svc)
	case egv1a1.KindBackend:
		ds = t.processBackendDestinationSetting(name, backendRef.BackendObjectReference, backendNamespace, protocol)
	default:
		// Handle custom backend resources defined in extension manager
		if t.isCustomBackendResource(backendRef.Group, KindDerefOr(backendRef.Kind, resource.KindService)) {
			// Add the custom backend resource to ExtensionRefFilters so it can be processed by the extension system
			unstructuredRef = t.processBackendExtensions(backendRef.BackendObjectReference, backendNamespace, resources)

			// Check if the custom backend resource was found
			if unstructuredRef == nil {
				return emptyDS, nil, status.NewRouteStatusError(
					fmt.Errorf("custom backend %s %s/%s not found",
						KindDerefOr(backendRef.Kind, resource.KindService),
						backendNamespace,
						backendRef.Name),
					gwapiv1.RouteReasonBackendNotFound,
				).WithType(gwapiv1.RouteConditionResolvedRefs)
			}

			return &ir.DestinationSetting{
				Name:            name,
				Weight:          &weight,
				IsCustomBackend: true,
			}, unstructuredRef, nil
		}
	}

	ds.TLS = tls

	var filtersErr error
	ds.Filters, filtersErr = t.processDestinationFilters(routeType, backendRefContext, parentRef, route, resources, xdsIR)
	if filtersErr != nil {
		return emptyDS, nil, status.NewRouteStatusError(filtersErr, status.RouteReasonInvalidBackendFilters)
	}
	if err := validateDestinationSettings(ds, t.IsServiceRouting(envoyProxy, btpRoutingType), backendRef.Kind); err != nil {
		return emptyDS, nil, err
	}

	ds.Weight = &weight
	return ds, nil, nil
}

func validateDestinationSettings(destinationSettings *ir.DestinationSetting, isServiceRouting bool, kind *gwapiv1.Kind) status.Error {
	// TODO: support mixed endpointslice address type for the same backendRef
	switch KindDerefOr(kind, resource.KindService) {
	case egv1a1.KindBackend:
		if destinationSettings.AddressType != nil && *destinationSettings.AddressType == ir.MIXED {
			return status.NewRouteStatusError(
				fmt.Errorf("mixed FQDN and IP or Unix address type for the same backendRef is not supported"),
				status.RouteReasonUnsupportedAddressType)
		}
	case resource.KindService, resource.KindServiceImport:
		if !isServiceRouting && destinationSettings.AddressType != nil && *destinationSettings.AddressType == ir.MIXED {
			return status.NewRouteStatusError(
				fmt.Errorf("mixed endpointslice address type for the same backendRef is not supported"),
				status.RouteReasonUnsupportedAddressType)
		}
	}

	return nil
}

func (t *Translator) dnsDomain() string {
	if t.DNSDomain != "" {
		return t.DNSDomain
	}
	return config.DefaultDNSDomain
}

// isServiceHeadless reports true when a Kubernetes Service is headless.
func isServiceHeadless(service *corev1.Service) bool {
	if service == nil {
		return false
	}
	if service.Spec.ClusterIP == corev1.ClusterIPNone {
		return true
	}
	return false
}

// isServiceExternalName reports true when a Kubernetes Service is of type ExternalName.
// ExternalName Services have no ClusterIP and no EndpointSlices, so they cannot be
// translated into a valid backend and are not supported.
func isServiceExternalName(service *corev1.Service) bool {
	return service != nil && service.Spec.Type == corev1.ServiceTypeExternalName
}

func (t *Translator) processServiceImportDestinationSetting(
	name string,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	defaultProtocol ir.AppProtocol,
	envoyProxy *egv1a1.EnvoyProxy,
	btpRoutingType *egv1a1.RoutingType,
) (*ir.DestinationSetting, status.Error) {
	var (
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
		protocol  = defaultProtocol
	)

	serviceImport := t.GetServiceImport(backendNamespace, string(backendRef.Name))
	var servicePort mcsapiv1a1.ServicePort
	for _, port := range serviceImport.Spec.Ports {
		if port.Port == *backendRef.Port {
			servicePort = port
			break
		}
	}

	if servicePort.AppProtocol != nil {
		protocol = resolveBackendProtocol(*servicePort.AppProtocol, protocol)
	}
	// For WebSocket backends, force HTTP/1.1 upstream to ensure Envoy can establish a successful connection,
	// as WebSocket over HTTP/2 is not widely supported by upstreams and can lead to connection failures.
	forceHTTP1Upstream := shouldForceHTTP1Upstream(protocol, servicePort.AppProtocol)

	backendIps := serviceImport.Spec.IPs
	isHeadless := len(backendIps) == 0

	// Route to endpoints by default, or if service routing is enabled but ServiceImport is headless
	useEndpointRouting := !t.IsServiceRouting(envoyProxy, btpRoutingType) || isHeadless
	if useEndpointRouting {
		endpointSlices := t.GetEndpointSlicesForBackend(backendNamespace, string(backendRef.Name), resource.KindServiceImport)
		endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, servicePort.Name, getServicePortProtocol(servicePort.Protocol), nil)
		if len(endpoints) == 0 {
			return nil, status.NewRouteStatusError(
				fmt.Errorf("no ready endpoints for the related ServiceImport %s/%s", backendNamespace, backendRef.Name),
				status.RouteReasonEndpointsNotFound,
			)
		}
	} else {
		// Use ServiceImport IPs for routing
		for _, ip := range backendIps {
			ep := ir.NewDestEndpoint(nil, ip, uint32(*backendRef.Port), false, nil)
			endpoints = append(endpoints, ep)
		}
	}

	return &ir.DestinationSetting{
		Name:               name,
		Protocol:           protocol,
		ForceHTTP1Upstream: forceHTTP1Upstream,
		Endpoints:          endpoints,
		AddressType:        addrType,
		Metadata:           buildResourceMetadata(serviceImport, new(gwapiv1.SectionName(strconv.Itoa(int(*backendRef.Port))))),
	}, nil
}

func (t *Translator) processServiceDestinationSetting(
	name string,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	defaultProtocol ir.AppProtocol,
	envoyProxy *egv1a1.EnvoyProxy,
	btpRoutingType *egv1a1.RoutingType,
	btpEndpointHostname *egv1a1.BackendEndpointHostname,
) (*ir.DestinationSetting, status.Error) {
	var (
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
	)
	protocol := defaultProtocol

	service := t.GetService(backendNamespace, string(backendRef.Name))
	// ExternalName Services have no ClusterIP and no EndpointSlices, so they cannot be
	// translated into a valid backend.
	// Backend with FQDN endpoint should be used instead of ExternalName Service to route to external services.
	if isServiceExternalName(service) {
		return nil, status.NewRouteStatusError(
			fmt.Errorf("Service %s/%s is of type ExternalName, which is not supported as a backend; "+
				"use an Envoy Gateway Backend resource with an FQDN endpoint instead",
				backendNamespace, string(backendRef.Name)),
			gwapiv1.RouteReasonUnsupportedValue)
	}
	var servicePort corev1.ServicePort
	for _, port := range service.Spec.Ports {
		if port.Port == *backendRef.Port {
			servicePort = port
			break
		}
	}

	// support HTTPRouteBackendProtocolH2C/GRPC
	if servicePort.AppProtocol != nil {
		protocol = resolveBackendProtocol(*servicePort.AppProtocol, protocol)
	}
	// For WebSocket backends, force HTTP/1.1 upstream to ensure Envoy can establish a successful connection,
	// as WebSocket over HTTP/2 is not widely supported by upstreams and can lead to connection failures.
	forceHTTP1Upstream := shouldForceHTTP1Upstream(protocol, servicePort.AppProtocol)

	isHeadless := isServiceHeadless(service)

	// Route to endpoints by default, or if service routing is enabled but service is headless
	useEndpointRouting := !t.IsServiceRouting(envoyProxy, btpRoutingType) || isHeadless
	endpointHostname := t.serviceEndpointHostname(service, btpEndpointHostname)
	if useEndpointRouting {
		endpointSlices := t.GetEndpointSlicesForBackend(backendNamespace, string(backendRef.Name), KindDerefOr(backendRef.Kind, resource.KindService))
		endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, servicePort.Name, getServicePortProtocol(servicePort.Protocol), endpointHostname)
		if len(endpoints) == 0 {
			return nil, status.NewRouteStatusError(
				fmt.Errorf("no ready endpoints for the related Service %s/%s", backendNamespace, backendRef.Name),
				status.RouteReasonEndpointsNotFound,
			)
		}
	} else {
		// Use Service ClusterIP routing
		ep := ir.NewDestEndpoint(endpointHostname, service.Spec.ClusterIP, uint32(*backendRef.Port), false, nil)
		endpoints = append(endpoints, ep)
	}

	return &ir.DestinationSetting{
		Name:               name,
		Protocol:           protocol,
		ForceHTTP1Upstream: forceHTTP1Upstream,
		Endpoints:          endpoints,
		AddressType:        addrType,
		PreferLocal:        processPreferLocalZone(service),
		Metadata:           buildResourceMetadata(service, new(gwapiv1.SectionName(strconv.Itoa(int(*backendRef.Port))))),
	}, nil
}

func (t *Translator) serviceEndpointHostname(service *corev1.Service, endpointHostname *egv1a1.BackendEndpointHostname) *string {
	if endpointHostname == nil {
		return nil
	}

	switch endpointHostname.Type {
	case egv1a1.BackendEndpointHostnameTypeKubernetesService:
		if service == nil || service.Name == "" || service.Namespace == "" {
			return nil
		}
		return new(fmt.Sprintf("%s.%s.svc.%s", service.Name, service.Namespace, t.dnsDomain()))
	case egv1a1.BackendEndpointHostnameTypeStatic:
		if endpointHostname.Hostname == nil || *endpointHostname.Hostname == "" {
			return nil
		}
		return endpointHostname.Hostname
	default:
		return nil
	}
}

func getBackendFilters(routeType gwapiv1.Kind, backendRefContext BackendRefContext) (backendFilters any) {
	filters := backendRefContext.GetFilters()
	if filters == nil {
		return nil
	}
	switch routeType {
	case resource.KindHTTPRoute:
		if len(filters.([]gwapiv1.HTTPRouteFilter)) > 0 {
			return filters.([]gwapiv1.HTTPRouteFilter)
		}
	case resource.KindGRPCRoute:
		if len(filters.([]gwapiv1.GRPCRouteFilter)) > 0 {
			return filters.([]gwapiv1.GRPCRouteFilter)
		}
	}

	return nil
}

func processPreferLocalZone(svc *corev1.Service) *ir.PreferLocalZone {
	if svc == nil {
		return nil
	}

	if trafficDist := svc.Spec.TrafficDistribution; trafficDist != nil {
		return &ir.PreferLocalZone{
			MinEndpointsThreshold: new(uint64(1)),
			Force: &ir.ForceLocalZone{
				MinEndpointsInZoneThreshold: new(uint32(1)),
			},
		}
	}

	// Allows annotation values that align with Kubernetes defaults.
	// Ref:
	// https://kubernetes.io/docs/concepts/services-networking/topology-aware-routing/#enabling-topology-aware-routing
	// https://github.com/kubernetes/kubernetes/blob/9d9e1afdf78bce0a517cc22557457f942040ca19/staging/src/k8s.io/endpointslice/utils.go#L355-L368
	if val, ok := svc.Annotations[corev1.AnnotationTopologyMode]; ok && val == "Auto" || val == "auto" {
		return &ir.PreferLocalZone{
			MinEndpointsThreshold: new(uint64(3)),
			Force: &ir.ForceLocalZone{
				MinEndpointsInZoneThreshold: new(uint32(3)),
			},
		}
	}

	return nil
}

func (t *Translator) processDestinationFilters(routeType gwapiv1.Kind, backendRefContext BackendRefContext, parentRef *RouteParentContext, route RouteContext, resources *resource.Resources, xdsIR resource.XdsIRMap) (*ir.DestinationFilters, error) {
	backendFilters := getBackendFilters(routeType, backendRefContext)
	if backendFilters == nil {
		return nil, nil
	}

	var httpFiltersContext *HTTPFiltersContext
	var destFilters ir.DestinationFilters

	var errs []status.Error
	switch filters := backendFilters.(type) {
	case []gwapiv1.HTTPRouteFilter:
		httpFiltersContext, errs = t.ProcessHTTPFilters(parentRef, route, filters, 0, resources, xdsIR)
	case []gwapiv1.GRPCRouteFilter:
		httpFiltersContext, errs = t.ProcessGRPCFilters(parentRef, route, filters, resources, xdsIR)
	}
	if len(errs) > 0 {
		var err error
		for _, e := range errs {
			err = errors.Join(err, e)
		}
		return nil, err
	}
	applyHTTPFiltersContextToDestinationFilters(httpFiltersContext, &destFilters)

	return &destFilters, nil
}

func applyHTTPFiltersContextToDestinationFilters(httpFiltersContext *HTTPFiltersContext, destFilters *ir.DestinationFilters) {
	if len(httpFiltersContext.AddRequestHeaders) > 0 {
		destFilters.AddRequestHeaders = httpFiltersContext.AddRequestHeaders
	}
	if len(httpFiltersContext.RemoveRequestHeaders) > 0 {
		destFilters.RemoveRequestHeaders = httpFiltersContext.RemoveRequestHeaders
	}
	if len(httpFiltersContext.AddResponseHeaders) > 0 {
		destFilters.AddResponseHeaders = httpFiltersContext.AddResponseHeaders
	}
	if len(httpFiltersContext.RemoveResponseHeaders) > 0 {
		destFilters.RemoveResponseHeaders = httpFiltersContext.RemoveResponseHeaders
	}
	if httpFiltersContext.CredentialInjection != nil {
		destFilters.CredentialInjection = httpFiltersContext.CredentialInjection
	}
	if httpFiltersContext.URLRewrite != nil {
		destFilters.URLRewrite = httpFiltersContext.URLRewrite
	}
}

func inspectAppProtocolByRouteKind(kind gwapiv1.Kind) ir.AppProtocol {
	switch kind {
	case resource.KindUDPRoute:
		return ir.UDP
	case resource.KindHTTPRoute:
		return ir.HTTP
	case resource.KindTCPRoute:
		return ir.TCP
	case resource.KindGRPCRoute:
		return ir.GRPC
	case resource.KindTLSRoute:
		// TLSRoute is translated into an ir.TCPRoute. The upstream protocol is
		// plain TCP from Envoy's perspective; whether the upstream connection is
		// secured by TLS is governed by ds.TLS (e.g. via BackendTLSPolicy), not by
		// the AppProtocol. Returning ir.TCP keeps the IR semantically accurate and
		// avoids emitting HTTP protocol options on a TCP proxy cluster.
		return ir.TCP
	}
	return ir.TCP
}

// processAllowedListenersForParentRefs finds out if the route attaches to one of our Gateways' listeners or the attached
// XListenerSet's listeners, and if so, gets the list of listeners that allow it to attach for each parentRef.
func (t *Translator) processAllowedListenersForParentRefs(
	routeContext RouteContext,
	gateways []*GatewayContext,
) bool {
	var relevantRoute bool
	ns := gwapiv1.Namespace(routeContext.GetNamespace())
	for _, parentRef := range GetParentReferences(routeContext) {
		isRelevantParentRef, selectedListeners := GetReferencedListeners(ns, parentRef, gateways)

		// Parent ref is not to a Gateway that we control: skip it
		if !isRelevantParentRef {
			continue
		}
		relevantRoute = true

		parentRefCtx := GetRouteParentContext(routeContext, parentRef, t.GatewayControllerName)
		// Reset conditions since they will be recomputed during translation
		parentRefCtx.ResetConditions(routeContext)

		if len(selectedListeners) == 0 {
			routeStatus := GetRouteStatus(routeContext)
			status.SetRouteStatusCondition(routeStatus,
				parentRefCtx.routeParentStatusIdx,
				routeContext.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonNoMatchingParent,
				"No listeners match this parent ref",
			)
			continue
		}

		var allowedListeners []*ListenerContext
		for _, listener := range selectedListeners {
			acceptedKind := routeContext.GetRouteType()
			if listener.AllowsKind(gwapiv1.RouteGroupKind{Group: GroupPtr(gwapiv1.GroupName), Kind: acceptedKind}) &&
				listener.AllowsNamespace(t.GetNamespace(routeContext.GetNamespace())) {
				allowedListeners = append(allowedListeners, listener)
			}
		}

		if len(allowedListeners) == 0 {
			routeStatus := GetRouteStatus(routeContext)
			status.SetRouteStatusCondition(routeStatus,
				parentRefCtx.routeParentStatusIdx,
				routeContext.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonNotAllowedByListeners,
				"No listeners included by this parent ref allowed this attachment.",
			)
			continue
		}
		parentRefCtx.SetListeners(allowedListeners...)

		routeStatus := GetRouteStatus(routeContext)
		status.SetRouteStatusCondition(routeStatus,
			parentRefCtx.routeParentStatusIdx,
			routeContext.GetGeneration(),
			gwapiv1.RouteConditionAccepted,
			metav1.ConditionTrue,
			gwapiv1.RouteReasonAccepted,
			"Route is accepted",
		)
	}
	return relevantRoute
}

func getIREndpointsFromEndpointSlices(endpointSlices []*discoveryv1.EndpointSlice, portName string, portProtocol corev1.Protocol, endpointHostname *string) ([]*ir.DestinationEndpoint, *ir.DestinationAddressType) {
	var (
		dstEndpoints []*ir.DestinationEndpoint
		dstAddrType  *ir.DestinationAddressType
	)

	addrTypeMap := make(map[ir.DestinationAddressType]int)
	for _, endpointSlice := range endpointSlices {
		if endpointSlice.AddressType == discoveryv1.AddressTypeFQDN {
			addrTypeMap[ir.FQDN]++
		} else {
			addrTypeMap[ir.IP]++
		}
		endpoints := getIREndpointsFromEndpointSlice(endpointSlice, portName, portProtocol, endpointHostname)
		dstEndpoints = append(dstEndpoints, endpoints...)
	}

	for addrTypeState, addrTypeCounts := range addrTypeMap {
		if addrTypeCounts == len(endpointSlices) {
			dstAddrType = new(addrTypeState)
			break
		}
	}

	if len(addrTypeMap) > 0 && dstAddrType == nil {
		dstAddrType = new(ir.MIXED)
	}

	return dstEndpoints, dstAddrType
}

func getIREndpointsFromEndpointSlice(endpointSlice *discoveryv1.EndpointSlice, portName string, portProtocol corev1.Protocol, endpointHostname *string) []*ir.DestinationEndpoint {
	var endpoints []*ir.DestinationEndpoint
	for _, endpoint := range endpointSlice.Endpoints {
		for _, endpointPort := range endpointSlice.Ports {
			// Check if the endpoint port matches the service port
			if *endpointPort.Name != portName || *endpointPort.Protocol != portProtocol {
				continue
			}
			conditions := endpoint.Conditions

			// Unknown Serving/Terminating (nil) should fall-back to Ready, see https://pkg.go.dev/k8s.io/api/discovery/v1#EndpointConditions
			// So drain the endpoint if:
			// 1. Both `Terminating` and `Serving` are != null, and either `Terminating=true` or `Serving=false`
			// 2. Or `Ready=false`
			var draining bool
			if conditions.Serving != nil && conditions.Terminating != nil {
				draining = *conditions.Terminating || !*conditions.Serving
			} else {
				draining = conditions.Ready != nil && !*conditions.Ready
			}

			for _, address := range endpoint.Addresses {
				ep := ir.NewDestEndpoint(endpointHostname, address, uint32(*endpointPort.Port), draining, endpoint.Zone)
				endpoints = append(endpoints, ep)
			}

		}
	}

	return endpoints
}

// isCustomBackendResource checks if the given group and kind match any of the configured custom backend resources
func (t *Translator) isCustomBackendResource(group *gwapiv1.Group, kind string) bool {
	groupStr := GroupDerefOr(group, "")
	for _, gk := range t.ExtensionGroupKinds {
		if gk.Group == groupStr && gk.Kind == kind {
			return true
		}
	}
	return false
}

// addCustomBackendToExtensionRefs adds custom backend resources to the ExtensionRefFilters
// so they can be processed by the extension system
func (t *Translator) processBackendExtensions(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	resources *resource.Resources,
) *ir.UnstructuredRef { // This list of resources will be empty unless an extension is loaded (and introduces resources)
	for _, res := range resources.ExtensionRefFilters {
		if res.GetKind() == string(*backendRef.Kind) && res.GetName() == string(backendRef.Name) && res.GetNamespace() == backendNamespace {
			apiVers := res.GetAPIVersion()
			// To get only the group we cut off the version.
			// This could be a one liner but just to be safe we check that the APIVersion is properly formatted
			idx := strings.IndexByte(apiVers, '/')
			if idx != -1 {
				group := apiVers[:idx]
				if group == string(*backendRef.Group) {
					res := res // Capture loop variable
					return &ir.UnstructuredRef{Object: &res}
				}
			}
		}
	}
	return nil
}

func (t *Translator) getTargetBackendReference(
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
) gwapiv1.LocalPolicyTargetReferenceWithSectionName {
	ref := gwapiv1.LocalPolicyTargetReferenceWithSectionName{
		LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
			Group: func() gwapiv1.Group {
				if backendRef.Group == nil || *backendRef.Group == "" {
					return ""
				}
				return *backendRef.Group
			}(),
			Kind: func() gwapiv1.Kind {
				if backendRef.Kind == nil || *backendRef.Kind == resource.KindService {
					return "Service"
				}
				return *backendRef.Kind
			}(),
			Name: backendRef.Name,
		},
	}
	if backendRef.Port == nil {
		return ref
	}

	switch {
	case backendRef.Kind == nil || *backendRef.Kind == resource.KindService:
		if service := t.GetService(backendNamespace, string(backendRef.Name)); service != nil {
			for _, port := range service.Spec.Ports {
				if port.Port == *backendRef.Port {
					if port.Name != "" {
						ref.SectionName = SectionNamePtr(port.Name)
						break
					}
				}
			}
		}

	case *backendRef.Kind == resource.KindServiceImport:
		if si := t.GetServiceImport(backendNamespace, string(backendRef.Name)); si != nil {
			for _, port := range si.Spec.Ports {
				if port.Port == *backendRef.Port {
					if port.Name != "" {
						ref.SectionName = SectionNamePtr(port.Name)
						break
					}
				}
			}
		}

	default:
		// Set the section name to the port number if the backend is a EG Backend
		ref.SectionName = SectionNamePtr(strconv.Itoa(int(*backendRef.Port)))
	}

	return ref
}

func (t *Translator) processBackendDestinationSetting(
	name string,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	defaultProtocol ir.AppProtocol,
) *ir.DestinationSetting {
	var dstAddrType *ir.DestinationAddressType
	protocol := defaultProtocol
	forceHTTP1Upstream := false

	addrTypeMap := make(map[ir.DestinationAddressType]int)
	backend := t.GetBackend(backendNamespace, string(backendRef.Name))
	for _, ap := range backend.Spec.AppProtocols {
		protocol = resolveBackendProtocol(string(ap), protocol)
		// For WebSocket backends, force HTTP/1.1 upstream to ensure Envoy can establish a successful connection,
		// as WebSocket over HTTP/2 is not widely supported by upstreams and can lead to connection failures.
		forceHTTP1Upstream = forceHTTP1Upstream || shouldForceHTTP1Upstream(protocol, (*string)(&ap))
	}

	ds := &ir.DestinationSetting{Name: name}

	// There is only one backend if it is a dynamic resolver
	if backend.Spec.Type != nil && *backend.Spec.Type == egv1a1.BackendTypeDynamicResolver {
		ds.IsDynamicResolver = true
		ds.Protocol = protocol
		ds.ForceHTTP1Upstream = forceHTTP1Upstream
		return ds
	}

	dstEndpoints := make([]*ir.DestinationEndpoint, 0, len(backend.Spec.Endpoints))

	for _, bep := range backend.Spec.Endpoints {
		var irde *ir.DestinationEndpoint
		switch {
		case bep.IP != nil:
			ip := net.ParseIP(bep.IP.Address)
			if ip != nil {
				addrTypeMap[ir.IP]++
				irde = ir.NewDestEndpoint(bep.Hostname, bep.IP.Address, uint32(bep.IP.Port), false, bep.Zone)
			}
		case bep.FQDN != nil:
			addrTypeMap[ir.FQDN]++
			irde = ir.NewDestEndpoint(bep.Hostname, bep.FQDN.Hostname, uint32(bep.FQDN.Port), false, bep.Zone)
		case bep.Unix != nil:
			addrTypeMap[ir.UDS]++
			irde = &ir.DestinationEndpoint{
				Hostname: bep.Hostname,
				Path:     new(bep.Unix.Path),
				Zone:     bep.Zone,
			}
		}

		dstEndpoints = append(dstEndpoints, irde)
	}

	for addrTypeState, addrTypeCounts := range addrTypeMap {
		if addrTypeCounts == len(backend.Spec.Endpoints) {
			dstAddrType = new(addrTypeState)
			break
		}
	}

	// more than 1 type of addr
	if len(addrTypeMap) > 0 && dstAddrType == nil {
		// if one of the types is FQDN, the other is UDS/IP, so mixed endpoints
		if _, hasFQDN := addrTypeMap[ir.FQDN]; hasFQDN {
			dstAddrType = new(ir.MIXED)
		} else { // otherwise
			dstAddrType = new(ir.STATIC)
		}
	}

	ds.Endpoints = dstEndpoints
	ds.AddressType = dstAddrType
	ds.Protocol = protocol
	ds.ForceHTTP1Upstream = forceHTTP1Upstream

	if backend.Spec.Fallback != nil {
		// set only the secondary priority, the backend defaults to a primary priority if unset.
		if ptr.Deref(backend.Spec.Fallback, false) {
			ds.Priority = new(uint32(1))
		}
	}

	ds.Metadata = buildResourceMetadata(backend, nil)

	return ds
}

// resolveBackendProtocol computes the upstream ir.AppProtocol for a backend from the
// backend's own appProtocol and the route's default (fallback) protocol. It recognizes
// both the Kubernetes Service convention ("kubernetes.io/*") and the Envoy Gateway
// Backend convention ("gateway.envoyproxy.io/*").
//
// backendAppProtocol describes an HTTP-layer protocol (h2c/ws/grpc), so it only refines
// the upstream protocol of HTTP-based routes. For non-HTTP routes (e.g. TCPRoute, TLSRoute
// and UDPRoute, whose default protocol is TCP/UDP), the route is a raw L4 proxy and the
// appProtocol is irrelevant; returning the default avoids emitting HTTP protocol options
// on an L4 cluster.
func resolveBackendProtocol(backendAppProtocol string, defaultProtocol ir.AppProtocol) ir.AppProtocol {
	// The backendAppProtocol is only relevant for HTTP-based routes, so if the default protocol is not HTTP, return the default.
	if defaultProtocol != ir.HTTP {
		return defaultProtocol
	}
	switch {
	case backendAppProtocol == "kubernetes.io/h2c" || backendAppProtocol == string(egv1a1.AppProtocolTypeH2C):
		return ir.HTTP2
	// HTTPRoute can route to gRPC backends, returning ir.GRPC allows the IR to emit gRPC protocol options on the cluster.
	// Kubernetes does not standardize grpc as a Kubernetes appProtocol value, but some projects like Istio uses "grpc" as a convention for gRPC backends, so we recognize it here.
	case backendAppProtocol == "grpc":
		return ir.GRPC
	default:
		return defaultProtocol
	}
}

// shouldForceHTTP1Upstream reports whether the upstream connection should be forced to
// HTTP/1.1. WebSocket over HTTP/2 is not widely supported by upstreams and can lead to
// connection failures, so a WebSocket backend on an HTTP-based route must use HTTP/1.1.
func shouldForceHTTP1Upstream(appProtocol ir.AppProtocol, backendAppProtocol *string) bool {
	return backendAppProtocol != nil && isHTTPProtocol(appProtocol) && isWebSocketAppProtocol(*backendAppProtocol)
}

func isHTTPProtocol(appProtocol ir.AppProtocol) bool {
	switch appProtocol {
	case ir.HTTP, ir.HTTP2, ir.GRPC:
		return true
	default:
		return false
	}
}

// isWebSocketAppProtocol reports whether the given appProtocol denotes WebSocket
// traffic, covering both the Kubernetes Service convention ("kubernetes.io/ws[s]")
// and the Envoy Gateway Backend convention ("gateway.envoyproxy.io/ws[s]").
func isWebSocketAppProtocol(ap string) bool {
	switch ap {
	case "kubernetes.io/ws", "kubernetes.io/wss",
		string(egv1a1.AppProtocolTypeWS), string(egv1a1.AppProtocolTypeWSS):
		return true
	default:
		return false
	}
}

func getStatPattern(routeContext RouteContext, parentRef *RouteParentContext, controllerName string) string {
	var pattern string
	var envoyProxy *egv1a1.EnvoyProxy
	gatewayCtx := GetRouteParentContext(routeContext, *parentRef.ParentReference, controllerName).GetGateway()
	if gatewayCtx != nil {
		envoyProxy = gatewayCtx.envoyProxy
	}
	if envoyProxy != nil && envoyProxy.Spec.Telemetry != nil && envoyProxy.Spec.Telemetry.Metrics != nil &&
		envoyProxy.Spec.Telemetry.Metrics.ClusterStatName != nil {
		pattern = *envoyProxy.Spec.Telemetry.Metrics.ClusterStatName
	}
	return pattern
}

func buildStatName(pattern string, route RouteContext, ruleName *gwapiv1.SectionName, idx int, refs []string) string {
	statName := strings.ReplaceAll(pattern, egv1a1.StatFormatterRouteName, route.GetName())
	statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteNamespace, route.GetNamespace())
	statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteKind, strings.ToLower(route.GetObjectKind().GroupVersionKind().Kind))
	if ruleName == nil {
		statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteRuleName, "-")
	} else {
		statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteRuleName, string(*ruleName))
	}
	statName = strings.ReplaceAll(statName, egv1a1.StatFormatterRouteRuleNumber, fmt.Sprintf("%d", idx))
	statName = strings.ReplaceAll(statName, egv1a1.StatFormatterBackendRefs, strings.Join(refs, "|"))
	return statName
}
