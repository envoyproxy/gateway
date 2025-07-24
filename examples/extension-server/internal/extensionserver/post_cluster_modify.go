// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extensionserver

import (
	"context"
	"encoding/json"
	"log/slog"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	inferencev1alpha2 "sigs.k8s.io/gateway-api-inference-extension/api/v1alpha2"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

// PostClusterModify is called after Envoy Gateway is done generating
// a Cluster xDS configuration and before that configuration is passed on to
// Envoy Proxy.
// This implementation modifies the cluster for InferencePool custom backends.
func (s *Server) PostClusterModify(ctx context.Context, req *pb.PostClusterModifyRequest) (*pb.PostClusterModifyResponse, error) {
	s.log.Info("postClusterModify callback was invoked", slog.String("cluster_name", req.Cluster.Name))

	// Parse extension resources to find InferencePool configurations
	var inferencePoolConfigs []*inferencev1alpha2.InferencePool
	for _, ext := range req.PostClusterContext.BackendExtensionResources {
		// Parse the JSON to check the kind and apiVersion
		var resourceInfo map[string]interface{}
		if err := json.Unmarshal(ext.GetUnstructuredBytes(), &resourceInfo); err != nil {
			s.log.Error("failed to unmarshal extension resource", slog.String("error", err.Error()))
			continue
		}

		kind, _ := resourceInfo["kind"].(string)
		apiVersion, _ := resourceInfo["apiVersion"].(string)

		s.log.Info("processing extension resource for cluster modification",
			slog.String("kind", kind),
			slog.String("apiVersion", apiVersion))

		// Check if it's an InferencePool
		if kind == "InferencePool" && apiVersion == "sigs.k8s.io/gateway-api-inference-extension/v1alpha2" {
			// Now unmarshal directly to InferencePool type
			var pool inferencev1alpha2.InferencePool
			if err := json.Unmarshal(ext.GetUnstructuredBytes(), &pool); err != nil {
				s.log.Error("failed to unmarshal InferencePool", slog.String("error", err.Error()))
				continue
			}

			s.log.Info("found InferencePool for cluster modification",
				slog.String("name", pool.GetName()),
				slog.String("namespace", pool.GetNamespace()),
				slog.Int("targetPortNumber", int(pool.Spec.TargetPortNumber)))

			inferencePoolConfigs = append(inferencePoolConfigs, &pool)
		}
	}
	if len(inferencePoolConfigs) == 1 {
		// Modify the cluster based on InferencePool configurations
		modifiedCluster := s.modifyClusterForInferencePool(req.Cluster, inferencePoolConfigs[0])
		s.log.Info("successfully processed cluster modification",
			slog.String("cluster_name", req.Cluster.Name),
			slog.Int("inference_pools", len(inferencePoolConfigs)))

		return &pb.PostClusterModifyResponse{
			Cluster: modifiedCluster,
		}, nil
	}

	return &pb.PostClusterModifyResponse{
		Cluster: req.Cluster,
	}, nil
}

// modifyClusterForInferencePool modifies an existing cluster based on InferencePool configurations
func (s *Server) modifyClusterForInferencePool(cluster *clusterv3.Cluster, pool *inferencev1alpha2.InferencePool) *clusterv3.Cluster {
	s.log.Info("modifying cluster for InferencePool",
		slog.String("cluster_name", cluster.Name),
		slog.String("inference_pool", pool.GetName()))

	// Convert to ORIGINAL_DST cluster type
	modifiedCluster := s.convertToOriginalDestCluster(cluster, pool)
	return modifiedCluster
}

// convertToOriginalDestCluster converts a regular cluster to an ORIGINAL_DST cluster for InferencePool
func (s *Server) convertToOriginalDestCluster(originalCluster *clusterv3.Cluster, pool *inferencev1alpha2.InferencePool) *clusterv3.Cluster {
	originalCluster.LbPolicy = clusterv3.Cluster_CLUSTER_PROVIDED
	originalCluster.ClusterDiscoveryType = &clusterv3.Cluster_Type{
		Type: clusterv3.Cluster_ORIGINAL_DST,
	}
	originalCluster.LbConfig = &clusterv3.Cluster_OriginalDstLbConfig_{
		OriginalDstLbConfig: &clusterv3.Cluster_OriginalDstLbConfig{
			UseHttpHeader:  true,
			HttpHeaderName: "x-gateway-destination-endpoint",
		},
	}
	originalCluster.ConnectTimeout = durationpb.New(10 * 1000000000)
	originalCluster.EdsClusterConfig = nil
	originalCluster.LoadAssignment = nil

	s.log.Info("successfully converted cluster to ORIGINAL_DST type",
		slog.String("cluster_name", originalCluster.Name),
		slog.String("inference_pool", pool.GetName()))

	return originalCluster
}
