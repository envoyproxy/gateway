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
