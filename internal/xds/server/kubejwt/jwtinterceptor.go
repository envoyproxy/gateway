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

	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	"github.com/envoyproxy/gateway/internal/xds/cache"
)

// JWTAuthInterceptor verifies Kubernetes Service Account JWT tokens in gRPC requests.
type JWTAuthInterceptor struct {
	clientset *kubernetes.Clientset
	issuer    string
	cache     cache.SnapshotCacheWithCallbacks
}

// NewJWTAuthInterceptor initializes a new JWTAuthInterceptor.
func NewJWTAuthInterceptor(clientset *kubernetes.Clientset, issuer string, cache cache.SnapshotCacheWithCallbacks) *JWTAuthInterceptor {
	return &JWTAuthInterceptor{
		clientset: clientset,
		issuer:    issuer,
		cache:     cache,
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

	proxyMetadata, err := processProxyMetadata(md)
	if err != nil {
		return fmt.Errorf("failed to extract node info: %w", err)
	}

	err = i.validateKubeJWT(ctx, proxyMetadata)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	return nil
}

type proxyMetadata struct {
	token  string
	nodeID string
	irKey  string
}

func processProxyMetadata(md metadata.MD) (*proxyMetadata, error) {
	authHeader, exists := md["authorization"]
	if !exists || len(authHeader) == 0 {
		return nil, fmt.Errorf("missing authorization token in metadata: %s", md)
	}
	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")

	irKey, exists := md[bootstrap.EnvoyIrKeyHeader]
	if !exists || len(irKey) == 0 {
		return nil, fmt.Errorf("missing ir key in metadata: %s", md)
	}

	nodeID, exists := md[bootstrap.EnvoyNodeIDHeader]
	if !exists || len(nodeID) == 0 {
		return nil, fmt.Errorf("missing node ID in metadata: %s", md)
	}

	return &proxyMetadata{
		token:  tokenStr,
		nodeID: nodeID[0],
		irKey:  irKey[0],
	}, nil
}
