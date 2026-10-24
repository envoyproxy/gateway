// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"slices"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// Extensions read this namespace from the selected route.
// Lua reads the same bindings under its filter name.
const extensionBackendsMetadataNamespace = "gateway.envoyproxy.io/extension-backends"

func patchExtensionBackendMetadata(route *routev3.Route, irRoute *ir.HTTPRoute) {
	if irRoute.EnvoyExtensions == nil || len(irRoute.EnvoyExtensions.Backends) == 0 {
		return
	}
	bindings := &structpb.Struct{Fields: make(map[string]*structpb.Value, len(irRoute.EnvoyExtensions.Backends))}
	for alias, destination := range irRoute.EnvoyExtensions.Backends {
		bindings.Fields[alias] = structpb.NewStringValue(destination.Name)
	}
	if route.Metadata == nil {
		route.Metadata = &corev3.Metadata{}
	}
	if route.Metadata.FilterMetadata == nil {
		route.Metadata.FilterMetadata = make(map[string]*structpb.Struct)
	}
	route.Metadata.FilterMetadata[extensionBackendsMetadataNamespace] = bindings
	// Lua's route:metadata() reads the namespace matching its filter name.
	for _, lua := range irRoute.EnvoyExtensions.Luas {
		route.Metadata.FilterMetadata[luaFilterName(lua)] = bindings
	}
}

func patchExtensionBackendResources(tCtx *types.ResourceVersionTable, listeners []*ir.HTTPListener) error {
	var errs error
	for _, listener := range listeners {
		for _, route := range listener.Routes {
			if route.EnvoyExtensions == nil {
				continue
			}
			aliases := make([]string, 0, len(route.EnvoyExtensions.Backends))
			for alias := range route.EnvoyExtensions.Backends {
				aliases = append(aliases, alias)
			}
			slices.Sort(aliases)
			// The same policy can serve many routes. Existing helpers reuse its resources.
			for _, alias := range aliases {
				destination := route.EnvoyExtensions.Backends[alias]
				if err := createExtServiceXDSCluster(&destination.RouteDestination, destination.Traffic, tCtx); err != nil {
					errs = errors.Join(errs, err)
				}
				if err := processClientCertificates(tCtx, destination.Settings); err != nil {
					errs = errors.Join(errs, err)
				}
			}
		}
	}
	return errs
}
