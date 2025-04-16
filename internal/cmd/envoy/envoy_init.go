// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import (
	"context"
	"fmt"
	"os"
	"time"

	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	DefaultEnvoyInitVolumeName     = "envoyconfigs"
	DefaultEnvoyInitConfigDir      = "/envoyconfigs"
	DefaultEnvoyInitConfigFilename = "config.json"
)

func EnvoyInit() error {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return fmt.Errorf("NODE_NAME environment variable is required")
	}

	node, err := getNode(nodeName)
	if err != nil {
		return fmt.Errorf("error getting node %q: %w", nodeName, err)
	}

	zone, err := buildLocalityZone(node)
	if err != nil {
		return fmt.Errorf("error getting node topology zone: %w", err)
	}

	jsonData, err := buildBootstrapCfg(zone)
	if err != nil {
		return fmt.Errorf("error building bootstrap config: %w", err)
	}

	if err = os.WriteFile(fmt.Sprintf("%s/%s", DefaultEnvoyInitConfigDir, DefaultEnvoyInitConfigFilename), jsonData, 0o600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Println("Successfully built service locality configuration.")
	return nil
}

func getNode(nodeName string) (*corev1.Node, error) {
	c, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to build clientset: %w", err)
	}
	return c.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
}

func getClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// buildLocalityZone configures the envoy locality zone using the Kubernetes node topology labels.
func buildLocalityZone(node *corev1.Node) (string, error) {
	zone, exists := node.Labels[corev1.LabelTopologyZone]
	if !exists {
		return "", fmt.Errorf("zone label %q not found on node %q", corev1.LabelTopologyZone, node.Name)
	}
	return zone, nil
}

func buildBootstrapCfg(zone string) ([]byte, error) {
	clusterName := "local_cluster"
	config := &bootstrapv3.Bootstrap{
		Node: &corev3.Node{
			Locality: &corev3.Locality{
				Zone: zone,
			},
		},
		ClusterManager: &bootstrapv3.ClusterManager{
			LocalClusterName: clusterName,
		},
		StaticResources: &bootstrapv3.Bootstrap_StaticResources{
			Clusters: []*clusterv3.Cluster{{
				Name: clusterName,
				ClusterDiscoveryType: &clusterv3.Cluster_Type{
					Type: clusterv3.Cluster_STATIC,
				},
				ConnectTimeout: durationpb.New(time.Second),
				LoadAssignment: &endpointv3.ClusterLoadAssignment{
					ClusterName: clusterName,
					Endpoints: []*endpointv3.LocalityLbEndpoints{{
						LbEndpoints: []*endpointv3.LbEndpoint{{
							HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
								Endpoint: &endpointv3.Endpoint{
									Address: &corev3.Address{
										Address: &corev3.Address_SocketAddress{
											SocketAddress: &corev3.SocketAddress{
												Protocol: corev3.SocketAddress_TCP,
												Address:  "0.0.0.0",
												PortSpecifier: &corev3.SocketAddress_PortValue{
													PortValue: 10080,
												},
											},
										},
									},
								},
							},
							LoadBalancingWeight: wrapperspb.UInt32(1),
						}},
						LoadBalancingWeight: wrapperspb.UInt32(1),
						Locality: &corev3.Locality{
							Zone: zone,
						},
					}},
				},
			}},
		},
	}

	mo := protojson.MarshalOptions{
		Multiline:       true,
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}
	return mo.Marshal(config)
}
