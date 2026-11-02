// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package filters

import (
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	zstdv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/compressor/v3"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

// CompressorLibrary builds the compressor library config for the provided compressor type.
// Only the settings matching the compressor type are applied, unset settings use the Envoy defaults.
func CompressorLibrary(compressorType egv1a1.CompressorType, gzip *egv1a1.GzipCompressor,
	brotli *egv1a1.BrotliCompressor, zstd *egv1a1.ZstdCompressor,
) (*corev3.TypedExtensionConfig, error) {
	var (
		extensionName string
		extensionMsg  protobuf.Message
		extensionAny  *anypb.Any
		err           error
	)

	switch compressorType {
	case egv1a1.BrotliCompressorType:
		extensionName = "envoy.compression.brotli.compressor"
		extensionMsg, err = buildBrotliProto(brotli)
	case egv1a1.GzipCompressorType:
		extensionName = "envoy.compression.gzip.compressor"
		extensionMsg, err = buildGzipProto(gzip)
	case egv1a1.ZstdCompressorType:
		extensionName = "envoy.compression.zstd.compressor"
		extensionMsg, err = buildZstdProto(zstd)
	default:
		return nil, fmt.Errorf("unsupported compressor type: %s", compressorType)
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
