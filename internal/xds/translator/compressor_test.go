// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	zstdv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/compressor/v3"
	compressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildCompressorFilter(t *testing.T) {
	tests := []struct {
		name            string
		compression     *ir.Compression
		expectedName    string
		expectedExtName string
		validateProto   func(*testing.T, *compressorv3.Compressor)
	}{
		{
			name: "gzip compressor",
			compression: &ir.Compression{
				Type: egv1a1.GzipCompressorType,
			},
			expectedName:    "envoy.filters.http.compressor.gzip",
			expectedExtName: "envoy.compression.gzip.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				gzip := &gzipv3.Gzip{}
				require.NoError(t, c.CompressorLibrary.TypedConfig.UnmarshalTo(gzip))
				assert.False(t, c.ChooseFirst)
				assert.Nil(t, c.ResponseDirectionConfig)
			},
		},
		{
			name: "brotli compressor",
			compression: &ir.Compression{
				Type: egv1a1.BrotliCompressorType,
			},
			expectedName:    "envoy.filters.http.compressor.brotli",
			expectedExtName: "envoy.compression.brotli.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				brotli := &brotliv3.Brotli{}
				require.NoError(t, c.CompressorLibrary.TypedConfig.UnmarshalTo(brotli))
			},
		},
		{
			name: "zstd compressor",
			compression: &ir.Compression{
				Type: egv1a1.ZstdCompressorType,
			},
			expectedName:    "envoy.filters.http.compressor.zstd",
			expectedExtName: "envoy.compression.zstd.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				zstd := &zstdv3.Zstd{}
				require.NoError(t, c.CompressorLibrary.TypedConfig.UnmarshalTo(zstd))
			},
		},
		{
			name: "with choose first",
			compression: &ir.Compression{
				Type:        egv1a1.GzipCompressorType,
				ChooseFirst: true,
			},
			expectedName:    "envoy.filters.http.compressor.gzip",
			expectedExtName: "envoy.compression.gzip.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				assert.True(t, c.ChooseFirst)
			},
		},
		{
			name: "with min content length",
			compression: &ir.Compression{
				Type:             egv1a1.GzipCompressorType,
				MinContentLength: ptr.To(uint32(1024)),
			},
			expectedName:    "envoy.filters.http.compressor.gzip",
			expectedExtName: "envoy.compression.gzip.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				require.NotNil(t, c.ResponseDirectionConfig)
				require.NotNil(t, c.ResponseDirectionConfig.CommonConfig)
				require.NotNil(t, c.ResponseDirectionConfig.CommonConfig.MinContentLength)
				assert.Equal(t, uint32(1024), c.ResponseDirectionConfig.CommonConfig.MinContentLength.Value)
			},
		},
		{
			name: "with all options",
			compression: &ir.Compression{
				Type:             egv1a1.BrotliCompressorType,
				ChooseFirst:      true,
				MinContentLength: ptr.To(uint32(2048)),
			},
			expectedName:    "envoy.filters.http.compressor.brotli",
			expectedExtName: "envoy.compression.brotli.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				assert.True(t, c.ChooseFirst)
				require.NotNil(t, c.ResponseDirectionConfig)
				assert.Equal(t, uint32(2048), c.ResponseDirectionConfig.CommonConfig.MinContentLength.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := buildCompressorFilter(tt.compression)
			require.NoError(t, err)
			require.NotNil(t, filter)

			assert.Equal(t, tt.expectedName, filter.Name)
			assert.True(t, filter.Disabled)

			compressorProto := &compressorv3.Compressor{}
			require.NoError(t, filter.GetTypedConfig().UnmarshalTo(compressorProto))
			assert.Equal(t, tt.expectedExtName, compressorProto.CompressorLibrary.Name)

			if tt.validateProto != nil {
				tt.validateProto(t, compressorProto)
			}
		})
	}
}

func TestCompressorFilterName(t *testing.T) {
	tests := []struct {
		compressorType egv1a1.CompressorType
		want           string
	}{
		{egv1a1.GzipCompressorType, "envoy.filters.http.compressor.gzip"},
		{egv1a1.BrotliCompressorType, "envoy.filters.http.compressor.brotli"},
		{egv1a1.ZstdCompressorType, "envoy.filters.http.compressor.zstd"},
	}

	for _, tt := range tests {
		t.Run(string(tt.compressorType), func(t *testing.T) {
			assert.Equal(t, tt.want, compressorFilterName(tt.compressorType))
		})
	}
}

func TestBuildGzipConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *egv1a1.Compression
		validate func(*testing.T, *gzipv3.Gzip)
	}{
		{
			name: "nil config",
			cfg:  nil,
			validate: func(t *testing.T, gzip *gzipv3.Gzip) {
				assert.Nil(t, gzip.MemoryLevel)
				assert.Nil(t, gzip.WindowBits)
				assert.Nil(t, gzip.ChunkSize)
			},
		},
		{
			name: "nil gzip config",
			cfg:  &egv1a1.Compression{},
			validate: func(t *testing.T, gzip *gzipv3.Gzip) {
				assert.Nil(t, gzip.MemoryLevel)
			},
		},
		{
			name: "with all fields",
			cfg: &egv1a1.Compression{
				Gzip: &egv1a1.GzipCompressor{
					MemoryLevel:         ptr.To(uint32(8)),
					WindowBits:          ptr.To(uint32(15)),
					ChunkSize:           ptr.To(uint32(4096)),
					CompressionLevel:    ptr.To("BEST_SPEED"),
					CompressionStrategy: ptr.To("FILTERED"),
				},
			},
			validate: func(t *testing.T, gzip *gzipv3.Gzip) {
				assert.Equal(t, uint32(8), gzip.MemoryLevel.Value)
				assert.Equal(t, uint32(15), gzip.WindowBits.Value)
				assert.Equal(t, uint32(4096), gzip.ChunkSize.Value)
				assert.Equal(t, gzipv3.Gzip_BEST_SPEED, gzip.CompressionLevel)
				assert.Equal(t, gzipv3.Gzip_FILTERED, gzip.CompressionStrategy)
			},
		},
		{
			name: "with partial fields",
			cfg: &egv1a1.Compression{
				Gzip: &egv1a1.GzipCompressor{
					CompressionLevel: ptr.To("COMPRESSION_LEVEL_6"),
				},
			},
			validate: func(t *testing.T, gzip *gzipv3.Gzip) {
				assert.Nil(t, gzip.MemoryLevel)
				assert.Equal(t, gzipv3.Gzip_COMPRESSION_LEVEL_6, gzip.CompressionLevel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gzip := buildGzipConfig(tt.cfg)
			require.NotNil(t, gzip)
			tt.validate(t, gzip)
		})
	}
}

func TestBuildBrotliConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *egv1a1.Compression
		validate func(*testing.T, *brotliv3.Brotli)
	}{
		{
			name: "nil config",
			cfg:  nil,
			validate: func(t *testing.T, brotli *brotliv3.Brotli) {
				assert.Nil(t, brotli.Quality)
				assert.Nil(t, brotli.WindowBits)
				assert.Nil(t, brotli.ChunkSize)
			},
		},
		{
			name: "nil brotli config",
			cfg:  &egv1a1.Compression{},
			validate: func(t *testing.T, brotli *brotliv3.Brotli) {
				assert.Nil(t, brotli.Quality)
			},
		},
		{
			name: "with all fields",
			cfg: &egv1a1.Compression{
				Brotli: &egv1a1.BrotliCompressor{
					Quality:                       ptr.To(uint32(5)),
					WindowBits:                    ptr.To(uint32(22)),
					InputBlockBits:                ptr.To(uint32(24)),
					ChunkSize:                     ptr.To(uint32(8192)),
					EncoderMode:                   ptr.To("TEXT"),
					DisableLiteralContextModeling: ptr.To(true),
				},
			},
			validate: func(t *testing.T, brotli *brotliv3.Brotli) {
				assert.Equal(t, uint32(5), brotli.Quality.Value)
				assert.Equal(t, uint32(22), brotli.WindowBits.Value)
				assert.Equal(t, uint32(24), brotli.InputBlockBits.Value)
				assert.Equal(t, uint32(8192), brotli.ChunkSize.Value)
				assert.Equal(t, brotliv3.Brotli_TEXT, brotli.EncoderMode)
				assert.True(t, brotli.DisableLiteralContextModeling)
			},
		},
		{
			name: "with partial fields",
			cfg: &egv1a1.Compression{
				Brotli: &egv1a1.BrotliCompressor{
					Quality:     ptr.To(uint32(11)),
					EncoderMode: ptr.To("GENERIC"),
				},
			},
			validate: func(t *testing.T, brotli *brotliv3.Brotli) {
				assert.Equal(t, uint32(11), brotli.Quality.Value)
				assert.Equal(t, brotliv3.Brotli_GENERIC, brotli.EncoderMode)
				assert.Nil(t, brotli.WindowBits)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			brotli := buildBrotliConfig(tt.cfg)
			require.NotNil(t, brotli)
			tt.validate(t, brotli)
		})
	}
}

func TestBuildZstdConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *egv1a1.Compression
		validate func(*testing.T, *zstdv3.Zstd)
	}{
		{
			name: "nil config",
			cfg:  nil,
			validate: func(t *testing.T, zstd *zstdv3.Zstd) {
				assert.Nil(t, zstd.CompressionLevel)
				assert.Nil(t, zstd.ChunkSize)
				assert.False(t, zstd.EnableChecksum)
			},
		},
		{
			name: "nil zstd config",
			cfg:  &egv1a1.Compression{},
			validate: func(t *testing.T, zstd *zstdv3.Zstd) {
				assert.Nil(t, zstd.CompressionLevel)
			},
		},
		{
			name: "with all fields",
			cfg: &egv1a1.Compression{
				Zstd: &egv1a1.ZstdCompressor{
					CompressionLevel: ptr.To(uint32(10)),
					ChunkSize:        ptr.To(uint32(16384)),
					EnableChecksum:   ptr.To(true),
					Strategy:         ptr.To("BTULTRA"),
				},
			},
			validate: func(t *testing.T, zstd *zstdv3.Zstd) {
				assert.Equal(t, uint32(10), zstd.CompressionLevel.Value)
				assert.Equal(t, uint32(16384), zstd.ChunkSize.Value)
				assert.True(t, zstd.EnableChecksum)
				assert.Equal(t, zstdv3.Zstd_BTULTRA, zstd.Strategy)
			},
		},
		{
			name: "with partial fields",
			cfg: &egv1a1.Compression{
				Zstd: &egv1a1.ZstdCompressor{
					Strategy:       ptr.To("LAZY"),
					EnableChecksum: ptr.To(false),
				},
			},
			validate: func(t *testing.T, zstd *zstdv3.Zstd) {
				assert.Nil(t, zstd.CompressionLevel)
				assert.Equal(t, zstdv3.Zstd_LAZY, zstd.Strategy)
				assert.False(t, zstd.EnableChecksum)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zstd := buildZstdConfig(tt.cfg)
			require.NotNil(t, zstd)
			tt.validate(t, zstd)
		})
	}
}

