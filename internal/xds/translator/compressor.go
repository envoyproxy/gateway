// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"encoding/json"
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
	"github.com/envoyproxy/gateway/internal/utils"
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

	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.Compression != nil {
			for _, irComp := range route.Traffic.Compression {
				filterName, err := compressorFilterName(irComp)
				if err != nil {
					return err
				}
				if !hcmContainsFilter(mgr, filterName) {
					filter, err := buildCompressorFilter(irComp)
					if err != nil {
						return err
					}
					mgr.HttpFilters = append(mgr.HttpFilters, filter)
				}
			}
		}
	}

	return nil
}

func compressorFilterName(compression *ir.Compression) (string, error) {
	filterName := fmt.Sprintf("%s.%s", egv1a1.EnvoyFilterCompressor.String(), strings.ToLower(string(compression.Type)))
	if !hasCustomCompressorSettings(compression) {
		return filterName, nil
	}

	compressionJSON, err := json.Marshal(compression)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", filterName, utils.Digest32(string(compressionJSON))), nil
}

func hasCustomCompressorSettings(compression *ir.Compression) bool {
	return compression.MinContentLength != nil ||
		compression.Gzip != nil ||
		compression.Brotli != nil ||
		compression.Zstd != nil
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
		extensionMsg = buildBrotliProto(compression.Brotli)
	case egv1a1.GzipCompressorType:
		extensionName = "envoy.compression.gzip.compressor"
		extensionMsg = buildGzipProto(compression.Gzip)
	case egv1a1.ZstdCompressorType:
		extensionName = "envoy.compression.zstd.compressor"
		extensionMsg = buildZstdProto(compression.Zstd)
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

	filterName, err := compressorFilterName(compression)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: filterName,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: compressorAny,
		},
		Disabled: true,
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
func buildGzipProto(gzip *egv1a1.GzipCompressor) *gzipv3.Gzip {
	gzipProto := &gzipv3.Gzip{}
	if gzip == nil {
		return gzipProto
	}

	if gzip.CompressionLevel != nil {
		// The API compression level 1-9 maps directly to the proto enum values,
		// e.g. 9 -> COMPRESSION_LEVEL_9.
		gzipProto.CompressionLevel = gzipv3.Gzip_CompressionLevel(*gzip.CompressionLevel)
	}
	if gzip.CompressionStrategy != nil {
		gzipProto.CompressionStrategy = gzipCompressionStrategies[*gzip.CompressionStrategy]
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

	return gzipProto
}

var brotliEncoderModes = map[egv1a1.BrotliEncoderMode]brotliv3.Brotli_EncoderMode{
	egv1a1.BrotliEncoderModeDefault: brotliv3.Brotli_DEFAULT,
	egv1a1.BrotliEncoderModeGeneric: brotliv3.Brotli_GENERIC,
	egv1a1.BrotliEncoderModeText:    brotliv3.Brotli_TEXT,
	egv1a1.BrotliEncoderModeFont:    brotliv3.Brotli_FONT,
}

// buildBrotliProto builds the Brotli compressor library config from the API config.
func buildBrotliProto(brotli *egv1a1.BrotliCompressor) *brotliv3.Brotli {
	brotliProto := &brotliv3.Brotli{}
	if brotli == nil {
		return brotliProto
	}

	if brotli.Quality != nil {
		brotliProto.Quality = wrapperspb.UInt32(*brotli.Quality)
	}
	if brotli.EncoderMode != nil {
		brotliProto.EncoderMode = brotliEncoderModes[*brotli.EncoderMode]
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

	return brotliProto
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
func buildZstdProto(zstd *egv1a1.ZstdCompressor) *zstdv3.Zstd {
	zstdProto := &zstdv3.Zstd{}
	if zstd == nil {
		return zstdProto
	}

	if zstd.CompressionLevel != nil {
		zstdProto.CompressionLevel = wrapperspb.UInt32(*zstd.CompressionLevel)
	}
	if zstd.EnableChecksum != nil {
		zstdProto.EnableChecksum = *zstd.EnableChecksum
	}
	if zstd.Strategy != nil {
		zstdProto.Strategy = zstdStrategies[*zstd.Strategy]
	}
	if zstd.ChunkSize != nil {
		zstdProto.ChunkSize = wrapperspb.UInt32(*zstd.ChunkSize)
	}

	return zstdProto
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

	// Overwrite the HCM level filter config with the per route filter config.
	perFilterCfg := route.GetTypedPerFilterConfig()
	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	compressorProto := compressorPerRouteConfig()
	for _, irComp := range irRoute.Traffic.Compression {
		filterName, err := compressorFilterName(irComp)
		if err != nil {
			return err
		}
		if _, ok := perFilterCfg[filterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				filterName, route)
		}

		compressorAny, err := proto.ToAnyWithValidation(compressorProto)
		if err != nil {
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
