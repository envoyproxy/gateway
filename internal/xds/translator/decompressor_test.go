// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	brotlidecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/decompressor/v3"
	gzipdecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/decompressor/v3"
	zstddecompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/zstd/decompressor/v3"
	decompressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/decompressor/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildDecompressorFilter(t *testing.T) {
	tests := []struct {
		name            string
		decompression   *ir.Decompressor
		expectedName    string
		expectedExtName string
		validateProto   func(*testing.T, *decompressorv3.Decompressor)
	}{
		{
			name: "gzip decompressor",
		decompression: &ir.Decompressor{
			Type: egv1a1.GzipDecompressorType,
		},
		expectedName:    "envoy.filters.http.decompressor.gzip",
		expectedExtName: "envoy.compression.gzip.decompressor",
			validateProto: func(t *testing.T, d *decompressorv3.Decompressor) {
				gzip := &gzipdecompressorv3.Gzip{}
				require.NoError(t, d.DecompressorLibrary.TypedConfig.UnmarshalTo(gzip))
				assert.NotNil(t, d.RequestDirectionConfig)
				assert.NotNil(t, d.ResponseDirectionConfig)
			},
		},
		{
			name: "brotli decompressor",
		decompression: &ir.Decompressor{
			Type: egv1a1.BrotliDecompressorType,
		},
		expectedName:    "envoy.filters.http.decompressor.brotli",
		expectedExtName: "envoy.compression.brotli.decompressor",
			validateProto: func(t *testing.T, d *decompressorv3.Decompressor) {
				brotli := &brotlidecompressorv3.Brotli{}
				require.NoError(t, d.DecompressorLibrary.TypedConfig.UnmarshalTo(brotli))
				assert.NotNil(t, d.RequestDirectionConfig)
				assert.NotNil(t, d.ResponseDirectionConfig)
			},
		},
		{
			name: "zstd decompressor",
		decompression: &ir.Decompressor{
			Type: egv1a1.ZstdDecompressorType,
		},
		expectedName:    "envoy.filters.http.decompressor.zstd",
		expectedExtName: "envoy.compression.zstd.decompressor",
			validateProto: func(t *testing.T, d *decompressorv3.Decompressor) {
				zstd := &zstddecompressorv3.Zstd{}
				require.NoError(t, d.DecompressorLibrary.TypedConfig.UnmarshalTo(zstd))
				assert.NotNil(t, d.RequestDirectionConfig)
				assert.NotNil(t, d.ResponseDirectionConfig)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := buildDecompressorFilter(tt.decompression)
			require.NoError(t, err)
			require.NotNil(t, filter)

			assert.Equal(t, tt.expectedName, filter.Name)
			assert.True(t, filter.Disabled)

			decompressorProto := &decompressorv3.Decompressor{}
			require.NoError(t, filter.GetTypedConfig().UnmarshalTo(decompressorProto))
			assert.Equal(t, tt.expectedExtName, decompressorProto.DecompressorLibrary.Name)

			if tt.validateProto != nil {
				tt.validateProto(t, decompressorProto)
			}
		})
	}
}

func TestDecompressorFilterName(t *testing.T) {
	tests := []struct {
		decompressorType egv1a1.DecompressorType
		want             string
	}{
		{egv1a1.GzipDecompressorType, "envoy.filters.http.decompressor.gzip"},
		{egv1a1.BrotliDecompressorType, "envoy.filters.http.decompressor.brotli"},
		{egv1a1.ZstdDecompressorType, "envoy.filters.http.decompressor.zstd"},
	}

	for _, tt := range tests {
		t.Run(string(tt.decompressorType), func(t *testing.T) {
			assert.Equal(t, tt.want, decompressorFilterName(tt.decompressorType))
		})
	}
}

