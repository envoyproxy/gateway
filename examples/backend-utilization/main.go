// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	orcadata "github.com/cncf/xds/go/xds/data/orca/v3"
	orcaservice "github.com/cncf/xds/go/xds/service/orca/v3"
	"google.golang.org/grpc"
)

const (
	// Interval used when Envoy does not ask for a specific report_interval on the
	// OOB stream.
	defaultReportInterval = 10 * time.Second
	// Floor on the requested interval, so a misconfigured reporting period cannot
	// turn the example into a busy loop.
	minReportInterval = 100 * time.Millisecond
)

type response struct {
	Path        string              `json:"path"`
	Host        string              `json:"host"`
	Method      string              `json:"method"`
	Protocol    string              `json:"proto"`
	Headers     map[string][]string `json:"headers"`
	Namespace   string              `json:"namespace"`
	Pod         string              `json:"pod"`
	ServiceName string              `json:"service_name"`
}

// orcaService implements xds.service.orca.v3.OpenRcaService, the out-of-band
// (OOB) side of ORCA reporting. Envoy opens one StreamCoreMetrics stream per
// endpoint and receives a report every report_interval, independent of request
// traffic.
type orcaService struct {
	orcaservice.UnimplementedOpenRcaServiceServer
	report *orcadata.OrcaLoadReport
}

func (s *orcaService) StreamCoreMetrics(req *orcaservice.OrcaLoadReportRequest, stream grpc.ServerStreamingServer[orcadata.OrcaLoadReport]) error {
	interval := defaultReportInterval
	if d := req.GetReportInterval().AsDuration(); d > 0 {
		interval = d
	}
	interval = max(interval, minReportInterval)

	log.Printf("OOB stream opened, reporting every %s", interval)
	defer log.Printf("OOB stream closed")

	// Report once up front so Envoy does not have to wait a full period for the
	// first sample.
	if err := stream.Send(s.report); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-ticker.C:
			if err := stream.Send(s.report); err != nil {
				return err
			}
		}
	}
}

func newORCAServer(cpuUtil float64) *grpc.Server {
	srv := grpc.NewServer()
	orcaservice.RegisterOpenRcaServiceServer(srv, &orcaService{
		report: &orcadata.OrcaLoadReport{
			CpuUtilization: cpuUtil,
			RpsFractional:  1000,
		},
	})
	return srv
}

func serveORCA(addr string, cpuUtil float64) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	log.Printf("Starting ORCA OpenRcaService on %s", addr)
	return newORCAServer(cpuUtil).Serve(lis)
}

func main() {
	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("NAMESPACE")
	serviceName := os.Getenv("SERVICE_NAME")

	cpuUtil := os.Getenv("ORCA_CPU_UTILIZATION")
	if cpuUtil == "" {
		cpuUtil = "0.0"
	}
	cpuUtilValue, err := strconv.ParseFloat(cpuUtil, 64)
	if err != nil {
		log.Fatalf("invalid ORCA_CPU_UTILIZATION %q: %v", cpuUtil, err)
	}

	// In-band reporting is on by default. Set ORCA_INBAND=false to report only
	// out-of-band, which is how the OOB path is exercised in isolation.
	inBand := true
	if v := os.Getenv("ORCA_INBAND"); v != "" {
		if inBand, err = strconv.ParseBool(v); err != nil {
			log.Fatalf("invalid ORCA_INBAND %q: %v", v, err)
		}
	}

	grpcPort := os.Getenv("ORCA_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "3001"
	}
	go func() {
		log.Fatal(serveORCA(net.JoinHostPort("", grpcPort), cpuUtilValue))
	}()

	orcaHeader := fmt.Sprintf(`JSON {"cpu_utilization": %s, "rps_fractional": 1000}`, cpuUtil)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if inBand {
			w.Header().Set("endpoint-load-metrics", orcaHeader)
		}

		headers := make(map[string][]string, len(r.Header))
		maps.Copy(headers, r.Header)

		resp := response{
			Path:        r.URL.Path,
			Host:        r.Host,
			Method:      r.Method,
			Protocol:    r.Proto,
			Headers:     headers,
			Namespace:   namespace,
			Pod:         podName,
			ServiceName: serviceName,
		}

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Printf("Starting ORCA backend on :3000 (cpu_utilization=%s, in-band=%t)", cpuUtil, inBand)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
