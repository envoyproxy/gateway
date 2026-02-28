// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"net"
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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
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
	ProcessTCPRoutes(tcpRoutes []*gwapiv1a2.TCPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*TCPRouteContext
	ProcessUDPRoutes(udpRoutes []*gwapiv1a2.UDPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*UDPRouteContext
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
		routeRoutes, errs, unacceptedRules := t.processHTTPRouteRules(httpRoute, parentRef, resources)
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
		hasHostnameIntersection := t.processHTTPRouteParentRefListener(httpRoute, routeRoutes, parentRef, xdsIR)
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

func (t *Translator) processHTTPRouteRules(httpRoute *HTTPRouteContext, parentRef *RouteParentContext, resources *resource.Resources) ([]*ir.HTTPRoute, []status.Error, []int) {
	var (
		irRoutes       []*ir.HTTPRoute
		errorCollector = &status.TypedErrorCollector{}
	)
	pattern := getStatPattern(httpRoute, parentRef, t.GatewayControllerName)

	// process each HTTPRouteRule, generate a unique Xds IR HTTPRoute per match of the rule
	unacceptedRules := sets.NewInt()
	for ruleIdx, rule := range httpRoute.Spec.Rules {
		// process HTTP Route filters first, so that the filters can be applied to the IR route later
		var processFilterError error
		httpFiltersContext, errs := t.ProcessHTTPFilters(parentRef, httpRoute, rule.Filters, ruleIdx, resources)
		if len(errs) > 0 {
			for _, err := range errs {
				errorCollector.Add(err)
				processFilterError = errors.Join(processFilterError, err)
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

		// process each backendRef, and calculate the destination settings for this rule
		destName := irRouteDestinationName(httpRoute, ruleIdx)
		allDs := make([]*ir.DestinationSetting, 0, len(rule.BackendRefs))
		var processDestinationError error
		failedNoReadyEndpoints := false
		hasDynamicResolver := false
		backendRefNames := make([]string, len(rule.BackendRefs))
		backendCustomRefs := make([]*ir.UnstructuredRef, 0, len(rule.BackendRefs))
		for i := range rule.BackendRefs {
			settingName := irDestinationSettingName(destName, i)
			backendRefCtx := BackendRefWithFilters{
				BackendRef: &rule.BackendRefs[i].BackendRef,
				Filters:    rule.BackendRefs[i].Filters,
			}
			// ds will never be nil here because processDestination returns an empty DestinationSetting for invalid backendRefs.
			ds, unstructuredRef, err := t.processDestination(settingName, backendRefCtx, parentRef, httpRoute, resources, rule.Name)
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
					ds.Invalid = true
					processDestinationError = err
				}
			}
			if unstructuredRef != nil {
				backendCustomRefs = append(backendCustomRefs, unstructuredRef)
			}
			// skip backendRefs with weight 0 as they do not affect the traffic distribution
			if ds.Weight != nil && *ds.Weight == 0 {
				continue
			}
			allDs = append(allDs, ds)

			// check if there is a dynamic resolver in the backendRefs
			if ds.IsDynamicResolver {
				hasDynamicResolver = true
			}
			backendNamespace := NamespaceDerefOr(rule.BackendRefs[i].Namespace, httpRoute.GetNamespace())
			backendRefNames[i] = fmt.Sprintf("%s/%s", backendNamespace, rule.BackendRefs[i].Name)
		}

		// process each IR route generated for this rule, and set its destination
		destination := &ir.RouteDestination{
			Settings: allDs,
			Metadata: routeRuleMetadata,
		}
		switch {
		// return 500 if any filter processing error occurred
		case processFilterError != nil:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
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
		case processDestinationError != nil && destination.ToBackendWeights().Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
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
		case failedNoReadyEndpoints && destination.ToBackendWeights().Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(503)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 503 direct response in routes due to no ready endpoints",
					"routes", sets.List(routesWithDirectResponse))
			}
		// return 500 if the weight of all the valid destination settings(endpoints list is not empty) is 0
		case destination.ToBackendWeights().Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 500 direct response in routes due to all valid destinations having 0 weight",
					"routes", sets.List(routesWithDirectResponse))
			}
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
					StatusCode: ptr.To(uint32(500)),
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
		default:
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// processing any destinations for this route.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				destination := &ir.RouteDestination{
					Name:     destName,
					Settings: allDs,
					Metadata: routeRuleMetadata,
				}
				irRoute.Destination = destination
			}
		}

		// finalize the IR routes for this rule
		for _, irRoute := range ruleRoutes {
			// add custom backend refs if any
			if len(backendCustomRefs) > 0 {
				irRoute.ExtensionRefs = append(irRoute.ExtensionRefs, backendCustomRefs...)
			}

			// set the stat name for this route
			if irRoute.Destination != nil && pattern != "" {
				irRoute.Destination.StatName = ptr.To(buildStatName(pattern, httpRoute, rule.Name, ruleIdx, backendRefNames))
			}
		}

		irRoutes = append(irRoutes, ruleRoutes...)
	}
	if errorCollector.Empty() {
		return irRoutes, nil, nil
	}

	return irRoutes, errorCollector.GetAllErrors(), unacceptedRules.List()
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
		res.NumRetries = ptr.To(uint32(*retry.Attempts))
	}
	if retry.Backoff != nil {
		backoff, err := time.ParseDuration(string(*retry.Backoff))
		if err == nil {
			res.PerRetry = &ir.PerRetryPolicy{
				BackOff: &ir.BackOffPolicy{
					BaseInterval: ir.MetaV1DurationPtr(backoff),
				},
			}
			// xref: https://gateway-api.sigs.k8s.io/geps/gep-1742/#timeout-values
			if rule.Timeouts != nil && rule.Timeouts.BackendRequest != nil {
				backendRequestTimeout, err := time.ParseDuration(string(*rule.Timeouts.BackendRequest))
				if err == nil {
					res.PerRetry.Timeout = ir.MetaV1DurationPtr(backendRequestTimeout)
				}
			}
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
		if rule.SessionPersistence.IdleTimeout != nil {
			return nil, status.NewRouteStatusError(
				fmt.Errorf("idle timeout is not supported in envoy gateway"),
				status.RouteReasonUnsupportedSetting,
			)
		}

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
					Exact: ptr.To(headerMatch.Value),
				})
			case gwapiv1.HeaderMatchRegularExpression:
				if err := regex.Validate(headerMatch.Value); err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:      string(headerMatch.Name),
					SafeRegex: ptr.To(headerMatch.Value),
				})
			}
		}
		for _, queryParamMatch := range match.QueryParams {
			switch QueryParamMatchTypeDerefOr(queryParamMatch.Type, gwapiv1.QueryParamMatchExact) {
			case gwapiv1.QueryParamMatchExact:
				irRoute.QueryParamMatches = append(irRoute.QueryParamMatches, &ir.StringMatch{
					Name:  string(queryParamMatch.Name),
					Exact: ptr.To(queryParamMatch.Value),
				})
			case gwapiv1.QueryParamMatchRegularExpression:
				if err := regex.Validate(queryParamMatch.Value); err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				irRoute.QueryParamMatches = append(irRoute.QueryParamMatches, &ir.StringMatch{
					Name:      string(queryParamMatch.Name),
					SafeRegex: ptr.To(queryParamMatch.Value),
				})
			}
		}

		if match.Method != nil {
			irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
				Name:  ":method",
				Exact: ptr.To(string(*match.Method)),
			})
		}
		for _, cookieMatch := range match.cookies {
			sm := &ir.StringMatch{
				Name: cookieMatch.Name,
			}
			matchType := egv1a1.CookieMatchExact
			if cookieMatch.Type != nil {
				matchType = *cookieMatch.Type
			}
			switch matchType {
			case egv1a1.CookieMatchExact:
				sm.Exact = ptr.To(cookieMatch.Value)
			case egv1a1.CookieMatchRegularExpression:
				if err := regex.Validate(cookieMatch.Value); err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				sm.SafeRegex = ptr.To(cookieMatch.Value)
			default:
				return nil, status.NewRouteStatusError(
					fmt.Errorf("unsupported cookie match type %q", matchType),
					gwapiv1.RouteReasonUnsupportedValue,
				)
			}
			irRoute.CookieMatches = append(irRoute.CookieMatches, sm)
		}
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
					Exact: ptr.To("OPTIONS"),
				},
				{
					Name:      "origin",
					SafeRegex: ptr.To(".*"),
				},
				{
					Name:      "access-control-request-method",
					SafeRegex: ptr.To(".*"),
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
		routeRoutes, errs, unacceptedRules := t.processGRPCRouteRules(grpcRoute, parentRef, resources)
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
		hasHostnameIntersection := t.processHTTPRouteParentRefListener(grpcRoute, routeRoutes, parentRef, xdsIR)
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

func (t *Translator) processGRPCRouteRules(grpcRoute *GRPCRouteContext, parentRef *RouteParentContext, resources *resource.Resources) ([]*ir.HTTPRoute, []status.Error, []int) {
	var (
		irRoutes       []*ir.HTTPRoute
		errorCollector = &status.TypedErrorCollector{}
	)
	pattern := getStatPattern(grpcRoute, parentRef, t.GatewayControllerName)

	// compute matches, filters, backends
	unacceptedRules := sets.NewInt()
	for ruleIdx := range grpcRoute.Spec.Rules {
		rule := &grpcRoute.Spec.Rules[ruleIdx]

		// process GRPC route filters first, so that the filters can be applied to the IR route later
		var processFilterError error
		httpFiltersContext, errs := t.ProcessGRPCFilters(parentRef, grpcRoute, rule.Filters, resources)
		if len(errs) > 0 {
			for _, err := range errs {
				errorCollector.Add(err)
				processFilterError = errors.Join(processFilterError, err)
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

		// process each backendRef, and calculate the destination settings for this rule
		destName := irRouteDestinationName(grpcRoute, ruleIdx)
		allDs := make([]*ir.DestinationSetting, 0, len(rule.BackendRefs))
		var processDestinationError error
		failedNoReadyEndpoints := false

		backendRefNames := make([]string, len(rule.BackendRefs))
		for i := range rule.BackendRefs {
			settingName := irDestinationSettingName(destName, i)
			backendRefCtx := BackendRefWithFilters{
				BackendRef: &rule.BackendRefs[i].BackendRef,
				Filters:    rule.BackendRefs[i].Filters,
			}
			// ds will never be nil here because processDestination returns an empty DestinationSetting for invalid backendRefs.
			ds, _, err := t.processDestination(settingName, backendRefCtx, parentRef, grpcRoute, resources, rule.Name)
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
			if ds.Weight != nil && *ds.Weight == 0 {
				continue
			}
			allDs = append(allDs, ds)
			backendNamespace := NamespaceDerefOr(rule.BackendRefs[i].Namespace, grpcRoute.GetNamespace())
			backendRefNames[i] = fmt.Sprintf("%s/%s", backendNamespace, rule.BackendRefs[i].Name)
		}

		// process each ir route
		destination := &ir.RouteDestination{
			Settings: allDs,
			Metadata: buildResourceMetadata(grpcRoute, rule.Name),
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
					StatusCode: ptr.To(uint32(500)),
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
		case processDestinationError != nil && destination.ToBackendWeights().Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
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
		case failedNoReadyEndpoints && destination.ToBackendWeights().Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(503)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Info("setting 503 direct response in routes due to no ready endpoints",
					"routes", sets.List(routesWithDirectResponse))
			}
		// return 500 if the weight of all the valid destination settings(endpoints list is not empty) is 0
		case destination.ToBackendWeights().Valid == 0:
			routesWithDirectResponse := sets.New[string]()
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// the direct response from errors.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
				routesWithDirectResponse.Insert(irRoute.Name)
			}
			if len(routesWithDirectResponse) > 0 {
				t.Logger.Error(errors.New("all valid destinations have 0 weight"), "setting 500 direct response in routes due to all valid destinations having 0 weight",
					"routes", sets.List(routesWithDirectResponse))
			}
		default:
			for _, irRoute := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// processing any destinations for this route.
				if irRoute.DirectResponse != nil || irRoute.Redirect != nil {
					continue
				}
				destination := &ir.RouteDestination{
					Name:     destName,
					Settings: allDs,
					Metadata: buildResourceMetadata(grpcRoute, rule.Name),
				}
				irRoute.Destination = destination
			}

		}

		// finalize the IR routes for this rule
		for _, irRoute := range ruleRoutes {
			irRoute.IsHTTP2 = true

			// set the stat name for this route
			if irRoute.Destination != nil && pattern != "" {
				irRoute.Destination.StatName = ptr.To(buildStatName(pattern, grpcRoute, rule.Name, ruleIdx, backendRefNames))
			}
		}

		irRoutes = append(irRoutes, ruleRoutes...)
	}

	if errorCollector.Empty() {
		return irRoutes, nil, nil
	}

	return irRoutes, errorCollector.GetAllErrors(), unacceptedRules.List()
}

