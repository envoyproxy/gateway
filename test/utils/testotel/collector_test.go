// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package testotel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	collectlogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestGRPCCollector_CapturesLogsHeaders(t *testing.T) {
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

	require.True(t, collector.AssertHeaderReceived("authorization", "Bearer log-api-key"))
	require.True(t, collector.AssertHeaderReceived("x-tenant-id", "tenant-123"))

	log := collector.TakeLog()
	require.NotNil(t, log)
}

func TestGRPCCollector_WaitForHeader(t *testing.T) {
	collector, err := StartGRPCCollector()
	require.NoError(t, err)
	defer collector.Close()

	// Should timeout for nonexistent header
	found := collector.WaitForHeader("nonexistent", "value", 100*time.Millisecond)
	require.False(t, found)

	// Send a request asynchronously
	go func() {
		time.Sleep(50 * time.Millisecond)

		conn, err := grpc.NewClient(
			collector.Address(),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return
		}
		defer conn.Close()

		logsClient := collectlogsv1.NewLogsServiceClient(conn)
		ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
			"x-async-header", "async-value",
		))
		_, _ = logsClient.Export(ctx, &collectlogsv1.ExportLogsServiceRequest{})
	}()

	found = collector.WaitForHeader("x-async-header", "async-value", 2*time.Second)
	require.True(t, found)
}

func TestGRPCCollector_ClearMetadata(t *testing.T) {
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
		"x-test", "value",
	))
	_, err = logsClient.Export(ctx, &collectlogsv1.ExportLogsServiceRequest{})
	require.NoError(t, err)

	require.True(t, collector.AssertHeaderReceived("x-test", "value"))

	collector.ClearMetadata()

	require.False(t, collector.AssertHeaderReceived("x-test", "value"))
	require.Empty(t, collector.GetReceivedMetadata())
}
