// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	zstdv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/compressor/v3"
	compressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

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
		filter *hcmv3.HttpFilter
		err    error
	)

	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.Compression != nil {
			for _, irComp := range route.Traffic.Compression {
				filterName := compressorFilterName(irComp.Type)
				if !hcmContainsFilter(mgr, filterName) {
					if filter, err = buildCompressorFilter(irComp); err != nil {
						return err
					}
					mgr.HttpFilters = append(mgr.HttpFilters, filter)
				}
			}
		}
	}

	return err
}

func compressorFilterName(compressorType egv1a1.CompressorType) string {
	return fmt.Sprintf("%s.%s", egv1a1.EnvoyFilterCompressor.String(), strings.ToLower(string(compressorType)))
}

// buildCompressorFilter builds a compressor filter with the provided compressionType.
func buildCompressorFilter(compression *ir.Compression) (*hcmv3.HttpFilter, error) {
	var (
		compressorProto *compressorv3.Compressor
		extensionName   string
		extensionMsg    protobuf.Message
		extensionAny    *anypb.Any
		compressorAny   *anypb.Any
		err             error
	)

	switch compression.Type {
	case egv1a1.BrotliCompressorType:
		extensionName = "envoy.compression.brotli.compressor"
		extensionMsg = buildBrotliConfig(compression.Config)
	case egv1a1.GzipCompressorType:
		extensionName = "envoy.compression.gzip.compressor"
		extensionMsg = buildGzipConfig(compression.Config)
	case egv1a1.ZstdCompressorType:
		extensionName = "envoy.compression.zstd.compressor"
		extensionMsg = buildZstdConfig(compression.Config)
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

	if compression.ChooseFirst {
		compressorProto.ChooseFirst = true
	}

	if compression.MinContentLength != nil {
		compressorProto.ResponseDirectionConfig = &compressorv3.Compressor_ResponseDirectionConfig{
			CommonConfig: &compressorv3.Compressor_CommonDirectionConfig{
				MinContentLength: wrapperspb.UInt32(*compression.MinContentLength),
			},
		}
	}

	if compressorAny, err = proto.ToAnyWithValidation(compressorProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: compressorFilterName(compression.Type),
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
		perFilterCfg  map[string]*anypb.Any
		compressorAny *anypb.Any
		err           error
	)

	// Overwrite the HCM level filter config with the per route filter config.
	perFilterCfg = route.GetTypedPerFilterConfig()
	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	compressorProto := compressorPerRouteConfig()
	for _, irComp := range irRoute.Traffic.Compression {
		filterName := compressorFilterName(irComp.Type)
		if _, ok := perFilterCfg[filterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				filterName, route)
		}

		if compressorAny, err = proto.ToAnyWithValidation(compressorProto); err != nil {
			return err
		}

		route.TypedPerFilterConfig[filterName] = compressorAny
	}

	// Ensure accept-encoding from the request to prevent double compression.
	if !slices.Contains(route.RequestHeadersToRemove, "accept-encoding") {
		route.RequestHeadersToRemove = append(route.RequestHeadersToRemove, "accept-encoding")
	}

	return nil
}

// Enable compression on this route if compression is configured.
func compressorPerRouteConfig() *compressorv3.CompressorPerRoute {
	return &compressorv3.CompressorPerRoute{
		Override: &compressorv3.CompressorPerRoute_Overrides{
			Overrides: &compressorv3.CompressorOverrides{
				ResponseDirectionConfig: &compressorv3.ResponseDirectionOverrides{},
			},
		},
	}
}

func buildGzipConfig(cfg *egv1a1.Compression) *gzipv3.Gzip {
	gzip := &gzipv3.Gzip{}
	if cfg == nil || cfg.Gzip == nil {
		return gzip
	}

	if cfg.Gzip.MemoryLevel != nil {
		gzip.MemoryLevel = wrapperspb.UInt32(*cfg.Gzip.MemoryLevel)
	}
	if cfg.Gzip.WindowBits != nil {
		gzip.WindowBits = wrapperspb.UInt32(*cfg.Gzip.WindowBits)
	}
	if cfg.Gzip.ChunkSize != nil {
		gzip.ChunkSize = wrapperspb.UInt32(*cfg.Gzip.ChunkSize)
	}
	if cfg.Gzip.CompressionLevel != nil {
		gzip.CompressionLevel = mapGzipCompressionLevel(*cfg.Gzip.CompressionLevel)
	}
	if cfg.Gzip.CompressionStrategy != nil {
		gzip.CompressionStrategy = mapGzipCompressionStrategy(*cfg.Gzip.CompressionStrategy)
	}
	return gzip
}