func TestMapGzipCompressionLevel(t *testing.T) {
	tests := []struct {
		level string
		want  gzipv3.Gzip_CompressionLevel
	}{
		{"BEST_SPEED", gzipv3.Gzip_BEST_SPEED},
		{"BEST_COMPRESSION", gzipv3.Gzip_BEST_COMPRESSION},
		{"COMPRESSION_LEVEL_1", gzipv3.Gzip_COMPRESSION_LEVEL_1},
		{"COMPRESSION_LEVEL_2", gzipv3.Gzip_COMPRESSION_LEVEL_2},
		{"COMPRESSION_LEVEL_3", gzipv3.Gzip_COMPRESSION_LEVEL_3},
		{"COMPRESSION_LEVEL_4", gzipv3.Gzip_COMPRESSION_LEVEL_4},
		{"COMPRESSION_LEVEL_5", gzipv3.Gzip_COMPRESSION_LEVEL_5},
		{"COMPRESSION_LEVEL_6", gzipv3.Gzip_COMPRESSION_LEVEL_6},
		{"COMPRESSION_LEVEL_7", gzipv3.Gzip_COMPRESSION_LEVEL_7},
		{"COMPRESSION_LEVEL_8", gzipv3.Gzip_COMPRESSION_LEVEL_8},
		{"COMPRESSION_LEVEL_9", gzipv3.Gzip_COMPRESSION_LEVEL_9},
		{"UNKNOWN", gzipv3.Gzip_DEFAULT_COMPRESSION},
		{"", gzipv3.Gzip_DEFAULT_COMPRESSION},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			assert.Equal(t, tt.want, mapGzipCompressionLevel(tt.level))
		})
	}
}

func TestMapGzipCompressionStrategy(t *testing.T) {
	tests := []struct {
		strategy string
		want     gzipv3.Gzip_CompressionStrategy
	}{
		{"FILTERED", gzipv3.Gzip_FILTERED},
		{"HUFFMAN_ONLY", gzipv3.Gzip_HUFFMAN_ONLY},
		{"RLE", gzipv3.Gzip_RLE},
		{"FIXED", gzipv3.Gzip_FIXED},
		{"UNKNOWN", gzipv3.Gzip_DEFAULT_STRATEGY},
		{"", gzipv3.Gzip_DEFAULT_STRATEGY},
	}

	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			assert.Equal(t, tt.want, mapGzipCompressionStrategy(tt.strategy))
		})
	}
}

func TestMapBrotliEncoderMode(t *testing.T) {
	tests := []struct {
		mode string
		want brotliv3.Brotli_EncoderMode
	}{
		{"GENERIC", brotliv3.Brotli_GENERIC},
		{"TEXT", brotliv3.Brotli_TEXT},
		{"FONT", brotliv3.Brotli_FONT},
		{"UNKNOWN", brotliv3.Brotli_DEFAULT},
		{"", brotliv3.Brotli_DEFAULT},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			assert.Equal(t, tt.want, mapBrotliEncoderMode(tt.mode))
		})
	}
}

func TestMapZstdStrategy(t *testing.T) {
	tests := []struct {
		strategy string
		want     zstdv3.Zstd_Strategy
	}{
		{"FAST", zstdv3.Zstd_FAST},
		{"DFAST", zstdv3.Zstd_DFAST},
		{"GREEDY", zstdv3.Zstd_GREEDY},
		{"LAZY", zstdv3.Zstd_LAZY},
		{"LAZY2", zstdv3.Zstd_LAZY2},
		{"BTLAZY2", zstdv3.Zstd_BTLAZY2},
		{"BTOPT", zstdv3.Zstd_BTOPT},
		{"BTULTRA", zstdv3.Zstd_BTULTRA},
		{"BTULTRA2", zstdv3.Zstd_BTULTRA2},
		{"UNKNOWN", zstdv3.Zstd_DEFAULT},
		{"", zstdv3.Zstd_DEFAULT},
	}

	for _, tt := range tests {
		t.Run(tt.strategy, func(t *testing.T) {
			assert.Equal(t, tt.want, mapZstdStrategy(tt.strategy))
		})
	}
}
