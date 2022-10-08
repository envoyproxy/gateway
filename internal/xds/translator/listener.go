package translator

import (
	"errors"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	tls_inspector "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/listener/tls_inspector/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tcp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildXdsListener(httpListener *ir.HTTPListener) (*listener.Listener, error) {
	if httpListener == nil {
		return nil, errors.New("http listener is nil")
	}

	routerAny, err := anypb.New(&router.Router{})
	if err != nil {
		return nil, err
	}

	// HTTP filter configuration
	mgr := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource: makeConfigSource(),
				// Configure route name to be found via RDS.
				RouteConfigName: getXdsRouteName(httpListener.Name),
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
		return nil, err
	}

	return &listener.Listener{
		Name: getXdsListenerName(httpListener.Name, httpListener.Port),
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  httpListener.Address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: httpListener.Port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Filters: []*listener.Filter{{
				Name: wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: mgrAny,
				},
			}},
		}},
	}, nil
}

func buildXdsTCPListener(clusterName string, tcpListener *ir.TCPListener) (*listener.Listener, error) {
	if tcpListener == nil {
		return nil, errors.New("http listener is nil")
	}

	statPrefix := "tcp"
	if tcpListener.TLS != nil {
		statPrefix = "passthrough"
	}
	mgr := &tcp.TcpProxy{
		StatPrefix: statPrefix,
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
	}
	mgrAny, err := anypb.New(mgr)
	if err != nil {
		return nil, err
	}

	filterChain := &listener.FilterChain{
		Filters: []*listener.Filter{{
			Name: wellknown.TCPProxy,
			ConfigType: &listener.Filter_TypedConfig{
				TypedConfig: mgrAny,
			},
		}},
	}
	if tcpListener.TLS != nil {
		filterChain.FilterChainMatch = &listener.FilterChainMatch{
			ServerNames: tcpListener.TLS.SNIs,
		}
	}

	xdsListener := &listener.Listener{
		Name: getXdsListenerName(tcpListener.Name, tcpListener.Port),
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  tcpListener.Address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: tcpListener.Port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{filterChain},
	}

	if tcpListener.TLS != nil {
		tlsInspector := &tls_inspector.TlsInspector{}
		tlsInspectorAny, err := anypb.New(tlsInspector)
		if err != nil {
			return nil, err
		}

		xdsListener.ListenerFilters = []*listener.ListenerFilter{{
			Name: wellknown.TlsInspector,
			ConfigType: &listener.ListenerFilter_TypedConfig{
				TypedConfig: tlsInspectorAny,
			},
		}}
	}

	return xdsListener, nil
}

func buildXdsDownstreamTLSSocket(listenerName string,
	tlsConfig *ir.TLSListenerConfig) (*core.TransportSocket, error) {
	tlsCtx := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: []*tls.SdsSecretConfig{{
				// Generate key name for this listener. The actual key will be
				// delivered to Envoy via SDS.
				Name:      getXdsSecretName(listenerName),
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
	tlsConfig *ir.TLSListenerConfig) (*tls.Secret, error) {
	// Build the tls secret
	return &tls.Secret{
		Name: getXdsSecretName(listenerName),
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
	}, nil
}
