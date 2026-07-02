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

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/cert"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	// rateLimitClientTLSCACertFilename is the ratelimit ca cert file.
	rateLimitClientTLSCACertFilename = "/certs/ca.crt"
	wasmHTTPServiceClusterName       = "wasm_cluster"
	wasmHTTPServiceHost              = "envoy-gateway"
	wasmHTTPServicePort              = 18002

	// SystemTrustStoreSecretName is the shared SDS secret name for the system CA trust store.
	SystemTrustStoreSecretName = "system_ca_certificates" //nolint:gosec // not a credential
)

// patchGlobalResources builds and appends global resources shared across listeners and routes.
func (t *Translator) patchGlobalResources(tCtx *types.ResourceVersionTable, irXds *ir.Xds) error {
	var errs error

	if irXds.GlobalResources != nil && irXds.GlobalResources.EnvoyClientCertificate != nil {
		// Create the envoy client TLS secret for control plane connections (rate limit, wasm).
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

// ensureSystemTrustStoreSecret creates the shared system trust store SDS secret if not yet present.
func ensureSystemTrustStoreSecret(tCtx *types.ResourceVersionTable) error {
	if findXdsSecret(tCtx, SystemTrustStoreSecretName) != nil {
		return nil
	}
	return tCtx.AddXdsResource(resourcev3.SecretType, &tlsv3.Secret{
		Name: SystemTrustStoreSecretName,
		Type: &tlsv3.Secret_ValidationContext{
			ValidationContext: &tlsv3.CertificateValidationContext{
				TrustedCa: &corev3.DataSource{
					Specifier: &corev3.DataSource_Filename{Filename: cert.SystemCertPath},
				},
			},
		},
	})
}

// validateSystemTrustStoreSecret verifies system_ca_certificates has expected content if present.
// Called once at end of Translate() to catch conflicts from EnvoyPatchPolicy or extensions.
func validateSystemTrustStoreSecret(tCtx *types.ResourceVersionTable) error {
	existing := findXdsSecret(tCtx, SystemTrustStoreSecretName)
	if existing == nil {
		return nil
	}
	vc := existing.GetValidationContext()
	if vc == nil || vc.GetTrustedCa().GetFilename() != cert.SystemCertPath {
		return fmt.Errorf("secret name %q is reserved for the system trust store and cannot be used by other resources", SystemTrustStoreSecretName)
	}
	return nil
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
		Weight:    new(uint32(1)),
		Protocol:  ir.GRPC,
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, host, port, false, nil)},
		Name:      destinationSettingName(clusterName),
		// TODO: tracked with issue #6861
		Metadata: nil,
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
		metadata:     ds.Metadata,
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
		Weight:    new(uint32(1)),
		Protocol:  ir.GRPC,
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, wasmHTTPServiceFQDN(t.ControllerNamespace), wasmHTTPServicePort, false, nil)},
		Name:      destinationSettingName(wasmHTTPServiceClusterName),
		// TODO: tracked with issue #6861
		Metadata: nil,
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
		metadata:     ds.Metadata,
	})
}

func wasmHTTPServiceFQDN(controllerNamespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local", wasmHTTPServiceHost, controllerNamespace)
}
