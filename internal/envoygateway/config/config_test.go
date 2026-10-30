// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

var (
	TLSSecretKind       = gwapiv1.Kind("Secret")
	TLSUnrecognizedKind = gwapiv1.Kind("Unrecognized")
)

// differentXdsClusterAddressBootstrap is a full bootstrap replacement whose
// xds_cluster load assignment points somewhere other than Envoy Gateway's xDS
// server. bootstrap.Validate rejects it; see the analogous testdata under
// internal/xds/bootstrap/testdata/validate for the same fixture.
var differentXdsClusterAddressBootstrap = `
admin:
  accessLog:
  - name: envoy.access_loggers.file
    typedConfig:
      '@type': type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog
      path: /dev/null
  address:
    socketAddress:
      address: 127.0.0.1
      portValue: 19000
dynamicResources:
  adsConfig:
    apiType: DELTA_GRPC
    grpcServices:
    - envoyGrpc:
        clusterName: xds_cluster
    setNodeOnFirstMessageOnly: true
    transportApiVersion: V3
  ldsConfig:
    ads: {}
    resourceApiVersion: V3
  cdsConfig:
    ads: {}
    resourceApiVersion: V3
layeredRuntime:
  layers:
  - name: runtime-0
    rtdsLayer:
      name: runtime-0
      rtdsConfig:
        ads: {}
        resourceApiVersion: V3
staticResources:
  clusters:
  - connectTimeout: 10s
    loadAssignment:
      clusterName: xds_cluster
      endpoints:
      - lbEndpoints:
        - endpoint:
            address:
              socketAddress:
                address: fake-envoy-gateway
                portValue: 18000
    name: xds_cluster
    transportSocket:
      name: envoy.transport_sockets.tls
      typedConfig:
        '@type': type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext
        commonTlsContext:
          tlsCertificateSdsSecretConfigs:
          - name: xds_certificate
            sdsConfig:
              pathConfigSource:
                path: /sds/xds-certificate.json
              resourceApiVersion: V3
          tlsParams:
            tlsMaximumProtocolVersion: TLSv1_3
          validationContextSdsSecretConfig:
            name: xds_trusted_ca
            sdsConfig:
              pathConfigSource:
                path: /sds/xds-trusted-ca.json
              resourceApiVersion: V3
    type: STRICT_DNS
    typedExtensionProtocolOptions:
      envoy.extensions.upstreams.http.v3.HttpProtocolOptions:
        '@type': type.googleapis.com/envoy.extensions.upstreams.http.v3.HttpProtocolOptions
        explicitHttpConfig:
          http2ProtocolOptions: {}
`

func TestValidate(t *testing.T) {
	cfg, err := New(os.Stdout, os.Stderr)
	require.NoError(t, err)

	testCases := []struct {
		name   string
		cfg    *Server
		expect bool
	}{
		{
			name:   "nil cfg",
			cfg:    nil,
			expect: false,
		},
		{
			name:   "default",
			cfg:    cfg,
			expect: true,
		},
		{
			name: "empty namespace",
			cfg: &Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Gateway:  egv1a1.DefaultGateway(),
						Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					},
				},
				ControllerNamespace: "",
			},
			expect: false,
		},
		{
			name: "unspecified envoy gateway",
			cfg: &Server{
				ControllerNamespace: "test-ns",
				Logger:              logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
			},
			expect: false,
		},
		{
			name: "default envoyProxy bootstrap override changes the xds cluster",
			cfg: &Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Gateway:  egv1a1.DefaultGateway(),
						Provider: egv1a1.DefaultEnvoyGatewayProvider(),
						EnvoyProxy: &egv1a1.EnvoyProxySpec{
							Bootstrap: &egv1a1.ProxyBootstrap{
								Value: &differentXdsClusterAddressBootstrap,
							},
						},
					},
				},
				ControllerNamespace: "test-ns",
			},
			expect: false,
		},
		{
			// Mirrors the EnvoyProxySpec CEL rule "mergeGateways and mergeBackends
			// cannot both be enabled". EnvoyGateway config files never go through
			// Kubernetes admission, so this embedded default has to be rejected by
			// Go-level validation instead, or the translator would later mark every
			// affected Gateway NotAccepted for this exact conflict.
			name: "default envoyProxy mergeGateways and mergeBackends both enabled",
			cfg: &Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Gateway:  egv1a1.DefaultGateway(),
						Provider: egv1a1.DefaultEnvoyGatewayProvider(),
						EnvoyProxy: &egv1a1.EnvoyProxySpec{
							MergeGateways: new(true),
							MergeBackends: &egv1a1.MergeBackendsConfig{},
						},
					},
				},
				ControllerNamespace: "test-ns",
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.cfg.Validate()
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
