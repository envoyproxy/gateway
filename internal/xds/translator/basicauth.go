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

// patchHCM builds and appends the basic_auth Filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
func (*basicAuth) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}
	if hcmContainsFilter(mgr, basicAuthFilter) {
		return nil
	}

	var (
		irBasicAuth *ir.BasicAuth
		filter      *hcmv3.HttpFilter
		err         error
	)

	for _, route := range irListener.Routes {
		if route.Security != nil && route.Security.BasicAuth != nil {
			irBasicAuth = route.Security.BasicAuth
			break
		}
	}
	if irBasicAuth == nil {
		return nil
	}

	// We use the first route that contains the basicAuth config to build the filter.
	// The HCM-level filter config doesn't matter since it is overridden at the route level.
	if filter, err = buildHCMBasicAuthFilter(irBasicAuth); err != nil {
		return err
	}
	mgr.HttpFilters = append(mgr.HttpFilters, filter)
	return err
}

// buildHCMBasicAuthFilter returns a basic_auth HTTP filter from the provided IR HTTPRoute.
func buildHCMBasicAuthFilter(basicAuth *ir.BasicAuth) (*hcmv3.HttpFilter, error) {
	var (
		basicAuthProto *basicauthv3.BasicAuth
		basicAuthAny   *anypb.Any
		err            error
	)

	basicAuthProto = &basicauthv3.BasicAuth{
		Users: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineBytes{
				InlineBytes: basicAuth.Users,
			},
		},
	}
	if err = basicAuthProto.ValidateAll(); err != nil {
		return nil, err
	}
	if basicAuthAny, err = anypb.New(basicAuthProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: basicAuthFilter,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: basicAuthAny,
		},
		Disabled: true,
	}, nil
}

func (*basicAuth) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the basicAuth config if applicable.
// Note: this method overwrites the HCM level filter config with the per route filter config.
func (*basicAuth) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.BasicAuth == nil {
		return nil
	}

	var (
		perFilterCfg map[string]*anypb.Any
		basicAuthAny *anypb.Any
		err          error
	)

	perFilterCfg = route.GetTypedPerFilterConfig()
	if _, ok := perFilterCfg[basicAuthFilter]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf("route already contains filter config: %s, %+v",
			basicAuthFilter, route)
	}

	// Overwrite the HCM level filter config with the per route filter config.
	basicAuthProto := basicAuthPerRouteConfig(irRoute.Security.BasicAuth)
	if err = basicAuthProto.ValidateAll(); err != nil {
		return err
	}

	if basicAuthAny, err = anypb.New(basicAuthProto); err != nil {
		return err
	}

	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[basicAuthFilter] = basicAuthAny

	return nil
}

func basicAuthPerRouteConfig(basicAuth *ir.BasicAuth) *basicauthv3.BasicAuthPerRoute {
	return &basicauthv3.BasicAuthPerRoute{
		Users: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineBytes{
				InlineBytes: basicAuth.Users,
			},
		},
	}
}
