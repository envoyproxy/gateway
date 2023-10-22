// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	corsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
)

// patchHCMWithCorsFilter builds and appends the Cors Filter to the HTTP
// Connection Manager if applicable.
func patchHCMWithCorsFilter(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsCors(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == wellknown.CORS {
			return nil
		}
	}

	corsFilter, err := buildHCMCorsFilter()
	if err != nil {
		return err
	}

	// Ensure the cors filter is the first one in the filter chain.
	mgr.HttpFilters = append([]*hcmv3.HttpFilter{corsFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMCorsFilter returns a Cors filter from the provided IR listener.
func buildHCMCorsFilter() (*hcmv3.HttpFilter, error) {
	corsProto := &corsv3.Cors{}

	corsAny, err := anypb.New(corsProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: wellknown.CORS,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: corsAny,
		},
	}, nil
}

// listenerContainsCors returns true if the provided listener has Cors
// policies attached to its routes.
func listenerContainsCors(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if route.Cors != nil {
			return true
		}
	}

	return false
}

// patchRouteWithCorsConfig patches the provided route with the Cors config if
// applicable.
func patchRouteWithCorsConfig(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Cors == nil {
		return nil
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[wellknown.CORS]; ok {
		// This should not happen since this is the only place where the cors
		// filter is added in a route.
		return fmt.Errorf("route already contains cors config: %+v", route)
	}

	var (
		allowOrigins     []*matcherv3.StringMatcher
		allowMethods     string
		allowHeaders     string
		exposeHeaders    string
		maxAge           string
		allowCredentials *wrappers.BoolValue
	)

	//nolint:gocritic

	for _, origin := range irRoute.Cors.AllowOrigins {
		allowOrigins = append(allowOrigins, buildXdsStringMatcher(origin))
	}

	allowMethods = strings.Join(irRoute.Cors.AllowMethods, ", ")
	allowHeaders = strings.Join(irRoute.Cors.AllowHeaders, ", ")
	exposeHeaders = strings.Join(irRoute.Cors.ExposeHeaders, ", ")
	maxAge = strconv.Itoa(int(irRoute.Cors.MaxAge.Seconds()))

	routeCfgProto := &corsv3.CorsPolicy{
		AllowOriginStringMatch: allowOrigins,
		AllowMethods:           allowMethods,
		AllowHeaders:           allowHeaders,
		ExposeHeaders:          exposeHeaders,
		MaxAge:                 maxAge,
		AllowCredentials:       allowCredentials,
	}

	routeCfgAny, err := anypb.New(routeCfgProto)
	if err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[wellknown.CORS] = routeCfgAny

	return nil
}
