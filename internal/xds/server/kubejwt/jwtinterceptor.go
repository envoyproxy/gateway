// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubejwt

import (
	"context"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/kubernetes"

	"github.com/envoyproxy/gateway/internal/logging"
)

// JWTAuthInterceptor verifies Kubernetes Service Account JWT tokens in gRPC requests.
type JWTAuthInterceptor struct {
	clientset *kubernetes.Clientset
	issuer    string
	audience  string
	logger    logging.Logger
}

// NewJWTAuthInterceptor initializes a new JWTAuthInterceptor.
func NewJWTAuthInterceptor(logger logging.Logger, clientset *kubernetes.Clientset, issuer, audience string) *JWTAuthInterceptor {
	return &JWTAuthInterceptor{
		clientset: clientset,
		issuer:    issuer,
		audience:  audience,
		logger:    logger.WithName("jwt-auth-interceptor"),
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx         context.Context
	interceptor *JWTAuthInterceptor
	// authNode is the node authenticated on the first request. It pins the stream:
	// later requests may omit the node (the xDS server reuses the first one) but must
	// not switch to a different node.
	authNode *corev3.Node
}

func (i *JWTAuthInterceptor) authenticate(ctx context.Context, msg any) (*corev3.Node, error) {
	node, err := extractNode(msg)
	if err != nil {
		return nil, err
	}
	nodeID := node.GetId()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata for node %s", nodeID)
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing authorization header for node %s", nodeID)
	}
	token := strings.TrimPrefix(authHeader[0], "Bearer ")

	if err := i.validateKubeJWT(ctx, token, nodeID, node.GetCluster()); err != nil {
		if s, ok := status.FromError(err); ok {
			return nil, status.Errorf(s.Code(), "failed to validate the token for node %s: %s", nodeID, s.Message())
		}
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate the token for node %s: %v", nodeID, err)
	}

	return node, nil
}

func (w *wrappedStream) RecvMsg(m any) error {
	if err := w.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	if w.authNode == nil {
		node, err := w.interceptor.authenticate(w.ctx, m)
		if err != nil {
			return err
		}
		w.authNode = node
		return nil
	}

	// The xDS server keys watches off the per-request node, and only the first
	// request is guaranteed to carry it. A later request may omit the node, but if
	// it carries one it must match the authenticated node: otherwise an
	// authenticated stream could pull another proxy's resources by switching node.
	if node := requestNode(m); node != nil && node.GetId() != "" &&
		(node.GetId() != w.authNode.GetId() || node.GetCluster() != w.authNode.GetCluster()) {
		return status.Errorf(codes.Unauthenticated,
			"node changed within stream: authenticated %s/%s, got %s/%s",
			w.authNode.GetCluster(), w.authNode.GetId(), node.GetCluster(), node.GetId())
	}

	return nil
}

func newWrappedStream(s grpc.ServerStream, ctx context.Context, interceptor *JWTAuthInterceptor) grpc.ServerStream {
	return &wrappedStream{ServerStream: s, ctx: ctx, interceptor: interceptor}
}

// Stream intercepts streaming gRPC calls for authorization.
func (i *JWTAuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapped := newWrappedStream(ss, ss.Context(), i)
		return handler(srv, wrapped)
	}
}

// Unary intercepts unary gRPC calls for authorization.
func (i *JWTAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, err := i.authenticate(ctx, req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// requestNode returns the node carried by a discovery request, or nil for an
// unknown message type or a request without a node.
func requestNode(m any) *corev3.Node {
	switch req := m.(type) {
	case *discoveryv3.DeltaDiscoveryRequest:
		return req.GetNode()
	case *discoveryv3.DiscoveryRequest:
		return req.GetNode()
	default:
		return nil
	}
}

// extractNode returns the node of a discovery request, requiring a non-empty
// node ID. The first request on a stream must carry it.
func extractNode(m any) (*corev3.Node, error) {
	switch m.(type) {
	case *discoveryv3.DeltaDiscoveryRequest, *discoveryv3.DiscoveryRequest:
		if node := requestNode(m); node != nil && node.GetId() != "" {
			return node, nil
		}
		return nil, status.Error(codes.InvalidArgument, "missing node ID")
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unexpected message type: %T", m)
	}
}
