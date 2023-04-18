// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"encoding/base64"
	"errors"

	xdscore "github.com/cncf/xds/go/xds/core/v3"
	matcher "github.com/cncf/xds/go/xds/type/matcher/v3"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	v31 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	cors "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	grpc_webv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_web/v3"
	"github.com/golang/protobuf/ptypes/wrappers"

	health_check "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/health_check/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	tls_inspector "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	udpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	grpc_json_transcoder "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_json_transcoder/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener.proto#envoy-v3-api-field-config-listener-v3-listener-per-connection-buffer-limit-bytes
	tcpListenerPerConnectionBufferLimitBytes = 32768
)

func buildXdsTCPListener(name, address string, port uint32) *listenerv3.Listener {
	accesslogAny, _ := anypb.New(stdoutFileAccessLog)
	return &listenerv3.Listener{
		Name: name,
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
				Filter:     listenerAccessLogFilter,
			},
		},
		PerConnectionBufferLimitBytes: wrapperspb.UInt32(tcpListenerPerConnectionBufferLimitBytes),
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
	}
}

func (t *Translator) addXdsHTTPFilterChain(xdsListener *listenerv3.Listener, irListener *ir.HTTPListener) error {
	routerAny, err := anypb.New(&router.Router{})
	if err != nil {
		return err
	}

	accesslogAny, err := anypb.New(stdoutFileAccessLog)
	if err != nil {
		return err
	}

	// HTTP filter configuration
	var statPrefix string
	if irListener.TLS != nil {
		statPrefix = "https"
	} else {
		statPrefix = "http"
	}

	mgr := &hcmv3.HttpConnectionManager{
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
			},
		},
		CodecType:  hcmv3.HttpConnectionManager_AUTO,
		StatPrefix: statPrefix,
		RouteSpecifier: &hcmv3.HttpConnectionManager_Rds{
			Rds: &hcmv3.Rds{
				ConfigSource: makeConfigSource(),
				// Configure route name to be found via RDS.
				RouteConfigName: irListener.Name,
			},
		},
		// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for
		UseRemoteAddress: &wrappers.BoolValue{Value: true},
		// Use only router.
		HttpFilters: []*hcmv3.HttpFilter{
			{
				Name:       wellknown.Router,
				ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: routerAny},
			},
		},
	}

	healthCheck := &health_check.HealthCheck{
		PassThroughMode: &wrappers.BoolValue{Value: false},
		Headers: []*v31.HeaderMatcher{
			{
				Name: ":path",
				HeaderMatchSpecifier: &v31.HeaderMatcher_ExactMatch{
					ExactMatch: "/status",
				},
			},
		},
	}

	healthCheckAny, err := anypb.New(healthCheck)
	if err != nil {
		return err
	}

	healthChecFilter := &hcmv3.HttpFilter{
		Name:       wellknown.HealthCheck,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: healthCheckAny},
	}

	mgr.HttpFilters = append([]*hcmv3.HttpFilter{healthChecFilter}, mgr.HttpFilters...)

	for _, route := range irListener.Routes {
		if route.CorsPolicy != nil || irListener.CorsPolicy != nil {
			corsAny, err := anypb.New(&cors.Cors{})

			if err != nil {
				return err
			}

			corsFilter := &hcmv3.HttpFilter{
				Name:       wellknown.CORS,
				ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: corsAny},
			}
			mgr.HttpFilters = append([]*hcmv3.HttpFilter{corsFilter}, mgr.HttpFilters...)
			break
		}
	}

	// TODO: Make this a generic interface for all API Gateway features.
	//       https://github.com/envoyproxy/gateway/issues/882
	t.patchHCMWithRateLimit(mgr, irListener)

	// Add the jwt authn filter, if needed.
	if err := patchHCMWithJwtAuthnFilter(mgr, irListener); err != nil {
		return err
	}

	// add GrpcJSONTranscoderFilter to httpFilters
	if irListener.GrpcJSONTranscoderFilters != nil || len(irListener.GrpcJSONTranscoderFilters) > 0 {
		for _, filter := range irListener.GrpcJSONTranscoderFilters {
			bytt, err := base64.StdEncoding.DecodeString(filter.ProtoDescriptorBin)

			if err != nil {
				return err
			}

			grpcJSONTranscoderAny, err := anypb.New(&grpc_json_transcoder.GrpcJsonTranscoder{
				AutoMapping:       filter.AutoMapping,
				ConvertGrpcStatus: true,
				Services:          filter.Services,
				PrintOptions: &grpc_json_transcoder.GrpcJsonTranscoder_PrintOptions{
					AddWhitespace:              filter.PrintOptions.AddWhitespace,
					AlwaysPrintPrimitiveFields: filter.PrintOptions.AlwaysPrintPrimitiveFields,
					AlwaysPrintEnumsAsInts:     filter.PrintOptions.AlwaysPrintEnumsAsInts,
					PreserveProtoFieldNames:    filter.PrintOptions.PreserveProtoFieldNames,
				},
				DescriptorSet: &grpc_json_transcoder.GrpcJsonTranscoder_ProtoDescriptorBin{
					ProtoDescriptorBin: bytt,
				},
			})

			if err != nil {
				return err
			}

			grpcJSONTranscoderFilter := &hcmv3.HttpFilter{
				Name:       wellknown.GRPCJSONTranscoder,
				ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: grpcJSONTranscoderAny},
			}
			mgr.HttpFilters = append([]*hcmv3.HttpFilter{grpcJSONTranscoderFilter}, mgr.HttpFilters...)
		}
	}

	if irListener.IsHTTP2 {
		// Set codec to HTTP2
		mgr.CodecType = hcmv3.HttpConnectionManager_HTTP2

		// Enable grpc-web filter for HTTP2
		grpcWebAny, err := anypb.New(&grpc_webv3.GrpcWeb{})
		if err != nil {
			return err
		}

		grpcWebFilter := &hcmv3.HttpFilter{
			Name:       wellknown.GRPCWeb,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: grpcWebAny},
		}
		// Ensure router is the last filter
		mgr.HttpFilters = append([]*hcmv3.HttpFilter{grpcWebFilter}, mgr.HttpFilters...)
	} else {
		// Allow websocket upgrades for HTTP 1.1
		// Reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Protocol_upgrade_mechanism
		mgr.UpgradeConfigs = []*hcmv3.HttpConnectionManager_UpgradeConfig{
			{
				UpgradeType: "websocket",
			},
		}
	}

	mgrAny, err := anypb.New(mgr)
	if err != nil {
		return err
	}

	filterChain := &listenerv3.FilterChain{
		Filters: []*listenerv3.Filter{{
			Name: wellknown.HTTPConnectionManager,
			ConfigType: &listenerv3.Filter_TypedConfig{
				TypedConfig: mgrAny,
			},
		}},
	}

	if irListener.TLS != nil {
		tSocket, err := buildXdsDownstreamTLSSocket(irListener.TLS)
		if err != nil {
			return err
		}
		filterChain.TransportSocket = tSocket
		if err := addServerNamesMatch(xdsListener, filterChain, irListener.Hostnames); err != nil {
			return err
		}

		xdsListener.FilterChains = append(xdsListener.FilterChains, filterChain)
	} else {
		// Add the HTTP filter chain as the default filter chain
		// Make sure one does not exist
		if xdsListener.DefaultFilterChain != nil {
			return errors.New("default filter chain already exists")
		}
		xdsListener.DefaultFilterChain = filterChain
	}

	return nil
}

