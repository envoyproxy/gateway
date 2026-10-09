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
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

// The alternate protocols cache filter records alt-svc response headers into the routed
// cluster's cache, which is what lets HTTP/3 mode Auto discover backends that speak HTTP/3.
// Envoy registers it as a downstream filter only, so it lives in the HCM chain, disabled by
// default and enabled on the routes that need it. It is patched outside the httpFilter
// registry because deciding whether a route uses HTTP/3 needs the backend cluster index.

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

// routeHasHTTP3Cluster reports whether the route proxies to a cluster that the cluster builder
// gives QUIC and whose HTTP/3 settings accept returns true for. The route's own Traffic.HTTP3
// governs its route-scoped cluster, the one built from Destination.Settings; a merged cluster
// behind a BackendClusterRef carries its own, since it is shared with other routes and may have
// been accepted or dropped independently. Redirect and direct-response routes never dial.
func routeHasHTTP3Cluster(route *ir.HTTPRoute, backendIndex backendClusterIndex, accept func(*ir.BackendHTTP3Settings) bool) bool {
	if route == nil || route.Destination == nil {
		return false
	}
	if settings := routeClusterHTTP3Settings(route); settings != nil && accept(settings) {
		return true
	}
	for _, bc := range resolveBackendClusters(route.Destination, backendIndex) {
		if settings := backendClusterHTTP3Settings(bc); settings != nil && accept(settings) {
			return true
		}
	}
	return false
}

func routeUsesUpstreamHTTP3(route *ir.HTTPRoute, backendIndex backendClusterIndex) bool {
	return routeHasHTTP3Cluster(route, backendIndex, func(*ir.BackendHTTP3Settings) bool { return true })
}

// routeUsesAutoHTTP3 reports whether any cluster the route reaches uses mode Auto, so discovery
// is enabled regardless of backend order or the route-scoped cluster's mode.
func routeUsesAutoHTTP3(route *ir.HTTPRoute, backendIndex backendClusterIndex) bool {
	return routeHasHTTP3Cluster(route, backendIndex, func(settings *ir.BackendHTTP3Settings) bool {
		return settings.Mode == string(egv1a1.BackendHTTP3ModeAuto)
	})
}

// routeClusterHTTP3Settings returns the HTTP/3 settings the route-scoped cluster is built
// with, or nil when there is no such cluster or the cluster builder would not give it QUIC.
// Sharing ir.CanUseHTTP3 with the cluster builder keeps the alt-svc strip and the alternate
// protocols cache from being enabled on a route whose cluster never speaks HTTP/3, even when
// the IR still carries HTTP3 for backends that cannot use it.
func routeClusterHTTP3Settings(route *ir.HTTPRoute) *ir.BackendHTTP3Settings {
	if len(route.Destination.Settings) == 0 || route.Traffic == nil || route.Traffic.HTTP3 == nil {
		return nil
	}
	if !ir.CanUseHTTP3(route.Destination.Settings, ptr.Deref(route.UseClientProtocol, false), route.Traffic.ProxyProtocol != nil) {
		return nil
	}
	return route.Traffic.HTTP3
}

// backendClusterHTTP3Settings is routeClusterHTTP3Settings for a merged cluster.
func backendClusterHTTP3Settings(bc *ir.BackendCluster) *ir.BackendHTTP3Settings {
	if bc == nil || bc.Traffic == nil || bc.Traffic.HTTP3 == nil {
		return nil
	}
	if !ir.CanUseHTTP3([]*ir.DestinationSetting{bc.Setting}, ptr.Deref(bc.UseClientProtocol, false), bc.Traffic.ProxyProtocol != nil) {
		return nil
	}
	return bc.Traffic.HTTP3
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
