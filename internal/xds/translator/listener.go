// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	xdscore "github.com/cncf/xds/go/xds/core/v3"
	matcher "github.com/cncf/xds/go/xds/type/matcher/v3"
	mutation_rulesv3 "github.com/envoyproxy/go-control-plane/envoy/config/common/mutation_rules/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tls_inspectorv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	connection_limitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/connection_limit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	udpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	early_header_mutationv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/early_header_mutation/header_mutation/v3"
	preservecasev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/header_formatters/preserve_case/v3"
	customheaderv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/original_ip_detection/custom_header/v3"
	xffv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/original_ip_detection/xff/v3"
	quicv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/quic/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	protobuf "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
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
	networkConnectionLimit                    = "envoy.filters.network.connection_limit"
	defaultMaxAcceptConnectionsPerSocketEvent = 1
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
		preservecaseAny, _ := proto.ToAnyWithValidation(&preservecasev3.PreserveCaseFormatterConfig{})
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

func http2ProtocolOptions(opts *ir.HTTP2Settings) *corev3.Http2ProtocolOptions {
	if opts == nil {
		opts = &ir.HTTP2Settings{}
	}

	out := &corev3.Http2ProtocolOptions{
		MaxConcurrentStreams: &wrapperspb.UInt32Value{
			Value: ptr.Deref(opts.MaxConcurrentStreams, http2MaxConcurrentStreamsLimit),
		},
		InitialStreamWindowSize: &wrapperspb.UInt32Value{
			Value: ptr.Deref(opts.InitialStreamWindowSize, http2InitialStreamWindowSize),
		},
		InitialConnectionWindowSize: &wrapperspb.UInt32Value{
			Value: ptr.Deref(opts.InitialConnectionWindowSize, http2InitialConnectionWindowSize),
		},
	}

	if opts.ResetStreamOnError != nil {
		out.OverrideStreamErrorOnInvalidHttpMessage = &wrapperspb.BoolValue{
			Value: *opts.ResetStreamOnError,
		}
	}

	return out
}

// xffNumTrustedHops returns the number of hops to be configured in proxy
// Need to decrement number of hops configured by EGW user by 1 for backward compatibility
// See for more: https://github.com/envoyproxy/envoy/issues/34241
func xffNumTrustedHops(clientIPDetection *ir.ClientIPDetectionSettings) uint32 {
	if clientIPDetection != nil && clientIPDetection.XForwardedFor != nil &&
		clientIPDetection.XForwardedFor.NumTrustedHops != nil && *clientIPDetection.XForwardedFor.NumTrustedHops > 0 {
		return *clientIPDetection.XForwardedFor.NumTrustedHops - 1
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

		customHeaderConfigAny, _ := proto.ToAnyWithValidation(&customheaderv3.CustomHeaderConfig{
			HeaderName:       clientIPDetection.CustomHeader.Name,
			RejectWithStatus: rejectWithStatus,

			AllowExtensionToSetAddressAsTrusted: true,
		})

		extensionConfig = append(extensionConfig, &corev3.TypedExtensionConfig{
			Name:        "envoy.extensions.http.original_ip_detection.custom_header",
			TypedConfig: customHeaderConfigAny,
		})
	} else if clientIPDetection.XForwardedFor != nil {
		var xffHeaderConfigAny *anypb.Any
		if clientIPDetection.XForwardedFor.TrustedCIDRs != nil {
			trustedCidrs := make([]*corev3.CidrRange, 0)
			for _, cidr := range clientIPDetection.XForwardedFor.TrustedCIDRs {
				ip, nw, _ := net.ParseCIDR(string(cidr))
				prefixLen, _ := nw.Mask.Size()
				trustedCidrs = append(trustedCidrs, &corev3.CidrRange{
					AddressPrefix: ip.String(),
					PrefixLen:     wrapperspb.UInt32(uint32(prefixLen)),
				})
			}
			xffHeaderConfigAny, _ = proto.ToAnyWithValidation(&xffv3.XffConfig{
				XffTrustedCidrs: &xffv3.XffTrustedCidrs{
					Cidrs: trustedCidrs,
				},
				SkipXffAppend: wrapperspb.Bool(false),
			})
		} else if clientIPDetection.XForwardedFor.NumTrustedHops != nil {
			xffHeaderConfigAny, _ = proto.ToAnyWithValidation(&xffv3.XffConfig{
				XffNumTrustedHops: xffNumTrustedHops(clientIPDetection),
				SkipXffAppend:     wrapperspb.Bool(false),
			})
		}
		extensionConfig = append(extensionConfig, &corev3.TypedExtensionConfig{
			Name:        "envoy.extensions.http.original_ip_detection.xff",
			TypedConfig: xffHeaderConfigAny,
		})
	}

	return extensionConfig
}

