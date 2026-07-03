// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	zstdv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/compressor/v3"
	compressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	protobuf "google.golang.org/protobuf/proto"

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
				MinContentLength: new(uint32(2048)),
			},
			expectedName:    "envoy.filters.http.compressor.brotli",
			expectedExtName: "envoy.compression.brotli.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				assert.True(t, c.ChooseFirst)
				require.NotNil(t, c.ResponseDirectionConfig)
				assert.Equal(t, uint32(2048), c.ResponseDirectionConfig.CommonConfig.MinContentLength.Value)
			},
		},
		{
			name: "custom settings are not applied to the filter",
			compression: &ir.Compression{
				Type: egv1a1.GzipCompressorType,
				Gzip: &egv1a1.GzipCompressor{
					CompressionLevel: new(uint32(9)),
				},
			},
			expectedName:    "envoy.filters.http.compressor.gzip",
			expectedExtName: "envoy.compression.gzip.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				gzip := &gzipv3.Gzip{}
				require.NoError(t, c.CompressorLibrary.TypedConfig.UnmarshalTo(gzip))
				assert.True(t, protobuf.Equal(&gzipv3.Gzip{}, gzip))
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

func TestBuildCompressorLibrary(t *testing.T) {
	tests := []struct {
		name            string
		compression     *ir.Compression
		expectedExtName string
		expectedError   string
		validateProto   func(*testing.T, *corev3.TypedExtensionConfig)
	}{
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
			expectedExtName: "envoy.compression.gzip.compressor",
			validateProto: func(t *testing.T, c *corev3.TypedExtensionConfig) {
				gzip := &gzipv3.Gzip{}
				require.NoError(t, c.TypedConfig.UnmarshalTo(gzip))
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
			expectedExtName: "envoy.compression.brotli.compressor",
			validateProto: func(t *testing.T, c *corev3.TypedExtensionConfig) {
				brotli := &brotliv3.Brotli{}
				require.NoError(t, c.TypedConfig.UnmarshalTo(brotli))
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
			expectedExtName: "envoy.compression.zstd.compressor",
			validateProto: func(t *testing.T, c *corev3.TypedExtensionConfig) {
				zstd := &zstdv3.Zstd{}
				require.NoError(t, c.TypedConfig.UnmarshalTo(zstd))
				assert.Equal(t, uint32(22), zstd.CompressionLevel.Value)
				assert.True(t, zstd.EnableChecksum)
				assert.Equal(t, zstdv3.Zstd_BTULTRA2, zstd.Strategy)
				assert.Equal(t, uint32(65536), zstd.ChunkSize.Value)
			},
		},
		{
			name: "unsupported gzip compression strategy",
			compression: &ir.Compression{
				Type: egv1a1.GzipCompressorType,
				Gzip: &egv1a1.GzipCompressor{
					CompressionStrategy: new(egv1a1.GzipCompressionStrategy("Unknown")),
				},
			},
			expectedError: "unsupported gzip compression strategy: Unknown",
		},
		{
			name: "unsupported brotli encoder mode",
			compression: &ir.Compression{
				Type: egv1a1.BrotliCompressorType,
				Brotli: &egv1a1.BrotliCompressor{
					EncoderMode: new(egv1a1.BrotliEncoderMode("Unknown")),
				},
			},
			expectedError: "unsupported brotli encoder mode: Unknown",
		},
		{
			name: "unsupported zstd compression strategy",
			compression: &ir.Compression{
				Type: egv1a1.ZstdCompressorType,
				Zstd: &egv1a1.ZstdCompressor{
					Strategy: new(egv1a1.ZstdCompressionStrategy("Unknown")),
				},
			},
			expectedError: "unsupported zstd compression strategy: Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			library, err := buildCompressorLibrary(tt.compression)
			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedExtName, library.Name)

			if tt.validateProto != nil {
				tt.validateProto(t, library)
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