func TestDecompressorPatchHCM(t *testing.T) {
	d := &decompressor{}

	t.Run("nil hcm returns error", func(t *testing.T) {
		err := d.patchHCM(nil, &ir.HTTPListener{})
		require.Error(t, err)
	})

	t.Run("nil listener returns error", func(t *testing.T) {
		err := d.patchHCM(&hcmv3.HttpConnectionManager{}, nil)
		require.Error(t, err)
	})

	t.Run("no routes with decompression is noop", func(t *testing.T) {
		mgr := &hcmv3.HttpConnectionManager{}
		listener := &ir.HTTPListener{
			Routes: []*ir.HTTPRoute{
				{Traffic: &ir.TrafficFeatures{}},
			},
		}
		err := d.patchHCM(mgr, listener)
		require.NoError(t, err)
		assert.Empty(t, mgr.HttpFilters)
	})

	t.Run("adds decompressor filter for route with decompression", func(t *testing.T) {
		mgr := &hcmv3.HttpConnectionManager{}
		listener := &ir.HTTPListener{
			Routes: []*ir.HTTPRoute{
				{
					Traffic: &ir.TrafficFeatures{
						Decompressor: []*ir.Decompressor{
							{Type: egv1a1.GzipDecompressorType},
						},
					},
				},
			},
		}
		err := d.patchHCM(mgr, listener)
		require.NoError(t, err)
		require.Len(t, mgr.HttpFilters, 1)
		assert.Equal(t, "envoy.filters.http.decompressor.gzip", mgr.HttpFilters[0].Name)
		assert.True(t, mgr.HttpFilters[0].Disabled)
	})

	t.Run("adds multiple decompressor filters for different types", func(t *testing.T) {
		mgr := &hcmv3.HttpConnectionManager{}
		listener := &ir.HTTPListener{
			Routes: []*ir.HTTPRoute{
				{
					Traffic: &ir.TrafficFeatures{
						Decompressor: []*ir.Decompressor{
							{Type: egv1a1.GzipDecompressorType},
							{Type: egv1a1.BrotliDecompressorType},
						},
					},
				},
			},
		}
		err := d.patchHCM(mgr, listener)
		require.NoError(t, err)
		require.Len(t, mgr.HttpFilters, 2)
	})

	t.Run("does not duplicate existing filter", func(t *testing.T) {
		mgr := &hcmv3.HttpConnectionManager{}
		listener := &ir.HTTPListener{
			Routes: []*ir.HTTPRoute{
				{
					Traffic: &ir.TrafficFeatures{
						Decompressor: []*ir.Decompressor{
							{Type: egv1a1.GzipDecompressorType},
						},
					},
				},
				{
					Traffic: &ir.TrafficFeatures{
						Decompressor: []*ir.Decompressor{
							{Type: egv1a1.GzipDecompressorType},
						},
					},
				},
			},
		}
		err := d.patchHCM(mgr, listener)
		require.NoError(t, err)
		require.Len(t, mgr.HttpFilters, 1)
	})
}

func TestDecompressorPatchRoute(t *testing.T) {
	d := &decompressor{}

	t.Run("nil route returns error", func(t *testing.T) {
		err := d.patchRoute(nil, &ir.HTTPRoute{}, nil)
		require.Error(t, err)
	})

	t.Run("nil ir route returns error", func(t *testing.T) {
		err := d.patchRoute(&routev3.Route{}, nil, nil)
		require.Error(t, err)
	})

	t.Run("no traffic config is noop", func(t *testing.T) {
		route := &routev3.Route{}
		irRoute := &ir.HTTPRoute{}
		err := d.patchRoute(route, irRoute, nil)
		require.NoError(t, err)
		assert.Nil(t, route.TypedPerFilterConfig)
	})

	t.Run("empty decompression is noop", func(t *testing.T) {
		route := &routev3.Route{}
		irRoute := &ir.HTTPRoute{
			Traffic: &ir.TrafficFeatures{
				Decompressor: []*ir.Decompressor{},
			},
		}
		err := d.patchRoute(route, irRoute, nil)
		require.NoError(t, err)
		assert.Nil(t, route.TypedPerFilterConfig)
	})

	t.Run("adds per-route filter config for gzip decompression", func(t *testing.T) {
		route := &routev3.Route{}
		irRoute := &ir.HTTPRoute{
			Traffic: &ir.TrafficFeatures{
				Decompressor: []*ir.Decompressor{
					{Type: egv1a1.GzipDecompressorType},
				},
			},
		}
		err := d.patchRoute(route, irRoute, nil)
		require.NoError(t, err)
		require.NotNil(t, route.TypedPerFilterConfig)
		assert.Contains(t, route.TypedPerFilterConfig, "envoy.filters.http.decompressor.gzip")
	})

	t.Run("adds per-route config for multiple decompressor types", func(t *testing.T) {
		route := &routev3.Route{}
		irRoute := &ir.HTTPRoute{
			Traffic: &ir.TrafficFeatures{
				Decompressor: []*ir.Decompressor{
					{Type: egv1a1.GzipDecompressorType},
					{Type: egv1a1.BrotliDecompressorType},
					{Type: egv1a1.ZstdDecompressorType},
				},
			},
		}
		err := d.patchRoute(route, irRoute, nil)
		require.NoError(t, err)
		require.Len(t, route.TypedPerFilterConfig, 3)
		assert.Contains(t, route.TypedPerFilterConfig, "envoy.filters.http.decompressor.gzip")
		assert.Contains(t, route.TypedPerFilterConfig, "envoy.filters.http.decompressor.brotli")
		assert.Contains(t, route.TypedPerFilterConfig, "envoy.filters.http.decompressor.zstd")
	})

	t.Run("returns error if filter config already exists", func(t *testing.T) {
		route := &routev3.Route{}
		irRoute := &ir.HTTPRoute{
			Traffic: &ir.TrafficFeatures{
				Decompressor: []*ir.Decompressor{
					{Type: egv1a1.GzipDecompressorType},
				},
			},
		}
		// Add config the first time
		err := d.patchRoute(route, irRoute, nil)
		require.NoError(t, err)
		// Adding it again should error
		err = d.patchRoute(route, irRoute, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "route already contains filter config")
	})
}
