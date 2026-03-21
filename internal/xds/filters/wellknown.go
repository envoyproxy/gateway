// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package filters

import (
	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	setfilterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/set_filter_state/v3"
	grpcstats "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_stats/v3"
	grpcweb "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_web/v3"
	healthcheck "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/health_check/v3"
	luafilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/lua/v3"
	httprouter "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	setfilterstate "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/set_filter_state/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/utils/proto"
)

var GRPCWeb, GRPCStats *hcm.HttpFilter

func init() {
	any, err := proto.ToAnyWithValidation(&grpcweb.GrpcWeb{})
	if err != nil {
		panic(err)
	}
	GRPCWeb = &hcm.HttpFilter{
		Name: wellknown.GRPCWeb,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: any,
		},
	}

	any, err = proto.ToAnyWithValidation(&grpcstats.FilterConfig{
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

func GenerateRouterFilter(enableEnvoyHeaders bool, upstreamAccessLogs []*accesslogv3.AccessLog) (*hcm.HttpFilter, error) {
	routerFilter := &httprouter.Router{
		SuppressEnvoyHeaders: !enableEnvoyHeaders,
		UpstreamLog:          upstreamAccessLogs,
	}
	if len(upstreamAccessLogs) > 0 {
		routerFilter.UpstreamLogOptions = &httprouter.Router_UpstreamAccessLogOptions{
			FlushUpstreamLogOnUpstreamStream: true,
		}
	}
	anyCfg, err := proto.ToAnyWithValidation(routerFilter)
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

const DownstreamProtocolKey = "eg.downstream_protocol"

func GenerateSetDownstreamProtocolFilter() *hcm.HttpFilter {
	anyCfg, _ := proto.ToAnyWithValidation(&setfilterstate.Config{
		OnRequestHeaders: []*setfilterstatev3.FilterStateValue{
			{
				Key: &setfilterstatev3.FilterStateValue_ObjectKey{
					ObjectKey: DownstreamProtocolKey,
				},
				FactoryKey:         "envoy.string",
				SkipIfEmpty:        true,
				SharedWithUpstream: setfilterstatev3.FilterStateValue_ONCE,
				Value: &setfilterstatev3.FilterStateValue_FormatString{
					FormatString: &corev3.SubstitutionFormatString{
						Format: &corev3.SubstitutionFormatString_TextFormatSource{
							TextFormatSource: &corev3.DataSource{
								Specifier: &corev3.DataSource_InlineString{
									InlineString: "%PROTOCOL%",
								},
							},
						},
					},
				},
			},
		},
	})
	return &hcm.HttpFilter{
		Name: "set downstream protocol",
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: anyCfg,
		},
	}
}

// GenerateClearRouteCacheFilter creates a filter that clears the route cache for each request.
// This is a TEMPORARY workaround for Envoy's route caching behavior with filter state changes.
// TODO: Remove this filter entirely after https://github.com/envoyproxy/envoy/issues/44035 is fixed.
func GenerateClearRouteCacheFilter() *hcm.HttpFilter {
	anyCfg, _ := proto.ToAnyWithValidation(&luafilterv3.Lua{
		DefaultSourceCode: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineString{
				InlineString: `function envoy_on_request(handle)
  handle:clearRouteCache()
end
`,
			},
		},
	})
	return &hcm.HttpFilter{
		Name: "clear route cache",
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: anyCfg,
		},
	}
}

func GenerateHealthCheckFilter(checkPath string) (*hcm.HttpFilter, error) {
	anyCfg, err := proto.ToAnyWithValidation(&healthcheck.HealthCheck{
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
