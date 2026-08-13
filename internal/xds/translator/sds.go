// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

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

func sdsClusterNameFromURL(url string) string {
	address := strings.TrimPrefix(url, "unix://")
	hash := sha256.Sum256([]byte(address))
	const maxReadablePrefixLength = 48

	hashSuffix := hex.EncodeToString(hash[:16])
	readablePrefix := strings.Trim(strings.ReplaceAll(address, "/", "_"), "_")
	for len(readablePrefix) > maxReadablePrefixLength {
		_, size := utf8.DecodeLastRuneInString(readablePrefix)
		readablePrefix = readablePrefix[:len(readablePrefix)-size]
	}
	if readablePrefix != "" {
		return fmt.Sprintf("sds_%s_%s", readablePrefix, hashSuffix)
	}
	return fmt.Sprintf("sds_%s", hashSuffix)
}

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

func buildSDSCluster(sdsURL string) *cluster.Cluster {
	clusterName := sdsClusterNameFromURL(sdsURL)
	pipePath := strings.TrimPrefix(sdsURL, "unix://")

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
												Path: pipePath,
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
		Http2ProtocolOptions: &corev3.Http2ProtocolOptions{}, //nolint:staticcheck
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
	sdsURLs := make(map[string]struct{})

	collectSDSURLs := func(dest []*ir.DestinationSetting) {
		for _, d := range dest {
			if d.TLS == nil {
				continue
			}
			if caCert := d.TLS.CACertificate; caCert != nil && caCert.SDS != nil && caCert.SDS.GetURL() != "" {
				sdsURLs[caCert.SDS.GetURL()] = struct{}{}
			}
			for _, cert := range d.TLS.ClientCertificates {
				if cert.SDS != nil && cert.SDS.GetURL() != "" {
					sdsURLs[cert.SDS.GetURL()] = struct{}{}
				}
			}
		}
	}

	for _, bc := range xdsIR.BackendClusters {
		collectSDSURLs([]*ir.DestinationSetting{bc.Setting})
	}

	for _, httpListener := range xdsIR.HTTP {
		if httpListener.TLS != nil {
			for _, cert := range httpListener.TLS.Certificates {
				if cert.SDS != nil && cert.SDS.GetURL() != "" {
					sdsURLs[cert.SDS.GetURL()] = struct{}{}
				}
			}
		}

		for _, route := range httpListener.Routes {
			if route.Destination != nil {
				collectSDSURLs(route.Destination.Settings)
			}
		}
	}

	for _, tcpListener := range xdsIR.TCP {
		for _, route := range tcpListener.Routes {
			if route.TLS != nil && route.TLS.Terminate != nil {
				for _, cert := range route.TLS.Terminate.Certificates {
					if cert.SDS != nil && cert.SDS.GetURL() != "" {
						sdsURLs[cert.SDS.GetURL()] = struct{}{}
					}
				}
			}

			if route.Destination != nil {
				collectSDSURLs(route.Destination.Settings)
			}
		}
	}

	urls := make([]string, 0, len(sdsURLs))
	for url := range sdsURLs {
		urls = append(urls, url)
	}
	slices.Sort(urls)

	for _, url := range urls {
		if err := createSDSCluster(tCtx, url); err != nil {
			return err
		}
	}

	return nil
}
