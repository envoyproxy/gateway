// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	xdscore "github.com/cncf/xds/go/xds/core/v3"
	matcher "github.com/cncf/xds/go/xds/type/matcher/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tls_inspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	udpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/protocov"
	xdsfilters "github.com/envoyproxy/gateway/internal/xds/filters"
)

const (
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/listener/v3/listener.proto#envoy-v3-api-field-config-listener-v3-listener-per-connection-buffer-limit-bytes
	tcpListenerPerConnectionBufferLimitBytes = 32768
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-http2protocoloptions-max-concurrent-streams
	http2MaxConcurrentStreamsLimit = 100
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-http2protocoloptions-initial-stream-window-size
	http2InitialStreamWindowSize = 65536 // 64 KiB
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/protocol.proto#envoy-v3-api-field-config-core-v3-http2protocoloptions-initial-connection-window-size
	http2InitialConnectionWindowSize = 1048576 // 1 MiB
)

func http2ProtocolOptions() *corev3.Http2ProtocolOptions {
	return &corev3.Http2ProtocolOptions{
		MaxConcurrentStreams: &wrappers.UInt32Value{
			Value: http2MaxConcurrentStreamsLimit,
		},
		InitialStreamWindowSize: &wrappers.UInt32Value{
			Value: http2InitialStreamWindowSize,
		},
		InitialConnectionWindowSize: &wrappers.UInt32Value{
			Value: http2InitialConnectionWindowSize,
		},
	}
}

func buildXdsTCPListener(name, address string, port uint32, accesslog *ir.AccessLog) *listenerv3.Listener {
	al := buildXdsAccessLog(accesslog, true)
	return &listenerv3.Listener{
		Name:                          name,
		AccessLog:                     al,
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

func (t *Translator) addXdsHTTPFilterChain(xdsListener *listenerv3.Listener, irListener *ir.HTTPListener,
	accesslog *ir.AccessLog, tracing *ir.Tracing) error {
	al := buildXdsAccessLog(accesslog, false)

	hcmTracing, err := buildHCMTracing(tracing)
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
		AccessLog:  al,
		CodecType:  hcmv3.HttpConnectionManager_AUTO,
		StatPrefix: statPrefix,
		RouteSpecifier: &hcmv3.HttpConnectionManager_Rds{
			Rds: &hcmv3.Rds{
				ConfigSource: makeConfigSource(),
				// Configure route name to be found via RDS.
				RouteConfigName: irListener.Name,
			},
		},
		// Add HTTP2 protocol options
		// Set it by default to also support HTTP1.1 to HTTP2 Upgrades
		Http2ProtocolOptions: http2ProtocolOptions(),
		// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for
		UseRemoteAddress: &wrappers.BoolValue{Value: true},
		// normalize paths according to RFC 3986
		NormalizePath: &wrapperspb.BoolValue{Value: true},
		// merge adjacent slashes in the path
		MergeSlashes:                 true,
		PathWithEscapedSlashesAction: hcmv3.HttpConnectionManager_UNESCAPE_AND_REDIRECT,
		CommonHttpProtocolOptions: &corev3.HttpProtocolOptions{
			HeadersWithUnderscoresAction: corev3.HttpProtocolOptions_REJECT_REQUEST,
		},
		Tracing: hcmTracing,
	}

	if irListener.IsHTTP2 {
		mgr.HttpFilters = append(mgr.HttpFilters, xdsfilters.GRPCWeb)
		// always enable grpc stats filter
		mgr.HttpFilters = append(mgr.HttpFilters, xdsfilters.GRPCStats)
	} else {
		// Allow websocket upgrades for HTTP 1.1
		// Reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Protocol_upgrade_mechanism
		mgr.UpgradeConfigs = []*hcmv3.HttpConnectionManager_UpgradeConfig{
			{
				UpgradeType: "websocket",
			},
		}
	}

	// TODO: Make this a generic interface for all API Gateway features.
	//       https://github.com/envoyproxy/gateway/issues/882
	t.patchHCMWithRateLimit(mgr, irListener)

	// Add the jwt authn filter, if needed.
	if err := patchHCMWithJwtAuthnFilter(mgr, irListener); err != nil {
		return err
	}

	// Make sure the router filter is the last one.
	mgr.HttpFilters = append(mgr.HttpFilters, xdsfilters.HTTPRouter)
	mgrAny, err := protocov.ToAnyWithError(mgr)
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

func addXdsTCPFilterChain(xdsListener *listenerv3.Listener, irListener *ir.TCPListener, clusterName string, accesslog *ir.AccessLog) error {
	if irListener == nil {
		return errors.New("tcp listener is nil")
	}

	isTLSPassthrough := irListener.TLS != nil && irListener.TLS.Passthrough != nil
	isTLSTerminate := irListener.TLS != nil && irListener.TLS.Terminate != nil
	statPrefix := "tcp"
	if isTLSPassthrough {
		statPrefix = "passthrough"
	}

	if isTLSTerminate {
		statPrefix = "terminate"
	}

	mgr := &tcpv3.TcpProxy{
		AccessLog:  buildXdsAccessLog(accesslog, false),
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

	if isTLSPassthrough {
		if err := addServerNamesMatch(xdsListener, filterChain, irListener.TLS.Passthrough.SNIs); err != nil {
			return err
		}
	}

	if isTLSTerminate {
		tSocket, err := buildXdsDownstreamTLSSocket(irListener.TLS.Terminate)
		if err != nil {
			return err
		}
		filterChain.TransportSocket = tSocket
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

	tlsInspector := &tls_inspectorv3.TlsInspector{}
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
			AlpnProtocols:                  []string{"h2", "http/1.1"},
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

func buildXdsUDPListener(clusterName string, udpListener *ir.UDPListener, accesslog *ir.AccessLog) (*listenerv3.Listener, error) {
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

	udpProxy := &udpv3.UdpProxyConfig{
		StatPrefix: statPrefix,
		AccessLog:  buildXdsAccessLog(accesslog, false),
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
		Name:      udpListener.Name,
		AccessLog: buildXdsAccessLog(accesslog, true),
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
