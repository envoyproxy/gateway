// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

type AppProtocol string

const (
	// GRPC declares that the port carries gRPC traffic.
	GRPC AppProtocol = "GRPC"
	// GRPCWeb declares that the port carries gRPC traffic.
	GRPCWeb AppProtocol = "GRPC-Web"
	// HTTP declares that the port carries HTTP/1.1 traffic.
	// Note that HTTP/1.0 or earlier may not be supported by the proxy.
	HTTP AppProtocol = "HTTP"
	// HTTP2 declares that the port carries HTTP/2 traffic.
	HTTP2 AppProtocol = "HTTP2"
	// HTTPS declares that the port carries HTTPS traffic.
	HTTPS AppProtocol = "HTTPS"
	// TCP declares the port uses TCP.
	// This is the default protocol for a service port.
	TCP AppProtocol = "TCP"
	// UDP declares that the port uses UDP.
	UDP AppProtocol = "UDP"
)
