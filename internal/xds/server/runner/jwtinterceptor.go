// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// JWTClaims defines the expected claims in the JWT token.
type JWTClaims struct {
	jwt.RegisteredClaims
}

// JWTAuthInterceptor verifies JWT tokens in gRPC requests.
type JWTAuthInterceptor struct{}

// NewJWTAuthInterceptor initializes a new JWTAuthInterceptor.
func NewJWTAuthInterceptor() *JWTAuthInterceptor {
	return &JWTAuthInterceptor{}
}

// Unary intercepts unary gRPC calls for authentication.
func (i *JWTAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := i.authorize(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// authorize validates the JWT token from the metadata.
func (i *JWTAuthInterceptor) authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing metadata")
	}

	authHeader, exists := md["authorization"]
	if !exists || len(authHeader) == 0 {
		return fmt.Errorf("missing authorization token")
	}

	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})

	if err != nil || !token.Valid {
		return fmt.Errorf("invalid or expired token")
	}

	return nil
}
