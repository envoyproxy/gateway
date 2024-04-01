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
	"k8s.io/utils/ptr"

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
//     is configured with the route's per-route configuration. The filter is
//     disabled by default and is enabled on the route level.
//   - PatchRouteWithPerRouteConfig: EG enables the corresponding filter for each
//     route in the typedFilterConfig of that route.
//
// The filter types that haven't native per-route support: oauth2, basic authn, ext_authz.
// Note: The filter types that have native per-route configuration support should
// always se their own native per-route configuration.
type httpFilter interface {
	// patchHCM patches the HttpConnectionManager with the filter.
	// Note: this method may be called multiple times for the same filter, please
	// make sure to avoid duplicate additions of the same filter.
	patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error

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
// For example, the fault filter should be placed in the first position because
// it doesn't rely on the functionality of other filters, and rejecting early can save computation costs
// for the remaining filters, the cors filter should be put at the second to avoid unnecessary
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
	// When the fault filter is configured to be at the first, the computation of
	// the remaining filters is skipped when rejected early
	switch {
	case filter.Name == wellknown.Fault:
		order = 1
	case filter.Name == wellknown.CORS:
		order = 2
	case isFilterType(filter, extAuthFilter):
		order = 3
	case isFilterType(filter, basicAuthFilter):
		order = 4
	case isFilterType(filter, oauth2Filter):
		order = 5
	case filter.Name == jwtAuthn:
		order = 6
	case filter.Name == localRateLimitFilter:
		order = 7
	case filter.Name == wellknown.HTTPRateLimit:
		order = 8
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

	// Add the router filter if it doesn't exist.
	hasRouter := false
	for _, filter := range mgr.HttpFilters {
		if filter.Name == wellknown.Router {
			hasRouter = true
			break
		}
	}
	if !hasRouter {
		headerSettings := ptr.Deref(irListener.Headers, ir.HeaderSettings{})
		mgr.HttpFilters = append(mgr.HttpFilters, filters.GenerateRouterFilter(headerSettings.EnableEnvoyHeaders))
	}

	// Sort the filters in the correct order.
	mgr.HttpFilters = sortHTTPFilters(mgr.HttpFilters)
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

// isFilterType returns true if the filter is the provided filter type.
func isFilterType(filter *hcmv3.HttpFilter, filterType string) bool {
	// Multiple filters of the same types are added to the HCM filter chain, one for each
	// route. The filter name is prefixed with the filter type, for example:
	// "envoy.filters.http.oauth2_first-route".
	return strings.HasPrefix(filter.Name, filterType)
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