func (t *Translator) processGRPCRouteRule(grpcRoute *GRPCRouteContext, ruleIdx int, httpFiltersContext *HTTPFiltersContext, rule *gwapiv1.GRPCRouteRule) ([]*ir.HTTPRoute, status.Error) {
	capacity := len(rule.Matches)
	if capacity == 0 {
		capacity = 1
	}
	ruleRoutes := make([]*ir.HTTPRoute, 0, capacity)

	// If no matches are specified, the implementation MUST match every gRPC request.
	if len(rule.Matches) == 0 {
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
	for matchIdx, match := range rule.Matches {
		irRoute := &ir.HTTPRoute{
			Name: irRouteName(grpcRoute, ruleIdx, matchIdx),
		}
		irRoute.Metadata = buildResourceMetadata(grpcRoute, rule.Name)
		for _, headerMatch := range match.Headers {
			switch GRPCHeaderMatchTypeDerefOr(headerMatch.Type, gwapiv1.GRPCHeaderMatchExact) {
			case gwapiv1.GRPCHeaderMatchExact:
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:  string(headerMatch.Name),
					Exact: ptr.To(headerMatch.Value),
				})
			case gwapiv1.GRPCHeaderMatchRegularExpression:
				if err := regex.Validate(headerMatch.Value); err != nil {
					return nil, status.NewRouteStatusError(err, gwapiv1.RouteReasonUnsupportedValue)
				}
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:      string(headerMatch.Name),
					SafeRegex: ptr.To(headerMatch.Value),
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

		ruleRoutes = append(ruleRoutes, irRoute)
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
	}
	return ruleRoutes, nil
}

func (t *Translator) processGRPCRouteMethodExact(method *gwapiv1.GRPCMethodMatch, irRoute *ir.HTTPRoute) {
	switch {
	case method.Service != nil && method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			Exact: ptr.To(fmt.Sprintf("/%s/%s", *method.Service, *method.Method)),
		}
	case method.Method != nil:
		// Use a header match since the PathMatch doesn't support Suffix matching
		irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
			Name:   ":path",
			Suffix: ptr.To(fmt.Sprintf("/%s", *method.Method)),
		})
	case method.Service != nil:
		irRoute.PathMatch = &ir.StringMatch{
			Prefix: ptr.To(fmt.Sprintf("/%s", *method.Service)),
		}
	}
}