func addServerNamesMatch(xdsListener *listenerv3.Listener, filterChain *listenerv3.FilterChain, hostnames []string) error {
	// Dont add a filter chain match if the hostname is a wildcard character.
	if len(hostnames) > 0 && hostnames[0] != "*" {
		filterChain.FilterChainMatch = &listenerv3.FilterChainMatch{
			ServerNames: hostnames,
		}

		if err := addXdsTLSInspectorFilter(xdsListener); err != nil {
			return err
		}
	}

	return nil
}

// findXdsHTTPRouteConfigName finds the name of the route config associated with the
// http connection manager within the default filter chain and returns an empty string if
// not found.
func findXdsHTTPRouteConfigName(xdsListener *listenerv3.Listener) string {
	if xdsListener == nil || xdsListener.DefaultFilterChain == nil || xdsListener.DefaultFilterChain.Filters == nil {
		return ""
	}

	for _, filter := range xdsListener.DefaultFilterChain.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			m := new(hcmv3.HttpConnectionManager)
			if err := filter.GetTypedConfig().UnmarshalTo(m); err != nil {
				return ""
			}
			rds := m.GetRds()
			if rds == nil {
				return ""
			}
			return rds.GetRouteConfigName()
		}
	}
	return ""
}

func addXdsTCPFilterChain(xdsListener *listenerv3.Listener, irListener *ir.TCPListener, clusterName string) error {
	if irListener == nil {
		return errors.New("tcp listener is nil")
	}

	statPrefix := "tcp"
	if irListener.TLS != nil {
		statPrefix = "passthrough"
	}

	accesslogAny, err := anypb.New(stdoutFileAccessLog)
	if err != nil {
		return err
	}

	mgr := &tcpv3.TcpProxy{
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
			},
		},
		StatPrefix: statPrefix,
		ClusterSpecifier: &tcpv3.TcpProxy_Cluster{
			Cluster: clusterName,
		},
	}
	mgrAny, err := anypb.New(mgr)
	if err != nil {
		return err
	}

	filterChain := &listenerv3.FilterChain{
		Filters: []*listenerv3.Filter{{
			Name: wellknown.TCPProxy,
			ConfigType: &listenerv3.Filter_TypedConfig{
				TypedConfig: mgrAny,
			},
		}},
	}

	if irListener.TLS != nil {
		if err := addServerNamesMatch(xdsListener, filterChain, irListener.TLS.SNIs); err != nil {
			return err
		}
	}

	xdsListener.FilterChains = append(xdsListener.FilterChains, filterChain)

	return nil
}

