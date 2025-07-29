// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"container/list"
	"fmt"
	"sort"
	"strconv"
	"strings"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/filters"
	"github.com/envoyproxy/gateway/internal/xds/types"
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
// The filter types that haven't native per-route support: oauth2, ext_authz.
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
// For example, the health_check filter should be placed in the first position because external load
// balancer determines whether envoy should receive traffic based on the health check result which
// only depending on the current draining state of the envoy, result should not be affected by other
// filters, or else user traffic disruption may happen.
// the fault filter should be placed in the second position because
// it doesn't rely on the functionality of other filters, and rejecting early can save computation costs
// for the remaining filters, the cors filter should be put at the third to avoid unnecessary
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
	// Important: After adding new filter types, don't forget to modify the validation rule of the EnvoyFilter type in the API
	switch {
	case isFilterType(filter, egv1a1.EnvoyFilterHealthCheck):
		order = 0
	case isFilterType(filter, egv1a1.EnvoyFilterFault):
		order = 1
	case isFilterType(filter, egv1a1.EnvoyFilterCORS):
		order = 2
	case isFilterType(filter, egv1a1.EnvoyFilterExtAuthz):
		order = 3
	case isFilterType(filter, egv1a1.EnvoyFilterAPIKeyAuth):
		order = 4
	case isFilterType(filter, egv1a1.EnvoyFilterBasicAuth):
		order = 5
	case isFilterType(filter, egv1a1.EnvoyFilterOAuth2):
		order = 6
	case isFilterType(filter, egv1a1.EnvoyFilterJWTAuthn):
		order = 7
	case isFilterType(filter, egv1a1.EnvoyFilterSessionPersistence):
		order = 8
	case isFilterType(filter, egv1a1.EnvoyFilterBuffer):
		order = 9
	case isFilterType(filter, egv1a1.EnvoyFilterLua):
		order = 10 + mustGetFilterIndex(filter.Name)
	case isFilterType(filter, egv1a1.EnvoyFilterExtProc):
		order = 100 + mustGetFilterIndex(filter.Name)
	case isFilterType(filter, egv1a1.EnvoyFilterWasm):
		order = 200 + mustGetFilterIndex(filter.Name)
	case isFilterType(filter, egv1a1.EnvoyFilterRBAC):
		order = 301
	case isFilterType(filter, egv1a1.EnvoyFilterLocalRateLimit):
		order = 302
	case isFilterType(filter, egv1a1.EnvoyFilterRateLimit):
		order = 303
	case isFilterType(filter, egv1a1.EnvoyFilterCustomResponse):
		order = 304
	case isFilterType(filter, egv1a1.EnvoyFilterCredentialInjector):
		order = 305
	case isFilterType(filter, egv1a1.EnvoyFilterCompressor):
		order = 306
	case isFilterType(filter, egv1a1.EnvoyFilterRouter):
		order = 307
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
	// Sort on name if the order is equal
	// to keep the order stable and avoiding
	// listener drains
	if o[i].order == o[j].order {
		return o[i].filter.Name < o[j].filter.Name
	}

	// Sort on order
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
func sortHTTPFilters(filters []*hcmv3.HttpFilter, filterOrder []egv1a1.FilterPosition) []*hcmv3.HttpFilter {
	// Sort the filters in the default order.
	orderedFilters := make(OrderedHTTPFilters, len(filters))
	for i := 0; i < len(filters); i++ {
		orderedFilters[i] = newOrderedHTTPFilter(filters[i])
	}

	sort.Sort(orderedFilters)

	// Use a linked list to sort the filters in the custom order.
	l := list.New()
	for i := 0; i < len(orderedFilters); i++ {
		l.PushBack(orderedFilters[i].filter)
	}

	// Sort the filters in the custom order.
	for i := 0; i < len(filterOrder); i++ {
		var (
			// The filter name in the filterOrder is the filter type.
			// For example, "envoy.filters.http.oauth2".
			filterType = string(filterOrder[i].Name)
			// currentFilters holds all the filters of the specified filter type
			// in the custom FilterOrder that we are currently processing.
			//
			// We need an array to store the filters because there may be multiple
			// filters of the same filter type for a specific HTTPRoute.
			// For example, there may be multiple wasm filters or extProc filters, for
			// different custom extensions.
			currentFilters []*list.Element
		)

		// Find all the filters for the current filter type in the custom FilterOrder.
		//
		// The real filter name is a generated name prefixed with the filter type,
		// for example,"envoy.filters.http.oauth2/securitypolicy/default/policy-for-http-route-1".
		for element := l.Front(); element != nil; element = element.Next() {
			if isFilterType(element.Value.(*hcmv3.HttpFilter), egv1a1.EnvoyFilter(filterType)) {
				currentFilters = append(currentFilters, element)
			}
		}

		// Skip if there are no filters found for the filter type in a custom order.
		if len(currentFilters) == 0 {
			continue
		}

		switch {
		// Move all the current filters before the first filter of the filter type
		// specified in the `FilterOrder.Before` field.
		case filterOrder[i].Before != nil:
			for element := l.Front(); element != nil; element = element.Next() {
				if isFilterType(element.Value.(*hcmv3.HttpFilter), *filterOrder[i].Before) {
					for _, filter := range currentFilters {
						l.MoveBefore(filter, element)
					}
					break
				}
			}
		// Move all the current filters after the last filter of the filter type
		// specified in the `FilterOrder.After` field.
		case filterOrder[i].After != nil:
			var afterFilter *list.Element
			for element := l.Front(); element != nil; element = element.Next() {
				if isFilterType(element.Value.(*hcmv3.HttpFilter), *filterOrder[i].After) {
					afterFilter = element
				}
			}
			if afterFilter != nil {
				for i := range currentFilters {
					l.MoveAfter(currentFilters[len(currentFilters)-1-i], afterFilter)
				}
			}
		}
	}

	// Collect the sorted filters.
	i := 0
	for element := l.Front(); element != nil; element = element.Next() {
		filters[i] = element.Value.(*hcmv3.HttpFilter)
		i++
	}

	return filters
}

// patchHCMWithFilters builds and appends HTTP Filters to the HTTP connection
// manager.
// Important: don't forget to set the order for newly added filters in the
// newOrderedHTTPFilter method.
func (t *Translator) patchHCMWithFilters(
	mgr *hcmv3.HttpConnectionManager,
	irListener *ir.HTTPListener,
) error {
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
		if filter.Name == string(egv1a1.EnvoyFilterRouter) {
			hasRouter = true
			break
		}
	}
	if !hasRouter {
		headerSettings := ptr.Deref(irListener.Headers, ir.HeaderSettings{})
		routerFilter, err := filters.GenerateRouterFilter(headerSettings.EnableEnvoyHeaders)
		if err != nil {
			return err
		}
		mgr.HttpFilters = append(mgr.HttpFilters, routerFilter)
	}

	// Sort the filters in the correct order.
	mgr.HttpFilters = sortHTTPFilters(mgr.HttpFilters, t.FilterOrder)
	return nil
}

