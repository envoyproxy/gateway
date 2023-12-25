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

	"github.com/envoyproxy/gateway/internal/xds/filters"
	"github.com/envoyproxy/gateway/internal/xds/types"

	"github.com/envoyproxy/gateway/internal/ir"
)

var httpFilters []httpFilter

// registerHTTPFilter registers the provided HTTP filter.
func registerHTTPFilter(filter httpFilter) {
	httpFilters = append(httpFilters, filter)
}

// httpFilter is the interface for all the HTTP filters.
//
// There are two ways to support per-route configuration for an HTTP filter:
// - For the filters with native per-route configuration support:
//   - patchHCM: EG adds the filter to the HCM filter chain only once.
//   - patchRoute: EG adds the filter's native per-route configuration to each
//     route's typedFilterConfig.
//
// - For the filters without native per-route configuration support:
//   - patchHCM: EG adds a filter for each route in the HCM filter chain, the
//     filter name is prefixed with the filter's type name, for example,
//     "envoy.filters.http.oauth2", and suffixed with the route name. Each filter
//     is configured with the route's per-route configuration.
//   - patchRouteConfig: EG disables all the filters of this type in the
//     typedFilterConfig of the route config.
//   - PatchRouteWithPerRouteConfig: EG enables the corresponding filter for each
//     route in the typedFilterConfig of that route.
//
// The filter types that haven't native per-route support: oauth2, basic authn
// Note: The filter types that have native per-route configuration support should
// always se their own native per-route configuration.
type httpFilter interface {
	// patchHCM patches the HttpConnectionManager with the filter.
	patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error

	// patchRouteConfig patches the provided RouteConfiguration with a filter's
	// RouteConfiguration level configuration.
	patchRouteConfig(rc *routev3.RouteConfiguration, irListener *ir.HTTPListener) error

	// patchRoute patches the provide Route with a filter's Route level configuration.
	patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error

	// patchResources adds all the other needed resources referenced by this
	// filter to the resource version table.
	// for example:
	// - a jwt filter needs to add the cluster for the jwks.
	// - an oidc filter needs to add the cluster for token endpoint and the secret
	//   for the oauth2 client secret and the hmac secret.
	patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error
}

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
	case filter.Name == basicAuthFilter:
		order = 2
	case isOAuth2Filter(filter):
		order = 3
	case filter.Name == jwtAuthn:
		order = 4
	case filter.Name == wellknown.Fault:
		order = 5
	case filter.Name == localRateLimitFilter:
		order = 6
	case filter.Name == wellknown.HTTPRateLimit:
		order = 7
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
	//
	// Important: don't forget to set the order for new filters in the
	// newOrderedHTTPFilter method.
	for _, filter := range httpFilters {
		if err := filter.patchHCM(mgr, irListener); err != nil {
			return err
		}
	}

	// RateLimit filter is handled separately because it relies on the global
	// rate limit server configuration.
	t.patchHCMWithRateLimit(mgr, irListener)

	// Add the router filter
	mgr.HttpFilters = append(mgr.HttpFilters, filters.GenerateRouterFilter(irListener.SuppressEnvoyHeaders))

	// Sort the filters in the correct order.
	mgr.HttpFilters = sortHTTPFilters(mgr.HttpFilters)
	return nil
}

// patchRouteCfgWithPerRouteConfig appends per-route filter configurations to the
// provided listener's RouteConfiguration.
// This method is used to disable the filters without native per-route support.
// The disabled filters will be enabled by route in the patchRouteWithPerRouteConfig
// method.
func patchRouteCfgWithPerRouteConfig(
	routeCfg *routev3.RouteConfiguration,
	irListener *ir.HTTPListener) error {
	// Only supports the oauth2 filter for now, other filters will be added later.
	for _, filter := range httpFilters {
		if err := filter.patchRouteConfig(routeCfg, irListener); err != nil {
			return err
		}
	}
	return nil
}

// patchRouteWithPerRouteConfig appends per-route filter configuration to the
// provided route.
func patchRouteWithPerRouteConfig(
	route *routev3.Route,
	irRoute *ir.HTTPRoute) error {

	for _, filter := range httpFilters {
		if err := filter.patchRoute(route, irRoute); err != nil {
			return err
		}
	}

	// RateLimit filter is handled separately because it relies on the global
	// rate limit server configuration.
	if err :=
		patchRouteWithRateLimit(route.GetRoute(), irRoute); err != nil {
		return nil
	}

	return nil
}

// isOAuth2Filter returns true if the provided filter is an OAuth2 filter.
func isOAuth2Filter(filter *hcmv3.HttpFilter) bool {
	// Multiple oauth2 filters are added to the HCM filter chain, one for each
	// route. The oauth2 filter name is prefixed with "envoy.filters.http.oauth2".
	return strings.HasPrefix(filter.Name, oauth2Filter)
}

// patchResources adds all the other needed resources referenced by this
// filter to the resource version table.
// for example:
// - a jwt filter needs to add the cluster for the jwks.
// - an oidc filter needs to add the secret for the oauth2 client secret.
func patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	for _, filter := range httpFilters {
		if err := filter.patchResources(tCtx, routes); err != nil {
			return err
		}
	}
	return nil
}
