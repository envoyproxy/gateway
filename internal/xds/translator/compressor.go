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
	compressorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/filters"
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
		filter     *hcmv3.HttpFilter
		filterName string
		err        error
	)

	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.Compression != nil {
			for _, irComp := range route.Traffic.Compression {
				if filterName, err = compressorFilterName(irComp); err != nil {
					return err
				}
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

// compressorFilterName returns the name of the listener-level compressor filter for the compression.
func compressorFilterName(compression *ir.Compression) (string, error) {
	name := fmt.Sprintf("%s.%s", egv1a1.EnvoyFilterCompressor.String(), strings.ToLower(string(compression.Type)))
	if len(compression.ContentTypes) == 0 {
		return name, nil
	}

	// CompressorPerRoute can't override the content types, so every distinct filter config with
	// content types gets a filter of its own, named after a hash of the config. Routes without
	// content types keep sharing a single filter per compressor type.
	compressorProto, err := buildCompressorProto(compression)
	if err != nil {
		return "", err
	}
	hash, err := protoHash(compressorProto)
	if err != nil {
		return "", err
	}
	return name + "/" + hash, nil
}

// buildCompressorProto builds the listener-level config of the compressor filter.
func buildCompressorProto(compression *ir.Compression) (*compressorv3.Compressor, error) {
	// The filter is shared by multiple routes, so it always uses the default compressor
	// library settings. Routes with custom settings override the compressor library.
	compressorLibrary, err := filters.CompressorLibrary(compression.Type, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	compressorProto := &compressorv3.Compressor{
		CompressorLibrary: compressorLibrary,
		ChooseFirst:       compression.ChooseFirst,
	}

	if compression.MinContentLength != nil || len(compression.ContentTypes) > 0 {
		commonConfig := &compressorv3.Compressor_CommonDirectionConfig{
			ContentTypeMatcher: filters.CompressorContentTypeMatchers(compression.ContentTypes),
		}
		if compression.MinContentLength != nil {
			commonConfig.MinContentLength = wrapperspb.UInt32(*compression.MinContentLength)
		}
		compressorProto.ResponseDirectionConfig = &compressorv3.Compressor_ResponseDirectionConfig{
			CommonConfig: commonConfig,
		}
	}

	return compressorProto, nil
}

// buildCompressorFilter builds a compressor filter with the provided compressionType.
func buildCompressorFilter(compression *ir.Compression) (*hcmv3.HttpFilter, error) {
	filterName, err := compressorFilterName(compression)
	if err != nil {
		return nil, err
	}

	compressorProto, err := buildCompressorProto(compression)
	if err != nil {
		return nil, err
	}

	compressorAny, err := proto.ToAnyWithValidation(compressorProto)
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
		filterName    string
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
		if filterName, err = compressorFilterName(irComp); err != nil {
			return err
		}
		if _, ok := perFilterCfg[filterName]; ok {
			// This should not happen since this is the only place where the filter
			// config is added in a route.
			return fmt.Errorf("route already contains filter config: %s, %+v",
				filterName, route)
		}

		var compressorLibrary *corev3.TypedExtensionConfig
		if irComp.Gzip != nil || irComp.Brotli != nil || irComp.Zstd != nil {
			if compressorLibrary, err = filters.CompressorLibrary(irComp.Type, irComp.Gzip, irComp.Brotli, irComp.Zstd); err != nil {
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
