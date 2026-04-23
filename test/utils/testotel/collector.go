// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package testotel provides test utilities for OpenTelemetry tests.
package testotel

import (
	"context"
	"fmt"
	"net"
	"time"

	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectmetricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collecttracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const otlpTimeout = 5 * time.Second

// GRPCCollector is a test OTLP gRPC collector that captures logs, traces, and metrics.
// gRPC metadata is injected as attributes on each signal for easy assertion.
type GRPCCollector struct {
	server   *grpc.Server
	listener net.Listener
	port     int

	logCh     chan *logsv1.ResourceLogs
	traceCh   chan *tracev1.ResourceSpans
	metricsCh chan *metricsv1.ResourceMetrics
}

type logsServer struct {
	collectlogsv1.UnimplementedLogsServiceServer
	collector *GRPCCollector
}

type traceServer struct {
	collecttracev1.UnimplementedTraceServiceServer
	collector *GRPCCollector
}

type metricsServer struct {
	collectmetricsv1.UnimplementedMetricsServiceServer
	collector *GRPCCollector
}

// StartGRPCCollector starts a test OTLP gRPC collector on an available port.
func StartGRPCCollector() (*GRPCCollector, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	c := &GRPCCollector{
		listener:  listener,
		port:      listener.Addr().(*net.TCPAddr).Port,
		logCh:     make(chan *logsv1.ResourceLogs, 100),
		traceCh:   make(chan *tracev1.ResourceSpans, 100),
		metricsCh: make(chan *metricsv1.ResourceMetrics, 100),
	}

	c.server = grpc.NewServer()

	collectlogsv1.RegisterLogsServiceServer(c.server, &logsServer{collector: c})
	collecttracev1.RegisterTraceServiceServer(c.server, &traceServer{collector: c})
	collectmetricsv1.RegisterMetricsServiceServer(c.server, &metricsServer{collector: c})

	go func() {
		_ = c.server.Serve(listener)
	}()

	return c, nil
}

// Export implements the LogsService Export RPC.
// gRPC metadata is injected as attributes on each log record.
func (s *logsServer) Export(ctx context.Context, req *collectlogsv1.ExportLogsServiceRequest) (*collectlogsv1.ExportLogsServiceResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	attrs := metadataToAttributes(md)
	for _, resourceLogs := range req.ResourceLogs {
		for _, scopeLogs := range resourceLogs.ScopeLogs {
			for _, log := range scopeLogs.LogRecords {
				log.Attributes = append(log.Attributes, attrs...)
			}
		}
		select {
		case s.collector.logCh <- resourceLogs:
		case <-time.After(otlpTimeout):
			fmt.Println("Warning: Dropping logs due to timeout")
		}
	}
	return &collectlogsv1.ExportLogsServiceResponse{}, nil
}

// Export implements the TraceService Export RPC.
// gRPC metadata is injected as attributes on each span.
func (s *traceServer) Export(ctx context.Context, req *collecttracev1.ExportTraceServiceRequest) (*collecttracev1.ExportTraceServiceResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	attrs := metadataToAttributes(md)
	for _, resourceSpans := range req.ResourceSpans {
		for _, scopeSpans := range resourceSpans.ScopeSpans {
			for _, span := range scopeSpans.Spans {
				span.Attributes = append(span.Attributes, attrs...)
			}
		}
		select {
		case s.collector.traceCh <- resourceSpans:
		case <-time.After(otlpTimeout):
			fmt.Println("Warning: Dropping spans due to timeout")
		}
	}
	return &collecttracev1.ExportTraceServiceResponse{}, nil
}

// Export implements the MetricsService Export RPC.
// gRPC metadata is injected as resource attributes.
func (s *metricsServer) Export(ctx context.Context, req *collectmetricsv1.ExportMetricsServiceRequest) (*collectmetricsv1.ExportMetricsServiceResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	attrs := metadataToAttributes(md)
	for _, resourceMetrics := range req.ResourceMetrics {
		if resourceMetrics.Resource == nil {
			resourceMetrics.Resource = &resourcev1.Resource{}
		}
		resourceMetrics.Resource.Attributes = append(resourceMetrics.Resource.Attributes, attrs...)
		select {
		case s.collector.metricsCh <- resourceMetrics:
		case <-time.After(otlpTimeout):
			fmt.Println("Warning: Dropping metrics due to timeout")
		}
	}
	return &collectmetricsv1.ExportMetricsServiceResponse{}, nil
}

// metadataToAttributes converts gRPC metadata to OTel attributes with "grpc.metadata." prefix.
func metadataToAttributes(md metadata.MD) []*commonv1.KeyValue {
	var attrs []*commonv1.KeyValue
	for key, values := range md {
		if len(values) > 0 {
			attrs = append(attrs, &commonv1.KeyValue{
				Key:   "grpc.metadata." + key,
				Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: values[0]}},
			})
		}
	}
	return attrs
}

// GetAttributeString returns the string value for the given key, or empty string if not found.
func GetAttributeString(attrs []*commonv1.KeyValue, key string) string {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Value.GetStringValue()
		}
	}
	return ""
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

// TakeResourceLogs returns resource logs or nil if none were recorded within the timeout.
func (c *GRPCCollector) TakeResourceLogs() *logsv1.ResourceLogs {
	select {
	case resourceLogs := <-c.logCh:
		return resourceLogs
	case <-time.After(otlpTimeout):
		return nil
	}
}

// TakeResourceSpans returns resource spans or nil if none were recorded within the timeout.
func (c *GRPCCollector) TakeResourceSpans() *tracev1.ResourceSpans {
	select {
	case resourceSpans := <-c.traceCh:
		return resourceSpans
	case <-time.After(otlpTimeout):
		return nil
	}
}

// GetResourceAttribute returns the string value for the given key from resource attributes.
func GetResourceAttribute(r *resourcev1.Resource, key string) string {
	if r == nil {
		return ""
	}
	return GetAttributeString(r.Attributes, key)
}

// TakeSpan returns a single span or nil if none were recorded within the timeout.
func (c *GRPCCollector) TakeSpan() *tracev1.Span {
	select {
	case resourceSpans := <-c.traceCh:
		if len(resourceSpans.ScopeSpans) == 0 || len(resourceSpans.ScopeSpans[0].Spans) == 0 {
			return nil
		}
		return resourceSpans.ScopeSpans[0].Spans[0]
	case <-time.After(otlpTimeout):
		return nil
	}
}

// TakeMetric returns resource metrics or nil if none were recorded within the timeout.
func (c *GRPCCollector) TakeMetric() *metricsv1.ResourceMetrics {
	select {
	case resourceMetrics := <-c.metricsCh:
		return resourceMetrics
	case <-time.After(otlpTimeout):
		return nil
	}
}

// Close shuts down the collector and cleans up resources.
func (c *GRPCCollector) Close() {
	c.server.GracefulStop()
	c.listener.Close()
	close(c.logCh)
	close(c.traceCh)
	close(c.metricsCh)
}