func (t *Translator) processGRPCRouteMethodRegularExpression(method *gwapiv1.GRPCMethodMatch, irRoute *ir.HTTPRoute) {
	switch {
	case method.Service != nil && method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: ptr.To(fmt.Sprintf("/%s/%s", *method.Service, *method.Method)),
		}
	case method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: ptr.To(fmt.Sprintf("/%s/%s", validServiceName, *method.Method)),
		}
	case method.Service != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: ptr.To(fmt.Sprintf("/%s/%s", *method.Service, validMethodName)),
		}
	}
}

func (t *Translator) processHTTPRouteParentRefListener(route RouteContext, routeRoutes []*ir.HTTPRoute, parentRef *RouteParentContext, xdsIR resource.XdsIRMap) bool {
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

		var perHostRoutes []*ir.HTTPRoute
		for _, host := range hosts {
			for _, routeRoute := range routeRoutes {
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
				hostRoute := routeRoute.DeepCopy()
				hostRoute.Name = fmt.Sprintf("%s/%s", routeRoute.Name, underscoredHost)
				hostRoute.Hostname = host
				perHostRoutes = append(perHostRoutes, hostRoute)
			}
		}
		irKey := t.getIRKey(listener.gateway.Gateway)
		irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))

		if irListener != nil {
			if route.GetRouteType() == resource.KindGRPCRoute {
				irListener.IsHTTP2 = true
			}
			irListener.Routes = append(irListener.Routes, perHostRoutes...)
		}
	}

	return hasHostnameIntersection
}

