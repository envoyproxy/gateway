// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

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
)

// patchGlobalResources builds and appends the global resources that are shared across listeners and routes.
// for example, the envoy client certificate and the OIDC HMAC secret.
func (t *Translator) patchGlobalResources(tCtx *types.ResourceVersionTable, globalResources *ir.GlobalResources, metrics *ir.Metrics) error {
	var errs error

	if err := t.createRateLimitServiceCluster(tCtx, &globalResources.EnvoyClientCertificate, metrics); err != nil {
		errs = errors.Join(errs, err)
	}
	if err := createEnvoyClientTLSCertSecret(tCtx, globalResources); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}

func createEnvoyClientTLSCertSecret(tCtx *types.ResourceVersionTable, globalResources *ir.GlobalResources) error {
	if err := tCtx.AddXdsResource(
		resourcev3.SecretType,
		buildXdsTLSCertSecret(globalResources.EnvoyClientCertificate)); err != nil {
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
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(host, port, false, nil)},
		Name:      destinationSettingName(clusterName),
	}

	tSocket, err := buildRateLimitTLSocket(envoyClientCertificate)
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

// buildRateLimitTLSocket builds the TLS socket for the rate limit service.
func buildRateLimitTLSocket(envoyClientCertificate *ir.TLSCertificate) (*corev3.TransportSocket, error) {
	tlsCtx := &tlsv3.UpstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
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
