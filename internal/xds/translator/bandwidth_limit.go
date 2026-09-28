// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	bwlimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/bandwidth_limit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var (
	// Two named filter instances allow independent limits per direction.
	bandwidthLimitRequestFilterName  = perRouteFilterName(egv1a1.EnvoyFilterBandwidthLimit, "request")
	bandwidthLimitResponseFilterName = perRouteFilterName(egv1a1.EnvoyFilterBandwidthLimit, "response")
)

const bandwidthLimitStatPrefix = "http_bandwidth_limiter"

func init() {
	registerHTTPFilter(&bandwidthLimit{})
}

type bandwidthLimit struct{}

var _ httpFilter = &bandwidthLimit{}

// patchHCM builds and appends the bandwidthLimit Filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
// The filter is disabled by default. It is enabled on the route level.
//
// Two separate named filter instances (one for request, one for response) are registered
// because the Envoy bandwidth_limit filter currently does not support independent per-direction
// limits within a single filter instance. By registering two instances with distinct names, each route can
// apply different limits per direction via TypedPerFilterConfig.
//
// Reference: https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/bandwidth_limit/v3/bandwidth_limit.proto#extensions-filters-http-bandwidth-limit-v3-bandwidthlimit
//
// Once Envoy supports per-direction configuration in a single filter instance, these can be
// consolidated into one.
func (*bandwidthLimit) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	needsRequest := listenerContainsBandwidthLimitRequest(irListener)
	needsResponse := listenerContainsBandwidthLimitResponse(irListener)
	if !needsRequest && !needsResponse {
		return nil
	}

	bandwidthLimitProto, err := anypb.New(&bwlimitv3.BandwidthLimit{
		StatPrefix: bandwidthLimitStatPrefix,
		EnableMode: bwlimitv3.BandwidthLimit_DISABLED,
	})
	if err != nil {
		return err
	}

	if needsRequest && !hcmContainsFilter(mgr, bandwidthLimitRequestFilterName) {
		mgr.HttpFilters = append(mgr.HttpFilters, &hcmv3.HttpFilter{
			Name:       bandwidthLimitRequestFilterName,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: bandwidthLimitProto},
			Disabled:   true,
		})
	}
	if needsResponse && !hcmContainsFilter(mgr, bandwidthLimitResponseFilterName) {
		mgr.HttpFilters = append(mgr.HttpFilters, &hcmv3.HttpFilter{
			Name:       bandwidthLimitResponseFilterName,
			Disabled:   true,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: bandwidthLimitProto},
		})
	}
	return nil
}

func listenerContainsBandwidthLimitRequest(irListener *ir.HTTPListener) bool {
	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.BandwidthLimit != nil && route.Traffic.BandwidthLimit.Request != nil {
			return true
		}
	}
	return false
}

func listenerContainsBandwidthLimitResponse(irListener *ir.HTTPListener) bool {
	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.BandwidthLimit != nil && route.Traffic.BandwidthLimit.Response != nil {
			return true
		}
	}
	return false
}

func (*bandwidthLimit) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

func (*bandwidthLimit) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Traffic == nil || irRoute.Traffic.BandwidthLimit == nil {
		return nil
	}

	bl := irRoute.Traffic.BandwidthLimit

	// Overwrite the HCM level filter config with the per route filter config.
	if route.GetTypedPerFilterConfig() == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	if bl.Request != nil {
		if _, ok := route.TypedPerFilterConfig[bandwidthLimitRequestFilterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				bandwidthLimitRequestFilterName, route)
		}
		proto := buildBandwidthLimitRequestProto(bl.Request)
		bandwidthReqAny, err := anypb.New(proto)
		if err != nil {
			return err
		}
		route.TypedPerFilterConfig[bandwidthLimitRequestFilterName] = bandwidthReqAny
	}

	if bl.Response != nil {
		if _, ok := route.TypedPerFilterConfig[bandwidthLimitResponseFilterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				bandwidthLimitResponseFilterName, route)
		}
		proto := buildBandwidthLimitResponseProto(bl.Response)
		bandwidthResAny, err := anypb.New(proto)
		if err != nil {
			return err
		}
		route.TypedPerFilterConfig[bandwidthLimitResponseFilterName] = bandwidthResAny
	}
	return nil
}

func buildBandwidthLimitRequestProto(cfg *ir.BandwidthLimitConfig) *bwlimitv3.BandwidthLimit {
	return &bwlimitv3.BandwidthLimit{
		StatPrefix: bandwidthLimitStatPrefix,
		EnableMode: bwlimitv3.BandwidthLimit_REQUEST,
		LimitKbps:  &wrapperspb.UInt64Value{Value: cfg.LimitKibps},
	}
}

func buildBandwidthLimitResponseProto(cfg *ir.BandwidthLimitConfig) *bwlimitv3.BandwidthLimit {
	proto := &bwlimitv3.BandwidthLimit{
		StatPrefix: bandwidthLimitStatPrefix,
		EnableMode: bwlimitv3.BandwidthLimit_RESPONSE,
		LimitKbps:  &wrapperspb.UInt64Value{Value: cfg.LimitKibps},
	}
	if cfg.ResponseTrailers != nil {
		proto.EnableResponseTrailers = true
		if cfg.ResponseTrailers.Prefix != nil {
			proto.ResponseTrailerPrefix = *cfg.ResponseTrailers.Prefix
		}
	}
	return proto
}
