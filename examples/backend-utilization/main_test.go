// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"context"
	"net"
	"testing"
	"time"

	orcaservice "github.com/cncf/xds/go/xds/service/orca/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TestStreamCoreMetrics exercises the same OpenRcaService stream that Envoy opens
// for out-of-band ORCA reporting, so the OOB path can be checked without a cluster.
func TestStreamCoreMetrics(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := newORCAServer(0.9)
	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("server stopped: %v", err)
		}
	}()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	stream, err := orcaservice.NewOpenRcaServiceClient(conn).StreamCoreMetrics(ctx,
		&orcaservice.OrcaLoadReportRequest{ReportInterval: durationpb.New(minReportInterval)})
	if err != nil {
		t.Fatalf("StreamCoreMetrics: %v", err)
	}

	// One report is sent immediately, the rest on the requested interval.
	for i := range 3 {
		report, err := stream.Recv()
		if err != nil {
			t.Fatalf("recv report %d: %v", i, err)
		}
		if got := report.GetCpuUtilization(); got != 0.9 {
			t.Errorf("report %d: cpu_utilization = %v, want 0.9", i, got)
		}
		if got := report.GetRpsFractional(); got != 1000 {
			t.Errorf("report %d: rps_fractional = %v, want 1000", i, got)
		}
	}
}
