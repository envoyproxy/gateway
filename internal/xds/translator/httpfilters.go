// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"sort"
	"strings"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/envoyproxy/gateway/internal/ir"
	xdsfilters "github.com/envoyproxy/gateway/internal/xds/filters"
)

type OrderedHTTPFilter struct {
	filter *hcmv3.HttpFilter
	order  int
}

type OrderedHTTPFilters []*OrderedHTTPFilter

// newOrderedHTTPFilter gives each HTTP filter a rational order.
// This is needed because the order of the filters is important.
// For example, the cors filter should be put at the first to avoid unnecessary
// processing of other filters for unauthorized cross-region access.
// The router filter must be the last one since it's a terminal filter.
//
// Important: please modify this method and set the order for the new filter
// when adding a new filter in the HCM filter chain.
// If the order is not explicitly specified in this method, a filter will be set
// a default order 50.
func newOrderedHTTPFilter(filter *hcmv3.HttpFilter) *OrderedHTTPFilter {
	order := 50

	// Set a rational order for all the filters.
	switch {
	case filter.Name == wellknown.CORS:
		order = 1
	case isOAuth2Filter(filter):
		order = 2
	case filter.Name == jwtAuthnFilter:
		order = 3
	case filter.Name == wellknown.HTTPRateLimit:
		order = 4
	case filter.Name == wellknown.Router:
		order = 100
	}

	return &OrderedHTTPFilter{
		filter: filter,
		order:  order,
	}
}

// sort.Interface implementation.
func (o OrderedHTTPFilters) Len() int {
	return len(o)
}

func (o OrderedHTTPFilters) Less(i, j int) bool {
	return o[i].order < o[j].order
}

func (o OrderedHTTPFilters) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

// sortHTTPFilters sorts the HTTP filters in the correct order.
// This is needed because the order of the filters is important.
// For example, the cors filter should be put at the first to avoid unnecessary
// processing of other filters for unauthorized cross-region access.
// The router filter must be the last one since it's a terminal filter.
func sortHTTPFilters(filters []*hcmv3.HttpFilter) []*hcmv3.HttpFilter {
	orderedFilters := make(OrderedHTTPFilters, len(filters))
	for i := 0; i < len(filters); i++ {
		orderedFilters[i] = newOrderedHTTPFilter(filters[i])
	}
	sort.Sort(orderedFilters)

	for i := 0; i < len(filters); i++ {
		filters[i] = orderedFilters[i].filter
	}
	return filters
}

// patchHCMWithFilters builds and appends HTTP Filters to the HTTP connection
// manager.
// Important: don't forget to set the order for newly added filters in the
// newOrderedHTTPFilter method.
func (t *Translator) patchHCMWithFilters(
	mgr *hcmv3.HttpConnectionManager,
	irListener *ir.HTTPListener) error {
	// The order of filter patching is not relevant here.
	// All the filters will be sorted in correct order after the patching is done.
	// Important: don't forget to set the order for new filters in the
	// newOrderedHTTPFilter method.
	// TODO: Make this a generic interface for all API Gateway features.
	//       https://github.com/envoyproxy/gateway/issues/882
	t.patchHCMWithRateLimit(mgr, irListener)

	// Add the cors filter, if needed
	if err := patchHCMWithCORSFilter(mgr, irListener); err != nil {
		return err
	}

	// Add the jwt authn filter, if needed.
	if err := patchHCMWithJWTAuthnFilter(mgr, irListener); err != nil {
		return err
	}

	// Add oauth2 filters, if needed.
	if err := patchHCMWithOAuth2Filters(mgr, irListener); err != nil {
		return err
	}

	// Add the router filter
	mgr.HttpFilters = append(mgr.HttpFilters, xdsfilters.HTTPRouter)

	// Sort the filters in the correct order.
	mgr.HttpFilters = sortHTTPFilters(mgr.HttpFilters)
	return nil
}

// patchRouteCfgWithPerRouteConfig appends per-route filter configurations to the
// route config.
// This is a generic way to add per-route filter configurations for all filters
// that has none-native per-route configuration support.
//   - For the filter type that without native per-route configuration support, EG
//     adds a filter for each route in the HCM filter chain.
//   - patchRouteCfgWithPerRouteConfig disables all the filters in the
//     typedFilterConfig of the route config.
//   - PatchRouteWithPerRouteConfig enables the corresponding oauth2 filter for each
//     route in the typedFilterConfig of the route.
//
// The filter types that have non-native per-route support: oauth2, basic authn
// Note: The filter types that have native per-route configuration support should
// use their own native per-route configuration.
func patchRouteCfgWithPerRouteConfig(
	routeCfg *routev3.RouteConfiguration,
	irListener *ir.HTTPListener) error {
	// Only supports the oauth2 filter for now, other filters will be added later.
	return patchRouteCfgWithOAuth2Filter(routeCfg, irListener)
}

// patchRouteWithPerRouteConfig appends per-route filter configurations to the route.
func patchRouteWithPerRouteConfig(
	route *routev3.Route,
	irRoute *ir.HTTPRoute) error {
	// TODO: Convert this into a generic interface for API Gateway features.
	//       https://github.com/envoyproxy/gateway/issues/882
	if err :=
		patchRouteWithRateLimit(route.GetRoute(), irRoute); err != nil {
		return nil
	}

	// Add the cors per route config to the route, if needed.
	if err := patchRouteWithCORS(route, irRoute); err != nil {
		return err
	}

	// Add the jwt per route config to the route, if needed.
	if err := patchRouteWithJWT(route, irRoute); err != nil {
		return err
	}

	// Add the oauth2 per route config to the route, if needed.
	if err := patchRouteWithOAuth2(route, irRoute); err != nil {
		return err
	}

	return nil
}

// isOAuth2Filter returns true if the provided filter is an OAuth2 filter.
func isOAuth2Filter(filter *hcmv3.HttpFilter) bool {
	// Multiple oauth2 filters are added to the HCM filter chain, one for each
	// route. The oauth2 filter name is prefixed with "envoy.filters.http.oauth2".
	return strings.HasPrefix(filter.Name, oauth2Filter)
}
