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
	connection_limitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/connection_limit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	udpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	preservecasev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/header_formatters/preserve_case/v3"
	customheaderv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/original_ip_detection/custom_header/v3"
	quicv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/quic/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/ptr"

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
	// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/network/connection_limit/v3/connection_limit.proto
	networkConnectionLimit = "envoy.filters.network.connection_limit"
)

func http1ProtocolOptions(opts *ir.HTTP1Settings) *corev3.Http1ProtocolOptions {
	if opts == nil {
		return nil
	}
	if !opts.EnableTrailers && !opts.PreserveHeaderCase && opts.HTTP10 == nil {
		return nil
	}
	// If PreserveHeaderCase is true and EnableTrailers is false then setting the EnableTrailers field to false
	// is simply keeping it at its default value of "disabled".
	r := &corev3.Http1ProtocolOptions{
		EnableTrailers: opts.EnableTrailers,
	}
	if opts.PreserveHeaderCase {
		preservecaseAny, _ := anypb.New(&preservecasev3.PreserveCaseFormatterConfig{})
		r.HeaderKeyFormat = &corev3.Http1ProtocolOptions_HeaderKeyFormat{
			HeaderFormat: &corev3.Http1ProtocolOptions_HeaderKeyFormat_StatefulFormatter{
				StatefulFormatter: &corev3.TypedExtensionConfig{
					Name:        "preserve_case",
					TypedConfig: preservecaseAny,
				},
			},
		}
	}
	if opts.HTTP10 != nil {
		r.AcceptHttp_10 = true
		r.DefaultHostForHttp_10 = ptr.Deref(opts.HTTP10.DefaultHost, "")
	}
	return r
}

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

func xffNumTrustedHops(clientIPDetection *ir.ClientIPDetectionSettings) uint32 {
	if clientIPDetection != nil && clientIPDetection.XForwardedFor != nil && clientIPDetection.XForwardedFor.NumTrustedHops != nil {
		return *clientIPDetection.XForwardedFor.NumTrustedHops
	}
	return 0
}

func originalIPDetectionExtensions(clientIPDetection *ir.ClientIPDetectionSettings) []*corev3.TypedExtensionConfig {
	// Return early if settings are nil
	if clientIPDetection == nil {
		return nil
	}

	var extensionConfig []*corev3.TypedExtensionConfig

	// Custom header extension
	if clientIPDetection.CustomHeader != nil {
		var rejectWithStatus *typev3.HttpStatus
		if ptr.Deref(clientIPDetection.CustomHeader.FailClosed, false) {
			rejectWithStatus = &typev3.HttpStatus{Code: typev3.StatusCode_Forbidden}
		}

		customHeaderConfigAny, _ := anypb.New(&customheaderv3.CustomHeaderConfig{
			HeaderName:       clientIPDetection.CustomHeader.Name,
			RejectWithStatus: rejectWithStatus,

			AllowExtensionToSetAddressAsTrusted: true,
		})

		extensionConfig = append(extensionConfig, &corev3.TypedExtensionConfig{
			Name:        "envoy.extensions.http.original_ip_detection.custom_header",
			TypedConfig: customHeaderConfigAny,
		})
	}

	return extensionConfig
}

// buildXdsTCPListener creates a xds Listener resource
// TODO: Improve function parameters
func buildXdsTCPListener(name, address string, port uint32, keepalive *ir.TCPKeepalive, accesslog *ir.AccessLog) *listenerv3.Listener {
	socketOptions := buildTCPSocketOptions(keepalive)
	al := buildXdsAccessLog(accesslog, true)
	return &listenerv3.Listener{
		Name:                          name,
		AccessLog:                     al,
		SocketOptions:                 socketOptions,
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
		// Remove /healthcheck/fail from endpoints that trigger a drain of listeners for better control
		// over the drain process while still allowing the healthcheck to be failed during pod shutdown.
		DrainType: listenerv3.Listener_MODIFY_ONLY,
	}
}

// buildXdsQuicListener creates a xds Listener resource for quic
func buildXdsQuicListener(name, address string, port uint32, accesslog *ir.AccessLog) *listenerv3.Listener {
	xdsListener := &listenerv3.Listener{
		Name:      name + "-quic",
		AccessLog: buildXdsAccessLog(accesslog, true),
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_UDP,
					Address:  address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		UdpListenerConfig: &listenerv3.UdpListenerConfig{
			DownstreamSocketConfig: &corev3.UdpSocketConfig{},
			QuicOptions:            &listenerv3.QuicProtocolOptions{},
		},
		// Remove /healthcheck/fail from endpoints that trigger a drain of listeners for better control
		// over the drain process while still allowing the healthcheck to be failed during pod shutdown.
		DrainType: listenerv3.Listener_MODIFY_ONLY,
	}

	return xdsListener
}

