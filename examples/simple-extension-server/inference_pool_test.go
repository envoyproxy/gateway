// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"log/slog"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

const testPool = `{"apiVersion":"inference.networking.k8s.io/v1","kind":"InferencePool",
"metadata":{"name":"pool-a","namespace":"default"},"spec":{"targetPorts":[{"number":8000}]}}`

func clusterRequest(resources ...string) *pb.PostClusterModifyRequest {
	ctx := &pb.PostClusterExtensionContext{}
	for _, r := range resources {
		ctx.BackendExtensionResources = append(ctx.BackendExtensionResources,
			&pb.ExtensionResource{UnstructuredBytes: []byte(r)})
	}
	return &pb.PostClusterModifyRequest{
		Cluster:            &clusterv3.Cluster{Name: "httproute/default/route/rule/0/backend/0"},
		PostClusterContext: ctx,
	}
}

func TestPostClusterModifyInferencePool(t *testing.T) {
	s := New(slog.Default(), nil, "extserver", true)

	resp, err := s.PostClusterModify(context.Background(), clusterRequest(testPool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c := resp.Cluster
	if c.GetType() != clusterv3.Cluster_ORIGINAL_DST {
		t.Fatalf("expected ORIGINAL_DST cluster, got %v", c.GetType())
	}
	if c.GetLbPolicy() != clusterv3.Cluster_CLUSTER_PROVIDED {
		t.Fatalf("expected CLUSTER_PROVIDED lb policy, got %v", c.GetLbPolicy())
	}
	cfg := c.GetOriginalDstLbConfig()
	if !cfg.GetUseHttpHeader() || cfg.GetHttpHeaderName() != destinationEndpointHeader {
		t.Fatalf("unexpected original dst config %v", cfg)
	}
	if c.GetLoadAssignment() != nil {
		t.Fatalf("expected no load assignment, got %v", c.GetLoadAssignment())
	}
}

func TestPostClusterModifyInferencePoolHeaderAnnotation(t *testing.T) {
	s := New(slog.Default(), nil, "extserver", true)
	pool := `{"apiVersion":"inference.networking.k8s.io/v1","kind":"InferencePool",
"metadata":{"name":"pool-a","namespace":"default","annotations":{"` + destinationHeaderAnnotation + `":"x-pool-a"}}}`

	resp, err := s.PostClusterModify(context.Background(), clusterRequest(pool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if h := resp.Cluster.GetOriginalDstLbConfig().GetHttpHeaderName(); h != "x-pool-a" {
		t.Fatalf("expected header x-pool-a, got %q", h)
	}
}

func TestPostClusterModifyInferencePoolRejectsMultiplePools(t *testing.T) {
	s := New(slog.Default(), nil, "extserver", true)

	if _, err := s.PostClusterModify(context.Background(), clusterRequest(testPool, testPool)); err == nil {
		t.Fatal("expected an error for a cluster with two InferencePools")
	}
}

func TestPostClusterModifyInferencePoolDisabled(t *testing.T) {
	s := New(slog.Default(), nil, "extserver", false)

	resp, err := s.PostClusterModify(context.Background(), clusterRequest(testPool, testPool))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Cluster.GetType() != clusterv3.Cluster_STATIC || resp.Cluster.GetLoadAssignment() != nil {
		t.Fatalf("expected the cluster to be left unchanged, got %v", resp.Cluster)
	}
}
