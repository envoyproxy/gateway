// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	wasmfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/wasm/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	wasmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/wasm/v3"
	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	wasmFilter  = "envoy.filters.http.wasm"
	vmRuntimeV8 = "envoy.wasm.runtime.v8"
)

func init() {
	registerHTTPFilter(&wasm{})
}

type wasm struct {
}

var _ httpFilter = &wasm{}

// patchHCM builds and appends the wasm Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates a wasm filter for each route that contains an wasm config.
// The filter is disabled by default. It is enabled on the route level.
func (*wasm) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsWasm(route) {
			continue
		}
		for _, ep := range route.Wasms {
			if hcmContainsFilter(mgr, wasmFilterName(ep)) {
				continue
			}
			filter, err := buildHCMWasmFilter(ep)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}

	return errs
}

// buildHCMWasmFilter returns a wasm HTTP filter from the provided IR HTTPRoute.
func buildHCMWasmFilter(wasm ir.Wasm) (*hcmv3.HttpFilter, error) {
	var (
		wasmProto *wasmfilterv3.Wasm
		wasmAny   *anypb.Any
		err       error
	)

	if wasmProto, err = wasmConfig(wasm); err != nil {
		return nil, err
	}
	if err = wasmProto.ValidateAll(); err != nil {
		return nil, err
	}
	if wasmAny, err = anypb.New(wasmProto); err != nil {
		return nil, err
	}

	// All wasm filters for all Routes are aggregated on HCM and disabled by default
	// Per-route config is used to enable the relevant filters on appropriate routes
	return &hcmv3.HttpFilter{
		Name:     wasmFilterName(wasm),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: wasmAny,
		},
	}, nil
}

func wasmFilterName(wasm ir.Wasm) string {
	return perRouteFilterName(wasmFilter, wasm.Name)
}

func wasmConfig(wasm ir.Wasm) (*wasmfilterv3.Wasm, error) {
	var (
		uc           *urlCluster
		pluginConfig = ""
		configAny    *anypb.Any
		filterConfig *wasmfilterv3.Wasm
		err          error
	)

	// We only support HTTP Wasm code source for now
	if uc, err = url2Cluster(wasm.HTTPWasmCode.URL); err != nil {
		return nil, err
	}

	if wasm.Config != nil {
		pluginConfig = string(wasm.Config.Raw)
	}

	if configAny, err = anypb.New(wrapperspb.String(pluginConfig)); err != nil {
		return nil, err
	}

	filterConfig = &wasmfilterv3.Wasm{
		Config: &wasmv3.PluginConfig{
			Name: wasm.WasmName,
			Vm: &wasmv3.PluginConfig_VmConfig{
				VmConfig: &wasmv3.VmConfig{
					VmId:    wasm.Name, // Do not share VMs across different filters
					Runtime: vmRuntimeV8,
					Code: &corev3.AsyncDataSource{
						Specifier: &corev3.AsyncDataSource_Remote{
							Remote: &corev3.RemoteDataSource{
								HttpUri: &corev3.HttpUri{
									Uri: wasm.HTTPWasmCode.URL,
									HttpUpstreamType: &corev3.HttpUri_Cluster{
										Cluster: uc.name,
									},
									Timeout: &duration.Duration{
										Seconds: defaultExtServiceRequestTimeout,
									},
								},
								Sha256: wasm.HTTPWasmCode.SHA256,
							},
						},
					},
				},
			},
			Configuration: configAny,
			FailOpen:      wasm.FailOpen,
		},
	}

	if wasm.RootID != nil {
		filterConfig.Config.RootId = *wasm.RootID
	}

	return filterConfig, nil
}

// routeContainsWasm returns true if Wasms exists for the provided route.
func routeContainsWasm(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	return len(irRoute.Wasms) > 0
}

// patchResources patches the cluster resources for the http wasm code source.
func (*wasm) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var err, errs error
	for _, route := range routes {
		if !routeContainsWasm(route) {
			continue
		}

		for _, w := range route.Wasms {
			if err = addClusterFromURL(w.HTTPWasmCode.URL, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

// patchRoute patches the provided route with the wasm config if applicable.
// Note: this method enables the corresponding wasm filter for the provided route.
func (*wasm) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	for _, ep := range irRoute.Wasms {
		filterName := wasmFilterName(ep)
		if err := enableFilterOnRoute(route, filterName); err != nil {
			return err
		}
	}
	return nil
}
