// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/rsa"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// JWTClaims defines the expected claims in the Kubernetes Service Account JWT token.
type JWTClaims struct {
	jwt.RegisteredClaims
	Namespace string `json:"kubernetes.io/serviceaccount/namespace"`
	PodName   string `json:"kubernetes.io/serviceaccount/pod.name"`
	PodUID    string `json:"kubernetes.io/serviceaccount/pod.uid"`
}

// JWTAuthInterceptor verifies Kubernetes Service Account JWT tokens in gRPC requests.
type JWTAuthInterceptor struct {
	publicKey *rsa.PublicKey
	issuer    string
}

// NewJWTAuthInterceptor initializes a new JWTAuthInterceptor.
func NewJWTAuthInterceptor(publicKey *rsa.PublicKey, issuer string) *JWTAuthInterceptor {
	return &JWTAuthInterceptor{
		publicKey: publicKey,
		issuer:    issuer,
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
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return i.publicKey, nil
	})
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("invalid or expired token")
	}

	// Validate claims
	if claims.Issuer != i.issuer {
		return fmt.Errorf("unexpected issuer: %s", claims.Issuer)
	}

	return nil
}
