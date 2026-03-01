// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	dmconfigv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/dynamic_modules/v3"
	dmfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/dynamic_modules/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&dynamicModule{})
}

type dynamicModule struct{}

var _ httpFilter = &dynamicModule{}

// patchHCM builds and appends the dynamic module filters to the HTTP Connection Manager
// if applicable, and they do not already exist.
// Note: this method creates a filter for each route that contains a dynamic module config.
// The filter is disabled by default and enabled on the route level.
func (*dynamicModule) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsDynamicModule(route) {
			continue
		}
		for _, dm := range route.EnvoyExtensions.DynamicModules {
			if hcmContainsFilter(mgr, dynamicModuleFilterName(&dm)) {
				continue
			}
			filter, err := buildHCMDynamicModuleFilter(&dm)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}

	return errs
}

// buildHCMDynamicModuleFilter returns a dynamic module HTTP filter from the provided IR DynamicModule.
func buildHCMDynamicModuleFilter(dm *ir.DynamicModule) (*hcmv3.HttpFilter, error) {
	dmProto, err := dynamicModuleConfig(dm)
	if err != nil {
		return nil, err
	}

	dmAny, err := anypb.New(dmProto)
	if err != nil {
		return nil, err
	}

	// All dynamic module filters for all Routes are aggregated on HCM and disabled by default.
	// Per-route config is used to enable the relevant filters on appropriate routes.
	return &hcmv3.HttpFilter{
		Name:     dynamicModuleFilterName(dm),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: dmAny,
		},
	}, nil
}

func dynamicModuleFilterName(dm *ir.DynamicModule) string {
	return perRouteFilterName(egv1a1.EnvoyFilterDynamicModules, dm.Name)
}

func dynamicModuleConfig(dm *ir.DynamicModule) (*dmfilterv3.DynamicModuleFilter, error) {
	filterConfig := &dmfilterv3.DynamicModuleFilter{
		DynamicModuleConfig: &dmconfigv3.DynamicModuleConfig{
			Name:         dm.ModuleName,
			DoNotClose:   dm.DoNotClose,
			LoadGlobally: dm.LoadGlobally,
		},
		FilterName:     dm.FilterName,
		TerminalFilter: dm.TerminalFilter,
	}

	if dm.Config != nil && dm.Config.Raw != nil {
		configAny, err := anypb.New(wrapperspb.String(string(dm.Config.Raw)))
		if err != nil {
			return nil, err
		}
		filterConfig.FilterConfig = configAny
	}

	return filterConfig, nil
}

// routeContainsDynamicModule returns true if DynamicModules exist for the provided route.
func routeContainsDynamicModule(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}
	return irRoute.EnvoyExtensions != nil && len(irRoute.EnvoyExtensions.DynamicModules) > 0
}

// patchResources is a no-op for dynamic modules: they are loaded from the local filesystem.
func (*dynamicModule) patchResources(_ *types.ResourceVersionTable, _ []*ir.HTTPRoute) error {
	return nil
}

// patchRoute enables the corresponding dynamic module filter for the provided route.
func (*dynamicModule) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.EnvoyExtensions == nil {
		return nil
	}

	for _, dm := range irRoute.EnvoyExtensions.DynamicModules {
		filterName := dynamicModuleFilterName(&dm)
		if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
			Config: &anypb.Any{},
		}); err != nil {
			return err
		}
	}
	return nil
}