func buildBrotliConfig(cfg *egv1a1.Compression) *brotliv3.Brotli {
	brotli := &brotliv3.Brotli{}
	if cfg == nil || cfg.Brotli == nil {
		return brotli
	}

	if cfg.Brotli.Quality != nil {
		brotli.Quality = wrapperspb.UInt32(*cfg.Brotli.Quality)
	}
	if cfg.Brotli.WindowBits != nil {
		brotli.WindowBits = wrapperspb.UInt32(*cfg.Brotli.WindowBits)
	}
	if cfg.Brotli.InputBlockBits != nil {
		brotli.InputBlockBits = wrapperspb.UInt32(*cfg.Brotli.InputBlockBits)
	}
	if cfg.Brotli.ChunkSize != nil {
		brotli.ChunkSize = wrapperspb.UInt32(*cfg.Brotli.ChunkSize)
	}
	if cfg.Brotli.EncoderMode != nil {
		brotli.EncoderMode = mapBrotliEncoderMode(*cfg.Brotli.EncoderMode)
	}
	if cfg.Brotli.DisableLiteralContextModeling != nil {
		brotli.DisableLiteralContextModeling = *cfg.Brotli.DisableLiteralContextModeling
	}
	return brotli
}

func buildZstdConfig(cfg *egv1a1.Compression) *zstdv3.Zstd {
	zstd := &zstdv3.Zstd{}
	if cfg == nil || cfg.Zstd == nil {
		return zstd
	}

	if cfg.Zstd.CompressionLevel != nil {
		zstd.CompressionLevel = wrapperspb.UInt32(*cfg.Zstd.CompressionLevel)
	}
	if cfg.Zstd.ChunkSize != nil {
		zstd.ChunkSize = wrapperspb.UInt32(*cfg.Zstd.ChunkSize)
	}
	if cfg.Zstd.EnableChecksum != nil {
		zstd.EnableChecksum = *cfg.Zstd.EnableChecksum
	}
	if cfg.Zstd.Strategy != nil {
		zstd.Strategy = mapZstdStrategy(*cfg.Zstd.Strategy)
	}
	return zstd
}

func mapGzipCompressionLevel(level string) gzipv3.Gzip_CompressionLevel {
	switch level {
	case "BEST_SPEED":
		return gzipv3.Gzip_BEST_SPEED
	case "BEST_COMPRESSION":
		return gzipv3.Gzip_BEST_COMPRESSION
	case "COMPRESSION_LEVEL_1":
		return gzipv3.Gzip_COMPRESSION_LEVEL_1
	case "COMPRESSION_LEVEL_2":
		return gzipv3.Gzip_COMPRESSION_LEVEL_2
	case "COMPRESSION_LEVEL_3":
		return gzipv3.Gzip_COMPRESSION_LEVEL_3
	case "COMPRESSION_LEVEL_4":
		return gzipv3.Gzip_COMPRESSION_LEVEL_4
	case "COMPRESSION_LEVEL_5":
		return gzipv3.Gzip_COMPRESSION_LEVEL_5
	case "COMPRESSION_LEVEL_6":
		return gzipv3.Gzip_COMPRESSION_LEVEL_6
	case "COMPRESSION_LEVEL_7":
		return gzipv3.Gzip_COMPRESSION_LEVEL_7
	case "COMPRESSION_LEVEL_8":
		return gzipv3.Gzip_COMPRESSION_LEVEL_8
	case "COMPRESSION_LEVEL_9":
		return gzipv3.Gzip_COMPRESSION_LEVEL_9
	default:
		return gzipv3.Gzip_DEFAULT_COMPRESSION
	}
}

func mapGzipCompressionStrategy(strategy string) gzipv3.Gzip_CompressionStrategy {
	switch strategy {
	case "FILTERED":
		return gzipv3.Gzip_FILTERED
	case "HUFFMAN_ONLY":
		return gzipv3.Gzip_HUFFMAN_ONLY
	case "RLE":
		return gzipv3.Gzip_RLE
	case "FIXED":
		return gzipv3.Gzip_FIXED
	default:
		return gzipv3.Gzip_DEFAULT_STRATEGY
	}
}

func mapBrotliEncoderMode(mode string) brotliv3.Brotli_EncoderMode {
	switch mode {
	case "GENERIC":
		return brotliv3.Brotli_GENERIC
	case "TEXT":
		return brotliv3.Brotli_TEXT
	case "FONT":
		return brotliv3.Brotli_FONT
	default:
		return brotliv3.Brotli_DEFAULT
	}
}

func mapZstdStrategy(strategy string) zstdv3.Zstd_Strategy {
	switch strategy {
	case "FAST":
		return zstdv3.Zstd_FAST
	case "DFAST":
		return zstdv3.Zstd_DFAST
	case "GREEDY":
		return zstdv3.Zstd_GREEDY
	case "LAZY":
		return zstdv3.Zstd_LAZY
	case "LAZY2":
		return zstdv3.Zstd_LAZY2
	case "BTLAZY2":
		return zstdv3.Zstd_BTLAZY2
	case "BTOPT":
		return zstdv3.Zstd_BTOPT
	case "BTULTRA":
		return zstdv3.Zstd_BTULTRA
	case "BTULTRA2":
		return zstdv3.Zstd_BTULTRA2
	default:
		return zstdv3.Zstd_DEFAULT
	}
}
