// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package filters

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	brotliv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/brotli/compressor/v3"
	gzipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	compressor "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	grpcstats "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_stats/v3"
	grpcweb "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_web/v3"
	healthcheck "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/health_check/v3"
	httprouter "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/protocov"
)

var GRPCWeb, GRPCStats *hcm.HttpFilter

func init() {
	any, err := protocov.ToAnyWithValidation(&grpcweb.GrpcWeb{})
	if err != nil {
		panic(err)
	}
	GRPCWeb = &hcm.HttpFilter{
		Name: wellknown.GRPCWeb,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: any,
		},
	}

	any, err = protocov.ToAnyWithValidation(&grpcstats.FilterConfig{
		EmitFilterState: true,
		PerMethodStatSpecifier: &grpcstats.FilterConfig_StatsForAllMethods{
			StatsForAllMethods: &wrapperspb.BoolValue{Value: true},
		},
	})
	if err != nil {
		panic(err)
	}
	GRPCStats = &hcm.HttpFilter{
		Name: wellknown.HTTPGRPCStats,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: any,
		},
	}
}

func GenerateRouterFilter(enableEnvoyHeaders bool) (*hcm.HttpFilter, error) {
	anyCfg, err := protocov.ToAnyWithValidation(&httprouter.Router{
		SuppressEnvoyHeaders: !enableEnvoyHeaders,
	})
	if err != nil {
		return nil, err
	}
	return &hcm.HttpFilter{
		Name: wellknown.Router,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: anyCfg,
		},
	}, nil
}

func GenerateHealthCheckFilter(checkPath string) (*hcm.HttpFilter, error) {
	anyCfg, err := protocov.ToAnyWithValidation(&healthcheck.HealthCheck{
		PassThroughMode: &wrapperspb.BoolValue{Value: false},
		Headers: []*routev3.HeaderMatcher{
			{
				Name: ":path",
				HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
					StringMatch: &matcherv3.StringMatcher{
						MatchPattern: &matcherv3.StringMatcher_Exact{
							Exact: checkPath,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &hcm.HttpFilter{
		Name: wellknown.HealthCheck,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: anyCfg,
		},
	}, nil
}

func GenerateCompressorFilter(compressorType egv1a1.CompressorType) (*hcm.HttpFilter, error) {
	var libCfg *anypb.Any
	var err error
	switch compressorType {
	case egv1a1.GzipCompressorType:
		libCfg, err = protocov.ToAnyWithValidation(&gzipv3.Gzip{})
	case egv1a1.BrotliCompressorType:
		libCfg, err = protocov.ToAnyWithValidation(&brotliv3.Brotli{})
	// TODO: support zstd
	default:
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	anyCfg, err := protocov.ToAnyWithValidation(&compressor.Compressor{
		CompressorLibrary: &corev3.TypedExtensionConfig{
			Name:        "text_optimized",
			TypedConfig: libCfg,
		},
	})
	if err != nil {
		return nil, err
	}

	return &hcm.HttpFilter{
		Name: "envoy.filters.http.compression",
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: anyCfg,
		},
	}, nil
}

func GenerateCompressorPerRouteFilter(disabled bool) (*compressor.CompressorPerRoute, error) {
	if disabled {
		return &compressor.CompressorPerRoute{
			Override: &compressor.CompressorPerRoute_Disabled{
				Disabled: true,
			},
		}, nil
	}
	return &compressor.CompressorPerRoute{
		Override: &compressor.CompressorPerRoute_Overrides{
			Overrides: &compressor.CompressorOverrides{
				ResponseDirectionConfig: &compressor.ResponseDirectionOverrides{},
			},
		},
	}, nil
}
