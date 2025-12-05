// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package testotel provides test utilities for OpenTelemetry access log tests.
// It implements a gRPC-based OTLP collector that captures metadata headers for testing.
package testotel

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// otlpTimeout is the timeout for reading logs.
const otlpTimeout = 5 * time.Second

// ReceivedMetadata represents metadata captured from a gRPC request.
type ReceivedMetadata struct {
	// Headers contains all metadata headers received with the request.
	Headers map[string][]string
	// Timestamp when the request was received.
	Timestamp time.Time
}

// GRPCCollector is a test OTLP gRPC collector that captures logs and metadata headers.
type GRPCCollector struct {
	server   *grpc.Server
	listener net.Listener
	port     int

	logCh chan *logsv1.ResourceLogs

	mu               sync.Mutex
	receivedMetadata []ReceivedMetadata
}

// logsServer handles logs export requests.
type logsServer struct {
	collectlogsv1.UnimplementedLogsServiceServer
	collector *GRPCCollector
}

// StartGRPCCollector starts a test OTLP gRPC collector on an available port.
func StartGRPCCollector() (*GRPCCollector, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	c := &GRPCCollector{
		listener:         listener,
		port:             listener.Addr().(*net.TCPAddr).Port,
		logCh:            make(chan *logsv1.ResourceLogs, 100),
		receivedMetadata: make([]ReceivedMetadata, 0),
	}

	c.server = grpc.NewServer(
		grpc.UnaryInterceptor(c.metadataInterceptor),
	)

	collectlogsv1.RegisterLogsServiceServer(c.server, &logsServer{collector: c})

	go func() {
		_ = c.server.Serve(listener)
	}()

	return c, nil
}

// metadataInterceptor captures gRPC metadata from incoming requests.
func (c *GRPCCollector) metadataInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		c.mu.Lock()
		c.receivedMetadata = append(c.receivedMetadata, ReceivedMetadata{
			Headers:   md,
			Timestamp: time.Now(),
		})
		c.mu.Unlock()
	}
	return handler(ctx, req)
}

// Export implements the LogsService Export RPC.
func (s *logsServer) Export(_ context.Context, req *collectlogsv1.ExportLogsServiceRequest) (*collectlogsv1.ExportLogsServiceResponse, error) {
	for _, resourceLogs := range req.ResourceLogs {
		select {
		case s.collector.logCh <- resourceLogs:
		case <-time.After(otlpTimeout):
			fmt.Println("Warning: Dropping logs due to timeout")
		}
	}
	return &collectlogsv1.ExportLogsServiceResponse{}, nil
}

// Port returns the port the collector is listening on.
func (c *GRPCCollector) Port() int {
	return c.port
}

// Address returns the full address (host:port) the collector is listening on.
func (c *GRPCCollector) Address() string {
	return fmt.Sprintf("127.0.0.1:%d", c.port)
}

// TakeLog returns a single log record or nil if none were recorded within the timeout.
func (c *GRPCCollector) TakeLog() *logsv1.LogRecord {
	select {
	case resourceLogs := <-c.logCh:
		if len(resourceLogs.ScopeLogs) == 0 || len(resourceLogs.ScopeLogs[0].LogRecords) == 0 {
			return nil
		}
		return resourceLogs.ScopeLogs[0].LogRecords[0]
	case <-time.After(otlpTimeout):
		return nil
	}
}

// GetReceivedMetadata returns a copy of all received metadata.
func (c *GRPCCollector) GetReceivedMetadata() []ReceivedMetadata {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]ReceivedMetadata, len(c.receivedMetadata))
	copy(result, c.receivedMetadata)
	return result
}

// AssertHeaderReceived checks if a specific header with the expected value was received.
// Returns true if the header was found with the expected value in any request.
func (c *GRPCCollector) AssertHeaderReceived(headerName, expectedValue string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, rm := range c.receivedMetadata {
		if values, ok := rm.Headers[headerName]; ok {
			for _, v := range values {
				if v == expectedValue {
					return true
				}
			}
		}
	}
	return false
}

// WaitForHeader waits for a specific header to be received within the timeout.
// Returns true if the header was found, false if timeout occurred.
func (c *GRPCCollector) WaitForHeader(headerName, expectedValue string, timeout time.Duration) bool {
	deadline := time.After(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return false
		case <-ticker.C:
			if c.AssertHeaderReceived(headerName, expectedValue) {
				return true
			}
		}
	}
}

// ClearMetadata clears all received metadata.
func (c *GRPCCollector) ClearMetadata() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.receivedMetadata = make([]ReceivedMetadata, 0)
}

// Close shuts down the collector and cleans up resources.
func (c *GRPCCollector) Close() {
	c.server.GracefulStop()
	c.listener.Close()
	close(c.logCh)
}
