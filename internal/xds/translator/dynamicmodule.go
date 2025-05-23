// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	dynamicmodulesconfigv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/dynamic_modules/v3"
	dynamicmodulesv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/dynamic_modules/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&dynamicModule{})
}

type dynamicModule struct{}

var _ httpFilter = &dynamicModule{}

// patchHCM builds and appends the dynamic module Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates a dynamic module filter for each route that contains a dynamic module config.
// The filter is disabled by default. It is enabled on the route level.
func (*dynamicModule) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	// Check if any route has dynamic module extensions
	hasDynamicModule := false
	for _, route := range irListener.Routes {
		if route.EnvoyExtensions != nil && len(route.EnvoyExtensions.DynamicModules) > 0 {
			hasDynamicModule = true
			break
		}
	}

	if !hasDynamicModule {
		return nil
	}

	// Create a map to track which dynamic modules have been added to the HCM
	dynamicModuleMap := make(map[string]bool)

	// Add dynamic module filters for each route that has dynamic module extensions
	for _, route := range irListener.Routes {
		if route.EnvoyExtensions == nil || len(route.EnvoyExtensions.DynamicModules) == 0 {
			continue
		}

		for _, dm := range route.EnvoyExtensions.DynamicModules {
			filterName := getDynamicModuleFilterName(dm)
			if dynamicModuleMap[filterName] {
				continue
			}

			dynamicModuleMap[filterName] = true

			// Create the dynamic module filter
			dynamicModuleFilter, err := createDynamicModuleFilter(dm)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			// Add the filter to the HCM
			httpFilter := &hcmv3.HttpFilter{
				Name: filterName,
				ConfigType: &hcmv3.HttpFilter_TypedConfig{
					TypedConfig: dynamicModuleFilter,
				},
				// Disable the filter by default, it will be enabled on the route level
				Disabled: true,
			}

			mgr.HttpFilters = append(mgr.HttpFilters, httpFilter)
		}
	}

	return errs
}

// patchRoute patches the provided route with the dynamic module config if applicable.
// Note: this method enables the corresponding dynamic module filter for the provided route.
func (*dynamicModule) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
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
		filterName := getDynamicModuleFilterName(dm)

		// Enable the filter on the route
		if route.TypedPerFilterConfig == nil {
			route.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}

		// Create an empty struct for the per-route config
		emptyStruct, err := structpb.NewStruct(map[string]interface{}{})
		if err != nil {
			return err
		}

		emptyConfigAny, err := anypb.New(emptyStruct)
		if err != nil {
			return err
		}

		route.TypedPerFilterConfig[filterName] = emptyConfigAny
	}
	return nil
}

// patchResources adds all the other needed resources referenced by this
// filter to the resource version table.
func (*dynamicModule) patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	// Dynamic modules don't need any additional resources
	return nil
}

// createDynamicModuleFilter creates a dynamic module filter for the provided dynamic module.
func createDynamicModuleFilter(dm ir.DynamicModule) (*anypb.Any, error) {
	// First, create the DynamicModuleConfig for the module itself
	dynamicModuleConfigProto := &dynamicmodulesconfigv3.DynamicModuleConfig{
		Name:       dm.Module,
		DoNotClose: dm.DoNotClose,
	}

	// Create the dynamic module filter
	dynamicModuleFilter := &dynamicmodulesv3.DynamicModuleFilter{
		DynamicModuleConfig: dynamicModuleConfigProto,
		FilterName:          dm.ExtensionName,
	}

	// Set the filter config if specified
	if dm.ExtensionConfig != nil {
		// Create a struct for the config
		configStruct, err := structpb.NewStruct(map[string]interface{}{
			"config": *dm.ExtensionConfig,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create config struct: %w", err)
		}

		// Convert the struct to Any
		configAny, err := anypb.New(configStruct)
		if err != nil {
			return nil, fmt.Errorf("failed to convert config to Any: %w", err)
		}

		dynamicModuleFilter.FilterConfig = configAny
	}

	return anypb.New(dynamicModuleFilter)
}

// getDynamicModuleFilterName returns the filter name for the provided dynamic module.
func getDynamicModuleFilterName(dm ir.DynamicModule) string {
	return perRouteFilterName(egv1a1.EnvoyFilterDynamicModules, dm.ExtensionName)
}