// addXdsTLSInspectorFilter adds a Tls Inspector filter if it does not yet exist.
func addXdsTLSInspectorFilter(xdsListener *listenerv3.Listener) error {
	// Return early if it exists
	for _, filter := range xdsListener.ListenerFilters {
		if filter.Name == wellknown.TlsInspector {
			return nil
		}
	}

	tlsInspector := &tls_inspector.TlsInspector{}
	tlsInspectorAny, err := anypb.New(tlsInspector)
	if err != nil {
		return err
	}

	filter := &listenerv3.ListenerFilter{
		Name: wellknown.TlsInspector,
		ConfigType: &listenerv3.ListenerFilter_TypedConfig{
			TypedConfig: tlsInspectorAny,
		},
	}

	xdsListener.ListenerFilters = append(xdsListener.ListenerFilters, filter)

	return nil
}

func buildXdsDownstreamTLSSocket(tlsConfigs []*ir.TLSListenerConfig) (*corev3.TransportSocket, error) {
	tlsCtx := &tlsv3.DownstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{},
		},
	}

	for _, tlsConfig := range tlsConfigs {
		tlsCtx.CommonTlsContext.TlsCertificateSdsSecretConfigs = append(
			tlsCtx.CommonTlsContext.TlsCertificateSdsSecretConfigs,
			&tlsv3.SdsSecretConfig{
				Name:      tlsConfig.Name,
				SdsConfig: makeConfigSource(),
			})
	}

	tlsCtxAny, err := anypb.New(tlsCtx)
	if err != nil {
		return nil, err
	}

	return &corev3.TransportSocket{
		Name: wellknown.TransportSocketTls,
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: tlsCtxAny,
		},
	}, nil
}

func buildXdsDownstreamTLSSecret(tlsConfig *ir.TLSListenerConfig) *tlsv3.Secret {
	// Build the tls secret
	return &tlsv3.Secret{
		Name: tlsConfig.Name,
		Type: &tlsv3.Secret_TlsCertificate{
			TlsCertificate: &tlsv3.TlsCertificate{
				CertificateChain: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{InlineBytes: tlsConfig.ServerCertificate},
				},
				PrivateKey: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{InlineBytes: tlsConfig.PrivateKey},
				},
			},
		},
	}
}

func buildXdsUDPListener(clusterName string, udpListener *ir.UDPListener) (*listenerv3.Listener, error) {
	if udpListener == nil {
		return nil, errors.New("udp listener is nil")
	}

	statPrefix := "service"

	route := &udpv3.Route{
		Cluster: clusterName,
	}
	routeAny, err := anypb.New(route)
	if err != nil {
		return nil, err
	}
	accesslogAny, _ := anypb.New(stdoutFileAccessLog)
	udpProxy := &udpv3.UdpProxyConfig{
		StatPrefix: statPrefix,
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
			},
		},
		RouteSpecifier: &udpv3.UdpProxyConfig_Matcher{
			Matcher: &matcher.Matcher{
				OnNoMatch: &matcher.Matcher_OnMatch{
					OnMatch: &matcher.Matcher_OnMatch_Action{
						Action: &xdscore.TypedExtensionConfig{
							Name:        "route",
							TypedConfig: routeAny,
						},
					},
				},
			},
		},
	}
	udpProxyAny, err := anypb.New(udpProxy)
	if err != nil {
		return nil, err
	}

	xdsListener := &listenerv3.Listener{
		Name: udpListener.Name,
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
			},
		},
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_UDP,
					Address:  udpListener.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: udpListener.Port,
					},
				},
			},
		},
		ListenerFilters: []*listenerv3.ListenerFilter{{
			Name: "envoy.filters.udp_listener.udp_proxy",
			ConfigType: &listenerv3.ListenerFilter_TypedConfig{
				TypedConfig: udpProxyAny,
			},
		}},
	}

	return xdsListener, nil
}

// Point to xds cluster.
func makeConfigSource() *corev3.ConfigSource {
	source := &corev3.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &corev3.ConfigSource_Ads{
		Ads: &corev3.AggregatedConfigSource{},
	}
	return source
}
