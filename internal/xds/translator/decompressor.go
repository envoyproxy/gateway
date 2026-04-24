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
	brotlidecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/decompressor/v3"
	gzipdecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/decompressor/v3"
	zstddecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/decompressor/v3"
	decompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/decompressor/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&decompressor{})
}

type decompressor struct{}

var _ httpFilter = &decompressor{}

// patchHCM builds and appends the decompressor Filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
func (*decompressor) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var (
		filter *hcmv3.HttpFilter
		err    error
	)

	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.Decompressor != nil {
			for _, irDecomp := range route.Traffic.Decompressor {
				filterName := decompressorFilterName(irDecomp.Type)
				if !hcmContainsFilter(mgr, filterName) {
					if filter, err = buildDecompressorFilter(irDecomp); err != nil {
						return err
					}
					mgr.HttpFilters = append(mgr.HttpFilters, filter)
				}
			}
		}
	}

	return err
}

func decompressorFilterName(decompressorType egv1a1.DecompressorType) string {
	return fmt.Sprintf("%s.%s", egv1a1.EnvoyFilterDecompressor.String(), strings.ToLower(string(decompressorType)))
}

// buildDecompressorFilter builds a decompressor filter with the provided decompressor type.
func buildDecompressorFilter(decompression *ir.Decompressor) (*hcmv3.HttpFilter, error) {
	var (
		decompressorProto *decompressorv3.Decompressor
		extensionName     string
		extensionMsg      protobuf.Message
		extensionAny      *anypb.Any
		decompressorAny   *anypb.Any
		err               error
	)

	switch decompression.Type {
	case egv1a1.BrotliDecompressorType:
		extensionName = "envoy.compression.brotli.decompressor"
		extensionMsg = &brotlidecompressorv3.Brotli{}
	case egv1a1.GzipDecompressorType:
		extensionName = "envoy.compression.gzip.decompressor"
		extensionMsg = &gzipdecompressorv3.Gzip{}
	case egv1a1.ZstdDecompressorType:
		extensionName = "envoy.compression.zstd.decompressor"
		extensionMsg = &zstddecompressorv3.Zstd{}
	}

	if extensionAny, err = proto.ToAnyWithValidation(extensionMsg); err != nil {
		return nil, err
	}

	decompressorProto = &decompressorv3.Decompressor{
		DecompressorLibrary: &corev3.TypedExtensionConfig{
			Name:        extensionName,
			TypedConfig: extensionAny,
		},
		// Enable request decompression (decompress compressed requests from clients).
		RequestDirectionConfig: &decompressorv3.Decompressor_RequestDirectionConfig{},
		// Enable response decompression (decompress compressed responses from backends).
		ResponseDirectionConfig: &decompressorv3.Decompressor_ResponseDirectionConfig{},
	}

	if decompressorAny, err = proto.ToAnyWithValidation(decompressorProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: decompressorFilterName(decompression.Type),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: decompressorAny,
		},
		// The decompressor filter is disabled by default and enabled per-route.
		Disabled: true,
	}, nil
}

func (*decompressor) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the decompressor config if applicable.
func (*decompressor) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Traffic == nil || len(irRoute.Traffic.Decompressor) == 0 {
		return nil
	}

	var (
		perFilterCfg    map[string]*anypb.Any
		decompressorAny *anypb.Any
		err             error
	)

	// Overwrite the HCM level filter config with the per route filter config.
	perFilterCfg = route.GetTypedPerFilterConfig()
	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	decompressorProto := decompressorPerRouteConfig()
	for _, irDecomp := range irRoute.Traffic.Decompressor {
		filterName := decompressorFilterName(irDecomp.Type)
		if _, ok := perFilterCfg[filterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				filterName, route)
		}

		if decompressorAny, err = proto.ToAnyWithValidation(decompressorProto); err != nil {
			return err
		}

		route.TypedPerFilterConfig[filterName] = decompressorAny
	}

	return nil
}

// decompressorPerRouteConfig returns a per-route config that enables the decompressor filter.
// The decompressor filter uses the route-level FilterConfig to enable/disable per route.
func decompressorPerRouteConfig() *routev3.FilterConfig {
	return &routev3.FilterConfig{}
}
