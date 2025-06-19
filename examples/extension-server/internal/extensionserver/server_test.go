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
	expectedClusterName := "inferencepool/default/test-inference-pool/8000"
	if resp.Route.GetRoute().GetCluster() != expectedClusterName {
		t.Errorf("expected cluster to be '%s', got %s", expectedClusterName, resp.Route.GetRoute().GetCluster())
	}

	// Check that timeout was set to 120 seconds
	expectedTimeout := durationpb.New(120 * 1000000000) // 120 seconds in nanoseconds
	if resp.Route.GetRoute().GetTimeout().GetSeconds() != expectedTimeout.GetSeconds() {
		t.Errorf("expected timeout to be 120s, got %ds", resp.Route.GetRoute().GetTimeout().GetSeconds())
	}

	// Check that a cluster was returned
	if len(resp.Clusters) != 1 {
		t.Errorf("expected 1 cluster to be returned, got %d", len(resp.Clusters))
		return // Exit early to avoid index out of bounds
	}

	// Verify the cluster configuration
	cluster := resp.Clusters[0]
	if cluster.GetName() != expectedClusterName {
		t.Errorf("expected cluster name to be '%s', got %s", expectedClusterName, cluster.GetName())
	}

	if cluster.GetType().String() != "ORIGINAL_DST" {
		t.Errorf("expected cluster type to be ORIGINAL_DST, got %s", cluster.GetType().String())
	}

	// Check original destination load balancer config
	if cluster.GetLbConfig() == nil {
		t.Error("expected LbConfig to be set")
	} else {
		if origDstConfig := cluster.GetLbConfig().(*clusterv3.Cluster_OriginalDstLbConfig_); origDstConfig != nil {
			if !origDstConfig.OriginalDstLbConfig.GetUseHttpHeader() {
				t.Error("expected UseHttpHeader to be true")
			}
			if origDstConfig.OriginalDstLbConfig.GetHttpHeaderName() != "target-pod" {
				t.Errorf("expected HttpHeaderName to be 'target-pod', got %s", origDstConfig.OriginalDstLbConfig.GetHttpHeaderName())
			}
		} else {
			t.Error("expected OriginalDstLbConfig to be set")
		}
	}
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

	// Verify that no clusters were returned
	if len(resp.Clusters) != 0 {
		t.Errorf("expected 0 clusters to be returned, got %d", len(resp.Clusters))
	}
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

	// Call PostRouteModify - should not fail but should not modify the route
	resp, err := server.PostRouteModify(context.Background(), req)
	if err != nil {
		t.Fatalf("PostRouteModify failed: %v", err)
	}

	// Verify that no clusters were returned (since invalid resource was ignored)
	if len(resp.Clusters) != 0 {
		t.Errorf("expected 0 clusters to be returned, got %d", len(resp.Clusters))
	}
}
