// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	dfpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/dynamic_forward_proxy/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&dynamicForwardProxy{})
}

type dynamicForwardProxy struct{}

var _ httpFilter = &dynamicForwardProxy{}

// patchHCM updates the HTTPConnectionManager with a Basic Auth HTTP filter for routes requiring authentication.
// It scans through all routes in the provided HTTPListener, and if a route has a BasicAuth configuration,
// it checks for the presence of a corresponding filter in the manager. If the filter is not already present,
// it generates and appends a new Basic Auth filter.
// The function returns an error if either the HTTPConnectionManager or HTTPListener is nil, or if an error occurs
// during the filter creation process.
func (*dynamicForwardProxy) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error

	needsDFPFilter := false
	for _, route := range irListener.Routes {
		if route.IsDynamicResolverRoute() && requiresDFPFilter(route) {
			needsDFPFilter = true
			break
		}
	}

	if needsDFPFilter {
		// Only generates one DFP filter, as its config is always empty
		// and only per-route config may change depending for routes that require host rewrites
		if hcmContainsFilter(mgr, dynamicForwardProxyFilterName()) {
			return nil
		}

		filter, err := buildDynamicForwardProxyFilter()
		if err != nil {
			errs = errors.Join(errs, err)
			return errs
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildDynamicForwardProxyFilter returns a basic_auth HTTP filter from the provided IR HTTPRoute.
func buildDynamicForwardProxyFilter() (*hcmv3.HttpFilter, error) {
	var (
		dfpProto *dfpv3.FilterConfig
		dfpAny   *anypb.Any
		err      error
	)

	dfpProto = &dfpv3.FilterConfig{}

	if dfpAny, err = proto.ToAnyWithValidation(dfpProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: dynamicForwardProxyFilterName(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: dfpAny,
		},
		Disabled: true,
	}, nil
}

func dynamicForwardProxyFilterName() string {
	return perRouteFilterName(egv1a1.EnvoyFilterDynamicForwardProxy, "shared")
}

func (*dynamicForwardProxy) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the basicAuth config if applicable.
// Note: this method overwrites the HCM level filter config with the per route filter config.
func (*dynamicForwardProxy) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	isDynamicResolver := irRoute.IsDynamicResolverRoute()
	requiresDFPHostRewrite := requiresDFPFilter(irRoute)
	if !isDynamicResolver || !requiresDFPHostRewrite {
		return nil
	}

	var (
		perFilterCfg map[string]*anypb.Any
		basicAuthAny *anypb.Any
		err          error
	)
	filterName := dynamicForwardProxyFilterName()
	perFilterCfg = route.GetTypedPerFilterConfig()
	if _, ok := perFilterCfg[filterName]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf("route already contains filter config: %s, %+v",
			egv1a1.EnvoyFilterDynamicForwardProxy.String(), route)
	}

	// Overwrite the HCM level filter config with the per route filter config.
	dfpProto := dynamicForwardProxyPerRouteConfig(irRoute)
	if basicAuthAny, err = proto.ToAnyWithValidation(dfpProto); err != nil {
		return err
	}

	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[filterName] = basicAuthAny

	return nil
}

// DFP filter is needed in filter chain (in additioan to DFP cluster) only if
// certain features are required, e.g. host header rewrite
func requiresDFPFilter(irRoute *ir.HTTPRoute) bool {
	return irRoute.URLRewrite != nil && irRoute.URLRewrite.Host != nil &&
		(irRoute.URLRewrite.Host.Name != nil || irRoute.URLRewrite.Host.Header != nil)
}

func dynamicForwardProxyPerRouteConfig(irRoute *ir.HTTPRoute) *dfpv3.PerRouteConfig {
	dfpPerRouteConfig := &dfpv3.PerRouteConfig{}
	if irRoute.URLRewrite != nil && irRoute.URLRewrite.Host != nil {
		if irRoute.URLRewrite.Host.Name != nil {
			dfpPerRouteConfig.HostRewriteSpecifier = &dfpv3.PerRouteConfig_HostRewriteLiteral{
				HostRewriteLiteral: *irRoute.URLRewrite.Host.Name,
			}
		}
		if irRoute.URLRewrite.Host.Header != nil {
			dfpPerRouteConfig.HostRewriteSpecifier = &dfpv3.PerRouteConfig_HostRewriteHeader{
				HostRewriteHeader: *irRoute.URLRewrite.Host.Header,
			}
		}
	}
	return dfpPerRouteConfig
}
