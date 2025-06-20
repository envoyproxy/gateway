// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubejwt

import (
	"context"
	"fmt"
	"strings"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"k8s.io/client-go/kubernetes"

	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/xds/cache"
)

// JWTAuthInterceptor verifies Kubernetes Service Account JWT tokens in gRPC requests.
type JWTAuthInterceptor struct {
	clientset *kubernetes.Clientset
	issuer    string
	audience  string
	cache     cache.SnapshotCacheWithCallbacks
	logger    logging.Logger
}

// NewJWTAuthInterceptor initializes a new JWTAuthInterceptor.
func NewJWTAuthInterceptor(logger logging.Logger, clientset *kubernetes.Clientset, issuer, audience string, cache cache.SnapshotCacheWithCallbacks) *JWTAuthInterceptor {
	return &JWTAuthInterceptor{
		clientset: clientset,
		issuer:    issuer,
		audience:  audience,
		cache:     cache,
		logger:    logger.WithName("jwt-auth-interceptor"),
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx         context.Context
	interceptor *JWTAuthInterceptor
	validated   bool
}

func (w *wrappedStream) RecvMsg(m any) error {
	err := w.ServerStream.RecvMsg(m)
	if err != nil {
		return err
	}

	if !w.validated {
		if req, ok := m.(*discoveryv3.DeltaDiscoveryRequest); ok {
			if req.Node == nil || req.Node.Id == "" {
				return fmt.Errorf("missing node ID in request")
			}
			nodeID := req.Node.Id

			md, ok := metadata.FromIncomingContext(w.ctx)
			if !ok {
				return fmt.Errorf("missing metadata")
			}

			authHeader := md.Get("authorization")
			if len(authHeader) == 0 {
				return fmt.Errorf("missing authorization token in metadata: %s", md)
			}
			token := strings.TrimPrefix(authHeader[0], "Bearer ")

			if err := w.interceptor.validateKubeJWT(w.ctx, token, nodeID); err != nil {
				w.interceptor.logger.Error(err, "failed to validate token")
				return fmt.Errorf("failed to validate token: %w", err)
			}

			w.validated = true
		}
	}

	return nil
}

func newWrappedStream(s grpc.ServerStream, ctx context.Context, interceptor *JWTAuthInterceptor) grpc.ServerStream {
	return &wrappedStream{s, ctx, interceptor, false}
}

// Stream intercepts streaming gRPC calls for authorization.
func (i *JWTAuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapped := newWrappedStream(ss, ss.Context(), i)
		return handler(srv, wrapped)
	}
}
