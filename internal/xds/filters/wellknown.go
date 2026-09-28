// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package filters

import (
	accesslogv3 "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	setfilterstatecommonv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/set_filter_state/v3"
	grpcstats "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_stats/v3"
	grpcweb "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_web/v3"
	healthcheck "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/health_check/v3"
	httprouter "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	setfilterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/set_filter_state/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/utils/proto"
)

const (
	// DownstreamProtocolFilterName is the name of the HTTP filter that stores the downstream HTTP protocol
	// in the filter state under DownstreamProtocolKey.
	DownstreamProtocolFilterName = "envoy.filters.http.set_filter_state/downstream_protocol"
	// DownstreamProtocolKey is the filter state key holding the downstream HTTP protocol, e.g. "HTTP/2".
	DownstreamProtocolKey = "envoy-gateway.downstream_protocol"
)

var GRPCWeb, GRPCStats, DownstreamProtocol *hcm.HttpFilter

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

	any, err = proto.ToAnyWithValidation(&setfilterstatev3.Config{
		OnRequestHeaders: []*setfilterstatecommonv3.FilterStateValue{
			{
				Key:        &setfilterstatecommonv3.FilterStateValue_ObjectKey{ObjectKey: DownstreamProtocolKey},
				FactoryKey: "envoy.string",
				Value: &setfilterstatecommonv3.FilterStateValue_FormatString{
					FormatString: &corev3.SubstitutionFormatString{
						Format: &corev3.SubstitutionFormatString_TextFormatSource{
							TextFormatSource: &corev3.DataSource{
								Specifier: &corev3.DataSource_InlineString{InlineString: "%PROTOCOL%"},
							},
						},
					},
				},
			},
		},
		// The HCM resolves the route before running the HTTP filters, so the route cache must be
		// cleared for routes that match on this filter state to be selected.
		ClearRouteCache: true,
	})
	if err != nil {
		panic(err)
	}
	DownstreamProtocol = &hcm.HttpFilter{
		Name: DownstreamProtocolFilterName,
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
