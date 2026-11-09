// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// grpc-health-backend is a plaintext gRPC server implementing the
// grpc.health.v1 Health service. Like gRPC servers built on recent versions of
// golang.org/x/net, it rejects requests whose :authority header is not a valid
// RFC 3986 authority, so it can be used to verify that active gRPC health
// checks send a well-formed authority.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 9000, "gRPC port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", port, err)
	}

	log.Printf("listening on %s", lis.Addr())
	if err := newServer().Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func newServer() *grpc.Server {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			if err := checkAuthority(ctx); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}),
		grpc.StreamInterceptor(func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			if err := checkAuthority(ss.Context()); err != nil {
				return err
			}
			return handler(srv, ss)
		}),
	)
	// The health server reports the overall server as SERVING by default.
	healthpb.RegisterHealthServer(srv, health.NewServer())
	return srv
}

// checkAuthority rejects requests whose :authority header is not a valid
// authority, such as the Envoy cluster name (grpcroute/ns/name/rule/0) that
// Envoy falls back to when a gRPC health check has no authority configured.
func checkAuthority(ctx context.Context) error {
	var authority string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if values := md.Get(":authority"); len(values) > 0 {
			authority = values[0]
		}
	}
	if !validAuthority(authority) {
		log.Printf("rejected request with malformed :authority %q", authority)
		return status.Errorf(codes.InvalidArgument, "malformed :authority %q", authority)
	}
	return nil
}

// validAuthority reports whether authority is made of a host and an optional
// port only. Anything url.Parse does not consider part of the host, such as a
// path, a query or user info, makes the authority invalid.
func validAuthority(authority string) bool {
	if authority == "" {
		return false
	}
	u, err := url.Parse("//" + authority)
	return err == nil && u.User == nil && u.Host == authority
}
