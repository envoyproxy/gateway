// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	alternateprotocolscachev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/alternate_protocols_cache/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

// The alternate protocols cache filter records alt-svc response headers into the routed
// cluster's cache, which is what lets HTTP/3 mode Auto discover backends that speak HTTP/3.
// Envoy registers it as a downstream filter only, so it lives in the HCM chain, disabled by
// default and enabled on the routes that need it.
//
// @TODO (Robin) this can be simplified once Envoy supports registering it as an upstream filter: https://github.com/envoyproxy/envoy/issues/48365
//
// patchHCMWithAlternateProtocolsCache adds the filter to the HCM when any route on the
// listener uses HTTP/3 mode Auto.
func (t *Translator) patchHCMWithAlternateProtocolsCache(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	filterName := egv1a1.EnvoyFilterAlternateProtocolsCache.String()
	if hcmContainsFilter(mgr, filterName) {
		return nil
	}
	for _, route := range irListener.Routes {
		if routeUsesAutoHTTP3(route, t.backendIndex) {
			filter, err := buildAlternateProtocolsCacheFilter(filterName)
			if err != nil {
				return err
			}
			mgr.HttpFilters = append(mgr.HttpFilters, filter)
			return nil
		}
	}
	return nil
}

// patchRouteWithAlternateProtocolsCache enables the filter on a route that uses HTTP/3 mode Auto.
func patchRouteWithAlternateProtocolsCache(route *routev3.Route, irRoute *ir.HTTPRoute, backendIndex backendClusterIndex) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if !routeUsesAutoHTTP3(irRoute, backendIndex) {
		return nil
	}
	return enableFilterOnRoute(route, egv1a1.EnvoyFilterAlternateProtocolsCache.String(), &routev3.FilterConfig{
		Config: &anypb.Any{},
	})
}

// routeUsesUpstreamHTTP3 reports whether the route sends traffic to a cluster that gets QUIC:
// its own cluster, built from Destination.Settings, or a merged cluster it shares with other
// routes, which carries its own settings. Redirect and direct-response routes reach no cluster.
func routeUsesUpstreamHTTP3(route *ir.HTTPRoute, backendIndex backendClusterIndex) bool {
	if route == nil || route.Destination == nil {
		return false
	}
	if len(route.Destination.Settings) > 0 && route.Traffic != nil && route.Traffic.HTTP3 != nil {
		return true
	}
	for _, bc := range resolveBackendClusters(route.Destination, backendIndex) {
		if bc != nil && bc.Traffic != nil && bc.Traffic.HTTP3 != nil {
			return true
		}
	}
	return false
}

// routeUsesAutoHTTP3 reports whether any cluster the route reaches uses mode Auto.
func routeUsesAutoHTTP3(route *ir.HTTPRoute, backendIndex backendClusterIndex) bool {
	if route == nil || route.Destination == nil {
		return false
	}
	auto := string(egv1a1.BackendHTTP3ModeAuto)
	if len(route.Destination.Settings) > 0 && route.Traffic != nil && route.Traffic.HTTP3 != nil && route.Traffic.HTTP3.Mode == auto {
		return true
	}
	for _, bc := range resolveBackendClusters(route.Destination, backendIndex) {
		if bc != nil && bc.Traffic != nil && bc.Traffic.HTTP3 != nil && bc.Traffic.HTTP3.Mode == auto {
			return true
		}
	}
	return false
}

func listenerHasUpstreamHTTP3(listener *ir.HTTPListener, backendIndex backendClusterIndex) bool {
	if listener == nil {
		return false
	}
	for _, route := range listener.Routes {
		if routeUsesUpstreamHTTP3(route, backendIndex) {
			return true
		}
	}
	return false
}

// The filter's own alternate_protocols_cache_options field is deprecated and ignored:
// Envoy reads the cache options off the cluster the request was routed to.
func buildAlternateProtocolsCacheFilter(name string) (*hcmv3.HttpFilter, error) {
	cacheAny, err := proto.ToAnyWithValidation(&alternateprotocolscachev3.FilterConfig{})
	if err != nil {
		return nil, err
	}
	return &hcmv3.HttpFilter{
		Name: name,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: cacheAny,
		},
		Disabled: true,
	}, nil
}
