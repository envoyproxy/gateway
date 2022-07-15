package translator

import (
	"errors"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/pkg/ir"
)

func buildXdsListener(httpListener *ir.HTTPListener) (*listener.Listener, error) {
	if httpListener == nil {
		return nil, errors.New("http listener is nil")
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
			Name: wellknown.Router,
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
