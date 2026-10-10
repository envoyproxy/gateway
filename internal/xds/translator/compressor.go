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
		compressorProto   *compressorv3.Compressor
		compressorLibrary *corev3.TypedExtensionConfig
		compressorAny     *anypb.Any
		err               error
	)

	// The filter is shared by all the routes of the listener, so it always uses the default
	// compressor library settings. Routes with custom settings override the compressor library.
	if compressorLibrary, err = buildCompressorLibrary(&ir.Compression{Type: compression.Type}); err != nil {
		return nil, err
	}

	compressorProto = &compressorv3.Compressor{
		CompressorLibrary: compressorLibrary,
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

// buildCompressorLibrary builds the compressor library config with the settings of the provided compression.
func buildCompressorLibrary(compression *ir.Compression) (*corev3.TypedExtensionConfig, error) {
	var (
		extensionName string
		extensionMsg  protobuf.Message
		extensionAny  *anypb.Any
		err           error
	)

	switch compression.Type {
	case egv1a1.BrotliCompressorType:
		extensionName = "envoy.compression.brotli.compressor"
		extensionMsg, err = buildBrotliProto(compression.Brotli)
	case egv1a1.GzipCompressorType:
		extensionName = "envoy.compression.gzip.compressor"
		extensionMsg, err = buildGzipProto(compression.Gzip)
	case egv1a1.ZstdCompressorType:
		extensionName = "envoy.compression.zstd.compressor"
		extensionMsg, err = buildZstdProto(compression.Zstd)
	}
	if err != nil {
		return nil, err
	}

	if extensionAny, err = proto.ToAnyWithValidation(extensionMsg); err != nil {
		return nil, err
	}

	return &corev3.TypedExtensionConfig{
		Name:        extensionName,
		TypedConfig: extensionAny,
	}, nil
}

var gzipCompressionStrategies = map[egv1a1.GzipCompressionStrategy]gzipv3.Gzip_CompressionStrategy{
	egv1a1.GzipCompressionStrategyDefault:     gzipv3.Gzip_DEFAULT_STRATEGY,
	egv1a1.GzipCompressionStrategyFiltered:    gzipv3.Gzip_FILTERED,
	egv1a1.GzipCompressionStrategyHuffmanOnly: gzipv3.Gzip_HUFFMAN_ONLY,
	egv1a1.GzipCompressionStrategyRLE:         gzipv3.Gzip_RLE,
	egv1a1.GzipCompressionStrategyFixed:       gzipv3.Gzip_FIXED,
}

// buildGzipProto builds the Gzip compressor library config from the API config.
func buildGzipProto(gzip *egv1a1.GzipCompressor) (*gzipv3.Gzip, error) {
	gzipProto := &gzipv3.Gzip{}
	if gzip == nil {
		return gzipProto, nil
	}

	if gzip.CompressionLevel != nil {
		// The API compression level 1-9 maps directly to the proto enum values,
		// e.g. 9 -> COMPRESSION_LEVEL_9.
		gzipProto.CompressionLevel = gzipv3.Gzip_CompressionLevel(*gzip.CompressionLevel)
	}
	if gzip.CompressionStrategy != nil {
		strategy, ok := gzipCompressionStrategies[*gzip.CompressionStrategy]
		if !ok {
			return nil, fmt.Errorf("unsupported gzip compression strategy: %s", *gzip.CompressionStrategy)
		}
		gzipProto.CompressionStrategy = strategy
	}
	if gzip.MemoryLevel != nil {
		gzipProto.MemoryLevel = wrapperspb.UInt32(*gzip.MemoryLevel)
	}
	if gzip.WindowBits != nil {
		gzipProto.WindowBits = wrapperspb.UInt32(*gzip.WindowBits)
	}
	if gzip.ChunkSize != nil {
		gzipProto.ChunkSize = wrapperspb.UInt32(*gzip.ChunkSize)
	}

	return gzipProto, nil
}

var brotliEncoderModes = map[egv1a1.BrotliEncoderMode]brotliv3.Brotli_EncoderMode{
	egv1a1.BrotliEncoderModeDefault: brotliv3.Brotli_DEFAULT,
	egv1a1.BrotliEncoderModeGeneric: brotliv3.Brotli_GENERIC,
	egv1a1.BrotliEncoderModeText:    brotliv3.Brotli_TEXT,
	egv1a1.BrotliEncoderModeFont:    brotliv3.Brotli_FONT,
}

// buildBrotliProto builds the Brotli compressor library config from the API config.
func buildBrotliProto(brotli *egv1a1.BrotliCompressor) (*brotliv3.Brotli, error) {
	brotliProto := &brotliv3.Brotli{}
	if brotli == nil {
		return brotliProto, nil
	}

	if brotli.Quality != nil {
		brotliProto.Quality = wrapperspb.UInt32(*brotli.Quality)
	}
	if brotli.EncoderMode != nil {
		encoderMode, ok := brotliEncoderModes[*brotli.EncoderMode]
		if !ok {
			return nil, fmt.Errorf("unsupported brotli encoder mode: %s", *brotli.EncoderMode)
		}
		brotliProto.EncoderMode = encoderMode
	}
	if brotli.WindowBits != nil {
		brotliProto.WindowBits = wrapperspb.UInt32(*brotli.WindowBits)
	}
	if brotli.InputBlockBits != nil {
		brotliProto.InputBlockBits = wrapperspb.UInt32(*brotli.InputBlockBits)
	}
	if brotli.ChunkSize != nil {
		brotliProto.ChunkSize = wrapperspb.UInt32(*brotli.ChunkSize)
	}
	if brotli.DisableLiteralContextModeling != nil {
		brotliProto.DisableLiteralContextModeling = *brotli.DisableLiteralContextModeling
	}

	return brotliProto, nil
}

var zstdStrategies = map[egv1a1.ZstdCompressionStrategy]zstdv3.Zstd_Strategy{
	egv1a1.ZstdCompressionStrategyDefault:  zstdv3.Zstd_DEFAULT,
	egv1a1.ZstdCompressionStrategyFast:     zstdv3.Zstd_FAST,
	egv1a1.ZstdCompressionStrategyDFast:    zstdv3.Zstd_DFAST,
	egv1a1.ZstdCompressionStrategyGreedy:   zstdv3.Zstd_GREEDY,
	egv1a1.ZstdCompressionStrategyLazy:     zstdv3.Zstd_LAZY,
	egv1a1.ZstdCompressionStrategyLazy2:    zstdv3.Zstd_LAZY2,
	egv1a1.ZstdCompressionStrategyBTLazy2:  zstdv3.Zstd_BTLAZY2,
	egv1a1.ZstdCompressionStrategyBTOpt:    zstdv3.Zstd_BTOPT,
	egv1a1.ZstdCompressionStrategyBTUltra:  zstdv3.Zstd_BTULTRA,
	egv1a1.ZstdCompressionStrategyBTUltra2: zstdv3.Zstd_BTULTRA2,
}

// buildZstdProto builds the Zstd compressor library config from the API config.
func buildZstdProto(zstd *egv1a1.ZstdCompressor) (*zstdv3.Zstd, error) {
	zstdProto := &zstdv3.Zstd{}
	if zstd == nil {
		return zstdProto, nil
	}

	if zstd.CompressionLevel != nil {
		zstdProto.CompressionLevel = wrapperspb.UInt32(*zstd.CompressionLevel)
	}
	if zstd.EnableChecksum != nil {
		zstdProto.EnableChecksum = *zstd.EnableChecksum
	}
	if zstd.Strategy != nil {
		strategy, ok := zstdStrategies[*zstd.Strategy]
		if !ok {
			return nil, fmt.Errorf("unsupported zstd compression strategy: %s", *zstd.Strategy)
		}
		zstdProto.Strategy = strategy
	}
	if zstd.ChunkSize != nil {
		zstdProto.ChunkSize = wrapperspb.UInt32(*zstd.ChunkSize)
	}

	return zstdProto, nil
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
		perFilterCfg = make(map[string]*anypb.Any)
		route.TypedPerFilterConfig = perFilterCfg
	}

	for _, irComp := range irRoute.Traffic.Compression {
		filterName := compressorFilterName(irComp.Type)
		if _, ok := perFilterCfg[filterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				filterName, route)
		}

		var compressorLibrary *corev3.TypedExtensionConfig
		if irComp.Gzip != nil || irComp.Brotli != nil || irComp.Zstd != nil {
			if compressorLibrary, err = buildCompressorLibrary(irComp); err != nil {
				return err
			}
		}

		if compressorAny, err = proto.ToAnyWithValidation(compressorPerRouteConfig(compressorLibrary)); err != nil {
			return err
		}

		perFilterCfg[filterName] = compressorAny
	}

	// Ensure accept-encoding from the request to prevent double compression.
	if !slices.Contains(route.RequestHeadersToRemove, "accept-encoding") {
		route.RequestHeadersToRemove = append(route.RequestHeadersToRemove, "accept-encoding")
	}

	return nil
}

// Enable compression on this route if compression is configured, and override the
// compressor library if the route has custom compressor settings.
func compressorPerRouteConfig(compressorLibrary *corev3.TypedExtensionConfig) *compressorv3.CompressorPerRoute {
	return &compressorv3.CompressorPerRoute{
		Override: &compressorv3.CompressorPerRoute_Overrides{
			Overrides: &compressorv3.CompressorOverrides{
				ResponseDirectionConfig: &compressorv3.ResponseDirectionOverrides{},
				CompressorLibrary:       compressorLibrary,
			},
		},
	}
}
