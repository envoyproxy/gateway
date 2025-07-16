// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	// rateLimitClientTLSCACertFilename is the ratelimit ca cert file.
	rateLimitClientTLSCACertFilename = "/certs/ca.crt"
	wasmHTTPServiceClusterName       = "wasm_cluster"
	wasmHTTPServiceHost              = "envoy-gateway"
	wasmHTTPServicePort              = 18002
)

// patchGlobalResources builds and appends the global resources that are shared across listeners and routes.
// for example, the envoy client certificate and the OIDC HMAC secret.
func (t *Translator) patchGlobalResources(tCtx *types.ResourceVersionTable, irXds *ir.Xds) error {
	var errs error

	if irXds.GlobalResources != nil && irXds.GlobalResources.EnvoyClientCertificate != nil {
		// Create the envoy client TLS secret. It is used for envoy to establish a TLS connection with control plane components.
		if err := createEnvoyClientTLSCertSecret(tCtx, irXds.GlobalResources); err != nil {
			errs = errors.Join(errs, err)
		}

		if containsGlobalRateLimit(irXds.HTTP) {
			if err := t.createRateLimitServiceCluster(tCtx, irXds.GlobalResources.EnvoyClientCertificate, irXds.Metrics); err != nil {
				errs = errors.Join(errs, err)
			}
		}

		if containsWasm(irXds.HTTP) {
			if err := t.createWasmHTTPServiceCluster(tCtx, irXds.GlobalResources.EnvoyClientCertificate, irXds.Metrics); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

func containsGlobalRateLimit(httpListeners []*ir.HTTPListener) bool {
	for _, httpListener := range httpListeners {
		for _, route := range httpListener.Routes {
			if route.Traffic != nil &&
				route.Traffic.RateLimit != nil &&
				route.Traffic.RateLimit.Global != nil {
				return true
			}
		}
	}
	return false
}

func createEnvoyClientTLSCertSecret(tCtx *types.ResourceVersionTable, globalResources *ir.GlobalResources) error {
	if err := tCtx.AddXdsResource(
		resourcev3.SecretType,
		buildXdsTLSCertSecret(*globalResources.EnvoyClientCertificate)); err != nil {
		return err
	}
	return nil
}

func (t *Translator) createRateLimitServiceCluster(tCtx *types.ResourceVersionTable, envoyClientCertificate *ir.TLSCertificate, metrics *ir.Metrics) error {
	clusterName := getRateLimitServiceClusterName()
	// Create cluster if it does not exist
	host, port := t.getRateLimitServiceGrpcHostPort()
	ds := &ir.DestinationSetting{
		Weight:    ptr.To[uint32](1),
		Protocol:  ir.GRPC,
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, host, port, false, nil)},
		Name:      destinationSettingName(clusterName),
	}

	tSocket, err := buildEnvoyClientTLSSocket(envoyClientCertificate)
	if err != nil {
		return err
	}

	return addXdsCluster(tCtx, &xdsClusterArgs{
		name:         clusterName,
		settings:     []*ir.DestinationSetting{ds},
		tSocket:      tSocket,
		endpointType: EndpointTypeDNS,
		metrics:      metrics,
	})
}

// buildEnvoyClientTLSSocket builds the TLS socket for Envoy to connect to the control plane components.
func buildEnvoyClientTLSSocket(envoyClientCertificate *ir.TLSCertificate) (*corev3.TransportSocket, error) {
	tlsCtx := &tlsv3.UpstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
			TlsParams: &tlsv3.TlsParameters{
				TlsMaximumProtocolVersion: tlsv3.TlsParameters_TLSv1_3,
			},
			ValidationContextType: &tlsv3.CommonTlsContext_ValidationContext{
				ValidationContext: &tlsv3.CertificateValidationContext{
					TrustedCa: &corev3.DataSource{
						Specifier: &corev3.DataSource_Filename{Filename: rateLimitClientTLSCACertFilename},
					},
				},
			},
			TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{
				{
					Name:      envoyClientCertificate.Name,
					SdsConfig: makeConfigSource(),
				},
			},
		},
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

func containsWasm(httpListeners []*ir.HTTPListener) bool {
	for _, httpListener := range httpListeners {
		for _, route := range httpListener.Routes {
			if route.EnvoyExtensions != nil &&
				len(route.EnvoyExtensions.Wasms) > 0 {
				return true
			}
		}
	}
	return false
}

func (t *Translator) createWasmHTTPServiceCluster(tCtx *types.ResourceVersionTable, envoyClientCertificate *ir.TLSCertificate, metrics *ir.Metrics) error {
	ds := &ir.DestinationSetting{
		Weight:    ptr.To[uint32](1),
		Protocol:  ir.GRPC,
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, wasmHTTPServiceFQDN(t.ControllerNamespace), wasmHTTPServicePort, false, nil)},
		Name:      destinationSettingName(wasmHTTPServiceClusterName),
	}

	tSocket, err := buildEnvoyClientTLSSocket(envoyClientCertificate)
	if err != nil {
		return err
	}

	return addXdsCluster(tCtx, &xdsClusterArgs{
		name:         wasmHTTPServiceClusterName,
		settings:     []*ir.DestinationSetting{ds},
		tSocket:      tSocket,
		endpointType: EndpointTypeDNS,
		metrics:      metrics,
	})
}

func wasmHTTPServiceFQDN(controllerNamespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", wasmHTTPServiceHost, controllerNamespace)
}
