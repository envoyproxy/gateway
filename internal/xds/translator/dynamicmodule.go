// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	dmconfigv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/dynamic_modules/v3"
	dmfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/dynamic_modules/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
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

	addFilters := func(dms []ir.DynamicModule) {
		for i := range dms {
			dm := &dms[i]
			if hcmContainsFilter(mgr, dynamicModuleFilterName(dm)) {
				continue
			}
			filter, err := buildHCMDynamicModuleFilter(dm)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}

	// Listener-scoped DynamicModules are enabled at VirtualHost scope; route-scoped
	// DynamicModules are enabled per route. Both need their (disabled by default) filter
	// present on the HCM.
	if listenerContainsDynamicModule(irListener) {
		addFilters(irListener.EnvoyExtensions.DynamicModules)
	}
	for _, route := range irListener.Routes {
		if !routeContainsDynamicModule(route) {
			continue
		}
		addFilters(route.EnvoyExtensions.DynamicModules)
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
	moduleSource, err := dynamicModuleSource(dm)
	if err != nil {
		return nil, err
	}

	dmConfig := &dmconfigv3.DynamicModuleConfig{
		DoNotClose:   dm.DoNotClose,
		LoadGlobally: dm.LoadGlobally,
		Module:       moduleSource,
	}

	filterConfig := &dmfilterv3.DynamicModuleFilter{
		DynamicModuleConfig: dmConfig,
		FilterName:          dm.FilterName,
		TerminalFilter:      dm.TerminalFilter,
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

func dynamicModuleSource(dm *ir.DynamicModule) (*corev3.AsyncDataSource, error) {
	if dm.Remote != nil {
		uc, err := url2Cluster(dm.Remote.URL)
		if err != nil {
			return nil, err
		}

		return &corev3.AsyncDataSource{
			Specifier: &corev3.AsyncDataSource_Remote{
				Remote: &corev3.RemoteDataSource{
					HttpUri: &corev3.HttpUri{
						Uri: dm.Remote.URL,
						HttpUpstreamType: &corev3.HttpUri_Cluster{
							Cluster: uc.name,
						},
						Timeout: durationpb.New(defaultExtServiceRequestTimeout),
					},
					Sha256: dm.Remote.SHA256,
				},
			},
		}, nil
	}

	return &corev3.AsyncDataSource{
		Specifier: &corev3.AsyncDataSource_Local{
			Local: &corev3.DataSource{
				Specifier: &corev3.DataSource_Filename{
					Filename: dm.Path,
				},
			},
		},
	}, nil
}

// routeContainsDynamicModule returns true if DynamicModules exist for the provided route.
func routeContainsDynamicModule(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}
	return irRoute.EnvoyExtensions != nil && len(irRoute.EnvoyExtensions.DynamicModules) > 0
}

// listenerContainsDynamicModule returns true if DynamicModules exist at listener scope.
func listenerContainsDynamicModule(irListener *ir.HTTPListener) bool {
	return irListener != nil && irListener.EnvoyExtensions != nil && len(irListener.EnvoyExtensions.DynamicModules) > 0
}

// patchResources creates clusters for remote dynamic module sources.
func (*dynamicModule) patchResources(tCtx *types.ResourceVersionTable, irListener *ir.HTTPListener, routes []*ir.HTTPRoute) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	addClusters := func(dms []ir.DynamicModule) {
		for _, dm := range dms {
			if dm.Remote == nil {
				continue
			}
			if err := addClusterFromURL(dm.Remote.URL, nil, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	if listenerContainsDynamicModule(irListener) {
		addClusters(irListener.EnvoyExtensions.DynamicModules)
	}
	for _, route := range routes {
		if !routeContainsDynamicModule(route) {
			continue
		}
		addClusters(route.EnvoyExtensions.DynamicModules)
	}

	return errs
}

// patchRoute enables the corresponding dynamic module filter for the provided route.
//
// A nil EnvoyExtensions means no route-scoped policy owns this route: it keeps inheriting the
// listener-scoped DynamicModules delivered at VirtualHost scope by patchVirtualHost.
//
// A non-nil EnvoyExtensions means a more specific (xRoute or route rule) policy owns this route
// and fully replaces — never merges with — the listener-scoped policy. The extension count is
// intentionally not checked: an empty result (e.g. fail-open invalid Wasm) still represents a
// more specific policy that owns this route and must suppress the lower-scope DynamicModules.
func (*dynamicModule) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, irListener *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.EnvoyExtensions == nil {
		return nil
	}

	own := make(map[string]struct{}, len(irRoute.EnvoyExtensions.DynamicModules))
	for i := range irRoute.EnvoyExtensions.DynamicModules {
		own[dynamicModuleFilterName(&irRoute.EnvoyExtensions.DynamicModules[i])] = struct{}{}
	}

	if listenerContainsDynamicModule(irListener) {
		for i := range irListener.EnvoyExtensions.DynamicModules {
			filterName := dynamicModuleFilterName(&irListener.EnvoyExtensions.DynamicModules[i])
			// A single EnvoyExtensionPolicy may target both this listener and this route via
			// separate targetRefs, in which case the same filter name appears at both scopes and
			// the route re-enables it below instead of disabling it.
			if _, ok := own[filterName]; ok {
				continue
			}
			if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{Disabled: true}); err != nil {
				return err
			}
		}
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

// patchVirtualHost enables the listener-scoped DynamicModule filters at VirtualHost scope so a
// listener's policy does not bleed into virtual hosts belonging to a different listener that
// shares the same RouteConfiguration. Delivery via VirtualHost TypedPerFilterConfig goes through
// RDS, so policy changes do not trigger listener drains.
func (*dynamicModule) patchVirtualHost(vh *routev3.VirtualHost, httpListener *ir.HTTPListener) error {
	if !listenerContainsDynamicModule(httpListener) {
		return nil
	}

	for i := range httpListener.EnvoyExtensions.DynamicModules {
		dm := &httpListener.EnvoyExtensions.DynamicModules[i]
		if err := enableFilterOnVirtualHost(vh, dynamicModuleFilterName(dm), &routev3.FilterConfig{
			Config: &anypb.Any{},
		}); err != nil {
			return err
		}
	}
	return nil
}
