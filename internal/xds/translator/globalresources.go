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
	proto "google.golang.org/protobuf/proto"
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
)

// SystemTrustStoreSecretName is re-exported from ir for use within the xDS translator package.
const SystemTrustStoreSecretName = ir.SystemTrustStoreSecretName

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

// ensureSystemTrustStoreSecret ensures a system CA trust store SDS secret with the given name
// is present, creating it if not. When dedup is enabled, name is SystemTrustStoreSecretName and
// a single shared secret is used; when dedup is disabled, name is a per-policy name and each
// cluster gets its own idempotently-created copy pointing at the same system CA file path.
func ensureSystemTrustStoreSecret(tCtx *types.ResourceVersionTable, name string) error {
	if findXdsSecret(tCtx, name) != nil {
		return nil
	}
	secret := &tlsv3.Secret{
		Name: name,
		Type: &tlsv3.Secret_ValidationContext{
			ValidationContext: systemTrustStoreValidationContext(),
		},
	}
	if err := tCtx.AddXdsResource(resourcev3.SecretType, secret); err != nil {
		return err
	}
	if name == SystemTrustStoreSecretName {
		tCtx.SystemTrustStore = true
	}
	return nil
}

// systemTrustStoreValidationContext returns the canonical ValidationContext for the system CA trust store.
// Used both when emitting the secret and when validating it hasn't been tampered with.
func systemTrustStoreValidationContext() *tlsv3.CertificateValidationContext {
	return &tlsv3.CertificateValidationContext{
		TrustedCa: &corev3.DataSource{
			Specifier: &corev3.DataSource_Filename{Filename: cert.SystemCertPath},
		},
	}
}

// validateSystemTrustStoreSecret verifies system_ca_certificates is present and unmodified
// if it was emitted during translation. A no-op if the secret was never emitted.
// The full proto is compared against the canonical form so any field mutation by
// an extension or EnvoyPatchPolicy (e.g. adding trustChainVerification, SAN matchers,
// or other validation-context fields) is detected, not just filename changes.
func validateSystemTrustStoreSecret(tCtx *types.ResourceVersionTable) error {
	if !tCtx.SystemTrustStore {
		return nil
	}
	var matches []*tlsv3.Secret
	for _, r := range tCtx.XdsResources[resourcev3.SecretType] {
		if s, ok := r.(*tlsv3.Secret); ok && s.Name == SystemTrustStoreSecretName {
			matches = append(matches, s)
		}
	}
	switch len(matches) {
	case 0:
		return fmt.Errorf("secret %q was removed by a patch or extension but is still referenced by clusters", SystemTrustStoreSecretName)
	case 1:
		canonical := canonicalSystemTrustStoreSecret()
		if !proto.Equal(matches[0], canonical) {
			return fmt.Errorf("secret name %q is reserved for the system trust store and cannot be modified by patches or extensions", SystemTrustStoreSecretName)
		}
		return nil
	default:
		return fmt.Errorf("secret name %q appears %d times; at most one is allowed", SystemTrustStoreSecretName, len(matches))
	}
}

// canonicalSystemTrustStoreSecret returns the exact proto that validateSystemTrustStoreSecret
// expects to find for system_ca_certificates. Any deviation is rejected.
func canonicalSystemTrustStoreSecret() *tlsv3.Secret {
	return &tlsv3.Secret{
		Name: SystemTrustStoreSecretName,
		Type: &tlsv3.Secret_ValidationContext{
			ValidationContext: systemTrustStoreValidationContext(),
		},
	}
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