// buildXdsTCPListener creates a xds Listener resource
func (t *Translator) buildXdsTCPListener(
	listenerDetails *ir.CoreListenerDetails,
	keepalive *ir.TCPKeepalive,
	connection *ir.ClientConnection,
	accesslog *ir.AccessLog,
) (*listenerv3.Listener, error) {
	socketOptions := buildTCPSocketOptions(keepalive)
	al, err := buildXdsAccessLog(accesslog, ir.ProxyAccessLogTypeListener)
	if err != nil {
		return nil, err
	}
	bufferLimitBytes := buildPerConnectionBufferLimitBytes(connection)
	maxAcceptPerSocketEvent := buildMaxAcceptPerSocketEvent(connection)
	listener := &listenerv3.Listener{
		Name: xdsListenerName(
			listenerDetails.Name, listenerDetails.ExternalPort,
			corev3.SocketAddress_TCP, t.xdsNameSchemeV2()),
		AccessLog:                            al,
		SocketOptions:                        socketOptions,
		PerConnectionBufferLimitBytes:        bufferLimitBytes,
		MaxConnectionsToAcceptPerSocketEvent: maxAcceptPerSocketEvent,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_TCP,
					Address:  listenerDetails.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: listenerDetails.Port,
					},
				},
			},
		},
	}

	if listenerDetails.IPFamily != nil && *listenerDetails.IPFamily == egv1a1.DualStack {
		socketAddress := listener.Address.GetSocketAddress()
		socketAddress.Ipv4Compat = true
	}

	return listener, nil
}

// xdsListenerName returns the name of the xDS listener in two formats:
// 1. "tcp-80" if xdsNameSchemeV2 is true.
// 2. "default/gateway-1/http" if xdsNameSchemeV2 is false.
// The second format can cause unnecessary listener drains and will be removed in the future.
// https://github.com/envoyproxy/gateway/issues/6534
func xdsListenerName(name string, externalPort uint32, protocol corev3.SocketAddress_Protocol, xdsNameSchemeV2 bool) string {
	if xdsNameSchemeV2 {
		protocolType := "tcp"
		if protocol == corev3.SocketAddress_UDP {
			protocolType = "udp"
		}
		return fmt.Sprintf("%s-%d", protocolType, externalPort)
	}

	return name
}

func buildPerConnectionBufferLimitBytes(connection *ir.ClientConnection) *wrapperspb.UInt32Value {
	if connection != nil && connection.BufferLimitBytes != nil {
		return wrapperspb.UInt32(*connection.BufferLimitBytes)
	}
	return wrapperspb.UInt32(tcpListenerPerConnectionBufferLimitBytes)
}

func buildMaxAcceptPerSocketEvent(connection *ir.ClientConnection) *wrapperspb.UInt32Value {
	if connection == nil || connection.MaxAcceptPerSocketEvent == nil {
		return wrapperspb.UInt32(defaultMaxAcceptConnectionsPerSocketEvent)
	}
	if *connection.MaxAcceptPerSocketEvent == 0 {
		return nil
	}
	return wrapperspb.UInt32(*connection.MaxAcceptPerSocketEvent)
}

