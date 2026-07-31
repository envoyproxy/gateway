// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
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
	// Sanitize the URL to create a valid cluster name
	// For Unix domain sockets like "unix:///var/run/sds" or "/var/run/sds", create a meaningful name
	if strings.HasPrefix(url, "unix://") {
		// Unix domain socket with scheme - extract the path
		path := strings.TrimPrefix(url, "unix://")
		sanitized := strings.ReplaceAll(path, "/", "_")
		sanitized = strings.Trim(sanitized, "_")
		return fmt.Sprintf("sds_%s", sanitized)
	}
	if strings.HasPrefix(url, "/") {
		// Unix domain socket path without scheme
		sanitized := strings.ReplaceAll(url, "/", "_")
		sanitized = strings.Trim(sanitized, "_")
		return fmt.Sprintf("sds_%s", sanitized)
	}
	// For other URLs, use a hash
	hash := sha256.Sum256([]byte(url))
	return fmt.Sprintf("sds_%s", hex.EncodeToString(hash[:8]))
}

// createSDSCluster creates an SDS cluster for the given URL
func createSDSCluster(tCtx *types.ResourceVersionTable, sdsURL string) error {
	clusterName := sdsClusterNameFromURL(sdsURL)

	// Check if cluster already exists
	if tCtx.XdsResources[resourcev3.ClusterType] != nil {
		for _, resource := range tCtx.XdsResources[resourcev3.ClusterType] {
			if c, ok := resource.(*cluster.Cluster); ok && c.Name == clusterName {
				// Cluster already exists
				return nil
			}
		}
	}

	// Create the cluster based on the URL type
	pipePath := sdsURL
	// Extract path for Unix domain sockets
	if strings.HasPrefix(sdsURL, "unix://") {
		pipePath = strings.TrimPrefix(sdsURL, "unix://")
	}

	c := &cluster.Cluster{
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
		Http2ProtocolOptions: &corev3.Http2ProtocolOptions{},
	}

	if err := tCtx.AddXdsResource(resourcev3.ClusterType, c); err != nil {
		return err
	}
	return nil
}

// processSDSClusters scans the IR for SDS URLs and creates clusters for them
func processSDSClusters(tCtx *types.ResourceVersionTable, xdsIR *ir.Xds) error {
	sdsURLs := make(map[string]bool)

	collectSDSURLs := func(dest []*ir.DestinationSetting) {
		for _, d := range dest {
			if d.TLS == nil {
				continue
			}
			if caCert := d.TLS.CACertificate; caCert != nil && caCert.SDS != nil && caCert.SDS.GetURL() != "" {
				sdsURLs[caCert.SDS.GetURL()] = true
			}
			for _, cert := range d.TLS.ClientCertificates {
				if cert.SDS != nil && cert.SDS.GetURL() != "" {
					sdsURLs[cert.SDS.GetURL()] = true
				}
			}
		}
	}

	for _, bc := range xdsIR.BackendClusters {
		collectSDSURLs([]*ir.DestinationSetting{bc.Setting})
	}

	for _, httpListener := range xdsIR.HTTP {
		for _, route := range httpListener.Routes {
			if route.Destination != nil {
				collectSDSURLs(route.Destination.Settings)
			}
		}
	}
	for _, tcpListener := range xdsIR.TCP {
		for _, route := range tcpListener.Routes {
			if route.Destination != nil {
				collectSDSURLs(route.Destination.Settings)
			}
		}
	}

	for url := range sdsURLs {
		if err := createSDSCluster(tCtx, url); err != nil {
			return err
		}
	}

	return nil
}
