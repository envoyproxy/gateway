// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"slices"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
	xdsfilters "github.com/envoyproxy/gateway/internal/xds/filters"
)

func isGRPCRoute(route *ir.HTTPRoute) bool {
	return route.Metadata != nil && route.Metadata.Kind == "GRPCRoute"
}

func grpcWebEnabled(listener *ir.HTTPListener) bool {
	if listener.GRPC != nil && listener.GRPC.EnableGRPCWeb != nil {
		return *listener.GRPC.EnableGRPCWeb
	}
	return slices.ContainsFunc(listener.Routes, isGRPCRoute)
}

func patchRouteWithGRPCWeb(route *routev3.Route, irRoute *ir.HTTPRoute, listener *ir.HTTPListener) error {
	if isGRPCRoute(irRoute) || (listener.GRPC != nil && listener.GRPC.EnableGRPCWeb != nil) || !grpcWebEnabled(listener) {
		return nil
	}

	// Preserve all other per-route settings on shared listeners.
	config, err := anypb.New(&routev3.FilterConfig{Disabled: true})
	if err != nil {
		return err
	}
	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = map[string]*anypb.Any{}
	}
	route.TypedPerFilterConfig[xdsfilters.GRPCWeb.Name] = config
	return nil
}
