// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/ir"
)

var (
	_                RoutesTranslator = (*Translator)(nil)
	validServiceName                  = `(?i)\.?[a-z_][a-z_0-9]*(\.[a-z_][a-z_0-9]*)*`
	validMethodName                   = `[A-Za-z_][A-Za-z_0-9]*`
)

type RoutesTranslator interface {
	ProcessHTTPRoutes(httpRoutes []*v1beta1.HTTPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*HTTPRouteContext
	ProcessGRPCRoutes(grpcRoutes []*v1alpha2.GRPCRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*GRPCRouteContext
	ProcessTLSRoutes(tlsRoutes []*v1alpha2.TLSRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*TLSRouteContext
	ProcessTCPRoutes(tcpRoutes []*v1alpha2.TCPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*TCPRouteContext
	ProcessUDPRoutes(udpRoutes []*v1alpha2.UDPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*UDPRouteContext
}

func (t *Translator) ProcessHTTPRoutes(httpRoutes []*v1beta1.HTTPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*HTTPRouteContext {
	var relevantHTTPRoutes []*HTTPRouteContext

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

func (t *Translator) ProcessGRPCRoutes(grpcRoutes []*v1alpha2.GRPCRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*GRPCRouteContext {
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

func (t *Translator) processHTTPRouteParentRefs(httpRoute *HTTPRouteContext, resources *Resources, xdsIR XdsIRMap) {
	for _, parentRef := range httpRoute.ParentRefs {
		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		routeRoutes := t.processHTTPRouteRules(httpRoute, parentRef, resources)

		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(httpRoute, v1beta1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			parentRef.SetCondition(httpRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				v1beta1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(httpRoute, v1beta1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		var hasHostnameIntersection = t.processHTTPRouteParentRefListener(httpRoute, routeRoutes, parentRef, xdsIR)
		if !hasHostnameIntersection {
			parentRef.SetCondition(httpRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonNoMatchingListenerHostname,
				"There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).",
			)
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if parentRef.HTTPRoute != nil &&
			len(parentRef.HTTPRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			parentRef.SetCondition(httpRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionTrue,
				v1beta1.RouteReasonAccepted,
				"Route is accepted",
			)
		}

	}
}

func (t *Translator) processHTTPRouteRules(httpRoute *HTTPRouteContext, parentRef *RouteParentContext, resources *Resources) []*ir.HTTPRoute {
	var routeRoutes []*ir.HTTPRoute

	// compute matches, filters, backends
	for ruleIdx, rule := range httpRoute.Spec.Rules {
		httpFiltersContext := t.ProcessHTTPFilters(parentRef, httpRoute, rule.Filters, resources)

		// A rule is matched if any one of its matches
		// is satisfied (i.e. a logical "OR"), so generate
		// a unique Xds IR HTTPRoute per match.
		var ruleRoutes = t.processHTTPRouteRule(httpRoute, ruleIdx, httpFiltersContext, rule)

		for _, backendRef := range rule.BackendRefs {
			destinations, backendWeight := t.processRouteDestinations(backendRef.BackendRef, parentRef, httpRoute, resources)
			for _, route := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// processing any destinations for this route.
				if route.DirectResponse == nil && route.Redirect == nil {
					if len(destinations) > 0 {
						route.Destinations = append(route.Destinations, destinations...)
						route.BackendWeights.Valid += backendWeight

					} else {
						route.BackendWeights.Invalid += backendWeight
					}
				}
			}
		}

		// If the route has no valid backends then just use a direct response and don't fuss with weighted responses
		for _, ruleRoute := range ruleRoutes {
			if ruleRoute.BackendWeights.Invalid > 0 && len(ruleRoute.Destinations) == 0 {
				ruleRoute.DirectResponse = &ir.DirectResponse{
					StatusCode: 500,
				}
			}
		}

		// TODO handle:
		//	- sum of weights for valid backend refs is 0
		//	- etc.

		routeRoutes = append(routeRoutes, ruleRoutes...)
	}

	return routeRoutes
}

func (t *Translator) processHTTPRouteRule(httpRoute *HTTPRouteContext, ruleIdx int, httpFiltersContext *HTTPFiltersContext, rule v1beta1.HTTPRouteRule) []*ir.HTTPRoute {
	var ruleRoutes []*ir.HTTPRoute

	// If no matches are specified, the implementation MUST match every HTTP request.
	if len(rule.Matches) == 0 {
		irRoute := &ir.HTTPRoute{
			Name: routeName(httpRoute, ruleIdx, -1),
		}
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
		ruleRoutes = append(ruleRoutes, irRoute)
	}

	// A rule is matched if any one of its matches
	// is satisfied (i.e. a logical "OR"), so generate
	// a unique Xds IR HTTPRoute per match.
	for matchIdx, match := range rule.Matches {
		irRoute := &ir.HTTPRoute{
			Name: routeName(httpRoute, ruleIdx, matchIdx),
		}

		if match.Path != nil {
			switch PathMatchTypeDerefOr(match.Path.Type, v1beta1.PathMatchPathPrefix) {
			case v1beta1.PathMatchPathPrefix:
				irRoute.PathMatch = &ir.StringMatch{
					Prefix: match.Path.Value,
				}
			case v1beta1.PathMatchExact:
				irRoute.PathMatch = &ir.StringMatch{
					Exact: match.Path.Value,
				}
			case v1beta1.PathMatchRegularExpression:
				irRoute.PathMatch = &ir.StringMatch{
					SafeRegex: match.Path.Value,
				}
			}
		}
		for _, headerMatch := range match.Headers {
			switch HeaderMatchTypeDerefOr(headerMatch.Type, v1beta1.HeaderMatchExact) {
			case v1beta1.HeaderMatchExact:
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:  string(headerMatch.Name),
					Exact: StringPtr(headerMatch.Value),
				})
			case v1beta1.HeaderMatchRegularExpression:
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:      string(headerMatch.Name),
					SafeRegex: StringPtr(headerMatch.Value),
				})
			}
		}
		for _, queryParamMatch := range match.QueryParams {
			switch QueryParamMatchTypeDerefOr(queryParamMatch.Type, v1beta1.QueryParamMatchExact) {
			case v1beta1.QueryParamMatchExact:
				irRoute.QueryParamMatches = append(irRoute.QueryParamMatches, &ir.StringMatch{
					Name:  string(queryParamMatch.Name),
					Exact: StringPtr(queryParamMatch.Value),
				})
			case v1beta1.QueryParamMatchRegularExpression:
				irRoute.QueryParamMatches = append(irRoute.QueryParamMatches, &ir.StringMatch{
					Name:      string(queryParamMatch.Name),
					SafeRegex: StringPtr(queryParamMatch.Value),
				})
			}
		}

		if match.Method != nil {
			irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
				Name:  ":method",
				Exact: StringPtr(string(*match.Method)),
			})
		}
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
		ruleRoutes = append(ruleRoutes, irRoute)
	}

	return ruleRoutes
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
	if len(httpFiltersContext.Mirrors) > 0 {
		irRoute.Mirrors = httpFiltersContext.Mirrors
	}
	if httpFiltersContext.RequestAuthentication != nil {
		irRoute.RequestAuthentication = httpFiltersContext.RequestAuthentication
	}
	if httpFiltersContext.RateLimit != nil {
		irRoute.RateLimit = httpFiltersContext.RateLimit
	}
	if len(httpFiltersContext.ExtensionRefs) > 0 {
		irRoute.ExtensionRefs = httpFiltersContext.ExtensionRefs
	}

}

func (t *Translator) processGRPCRouteParentRefs(grpcRoute *GRPCRouteContext, resources *Resources, xdsIR XdsIRMap) {
	for _, parentRef := range grpcRoute.ParentRefs {

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		routeRoutes := t.processGRPCRouteRules(grpcRoute, parentRef, resources)

		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(grpcRoute, v1beta1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			parentRef.SetCondition(grpcRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				v1beta1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		if parentRef.HasCondition(grpcRoute, v1beta1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}
		var hasHostnameIntersection = t.processHTTPRouteParentRefListener(grpcRoute, routeRoutes, parentRef, xdsIR)
		if !hasHostnameIntersection {
			parentRef.SetCondition(grpcRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonNoMatchingListenerHostname,
				"There were no hostname intersections between the GRPCRoute and this parent ref's Listener(s).",
			)
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if parentRef.GRPCRoute != nil &&
			len(parentRef.GRPCRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			parentRef.SetCondition(grpcRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionTrue,
				v1beta1.RouteReasonAccepted,
				"Route is accepted",
			)
		}

	}
}

func (t *Translator) processGRPCRouteRules(grpcRoute *GRPCRouteContext, parentRef *RouteParentContext, resources *Resources) []*ir.HTTPRoute {
	var routeRoutes []*ir.HTTPRoute

	// compute matches, filters, backends
	for ruleIdx, rule := range grpcRoute.Spec.Rules {
		httpFiltersContext := t.ProcessGRPCFilters(parentRef, grpcRoute, rule.Filters, resources)

		// A rule is matched if any one of its matches
		// is satisfied (i.e. a logical "OR"), so generate
		// a unique Xds IR HTTPRoute per match.
		var ruleRoutes = t.processGRPCRouteRule(grpcRoute, ruleIdx, httpFiltersContext, rule)

		for _, backendRef := range rule.BackendRefs {
			destinations, backendWeight := t.processRouteDestinations(backendRef.BackendRef, parentRef, grpcRoute, resources)
			for _, route := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// processing any destinations for this route.
				if route.DirectResponse == nil && route.Redirect == nil {
					if len(destinations) > 0 {
						route.Destinations = append(route.Destinations, destinations...)
						route.BackendWeights.Valid += backendWeight

					} else {
						route.BackendWeights.Invalid += backendWeight
					}
				}
			}
		}

		// If the route has no valid backends then just use a direct response and don't fuss with weighted responses
		for _, ruleRoute := range ruleRoutes {
			if ruleRoute.BackendWeights.Invalid > 0 && len(ruleRoute.Destinations) == 0 {
				ruleRoute.DirectResponse = &ir.DirectResponse{
					StatusCode: 500,
				}
			}
		}

		// TODO handle:
		//	- sum of weights for valid backend refs is 0
		//	- etc.

		routeRoutes = append(routeRoutes, ruleRoutes...)
	}

	return routeRoutes
}

func (t *Translator) processGRPCRouteRule(grpcRoute *GRPCRouteContext, ruleIdx int, httpFiltersContext *HTTPFiltersContext, rule v1alpha2.GRPCRouteRule) []*ir.HTTPRoute {
	var ruleRoutes []*ir.HTTPRoute

	// If no matches are specified, the implementation MUST match every gRPC request.
	if len(rule.Matches) == 0 {
		irRoute := &ir.HTTPRoute{
			Name: routeName(grpcRoute, ruleIdx, -1),
		}
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
		ruleRoutes = append(ruleRoutes, irRoute)
	}

	// A rule is matched if any one of its matches
	// is satisfied (i.e. a logical "OR"), so generate
	// a unique Xds IR HTTPRoute per match.
	for matchIdx, match := range rule.Matches {
		irRoute := &ir.HTTPRoute{
			Name: routeName(grpcRoute, ruleIdx, matchIdx),
		}

		for _, headerMatch := range match.Headers {
			switch HeaderMatchTypeDerefOr(headerMatch.Type, v1beta1.HeaderMatchExact) {
			case v1beta1.HeaderMatchExact:
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:  string(headerMatch.Name),
					Exact: StringPtr(headerMatch.Value),
				})
			case v1beta1.HeaderMatchRegularExpression:
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:      string(headerMatch.Name),
					SafeRegex: StringPtr(headerMatch.Value),
				})
			}
		}

		if match.Method != nil {
			// GRPC's path is in the form of "/<service>/<method>"
			switch GRPCMethodMatchTypeDerefOr(match.Method.Type, v1alpha2.GRPCMethodMatchExact) {
			case v1alpha2.GRPCMethodMatchExact:
				t.processGRPCRouteMethodExact(match.Method, irRoute)
			case v1alpha2.GRPCMethodMatchRegularExpression:
				t.processGRPCRouteMethodRegularExpression(match.Method, irRoute)
			}
		}

		ruleRoutes = append(ruleRoutes, irRoute)
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
	}
	return ruleRoutes
}

func (t *Translator) processGRPCRouteMethodExact(method *v1alpha2.GRPCMethodMatch, irRoute *ir.HTTPRoute) {
	switch {
	case method.Service != nil && method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			Exact: StringPtr(fmt.Sprintf("/%s/%s", *method.Service, *method.Method)),
		}
	case method.Method != nil:
		// Use a header match since the PathMatch doesn't support Suffix matching
		irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
			Name:   ":path",
			Suffix: StringPtr(fmt.Sprintf("/%s", *method.Method)),
		})
	case method.Service != nil:
		irRoute.PathMatch = &ir.StringMatch{
			Prefix: StringPtr(fmt.Sprintf("/%s", *method.Service)),
		}
	}
}

func (t *Translator) processGRPCRouteMethodRegularExpression(method *v1alpha2.GRPCMethodMatch, irRoute *ir.HTTPRoute) {
	switch {
	case method.Service != nil && method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: StringPtr(fmt.Sprintf("/%s/%s", *method.Service, *method.Method)),
		}
	case method.Method != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: StringPtr(fmt.Sprintf("/%s/%s", validServiceName, *method.Method)),
		}
	case method.Service != nil:
		irRoute.PathMatch = &ir.StringMatch{
			SafeRegex: StringPtr(fmt.Sprintf("/%s/%s", *method.Service, validMethodName)),
		}
	}
}