func (t *Translator) addXdsHTTPFilterChain(xdsListener *listenerv3.Listener, irListener *ir.HTTPListener,
	accesslog *ir.AccessLog, tracing *ir.Tracing, http3Listener bool, connection *ir.Connection) error {
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

	// Client IP detection
	var useRemoteAddress = true
	var originalIPDetectionExtensions = originalIPDetectionExtensions(irListener.ClientIPDetection)
	if originalIPDetectionExtensions != nil {
		useRemoteAddress = false
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
		HttpProtocolOptions: http1ProtocolOptions(irListener.HTTP1),
		// Hide the Envoy proxy in the Server header by default
		ServerHeaderTransformation: hcmv3.HttpConnectionManager_PASS_THROUGH,
		// Add HTTP2 protocol options
		// Set it by default to also support HTTP1.1 to HTTP2 Upgrades
		Http2ProtocolOptions: http2ProtocolOptions(),
		// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for
		UseRemoteAddress:              &wrappers.BoolValue{Value: useRemoteAddress},
		XffNumTrustedHops:             xffNumTrustedHops(irListener.ClientIPDetection),
		OriginalIpDetectionExtensions: originalIPDetectionExtensions,
		// normalize paths according to RFC 3986
		NormalizePath:                &wrapperspb.BoolValue{Value: true},
		MergeSlashes:                 irListener.Path.MergeSlashes,
		PathWithEscapedSlashesAction: translateEscapePath(irListener.Path.EscapedSlashesAction),
		CommonHttpProtocolOptions: &corev3.HttpProtocolOptions{
			HeadersWithUnderscoresAction: buildHeadersWithUnderscoresAction(irListener.Headers),
		},
		Tracing: hcmTracing,
	}

	if irListener.Timeout != nil && irListener.Timeout.HTTP != nil {
		if irListener.Timeout.HTTP.RequestReceivedTimeout != nil {
			mgr.RequestTimeout = durationpb.New(irListener.Timeout.HTTP.RequestReceivedTimeout.Duration)
		}

		if irListener.Timeout.HTTP.IdleTimeout != nil {
			mgr.CommonHttpProtocolOptions.IdleTimeout = durationpb.New(irListener.Timeout.HTTP.IdleTimeout.Duration)
		}
	}

	// Add the proxy protocol filter if needed
	patchProxyProtocolFilter(xdsListener, irListener)

	if irListener.IsHTTP2 {
		mgr.HttpFilters = append(mgr.HttpFilters, xdsfilters.GRPCWeb)
		// always enable grpc stats filter
		mgr.HttpFilters = append(mgr.HttpFilters, xdsfilters.GRPCStats)
	}

	if http3Listener {
		mgr.CodecType = hcmv3.HttpConnectionManager_HTTP3
		mgr.Http3ProtocolOptions = &corev3.Http3ProtocolOptions{}
	}
	// Add HTTP filters to the HCM, the filters have already been sorted in the
	// correct order in the patchHCMWithFilters function.
	if err := t.patchHCMWithFilters(mgr, irListener); err != nil {
		return err
	}

	var filters []*listenerv3.Filter

	if connection != nil && connection.Limit != nil {
		cl := buildConnectionLimitFilter(statPrefix, connection)
		if clf, err := toNetworkFilter(networkConnectionLimit, cl); err == nil {
			filters = append(filters, clf)
		} else {
			return err
		}
	}

	if mgrf, err := toNetworkFilter(wellknown.HTTPConnectionManager, mgr); err == nil {
		filters = append(filters, mgrf)
	} else {
		return err
	}

	filterChain := &listenerv3.FilterChain{
		Filters: filters,
	}

	if irListener.TLS != nil {
		var tSocket *corev3.TransportSocket
		if http3Listener {
			tSocket, err = buildDownstreamQUICTransportSocket(irListener.TLS)
			if err != nil {
				return err
			}
		} else {
			tSocket, err = buildXdsDownstreamTLSSocket(irListener.TLS)
			if err != nil {
				return err
			}
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

func addXdsTCPFilterChain(xdsListener *listenerv3.Listener, irListener *ir.TCPListener, clusterName string, accesslog *ir.AccessLog, connection *ir.Connection) error {
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
		HashPolicy: buildTCPProxyHashPolicy(irListener.LoadBalancer),
	}

	var filters []*listenerv3.Filter

	if connection != nil && connection.Limit != nil {
		cl := buildConnectionLimitFilter(statPrefix, connection)
		if clf, err := toNetworkFilter(networkConnectionLimit, cl); err == nil {
			filters = append(filters, clf)
		} else {
			return err
		}
	}

	if mgrf, err := toNetworkFilter(wellknown.TCPProxy, mgr); err == nil {
		filters = append(filters, mgrf)
	} else {
		return err
	}

	filterChain := &listenerv3.FilterChain{
		Filters: filters,
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

func buildConnectionLimitFilter(statPrefix string, connection *ir.Connection) *connection_limitv3.ConnectionLimit {
	cl := &connection_limitv3.ConnectionLimit{
		StatPrefix:     statPrefix,
		MaxConnections: wrapperspb.UInt64(*connection.Limit.Value),
	}

	if connection.Limit.CloseDelay != nil {
		cl.Delay = durationpb.New(connection.Limit.CloseDelay.Duration)
	}
	return cl
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

func buildDownstreamQUICTransportSocket(tlsConfig *ir.TLSConfig) (*corev3.TransportSocket, error) {
	tlsCtx := &quicv3.QuicDownstreamTransport{
		DownstreamTlsContext: &tlsv3.DownstreamTlsContext{
			CommonTlsContext: &tlsv3.CommonTlsContext{
				TlsParams:     buildTLSParams(tlsConfig),
				AlpnProtocols: []string{"h3"},
			},
		},
	}

	for _, tlsConfig := range tlsConfig.Certificates {
		tlsCtx.DownstreamTlsContext.CommonTlsContext.TlsCertificateSdsSecretConfigs = append(
			tlsCtx.DownstreamTlsContext.CommonTlsContext.TlsCertificateSdsSecretConfigs,
			&tlsv3.SdsSecretConfig{
				Name:      tlsConfig.Name,
				SdsConfig: makeConfigSource(),
			})
	}

	if tlsConfig.CACertificate != nil {
		tlsCtx.DownstreamTlsContext.RequireClientCertificate = &wrappers.BoolValue{Value: true}
		tlsCtx.DownstreamTlsContext.CommonTlsContext.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
			ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
				Name:      tlsConfig.CACertificate.Name,
				SdsConfig: makeConfigSource(),
			},
		}
	}

	tlsCtxAny, err := anypb.New(tlsCtx)
	if err != nil {
		return nil, err
	}
	return &corev3.TransportSocket{
		Name: wellknown.TransportSocketQuic,
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: tlsCtxAny,
		},
	}, nil
}

func buildXdsDownstreamTLSSocket(tlsConfig *ir.TLSConfig) (*corev3.TransportSocket, error) {
	tlsCtx := &tlsv3.DownstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
			TlsParams:                      buildTLSParams(tlsConfig),
			AlpnProtocols:                  buildALPNProtocols(tlsConfig.ALPNProtocols),
			TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{},
		},
	}

	for _, tlsConfig := range tlsConfig.Certificates {
		tlsCtx.CommonTlsContext.TlsCertificateSdsSecretConfigs = append(
			tlsCtx.CommonTlsContext.TlsCertificateSdsSecretConfigs,
			&tlsv3.SdsSecretConfig{
				Name:      tlsConfig.Name,
				SdsConfig: makeConfigSource(),
			})
	}

	if tlsConfig.CACertificate != nil {
		tlsCtx.RequireClientCertificate = &wrappers.BoolValue{Value: true}
		tlsCtx.CommonTlsContext.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
			ValidationContextSdsSecretConfig: &tlsv3.SdsSecretConfig{
				Name:      tlsConfig.CACertificate.Name,
				SdsConfig: makeConfigSource(),
			},
		}
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

func buildTLSParams(tlsConfig *ir.TLSConfig) *tlsv3.TlsParameters {
	p := &tlsv3.TlsParameters{}
	isEmpty := true
	if tlsConfig.MinVersion != nil {
		p.TlsMinimumProtocolVersion = buildTLSVersion(tlsConfig.MinVersion)
		isEmpty = false
	}
	if tlsConfig.MaxVersion != nil {
		p.TlsMaximumProtocolVersion = buildTLSVersion(tlsConfig.MaxVersion)
		isEmpty = false
	}
	if len(tlsConfig.Ciphers) > 0 {
		p.CipherSuites = tlsConfig.Ciphers
		isEmpty = false
	}
	if len(tlsConfig.ECDHCurves) > 0 {
		p.EcdhCurves = tlsConfig.ECDHCurves
		isEmpty = false
	}
	if len(tlsConfig.SignatureAlgorithms) > 0 {
		p.SignatureAlgorithms = tlsConfig.SignatureAlgorithms
		isEmpty = false
	}
	if isEmpty {
		return nil
	}
	return p
}

func buildTLSVersion(version *ir.TLSVersion) tlsv3.TlsParameters_TlsProtocol {
	lookup := map[ir.TLSVersion]tlsv3.TlsParameters_TlsProtocol{
		ir.TLSv10: tlsv3.TlsParameters_TLSv1_0,
		ir.TLSv11: tlsv3.TlsParameters_TLSv1_1,
		ir.TLSv12: tlsv3.TlsParameters_TLSv1_2,
		ir.TLSv13: tlsv3.TlsParameters_TLSv1_3,
	}
	if r, found := lookup[*version]; found {
		return r
	}
	return tlsv3.TlsParameters_TLS_AUTO
}

func buildALPNProtocols(alpn []string) []string {
	if len(alpn) == 0 {
		out := []string{"h2", "http/1.1"}
		return out
	}
	return alpn
}

func buildXdsTLSCertSecret(tlsConfig ir.TLSCertificate) *tlsv3.Secret {
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

func buildXdsTLSCaCertSecret(caCertificate *ir.TLSCACertificate) *tlsv3.Secret {
	return &tlsv3.Secret{
		Name: caCertificate.Name,
		Type: &tlsv3.Secret_ValidationContext{
			ValidationContext: &tlsv3.CertificateValidationContext{
				TrustedCa: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{InlineBytes: caCertificate.Certificate},
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

func translateEscapePath(in ir.PathEscapedSlashAction) hcmv3.HttpConnectionManager_PathWithEscapedSlashesAction {

	lookup := map[ir.PathEscapedSlashAction]hcmv3.HttpConnectionManager_PathWithEscapedSlashesAction{
		ir.KeepUnchangedAction: hcmv3.HttpConnectionManager_KEEP_UNCHANGED,
		ir.RejectRequestAction: hcmv3.HttpConnectionManager_REJECT_REQUEST,
		ir.UnescapeAndRedirect: hcmv3.HttpConnectionManager_UNESCAPE_AND_REDIRECT,
		ir.UnescapeAndForward:  hcmv3.HttpConnectionManager_UNESCAPE_AND_FORWARD,
	}
	if r, found := lookup[in]; found {
		return r
	}
	return hcmv3.HttpConnectionManager_IMPLEMENTATION_SPECIFIC_DEFAULT
}

func toNetworkFilter(filterName string, filterProto proto.Message) (*listenerv3.Filter, error) {
	filterAny, err := protocov.ToAnyWithError(filterProto)
	if err != nil {
		return nil, err
	}

	return &listenerv3.Filter{
		Name: filterName,
		ConfigType: &listenerv3.Filter_TypedConfig{
			TypedConfig: filterAny,
		},
	}, nil
}

func buildTCPProxyHashPolicy(lb *ir.LoadBalancer) []*typev3.HashPolicy {
	// Return early
	if lb == nil || lb.ConsistentHash == nil {
		return nil
	}

	if lb.ConsistentHash.SourceIP != nil && *lb.ConsistentHash.SourceIP {
		hashPolicy := &typev3.HashPolicy{
			PolicySpecifier: &typev3.HashPolicy_SourceIp_{
				SourceIp: &typev3.HashPolicy_SourceIp{},
			},
		}

		return []*typev3.HashPolicy{hashPolicy}
	}

	return nil
}

func buildHeadersWithUnderscoresAction(in *ir.HeaderSettings) corev3.HttpProtocolOptions_HeadersWithUnderscoresAction {
	if in != nil {
		switch in.WithUnderscoresAction {
		case ir.WithUnderscoresActionAllow:
			return corev3.HttpProtocolOptions_ALLOW
		case ir.WithUnderscoresActionRejectRequest:
			return corev3.HttpProtocolOptions_REJECT_REQUEST
		case ir.WithUnderscoresActionDropHeader:
			return corev3.HttpProtocolOptions_DROP_HEADER
		}
	}
	return corev3.HttpProtocolOptions_REJECT_REQUEST
}
