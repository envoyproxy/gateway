// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const defaultConnectionTimeout = 10 * time.Second

func sdsSecretConfig(secretName, clusterName string) *tlsv3.SdsSecretConfig {
	return &tlsv3.SdsSecretConfig{
		Name: secretName,
		SdsConfig: &corev3.ConfigSource{
			ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
				ApiConfigSource: &corev3.ApiConfigSource{
					ApiType: corev3.ApiConfigSource_GRPC,
					GrpcServices: []*corev3.GrpcService{
						{
							TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
								EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
									ClusterName: clusterName,
								},
							},
						},
					},
				},
			},
		},
	}
}

// sdsClusterNameFromURL generates a unique cluster name from an SDS URL
func sdsClusterNameFromURL(url string) string {
	return ir.SDSClusterNameFromURL(url)
}

func buildSDSCluster(sdsURL string) *cluster.Cluster {
	clusterName := sdsClusterNameFromURL(sdsURL)
	return &cluster.Cluster{
		Name: clusterName,
		ClusterDiscoveryType: &cluster.Cluster_Type{
			Type: cluster.Cluster_STATIC,
		},
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: clusterName,
			Endpoints: []*endpoint.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpoint.LbEndpoint{
						{
							HostIdentifier: &endpoint.LbEndpoint_Endpoint{
								Endpoint: &endpoint.Endpoint{
									Address: &corev3.Address{
										Address: &corev3.Address_Pipe{
											Pipe: &corev3.Pipe{
												Path: sdsURL,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		ConnectTimeout:       durationpb.New(defaultConnectionTimeout),
		Http2ProtocolOptions: &corev3.Http2ProtocolOptions{},
	}
}

// createSDSCluster creates an SDS cluster for the given URL
func createSDSCluster(tCtx *types.ResourceVersionTable, sdsURL string) error {
	c := buildSDSCluster(sdsURL)

	if existing := findXdsCluster(tCtx, c.Name); existing != nil {
		if !proto.Equal(existing, c) {
			return fmt.Errorf("SDS cluster %q conflicts with an existing cluster", c.Name)
		}
		return nil
	}

	if err := tCtx.AddXdsResource(resourcev3.ClusterType, c); err != nil {
		return err
	}
	return nil
}

// processSDSClusters scans the IR for SDS URLs and creates clusters for them
func processSDSClusters(tCtx *types.ResourceVersionTable, xdsIR *ir.Xds) error {
	for _, httpListener := range xdsIR.HTTP {
		if httpListener.TLS != nil {
			for _, cert := range httpListener.TLS.Certificates {
				if cert.SDS != nil && cert.SDS.URL != "" {
					if err := createSDSCluster(tCtx, cert.SDS.URL); err != nil {
						return err
					}
				}
			}
		}

		for _, route := range httpListener.Routes {
			if route.Destination == nil {
				continue
			}
			for _, dest := range route.Destination.Settings {
				if dest.TLS != nil {
					if caCert := dest.TLS.CACertificate; caCert != nil {
						if caCert.SDS != nil && caCert.SDS.URL != "" {
							if err := createSDSCluster(tCtx, caCert.SDS.URL); err != nil {
								return err
							}
						}
					}
					for _, cert := range dest.TLS.ClientCertificates {
						if cert.SDS != nil && cert.SDS.URL != "" {
							if err := createSDSCluster(tCtx, cert.SDS.URL); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	for _, tcpListener := range xdsIR.TCP {
		for _, route := range tcpListener.Routes {
			if route.TLS != nil && route.TLS.Terminate != nil {
				for _, cert := range route.TLS.Terminate.Certificates {
					if cert.SDS != nil && cert.SDS.URL != "" {
						if err := createSDSCluster(tCtx, cert.SDS.URL); err != nil {
							return err
						}
					}
				}
			}

			if route.Destination == nil {
				continue
			}
			for _, dest := range route.Destination.Settings {
				if dest.TLS != nil {
					if caCert := dest.TLS.CACertificate; caCert != nil {
						if caCert.SDS != nil && caCert.SDS.URL != "" {
							if err := createSDSCluster(tCtx, caCert.SDS.URL); err != nil {
								return err
							}
						}
					}
					for _, cert := range dest.TLS.ClientCertificates {
						if cert.SDS != nil && cert.SDS.URL != "" {
							if err := createSDSCluster(tCtx, cert.SDS.URL); err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}
