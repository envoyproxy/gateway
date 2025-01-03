// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	compressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/protocov"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&compressor{})
}

type compressor struct{}

var _ httpFilter = &compressor{}

// patchHCM builds and appends the compressor Filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
func (*compressor) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}
	if hcmContainsFilter(mgr, egv1a1.EnvoyFilterCompressor.String()) {
		return nil
	}

	var (
		irCompression *ir.Compression
		filter        *hcmv3.HttpFilter
		err           error
	)

	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.Compression != nil {
			irCompression = route.Traffic.Compression
		}
	}
	if irCompression == nil {
		return nil
	}

	// The HCM-level filter config doesn't matter since it is overridden at the route level.
	if filter, err = buildHCMCompressorFilter(); err != nil {
		return err
	}
	mgr.HttpFilters = append(mgr.HttpFilters, filter)
	return err
}

// buildHCMCompressorFilter returns a Compressor HTTP filter from the provided IR HTTPRoute.
func buildHCMCompressorFilter() (*hcmv3.HttpFilter, error) {
	var (
		compressorProto *compressorv3.Compressor
		gzipAny         *anypb.Any
		compressorAny   *anypb.Any
		err             error
	)

	if gzipAny, err = protocov.ToAnyWithValidation(&gzipv3.Gzip{}); err != nil {
		return nil, err
	}

	compressorProto = &compressorv3.Compressor{
		CompressorLibrary: &corev3.TypedExtensionConfig{
			Name:        "envoy.compressor.gzip",
			TypedConfig: gzipAny,
		},
	}

	if compressorAny, err = protocov.ToAnyWithValidation(compressorProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterCompressor.String(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: compressorAny,
		},
		Disabled: true,
	}, nil
}

func (*compressor) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the compressor config if applicable.
// Note: this method overwrites the HCM level filter config with the per route filter config.
func (*compressor) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Traffic == nil || irRoute.Traffic.Compression == nil {
		return nil
	}

	var (
		perFilterCfg  map[string]*anypb.Any
		compressorAny *anypb.Any
		err           error
	)

	perFilterCfg = route.GetTypedPerFilterConfig()
	if _, ok := perFilterCfg[egv1a1.EnvoyFilterCompressor.String()]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf("route already contains filter config: %s, %+v",
			egv1a1.EnvoyFilterCompressor.String(), route)
	}

	// Overwrite the HCM level filter config with the per route filter config.
	compressorProto := compressorPerRouteConfig(irRoute.Traffic.Compression)

	if compressorAny, err = protocov.ToAnyWithValidation(compressorProto); err != nil {
		return err
	}

	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[egv1a1.EnvoyFilterCompressor.String()] = compressorAny

	return nil
}

func compressorPerRouteConfig(_ *ir.Compression) *compressorv3.CompressorPerRoute {
	// Enable compression on this route if compression is configured.
	return &compressorv3.CompressorPerRoute{
		Override: &compressorv3.CompressorPerRoute_Overrides{
			Overrides: &compressorv3.CompressorOverrides{
				ResponseDirectionConfig: &compressorv3.ResponseDirectionOverrides{},
			},
		},
	}
}
