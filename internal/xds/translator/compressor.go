// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	compressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
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

	var (
		brotli bool
		gzip   bool
		filter *hcmv3.HttpFilter
		err    error
	)

	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.Compression != nil {
			for _, irComp := range route.Traffic.Compression {
				if irComp.Type == egv1a1.BrotliCompressorType {
					brotli = true
				}
				if irComp.Type == egv1a1.GzipCompressorType {
					gzip = true
				}
			}
		}
	}

	// Add the compressor filters for all the compression types required by the routes.
	// All the compressor filters are disabled at the HCM level.
	// The per route filter config will enable the compressor filters for the routes that require them.
	if brotli {
		brotliFilterName := compressorFilterName(egv1a1.BrotliCompressorType)
		if !hcmContainsFilter(mgr, brotliFilterName) {
			if filter, err = buildCompressorFilter(egv1a1.BrotliCompressorType); err != nil {
				return err
			}
			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}

	if gzip {
		gzipFilterName := compressorFilterName(egv1a1.GzipCompressorType)
		if !hcmContainsFilter(mgr, gzipFilterName) {
			if filter, err = buildCompressorFilter(egv1a1.GzipCompressorType); err != nil {
				return err
			}
			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}

	return err
}

func compressorFilterName(compressorType egv1a1.CompressorType) string {
	return fmt.Sprintf("%s.%s", egv1a1.EnvoyFilterCompressor.String(), strings.ToLower(string(compressorType)))
}

// buildCompressorFilter builds a compressor filter with the provided compressionType.
func buildCompressorFilter(compressionType egv1a1.CompressorType) (*hcmv3.HttpFilter, error) {
	var (
		compressorProto *compressorv3.Compressor
		extensionName   string
		extensionMsg    protobuf.Message
		extensionAny    *anypb.Any
		compressorAny   *anypb.Any
		err             error
	)

	switch compressionType {
	case egv1a1.BrotliCompressorType:
		extensionName = "envoy.compression.brotli.compressor"
		extensionMsg = &brotliv3.Brotli{}
	case egv1a1.GzipCompressorType:
		extensionName = "envoy.compression.gzip.compressor"
		extensionMsg = &gzipv3.Gzip{}
	}

	if extensionAny, err = proto.ToAnyWithValidation(extensionMsg); err != nil {
		return nil, err
	}

	compressorProto = &compressorv3.Compressor{
		CompressorLibrary: &corev3.TypedExtensionConfig{
			Name:        extensionName,
			TypedConfig: extensionAny,
		},
	}

	if compressorAny, err = proto.ToAnyWithValidation(compressorProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: compressorFilterName(compressionType),
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
func (*compressor) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Traffic == nil || len(irRoute.Traffic.Compression) == 0 {
		return nil
	}

	var (
		brotli        bool
		gzip          bool
		perFilterCfg  map[string]*anypb.Any
		compressorAny *anypb.Any
		err           error
	)

	for _, irComp := range irRoute.Traffic.Compression {
		if irComp.Type == egv1a1.BrotliCompressorType {
			brotli = true
		}
		if irComp.Type == egv1a1.GzipCompressorType {
			gzip = true
		}
	}

	if !brotli && !gzip {
		return nil
	}

	// Overwrite the HCM level filter config with the per route filter config.
	perFilterCfg = route.GetTypedPerFilterConfig()
	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	compressorProto := compressorPerRouteConfig()
	if compressorAny, err = proto.ToAnyWithValidation(compressorProto); err != nil {
		return err
	}

	if brotli {
		brotliFilterName := compressorFilterName(egv1a1.BrotliCompressorType)
		if _, ok := perFilterCfg[brotliFilterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				brotliFilterName, route)
		}
		route.TypedPerFilterConfig[brotliFilterName] = compressorAny
	}
	if gzip {
		gzipFilterName := compressorFilterName(egv1a1.GzipCompressorType)
		if _, ok := perFilterCfg[gzipFilterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				gzipFilterName, route)
		}
		route.TypedPerFilterConfig[gzipFilterName] = compressorAny
	}

	return nil
}

func compressorPerRouteConfig() *compressorv3.CompressorPerRoute {
	// Enable compression on this route if compression is configured.
	return &compressorv3.CompressorPerRoute{
		Override: &compressorv3.CompressorPerRoute_Overrides{
			Overrides: &compressorv3.CompressorOverrides{
				ResponseDirectionConfig: &compressorv3.ResponseDirectionOverrides{},
			},
		},
	}
}
