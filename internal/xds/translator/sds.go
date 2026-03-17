// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func processSDSCluster(tCtx *types.ResourceVersionTable, sds *ir.SDS) error {
	if sds == nil {
		return nil
	}
	if err := addXdsCluster(tCtx, &xdsClusterArgs{
		name:         sds.Destination.Name,
		settings:     sds.Destination.Settings,
		tSocket:      nil,
		endpointType: buildEndpointType(sds.Destination.Settings),
	}); err != nil {
		return err
	}
	return nil
}

func sdsSecretConfig(name string) *tlsv3.SdsSecretConfig {
	return &tlsv3.SdsSecretConfig{
		Name: name,
		SdsConfig: &corev3.ConfigSource{
			ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
				ApiConfigSource: &corev3.ApiConfigSource{
					ApiType: corev3.ApiConfigSource_GRPC,
					GrpcServices: []*corev3.GrpcService{
						{
							TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
								EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
									ClusterName: "sds-grpc",
								},
							},
						},
					},
				},
			},
		},
	}
}
