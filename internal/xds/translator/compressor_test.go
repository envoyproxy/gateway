// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"strings"
	"testing"

	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	zstdv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/compressor/v3"
	compressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
				MinContentLength: new(uint32(1024)),
			},
			expectedName:    "envoy.filters.http.compressor.gzip/2ee8b42b",
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
				MinContentLength: new(uint32(2048)),
			},
			expectedName:    "envoy.filters.http.compressor.brotli/2965facd",
			expectedExtName: "envoy.compression.brotli.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				assert.True(t, c.ChooseFirst)
				require.NotNil(t, c.ResponseDirectionConfig)
				assert.Equal(t, uint32(2048), c.ResponseDirectionConfig.CommonConfig.MinContentLength.Value)
			},
		},
		{
			name: "gzip compressor with custom settings",
			compression: &ir.Compression{
				Type: egv1a1.GzipCompressorType,
				Gzip: &egv1a1.GzipCompressor{
					CompressionLevel:    new(uint32(9)),
					CompressionStrategy: new(egv1a1.GzipCompressionStrategyRLE),
					MemoryLevel:         new(uint32(8)),
					WindowBits:          new(uint32(15)),
					ChunkSize:           new(uint32(8192)),
				},
			},
			expectedName:    "envoy.filters.http.compressor.gzip/fcbb148f",
			expectedExtName: "envoy.compression.gzip.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				gzip := &gzipv3.Gzip{}
				require.NoError(t, c.CompressorLibrary.TypedConfig.UnmarshalTo(gzip))
				assert.Equal(t, gzipv3.Gzip_COMPRESSION_LEVEL_9, gzip.CompressionLevel)
				assert.Equal(t, gzipv3.Gzip_RLE, gzip.CompressionStrategy)
				assert.Equal(t, uint32(8), gzip.MemoryLevel.Value)
				assert.Equal(t, uint32(15), gzip.WindowBits.Value)
				assert.Equal(t, uint32(8192), gzip.ChunkSize.Value)
			},
		},
		{
			name: "brotli compressor with custom settings",
			compression: &ir.Compression{
				Type: egv1a1.BrotliCompressorType,
				Brotli: &egv1a1.BrotliCompressor{
					Quality:                       new(uint32(11)),
					EncoderMode:                   new(egv1a1.BrotliEncoderModeText),
					WindowBits:                    new(uint32(24)),
					InputBlockBits:                new(uint32(16)),
					ChunkSize:                     new(uint32(4096)),
					DisableLiteralContextModeling: new(true),
				},
			},
			expectedName:    "envoy.filters.http.compressor.brotli/e3c96b84",
			expectedExtName: "envoy.compression.brotli.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				brotli := &brotliv3.Brotli{}
				require.NoError(t, c.CompressorLibrary.TypedConfig.UnmarshalTo(brotli))
				assert.Equal(t, uint32(11), brotli.Quality.Value)
				assert.Equal(t, brotliv3.Brotli_TEXT, brotli.EncoderMode)
				assert.Equal(t, uint32(24), brotli.WindowBits.Value)
				assert.Equal(t, uint32(16), brotli.InputBlockBits.Value)
				assert.Equal(t, uint32(4096), brotli.ChunkSize.Value)
				assert.True(t, brotli.DisableLiteralContextModeling)
			},
		},
		{
			name: "zstd compressor with custom settings",
			compression: &ir.Compression{
				Type: egv1a1.ZstdCompressorType,
				Zstd: &egv1a1.ZstdCompressor{
					CompressionLevel: new(uint32(22)),
					EnableChecksum:   new(true),
					Strategy:         new(egv1a1.ZstdCompressionStrategyBTUltra2),
					ChunkSize:        new(uint32(65536)),
				},
			},
			expectedName:    "envoy.filters.http.compressor.zstd/b29b3f6f",
			expectedExtName: "envoy.compression.zstd.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				zstd := &zstdv3.Zstd{}
				require.NoError(t, c.CompressorLibrary.TypedConfig.UnmarshalTo(zstd))
				assert.Equal(t, uint32(22), zstd.CompressionLevel.Value)
				assert.True(t, zstd.EnableChecksum)
				assert.Equal(t, zstdv3.Zstd_BTULTRA2, zstd.Strategy)
				assert.Equal(t, uint32(65536), zstd.ChunkSize.Value)
			},
		},
		{
			name: "compressor settings for a different type are ignored",
			compression: &ir.Compression{
				Type: egv1a1.GzipCompressorType,
				Brotli: &egv1a1.BrotliCompressor{
					Quality: new(uint32(11)),
				},
			},
			expectedName:    "envoy.filters.http.compressor.gzip/f55d6fbc",
			expectedExtName: "envoy.compression.gzip.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				gzip := &gzipv3.Gzip{}
				require.NoError(t, c.CompressorLibrary.TypedConfig.UnmarshalTo(gzip))
				assert.Nil(t, gzip.MemoryLevel)
				assert.Nil(t, gzip.WindowBits)
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
	gzipCompression := &ir.Compression{
		Type: egv1a1.GzipCompressorType,
		Gzip: &egv1a1.GzipCompressor{
			CompressionLevel:    new(uint32(9)),
			CompressionStrategy: new(egv1a1.GzipCompressionStrategyRLE),
		},
	}
	identicalGzipCompression := &ir.Compression{
		Type: egv1a1.GzipCompressorType,
		Gzip: &egv1a1.GzipCompressor{
			CompressionLevel:    new(uint32(9)),
			CompressionStrategy: new(egv1a1.GzipCompressionStrategyRLE),
		},
	}
	differentGzipCompression := &ir.Compression{
		Type: egv1a1.GzipCompressorType,
		Gzip: &egv1a1.GzipCompressor{
			CompressionLevel:    new(uint32(1)),
			CompressionStrategy: new(egv1a1.GzipCompressionStrategyRLE),
		},
	}

	tests := []struct {
		name        string
		compression *ir.Compression
		want        string
	}{
		{
			name: "bare gzip",
			compression: &ir.Compression{
				Type: egv1a1.GzipCompressorType,
			},
			want: "envoy.filters.http.compressor.gzip",
		},
		{
			name: "bare brotli",
			compression: &ir.Compression{
				Type: egv1a1.BrotliCompressorType,
			},
			want: "envoy.filters.http.compressor.brotli",
		},
		{
			name: "bare zstd",
			compression: &ir.Compression{
				Type: egv1a1.ZstdCompressorType,
			},
			want: "envoy.filters.http.compressor.zstd",
		},
		{
			name: "choose first only",
			compression: &ir.Compression{
				Type:        egv1a1.GzipCompressorType,
				ChooseFirst: true,
			},
			want: "envoy.filters.http.compressor.gzip",
		},
		{
			name:        "custom gzip",
			compression: gzipCompression,
			want:        "envoy.filters.http.compressor.gzip/7a9fd8cc",
		},
		{
			name: "min content length",
			compression: &ir.Compression{
				Type:             egv1a1.GzipCompressorType,
				MinContentLength: new(uint32(1024)),
			},
			want: "envoy.filters.http.compressor.gzip/2ee8b42b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compressorFilterName(tt.compression)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	customName, err := compressorFilterName(gzipCompression)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(customName, "envoy.filters.http.compressor.gzip/"))
	assert.Len(t, strings.TrimPrefix(customName, "envoy.filters.http.compressor.gzip/"), 8)

	identicalName, err := compressorFilterName(identicalGzipCompression)
	require.NoError(t, err)
	assert.Equal(t, customName, identicalName)

	differentName, err := compressorFilterName(differentGzipCompression)
	require.NoError(t, err)
	assert.NotEqual(t, customName, differentName)
}
