// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
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
	ProcessTLSRoutes(tlsRoutes []*gwapiv1a2.TLSRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*TLSRouteContext
	ProcessTCPRoutes(tcpRoutes []*gwapiv1a2.TCPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*TCPRouteContext
	ProcessUDPRoutes(udpRoutes []*gwapiv1a2.UDPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*UDPRouteContext
}

func (t *Translator) ProcessHTTPRoutes(httpRoutes []*gwapiv1.HTTPRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*HTTPRouteContext {
	var relevantHTTPRoutes []*HTTPRouteContext

	// always sort initially by creation time stamp. Later on, additional sorting based on matcher type and
	// match length may occur.
	sort.Slice(httpRoutes, func(i, j int) bool {
		return httpRoutes[i].CreationTimestamp.Before(&(httpRoutes[j].CreationTimestamp))
	})

	for _, h := range httpRoutes {
		if h == nil {
			panic("received nil httproute")
		}
		httpRoute := &HTTPRouteContext{
			GatewayControllerName: t.GatewayControllerName,
			HTTPRoute:             h.DeepCopy(),
		}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(httpRoute, gateways, resources)
		if !relevantRoute {
			continue
		}

		relevantHTTPRoutes = append(relevantHTTPRoutes, httpRoute)

		t.processHTTPRouteParentRefs(httpRoute, resources, xdsIR)
	}

	return relevantHTTPRoutes
}

func (t *Translator) ProcessGRPCRoutes(grpcRoutes []*gwapiv1.GRPCRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*GRPCRouteContext {
	var relevantGRPCRoutes []*GRPCRouteContext

	for _, g := range grpcRoutes {
		if g == nil {
			panic("received nil grpcroute")
		}
		grpcRoute := &GRPCRouteContext{
			GatewayControllerName: t.GatewayControllerName,
			GRPCRoute:             g.DeepCopy(),
		}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(grpcRoute, gateways, resources)
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
		routeRoutes, err := t.processHTTPRouteRules(httpRoute, parentRef, resources)
		// TODO: zhaohuabing: according to the gateway api, the RouteConditionPartiallyInvalid condition should be set
		// to true when an HTTPRoute contains a combination of both valid and invalid rules.
		if err != nil {
			routeStatus := GetRouteStatus(httpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				httpRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				err.Reason(),
				status.Error2ConditionMsg(err),
			)
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

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(httpRoute, gwapiv1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
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

func (t *Translator) processHTTPRouteRules(httpRoute *HTTPRouteContext, parentRef *RouteParentContext, resources *resource.Resources) ([]*ir.HTTPRoute, StatusError) {
	var (
		irRoutes   []*ir.HTTPRoute
		envoyProxy *egv1a1.EnvoyProxy
		errs       = &MultiStatusError{errs: make([]StatusError, 0)}
	)

	gatewayCtx := httpRoute.ParentRefs[*parentRef.ParentReference].GetGateway()
	if gatewayCtx != nil {
		envoyProxy = gatewayCtx.envoyProxy
	}

	// process each HTTPRouteRule, generate a unique Xds IR HTTPRoute per match of the rule
	for ruleIdx, rule := range httpRoute.Spec.Rules {
		httpFiltersContext, err := t.ProcessHTTPFilters(parentRef, httpRoute, rule.Filters, ruleIdx, resources)
		if err != nil {
			errs.Add(&RouteStatusError{
				Wrapped:              fmt.Errorf("failed to process route rule %d: %w", ruleIdx, err),
				RouteConditionReason: err.Reason(),
			})
			continue
		}

		// The HTTPRouteRule matches are ORed, a rule is matched if any one of its matches is satisfied,
		// so generate a unique Xds IR HTTPRoute per match.
		ruleRoutes, err := t.processHTTPRouteRule(httpRoute, ruleIdx, httpFiltersContext, rule)
		if err != nil {
			errs.Add(&RouteStatusError{
				Wrapped:              fmt.Errorf("failed to process route rule %d: %w", ruleIdx, err),
				RouteConditionReason: err.Reason(),
			})
			continue
		}

		dstAddrTypeSet := make(sets.Set[ir.DestinationAddressType])
		destName := irRouteDestinationName(httpRoute, ruleIdx)
		allDs := []*ir.DestinationSetting{}
		failedProcessDestination := false
		hasDynamicResolver := false

		// process each backendRef, and calculate the destination settings for this rule
		for i, backendRef := range rule.BackendRefs {
			settingName := irDestinationSettingName(destName, i)
			ds, err := t.processDestination(settingName, backendRef, parentRef, httpRoute, resources)
			if err != nil {
				errs.Add(&RouteStatusError{
					Wrapped:              fmt.Errorf("failed to process route rule %d backendRef %d: %w", ruleIdx, i, err),
					RouteConditionReason: err.Reason(),
				})
				failedProcessDestination = true
				continue
			}

			// ds can be nil if the backendRef weight is 0
			if ds == nil {
				continue
			}
			allDs = append(allDs, ds)

			// check if there is a dynamic resolver in the backendRefs
			if ds.IsDynamicResolver {
				hasDynamicResolver = true
			}

			if !t.IsEnvoyServiceRouting(envoyProxy) && len(ds.Endpoints) > 0 && ds.AddressType != nil {
				dstAddrTypeSet.Insert(*ds.AddressType)
			}
		}

		// noValidBackendRef := false
		// process each ir route
		for _, irRoute := range ruleRoutes {
			destination := &ir.RouteDestination{
				Settings: allDs,
			}

			switch {
			// If the route already has a direct response or redirect configured, then it was from a filter so skip
			// processing any destinations for this route.
			case irRoute.DirectResponse != nil || irRoute.Redirect != nil:
			// return 500 if any destination setting is invalid
			// the error is already added to the error list when processing the destination
			case failedProcessDestination:
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
			// return 500 if the weight of all the valid destination settings(endpoints list is not empty) is 0
			case !hasDynamicResolver && destination.ToBackendWeights().Valid == 0:
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
			// A route can only have one destination if this destination is a dynamic resolver, because the behavior of
			// multiple destinations with one being a dynamic resolver just doesn't make sense.
			case hasDynamicResolver && len(rule.BackendRefs) > 1:
				irRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
			default:
				destination.Name = destName
				destination.Settings = allDs
				irRoute.Destination = destination
			}
		}

		if hasDynamicResolver && len(rule.BackendRefs) > 1 {
			errs.Add(&RouteStatusError{
				Wrapped:              fmt.Errorf("failed to process route rule %d: dynamic resolver is not supported for multiple backendRefs", ruleIdx),
				RouteConditionReason: RouteReasonInvalidBackendRef,
			})
		}

		// TODO: support mixed endpointslice address type between backendRefs
		if !t.IsEnvoyServiceRouting(envoyProxy) && (dstAddrTypeSet.Len() > 1 || dstAddrTypeSet.Has(ir.MIXED)) {
			errs.Add(&RouteStatusError{
				Wrapped:              fmt.Errorf("failed to process route rule %d: mixed endpointslice address type between backendRefs is not supported", ruleIdx),
				RouteConditionReason: RouteReasonInvalidBackendRef,
			})
		}

		// TODO handle:
		//	- sum of weights for valid backend refs is 0
		//	- etc.

		irRoutes = append(irRoutes, ruleRoutes...)
	}
	if errs.Empty() {
		return irRoutes, nil
	}
	return irRoutes, errs
}

func processRouteTrafficFeatures(irRoute *ir.HTTPRoute, rule gwapiv1.HTTPRouteRule) {
	processRouteTimeout(irRoute, rule)
	processRouteRetry(irRoute, rule)
}

func processRouteTimeout(irRoute *ir.HTTPRoute, rule gwapiv1.HTTPRouteRule) {
	if rule.Timeouts != nil {
		if rule.Timeouts.Request != nil {
			d, err := time.ParseDuration(string(*rule.Timeouts.Request))
			if err != nil {
				d, _ = time.ParseDuration(HTTPRequestTimeout)
			}
			irRoute.Timeout = ptr.To(metav1.Duration{Duration: d})
		}

		// Also set the IR Route Timeout to the backend request timeout
		// until we introduce retries, then set it to per try timeout
		if rule.Timeouts.BackendRequest != nil {
			d, err := time.ParseDuration(string(*rule.Timeouts.BackendRequest))
			if err != nil {
				d, _ = time.ParseDuration(HTTPRequestTimeout)
			}
			irRoute.Timeout = ptr.To(metav1.Duration{Duration: d})
		}
	}
}

func processRouteRetry(irRoute *ir.HTTPRoute, rule gwapiv1.HTTPRouteRule) {
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
					BaseInterval: ptr.To(metav1.Duration{Duration: backoff}),
				},
			}
			// xref: https://gateway-api.sigs.k8s.io/geps/gep-1742/#timeout-values
			if rule.Timeouts != nil && rule.Timeouts.BackendRequest != nil {
				backendRequestTimeout, err := time.ParseDuration(string(*rule.Timeouts.BackendRequest))
				if err == nil {
					res.PerRetry.Timeout = &metav1.Duration{Duration: backendRequestTimeout}
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

func (t *Translator) processHTTPRouteRule(httpRoute *HTTPRouteContext, ruleIdx int, httpFiltersContext *HTTPFiltersContext, rule gwapiv1.HTTPRouteRule) ([]*ir.HTTPRoute, StatusError) {
	var ruleRoutes []*ir.HTTPRoute

	// If no matches are specified, the implementation MUST match every HTTP request.
	if len(rule.Matches) == 0 {
		irRoute := &ir.HTTPRoute{
			Name: irRouteName(httpRoute, ruleIdx, -1),
		}
		irRoute.Metadata = buildRouteMetadata(httpRoute, rule.Name)
		processRouteTrafficFeatures(irRoute, rule)
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
		ruleRoutes = append(ruleRoutes, irRoute)
	}

	var sessionPersistence *ir.SessionPersistence
	if rule.SessionPersistence != nil {
		if rule.SessionPersistence.IdleTimeout != nil {
			return nil,
				&RouteStatusError{
					Wrapped:              fmt.Errorf("idle timeout is not supported in envoy gateway"),
					RouteConditionReason: RouteReasonUnsupportedSetting,
				}
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
					return nil, &RouteStatusError{
						Wrapped:              err,
						RouteConditionReason: gwapiv1.RouteReasonUnsupportedValue,
					}
				}
				sessionPersistence.Cookie.TTL = &metav1.Duration{Duration: ttl}
			}
		case *rule.SessionPersistence.Type == gwapiv1.HeaderBasedSessionPersistence:
			sessionPersistence = &ir.SessionPersistence{
				Header: &ir.HeaderBasedSessionPersistence{
					Name: sessionName,
				},
			}
		default:
			// Unknown session persistence type is specified.
			return nil, &RouteStatusError{
				Wrapped:              fmt.Errorf("unknown session persistence type %s", *rule.SessionPersistence.Type),
				RouteConditionReason: gwapiv1.RouteReasonUnsupportedValue,
			}
		}
	}

	// A rule is matched if any one of its matches
	// is satisfied (i.e. a logical "OR"), so generate
	// a unique Xds IR HTTPRoute per match.
	for matchIdx, match := range rule.Matches {
		irRoute := &ir.HTTPRoute{
			Name:               irRouteName(httpRoute, ruleIdx, matchIdx),
			SessionPersistence: sessionPersistence,
		}
		irRoute.Metadata = buildRouteMetadata(httpRoute, rule.Name)
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
					return nil, &RouteStatusError{
						Wrapped:              err,
						RouteConditionReason: gwapiv1.RouteReasonUnsupportedValue,
					}
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
					return nil, &RouteStatusError{
						Wrapped:              err,
						RouteConditionReason: gwapiv1.RouteReasonUnsupportedValue,
					}
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
					return nil, &RouteStatusError{
						Wrapped:              err,
						RouteConditionReason: gwapiv1.RouteReasonUnsupportedValue,
					}
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
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
		ruleRoutes = append(ruleRoutes, irRoute)
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

	if len(httpFiltersContext.ExtensionRefs) > 0 {
		irRoute.ExtensionRefs = httpFiltersContext.ExtensionRefs
	}
}

func (t *Translator) processGRPCRouteParentRefs(grpcRoute *GRPCRouteContext, resources *resource.Resources, xdsIR resource.XdsIRMap) {
	for _, parentRef := range grpcRoute.ParentRefs {

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		routeRoutes, err := t.processGRPCRouteRules(grpcRoute, parentRef, resources)
		if err != nil {
			routeStatus := GetRouteStatus(grpcRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				grpcRoute.GetGeneration(),
				gwapiv1.RouteConditionAccepted,
				metav1.ConditionFalse,
				gwapiv1.RouteReasonUnsupportedValue, // TODO: better reason
				status.Error2ConditionMsg(err),
			)
			continue
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

func (t *Translator) processGRPCRouteRules(grpcRoute *GRPCRouteContext, parentRef *RouteParentContext, resources *resource.Resources) ([]*ir.HTTPRoute, error) {
	var routeRoutes []*ir.HTTPRoute

	// compute matches, filters, backends
	for ruleIdx, rule := range grpcRoute.Spec.Rules {
		httpFiltersContext, err := t.ProcessGRPCFilters(parentRef, grpcRoute, rule.Filters, resources)
		if err != nil {
			return nil, err
		}
		// A rule is matched if any one of its matches
		// is satisfied (i.e. a logical "OR"), so generate
		// a unique Xds IR HTTPRoute per match.
		ruleRoutes, err := t.processGRPCRouteRule(grpcRoute, ruleIdx, httpFiltersContext, rule)
		if err != nil {
			return nil, err
		}

		destName := irRouteDestinationName(grpcRoute, ruleIdx)
		for i, backendRef := range rule.BackendRefs {
			settingName := irDestinationSettingName(destName, i)
			ds, err := t.processDestination(settingName, backendRef, parentRef, grpcRoute, resources)
			if ds == nil {
				continue
			}

			for _, route := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// processing any destinations for this route.
				if route.DirectResponse != nil || route.Redirect != nil {
					continue
				}

				//  disable associated routes to a backend ref in case some of its config was invalid
				if err != nil {
					route.DirectResponse = &ir.CustomResponse{
						StatusCode: ptr.To(uint32(500)),
					}
				}

				if route.Destination == nil {
					route.Destination = &ir.RouteDestination{
						Name: destName,
					}
				}
				route.Destination.Settings = append(route.Destination.Settings, ds)
			}
		}

		// If the route has no valid backends then just use a direct response and don't fuss with weighted responses
		for _, ruleRoute := range ruleRoutes {
			noValidBackends := ruleRoute.Destination == nil || ruleRoute.Destination.ToBackendWeights().Valid == 0
			if noValidBackends && ruleRoute.Redirect == nil {
				ruleRoute.DirectResponse = &ir.CustomResponse{
					StatusCode: ptr.To(uint32(500)),
				}
			}
			ruleRoute.IsHTTP2 = true
		}

		// TODO handle:
		//	- sum of weights for valid backend refs is 0
		//	- etc.

		routeRoutes = append(routeRoutes, ruleRoutes...)
	}

	return routeRoutes, nil
}

func (t *Translator) processGRPCRouteRule(grpcRoute *GRPCRouteContext, ruleIdx int, httpFiltersContext *HTTPFiltersContext, rule gwapiv1.GRPCRouteRule) ([]*ir.HTTPRoute, error) {
	var ruleRoutes []*ir.HTTPRoute

	// If no matches are specified, the implementation MUST match every gRPC request.
	if len(rule.Matches) == 0 {
		irRoute := &ir.HTTPRoute{
			Name: irRouteName(grpcRoute, ruleIdx, -1),
		}
		irRoute.Metadata = buildRouteMetadata(grpcRoute, rule.Name)
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
		irRoute.Metadata = buildRouteMetadata(grpcRoute, rule.Name)
		for _, headerMatch := range match.Headers {
			switch GRPCHeaderMatchTypeDerefOr(headerMatch.Type, gwapiv1.GRPCHeaderMatchExact) {
			case gwapiv1.GRPCHeaderMatchExact:
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:  string(headerMatch.Name),
					Exact: ptr.To(headerMatch.Value),
				})
			case gwapiv1.GRPCHeaderMatchRegularExpression:
				if err := regex.Validate(headerMatch.Value); err != nil {
					return nil, err
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
						return nil, err
					}
				}
				if match.Method.Method != nil {
					if err := regex.Validate(*match.Method.Method); err != nil {
						return nil, err
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
	var hasHostnameIntersection bool

	for _, listener := range parentRef.listeners {
		hosts := computeHosts(GetHostnames(route), listener)
		if len(hosts) == 0 {
			continue
		}
		hasHostnameIntersection = true

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
				hostRoute := &ir.HTTPRoute{
					Name:                  fmt.Sprintf("%s/%s", routeRoute.Name, underscoredHost),
					Metadata:              routeRoute.Metadata,
					Hostname:              host,
					PathMatch:             routeRoute.PathMatch,
					HeaderMatches:         routeRoute.HeaderMatches,
					QueryParamMatches:     routeRoute.QueryParamMatches,
					AddRequestHeaders:     routeRoute.AddRequestHeaders,
					RemoveRequestHeaders:  routeRoute.RemoveRequestHeaders,
					AddResponseHeaders:    routeRoute.AddResponseHeaders,
					RemoveResponseHeaders: routeRoute.RemoveResponseHeaders,
					Destination:           routeRoute.Destination,
					Redirect:              routeRoute.Redirect,
					DirectResponse:        routeRoute.DirectResponse,
					CredentialInjection:   routeRoute.CredentialInjection,
					URLRewrite:            routeRoute.URLRewrite,
					Mirrors:               routeRoute.Mirrors,
					ExtensionRefs:         routeRoute.ExtensionRefs,
					IsHTTP2:               routeRoute.IsHTTP2,
					SessionPersistence:    routeRoute.SessionPersistence,
					Timeout:               routeRoute.Timeout,
					Retry:                 routeRoute.Retry,
				}
				perHostRoutes = append(perHostRoutes, hostRoute)
			}
		}
		irKey := t.getIRKey(listener.gateway.Gateway)
		irListener := xdsIR[irKey].GetHTTPListener(irListenerName(listener))

		if irListener != nil {
			if GetRouteType(route) == resource.KindGRPCRoute {
				irListener.IsHTTP2 = true
			}
			irListener.Routes = append(irListener.Routes, perHostRoutes...)
		}
	}

	return hasHostnameIntersection
}

func buildRouteMetadata(route RouteContext, sectionName *gwapiv1.SectionName) *ir.ResourceMetadata {
	metadata := &ir.ResourceMetadata{
		Kind:        route.GetObjectKind().GroupVersionKind().Kind,
		Name:        route.GetName(),
		Namespace:   route.GetNamespace(),
		Annotations: filterEGPrefix(route.GetAnnotations()),
	}
	if sectionName != nil {
		metadata.SectionName = string(*sectionName)
	}
	return metadata
}

func filterEGPrefix(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		if strings.HasPrefix(k, egPrefix) {
			out[strings.TrimPrefix(k, egPrefix)] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (t *Translator) ProcessTLSRoutes(tlsRoutes []*gwapiv1a2.TLSRoute, gateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) []*TLSRouteContext {
	var relevantTLSRoutes []*TLSRouteContext

	for _, tls := range tlsRoutes {
		if tls == nil {
			panic("received nil tlsroute")
		}
		tlsRoute := &TLSRouteContext{
			GatewayControllerName: t.GatewayControllerName,
			TLSRoute:              tls.DeepCopy(),
		}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(tlsRoute, gateways, resources)
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
		var destSettings []*ir.DestinationSetting

		destName := irRouteDestinationName(tlsRoute, -1 /*rule index*/)
		// compute backends
		for _, rule := range tlsRoute.Spec.Rules {
			for i, backendRef := range rule.BackendRefs {
				settingName := irDestinationSettingName(destName, i)
				// not yet handled, requires to align with the conformance test - TLSRouteInvalidReferenceGrant.
				ds, _ := t.processDestination(settingName, backendRef, parentRef, tlsRoute, resources)
				if ds != nil {
					destSettings = append(destSettings, ds)
				}
			}

			// TODO handle:
			//	- no valid backend refs
			//	- sum of weights for valid backend refs is 0
			//	- returning 500's for invalid backend refs
			//	- etc.
		}

		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(tlsRoute, gwapiv1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			routeStatus := GetRouteStatus(tlsRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tlsRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				gwapiv1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(tlsRoute, gwapiv1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		var hasHostnameIntersection bool
		for _, listener := range parentRef.listeners {
			hosts := computeHosts(GetHostnames(tlsRoute), listener)
			if len(hosts) == 0 {
				continue
			}

			hasHostnameIntersection = true

			irKey := t.getIRKey(listener.gateway.Gateway)

			gwXdsIR := xdsIR[irKey]
			irListener := gwXdsIR.GetTCPListener(irListenerName(listener))
			if irListener != nil {
				irRoute := &ir.TCPRoute{
					Name: irTCPRouteName(tlsRoute),
					TLS: &ir.TLS{TLSInspectorConfig: &ir.TLSInspectorConfig{
						SNIs: hosts,
					}},
					Destination: &ir.RouteDestination{
						Name:     destName,
						Settings: destSettings,
					},
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
				"There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).",
			)
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
	var relevantUDPRoutes []*UDPRouteContext

	for _, u := range udpRoutes {
		if u == nil {
			panic("received nil udproute")
		}
		udpRoute := &UDPRouteContext{
			GatewayControllerName: t.GatewayControllerName,
			UDPRoute:              u.DeepCopy(),
		}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(udpRoute, gateways, resources)
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
		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		var destSettings []*ir.DestinationSetting

		// compute backends
		if len(udpRoute.Spec.Rules) != 1 {
			routeStatus := GetRouteStatus(udpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				udpRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				"InvalidRule",
				"One and only one rule is supported",
			)
			continue
		}

		destName := irRouteDestinationName(udpRoute, -1 /*rule index*/)
		for i, backendRef := range udpRoute.Spec.Rules[0].BackendRefs {
			settingName := irDestinationSettingName(destName, i)
			ds, err := t.processDestination(settingName, backendRef, parentRef, udpRoute, resources)
			// skip adding the route and provide the reason via route status.
			if err != nil {
				routeStatus := GetRouteStatus(udpRoute)
				status.SetRouteStatusCondition(routeStatus,
					parentRef.routeParentStatusIdx,
					udpRoute.GetGeneration(),
					gwapiv1.RouteConditionAccepted,
					metav1.ConditionFalse,
					"Failed to process the settings associated with the UDP route.",
					err.Error(),
				)
				continue
			}

			// Skip nil destination settings
			if ds != nil {
				destSettings = append(destSettings, ds)
			}
		}

		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(udpRoute, gwapiv1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			routeStatus := GetRouteStatus(udpRoute)
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
			if listener.AttachedRoutes() > 1 {
				continue
			}
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
	var relevantTCPRoutes []*TCPRouteContext

	for _, tcp := range tcpRoutes {
		if tcp == nil {
			panic("received nil tcproute")
		}
		tcpRoute := &TCPRouteContext{
			GatewayControllerName: t.GatewayControllerName,
			TCPRoute:              tcp.DeepCopy(),
		}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(tcpRoute, gateways, resources)
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
		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		var destSettings []*ir.DestinationSetting

		// compute backends
		if len(tcpRoute.Spec.Rules) != 1 {
			routeStatus := GetRouteStatus(tcpRoute)
			status.SetRouteStatusCondition(routeStatus,
				parentRef.routeParentStatusIdx,
				tcpRoute.GetGeneration(),
				gwapiv1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				"InvalidRule",
				"One and only one rule is supported",
			)
			continue
		}

		destName := irRouteDestinationName(tcpRoute, -1 /*rule index*/)
		for i, backendRef := range tcpRoute.Spec.Rules[0].BackendRefs {
			settingName := irDestinationSettingName(destName, i)
			ds, err := t.processDestination(settingName, backendRef, parentRef, tcpRoute, resources)
			// skip adding the route and provide the reason via route status.
			if err != nil {
				routeStatus := GetRouteStatus(tcpRoute)
				status.SetRouteStatusCondition(routeStatus,
					parentRef.routeParentStatusIdx,
					tcpRoute.GetGeneration(),
					gwapiv1.RouteConditionAccepted,
					metav1.ConditionFalse,
					"Failed to process the settings associated with the TCP route.",
					err.Error(),
				)
				continue
			}
			// Skip nil destination settings
			if ds != nil {
				destSettings = append(destSettings, ds)
			}
		}

		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(tcpRoute, gwapiv1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			routeStatus := GetRouteStatus(tcpRoute)
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
			if listener.AttachedRoutes() > 1 {
				continue
			}
			if !listener.IsReady() {
				continue
			}
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
					},
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
) (ds *ir.DestinationSetting, err StatusError) {
	routeType := GetRouteType(route)
	weight := uint32(1)
	backendRef := GetBackendRef(backendRefContext)
	if backendRef.Weight != nil {
		weight = uint32(*backendRef.Weight)
	}

	backendNamespace := NamespaceDerefOr(backendRef.Namespace, route.GetNamespace())
	err = t.validateBackendRef(backendRefContext, parentRef, route, resources, backendNamespace, routeType)
	{
		// return with empty endpoint means the backend is invalid and an error to fail the associated route.
		if err != nil {
			return nil, err
		}
	}

	// Skip processing backends with 0 weight
	if weight == 0 {
		return nil, nil
	}

	var envoyProxy *egv1a1.EnvoyProxy
	gatewayCtx := GetRouteParentContext(route, *parentRef.ParentReference).GetGateway()
	if gatewayCtx != nil {
		envoyProxy = gatewayCtx.envoyProxy
	}

	protocol := inspectAppProtocolByRouteKind(routeType)

	switch KindDerefOr(backendRef.Kind, resource.KindService) {
	case resource.KindServiceImport:
		ds = t.processServiceImportDestinationSetting(name, backendRef.BackendObjectReference, backendNamespace, protocol, resources, envoyProxy)

	case resource.KindService:
		ds = t.processServiceDestinationSetting(name, backendRef.BackendObjectReference, backendNamespace, protocol, resources, envoyProxy)
		svc := resources.GetService(backendNamespace, string(backendRef.Name))
		ds.IPFamily = getServiceIPFamily(svc)
		ds.ZoneAwareRoutingEnabled = isZoneAwareRoutingEnabled(svc)

	case egv1a1.KindBackend:
		ds = t.processBackendDestinationSetting(name, backendRef.BackendObjectReference, backendNamespace, protocol, resources)
	}

	var tlsErr error
	ds.TLS, tlsErr = t.applyBackendTLSSetting(
		backendRef.BackendObjectReference,
		backendNamespace,
		gwapiv1a2.ParentReference{
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
		return nil, &RouteStatusError{
			Wrapped:              tlsErr,
			RouteConditionReason: RouteReasonInvalidBackendTLS,
		}
	}

	var filtersErr error
	ds.Filters, filtersErr = t.processDestinationFilters(routeType, backendRefContext, parentRef, route, resources)
	if filtersErr != nil {
		return nil, &RouteStatusError{
			Wrapped:              filtersErr,
			RouteConditionReason: RouteReasonInvalidBackendFilters,
		}
	}

	if err := validateDestinationSettings(ds, t.IsEnvoyServiceRouting(envoyProxy), backendRef.Kind); err != nil {
		routeStatus := GetRouteStatus(route)
		status.SetRouteStatusCondition(routeStatus,
			parentRef.routeParentStatusIdx,
			route.GetGeneration(),
			gwapiv1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			gwapiv1.RouteReasonResolvedRefs,
			err.Error())
		return nil, err
	}

	ds.Weight = &weight
	return ds, nil
}

func validateDestinationSettings(destinationSettings *ir.DestinationSetting, endpointRoutingDisabled bool, kind *gwapiv1.Kind) StatusError {
	// TODO: support mixed endpointslice address type for the same backendRef
	switch KindDerefOr(kind, resource.KindService) {
	case egv1a1.KindBackend:
		if destinationSettings.AddressType != nil && *destinationSettings.AddressType == ir.MIXED {
			return &RouteStatusError{
				Wrapped:              fmt.Errorf("mixed FQDN and IP or Unix address type for the same backendRef is not supported"),
				RouteConditionReason: RouteReasonInvalidAddressType,
			}
		}
	case resource.KindService, resource.KindServiceImport:
		if !endpointRoutingDisabled && destinationSettings.AddressType != nil && *destinationSettings.AddressType == ir.MIXED {
			return &RouteStatusError{
				Wrapped:              fmt.Errorf("mixed endpointslice address type for the same backendRef is not supported"),
				RouteConditionReason: RouteReasonInvalidAddressType,
			}
		}
	}

	return nil
}

func (t *Translator) processServiceImportDestinationSetting(
	name string,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	protocol ir.AppProtocol,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) *ir.DestinationSetting {
	var (
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
	)

	serviceImport := resources.GetServiceImport(backendNamespace, string(backendRef.Name))
	var servicePort mcsapiv1a1.ServicePort
	for _, port := range serviceImport.Spec.Ports {
		if port.Port == int32(*backendRef.Port) {
			servicePort = port
			break
		}
	}

	if servicePort.AppProtocol != nil {
		protocol = serviceAppProtocolToIRAppProtocol(*servicePort.AppProtocol, protocol, false)
	}

	// Route to endpoints by default
	if !t.IsEnvoyServiceRouting(envoyProxy) {
		endpointSlices := resources.GetEndpointSlicesForBackend(backendNamespace, string(backendRef.Name), resource.KindServiceImport)
		endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, servicePort.Name, servicePort.Protocol)
	} else {
		// Fall back to Service ClusterIP routing
		backendIps := resources.GetServiceImport(backendNamespace, string(backendRef.Name)).Spec.IPs
		for _, ip := range backendIps {
			ep := ir.NewDestEndpoint(ip, uint32(*backendRef.Port), false, nil)
			endpoints = append(endpoints, ep)
		}
	}

	return &ir.DestinationSetting{
		Name:        name,
		Protocol:    protocol,
		Endpoints:   endpoints,
		AddressType: addrType,
	}
}

func (t *Translator) processServiceDestinationSetting(
	name string,
	backendRef gwapiv1.BackendObjectReference,
	backendNamespace string,
	protocol ir.AppProtocol,
	resources *resource.Resources,
	envoyProxy *egv1a1.EnvoyProxy,
) *ir.DestinationSetting {
	var (
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
	)

	service := resources.GetService(backendNamespace, string(backendRef.Name))
	var servicePort corev1.ServicePort
	for _, port := range service.Spec.Ports {
		if port.Port == int32(*backendRef.Port) {
			servicePort = port
			break
		}
	}

	// support HTTPRouteBackendProtocolH2C/GRPC
	if servicePort.AppProtocol != nil {
		protocol = serviceAppProtocolToIRAppProtocol(*servicePort.AppProtocol, protocol, true)
	}

	// Route to endpoints by default
	if !t.IsEnvoyServiceRouting(envoyProxy) {
		endpointSlices := resources.GetEndpointSlicesForBackend(backendNamespace, string(backendRef.Name), KindDerefOr(backendRef.Kind, resource.KindService))
		endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, servicePort.Name, servicePort.Protocol)
	} else {
		// Fall back to Service ClusterIP routing
		ep := ir.NewDestEndpoint(service.Spec.ClusterIP, uint32(*backendRef.Port), false, nil)
		endpoints = append(endpoints, ep)
	}

	return &ir.DestinationSetting{
		Name:                    name,
		Protocol:                protocol,
		Endpoints:               endpoints,
		AddressType:             addrType,
		ZoneAwareRoutingEnabled: isZoneAwareRoutingEnabled(service),
	}
}

func getBackendFilters(routeType gwapiv1.Kind, backendRefContext BackendRefContext) (backendFilters any) {
	filters := GetFilters(backendRefContext)
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

func isZoneAwareRoutingEnabled(svc *corev1.Service) bool {
	if trafficDist := svc.Spec.TrafficDistribution; trafficDist != nil {
		return *trafficDist == corev1.ServiceTrafficDistributionPreferClose
	}

	// Allows annotation values that align with Kubernetes defaults.
	// Ref:
	// https://kubernetes.io/docs/concepts/services-networking/topology-aware-routing/#enabling-topology-aware-routing
	// https://github.com/kubernetes/kubernetes/blob/9d9e1afdf78bce0a517cc22557457f942040ca19/staging/src/k8s.io/endpointslice/utils.go#L355-L368
	if val, ok := svc.Annotations[corev1.AnnotationTopologyMode]; ok {
		return val == "Auto" || val == "auto"
	}

	return false
}

func (t *Translator) processDestinationFilters(routeType gwapiv1.Kind, backendRefContext BackendRefContext, parentRef *RouteParentContext, route RouteContext, resources *resource.Resources) (*ir.DestinationFilters, error) {
	backendFilters := getBackendFilters(routeType, backendRefContext)
	if backendFilters == nil {
		return nil, nil
	}

	var httpFiltersContext *HTTPFiltersContext
	var destFilters ir.DestinationFilters

	var err error
	switch filters := backendFilters.(type) {
	case []gwapiv1.HTTPRouteFilter:
		httpFiltersContext, err = t.ProcessHTTPFilters(parentRef, route, filters, 0, resources)

	case []gwapiv1.GRPCRouteFilter:
		httpFiltersContext, err = t.ProcessGRPCFilters(parentRef, route, filters, resources)
		if err != nil {
			return &destFilters, err
		}
	}
	applyHTTPFiltersContextToDestinationFilters(httpFiltersContext, &destFilters)

	return &destFilters, err
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

// processAllowedListenersForParentRefs finds out if the route attaches to one of our
// Gateways' listeners, and if so, gets the list of listeners that allow it to
// attach for each parentRef.
func (t *Translator) processAllowedListenersForParentRefs(routeContext RouteContext, gateways []*GatewayContext, resources *resource.Resources) bool {
	var relevantRoute bool
	ns := gwapiv1.Namespace(routeContext.GetNamespace())
	for _, parentRef := range GetParentReferences(routeContext) {
		isRelevantParentRef, selectedListeners := GetReferencedListeners(ns, parentRef, gateways)

		// Parent ref is not to a Gateway that we control: skip it
		if !isRelevantParentRef {
			continue
		}
		relevantRoute = true

		parentRefCtx := GetRouteParentContext(routeContext, parentRef)
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
			acceptedKind := GetRouteType(routeContext)
			if listener.AllowsKind(gwapiv1.RouteGroupKind{Group: GroupPtr(gwapiv1.GroupName), Kind: acceptedKind}) &&
				listener.AllowsNamespace(resources.GetNamespace(routeContext.GetNamespace())) {
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

		// Its safe to increment AttachedRoutes since we've found a valid parentRef
		// and the listener allows this Route kind

		// Theoretically there should only be one parent ref per
		// Route that attaches to a given Listener, so fine to just increment here, but we
		// might want to check to ensure we're not double-counting.
		for _, listener := range allowedListeners {
			listener.IncrementAttachedRoutes()
		}

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
			if conditions.Serving != nil && conditions.Terminating != nil {
				// Check if the endpoint is serving
				if !*conditions.Serving {
					continue
				}
				// Drain the endpoint if it is being terminated
				draining := *conditions.Terminating
				for _, address := range endpoint.Addresses {
					ep := ir.NewDestEndpoint(address, uint32(*endpointPort.Port), draining, endpoint.Zone)
					endpoints = append(endpoints, ep)
				}
			} else if conditions.Ready == nil || *conditions.Ready {
				for _, address := range endpoint.Addresses {
					ep := ir.NewDestEndpoint(address, uint32(*endpointPort.Port), false, endpoint.Zone)
					endpoints = append(endpoints, ep)
				}
			}
		}
	}

	return endpoints
}

func getTargetBackendReference(backendRef gwapiv1a2.BackendObjectReference, backendNamespace string, resources *resource.Resources) gwapiv1a2.LocalPolicyTargetReferenceWithSectionName {
	ref := gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
		LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
			Group: func() gwapiv1a2.Group {
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
		if service := resources.GetService(backendNamespace, string(backendRef.Name)); service != nil {
			for _, port := range service.Spec.Ports {
				if port.Port == int32(*backendRef.Port) {
					if port.Name != "" {
						ref.SectionName = SectionNamePtr(port.Name)
						break
					}
				}
			}
		}

	case *backendRef.Kind == resource.KindServiceImport:
		if si := resources.GetServiceImport(backendNamespace, string(backendRef.Name)); si != nil {
			for _, port := range si.Spec.Ports {
				if port.Port == int32(*backendRef.Port) {
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
	resources *resource.Resources,
) *ir.DestinationSetting {
	var (
		dstEndpoints []*ir.DestinationEndpoint
		dstAddrType  *ir.DestinationAddressType
	)

	addrTypeMap := make(map[ir.DestinationAddressType]int)

	backend := resources.GetBackend(backendNamespace, string(backendRef.Name))
	ds := &ir.DestinationSetting{Name: name}

	if backend.Spec.Type != nil && *backend.Spec.Type == egv1a1.BackendTypeDynamicResolver {
		ds.IsDynamicResolver = true
	} else {
		for _, bep := range backend.Spec.Endpoints {
			var irde *ir.DestinationEndpoint
			switch {
			case bep.IP != nil:
				ip := net.ParseIP(bep.IP.Address)
				if ip != nil {
					addrTypeMap[ir.IP]++
					irde = &ir.DestinationEndpoint{
						Host: bep.IP.Address,
						Port: uint32(bep.IP.Port),
					}
				}
			case bep.FQDN != nil:
				addrTypeMap[ir.FQDN]++
				irde = &ir.DestinationEndpoint{
					Host: bep.FQDN.Hostname,
					Port: uint32(bep.FQDN.Port),
				}
			case bep.Unix != nil:
				addrTypeMap[ir.IP]++
				irde = &ir.DestinationEndpoint{
					Path: ptr.To(bep.Unix.Path),
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

		for _, ap := range backend.Spec.AppProtocols {
			protocol = backendAppProtocolToIRAppProtocol(ap, protocol)
		}
		ds.Endpoints = dstEndpoints
		ds.AddressType = dstAddrType
		ds.Protocol = protocol
	}

	if backend.Spec.Fallback != nil {
		// set only the secondary priority, the backend defaults to a primary priority if unset.
		if ptr.Deref(backend.Spec.Fallback, false) {
			ds.Priority = ptr.To(uint32(1))
		}
	}

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
