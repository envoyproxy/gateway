// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	wasmfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/wasm/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	wasmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/wasm/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	vmRuntimeV8 = "envoy.wasm.runtime.v8"

	// Envoy's built-in retry policy for remote Wasm code is a single retry with
	// a ~1s backoff, after which the fetch is never re-attempted and the filter
	// stays failed until the next config update. Use a more generous jittered
	// exponential backoff so the fetch survives transient unavailability of the
	// Wasm HTTP service (e.g. slow module downloads or pod startup).
	wasmFetchNumRetries   = uint32(10)
	wasmFetchBaseInterval = 1 * time.Second
	wasmFetchMaxInterval  = 30 * time.Second
)

func init() {
	registerHTTPFilter(&wasm{})
}

type wasm struct{}

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

	addFilters := func(wasms []ir.Wasm) {
		for i := range wasms {
			ep := &wasms[i]
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

	// Listener-scoped Wasms are enabled at VirtualHost scope; route-scoped Wasms are enabled
	// per route. Both need their (disabled by default) filter present on the HCM.
	if listenerContainsWasm(irListener) {
		addFilters(irListener.EnvoyExtensions.Wasms)
	}
	for _, route := range irListener.Routes {
		if !routeContainsWasm(route) {
			continue
		}
		addFilters(route.EnvoyExtensions.Wasms)
	}

	return errs
}

// buildHCMWasmFilter returns a wasm HTTP filter from the provided IR HTTPRoute.
func buildHCMWasmFilter(wasm *ir.Wasm) (*hcmv3.HttpFilter, error) {
	var (
		wasmProto *wasmfilterv3.Wasm
		wasmAny   *anypb.Any
		err       error
	)

	if wasmProto, err = wasmConfig(wasm); err != nil {
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

func wasmFilterName(wasm *ir.Wasm) string {
	return perRouteFilterName(egv1a1.EnvoyFilterWasm, wasm.Name)
}

func wasmConfig(wasm *ir.Wasm) (*wasmfilterv3.Wasm, error) {
	var (
		pluginConfig = ""
		configAny    *anypb.Any
		filterConfig *wasmfilterv3.Wasm
		err          error
	)

	if wasm.Config != nil {
		pluginConfig = string(wasm.Config.Raw)
	}

	if configAny, err = anypb.New(wrapperspb.String(pluginConfig)); err != nil {
		return nil, err
	}

	vmConfig := &wasmv3.VmConfig{
		VmId:    wasm.Name, // Do not share VMs across different filters
		Runtime: vmRuntimeV8,
		Code: &corev3.AsyncDataSource{
			Specifier: &corev3.AsyncDataSource_Remote{
				Remote: &corev3.RemoteDataSource{
					HttpUri: &corev3.HttpUri{
						Uri: wasm.Code.ServingURL,
						HttpUpstreamType: &corev3.HttpUri_Cluster{
							Cluster: wasmHTTPServiceClusterName,
						},
						Timeout: durationpb.New(defaultExtServiceRequestTimeout),
					},
					Sha256: wasm.Code.SHA256,
					RetryPolicy: &corev3.RetryPolicy{
						NumRetries: wrapperspb.UInt32(wasmFetchNumRetries),
						RetryBackOff: &corev3.BackoffStrategy{
							BaseInterval: durationpb.New(wasmFetchBaseInterval),
							MaxInterval:  durationpb.New(wasmFetchMaxInterval),
						},
					},
				},
			},
		},
	}

	if wasm.HostKeys != nil {
		vmConfig.EnvironmentVariables = &wasmv3.EnvironmentVariables{
			HostEnvKeys: wasm.HostKeys,
		}
	}

	filterConfig = &wasmfilterv3.Wasm{
		Config: &wasmv3.PluginConfig{
			Name: wasm.WasmName,
			Vm: &wasmv3.PluginConfig_VmConfig{
				VmConfig: vmConfig,
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

	return irRoute.EnvoyExtensions != nil && len(irRoute.EnvoyExtensions.Wasms) > 0
}

// listenerContainsWasm returns true if Wasms exist at listener scope.
func listenerContainsWasm(irListener *ir.HTTPListener) bool {
	return irListener != nil && irListener.EnvoyExtensions != nil && len(irListener.EnvoyExtensions.Wasms) > 0
}

// patchResources patches the cluster resources for the http wasm code source.
func (*wasm) patchResources(_ *types.ResourceVersionTable, _ *ir.HTTPListener, _ []*ir.HTTPRoute) error {
	// EG always serves the Wasm module through the built-in HTTP server, which
	// has been configured in the bootstrap configuration. So we don't need to
	// create a cluster for the Wasm module.
	return nil
}

// patchRoute patches the provided route with the wasm config if applicable.
//
// A nil EnvoyExtensions means no route-scoped policy owns this route: it keeps inheriting the
// listener-scoped Wasms delivered at VirtualHost scope by patchVirtualHost.
//
// A non-nil EnvoyExtensions means a more specific (xRoute or route rule) policy owns this route
// and fully replaces — never merges with — the listener-scoped policy. The extension count is
// intentionally not checked: an empty result (e.g. fail-open invalid Wasm) still represents a
// more specific policy that owns this route and must suppress the lower-scope Wasms.
func (*wasm) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, irListener *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.EnvoyExtensions == nil {
		return nil
	}

	own := make(map[string]struct{}, len(irRoute.EnvoyExtensions.Wasms))
	for i := range irRoute.EnvoyExtensions.Wasms {
		own[wasmFilterName(&irRoute.EnvoyExtensions.Wasms[i])] = struct{}{}
	}

	if listenerContainsWasm(irListener) {
		for i := range irListener.EnvoyExtensions.Wasms {
			filterName := wasmFilterName(&irListener.EnvoyExtensions.Wasms[i])
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

	for _, ep := range irRoute.EnvoyExtensions.Wasms {
		filterName := wasmFilterName(&ep)
		if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
			Config: &anypb.Any{},
		}); err != nil {
			return err
		}
	}
	return nil
}

// patchVirtualHost enables the listener-scoped Wasm filters at VirtualHost scope so a listener's
// policy does not bleed into virtual hosts belonging to a different listener that shares the same
// RouteConfiguration. Delivery via VirtualHost TypedPerFilterConfig goes through RDS, so policy
// changes do not trigger listener drains.
func (*wasm) patchVirtualHost(vh *routev3.VirtualHost, httpListener *ir.HTTPListener) error {
	if !listenerContainsWasm(httpListener) {
		return nil
	}

	for i := range httpListener.EnvoyExtensions.Wasms {
		ep := &httpListener.EnvoyExtensions.Wasms[i]
		if err := enableFilterOnVirtualHost(vh, wasmFilterName(ep), &routev3.FilterConfig{
			Config: &anypb.Any{},
		}); err != nil {
			return err
		}
	}
	return nil
}
