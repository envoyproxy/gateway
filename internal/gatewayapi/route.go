// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"strings"

	"github.com/envoyproxy/gateway/internal/ir"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

var _ RoutesTranslator = (*Translator)(nil)

type RoutesTranslator interface {
	ProcessHTTPRoutes(httpRoutes []*v1beta1.HTTPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*HTTPRouteContext
	ProcessGRPCRoutes(grpcRoutes []*v1alpha2.GRPCRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*GRPCRouteContext
	ProcessTLSRoutes(tlsRoutes []*v1alpha2.TLSRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*TLSRouteContext
	ProcessTCPRoutes(tcpRoutes []*v1alpha2.TCPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*TCPRouteContext
	ProcessUDPRoutes(udpRoutes []*v1alpha2.UDPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*UDPRouteContext
}

func (t *Translator) ProcessGRPCRoutes(grpcRoutes []*v1alpha2.GRPCRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*GRPCRouteContext {
	var relevantGRPCRoutes []*GRPCRouteContext

	for _, g := range grpcRoutes {
		if g == nil {
			panic("received nil grpcroute")
		}
		grpcRoute := &GRPCRouteContext{GRPCRoute: g}

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

func (t *Translator) ProcessHTTPRoutes(httpRoutes []*v1beta1.HTTPRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*HTTPRouteContext {
	var relevantHTTPRoutes []*HTTPRouteContext

	for _, h := range httpRoutes {
		if h == nil {
			panic("received nil httproute")
		}
		httpRoute := &HTTPRouteContext{HTTPRoute: h}

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

func (t *Translator) processGRPCRouteParentRefs(grpcRoute *GRPCRouteContext, resources *Resources, xdsIR XdsIRMap) {
	for _, parentRef := range grpcRoute.parentRefs {
		if !parentRef.IsAccepted(grpcRoute) {
			continue
		}

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		routeRoutes := t.processGRPCRoutes(grpcRoute, parentRef, resources)

		var hasHostnameIntersection = t.processGRPCRouteParentRefListener(grpcRoute, routeRoutes, parentRef, xdsIR)
		if !hasHostnameIntersection {
			parentRef.SetCondition(grpcRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonNoMatchingListenerHostname,
				"There were no hostname intersections between the HTTPRoute and this parent ref's Listener(s).",
			)
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if parentRef.grpcRoute != nil &&
			len(parentRef.grpcRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			parentRef.SetCondition(grpcRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionTrue,
				v1beta1.RouteReasonAccepted,
				"Route is accepted",
			)
		}
	}
}

func (t *Translator) processHTTPRouteParentRefs(httpRoute *HTTPRouteContext, resources *Resources, xdsIR XdsIRMap) {
	for _, parentRef := range httpRoute.parentRefs {
		// Skip parent refs that did not accept the route
		if !parentRef.IsAccepted(httpRoute) {
			continue
		}

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		routeRoutes := t.processHTTPRouteRules(httpRoute, parentRef, resources)

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
		if parentRef.httpRoute != nil &&
			len(parentRef.httpRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
			parentRef.SetCondition(httpRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionTrue,
				v1beta1.RouteReasonAccepted,
				"Route is accepted",
			)
		}
	}
}

func (t *Translator) processGRPCRoutes(grpcRoute *GRPCRouteContext, parentRef *RouteParentContext, resources *Resources) []*ir.HTTPRoute {
	var routeRoutes []*ir.HTTPRoute
	for ruleIdx, rule := range grpcRoute.Spec.Rules {
		var ruleRoutes []*ir.HTTPRoute

		// First see if there are any filters in the rules. Then apply those filters to any irRoutes.
		var directResponse *ir.DirectResponse
		addRequestHeaders := []ir.AddHeader{}
		removeRequestHeaders := []string{}

		addResponseHeaders := []ir.AddHeader{}
		removeResponseHeaders := []string{}

		// Process the filters for this route rule
		for _, filter := range rule.Filters {
			if directResponse != nil {
				break // If an invalid filter type has been configured then skip processing any more filters
			}
			switch filter.Type {
			case v1alpha2.GRPCRouteFilterRequestHeaderModifier:
				// Make sure the header modifier config actually exists
				headerModifier := filter.RequestHeaderModifier
				if headerModifier == nil {
					break
				}
				emptyFilterConfig := true // keep track of whether the provided config is empty or not

				// Add request headers
				if headersToAdd := headerModifier.Add; headersToAdd != nil {
					if len(headersToAdd) > 0 {
						emptyFilterConfig = false
					}
					for _, addHeader := range headersToAdd {
						emptyFilterConfig = false
						if addHeader.Name == "" {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								"RequestHeaderModifier Filter cannot add a header with an empty name",
							)
							// try to process the rest of the headers and produce a valid config.
							continue
						}
						// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
						if strings.Contains(string(addHeader.Name), "/") || strings.Contains(string(addHeader.Name), ":") {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								fmt.Sprintf("RequestHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: %q", string(addHeader.Name)),
							)
							continue
						}
						// Check if the header is a duplicate
						headerKey := string(addHeader.Name)
						canAddHeader := true
						for _, h := range addRequestHeaders {
							if strings.EqualFold(h.Name, headerKey) {
								canAddHeader = false
								break
							}
						}

						if !canAddHeader {
							continue
						}

						newHeader := ir.AddHeader{
							Name:   headerKey,
							Append: true,
							Value:  addHeader.Value,
						}

						addRequestHeaders = append(addRequestHeaders, newHeader)
					}
				}

				// Set headers
				if headersToSet := headerModifier.Set; headersToSet != nil {
					if len(headersToSet) > 0 {
						emptyFilterConfig = false
					}
					for _, setHeader := range headersToSet {

						if setHeader.Name == "" {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								"RequestHeaderModifier Filter cannot set a header with an empty name",
							)
							continue
						}
						// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
						if strings.Contains(string(setHeader.Name), "/") || strings.Contains(string(setHeader.Name), ":") {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								fmt.Sprintf("RequestHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: '%s'", string(setHeader.Name)),
							)
							continue
						}

						// Check if the header to be set has already been configured
						headerKey := string(setHeader.Name)
						canAddHeader := true
						for _, h := range addRequestHeaders {
							if strings.EqualFold(h.Name, headerKey) {
								canAddHeader = false
								break
							}
						}
						if !canAddHeader {
							continue
						}
						newHeader := ir.AddHeader{
							Name:   string(setHeader.Name),
							Append: false,
							Value:  setHeader.Value,
						}

						addRequestHeaders = append(addRequestHeaders, newHeader)
					}
				}

				// Remove request headers
				// As far as Envoy is concerned, it is ok to configure a header to be added/set and also in the list of
				// headers to remove. It will remove the original header if present and then add/set the header after.
				if headersToRemove := headerModifier.Remove; headersToRemove != nil {
					if len(headersToRemove) > 0 {
						emptyFilterConfig = false
					}
					for _, removedHeader := range headersToRemove {
						if removedHeader == "" {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								"RequestHeaderModifier Filter cannot remove a header with an empty name",
							)
							continue
						}

						canRemHeader := true
						for _, h := range removeRequestHeaders {
							if strings.EqualFold(h, removedHeader) {
								canRemHeader = false
								break
							}
						}
						if !canRemHeader {
							continue
						}

						removeRequestHeaders = append(removeRequestHeaders, removedHeader)

					}
				}

				// Update the status if the filter failed to configure any valid headers to add/remove
				if len(addRequestHeaders) == 0 && len(removeRequestHeaders) == 0 && !emptyFilterConfig {
					parentRef.SetCondition(grpcRoute,
						v1beta1.RouteConditionAccepted,
						metav1.ConditionFalse,
						v1beta1.RouteReasonUnsupportedValue,
						"RequestHeaderModifier Filter did not provide valid configuration to add/set/remove any headers",
					)
				}
			case v1alpha2.GRPCRouteFilterResponseHeaderModifier:
				// Make sure the header modifier config actually exists
				headerModifier := filter.ResponseHeaderModifier
				if headerModifier == nil {
					break
				}
				emptyFilterConfig := true // keep track of whether the provided config is empty or not

				// Add response headers
				if headersToAdd := headerModifier.Add; headersToAdd != nil {
					if len(headersToAdd) > 0 {
						emptyFilterConfig = false
					}
					for _, addHeader := range headersToAdd {
						emptyFilterConfig = false
						if addHeader.Name == "" {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								"ResponseHeaderModifier Filter cannot add a header with an empty name",
							)
							// try to process the rest of the headers and produce a valid config.
							continue
						}
						// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
						if strings.Contains(string(addHeader.Name), "/") || strings.Contains(string(addHeader.Name), ":") {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								fmt.Sprintf("ResponseHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: %q", string(addHeader.Name)),
							)
							continue
						}
						// Check if the header is a duplicate
						headerKey := string(addHeader.Name)
						canAddHeader := true
						for _, h := range addResponseHeaders {
							if strings.EqualFold(h.Name, headerKey) {
								canAddHeader = false
								break
							}
						}

						if !canAddHeader {
							continue
						}

						newHeader := ir.AddHeader{
							Name:   headerKey,
							Append: true,
							Value:  addHeader.Value,
						}

						addResponseHeaders = append(addResponseHeaders, newHeader)
					}
				}

				// Set headers
				if headersToSet := headerModifier.Set; headersToSet != nil {
					if len(headersToSet) > 0 {
						emptyFilterConfig = false
					}
					for _, setHeader := range headersToSet {

						if setHeader.Name == "" {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								"ResponseHeaderModifier Filter cannot set a header with an empty name",
							)
							continue
						}
						// Per Gateway API specification on HTTPHeaderName, : and / are invalid characters in header names
						if strings.Contains(string(setHeader.Name), "/") || strings.Contains(string(setHeader.Name), ":") {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								fmt.Sprintf("ResponseHeaderModifier Filter cannot set headers with a '/' or ':' character in them. Header: '%s'", string(setHeader.Name)),
							)
							continue
						}

						// Check if the header to be set has already been configured
						headerKey := string(setHeader.Name)
						canAddHeader := true
						for _, h := range addResponseHeaders {
							if strings.EqualFold(h.Name, headerKey) {
								canAddHeader = false
								break
							}
						}
						if !canAddHeader {
							continue
						}
						newHeader := ir.AddHeader{
							Name:   string(setHeader.Name),
							Append: false,
							Value:  setHeader.Value,
						}

						addResponseHeaders = append(addResponseHeaders, newHeader)
					}
				}

				// Remove response headers
				// As far as Envoy is concerned, it is ok to configure a header to be added/set and also in the list of
				// headers to remove. It will remove the original header if present and then add/set the header after.
				if headersToRemove := headerModifier.Remove; headersToRemove != nil {
					if len(headersToRemove) > 0 {
						emptyFilterConfig = false
					}
					for _, removedHeader := range headersToRemove {
						if removedHeader == "" {
							parentRef.SetCondition(grpcRoute,
								v1beta1.RouteConditionAccepted,
								metav1.ConditionFalse,
								v1beta1.RouteReasonUnsupportedValue,
								"ResponseHeaderModifier Filter cannot remove a header with an empty name",
							)
							continue
						}

						canRemHeader := true
						for _, h := range removeResponseHeaders {
							if strings.EqualFold(h, removedHeader) {
								canRemHeader = false
								break
							}
						}
						if !canRemHeader {
							continue
						}

						removeResponseHeaders = append(removeResponseHeaders, removedHeader)

					}
				}

				// Update the status if the filter failed to configure any valid headers to add/remove
				if len(addResponseHeaders) == 0 && len(removeResponseHeaders) == 0 && !emptyFilterConfig {
					parentRef.SetCondition(grpcRoute,
						v1beta1.RouteConditionAccepted,
						metav1.ConditionFalse,
						v1beta1.RouteReasonUnsupportedValue,
						"ResponseHeaderModifier Filter did not provide valid configuration to add/set/remove any headers",
					)
				}
			case v1alpha2.GRPCRouteFilterExtensionRef:
				// "If a reference to a custom filter type cannot be resolved, the filter MUST NOT be skipped.
				// Instead, requests that would have been processed by that filter MUST receive a HTTP error response."
				errMsg := fmt.Sprintf("Unknown custom filter type: %s", filter.Type)
				parentRef.SetCondition(grpcRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					errMsg,
				)
				directResponse = &ir.DirectResponse{
					Body:       &errMsg,
					StatusCode: 500,
				}
			default:
				// Unsupported filters.
				errMsg := fmt.Sprintf("Unsupported filter type: %s", filter.Type)
				parentRef.SetCondition(grpcRoute,
					v1beta1.RouteConditionAccepted,
					metav1.ConditionFalse,
					v1beta1.RouteReasonUnsupportedValue,
					errMsg,
				)
				directResponse = &ir.DirectResponse{
					Body:       &errMsg,
					StatusCode: 500,
				}
			}
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
				if match.Method.Method != nil {
					irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
						Name:  ":method",
						Exact: match.Method.Method,
					})
				}
				if match.Method.Service != nil {
					// set pathmach w
					irRoute.PathMatch = &ir.StringMatch{
						Prefix: StringPtr(fmt.Sprintf("/%s", (*match.Method.Service))),
					}
				}
			}

			// Add the direct response that were created earlier to all the irRoutes
			if directResponse != nil {
				irRoute.DirectResponse = directResponse
			}
			if len(addRequestHeaders) > 0 {
				irRoute.AddRequestHeaders = addRequestHeaders
			}
			if len(removeRequestHeaders) > 0 {
				irRoute.RemoveRequestHeaders = removeRequestHeaders
			}
			if len(addResponseHeaders) > 0 {
				irRoute.AddResponseHeaders = addResponseHeaders
			}
			if len(removeResponseHeaders) > 0 {
				irRoute.RemoveResponseHeaders = removeResponseHeaders
			}
			ruleRoutes = append(ruleRoutes, irRoute)
		}

		for _, backendRef := range rule.BackendRefs {
			destination, backendWeight := buildRuleRouteDest(backendRef.BackendRef, parentRef, grpcRoute, grpcRoute.Namespace, resources)
			for _, route := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// processing any destinations for this route.
				if route.DirectResponse == nil && route.Redirect == nil {
					if destination != nil {
						route.Destinations = append(route.Destinations, destination)
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

func (t *Translator) processHTTPRouteRules(httpRoute *HTTPRouteContext, parentRef *RouteParentContext, resources *Resources) []*ir.HTTPRoute {
	var routeRoutes []*ir.HTTPRoute

	// compute matches, filters, backends
	for ruleIdx, rule := range httpRoute.Spec.Rules {
		httpFiltersContext := t.ProcessHTTPFilters(parentRef, httpRoute, rule.Filters, resources)

		// A rule is matched if any one of its matches
		// is satisfied (i.e. a logical "OR"), so generate
		// a unique Xds IR HTTPRoute per match.
		var ruleRoutes []*ir.HTTPRoute = t.processHTTPRouteRule(httpRoute, ruleIdx, httpFiltersContext, rule)

		for _, backendRef := range rule.BackendRefs {
			destination, backendWeight := t.processRouteDestination(backendRef.BackendRef, parentRef, httpRoute, resources)
			for _, route := range ruleRoutes {
				// If the route already has a direct response or redirect configured, then it was from a filter so skip
				// processing any destinations for this route.
				if route.DirectResponse == nil && route.Redirect == nil {
					if destination != nil {
						route.Destinations = append(route.Destinations, destination)
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
					Name:  queryParamMatch.Name,
					Exact: StringPtr(queryParamMatch.Value),
				})
			case v1beta1.QueryParamMatchRegularExpression:
				irRoute.QueryParamMatches = append(irRoute.QueryParamMatches, &ir.StringMatch{
					Name:      queryParamMatch.Name,
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
		ruleRoutes = append(ruleRoutes, irRoute)
	}

	return ruleRoutes
}

func (t *Translator) processGRPCRouteParentRefListener(grpcRoute *GRPCRouteContext, routeRoutes []*ir.HTTPRoute, parentRef *RouteParentContext, xdsIR XdsIRMap) bool {
	var hasHostnameIntersection bool
	for _, listener := range parentRef.listeners {
		hosts := computeHosts(grpcRoute.GetHostnames(), listener.Hostname)
		if len(hosts) == 0 {
			continue
		}
		hasHostnameIntersection = true

		var perHostRoutes []*ir.HTTPRoute
		for _, host := range hosts {
			var headerMatches []*ir.StringMatch

			// If the intersecting host is more specific than the Listener's hostname,
			// add an additional header match to all of the routes for it
			if host != "*" && (listener.Hostname == nil || string(*listener.Hostname) != host) {
				// Hostnames that are prefixed with a wildcard label (*.)
				// are interpreted as a suffix match.
				if strings.HasPrefix(host, "*.") {
					headerMatches = append(headerMatches, &ir.StringMatch{
						Name:   ":authority",
						Suffix: StringPtr(host[2:]),
					})
				} else {
					headerMatches = append(headerMatches, &ir.StringMatch{
						Name:  ":authority",
						Exact: StringPtr(host),
					})
				}
			}

			for _, routeRoute := range routeRoutes {
				hostRoute := &ir.HTTPRoute{
					Name:                  fmt.Sprintf("%s-%s", routeRoute.Name, host),
					PathMatch:             routeRoute.PathMatch,
					HeaderMatches:         append(headerMatches, routeRoute.HeaderMatches...),
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
				}
				// Don't bother copying over the weights unless the route has invalid backends.
				if routeRoute.BackendWeights.Invalid > 0 {
					hostRoute.BackendWeights = routeRoute.BackendWeights
				}
				perHostRoutes = append(perHostRoutes, hostRoute)
			}
		}

		irKey := irStringKey(listener.gateway)
		irListener := xdsIR[irKey].GetHTTPListener(irHTTPListenerName(listener))
		if irListener != nil {
			irListener.IsHTTP2 = true
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

func (t *Translator) processHTTPRouteParentRefListener(httpRoute *HTTPRouteContext, routeRoutes []*ir.HTTPRoute, parentRef *RouteParentContext, xdsIR XdsIRMap) bool {
	var hasHostnameIntersection bool

	for _, listener := range parentRef.listeners {
		hosts := computeHosts(httpRoute.GetHostnames(), listener.Hostname)
		if len(hosts) == 0 {
			continue
		}
		hasHostnameIntersection = true

		var perHostRoutes []*ir.HTTPRoute
		for _, host := range hosts {
			var headerMatches []*ir.StringMatch

			// If the intersecting host is more specific than the Listener's hostname,
			// add an additional header match to all of the routes for it
			if host != "*" && (listener.Hostname == nil || string(*listener.Hostname) != host) {
				// Hostnames that are prefixed with a wildcard label (*.)
				// are interpreted as a suffix match.
				if strings.HasPrefix(host, "*.") {
					headerMatches = append(headerMatches, &ir.StringMatch{
						Name:   ":authority",
						Suffix: StringPtr(host[2:]),
					})
				} else {
					headerMatches = append(headerMatches, &ir.StringMatch{
						Name:  ":authority",
						Exact: StringPtr(host),
					})
				}
			}

			for _, routeRoute := range routeRoutes {
				hostRoute := &ir.HTTPRoute{
					Name:      fmt.Sprintf("%s-%s", routeRoute.Name, host),
					PathMatch: routeRoute.PathMatch,
					// HeaderMatches:         append(headerMatches, routeRoute.HeaderMatches...),
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
				}
				// Don't bother copying over the weights unless the route has invalid backends.
				if routeRoute.BackendWeights.Invalid > 0 {
					hostRoute.BackendWeights = routeRoute.BackendWeights
				}
				perHostRoutes = append(perHostRoutes, hostRoute)
			}
		}

		irKey := irStringKey(listener.gateway)
		irListener := xdsIR[irKey].GetHTTPListener(irHTTPListenerName(listener))
		if irListener != nil {
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
		tlsRoute := &TLSRouteContext{TLSRoute: tls}

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
	for _, parentRef := range tlsRoute.parentRefs {
		// Skip parent refs that did not accept the route
		if !parentRef.IsAccepted(tlsRoute) {
			continue
		}

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		var routeDestinations []*ir.RouteDestination

		// compute backends
		for _, rule := range tlsRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				backendRef := backendRef
				// TODO: [v1alpha2-v1beta1] Replace with NamespaceDerefOr when TLSRoute graduates to v1beta1.
				serviceNamespace := NamespaceDerefOrAlpha(backendRef.Namespace, tlsRoute.Namespace)
				service := resources.GetService(serviceNamespace, string(backendRef.Name))

				if !t.validateBackendRef(&backendRef, parentRef, tlsRoute, resources, serviceNamespace, KindTLSRoute) {
					continue
				}

				weight := uint32(1)
				if backendRef.Weight != nil {
					weight = uint32(*backendRef.Weight)
				}

				routeDestinations = append(routeDestinations, &ir.RouteDestination{
					Host:   service.Spec.ClusterIP,
					Port:   uint32(*backendRef.Port),
					Weight: weight,
				})
			}

			// TODO handle:
			//	- no valid backend refs
			//	- sum of weights for valid backend refs is 0
			//	- returning 500's for invalid backend refs
			//	- etc.
		}

		var hasHostnameIntersection bool
		for _, listener := range parentRef.listeners {
			hosts := computeHosts(tlsRoute.GetHostnames(), listener.Hostname)
			if len(hosts) == 0 {
				continue
			}

			hasHostnameIntersection = true

			irKey := irStringKey(listener.gateway)
			containerPort := servicePortToContainerPort(int32(listener.Port))
			// Create the TCP Listener while parsing the TLSRoute since
			// the listener directly links to a routeDestination.
			irListener := &ir.TCPListener{
				Name:    irTLSListenerName(listener, tlsRoute),
				Address: "0.0.0.0",
				Port:    uint32(containerPort),
				TLS: &ir.TLSInspectorConfig{
					SNIs: hosts,
				},
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
		if parentRef.tlsRoute != nil &&
			len(parentRef.tlsRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
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
		udpRoute := &UDPRouteContext{UDPRoute: u}

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
	for _, parentRef := range udpRoute.parentRefs {
		// Skip parent refs that did not accept the route
		if !parentRef.IsAccepted(udpRoute) {
			continue
		}

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
		// TODO: [v1alpha2-v1beta1] Replace with NamespaceDerefOr when UDPRoute graduates to v1beta1.
		serviceNamespace := NamespaceDerefOrAlpha(backendRef.Namespace, udpRoute.Namespace)
		service := resources.GetService(serviceNamespace, string(backendRef.Name))

		if !t.validateBackendRef(&backendRef, parentRef, udpRoute, resources, serviceNamespace, KindUDPRoute) {
			continue
		}

		// weight is not used in udp route destinations
		routeDestinations = append(routeDestinations, &ir.RouteDestination{
			Host: service.Spec.ClusterIP,
			Port: uint32(*backendRef.Port),
		})

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
			irKey := irStringKey(listener.gateway)
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
		if accepted && parentRef.udpRoute != nil &&
			len(parentRef.udpRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
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
		tcpRoute := &TCPRouteContext{TCPRoute: tcp}

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
	for _, parentRef := range tcpRoute.parentRefs {
		// Skip parent refs that did not accept the route
		if !parentRef.IsAccepted(tcpRoute) {
			continue
		}

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
		// TODO: [v1alpha2-v1beta1] Replace with NamespaceDerefOr when TCPRoute graduates to v1beta1.
		serviceNamespace := NamespaceDerefOrAlpha(backendRef.Namespace, tcpRoute.Namespace)
		service := resources.GetService(serviceNamespace, string(backendRef.Name))

		if !t.validateBackendRef(&backendRef, parentRef, tcpRoute, resources, serviceNamespace, KindTCPRoute) {
			continue
		}

		// weight is not used in tcp route destinations
		routeDestinations = append(routeDestinations, &ir.RouteDestination{
			Host: service.Spec.ClusterIP,
			Port: uint32(*backendRef.Port),
		})

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
			irKey := irStringKey(listener.gateway)
			containerPort := servicePortToContainerPort(int32(listener.Port))
			// Create the TCP Listener while parsing the TCPRoute since
			// the listener directly links to a routeDestination.
			irListener := &ir.TCPListener{
				Name:         irTCPListenerName(listener, tcpRoute),
				Address:      "0.0.0.0",
				Port:         uint32(containerPort),
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

		// If no negative conditions have been set, the route is considered "Accepted=True".
		if accepted && parentRef.tcpRoute != nil &&
			len(parentRef.tcpRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {
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

// processRouteDestination takes a backendRef and translates it into a destination or sets error statuses and
// returns the weight for the backend so that 500 error responses can be returned for invalid backends in
// the same proportion as the backend would have otherwise received
func (t *Translator) processRouteDestination(backendRef v1beta1.BackendRef,
	parentRef *RouteParentContext,
	httpRoute *HTTPRouteContext,
	resources *Resources) (destination *ir.RouteDestination, backendWeight uint32) {

	weight := uint32(1)
	if backendRef.Weight != nil {
		weight = uint32(*backendRef.Weight)
	}

	serviceNamespace := NamespaceDerefOr(backendRef.Namespace, httpRoute.Namespace)
	service := resources.GetService(serviceNamespace, string(backendRef.Name))

	if !t.validateBackendRef(&backendRef, parentRef, httpRoute, resources, serviceNamespace, KindHTTPRoute) {
		return nil, weight
	}

	return &ir.RouteDestination{
		Host:   service.Spec.ClusterIP,
		Port:   uint32(*backendRef.Port),
		Weight: weight,
	}, weight

}

// processAllowedListenersForParentRefs finds out if the route attaches to one of our
// Gateways' listeners, and if so, gets the list of listeners that allow it to
// attach for each parentRef.
func (t *Translator) processAllowedListenersForParentRefs(routeContext RouteContext, gateways []*GatewayContext, resources *Resources) bool {
	var relevantRoute bool

	for _, parentRef := range routeContext.GetParentReferences() {
		isRelevantParentRef, selectedListeners := GetReferencedListeners(parentRef, gateways)

		// Parent ref is not to a Gateway that we control: skip it
		if !isRelevantParentRef {
			continue
		}
		relevantRoute = true

		parentRefCtx := routeContext.GetRouteParentContext(parentRef)
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
			acceptedKind := routeContext.GetRouteType()
			if listener.AllowsKind(v1beta1.RouteGroupKind{Group: GroupPtr(v1beta1.GroupName), Kind: v1beta1.Kind(acceptedKind)}) &&
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

func buildRuleRouteDest(backendRef v1beta1.BackendRef,
	parentRef *RouteParentContext,
	route RouteContext,
	routeNs string,
	resources *Resources) (destination *ir.RouteDestination, backendWeight uint32) {

	weight := uint32(1)
	if backendRef.Weight != nil {
		weight = uint32(*backendRef.Weight)
	}

	if backendRef.Group != nil && *backendRef.Group != "" {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonInvalidKind,
			"Group is invalid, only the core API group (specified by omitting the group field or setting it to an empty string) is supported",
		)
		return nil, weight
	}

	if backendRef.Kind != nil && *backendRef.Kind != KindService {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonInvalidKind,
			"Kind is invalid, only Service is supported",
		)
		return nil, weight
	}

	if backendRef.Namespace != nil && string(*backendRef.Namespace) != "" && string(*backendRef.Namespace) != routeNs {
		if !isValidCrossNamespaceRef(
			crossNamespaceFrom{
				group:     v1beta1.GroupName,
				kind:      route.GetRouteType(),
				namespace: routeNs,
			},
			crossNamespaceTo{
				group:     "",
				kind:      KindService,
				namespace: string(*backendRef.Namespace),
				name:      string(backendRef.Name),
			},
			resources.ReferenceGrants,
		) {
			parentRef.SetCondition(route,
				v1beta1.RouteConditionResolvedRefs,
				metav1.ConditionFalse,
				v1beta1.RouteReasonRefNotPermitted,
				fmt.Sprintf("Backend ref to service %s/%s not permitted by any ReferenceGrant", *backendRef.Namespace, backendRef.Name),
			)
			return nil, weight
		}
	}

	if backendRef.Port == nil {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotSpecified",
			"A valid port number corresponding to a port on the Service must be specified",
		)
		return nil, weight
	}

	service := resources.GetService(NamespaceDerefOr(backendRef.Namespace, routeNs), string(backendRef.Name))
	if service == nil {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			v1beta1.RouteReasonBackendNotFound,
			fmt.Sprintf("Service %s/%s not found", NamespaceDerefOr(backendRef.Namespace, routeNs), string(backendRef.Name)),
		)
		return nil, weight
	}

	var portFound bool
	for _, port := range service.Spec.Ports {
		if port.Port == int32(*backendRef.Port) &&
			(port.Protocol == v1.ProtocolTCP || port.Protocol == "") { // Default protocol is TCP
			portFound = true
			break
		}
	}

	if !portFound {
		parentRef.SetCondition(route,
			v1beta1.RouteConditionResolvedRefs,
			metav1.ConditionFalse,
			"PortNotFound",
			fmt.Sprintf("Port %d not found on service %s/%s", *backendRef.Port, NamespaceDerefOr(backendRef.Namespace, routeNs), string(backendRef.Name)),
		)
		return nil, weight
	}

	return &ir.RouteDestination{
		Host:   service.Spec.ClusterIP,
		Port:   uint32(*backendRef.Port),
		Weight: weight,
	}, weight

}

func isValidCrossNamespaceRef(from crossNamespaceFrom, to crossNamespaceTo, referenceGrants []*v1alpha2.ReferenceGrant) bool {
	for _, referenceGrant := range referenceGrants {
		// The ReferenceGrant must be defined in the namespace of
		// the "to" (the referent).
		if referenceGrant.Namespace != to.namespace {
			continue
		}

		// Check if the ReferenceGrant has a matching "from".
		var fromAllowed bool
		for _, refGrantFrom := range referenceGrant.Spec.From {
			if string(refGrantFrom.Namespace) == from.namespace && string(refGrantFrom.Group) == from.group && string(refGrantFrom.Kind) == from.kind {
				fromAllowed = true
				break
			}
		}
		if !fromAllowed {
			continue
		}

		// Check if the ReferenceGrant has a matching "to".
		var toAllowed bool
		for _, refGrantTo := range referenceGrant.Spec.To {
			if string(refGrantTo.Group) == to.group && string(refGrantTo.Kind) == to.kind && (refGrantTo.Name == nil || *refGrantTo.Name == "" || string(*refGrantTo.Name) == to.name) {
				toAllowed = true
				break
			}
		}
		if !toAllowed {
			continue
		}

		// If we got here, both the "from" and the "to" were allowed by this
		// reference grant.
		return true
	}

	// If we got here, no reference policy or reference grant allowed both the "from" and "to".
	return false
}
