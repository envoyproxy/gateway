// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extensionserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/types/known/durationpb"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

func TestPostRouteModify_WithInferencePool(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	server := New(logger)

	// Marshal the InferencePool to JSON as unstructured
	unstructuredObj := map[string]interface{}{
		"kind":       "InferencePool",
		"apiVersion": "sigs.k8s.io/gateway-api-inference-extension/v1alpha2",
		"metadata": map[string]interface{}{
			"name":      "test-inference-pool",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"targetPortNumber": 8000,
			"selector": map[string]interface{}{
				"app": "vllm-llama3-8b-instruct",
			},
		},
	}

	inferencePoolBytes, err := json.Marshal(unstructuredObj)
	if err != nil {
		t.Fatalf("failed to marshal InferencePool: %v", err)
	}

	// Create a test route
	testRoute := &routev3.Route{
		Name: "test-route",
		Match: &routev3.RouteMatch{
			PathSpecifier: &routev3.RouteMatch_Prefix{
				Prefix: "/v1",
			},
		},
		Action: &routev3.Route_Route{
			Route: &routev3.RouteAction{
				ClusterSpecifier: &routev3.RouteAction_Cluster{
					Cluster: "original-cluster",
				},
			},
		},
	}

	// Create the request
	req := &pb.PostRouteModifyRequest{
		Route: testRoute,
		PostRouteContext: &pb.PostRouteExtensionContext{
			ExtensionResources: []*pb.ExtensionResource{
				{
					UnstructuredBytes: inferencePoolBytes,
				},
			},
			Hostnames: []string{"example.com"},
		},
	}

	// Call PostRouteModify
	resp, err := server.PostRouteModify(context.Background(), req)
	if err != nil {
		t.Fatalf("PostRouteModify failed: %v", err)
	}

	// Verify the response
	if resp.Route == nil {
		t.Fatal("expected route to be returned")
	}

	// Check that the cluster was changed to the expected InferencePool cluster name
	expectedClusterName := "endpointpicker_test-inference-pool_default_original_dst"
	if resp.Route.GetRoute().GetCluster() != expectedClusterName {
		t.Errorf("expected cluster to be '%s', got %s", expectedClusterName, resp.Route.GetRoute().GetCluster())
	}

	// Check that timeout was set to 120 seconds
	expectedTimeout := durationpb.New(120 * 1000000000) // 120 seconds in nanoseconds
	if resp.Route.GetRoute().GetTimeout().GetSeconds() != expectedTimeout.GetSeconds() {
		t.Errorf("expected timeout to be 120s, got %ds", resp.Route.GetRoute().GetTimeout().GetSeconds())
	}

	// Note: Cluster creation is now handled by PostClusterModify hook, not PostRouteModify
	// This test now only verifies that the route was modified correctly
}

func TestPostRouteModify_WithoutInferencePool(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	server := New(logger)

	// Create a test route
	testRoute := &routev3.Route{
		Name: "test-route",
		Match: &routev3.RouteMatch{
			PathSpecifier: &routev3.RouteMatch_Prefix{
				Prefix: "/v1",
			},
		},
		Action: &routev3.Route_Route{
			Route: &routev3.RouteAction{
				ClusterSpecifier: &routev3.RouteAction_Cluster{
					Cluster: "original-cluster",
				},
			},
		},
	}

	// Create a request without InferencePool
	req := &pb.PostRouteModifyRequest{
		Route: testRoute,
		PostRouteContext: &pb.PostRouteExtensionContext{
			ExtensionResources: []*pb.ExtensionResource{},
			Hostnames:          []string{"example.com"},
		},
	}

	// Call PostRouteModify
	resp, err := server.PostRouteModify(context.Background(), req)
	if err != nil {
		t.Fatalf("PostRouteModify failed: %v", err)
	}

	// Verify that the route was not modified
	if resp.Route.GetRoute().GetCluster() != "original-cluster" {
		t.Errorf("expected cluster to remain 'original-cluster', got %s", resp.Route.GetRoute().GetCluster())
	}

	// Note: PostRouteModify no longer returns clusters - they are handled by PostClusterModify hook
}

func TestPostRouteModify_WithInvalidExtensionResource(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	server := New(logger)

	// Create a test route
	testRoute := &routev3.Route{
		Name: "test-route",
	}

	// Create a request with invalid extension resource
	req := &pb.PostRouteModifyRequest{
		Route: testRoute,
		PostRouteContext: &pb.PostRouteExtensionContext{
			ExtensionResources: []*pb.ExtensionResource{
				{
					UnstructuredBytes: []byte("invalid json"),
				},
			},
		},
	}

	// Call PostRouteModify - should fail due to invalid JSON
	_, err := server.PostRouteModify(context.Background(), req)
	if err == nil {
		t.Fatal("PostRouteModify should have failed with invalid JSON")
	}

	// Note: PostRouteModify no longer returns clusters - they are handled by PostClusterModify hook
}

func TestPostClusterModify_WithInferencePool(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	server := New(logger)

	// Marshal the InferencePool to JSON as unstructured
	unstructuredObj := map[string]interface{}{
		"kind":       "InferencePool",
		"apiVersion": "sigs.k8s.io/gateway-api-inference-extension/v1alpha2",
		"metadata": map[string]interface{}{
			"name":      "test-inference-pool",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"targetPortNumber": 8000,
			"selector": map[string]interface{}{
				"app": "vllm-llama3-8b-instruct",
			},
		},
	}

	inferencePoolBytes, err := json.Marshal(unstructuredObj)
	if err != nil {
		t.Fatalf("failed to marshal InferencePool: %v", err)
	}

	// Create an existing cluster that should be modified
	existingCluster := &clusterv3.Cluster{
		Name: "inferencepool/default/test-inference-pool/8000",
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			Type: clusterv3.Cluster_EDS, // Original type, should be changed to ORIGINAL_DST
		},
		LbPolicy: clusterv3.Cluster_ROUND_ROBIN,
	}

	// Create the request
	req := &pb.PostClusterModifyRequest{
		Cluster: existingCluster, // Provide existing cluster to modify
		PostClusterContext: &pb.PostClusterExtensionContext{
			BackendExtensionResources: []*pb.ExtensionResource{
				{
					UnstructuredBytes: inferencePoolBytes,
				},
			},
		},
	}

	// Call PostClusterModify
	resp, err := server.PostClusterModify(context.Background(), req)
	if err != nil {
		t.Fatalf("PostClusterModify failed: %v", err)
	}

	// Verify the response
	if resp.Cluster == nil {
		t.Fatalf("expected a cluster to be returned, got nil")
	}

	// Check that the cluster name remains the same (it was modified, not replaced)
	expectedClusterName := "inferencepool/default/test-inference-pool/8000"
	if resp.Cluster.Name != expectedClusterName {
		t.Errorf("expected cluster name to be '%s', got %s", expectedClusterName, resp.Cluster.Name)
	}

	// Check that it's an ORIGINAL_DST cluster
	if resp.Cluster.GetType() != clusterv3.Cluster_ORIGINAL_DST {
		t.Errorf("expected cluster type to be ORIGINAL_DST, got %v", resp.Cluster.GetType())
	}
}
