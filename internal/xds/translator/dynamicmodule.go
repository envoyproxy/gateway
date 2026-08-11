// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

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

// patchHCM adds disabled envoy.filters.http.filter_chain placeholder filters to the HTTP
// Connection Manager: one for per-listener (per-connection) DynamicModule and one for per-route
// DynamicModule.
//
// Both placeholders are added together as soon as either scope has a DynamicModule policy
// anywhere on this listener, even if the other scope currently has none. This keeps the HCM's
// filter set stable across that kind of policy churn too: e.g. adding a per-listener
// DynamicModule policy later to a listener that already has per-route DynamicModule only changes
// route/virtual host TypedPerFilterConfig (an RDS update), never the listener's filter list
// (which would require an LDS update and a connection drain).
//
// DynamicModule has no native per-route override at all, while EG's EnvoyExtensionPolicy API
// allows an ordered list of DynamicModule filters per listener/route. The filter_chain filter
// wraps an ordered, named sub-chain of DynamicModule filters that is supplied separately (per
// virtual host for listener-scoped DynamicModule, per route for route-scoped DynamicModule).
func (*dynamicModule) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	hasListenerDynamicModule := listenerContainsDynamicModule(irListener)
	hasRouteDynamicModule := slices.ContainsFunc(irListener.Routes, routeContainsDynamicModule)
	if !hasListenerDynamicModule && !hasRouteDynamicModule {
		return nil
	}

	for _, filterName := range []string{eepListenerFCFilterName(), eepFCFilterName()} {
		if hcmContainsFilter(mgr, filterName) {
			continue
		}
		filter, err := buildHCMFilterChainFilter(filterName)
		if err != nil {
			return err
		}
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return nil
}

// dynamicModuleSubFilterName returns the stable top-level filter name for the per-route
// DynamicModule slot index. The index is the execution slot within the ordered
// EnvoyExtensionPolicy DynamicModule list, so route 0th modules always bind to the same
// listener-level filter.
func dynamicModuleSubFilterName(idx int) string {
	return perRouteFilterName(egv1a1.EnvoyFilterDynamicModules, strconv.Itoa(idx))
}

// dynamicModuleListenerSubFilterName returns the stable HCM-level filter name for a
// listener-level DynamicModule slot. Using the envoy.filters.http.dynamic_modules prefix
// (instead of the raw policy name) ensures sortHTTPFilters assigns it the correct order
// relative to route-level slots.
func dynamicModuleListenerSubFilterName(idx int) string {
	return fmt.Sprintf("%s/listener/%d", egv1a1.EnvoyFilterDynamicModules, idx)
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

	// A non-nil EnvoyExtensions means a more specific route policy owns this route and fully
	// replaces the listener-scoped policy. The extension count is intentionally not checked
	// here: an empty result (e.g. fail-open invalid Wasm) still represents a more specific
	// policy that owns this route and must suppress the lower-scope DynamicModule.
	if err := disableFilterOnRouteOnce(route, eepListenerFCFilterName()); err != nil {
		return err
	}

	var newFilters []*corev3.TypedExtensionConfig
	for idx := range irRoute.EnvoyExtensions.DynamicModules {
		cfg, err := dynamicModuleConfig(&irRoute.EnvoyExtensions.DynamicModules[idx])
		if err != nil {
			return err
		}
		cfgAny, err := anypb.New(cfg)
		if err != nil {
			return err
		}
		newFilters = append(newFilters, &corev3.TypedExtensionConfig{
			Name:        dynamicModuleSubFilterName(idx),
			TypedConfig: cfgAny,
		})
	}

	if len(newFilters) == 0 {
		return nil
	}

	merged, err := mergeFilterChainConfigPerRoute(route.GetTypedPerFilterConfig()[eepFCFilterName()], newFilters)
	if err != nil {
		return err
	}
	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[eepFCFilterName()] = merged
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

	filterName := eepListenerFCFilterName()
	existing := vh.GetTypedPerFilterConfig()[filterName]
	alreadyDelivered, err := filterChainAlreadyHasType(existing, egv1a1.EnvoyFilterDynamicModules)
	if err != nil {
		return err
	}
	if alreadyDelivered {
		return nil
	}

	var newFilters []*corev3.TypedExtensionConfig
	for idx := range httpListener.EnvoyExtensions.DynamicModules {
		cfg, err := dynamicModuleConfig(&httpListener.EnvoyExtensions.DynamicModules[idx])
		if err != nil {
			return err
		}
		cfgAny, err := anypb.New(cfg)
		if err != nil {
			return err
		}
		newFilters = append(newFilters, &corev3.TypedExtensionConfig{
			Name:        dynamicModuleListenerSubFilterName(idx),
			TypedConfig: cfgAny,
		})
	}

	merged, err := mergeFilterChainConfigPerRoute(existing, newFilters)
	if err != nil {
		return err
	}
	if vh.TypedPerFilterConfig == nil {
		vh.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	vh.TypedPerFilterConfig[filterName] = merged
	return nil
}