// buildXdsQuicListener creates a xds Listener resource for quic
func (t *Translator) buildXdsQuicListener(
	listenerDetails *ir.CoreListenerDetails,
	ipFamily *egv1a1.IPFamily,
	accesslog *ir.AccessLog,
) (*listenerv3.Listener, error) {
	log, err := buildXdsAccessLog(accesslog, ir.ProxyAccessLogTypeListener)
	if err != nil {
		return nil, err
	}
	// Keep the listener name compatible with the old naming scheme
	listenerName := listenerDetails.Name + "-quic"
	if t.xdsNameSchemeV2() {
		listenerName = xdsListenerName(listenerDetails.Name, listenerDetails.ExternalPort, corev3.SocketAddress_UDP, true)
	}
	xdsListener := &listenerv3.Listener{
		Name:      listenerName,
		AccessLog: log,
		Address: &corev3.Address{
			Address: &corev3.Address_SocketAddress{
				SocketAddress: &corev3.SocketAddress{
					Protocol: corev3.SocketAddress_UDP,
					Address:  listenerDetails.Address,
					PortSpecifier: &corev3.SocketAddress_PortValue{
						PortValue: listenerDetails.Port,
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

	if ipFamily != nil && *ipFamily == egv1a1.DualStack {
		socketAddress := xdsListener.Address.GetSocketAddress()
		socketAddress.Ipv4Compat = true
	}

	return xdsListener, nil
}

// addHCMToXDSListener adds a HCM filter to the listener's filter chain, and adds
// all the necessary HTTP filters to that HCM.
//
//   - If tls is not enabled, a HCM filter is added to the Listener's default TCP filter chain.
//     All the ir HTTP Listeners on the same address + port combination share the
//     same HCM + HTTP filters.
//   - If tls is enabled, a new TCP filter chain is created and added to the listener.
//     A HCM filter is added to the new TCP filter chain.
//     The newly created TCP filter chain is configured with a filter chain match to
//     match the server names(SNI) based on the listener's hostnames.
func (t *Translator) addHCMToXDSListener(
	xdsListener *listenerv3.Listener,
	irListener *ir.HTTPListener,
	accesslog *ir.AccessLog,
	tracing *ir.Tracing,
	http3Listener bool,
	connection *ir.ClientConnection,
) error {
	al, err := buildXdsAccessLog(accesslog, ir.ProxyAccessLogTypeRoute)
	if err != nil {
		return err
	}

	hcmTracing, err := buildHCMTracing(tracing)
	if err != nil {
		return err
	}

	// HTTP filter configuration
	// Client IP detection
	useRemoteAddress := true
	originalIPDetectionExtensions := originalIPDetectionExtensions(irListener.ClientIPDetection)
	if originalIPDetectionExtensions != nil {
		useRemoteAddress = false
	}
	statPrefix := hcmStatPrefix(irListener, t.xdsNameSchemeV2())
	mgr := &hcmv3.HttpConnectionManager{
		AccessLog:  al,
		CodecType:  hcmv3.HttpConnectionManager_AUTO,
		StatPrefix: statPrefix,
		RouteSpecifier: &hcmv3.HttpConnectionManager_Rds{
			Rds: &hcmv3.Rds{
				ConfigSource: makeConfigSource(),
				// Configure route name to be found via RDS.
				RouteConfigName: routeConfigName(irListener, t.xdsNameSchemeV2()),
			},
		},
		HttpProtocolOptions: http1ProtocolOptions(irListener.HTTP1),
		// Hide the Envoy proxy in the Server header by default
		ServerHeaderTransformation: hcmv3.HttpConnectionManager_PASS_THROUGH,
		// Add HTTP2 protocol options
		// Set it by default to also support HTTP1.1 to HTTP2 Upgrades
		Http2ProtocolOptions: http2ProtocolOptions(irListener.HTTP2),
		// https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_conn_man/headers#x-forwarded-for
		UseRemoteAddress:              &wrapperspb.BoolValue{Value: useRemoteAddress},
		OriginalIpDetectionExtensions: originalIPDetectionExtensions,
		// normalize paths according to RFC 3986
		NormalizePath:                &wrapperspb.BoolValue{Value: true},
		MergeSlashes:                 irListener.Path.MergeSlashes,
		PathWithEscapedSlashesAction: translateEscapePath(irListener.Path.EscapedSlashesAction),
		CommonHttpProtocolOptions: &corev3.HttpProtocolOptions{
			HeadersWithUnderscoresAction: buildHeadersWithUnderscoresAction(irListener.Headers),
		},
		Tracing:                       hcmTracing,
		ForwardClientCertDetails:      buildForwardClientCertDetailsAction(irListener.Headers),
		EarlyHeaderMutationExtensions: buildEarlyHeaderMutation(irListener.Headers),
	}

	if requestID := ptr.Deref(irListener.Headers, ir.HeaderSettings{}).RequestID; requestID != nil {
		switch *requestID {
		case ir.RequestIDActionPreserveOrGenerate:
			mgr.GenerateRequestId = wrapperspb.Bool(true)
			mgr.PreserveExternalRequestId = true
		case ir.RequestIDActionPreserve:
			mgr.GenerateRequestId = wrapperspb.Bool(false)
			mgr.PreserveExternalRequestId = true
		case ir.RequestIDActionDisable:
			mgr.GenerateRequestId = wrapperspb.Bool(false)
			mgr.PreserveExternalRequestId = false
		case ir.RequestIDActionGenerate:
			mgr.GenerateRequestId = wrapperspb.Bool(true)
			mgr.PreserveExternalRequestId = false
		}
	}

	if mgr.ForwardClientCertDetails == hcmv3.HttpConnectionManager_APPEND_FORWARD || mgr.ForwardClientCertDetails == hcmv3.HttpConnectionManager_SANITIZE_SET {
		mgr.SetCurrentClientCertDetails = buildSetCurrentClientCertDetails(irListener.Headers)
	}

	if irListener.Timeout != nil && irListener.Timeout.HTTP != nil {
		if irListener.Timeout.HTTP.RequestReceivedTimeout != nil {
			mgr.RequestTimeout = durationpb.New(irListener.Timeout.HTTP.RequestReceivedTimeout.Duration)
		}

		if irListener.Timeout.HTTP.IdleTimeout != nil {
			mgr.CommonHttpProtocolOptions.IdleTimeout = durationpb.New(irListener.Timeout.HTTP.IdleTimeout.Duration)
		}

		if irListener.Timeout.HTTP.StreamIdleTimeout != nil {
			mgr.StreamIdleTimeout = durationpb.New(irListener.Timeout.HTTP.StreamIdleTimeout.Duration)
		}
	}

	// Add the proxy protocol filter if needed
	patchProxyProtocolFilter(xdsListener, irListener.ProxyProtocol)

	if irListener.IsHTTP2 {
		mgr.HttpFilters = append(mgr.HttpFilters, xdsfilters.GRPCWeb, xdsfilters.GRPCStats)
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

	if connection != nil && connection.ConnectionLimit != nil {
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
			config := irListener.TLS.DeepCopy()
			// If the listener has overlapping TLS config with other listeners, we need to disable HTTP/2
			// to avoid the HTTP/2 Connection Coalescing issue (see https://gateway-api.sigs.k8s.io/geps/gep-3567/)
			// Note: if ALPN is explicitly set by the user using ClientTrafficPolicy, we keep it as is
			if irListener.TLSOverlaps && config.ALPNProtocols == nil {
				config.ALPNProtocols = []string{"http/1.1"}
			}
			tSocket, err = buildXdsDownstreamTLSSocket(config)
			if err != nil {
				return err
			}
		}
		filterChain.TransportSocket = tSocket
		filterChain.Name = httpsListenerFilterChainName(irListener)

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
		filterChain.Name = httpListenerDefaultFilterChainName(irListener, t.xdsNameSchemeV2())
		xdsListener.DefaultFilterChain = filterChain
	}

	return nil
}

func hcmStatPrefix(irListener *ir.HTTPListener, nameSchemeV2 bool) string {
	statPrefix := "http"
	if irListener.TLS != nil {
		statPrefix = "https"
	}

	if nameSchemeV2 {
		return fmt.Sprintf("%s-%d", statPrefix, irListener.ExternalPort)
	}
	return fmt.Sprintf("%s-%d", statPrefix, irListener.Port)
}

// use the same name for the route config as the filter chain name, as they're 1:1 mapping.
func routeConfigName(irListener *ir.HTTPListener, nameSchemeV2 bool) string {
	if irListener.TLS != nil {
		return httpsListenerFilterChainName(irListener)
	}
	return httpListenerDefaultFilterChainName(irListener, nameSchemeV2)
}

// port value is used for the default filter chain name for HTTP listeners, as multiple HTTP listeners are merged into
// one filter chain.
func httpListenerDefaultFilterChainName(irListener *ir.HTTPListener, nameSchemeV2 bool) string {
	if nameSchemeV2 {
		return fmt.Sprint("http-", irListener.ExternalPort)
	}
	// For backward compatibility, we use the listener name as the filter chain name.
	return irListener.Name
}

// irListener name is used as the filter chain name for HTTPS listener, as HTTPS Listener is 1:1 mapping to the filter chain.
// The Gateway API layer ensures that each listener has a unique combination of hostname and port.
func httpsListenerFilterChainName(irListener *ir.HTTPListener) string {
	return irListener.Name
}

// irRoute name is used as the filter chain name for TLS listener, as TLSRoute is 1:1 mapping to the filter chain.
func tlsListenerFilterChainName(irRoute *ir.TCPRoute) string {
	return irRoute.Name
}

func buildEarlyHeaderMutation(headers *ir.HeaderSettings) []*corev3.TypedExtensionConfig {
	if headers == nil || (len(headers.EarlyAddRequestHeaders) == 0 && len(headers.EarlyRemoveRequestHeaders) == 0) {
		return nil
	}

	var mutationRules []*mutation_rulesv3.HeaderMutation

	for _, header := range headers.EarlyAddRequestHeaders {
		var appendAction corev3.HeaderValueOption_HeaderAppendAction
		if header.Append {
			appendAction = corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
		} else {
			appendAction = corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
		}
		// Allow empty headers to be set, but don't add the config to do so unless necessary
		if len(header.Value) == 0 {
			mutationRules = append(mutationRules, &mutation_rulesv3.HeaderMutation{
				Action: &mutation_rulesv3.HeaderMutation_Append{
					Append: &corev3.HeaderValueOption{
						Header: &corev3.HeaderValue{
							Key: header.Name,
						},
						AppendAction:   appendAction,
						KeepEmptyValue: true,
					},
				},
			})
		} else {
			for _, val := range header.Value {
				mutationRules = append(mutationRules, &mutation_rulesv3.HeaderMutation{
					Action: &mutation_rulesv3.HeaderMutation_Append{
						Append: &corev3.HeaderValueOption{
							Header: &corev3.HeaderValue{
								Key:   header.Name,
								Value: val,
							},
							AppendAction:   appendAction,
							KeepEmptyValue: val == "",
						},
					},
				})
			}
		}
	}

	for _, header := range headers.EarlyRemoveRequestHeaders {
		mr := &mutation_rulesv3.HeaderMutation{
			Action: &mutation_rulesv3.HeaderMutation_Remove{
				Remove: header,
			},
		}

		mutationRules = append(mutationRules, mr)
	}

	earlyHeaderMutationAny, _ := proto.ToAnyWithValidation(&early_header_mutationv3.HeaderMutation{
		Mutations: mutationRules,
	})

	return []*corev3.TypedExtensionConfig{
		{
			Name:        "envoy.http.early_header_mutation.header_mutation",
			TypedConfig: earlyHeaderMutationAny,
		},
	}
}

func addServerNamesMatch(xdsListener *listenerv3.Listener, filterChain *listenerv3.FilterChain, hostnames []string) error {
	// Skip adding ServerNames match for:
	// 1. nil listeners
	// 2. UDP (QUIC) listeners used for HTTP3
	// 3. wildcard hostnames
	// TODO(zhaohuabing): https://github.com/envoyproxy/gateway/issues/5660#issuecomment-3130314740
	if xdsListener == nil || (xdsListener.GetAddress() != nil &&
		xdsListener.GetAddress().GetSocketAddress() != nil &&
		xdsListener.GetAddress().GetSocketAddress().GetProtocol() == corev3.SocketAddress_UDP) {
		return nil
	}

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

func hasHCMInDefaultFilterChain(xdsListener *listenerv3.Listener) bool {
	if xdsListener == nil || xdsListener.DefaultFilterChain == nil || xdsListener.DefaultFilterChain.Filters == nil {
		return false
	}
	for _, filter := range xdsListener.DefaultFilterChain.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			return true
		}
	}
	return false
}

func (t *Translator) addXdsTCPFilterChain(
	xdsListener *listenerv3.Listener, irRoute *ir.TCPRoute, clusterName string,
	accesslog *ir.AccessLog, timeout *ir.ClientTimeout, connection *ir.ClientConnection,
) error {
	if irRoute == nil {
		return errors.New("tcp listener is nil")
	}

	isTLSPassthrough := irRoute.TLS != nil && irRoute.TLS.TLSInspectorConfig != nil
	isTLSTerminate := irRoute.TLS != nil && irRoute.TLS.Terminate != nil
	statPrefix := "tcp"
	if isTLSPassthrough {
		statPrefix = "tls-passthrough"
	}

	if isTLSTerminate {
		statPrefix = "tls-terminate"
	}

	// Append port to the statPrefix.
	statPrefix = strings.Join([]string{statPrefix, strconv.Itoa(int(xdsListener.Address.GetSocketAddress().GetPortValue()))}, "-")
	al, error := buildXdsAccessLog(accesslog, ir.ProxyAccessLogTypeRoute)
	if error != nil {
		return error
	}
	mgr := &tcpv3.TcpProxy{
		AccessLog:  al,
		StatPrefix: statPrefix,
		ClusterSpecifier: &tcpv3.TcpProxy_Cluster{
			Cluster: clusterName,
		},
		HashPolicy: buildTCPProxyHashPolicy(irRoute.LoadBalancer),
	}

	if timeout != nil && timeout.TCP != nil {
		if timeout.TCP.IdleTimeout != nil {
			mgr.IdleTimeout = durationpb.New(timeout.TCP.IdleTimeout.Duration)
		}
	}

	var filters []*listenerv3.Filter

	if connection != nil && connection.ConnectionLimit != nil {
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
		Name:    tlsListenerFilterChainName(irRoute),
		Filters: filters,
	}

	if isTLSPassthrough {
		if err := addServerNamesMatch(xdsListener, filterChain, irRoute.TLS.TLSInspectorConfig.SNIs); err != nil {
			return err
		}
	}

	if isTLSTerminate {
		var snis []string
		if cfg := irRoute.TLS.TLSInspectorConfig; cfg != nil {
			snis = cfg.SNIs
		}
		if err := addServerNamesMatch(xdsListener, filterChain, snis); err != nil {
			return err
		}
		tSocket, err := buildXdsDownstreamTLSSocket(irRoute.TLS.Terminate)
		if err != nil {
			return err
		}
		filterChain.TransportSocket = tSocket
	}

	xdsListener.FilterChains = append(xdsListener.FilterChains, filterChain)

	return nil
}

func buildConnectionLimitFilter(statPrefix string, connection *ir.ClientConnection) *connection_limitv3.ConnectionLimit {
	cl := &connection_limitv3.ConnectionLimit{
		StatPrefix:     statPrefix,
		MaxConnections: wrapperspb.UInt64(*connection.ConnectionLimit.Value),
	}

	if connection.ConnectionLimit.CloseDelay != nil {
		cl.Delay = durationpb.New(connection.ConnectionLimit.CloseDelay.Duration)
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
	tlsInspectorAny, err := proto.ToAnyWithValidation(tlsInspector)
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
		tlsCtx.DownstreamTlsContext.RequireClientCertificate = &wrapperspb.BoolValue{Value: true}
		setTLSValidationContext(tlsConfig, tlsCtx.DownstreamTlsContext.CommonTlsContext)
	}

	setDownstreamTLSSessionSettings(tlsConfig, tlsCtx.DownstreamTlsContext)

	tlsCtxAny, err := proto.ToAnyWithValidation(tlsCtx)
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
		tlsCtx.RequireClientCertificate = &wrapperspb.BoolValue{Value: tlsConfig.RequireClientCertificate}
		setTLSValidationContext(tlsConfig, tlsCtx.CommonTlsContext)
	}

	setDownstreamTLSSessionSettings(tlsConfig, tlsCtx)

	tlsCtxAny, err := proto.ToAnyWithValidation(tlsCtx)
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

func setTLSValidationContext(tlsConfig *ir.TLSConfig, tlsCtx *tlsv3.CommonTlsContext) {
	sdsSecretConfig := &tlsv3.SdsSecretConfig{
		Name:      tlsConfig.CACertificate.Name,
		SdsConfig: makeConfigSource(),
	}
	if len(tlsConfig.VerifyCertificateSpki) > 0 || len(tlsConfig.VerifyCertificateHash) > 0 || len(tlsConfig.MatchTypedSubjectAltNames) > 0 {
		validationContext := &tlsv3.CertificateValidationContext{}
		validationContext.VerifyCertificateSpki = append(validationContext.VerifyCertificateSpki, tlsConfig.VerifyCertificateSpki...)
		validationContext.VerifyCertificateHash = append(validationContext.VerifyCertificateHash, tlsConfig.VerifyCertificateHash...)
		for _, match := range tlsConfig.MatchTypedSubjectAltNames {
			sanType := tlsv3.SubjectAltNameMatcher_OTHER_NAME
			oid := ""
			switch match.Name {
			case "":
				sanType = tlsv3.SubjectAltNameMatcher_SAN_TYPE_UNSPECIFIED
			case "EMAIL":
				sanType = tlsv3.SubjectAltNameMatcher_EMAIL
			case "DNS":
				sanType = tlsv3.SubjectAltNameMatcher_DNS
			case "URI":
				sanType = tlsv3.SubjectAltNameMatcher_URI
			case "IP_ADDRESS":
				sanType = tlsv3.SubjectAltNameMatcher_IP_ADDRESS
			default:
				oid = match.Name
			}
			validationContext.MatchTypedSubjectAltNames = append(validationContext.MatchTypedSubjectAltNames, &tlsv3.SubjectAltNameMatcher{
				SanType: sanType,
				Matcher: buildXdsStringMatcher(match),
				Oid:     oid,
			})
		}
		tlsCtx.ValidationContextType = &tlsv3.CommonTlsContext_CombinedValidationContext{
			CombinedValidationContext: &tlsv3.CommonTlsContext_CombinedCertificateValidationContext{
				DefaultValidationContext:         validationContext,
				ValidationContextSdsSecretConfig: sdsSecretConfig,
			},
		}
	} else {
		tlsCtx.ValidationContextType = &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
			ValidationContextSdsSecretConfig: sdsSecretConfig,
		}
	}
}

func setDownstreamTLSSessionSettings(tlsConfig *ir.TLSConfig, tlsCtx *tlsv3.DownstreamTlsContext) {
	if !tlsConfig.StatefulSessionResumption {
		tlsCtx.DisableStatefulSessionResumption = true
	}

	if !tlsConfig.StatelessSessionResumption {
		tlsCtx.SessionTicketKeysType = &tlsv3.DownstreamTlsContext_DisableStatelessSessionResumption{
			DisableStatelessSessionResumption: true,
		}
	}
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
	if alpn == nil { // not set - default to h2 and http/1.1
		out := []string{"h2", "http/1.1"}
		return out
	} else {
		return alpn
	}
}

func buildXdsTLSCertSecret(tlsConfig ir.TLSCertificate) *tlsv3.Secret {
	return &tlsv3.Secret{
		Name: tlsConfig.Name,
		Type: &tlsv3.Secret_TlsCertificate{
			TlsCertificate: &tlsv3.TlsCertificate{
				CertificateChain: &corev3.DataSource{
					Specifier: &corev3.DataSource_InlineBytes{InlineBytes: tlsConfig.Certificate},
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

func buildXdsUDPListener(
	clusterName string,
	udpListener *ir.UDPListener,
	accesslog *ir.AccessLog,
	xdsNameSchemeV2 bool,
) (*listenerv3.Listener, error) {
	if udpListener == nil {
		return nil, errors.New("udp listener is nil")
	}

	statPrefix := "service"

	route := &udpv3.Route{
		Cluster: clusterName,
	}
	routeAny, err := proto.ToAnyWithValidation(route)
	if err != nil {
		return nil, err
	}

	al, error := buildXdsAccessLog(accesslog, ir.ProxyAccessLogTypeRoute)
	if error != nil {
		return nil, error
	}
	udpProxy := &udpv3.UdpProxyConfig{
		StatPrefix: statPrefix,
		AccessLog:  al,
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
	udpProxyAny, err := proto.ToAnyWithValidation(udpProxy)
	if err != nil {
		return nil, err
	}

	if al, err = buildXdsAccessLog(accesslog, ir.ProxyAccessLogTypeListener); err != nil {
		return nil, err
	}
	xdsListener := &listenerv3.Listener{
		Name:      xdsListenerName(udpListener.Name, udpListener.ExternalPort, corev3.SocketAddress_UDP, xdsNameSchemeV2),
		AccessLog: al,
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

	if udpListener.IPFamily != nil && *udpListener.IPFamily == egv1a1.DualStack {
		socketAddress := xdsListener.Address.GetSocketAddress()
		socketAddress.Ipv4Compat = true
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

func toNetworkFilter(filterName string, filterProto protobuf.Message) (*listenerv3.Filter, error) {
	filterAny, err := proto.ToAnyWithValidation(filterProto)
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

func buildForwardClientCertDetailsAction(in *ir.HeaderSettings) hcmv3.HttpConnectionManager_ForwardClientCertDetails {
	if in != nil {
		if in.XForwardedClientCert != nil {
			switch in.XForwardedClientCert.Mode {
			case egv1a1.XFCCForwardModeSanitize:
				return hcmv3.HttpConnectionManager_SANITIZE
			case egv1a1.XFCCForwardModeForwardOnly:
				return hcmv3.HttpConnectionManager_FORWARD_ONLY
			case egv1a1.XFCCForwardModeAppendForward:
				return hcmv3.HttpConnectionManager_APPEND_FORWARD
			case egv1a1.XFCCForwardModeSanitizeSet:
				return hcmv3.HttpConnectionManager_SANITIZE_SET
			case egv1a1.XFCCForwardModeAlwaysForwardOnly:
				return hcmv3.HttpConnectionManager_ALWAYS_FORWARD_ONLY
			}
		}
	}
	return hcmv3.HttpConnectionManager_SANITIZE
}

func buildSetCurrentClientCertDetails(in *ir.HeaderSettings) *hcmv3.HttpConnectionManager_SetCurrentClientCertDetails {
	if in == nil {
		return nil
	}

	if in.XForwardedClientCert == nil {
		return nil
	}

	if len(in.XForwardedClientCert.CertDetailsToAdd) == 0 {
		return nil
	}

	clientCertDetails := &hcmv3.HttpConnectionManager_SetCurrentClientCertDetails{}
	for _, data := range in.XForwardedClientCert.CertDetailsToAdd {
		switch data {
		case egv1a1.XFCCCertDataCert:
			clientCertDetails.Cert = true
		case egv1a1.XFCCCertDataChain:
			clientCertDetails.Chain = true
		case egv1a1.XFCCCertDataDNS:
			clientCertDetails.Dns = true
		case egv1a1.XFCCCertDataSubject:
			clientCertDetails.Subject = &wrapperspb.BoolValue{Value: true}
		case egv1a1.XFCCCertDataURI:
			clientCertDetails.Uri = true
		}
	}

	return clientCertDetails
}
