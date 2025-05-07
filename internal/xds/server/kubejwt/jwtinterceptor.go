// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubejwt

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"k8s.io/client-go/kubernetes"

	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/cache"
)

// JWTAuthInterceptor verifies Kubernetes Service Account JWT tokens in gRPC requests.
type JWTAuthInterceptor struct {
	clientset *kubernetes.Clientset
	issuer    string
	cache     cache.SnapshotCacheWithCallbacks
	xds       *message.Xds
}

// NewJWTAuthInterceptor initializes a new JWTAuthInterceptor.
func NewJWTAuthInterceptor(clientset *kubernetes.Clientset, issuer string, cache cache.SnapshotCacheWithCallbacks, xds *message.Xds) *JWTAuthInterceptor {
	return &JWTAuthInterceptor{
		clientset: clientset,
		issuer:    issuer,
		cache:     cache,
		xds:       xds,
	}
}

// Stream intercepts streaming gRPC calls for authentication.
func (i *JWTAuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := i.authorize(ss.Context()); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

// authorize validates the Kubernetes Service Account JWT token from the metadata.
func (i *JWTAuthInterceptor) authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing metadata")
	}

	authHeader, exists := md["authorization"]
	if !exists || len(authHeader) == 0 {
		return fmt.Errorf("missing authorization token in metadata: %s", md)
	}

	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
	err := i.validateKubeJWT(ctx, tokenStr)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	return nil
}
