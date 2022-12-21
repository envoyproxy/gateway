// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	xdscore "github.com/cncf/xds/go/xds/core/v3"
	matcher "github.com/cncf/xds/go/xds/type/matcher/v3"
	accesslog "github.com/envoyproxy/go-control-plane/envoy/config/accesslog/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	tls_inspector "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	udp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/udp/udp_proxy/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildXdsTCPListener(name, address string, port uint32) *listener.Listener {
	accesslogAny, _ := anypb.New(stdoutFileAccessLog)
	return &listener.Listener{
		Name: name,
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
				Filter:     listenerAccessLogFilter,
			},
		},
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
	}
}

func addXdsHTTPFilterChain(xdsListener *listener.Listener, irListener *ir.HTTPListener) error {
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
	mgr := &hcm.HttpConnectionManager{
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
			},
		},
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: statPrefix,
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource: makeConfigSource(),
				// Configure route name to be found via RDS.
				RouteConfigName: irListener.Name,
			},
		},
		// Use only router.
		HttpFilters: []*hcm.HttpFilter{{
			Name:       wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerAny},
		}},
	}

	mgrAny, err := anypb.New(mgr)
	if err != nil {
		return err
	}

	filterChain := &listener.FilterChain{
		Filters: []*listener.Filter{{
			Name: wellknown.HTTPConnectionManager,
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: mgrAny,
			},
		}},
	}

	if irListener.TLS != nil {
		tSocket, err := buildXdsDownstreamTLSSocket(irListener.Name)
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

func addServerNamesMatch(xdsListener *listener.Listener, filterChain *listener.FilterChain, hostnames []string) error {
	// Dont add a filter chain match if the hostname is a wildcard character.
	if len(hostnames) > 0 && hostnames[0] != "*" {
		filterChain.FilterChainMatch = &listener.FilterChainMatch{
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
func findXdsHTTPRouteConfigName(xdsListener *listener.Listener) string {
	if xdsListener == nil || xdsListener.DefaultFilterChain == nil || xdsListener.DefaultFilterChain.Filters == nil {
		return ""
	}

	for _, filter := range xdsListener.DefaultFilterChain.Filters {
		if filter.Name == wellknown.HTTPConnectionManager {
			m := new(hcm.HttpConnectionManager)
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

func addXdsTCPFilterChain(xdsListener *listener.Listener, irListener *ir.TCPListener, clusterName string) error {
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

	mgr := &tcp.TcpProxy{
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
			},
		},
		StatPrefix: statPrefix,
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
	}
	mgrAny, err := anypb.New(mgr)
	if err != nil {
		return err
	}

	filterChain := &listener.FilterChain{
		Filters: []*listener.Filter{{
			Name: wellknown.TCPProxy,
			ConfigType: &listener.Filter_TypedConfig{
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
func addXdsTLSInspectorFilter(xdsListener *listener.Listener) error {
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

	filter := &listener.ListenerFilter{
		Name: wellknown.TlsInspector,
		ConfigType: &listener.ListenerFilter_TypedConfig{
			TypedConfig: tlsInspectorAny,
		},
	}

	xdsListener.ListenerFilters = append(xdsListener.ListenerFilters, filter)

	return nil
}

func buildXdsDownstreamTLSSocket(listenerName string) (*core.TransportSocket, error) {
	tlsCtx := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: []*tls.SdsSecretConfig{{
				// Generate key name for this listener. The actual key will be
				// delivered to Envoy via SDS.
				Name:      listenerName,
				SdsConfig: makeConfigSource(),
			}},
		},
	}

	tlsCtxAny, err := anypb.New(tlsCtx)
	if err != nil {
		return nil, err
	}

	return &core.TransportSocket{
		Name: wellknown.TransportSocketTls,
		ConfigType: &core.TransportSocket_TypedConfig{
			TypedConfig: tlsCtxAny,
		},
	}, nil
}

func buildXdsDownstreamTLSSecret(listenerName string,
	tlsConfig *ir.TLSListenerConfig) *tls.Secret {
	// Build the tls secret
	return &tls.Secret{
		Name: listenerName,
		Type: &tls.Secret_TlsCertificate{
			TlsCertificate: &tls.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: tlsConfig.ServerCertificate},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineBytes{InlineBytes: tlsConfig.PrivateKey},
				},
			},
		},
	}
}

func buildXdsUDPListener(clusterName string, udpListener *ir.UDPListener) (*listener.Listener, error) {
	if udpListener == nil {
		return nil, errors.New("udp listener is nil")
	}

	statPrefix := "service"

	route := &udp.Route{
		Cluster: clusterName,
	}
	routeAny, err := anypb.New(route)
	if err != nil {
		return nil, err
	}
	accesslogAny, _ := anypb.New(stdoutFileAccessLog)
	udpProxy := &udp.UdpProxyConfig{
		StatPrefix: statPrefix,
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
			},
		},
		RouteSpecifier: &udp.UdpProxyConfig_Matcher{
			Matcher: &matcher.Matcher{
				OnNoMatch: &matcher.Matcher_OnMatch{
					OnMatch: &matcher.Matcher_OnMatch_Action{
						Action: &xdscore.TypedExtensionConfig{
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

	filterChain := &listener.FilterChain{
		Filters: []*listener.Filter{{
			Name: "envoy.filters.udp_listener.udp_proxy",
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: udpProxyAny,
			},
		}},
	}

	xdsListener := &listener.Listener{
		Name: udpListener.Name,
		AccessLog: []*accesslog.AccessLog{
			{
				Name:       wellknown.FileAccessLog,
				ConfigType: &accesslog.AccessLog_TypedConfig{TypedConfig: accesslogAny},
			},
		},
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_UDP,
					Address:  udpListener.Address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: udpListener.Port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{filterChain},
	}

	return xdsListener, nil
}
