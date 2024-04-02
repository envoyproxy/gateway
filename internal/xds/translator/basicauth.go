// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

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
// The filter is disabled by default. It is enabled on the route level.
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

		// Only generates one BasicAuth Envoy filter for each unique name.
		// For example, if there are two routes under the same gateway with the
		// same BasicAuth config, only one BasicAuth filter will be generated.
		if hcmContainsFilter(mgr, basicAuthFilterName(route.BasicAuth)) {
			continue
		}

		filter, err := buildHCMBasicAuthFilter(route.BasicAuth)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildHCMBasicAuthFilter returns a basic_auth HTTP filter from the provided IR HTTPRoute.
func buildHCMBasicAuthFilter(basicAuth *ir.BasicAuth) (*hcmv3.HttpFilter, error) {
	basicAuthProto := basicAuthConfig(basicAuth)

	if err := basicAuthProto.ValidateAll(); err != nil {
		return nil, err
	}

	basicAuthAny, err := anypb.New(basicAuthProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     basicAuthFilterName(basicAuth),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: basicAuthAny,
		},
	}, nil
}

func basicAuthFilterName(basicAuth *ir.BasicAuth) string {
	return perRouteFilterName(basicAuthFilter, basicAuth.Name)
}

func basicAuthConfig(basicAuth *ir.BasicAuth) *basicauthv3.BasicAuth {
	return &basicauthv3.BasicAuth{
		Users: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineBytes{
				InlineBytes: basicAuth.Users,
			},
		},
	}
}

// routeContainsBasicAuth returns true if BasicAuth exists for the provided route.
func routeContainsBasicAuth(irRoute *ir.HTTPRoute) bool {
	if irRoute != nil && irRoute.BasicAuth != nil {
		return true
	}
	return false
}

func (*basicAuth) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
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
	filterName := basicAuthFilterName(irRoute.BasicAuth)
	if err := enableFilterOnRoute(route, filterName); err != nil {
		return err
	}
	return nil
}
