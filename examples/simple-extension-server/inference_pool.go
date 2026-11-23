// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"encoding/json"
	"log/slog"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

const (
	inferencePoolAPIVersion = "inference.networking.k8s.io/v1"
	inferencePoolKind       = "InferencePool"

	// destinationEndpointHeader is the header an endpoint picker sets to the chosen endpoint
	// (host:port). ORIGINAL_DST clusters built for InferencePools route to this endpoint.
	destinationEndpointHeader = "x-gateway-destination-endpoint"

	// destinationHeaderAnnotation overrides destinationEndpointHeader for one InferencePool. Tests
	// give each pool its own header, so the pod that answers shows which pool's cluster was picked.
	destinationHeaderAnnotation = "simple-extension-server.envoyproxy.io/destination-header"
)

// inferencePool holds the InferencePool fields this server needs.
type inferencePool struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name        string            `json:"name"`
		Namespace   string            `json:"namespace"`
		Annotations map[string]string `json:"annotations"`
	} `json:"metadata"`
}

// PostClusterModify is only meaningful when InferencePool support is enabled.
//
// For a cluster built from an InferencePool backendRef, it turns the cluster into an ORIGINAL_DST
// cluster that sends each request to the endpoint named in the x-gateway-destination-endpoint
// request header, the same way an endpoint picker (EPP) would. The pool's
// simple-extension-server.envoyproxy.io/destination-header annotation can name a different header. This server does not run an EPP,
// so the header must be set by the client (or by another filter). This is only meant to let e2e
// tests check what Envoy Gateway does with custom backends (one cluster per backendRef, weighted
// clusters, and the extension resources each cluster hook call receives), without deploying a
// real inference stack.
//
// Envoy Gateway passes each cluster only the custom backend resource it was built from, so
// receiving more than one InferencePool in a single call is reported as an error.
func (s *Server) PostClusterModify(_ context.Context, req *pb.PostClusterModifyRequest) (*pb.PostClusterModifyResponse, error) {
	s.logRequest("PostClusterModify", req)

	if !s.inferencePool || req.PostClusterContext == nil {
		return &pb.PostClusterModifyResponse{Cluster: req.Cluster}, nil
	}

	var pools []*inferencePool
	for _, res := range req.PostClusterContext.BackendExtensionResources {
		pool := &inferencePool{}
		if err := json.Unmarshal(res.GetUnstructuredBytes(), pool); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal extension resource: %v", err)
		}
		if pool.APIVersion == inferencePoolAPIVersion && pool.Kind == inferencePoolKind {
			pools = append(pools, pool)
		}
	}

	switch len(pools) {
	case 0:
		return &pb.PostClusterModifyResponse{Cluster: req.Cluster}, nil
	case 1:
	default:
		return nil, status.Errorf(codes.InvalidArgument,
			"cluster %s received %d InferencePools, expected at most one", req.Cluster.GetName(), len(pools))
	}

	pool := pools[0]
	header := destinationEndpointHeader
	if h := pool.Metadata.Annotations[destinationHeaderAnnotation]; h != "" {
		header = h
	}
	modifyClusterForInferencePool(req.Cluster, header)

	s.log.Info("PostClusterModify converted cluster to ORIGINAL_DST for InferencePool",
		slog.String("cluster", req.Cluster.GetName()),
		slog.String("inferencePool", pool.Metadata.Namespace+"/"+pool.Metadata.Name),
		slog.String("header", header))

	return &pb.PostClusterModifyResponse{Cluster: req.Cluster}, nil
}

// modifyClusterForInferencePool turns cluster into an ORIGINAL_DST cluster that routes each request
// to the host:port in the given request header.
func modifyClusterForInferencePool(cluster *clusterv3.Cluster, header string) {
	cluster.ClusterDiscoveryType = &clusterv3.Cluster_Type{Type: clusterv3.Cluster_ORIGINAL_DST}
	// ORIGINAL_DST only works with the cluster-provided load balancer.
	cluster.LbPolicy = clusterv3.Cluster_CLUSTER_PROVIDED
	cluster.LoadBalancingPolicy = nil
	cluster.CommonLbConfig = nil
	cluster.LbConfig = &clusterv3.Cluster_OriginalDstLbConfig_{
		OriginalDstLbConfig: &clusterv3.Cluster_OriginalDstLbConfig{
			UseHttpHeader:  true,
			HttpHeaderName: header,
		},
	}
	cluster.EdsClusterConfig = nil
	cluster.LoadAssignment = nil
}
