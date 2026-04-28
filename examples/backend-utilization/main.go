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
	"net/http"
	"os"
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

func main() {
	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("NAMESPACE")
	serviceName := os.Getenv("SERVICE_NAME")
	cpuUtil := os.Getenv("ORCA_CPU_UTILIZATION")
	if cpuUtil == "" {
		cpuUtil = "0.0"
	}

	orcaHeader := fmt.Sprintf(`JSON {"cpu_utilization": %s, "rps_fractional": 1000}`, cpuUtil)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("endpoint-load-metrics", orcaHeader)

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

	log.Printf("Starting ORCA backend on :3000 (cpu_utilization=%s)", cpuUtil)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
