// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func (t *Translator) ProcessCustomGRPCRoutes(grpcRoutes []*v1alpha2.CustomGRPCRoute, gateways []*GatewayContext, resources *Resources, xdsIR XdsIRMap) []*CustomGRPCRouteContext {
	var relevantGRPCRoutes []*CustomGRPCRouteContext

	for _, g := range grpcRoutes {
		if g == nil {
			panic("received nil grpcroute")
		}
		grpcRoute := &CustomGRPCRouteContext{CustomGRPCRoute: g}

		// Find out if this route attaches to one of our Gateway's listeners,
		// and if so, get the list of listeners that allow it to attach for each
		// parentRef.
		relevantRoute := t.processAllowedListenersForParentRefs(grpcRoute, gateways, resources)
		if !relevantRoute {
			continue
		}

		relevantGRPCRoutes = append(relevantGRPCRoutes, grpcRoute)

		t.processCustomGRPCRouteParentRefs(grpcRoute, resources, xdsIR)
	}

	return relevantGRPCRoutes
}

func (t *Translator) processCustomGRPCRouteParentRefs(grpcRoute *CustomGRPCRouteContext, resources *Resources, xdsIR XdsIRMap) {
	for _, parentRef := range grpcRoute.parentRefs {
		// Skip parent refs that did not accept the route
		if !parentRef.IsAccepted(grpcRoute) {
			continue
		}

		// Need to compute Route rules within the parentRef loop because
		// any conditions that come out of it have to go on each RouteParentStatus,
		// not on the Route as a whole.
		routeRoutes := t.processCustomGRPCRouteRules(grpcRoute, parentRef, resources)

		var hasHostnameIntersection = t.processHTTPRouteParentRefListener(grpcRoute, routeRoutes, parentRef, xdsIR)

		if !hasHostnameIntersection {
			parentRef.SetCondition(grpcRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionFalse,
				v1beta1.RouteReasonNoMatchingListenerHostname,
				"There were no hostname intersections between the CustomGRPCRoute and this parent ref's Listener(s).",
			)
		}

		// If no negative conditions have been set, the route is considered "Accepted=True".

		if parentRef.customgrpcRoute != nil &&
			len(parentRef.customgrpcRoute.Status.Parents[parentRef.routeParentStatusIdx].Conditions) == 0 {

			parentRef.SetCondition(grpcRoute,
				v1beta1.RouteConditionAccepted,
				metav1.ConditionTrue,
				v1beta1.RouteReasonAccepted,
				"Route is accepted",
			)
		}

	}
}

func (t *Translator) processCustomGRPCRouteRules(grpcRoute *CustomGRPCRouteContext, parentRef *RouteParentContext, resources *Resources) []*ir.HTTPRoute {
	var routeRoutes []*ir.HTTPRoute

	// compute matches, filters, backends
	for ruleIdx, rule := range grpcRoute.Spec.Rules {
		httpFiltersContext := t.ProcessCustomGRPCFilters(parentRef, grpcRoute, rule.Filters, resources)

		// A rule is matched if any one of its matches
		// is satisfied (i.e. a logical "OR"), so generate
		// a unique Xds IR HTTPRoute per match.
		var ruleRoutes = t.processCustomGRPCRouteRule(grpcRoute, ruleIdx, httpFiltersContext, rule)

		for _, backendRef := range rule.BackendRefs {
			destination, backendWeight := t.processRouteDestination(backendRef.BackendRef, parentRef, grpcRoute, resources)
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

func (t *Translator) processCustomGRPCRouteRule(grpcRoute *CustomGRPCRouteContext, ruleIdx int, httpFiltersContext *HTTPFiltersContext, rule v1alpha2.GRPCRouteRule) []*ir.HTTPRoute {
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
			// TODO: support regex match type after https://github.com/kubernetes-sigs/gateway-api/issues/1746 is resolved
			switch {
			case match.Method.Service != nil && match.Method.Method != nil:
				irRoute.PathMatch = &ir.StringMatch{
					Exact: StringPtr(fmt.Sprintf("/%s/%s", *match.Method.Service, *match.Method.Method)),
				}
			case match.Method.Method != nil:
				// Use a header match since the PathMatch doesn't support Suffix matching
				irRoute.HeaderMatches = append(irRoute.HeaderMatches, &ir.StringMatch{
					Name:   ":path",
					Suffix: StringPtr(fmt.Sprintf("/%s", *match.Method.Method)),
				})
			case match.Method.Service != nil:
				irRoute.PathMatch = &ir.StringMatch{
					Prefix: StringPtr(fmt.Sprintf("/%s", *match.Method.Service)),
				}
			}
		}

		ruleRoutes = append(ruleRoutes, irRoute)
		applyHTTPFiltersContextToIRRoute(httpFiltersContext, irRoute)
	}
	return ruleRoutes
}
