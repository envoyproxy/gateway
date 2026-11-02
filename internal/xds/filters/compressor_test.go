// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package filters

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	zstdv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/compressor/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestCompressorLibrary(t *testing.T) {
	tests := []struct {
		name            string
		compression     *egv1a1.Compression
		expectedExtName string
		expectedError   string
		validateProto   func(*testing.T, *corev3.TypedExtensionConfig)
	}{
		{
			name: "gzip compressor with custom settings",
			compression: &egv1a1.Compression{
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
			compression: &egv1a1.Compression{
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
			compression: &egv1a1.Compression{
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
			compression: &egv1a1.Compression{
				Type: egv1a1.GzipCompressorType,
				Gzip: &egv1a1.GzipCompressor{
					CompressionStrategy: new(egv1a1.GzipCompressionStrategy("Unknown")),
				},
			},
			expectedError: "unsupported gzip compression strategy: Unknown",
		},
		{
			name: "unsupported brotli encoder mode",
			compression: &egv1a1.Compression{
				Type: egv1a1.BrotliCompressorType,
				Brotli: &egv1a1.BrotliCompressor{
					EncoderMode: new(egv1a1.BrotliEncoderMode("Unknown")),
				},
			},
			expectedError: "unsupported brotli encoder mode: Unknown",
		},
		{
			name: "unsupported zstd compression strategy",
			compression: &egv1a1.Compression{
				Type: egv1a1.ZstdCompressorType,
				Zstd: &egv1a1.ZstdCompressor{
					Strategy: new(egv1a1.ZstdCompressionStrategy("Unknown")),
				},
			},
			expectedError: "unsupported zstd compression strategy: Unknown",
		},
		{
			name: "unsupported compressor type",
			compression: &egv1a1.Compression{
				Type: egv1a1.CompressorType("Unknown"),
			},
			expectedError: "unsupported compressor type: Unknown",
		},
		{
			name: "settings rejected by the Envoy validation",
			compression: &egv1a1.Compression{
				Type: egv1a1.GzipCompressorType,
				Gzip: &egv1a1.GzipCompressor{
					WindowBits: new(uint32(16)),
				},
			},
			expectedError: "invalid Gzip.WindowBits: value must be inside range [9, 15]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			library, err := CompressorLibrary(tt.compression.Type, tt.compression.Gzip, tt.compression.Brotli, tt.compression.Zstd)
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