// patchRouteWithPerRouteConfig appends per-route filter configuration to the
// provided route.
func patchRouteWithPerRouteConfig(
	route *routev3.Route,
	irRoute *ir.HTTPRoute,
) error {
	for _, filter := range httpFilters {
		if err := filter.patchRoute(route, irRoute); err != nil {
			return err
		}
	}

	// RateLimit filter is handled separately because it relies on the global
	// rate limit server configuration if costs are not provided.
	if err := patchRouteWithRateLimit(route, irRoute); err != nil {
		return nil
	}

	return nil
}

// isFilterType returns true if the filter is the provided filter type.
func isFilterType(filter *hcmv3.HttpFilter, filterType egv1a1.EnvoyFilter) bool {
	// Multiple filters of the same types are added to the HCM filter chain, one for each
	// route. The filter name is prefixed with the filter type, for example:
	// "envoy.filters.http.oauth2_first-route".
	return strings.HasPrefix(filter.Name, string(filterType))
}

// mustGetFilterIndex returns the index of the filter in its filter type.
func mustGetFilterIndex(filterName string) int {
	a := strings.Split(filterName, "/")
	index, err := strconv.Atoi(a[len(a)-1])
	if err != nil {
		panic(fmt.Errorf("cannot get filter index from %s :%w", filterName, err))
	}
	return index
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