func buildResourceMetadata(resource client.Object, sectionName *gwapiv1.SectionName) *ir.ResourceMetadata {
	metadata := &ir.ResourceMetadata{
		Kind:        resource.GetObjectKind().GroupVersionKind().Kind,
		Name:        resource.GetName(),
		Namespace:   resource.GetNamespace(),
		Annotations: ir.MapToSlice(filterEGPrefix(resource.GetAnnotations())),
	}
	if sectionName != nil {
		metadata.SectionName = string(*sectionName)
	}
	return metadata
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
			destSettings []*ir.DestinationSetting
			resolveErrs  = &status.MultiStatusError{}
			destName     = irRouteDestinationName(tlsRoute, -1 /*rule index*/)
		)

		// compute backends
		for _, rule := range tlsRoute.Spec.Rules {
			for i := range rule.BackendRefs {
				settingName := irDestinationSettingName(destName, i)
				backendRefCtx := DirectBackendRef{BackendRef: &rule.BackendRefs[i]}
				// ds will never be nil here because processDestination returns an empty DestinationSetting for invalid backendRefs.
				ds, _, err := t.processDestination(settingName, backendRefCtx, parentRef, tlsRoute, resources, rule.Name)
				if err != nil {
					resolveErrs.Add(err)
					continue
				}
				// skip backendRefs with weight 0 as they do not affect the traffic distribution
				if ds.Weight != nil && *ds.Weight > 0 {
					destSettings = append(destSettings, ds)
				}
			}

			// TODO handle:
			//	- no valid backend refs
			//	- sum of weights for valid backend refs is 0
			//	- returning 500's for invalid backend refs
			//	- etc.
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
					// Listener is in terminate mode.
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
					Destination: &ir.RouteDestination{
						Name:     destName,
						Settings: destSettings,
						Metadata: buildResourceMetadata(tlsRoute, nil),
					},
					Metadata: buildResourceMetadata(tlsRoute, nil),
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

func (t *Translator) ProcessUDPRoutes(udpRoutes []*gwapiv1a2.UDPRoute, gateways []*GatewayContext, resources *resource.Resources,
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
		var (
			destSettings []*ir.DestinationSetting
			resolveErrs  = &status.MultiStatusError{}
			destName     = irRouteDestinationName(udpRoute, -1 /*rule index*/)
		)

		for i := range udpRoute.Spec.Rules[0].BackendRefs {
			settingName := irDestinationSettingName(destName, i)
			backendRefCtx := DirectBackendRef{BackendRef: &udpRoute.Spec.Rules[0].BackendRefs[i]}
			// ds will never be nil here because processDestination returns an empty DestinationSetting for invalid backendRefs.
			ds, _, err := t.processDestination(settingName, backendRefCtx, parentRef, udpRoute, resources, udpRoute.Spec.Rules[0].Name)
			if err != nil {
				resolveErrs.Add(err)
				continue
			}

			// skip backendRefs with weight 0 as they do not affect the traffic distribution
			if ds.Weight != nil && *ds.Weight > 0 {
				destSettings = append(destSettings, ds)
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
			// only one route is allowed for a UDP listener
			if listener.AttachedRoutes() >= 1 {
				continue
			}
			listener.IncrementAttachedRoutes()
			if !listener.IsReady() {
				continue
			}
			accepted = true

			irKey := t.getIRKey(listener.gateway.Gateway)

			gwXdsIR := xdsIR[irKey]
			irListener := gwXdsIR.GetUDPListener(irListenerName(listener))
			if irListener != nil {
				irRoute := &ir.UDPRoute{
					Name: irUDPRouteName(udpRoute),
					Destination: &ir.RouteDestination{
						Name:     destName,
						Settings: destSettings,
						// udpRoute Must have a single rule, so can use index 0.
						Metadata: buildResourceMetadata(udpRoute, udpRoute.Spec.Rules[0].Name),
					},
				}
				irListener.Route = irRoute
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

func (t *Translator) ProcessTCPRoutes(tcpRoutes []*gwapiv1a2.TCPRoute, gateways []*GatewayContext, resources *resource.Resources,
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
			destSettings []*ir.DestinationSetting
			resolveErrs  = &status.MultiStatusError{}
			destName     = irRouteDestinationName(tcpRoute, -1 /*rule index*/)
		)

		for i := range tcpRoute.Spec.Rules[0].BackendRefs {
			settingName := irDestinationSettingName(destName, i)
			backendRefCtx := DirectBackendRef{BackendRef: &tcpRoute.Spec.Rules[0].BackendRefs[i]}
			ds, _, err := t.processDestination(settingName, backendRefCtx, parentRef, tcpRoute, resources, tcpRoute.Spec.Rules[0].Name)
			// skip adding the route and provide the reason via route status.
			if err != nil {
				resolveErrs.Add(err)
				continue
			}
			// skip backendRefs with weight 0 as they do not affect the traffic distribution
			if ds.Weight != nil && *ds.Weight > 0 {
				destSettings = append(destSettings, ds)
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
			// only one route is allowed for a TCP listener
			if listener.AttachedRoutes() >= 1 {
				continue
			}
			if !listener.IsReady() {
				continue
			}
			listener.IncrementAttachedRoutes()

			accepted = true
			irKey := t.getIRKey(listener.gateway.Gateway)

			gwXdsIR := xdsIR[irKey]
			irListener := gwXdsIR.GetTCPListener(irListenerName(listener))
			if irListener != nil {
				irRoute := &ir.TCPRoute{
					Name: irTCPRouteName(tcpRoute),
					Destination: &ir.RouteDestination{
						Name:     destName,
						Settings: destSettings,
						Metadata: buildResourceMetadata(tcpRoute, tcpRoute.Spec.Rules[0].Name),
					},
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
	routeRuleName *gwapiv1.SectionName,
) (ds *ir.DestinationSetting, unstructuredRef *ir.UnstructuredRef, err status.Error) {
	var (
		routeType  = route.GetRouteType()
		weight     = (uint32(ptr.Deref(backendRefContext.GetBackendRef().Weight, int32(1))))
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
	}

	// Skip processing backends with 0 weight
	if weight == 0 {
		return emptyDS, nil, nil
	}

	var envoyProxy *egv1a1.EnvoyProxy
	gatewayCtx := GetRouteParentContext(route, *parentRef.ParentReference, t.GatewayControllerName).GetGateway()
	if gatewayCtx != nil {
		envoyProxy = gatewayCtx.envoyProxy
	}

	// Resolve BTP RoutingType for this route/gateway combination
	var btpRoutingType *egv1a1.RoutingType
	if gatewayCtx != nil {
		btpRoutingType = t.BTPRoutingTypeIndex.LookupBTPRoutingType(
			route.GetRouteType(),
			types.NamespacedName{Namespace: route.GetNamespace(), Name: route.GetName()},
			types.NamespacedName{Namespace: gatewayCtx.GetNamespace(), Name: gatewayCtx.GetName()},
			parentRef.SectionName,
			routeRuleName,
		)
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
		envoyProxy,
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
		ds, err = t.processServiceDestinationSetting(name, backendRef.BackendObjectReference, backendNamespace, protocol, envoyProxy, btpRoutingType)
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
	ds.Filters, filtersErr = t.processDestinationFilters(routeType, backendRefContext, parentRef, route, resources)
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

func (t *Translator) processServiceImportDestinationSetting(
	name string,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	protocol ir.AppProtocol,
	envoyProxy *egv1a1.EnvoyProxy,
	btpRoutingType *egv1a1.RoutingType,
) (*ir.DestinationSetting, status.Error) {
	var (
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
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
		protocol = serviceAppProtocolToIRAppProtocol(*servicePort.AppProtocol, protocol, false)
	}

	backendIps := serviceImport.Spec.IPs
	isHeadless := len(backendIps) == 0

	// Route to endpoints by default, or if service routing is enabled but ServiceImport is headless
	useEndpointRouting := !t.IsServiceRouting(envoyProxy, btpRoutingType) || isHeadless
	if useEndpointRouting {
		endpointSlices := t.GetEndpointSlicesForBackend(backendNamespace, string(backendRef.Name), resource.KindServiceImport)
		endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, servicePort.Name, getServicePortProtocol(servicePort.Protocol))
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
		Name:        name,
		Protocol:    protocol,
		Endpoints:   endpoints,
		AddressType: addrType,
		Metadata:    buildResourceMetadata(serviceImport, ptr.To(gwapiv1.SectionName(strconv.Itoa(int(*backendRef.Port))))),
	}, nil
}

func (t *Translator) processServiceDestinationSetting(
	name string,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	protocol ir.AppProtocol,
	envoyProxy *egv1a1.EnvoyProxy,
	btpRoutingType *egv1a1.RoutingType,
) (*ir.DestinationSetting, status.Error) {
	var (
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
	)

	service := t.GetService(backendNamespace, string(backendRef.Name))
	var servicePort corev1.ServicePort
	for _, port := range service.Spec.Ports {
		if port.Port == *backendRef.Port {
			servicePort = port
			break
		}
	}

	// support HTTPRouteBackendProtocolH2C/GRPC
	if servicePort.AppProtocol != nil {
		protocol = serviceAppProtocolToIRAppProtocol(*servicePort.AppProtocol, protocol, true)
	}

	isHeadless := isServiceHeadless(service)

	// Route to endpoints by default, or if service routing is enabled but service is headless
	useEndpointRouting := !t.IsServiceRouting(envoyProxy, btpRoutingType) || isHeadless
	if useEndpointRouting {
		endpointSlices := t.GetEndpointSlicesForBackend(backendNamespace, string(backendRef.Name), KindDerefOr(backendRef.Kind, resource.KindService))
		endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, servicePort.Name, getServicePortProtocol(servicePort.Protocol))
		if len(endpoints) == 0 {
			return nil, status.NewRouteStatusError(
				fmt.Errorf("no ready endpoints for the related Service %s/%s", backendNamespace, backendRef.Name),
				status.RouteReasonEndpointsNotFound,
			)
		}
	} else {
		// Use Service ClusterIP routing
		ep := ir.NewDestEndpoint(nil, service.Spec.ClusterIP, uint32(*backendRef.Port), false, nil)
		endpoints = append(endpoints, ep)
	}

	return &ir.DestinationSetting{
		Name:        name,
		Protocol:    protocol,
		Endpoints:   endpoints,
		AddressType: addrType,
		PreferLocal: processPreferLocalZone(service),
		Metadata:    buildResourceMetadata(service, ptr.To(gwapiv1.SectionName(strconv.Itoa(int(*backendRef.Port))))),
	}, nil
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
			MinEndpointsThreshold: ptr.To[uint64](1),
			Force: &ir.ForceLocalZone{
				MinEndpointsInZoneThreshold: ptr.To[uint32](1),
			},
		}
	}

	// Allows annotation values that align with Kubernetes defaults.
	// Ref:
	// https://kubernetes.io/docs/concepts/services-networking/topology-aware-routing/#enabling-topology-aware-routing
	// https://github.com/kubernetes/kubernetes/blob/9d9e1afdf78bce0a517cc22557457f942040ca19/staging/src/k8s.io/endpointslice/utils.go#L355-L368
	if val, ok := svc.Annotations[corev1.AnnotationTopologyMode]; ok && val == "Auto" || val == "auto" {
		return &ir.PreferLocalZone{
			MinEndpointsThreshold: ptr.To[uint64](3),
			Force: &ir.ForceLocalZone{
				MinEndpointsInZoneThreshold: ptr.To[uint32](3),
			},
		}
	}

	return nil
}

func (t *Translator) processDestinationFilters(routeType gwapiv1.Kind, backendRefContext BackendRefContext, parentRef *RouteParentContext, route RouteContext, resources *resource.Resources) (*ir.DestinationFilters, error) {
	backendFilters := getBackendFilters(routeType, backendRefContext)
	if backendFilters == nil {
		return nil, nil
	}

	var httpFiltersContext *HTTPFiltersContext
	var destFilters ir.DestinationFilters

	var errs []status.Error
	switch filters := backendFilters.(type) {
	case []gwapiv1.HTTPRouteFilter:
		httpFiltersContext, errs = t.ProcessHTTPFilters(parentRef, route, filters, 0, resources)
	case []gwapiv1.GRPCRouteFilter:
		httpFiltersContext, errs = t.ProcessGRPCFilters(parentRef, route, filters, resources)
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
		return ir.HTTPS
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

		if !HasReadyListener(selectedListeners) {
			routeStatus := GetRouteStatus(routeContext)
			status.SetRouteStatusCondition(routeStatus,
				parentRefCtx.routeParentStatusIdx,
				routeContext.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				"NoReadyListeners",
				"There are no ready listeners for this parent ref",
			)
			continue
		}

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

func getIREndpointsFromEndpointSlices(endpointSlices []*discoveryv1.EndpointSlice, portName string, portProtocol corev1.Protocol) ([]*ir.DestinationEndpoint, *ir.DestinationAddressType) {
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
		endpoints := getIREndpointsFromEndpointSlice(endpointSlice, portName, portProtocol)
		dstEndpoints = append(dstEndpoints, endpoints...)
	}

	for addrTypeState, addrTypeCounts := range addrTypeMap {
		if addrTypeCounts == len(endpointSlices) {
			dstAddrType = ptr.To(addrTypeState)
			break
		}
	}

	if len(addrTypeMap) > 0 && dstAddrType == nil {
		dstAddrType = ptr.To(ir.MIXED)
	}

	return dstEndpoints, dstAddrType
}

func getIREndpointsFromEndpointSlice(endpointSlice *discoveryv1.EndpointSlice, portName string, portProtocol corev1.Protocol) []*ir.DestinationEndpoint {
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
				ep := ir.NewDestEndpoint(nil, address, uint32(*endpointPort.Port), draining, endpoint.Zone)
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
	protocol ir.AppProtocol,
) *ir.DestinationSetting {
	var dstAddrType *ir.DestinationAddressType

	addrTypeMap := make(map[ir.DestinationAddressType]int)
	backend := t.GetBackend(backendNamespace, string(backendRef.Name))
	for _, ap := range backend.Spec.AppProtocols {
		protocol = backendAppProtocolToIRAppProtocol(ap, protocol)
	}

	ds := &ir.DestinationSetting{Name: name}

	// There is only one backend if it is a dynamic resolver
	if backend.Spec.Type != nil && *backend.Spec.Type == egv1a1.BackendTypeDynamicResolver {
		ds.IsDynamicResolver = true
		ds.Protocol = protocol
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
				Path: ptr.To(bep.Unix.Path),
				Zone: bep.Zone,
			}
		}

		dstEndpoints = append(dstEndpoints, irde)
	}

	for addrTypeState, addrTypeCounts := range addrTypeMap {
		if addrTypeCounts == len(backend.Spec.Endpoints) {
			dstAddrType = ptr.To(addrTypeState)
			break
		}
	}

	if len(addrTypeMap) > 0 && dstAddrType == nil {
		dstAddrType = ptr.To(ir.MIXED)
	}

	ds.Endpoints = dstEndpoints
	ds.AddressType = dstAddrType
	ds.Protocol = protocol

	if backend.Spec.Fallback != nil {
		// set only the secondary priority, the backend defaults to a primary priority if unset.
		if ptr.Deref(backend.Spec.Fallback, false) {
			ds.Priority = ptr.To(uint32(1))
		}
	}

	ds.Metadata = buildResourceMetadata(backend, nil)

	return ds
}

// serviceAppProtocolToIRAppProtocol translates the appProtocol string into an ir.AppProtocol.
//
// When grpcCompatibility is enabled, `grpc` will be parsed as a valid option for HTTP2.
// See https://github.com/envoyproxy/gateway/issues/5485#issuecomment-2731322578.
func serviceAppProtocolToIRAppProtocol(ap string, defaultProtocol ir.AppProtocol, grpcCompatibility bool) ir.AppProtocol {
	switch {
	case ap == "kubernetes.io/h2c":
		return ir.HTTP2
	case ap == "grpc" && grpcCompatibility:
		return ir.GRPC
	default:
		return defaultProtocol
	}
}

func backendAppProtocolToIRAppProtocol(ap egv1a1.AppProtocolType, defaultProtocol ir.AppProtocol) ir.AppProtocol {
	switch ap {
	case egv1a1.AppProtocolTypeH2C:
		return ir.HTTP2
	case "grpc":
		return ir.GRPC
	default:
		return defaultProtocol
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
