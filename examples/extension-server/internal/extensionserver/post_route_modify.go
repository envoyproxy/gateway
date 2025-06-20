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

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	upstreamsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
		extensionResource := unstructured.Unstructured{}
		if err := extensionResource.UnmarshalJSON(ext.GetUnstructuredBytes()); err != nil {
			return &pb.PostRouteModifyResponse{
				Route: req.Route,
			}, err
		}

		s.log.Info("processing extension resource",
			slog.String("kind", extensionResource.GetKind()),
			slog.String("apiVersion", extensionResource.GetAPIVersion()))

		// Check if it's an InferencePool
		if extensionResource.GetKind() == "InferencePool" && extensionResource.GetAPIVersion() == "sigs.k8s.io/gateway-api-inference-extension/v1alpha2" {
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
	// Create the original_destination_cluster
	originalDestCluster := &clusterv3.Cluster{
		Name: clusterName,
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			Type: clusterv3.Cluster_ORIGINAL_DST,
		},
		LbPolicy:        clusterv3.Cluster_CLUSTER_PROVIDED,
		ConnectTimeout:  durationpb.New(6 * time.Second),
		DnsLookupFamily: clusterv3.Cluster_V4_ONLY,
		LbConfig: &clusterv3.Cluster_OriginalDstLbConfig_{
			OriginalDstLbConfig: &clusterv3.Cluster_OriginalDstLbConfig{
				UseHttpHeader:  true,
				HttpHeaderName: "x-gateway-destination-endpoint",
			},
		},
	}

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
		Route:    modifiedRoute,
		Clusters: []*clusterv3.Cluster{buildExtProcCluster(inferencePool), originalDestCluster},
	}, nil
}

func buildExtProcCluster(inferencePool *inferencev1alpha2.InferencePool) *clusterv3.Cluster {
	name := clusterNameExtProc(string(inferencePool.Spec.EndpointPickerConfig.ExtensionRef.Name), inferencePool.GetNamespace())
	c := &clusterv3.Cluster{
		Name:           name,
		ConnectTimeout: durationpb.New(10 * time.Second),
		ClusterDiscoveryType: &clusterv3.Cluster_Type{
			Type: clusterv3.Cluster_STRICT_DNS,
		},
		LbPolicy: clusterv3.Cluster_LEAST_REQUEST,
		LoadAssignment: &endpointv3.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints: []*endpointv3.LocalityLbEndpoints{{
				LbEndpoints: []*endpointv3.LbEndpoint{{
					HealthStatus: corev3.HealthStatus_HEALTHY,
					HostIdentifier: &endpointv3.LbEndpoint_Endpoint{
						Endpoint: &endpointv3.Endpoint{
							Address: &corev3.Address{
								Address: &corev3.Address_SocketAddress{
									SocketAddress: &corev3.SocketAddress{
										Address:  fmt.Sprintf("%s.%s.svc", inferencePool.Spec.EndpointPickerConfig.ExtensionRef.Name, inferencePool.GetNamespace()),
										Protocol: corev3.SocketAddress_TCP,
										PortSpecifier: &corev3.SocketAddress_PortValue{
											PortValue: uint32(*inferencePool.Spec.EndpointPickerConfig.ExtensionRef.PortNumber),
										},
									},
								},
							},
						},
					},
				}},
			}},
		},
		// Ensure Envoy accepts untrusted certificates.
		TransportSocket: &corev3.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &corev3.TransportSocket_TypedConfig{
				TypedConfig: func() *anypb.Any {
					tlsCtx := &tlsv3.UpstreamTlsContext{
						CommonTlsContext: &tlsv3.CommonTlsContext{
							ValidationContextType: &tlsv3.CommonTlsContext_ValidationContext{},
						},
					}
					anyTLS, _ := messageToAny(tlsCtx)
					return anyTLS
				}(),
			},
		},
	}

	http2Opts := &upstreamsv3.HttpProtocolOptions{
		UpstreamProtocolOptions: &upstreamsv3.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &upstreamsv3.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &upstreamsv3.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
					Http2ProtocolOptions: &corev3.Http2ProtocolOptions{},
				},
			},
		},
	}

	anyHttp2, _ := messageToAny(http2Opts)
	c.TypedExtensionProtocolOptions = map[string]*anypb.Any{
		"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": anyHttp2,
	}
	return c
}

func clusterNameExtProc(name, ns string) string {
	return fmt.Sprintf("endpointpicker_%s_%s_ext_proc", name, ns)
}

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
