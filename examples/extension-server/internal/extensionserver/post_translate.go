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
	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	extensionv1alpha1 "github.com/exampleorg/envoygateway-extension/api/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"log/slog"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"time"
)

const (
	spireAgentClusterName = "spire_agent"
)

var (
	spireAgentCluster = &clusterv3.Cluster{
		Name:                 spireAgentClusterName,
		ConnectTimeout:       durationpb.New(time.Second),
		Http2ProtocolOptions: &corev3.Http2ProtocolOptions{},
		LoadAssignment: &endpointV3.ClusterLoadAssignment{
			ClusterName: spireAgentClusterName,
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

	sdsConfigSource = &corev3.ConfigSource{
		ResourceApiVersion: corev3.ApiVersion_V3,
		ConfigSourceSpecifier: &corev3.ConfigSource_ApiConfigSource{
			ApiConfigSource: &corev3.ApiConfigSource{
				ApiType:             corev3.ApiConfigSource_GRPC,
				TransportApiVersion: corev3.ApiVersion_V3,
				GrpcServices: []*corev3.GrpcService{
					{
						TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
							EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
								ClusterName: spireAgentClusterName,
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
		return nil
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

	return s.addSdsConfigToCluster(cluster, backendMtlsPolicy)
}

func (s *Server) addSdsConfigToCluster(cluster *clusterv3.Cluster, policy *extensionv1alpha1.CustomBackendMtlsPolicy) error {
	s.log.Info("Adding SDS config to cluster", slog.String("cluster_name", cluster.GetName()))
	spiffeId := generateSpiffeID(policy)

	for _, transportSocketMatch := range cluster.TransportSocketMatches {
		if !isTlsTransportSocket(transportSocketMatch) {
			continue
		}

		if err := configureTransportSocketMatch(transportSocketMatch, spiffeId); err != nil {
			return fmt.Errorf("addSdsConfigToCluster for transport socket match %q: %v", transportSocketMatch.GetName(), err)
		}
		s.log.Info("Added SDS config to cluster",
			slog.String("cluster_name", cluster.GetName()),
			slog.String("transport_socket_match", transportSocketMatch.GetName()),
			slog.String("spiffe_id", spiffeId),
		)
	}

	return nil
}

func isTlsTransportSocket(match *clusterv3.Cluster_TransportSocketMatch) bool {
	return match.TransportSocket != nil && match.TransportSocket.Name == "envoy.transport_sockets.tls"
}

func configureTransportSocketMatch(match *clusterv3.Cluster_TransportSocketMatch, spiffeId string) error {
	upstreamTlsCtx := &tlsV3.UpstreamTlsContext{}
	if err := match.GetTransportSocket().GetTypedConfig().UnmarshalTo(upstreamTlsCtx); err != nil {
		return fmt.Errorf("unmarshal UpstreamTlsContext failed: %v", err)
	}

	if upstreamTlsCtx.GetCommonTlsContext() == nil {
		return fmt.Errorf("upstreamTlsCtx common tls context is nil")
	}

	upstreamTlsCtx.CommonTlsContext.TlsCertificateSdsSecretConfigs = []*tlsV3.SdsSecretConfig{
		{
			Name:      spiffeId,
			SdsConfig: sdsConfigSource,
		},
	}

	typedConfigAny, err := anypb.New(upstreamTlsCtx)
	if err != nil {
		return fmt.Errorf("marshal UpstreamTlsContext failed: %v", err)
	}

	match.TransportSocket.ConfigType = &corev3.TransportSocket_TypedConfig{
		TypedConfig: typedConfigAny,
	}

	return nil
}

func generateSpiffeID(policy *extensionv1alpha1.CustomBackendMtlsPolicy) string {
	return fmt.Sprintf("spiffe://%s/%s", policy.Spec.TrustDomain, policy.Spec.WorkloadIdentifier)
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
