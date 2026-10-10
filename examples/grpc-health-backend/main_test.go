// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func TestHealthCheckAuthority(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	srv := newServer()
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	tests := []struct {
		name      string
		authority string
		wantCode  codes.Code
	}{
		{
			name:      "valid authority",
			authority: "grpc.example.com",
			wantCode:  codes.OK,
		},
		{
			name:      "cluster name used as authority",
			authority: "grpcroute/gateway-conformance-infra/grpc-route/rule/0",
			wantCode:  codes.InvalidArgument,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn, err := grpc.NewClient("passthrough:///bufnet",
				grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
					return lis.DialContext(ctx)
				}),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithAuthority(tc.authority),
			)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}
			defer conn.Close()

			resp, err := healthpb.NewHealthClient(conn).Check(t.Context(), &healthpb.HealthCheckRequest{})
			if got := status.Code(err); got != tc.wantCode {
				t.Fatalf("Check() code = %v, want %v (err: %v)", got, tc.wantCode, err)
			}
			if tc.wantCode == codes.OK && resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
				t.Errorf("Check() status = %v, want SERVING", resp.GetStatus())
			}
		})
	}
}

func TestValidAuthority(t *testing.T) {
	tests := []struct {
		authority string
		want      bool
	}{
		{authority: "grpc.example.com", want: true},
		{authority: "grpc.example.com:8080", want: true},
		{authority: "grpc-health-backend.gateway-conformance-infra.svc.cluster.local", want: true},
		{authority: "10.0.0.1:9000", want: true},
		{authority: "[::1]:9000", want: true},
		{authority: "", want: false},
		{authority: "grpcroute/gateway-conformance-infra/grpc-route/rule/0", want: false},
		{authority: "grpc.example.com/path", want: false},
		{authority: "grpc.example.com?query", want: false},
		{authority: "user@grpc.example.com", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.authority, func(t *testing.T) {
			if got := validAuthority(tc.authority); got != tc.want {
				t.Errorf("validAuthority(%q) = %v, want %v", tc.authority, got, tc.want)
			}
		})
	}
}
