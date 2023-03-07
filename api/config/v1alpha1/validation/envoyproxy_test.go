// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func TestValidateEnvoyProxy(t *testing.T) {
	testCases := []struct {
		name     string
		obj      *egcfgv1a1.EnvoyProxy
		expected bool
	}{
		{
			name:     "nil envoyproxy",
			obj:      nil,
			expected: false,
		},
		{
			name: "nil provider",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: nil,
				},
			},
			expected: true,
		},
		{
			name: "unsupported provider",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.ResourceProvider{
						Type: egcfgv1a1.ProviderTypeFile,
					},
				},
			},
			expected: false,
		},
		{
			name: "valid user bootstrap",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: gatewayapi.StringPtr(`
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
  cdsConfig:
    ads: {}
layeredRuntime:
  layers:
  - name: runtime-0
    rtdsLayer:
      name: runtime-0
      rtdsConfig:
        ads: {}
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
                address: envoy-gateway
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
            `),
				},
			},
			expected: true,
		},
		{
			name: "user bootstrap with missing admin address",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: gatewayapi.StringPtr(`
          
					`),
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different dynamic resources",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: gatewayapi.StringPtr(`
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
    apiType: GRPC
    grpcServices:
    - envoyGrpc:
        clusterName: xds_cluster
    setNodeOnFirstMessageOnly: true
    transportApiVersion: V3
  ldsConfig:
    ads: {}
  cdsConfig:
    ads: {}
layeredRuntime:
  layers:
  - name: runtime-0
    rtdsLayer:
      name: runtime-0
      rtdsConfig:
        ads: {}
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
                address: envoy-gateway
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
            `),
				},
			},
			expected: false,
		},
		{
			name: "user bootstrap with different xds_cluster endpoint",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Bootstrap: gatewayapi.StringPtr(`
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
  cdsConfig:
    ads: {}
layeredRuntime:
  layers:
  - name: runtime-0
    rtdsLayer:
      name: runtime-0
      rtdsConfig:
        ads: {}
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
            `),
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEnvoyProxy(tc.obj)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
