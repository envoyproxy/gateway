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
// Note: this method creates a filter for each route that contains a dynamic module config, plus
// one for each DynamicModule a Gateway/Listener-scoped policy attached directly to the listener.
// The filter is disabled by default and enabled on the route level.
func (*dynamicModule) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	addDynamicModuleFilters := func(dms []ir.DynamicModule) {
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

	if irListener.EnvoyExtensions != nil {
		addDynamicModuleFilters(irListener.EnvoyExtensions.DynamicModules)
	}
	for _, route := range irListener.Routes {
		if !routeContainsDynamicModule(route) {
			continue
		}
		addDynamicModuleFilters(route.EnvoyExtensions.DynamicModules)
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

// patchResources creates clusters for remote dynamic module sources.
func (*dynamicModule) patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsDynamicModule(route) {
			continue
		}

		for _, dm := range route.EnvoyExtensions.DynamicModules {
			if dm.Remote == nil {
				continue
			}

			if err := addClusterFromURL(dm.Remote.URL, nil, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

// patchResources creates clusters for remote dynamic module sources a Gateway/Listener-scoped
// policy attached directly to irListener, as opposed to a specific route. Called directly from
// the translator alongside patchResources, the same way rate limiting is handled outside the
// generic per-filter loop, since patchResources has no listener context.
func patchDynamicModuleListenerResources(tCtx *types.ResourceVersionTable, irListener *ir.HTTPListener) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}
	if irListener == nil || irListener.EnvoyExtensions == nil {
		return nil
	}

	var errs error
	for _, dm := range irListener.EnvoyExtensions.DynamicModules {
		if dm.Remote == nil {
			continue
		}
		if err := addClusterFromURL(dm.Remote.URL, nil, tCtx); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

// patchRoute enables the corresponding dynamic module filter for the provided route. Routes with
// no DynamicModule of their own inherit whatever a Gateway/Listener-scoped policy attached to irListener.
func (*dynamicModule) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, irListener *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	extensions := effectiveEnvoyExtensions(irRoute, irListener)
	if extensions == nil {
		return nil
	}

	for i := range extensions.DynamicModules {
		dm := &extensions.DynamicModules[i]
		filterName := dynamicModuleFilterName(dm)
		if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
			Config: &anypb.Any{},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (*dynamicModule) patchVirtualHost(_ *routev3.VirtualHost, _ *ir.HTTPListener) error {
	return nil
}
