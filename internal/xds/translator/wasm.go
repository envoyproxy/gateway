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

// patchHCM adds disabled envoy.filters.http.filter_chain placeholder filters to the HTTP
// Connection Manager: one for per-listener (per-connection) Wasm and one for per-route Wasm.
//
// Both placeholders are added together as soon as either scope has a Wasm policy anywhere on
// this listener, even if the other scope currently has none. This keeps the HCM's filter set
// stable across that kind of policy churn too: e.g. adding a per-listener Wasm policy later to
// a listener that already has per-route Wasm only changes route/virtual host
// TypedPerFilterConfig (an RDS update), never the listener's filter list (which would require
// an LDS update and a connection drain).
//
// Wasm has no native per-route override at all, while EG's EnvoyExtensionPolicy API allows an
// ordered list of Wasm filters per listener/route. The filter_chain filter wraps an ordered,
// named sub-chain of Wasm filters that is supplied separately (per virtual host for
// listener-scoped Wasm, per route for route-scoped Wasm).
func (*wasm) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	hasListenerWasm := listenerContainsWasm(irListener)
	hasRouteWasm := slices.ContainsFunc(irListener.Routes, routeContainsWasm)
	if !hasListenerWasm && !hasRouteWasm {
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

// wasmSubFilterName returns the stable top-level filter name for the per-route Wasm slot index.
// The index is the execution slot within the ordered EnvoyExtensionPolicy Wasm list, so route
// 0th modules always bind to the same listener-level filter.
func wasmSubFilterName(idx int) string {
	return perRouteFilterName(egv1a1.EnvoyFilterWasm, strconv.Itoa(idx))
}

// wasmListenerSubFilterName returns the stable HCM-level filter name for a listener-level Wasm
// slot. Using the envoy.filters.http.wasm prefix (instead of the raw policy name) ensures
// sortHTTPFilters assigns it the correct order relative to route-level slots.
func wasmListenerSubFilterName(idx int) string {
	return fmt.Sprintf("%s/listener/%d", egv1a1.EnvoyFilterWasm, idx)
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
func (*wasm) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
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
	// policy that owns this route and must suppress the lower-scope Wasm.
	if err := disableFilterOnRouteOnce(route, eepListenerFCFilterName()); err != nil {
		return err
	}

	var newFilters []*corev3.TypedExtensionConfig
	for idx := range irRoute.EnvoyExtensions.Wasms {
		cfg, err := wasmConfig(&irRoute.EnvoyExtensions.Wasms[idx])
		if err != nil {
			return err
		}
		cfgAny, err := anypb.New(cfg)
		if err != nil {
			return err
		}
		newFilters = append(newFilters, &corev3.TypedExtensionConfig{
			Name:        wasmSubFilterName(idx),
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

// patchVirtualHost enables the listener-scoped Wasm filters at VirtualHost scope so a listener's
// policy does not bleed into virtual hosts belonging to a different listener that shares the same
// RouteConfiguration. Delivery via VirtualHost TypedPerFilterConfig goes through RDS, so policy
// changes do not trigger listener drains.
func (*wasm) patchVirtualHost(vh *routev3.VirtualHost, httpListener *ir.HTTPListener) error {
	if !listenerContainsWasm(httpListener) {
		return nil
	}

	filterName := eepListenerFCFilterName()
	existing := vh.GetTypedPerFilterConfig()[filterName]
	alreadyDelivered, err := filterChainAlreadyHasType(existing, egv1a1.EnvoyFilterWasm)
	if err != nil {
		return err
	}
	if alreadyDelivered {
		return nil
	}

	var newFilters []*corev3.TypedExtensionConfig
	for idx := range httpListener.EnvoyExtensions.Wasms {
		cfg, err := wasmConfig(&httpListener.EnvoyExtensions.Wasms[idx])
		if err != nil {
			return err
		}
		cfgAny, err := anypb.New(cfg)
		if err != nil {
			return err
		}
		newFilters = append(newFilters, &corev3.TypedExtensionConfig{
			Name:        wasmListenerSubFilterName(idx),
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
