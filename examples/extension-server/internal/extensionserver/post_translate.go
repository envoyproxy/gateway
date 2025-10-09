// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extensionserver

import (
	"context"
	"fmt"
	pb "github.com/envoyproxy/gateway/proto/extension"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointV3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	extensionv1alpha1 "github.com/exampleorg/envoygateway-extension/api/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
	"log/slog"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"time"
)

var (
	spireAgentCluster = &clusterv3.Cluster{
		Name:                 "spire_agent",
		ConnectTimeout:       durationpb.New(time.Second),
		Http2ProtocolOptions: &corev3.Http2ProtocolOptions{},
		LoadAssignment: &endpointV3.ClusterLoadAssignment{
			ClusterName: "spire_agent",
			Endpoints: []*endpointV3.LocalityLbEndpoints{
				{
					LbEndpoints: []*endpointV3.LbEndpoint{
						{
							HostIdentifier: &endpointV3.LbEndpoint_Endpoint{
								Endpoint: &endpointV3.Endpoint{
									Address: &corev3.Address{
										Address: &corev3.Address_Pipe{
											Pipe: &corev3.Pipe{
												Path: "/var/run/spiffe-workload-api/spire-agent.sock",
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
	}
)

func (s *Server) PostTranslateModify(ctx context.Context, req *pb.PostTranslateModifyRequest) (*pb.PostTranslateModifyResponse, error) {
	s.log.Info("PostTranslate callback was invoked",
		slog.Int("len(extension_resources)", len(req.GetPostTranslateContext().GetExtensionResources())),
	)

	var backendMtlsPolicy *extensionv1alpha1.CustomBackendMtlsPolicy

	for _, untyped := range req.GetPostTranslateContext().GetExtensionResources() {
		policy, err := s.getCustomBackendMtlsPolicy(untyped)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Backend MTLS Policy Error: %v", err)
		}
		if policy == nil {
			continue
		}

		backendMtlsPolicy = policy
		break
	}

	if backendMtlsPolicy == nil {
		return &pb.PostTranslateModifyResponse{
			Clusters:  req.GetClusters(),
			Secrets:   req.GetSecrets(),
			Listeners: req.GetListeners(),
			Routes:    req.GetRoutes(),
		}, nil
	}

	s.log.Info("CustomBackendMtlsPolicy found, adding spire agent xDS cluster and potentially adding SDS config to each backend cluster")

	clusters := req.GetClusters()
	for _, cluster := range clusters {
		err := s.maybeAddSdsConfigToCluster(cluster, backendMtlsPolicy)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "maybeAddSdsConfigToCluster for cluster %q: %v", cluster.GetName(), err)
		}
	}

	clusters = append(clusters, spireAgentCluster)
	return &pb.PostTranslateModifyResponse{
		Clusters:  clusters,
		Secrets:   req.GetSecrets(),
		Listeners: req.GetListeners(),
		Routes:    req.GetRoutes(),
	}, nil
}

func (s *Server) maybeAddSdsConfigToCluster(cluster *clusterv3.Cluster, backendMtlsPolicy *extensionv1alpha1.CustomBackendMtlsPolicy) error {
	backendRef := parseGatewayApiMetadata(cluster)
	if backendRef == nil {
		return fmt.Errorf("could not parse backend metadata")
	}

	shouldModifyCluster := false
	for _, targetRoute := range backendMtlsPolicy.Spec.TargetRoutes {
		if backendRef.Group == targetRoute.Group && backendRef.Name == targetRoute.Name && backendRef.Kind == targetRoute.Kind {
			shouldModifyCluster = true
			break
		}
	}

	if !shouldModifyCluster {
		return nil
	}

	s.log.Info("Adding SDS config to cluster", slog.String("cluster_name", cluster.GetName()))
	return nil
}

func parseGatewayApiMetadata(cluster *clusterv3.Cluster) *gwapiv1a2.LocalPolicyTargetReference {
	metadata := cluster.GetMetadata().GetFilterMetadata()
	if s, ok := metadata["envoy-gateway"]; ok {
		if v, ok := s.GetFields()["resources"]; ok {
			for _, value := range v.GetListValue().GetValues() {
				kind := value.GetStructValue().GetFields()["kind"].GetStringValue()
				name := value.GetStructValue().GetFields()["name"].GetStringValue()
				return &gwapiv1a2.LocalPolicyTargetReference{
					Group: "gateway.networking.k8s.io",
					Kind:  gwapiv1a2.Kind(kind),
					Name:  gwapiv1a2.ObjectName(name),
				}
			}
		}
	}
	return nil
}