func (t *Translator) processHTTPRouteParentRefListener(route RouteContext, routeRoutes []*ir.HTTPRoute, parentRef *RouteParentContext, xdsIR XdsIRMap) bool {
	var hasHostnameIntersection bool

	for _, listener := range parentRef.listeners {
		hosts := computeHosts(GetHostnames(route), listener.Hostname)
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

				hostRoute := &ir.HTTPRoute{
					Name:                  fmt.Sprintf("%s-%s", routeRoute.Name, host),
					Hostname:              host,
					PathMatch:             routeRoute.PathMatch,
					HeaderMatches:         routeRoute.HeaderMatches,
					QueryParamMatches:     routeRoute.QueryParamMatches,
					AddRequestHeaders:     routeRoute.AddRequestHeaders,
					RemoveRequestHeaders:  routeRoute.RemoveRequestHeaders,
					AddResponseHeaders:    routeRoute.AddResponseHeaders,
					RemoveResponseHeaders: routeRoute.RemoveResponseHeaders,
					Destinations:          routeRoute.Destinations,
					Redirect:              routeRoute.Redirect,
					DirectResponse:        routeRoute.DirectResponse,
					URLRewrite:            routeRoute.URLRewrite,
					Mirrors:               routeRoute.Mirrors,
					RequestAuthentication: routeRoute.RequestAuthentication,
					RateLimit:             routeRoute.RateLimit,
					ExtensionRefs:         routeRoute.ExtensionRefs,
				}
				if hostRoute.Hostname == "*" {
					hostRoute.Hostname = ""
				}
				// Don't bother copying over the weights unless the route has invalid backends.
				if routeRoute.BackendWeights.Invalid > 0 {
					hostRoute.BackendWeights = routeRoute.BackendWeights
				}
				perHostRoutes = append(perHostRoutes, hostRoute)
			}
		}

		irKey := irStringKey(listener.gateway.Namespace, listener.gateway.Name)
		irListener := xdsIR[irKey].GetHTTPListener(irHTTPListenerName(listener))
		if irListener != nil {
			if GetRouteType(route) == KindGRPCRoute {
				irListener.IsHTTP2 = true
			}
			irListener.Routes = append(irListener.Routes, perHostRoutes...)
		}
		// Theoretically there should only be one parent ref per
		// Route that attaches to a given Listener, so fine to just increment here, but we
		// might want to check to ensure we're not double-counting.
		if len(routeRoutes) > 0 {
			listener.IncrementAttachedRoutes()
		}
	}

	return hasHostnameIntersection
}

