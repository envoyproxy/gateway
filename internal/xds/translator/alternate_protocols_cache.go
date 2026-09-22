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
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&alternateProtocolsCache{})
}

// alternateProtocolsCache records alt-svc response headers into the routed cluster's
// alternate protocols cache, which is what lets HTTP/3 mode Auto discover backends
// that speak HTTP/3. Envoy registers it as a downstream filter only, so it lives in
// the HCM chain, disabled by default and enabled on the routes that need it.
type alternateProtocolsCache struct{}

var _ httpFilter = &alternateProtocolsCache{}

func (*alternateProtocolsCache) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
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
		if routeUsesAutoHTTP3(route) {
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

func (*alternateProtocolsCache) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

func (*alternateProtocolsCache) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if !routeUsesAutoHTTP3(irRoute) {
		return nil
	}
	return enableFilterOnRoute(route, egv1a1.EnvoyFilterAlternateProtocolsCache.String(), &routev3.FilterConfig{
		Config: &anypb.Any{},
	})
}

// A destination without TLS never gets a cache, so skip its route for consistency. A merged
// destination carries BackendClusterRefs rather than Settings and cannot be checked here,
// so it is enabled: the filter is a no-op on a cluster that has no cache anyway.
func routeUsesAutoHTTP3(route *ir.HTTPRoute) bool {
	if route == nil || route.Traffic == nil || route.Traffic.HTTP3 == nil ||
		route.Traffic.HTTP3.Mode != string(egv1a1.BackendHTTP3ModeAuto) || route.Destination == nil {
		return false
	}
	return len(route.Destination.Settings) == 0 || route.Destination.AllSettingsHaveTLS()
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
