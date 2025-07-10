// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package extensionserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	inferencev1alpha2 "sigs.k8s.io/gateway-api-inference-extension/api/v1alpha2"

	pb "github.com/envoyproxy/gateway/proto/extension"
)

// PostRouteModify is called after Envoy Gateway is done generating a
// Route xDS configuration and before that configuration is passed on to
// Envoy Proxy.
// This implementation detects InferencePool custom backends and creates
// an original_destination_cluster for routing.
func (s *Server) PostRouteModify(ctx context.Context, req *pb.PostRouteModifyRequest) (*pb.PostRouteModifyResponse, error) {
	s.log.Info("postRouteModify callback was invoked")

	// Check if there are any InferencePool extension resources
	var hasInferencePool bool
	var inferencePool *inferencev1alpha2.InferencePool
	for _, ext := range req.PostRouteContext.ExtensionResources {
		// Parse the JSON to check the kind and apiVersion
		var resourceInfo map[string]interface{}
		if err := json.Unmarshal(ext.GetUnstructuredBytes(), &resourceInfo); err != nil {
			return &pb.PostRouteModifyResponse{
				Route: req.Route,
			}, err
		}

		kind, _ := resourceInfo["kind"].(string)
		apiVersion, _ := resourceInfo["apiVersion"].(string)

		s.log.Info("processing extension resource",
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

			hasInferencePool = true
			inferencePool = &pool
			s.log.Info("found InferencePool backend",
				slog.String("name", pool.GetName()),
				slog.String("namespace", pool.GetNamespace()),
				slog.Int("targetPortNumber", int(pool.Spec.TargetPortNumber)))
			break
		}
	}

	// If no InferencePool found, return the route unchanged
	if !hasInferencePool {
		return &pb.PostRouteModifyResponse{
			Route: req.Route,
		}, nil
	}

	// Build cluster name using InferencePool information
	clusterName := clusterNameOriginalDst(inferencePool.GetName(), inferencePool.GetNamespace())

	// Modify the route to use the dynamically named cluster
	modifiedRoute := req.Route
	if modifiedRoute.GetRoute() != nil {
		modifiedRoute.GetRoute().ClusterSpecifier = &routev3.RouteAction_Cluster{
			Cluster: clusterName,
		}
		// Set route timeout to 120 seconds
		modifiedRoute.GetRoute().Timeout = durationpb.New(120 * time.Second)
	}

	s.log.Info("successfully modified route for InferencePool backend",
		slog.String("cluster", clusterName),
		slog.String("route_name", modifiedRoute.GetName()),
		slog.String("inference_pool_name", inferencePool.GetName()),
		slog.String("inference_pool_namespace", inferencePool.GetNamespace()),
		slog.Int("target_port", int(inferencePool.Spec.TargetPortNumber)))

	return &pb.PostRouteModifyResponse{
		Route: modifiedRoute,
	}, nil
}

// Note: buildExtProcCluster function removed as cluster creation is now handled by PostClusterModify hook

func clusterNameOriginalDst(name, ns string) string {
	return fmt.Sprintf("endpointpicker_%s_%s_original_dst", name, ns)
}

func messageToAny(msg proto.Message) (*anypb.Any, error) {
	anyPb := &anypb.Any{}
	err := anypb.MarshalFrom(anyPb, msg, proto.MarshalOptions{
		Deterministic: true,
	})
	return anyPb, err
}
