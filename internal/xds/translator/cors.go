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
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&cors{})
}

type cors struct {
}

var _ httpFilter = &cors{}

// patchHCM builds and appends the CORS Filter to the HTTP Connection Manager if
// applicable.
func (*cors) patchHCM(
	mgr *hcmv3.HttpConnectionManager,
	irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsCORS(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == wellknown.CORS {
			return nil
		}
	}

	corsFilter, err := buildHCMCORSFilter()
	if err != nil {
		return err
	}

	// Ensure the CORS filter is the first one in the filter chain.
	mgr.HttpFilters = append([]*hcmv3.HttpFilter{corsFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMCORSFilter returns a CORS filter from the provided IR listener.
func buildHCMCORSFilter() (*hcmv3.HttpFilter, error) {
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

// listenerContainsCORS returns true if the provided listener has CORS
// policies attached to its routes.
func listenerContainsCORS(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if route.Security != nil && route.Security.CORS != nil {
			return true
		}
	}

	return false
}

// patchRoute patches the provided route with the CORS config if applicable.
func (*cors) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.CORS == nil {
		return nil
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[wellknown.CORS]; ok {
		// This should not happen since this is the only place where the CORS
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
		c                = irRoute.Security.CORS
	)

	//nolint:gocritic

	for _, origin := range c.AllowOrigins {
		allowOrigins = append(allowOrigins, buildXdsStringMatcher(origin))
	}

	allowMethods = strings.Join(c.AllowMethods, ", ")
	allowHeaders = strings.Join(c.AllowHeaders, ", ")
	exposeHeaders = strings.Join(c.ExposeHeaders, ", ")
	if c.MaxAge != nil {
		maxAge = strconv.Itoa(int(c.MaxAge.Seconds()))
	}
	allowCredentials = &wrappers.BoolValue{Value: c.AllowCredentials}

	routeCfgProto := &corsv3.CorsPolicy{
		AllowOriginStringMatch:       allowOrigins,
		AllowMethods:                 allowMethods,
		AllowHeaders:                 allowHeaders,
		ExposeHeaders:                exposeHeaders,
		MaxAge:                       maxAge,
		AllowCredentials:             allowCredentials,
		ForwardNotMatchingPreflights: &wrappers.BoolValue{Value: false},
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

func (c *cors) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}
