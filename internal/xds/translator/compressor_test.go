// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	zstdv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/compressor/v3"
	compressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
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
			name: "with content types",
			compression: &ir.Compression{
				Type:             egv1a1.BrotliCompressorType,
				MinContentLength: new(uint32(100)),
				ContentTypes: []egv1a1.StringMatch{
					{Value: "image/svg+xml"},
					{Type: new(egv1a1.StringMatchPrefix), Value: "text/"},
					{Type: new(egv1a1.StringMatchSuffix), Value: "+json"},
					{Type: new(egv1a1.StringMatchRegularExpression), Value: "^font/(ttf|otf)$"},
				},
			},
			expectedName:    "envoy.filters.http.compressor.brotli/d439da1658375a0d",
			expectedExtName: "envoy.compression.brotli.compressor",
			validateProto: func(t *testing.T, c *compressorv3.Compressor) {
				require.NotNil(t, c.ResponseDirectionConfig)
				commonConfig := c.ResponseDirectionConfig.CommonConfig
				assert.Equal(t, uint32(100), commonConfig.MinContentLength.Value)
				assert.Empty(t, commonConfig.ContentType) //nolint:staticcheck // the deprecated field must stay unset
				want := []*matcherv3.StringMatcher{
					{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: "image/svg+xml"}, IgnoreCase: true},
					{MatchPattern: &matcherv3.StringMatcher_Prefix{Prefix: "text/"}, IgnoreCase: true},
					{MatchPattern: &matcherv3.StringMatcher_Suffix{Suffix: "+json"}, IgnoreCase: true},
					{MatchPattern: &matcherv3.StringMatcher_SafeRegex{SafeRegex: &matcherv3.RegexMatcher{Regex: "^font/(ttf|otf)$"}}},
				}
				require.Len(t, commonConfig.ContentTypeMatcher, len(want))
				for i := range want {
					assert.True(t, protobuf.Equal(want[i], commonConfig.ContentTypeMatcher[i]), "matcher %d: %v", i, commonConfig.ContentTypeMatcher[i])
				}
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

func TestCompressorPatchRouteDuplicateType(t *testing.T) {
	irRoute := &ir.HTTPRoute{
		Traffic: &ir.TrafficFeatures{
			Compression: []*ir.Compression{
				{Type: egv1a1.GzipCompressorType},
				{Type: egv1a1.GzipCompressorType},
			},
		},
	}

	err := (&compressor{}).patchRoute(&routev3.Route{}, irRoute, nil)
	require.ErrorContains(t, err, "route already contains filter config: envoy.filters.http.compressor.gzip")
}

func TestCompressorFilterName(t *testing.T) {
	svg := []egv1a1.StringMatch{{Value: "image/svg+xml"}}
	tests := []struct {
		name        string
		compression *ir.Compression
		want        string
	}{
		{"gzip", &ir.Compression{Type: egv1a1.GzipCompressorType}, "envoy.filters.http.compressor.gzip"},
		{"brotli", &ir.Compression{Type: egv1a1.BrotliCompressorType}, "envoy.filters.http.compressor.brotli"},
		{"zstd", &ir.Compression{Type: egv1a1.ZstdCompressorType}, "envoy.filters.http.compressor.zstd"},
		{
			"without content types the filter is shared regardless of the settings",
			&ir.Compression{
				Type:             egv1a1.GzipCompressorType,
				ChooseFirst:      true,
				MinContentLength: new(uint32(100)),
				Gzip:             &egv1a1.GzipCompressor{CompressionLevel: new(uint32(9))},
			},
			"envoy.filters.http.compressor.gzip",
		},
		{
			"content types",
			&ir.Compression{Type: egv1a1.GzipCompressorType, ContentTypes: svg},
			"envoy.filters.http.compressor.gzip/60eaac0969ddae0c",
		},
		{
			"an explicit exact match is the same filter as the default match type",
			&ir.Compression{
				Type:         egv1a1.GzipCompressorType,
				ContentTypes: []egv1a1.StringMatch{{Type: new(egv1a1.StringMatchExact), Value: "image/svg+xml"}},
			},
			"envoy.filters.http.compressor.gzip/60eaac0969ddae0c",
		},
		{
			"compressor library settings are overridden per route and share the filter",
			&ir.Compression{
				Type:         egv1a1.GzipCompressorType,
				ContentTypes: svg,
				Gzip:         &egv1a1.GzipCompressor{CompressionLevel: new(uint32(9))},
			},
			"envoy.filters.http.compressor.gzip/60eaac0969ddae0c",
		},
		{
			"min content length needs its own filter",
			&ir.Compression{Type: egv1a1.GzipCompressorType, ContentTypes: svg, MinContentLength: new(uint32(100))},
			"envoy.filters.http.compressor.gzip/00c0f2b192c7263d",
		},
		{
			"choose first needs its own filter",
			&ir.Compression{Type: egv1a1.GzipCompressorType, ContentTypes: svg, ChooseFirst: true},
			"envoy.filters.http.compressor.gzip/8b7cd03a8dfcc5ba",
		},
		{
			"other content types",
			&ir.Compression{Type: egv1a1.GzipCompressorType, ContentTypes: []egv1a1.StringMatch{{Value: "application/wasm"}}},
			"envoy.filters.http.compressor.gzip/ff22b42521632897",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compressorFilterName(tt.compression)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
