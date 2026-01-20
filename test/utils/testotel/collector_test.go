// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package testotel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectmetricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collecttracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestGRPCCollector_InjectsMetadataAsLogAttributes(t *testing.T) {
	collector, err := StartGRPCCollector()
	require.NoError(t, err)
	defer collector.Close()

	conn, err := grpc.NewClient(
		collector.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	logsClient := collectlogsv1.NewLogsServiceClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer log-api-key",
		"x-tenant-id", "tenant-123",
	))

	_, err = logsClient.Export(ctx, &collectlogsv1.ExportLogsServiceRequest{
		ResourceLogs: []*logsv1.ResourceLogs{
			{
				ScopeLogs: []*logsv1.ScopeLogs{
					{
						LogRecords: []*logsv1.LogRecord{
							{},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	log := collector.TakeLog()
	require.NotNil(t, log)
	require.Equal(t, "Bearer log-api-key", GetAttributeString(log.Attributes, "grpc.metadata.authorization"))
	require.Equal(t, "tenant-123", GetAttributeString(log.Attributes, "grpc.metadata.x-tenant-id"))
}

func TestGRPCCollector_InjectsMetadataAsSpanAttributes(t *testing.T) {
	collector, err := StartGRPCCollector()
	require.NoError(t, err)
	defer collector.Close()

	conn, err := grpc.NewClient(
		collector.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	traceClient := collecttracev1.NewTraceServiceClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer trace-api-key",
	))

	_, err = traceClient.Export(ctx, &collecttracev1.ExportTraceServiceRequest{
		ResourceSpans: []*tracev1.ResourceSpans{
			{
				ScopeSpans: []*tracev1.ScopeSpans{
					{
						Spans: []*tracev1.Span{
							{Name: "test-span"},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	span := collector.TakeSpan()
	require.NotNil(t, span)
	require.Equal(t, "Bearer trace-api-key", GetAttributeString(span.Attributes, "grpc.metadata.authorization"))
}

func TestGRPCCollector_InjectsMetadataAsResourceAttributes(t *testing.T) {
	collector, err := StartGRPCCollector()
	require.NoError(t, err)
	defer collector.Close()

	conn, err := grpc.NewClient(
		collector.Address(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	metricsClient := collectmetricsv1.NewMetricsServiceClient(conn)
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer metrics-api-key",
	))

	_, err = metricsClient.Export(ctx, &collectmetricsv1.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricsv1.ResourceMetrics{
			{
				ScopeMetrics: []*metricsv1.ScopeMetrics{
					{
						Metrics: []*metricsv1.Metric{
							{Name: "test-metric"},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	resourceMetrics := collector.TakeMetric()
	require.NotNil(t, resourceMetrics)
	require.NotNil(t, resourceMetrics.Resource)
	require.Equal(t, "Bearer metrics-api-key", GetAttributeString(resourceMetrics.Resource.Attributes, "grpc.metadata.authorization"))
}
