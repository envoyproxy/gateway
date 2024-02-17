// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	basicauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/basic_auth/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	basicAuthFilter = "envoy.filters.http.basic_auth"
)

func init() {
	registerHTTPFilter(&basicAuth{})
}

type basicAuth struct {
}

var _ httpFilter = &basicAuth{}

// patchHCM builds and appends the basic_auth Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an basic_auth filter for each route that contains an BasicAuth config.
func (*basicAuth) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsBasicAuth(route) {
			continue
		}

		filter, err := buildHCMBasicAuthFilter(route)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return nil
}

// buildHCMBasicAuthFilter returns a basic_auth HTTP filter from the provided IR HTTPRoute.
func buildHCMBasicAuthFilter(route *ir.HTTPRoute) (*hcmv3.HttpFilter, error) {
	basicAuthProto := basicAuthConfig(route)

	if err := basicAuthProto.ValidateAll(); err != nil {
		return nil, err
	}

	basicAuthAny, err := anypb.New(basicAuthProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: basicAuthFilterName(route),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: basicAuthAny,
		},
	}, nil
}

func basicAuthFilterName(route *ir.HTTPRoute) string {
	return perRouteFilterName(basicAuthFilter, route.Name)
}

func basicAuthConfig(route *ir.HTTPRoute) *basicauthv3.BasicAuth {
	return &basicauthv3.BasicAuth{
		Users: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineBytes{
				InlineBytes: route.BasicAuth.Users,
			},
		},
	}
}

// routeContainsBasicAuth returns true if BasicAuth exists for the provided route.
func routeContainsBasicAuth(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	if irRoute != nil &&
		irRoute.BasicAuth != nil {
		return true
	}

	return false
}

func (*basicAuth) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRouteCfg patches the provided route configuration with the basicAuth filter
// if applicable.
// Note: this method disables all the basicAuth filters by default. The filter will
// be enabled per-route in the typePerFilterConfig of the route.
func (*basicAuth) patchRouteConfig(routeCfg *routev3.RouteConfiguration, irListener *ir.HTTPListener) error {
	if routeCfg == nil {
		return errors.New("route configuration is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error
	for _, route := range irListener.Routes {
		if !routeContainsBasicAuth(route) {
			continue
		}

		filterName := basicAuthFilterName(route)
		filterCfg := routeCfg.TypedPerFilterConfig

		if _, ok := filterCfg[filterName]; ok {
			// This should not happen since this is the only place where the basicAuth
			// filter is added in a route.
			errs = errors.Join(errs, fmt.Errorf(
				"route config already contains basicAuth config: %+v", route))
			continue
		}

		// Disable all the filters by default. The filter will be enabled
		// per-route in the typePerFilterConfig of the route.
		routeCfgAny, err := anypb.New(&routev3.FilterConfig{Disabled: true})
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if filterCfg == nil {
			routeCfg.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}

		routeCfg.TypedPerFilterConfig[filterName] = routeCfgAny
	}
	return errs
}

// patchRoute patches the provided route with the basicAuth config if applicable.
// Note: this method enables the corresponding basicAuth filter for the provided route.
func (*basicAuth) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.BasicAuth == nil {
		return nil
	}
	if err := enableFilterOnRoute(basicAuthFilter, route, irRoute); err != nil {
		return err
	}
	return nil
}