func (t *Translator) ProcessTLSRoutes(tlsRoutes []*v1alpha2.TLSRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*TLSRouteContext {
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

func (t *Translator) processTLSRouteParentRefs(tlsRoute *TLSRouteContext, resources *Resources, xdsIR XdsIRMap) {
	for _, parentRef := range tlsRoute.ParentRefs {

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		var routeDestinations []*ir.RouteDestination

		// compute backends
		for _, rule := range tlsRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				backendRef := backendRef
				destinations, _ := t.processRouteDestinations(backendRef, parentRef, tlsRoute, resources)
				routeDestinations = append(routeDestinations, destinations...)
			}

			// TODO handle:
			//	- no valid backend refs
			//	- sum of weights for valid backend refs is 0
			//	- returning 500's for invalid backend refs
			//	- etc.
		}

		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(tlsRoute, v1beta1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			parentRef.SetCondition(tlsRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				v1beta1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(tlsRoute, v1beta1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		var hasHostnameIntersection bool
		for _, listener := range parentRef.listeners {
			hosts := computeHosts(GetHostnames(tlsRoute), listener.Hostname)
			if len(hosts) == 0 {
				continue
			}

			hasHostnameIntersection = true

			irKey := irStringKey(listener.gateway.Namespace, listener.gateway.Name)
			containerPort := servicePortToContainerPort(int32(listener.Port))
			// Create the TCP Listener while parsing the TLSRoute since
			// the listener directly links to a routeDestination.
			irListener := &ir.TCPListener{
				Name:    irTLSListenerName(listener, tlsRoute),
				Address: "0.0.0.0",
				Port:    uint32(containerPort),
				TLS: &ir.TLS{Passthrough: &ir.TLSInspectorConfig{
					SNIs: hosts,
				}},
				Destinations: routeDestinations,
			}
			gwXdsIR := xdsIR[irKey]
			gwXdsIR.TCP = append(gwXdsIR.TCP, irListener)

			// Theoretically there should only be one parent ref per
			// Route that attaches to a given Listener, so fine to just increment here, but we
			// might want to check to ensure we're not double-counting.
			if len(routeDestinations) > 0 {
				listener.IncrementAttachedRoutes()
			}
		}

		if !hasHostnameIntersection {
			parentRef.SetCondition(tlsRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonNoMatchingListenerHostname,
				"There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).",
			)
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if parentRef.TLSRoute != nil &&
			len(parentRef.TLSRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			parentRef.SetCondition(tlsRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionTrue,
				v1beta1.RouteReasonAccepted,
				"Route is accepted",
			)
		}
	}
}

func (t *Translator) ProcessUDPRoutes(udpRoutes []*v1alpha2.UDPRoute, gateways []*GatewayContext, resources *Resources,
	xdsIR XdsIRMap) []*UDPRouteContext {
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

func (t *Translator) processUDPRouteParentRefs(udpRoute *UDPRouteContext, resources *Resources, xdsIR XdsIRMap) {
	for _, parentRef := range udpRoute.ParentRefs {
		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		var routeDestinations []*ir.RouteDestination

		// compute backends
		if len(udpRoute.Spec.Rules) != 1 {
			parentRef.SetCondition(udpRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				"InvalidRule",
				"One and only one rule is supported",
			)
			continue
		}
		if len(udpRoute.Spec.Rules[0].BackendRefs) != 1 {
			parentRef.SetCondition(udpRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				"InvalidBackend",
				"One and only one backend is supported",
			)
			continue
		}

		backendRef := udpRoute.Spec.Rules[0].BackendRefs[0]
		destinations, _ := t.processRouteDestinations(backendRef, parentRef, udpRoute, resources)
		// Skip further processing if route destination is not valid
		if len(destinations) == 0 {
			continue
		}

		routeDestinations = append(routeDestinations, destinations...)
		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(udpRoute, v1beta1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			parentRef.SetCondition(udpRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				v1beta1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(udpRoute, v1beta1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		accepted := false
		for _, listener := range parentRef.listeners {
			// only one route is allowed for a UDP listener
			if listener.AttachedRoutes() > 0 {
				continue
			}
			if !listener.IsReady() {
				continue
			}
			accepted = true
			irKey := irStringKey(listener.gateway.Namespace, listener.gateway.Name)
			containerPort := servicePortToContainerPort(int32(listener.Port))
			// Create the UDP Listener while parsing the UDPRoute since
			// the listener directly links to a routeDestination.
			irListener := &ir.UDPListener{
				Name:         irUDPListenerName(listener, udpRoute),
				Address:      "0.0.0.0",
				Port:         uint32(containerPort),
				Destinations: routeDestinations,
			}
			gwXdsIR := xdsIR[irKey]
			gwXdsIR.UDP = append(gwXdsIR.UDP, irListener)

			// Theoretically there should only be one parent ref per
			// Route that attaches to a given Listener, so fine to just increment here, but we
			// might want to check to ensure we're not double-counting.
			if len(routeDestinations) > 0 {
				listener.IncrementAttachedRoutes()
			}
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if accepted && parentRef.UDPRoute != nil &&
			len(parentRef.UDPRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			parentRef.SetCondition(udpRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionTrue,
				v1beta1.RouteReasonAccepted,
				"Route is accepted",
			)
		}

		if !accepted {
			parentRef.SetCondition(udpRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonUnsupportedValue,
				"Multiple routes on the same UDP listener",
			)
		}
	}
}

func (t *Translator) ProcessTCPRoutes(tcpRoutes []*v1alpha2.TCPRoute, gateways []*GatewayContext, resources *Resources,
	xdsIR XdsIRMap) []*TCPRouteContext {
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

func (t *Translator) processTCPRouteParentRefs(tcpRoute *TCPRouteContext, resources *Resources, xdsIR XdsIRMap) {
	for _, parentRef := range tcpRoute.ParentRefs {

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		var routeDestinations []*ir.RouteDestination

		// compute backends
		if len(tcpRoute.Spec.Rules) != 1 {
			parentRef.SetCondition(tcpRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				"InvalidRule",
				"One and only one rule is supported",
			)
			continue
		}
		if len(tcpRoute.Spec.Rules[0].BackendRefs) != 1 {
			parentRef.SetCondition(tcpRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				"InvalidBackend",
				"One and only one backend is supported",
			)
			continue
		}

		backendRef := tcpRoute.Spec.Rules[0].BackendRefs[0]
		destinations, _ := t.processRouteDestinations(backendRef, parentRef, tcpRoute, resources)
		// Skip further processing if route destination is not valid
		if len(destinations) == 0 {
			continue
		}
		routeDestinations = append(routeDestinations, destinations...)
		// If no negative condition has been set for ResolvedRefs, set "ResolvedRefs=True"
		if !parentRef.HasCondition(tcpRoute, v1beta1.RouteConditionResolvedRefs, metav1.ConditionFalse) {
			parentRef.SetCondition(tcpRoute,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionTrue,
				v1beta1.RouteReasonResolvedRefs,
				"Resolved all the Object references for the Route",
			)
		}

		// Skip parent refs that did not accept the route
		if parentRef.HasCondition(tcpRoute, v1beta1.RouteConditionAccepted, metav1.ConditionFalse) {
			continue
		}

		accepted := false
		for _, listener := range parentRef.listeners {
			// only one route is allowed for a TCP listener
			if listener.AttachedRoutes() > 0 {
				continue
			}
			if !listener.IsReady() {
				continue
			}
			accepted = true
			irKey := irStringKey(listener.gateway.Namespace, listener.gateway.Name)
			containerPort := servicePortToContainerPort(int32(listener.Port))
			// Create the TCP Listener while parsing the TCPRoute since
			// the listener directly links to a routeDestination.
			irListener := &ir.TCPListener{
				Name:         irTCPListenerName(listener, tcpRoute),
				Address:      "0.0.0.0",
				Port:         uint32(containerPort),
				Destinations: routeDestinations,
				TLS:          &ir.TLS{Terminate: irTLSConfigs(listener.tlsSecrets)},
			}
			gwXdsIR := xdsIR[irKey]
			gwXdsIR.TCP = append(gwXdsIR.TCP, irListener)

			// Theoretically there should only be one parent ref per
			// Route that attaches to a given Listener, so fine to just increment here, but we
			// might want to check to ensure we're not double-counting.
			if len(routeDestinations) > 0 {
				listener.IncrementAttachedRoutes()
			}
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if accepted && parentRef.TCPRoute != nil &&
			len(parentRef.TCPRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			parentRef.SetCondition(tcpRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionTrue,
				v1beta1.RouteReasonAccepted,
				"Route is accepted",
			)
		}
		if !accepted {
			parentRef.SetCondition(tcpRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonUnsupportedValue,
				"Multiple routes on the same TCP listener",
			)
		}

	}
}

// processRouteDestinations takes a backendRef and translates it into route destinations or sets error statuses and
// returns the weight for the backend so that 500 error responses can be returned for invalid backends in
// the same proportion as the backend would have otherwise received
func (t *Translator) processRouteDestinations(backendRef v1beta1.BackendRef,
	parentRef *RouteParentContext,
	route RouteContext,
	resources *Resources) (destinations []*ir.RouteDestination, backendWeight uint32) {

	weight := uint32(1)
	if backendRef.Weight != nil {
		weight = uint32(*backendRef.Weight)
	}

	serviceNamespace := NamespaceDerefOr(backendRef.Namespace, route.GetNamespace())
	service := resources.GetService(serviceNamespace, string(backendRef.Name))

	routeType := GetRouteType(route)
	if !t.validateBackendRef(&backendRef, parentRef, route, resources, serviceNamespace, routeType) {
		return nil, weight
	}

	var dest *ir.RouteDestination
	// Weights are not relevant for TCP and UDP Routes
	if routeType == KindTCPRoute || routeType == KindUDPRoute {
		dest = ir.NewRouteDest(
			service.Spec.ClusterIP,
			uint32(*backendRef.Port))
	} else {
		dest = ir.NewRouteDestWithWeight(
			service.Spec.ClusterIP,
			uint32(*backendRef.Port),
			weight)
	}
	destinations = append(destinations, dest)
	return destinations, weight
}

// processAllowedListenersForParentRefs finds out if the route attaches to one of our
// Gateways' listeners, and if so, gets the list of listeners that allow it to
// attach for each parentRef.
func (t *Translator) processAllowedListenersForParentRefs(routeContext RouteContext, gateways []*GatewayContext, resources *Resources) bool {
	var relevantRoute bool

	for _, parentRef := range GetParentReferences(routeContext) {
		isRelevantParentRef, selectedListeners := GetReferencedListeners(parentRef, gateways)

		// Parent ref is not to a Gateway that we control: skip it
		if !isRelevantParentRef {
			continue
		}
		relevantRoute = true

		parentRefCtx := GetRouteParentContext(routeContext, parentRef)
		// Reset conditions since they will be recomputed during translation
		parentRefCtx.ResetConditions(routeContext)

		if len(selectedListeners) == 0 {
			parentRefCtx.SetCondition(routeContext,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonNoMatchingParent,
				"No listeners match this parent ref",
			)
			continue
		}

		if !HasReadyListener(selectedListeners) {
			parentRefCtx.SetCondition(routeContext,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				"NoReadyListeners",
				"There are no ready listeners for this parent ref",
			)
			continue
		}

		var allowedListeners []*ListenerContext
		for _, listener := range selectedListeners {
			acceptedKind := GetRouteType(routeContext)
			if listener.AllowsKind(v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: acceptedKind}) &&
				listener.AllowsNamespace(resources.GetNamespace(routeContext.GetNamespace())) {
				allowedListeners = append(allowedListeners, listener)
			}
		}

		if len(allowedListeners) == 0 {
			parentRefCtx.SetCondition(routeContext,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonNotAllowedByListeners,
				"No listeners included by this parent ref allowed this attachment.",
			)
			continue
		}

		parentRefCtx.SetListeners(allowedListeners...)

		parentRefCtx.SetCondition(routeContext,
			v1beta1.RouteConditionAccepted,
			metav1.ConditionTrue,
			v1beta1.RouteReasonAccepted,
			"Route is accepted",
		)
	}
	return relevantRoute
}
