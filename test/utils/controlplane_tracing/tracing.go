// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package controlplanetracing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

// getTempoHost returns the Tempo host address
// It tries to use the LoadBalancer IP first, then falls back to localhost (port-forward)
func getTempoHost(t *testing.T, c client.Client) (string, error) {
	// Verify Tempo service exists
	svc := corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "monitoring",
		Name:      "tempo",
	}, &svc); err != nil {
		// Fall back to eg-addons prefix if that fails
		if err := c.Get(context.Background(), types.NamespacedName{
			Namespace: "monitoring",
			Name:      "eg-addons-tempo",
		}, &svc); err != nil {
			return "", fmt.Errorf("failed to get tempo service: %w", err)
		}
	}

	// Try to use LoadBalancer IP if available (more reliable than port-forward)
	if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			if svc.Status.LoadBalancer.Ingress[0].IP != "" {
				host := svc.Status.LoadBalancer.Ingress[0].IP
				tlog.Logf(t, "using Tempo at %s:3100 (via LoadBalancer)", host)
				return host, nil
			}
		}
	}

	return "", fmt.Errorf("tempo loadbalancer IP not found")
}

// QueryControlPlaneTraces queries Tempo for control plane traces with the given service name
func QueryControlPlaneTraces(t *testing.T, c client.Client, serviceName string) (int, error) {
	host, err := getTempoHost(t, c)
	if err != nil {
		return -1, err
	}

	tempoURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, "3100"),
		Path:   "/api/search",
	}
	query := tempoURL.Query()
	query.Add("start", fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix()))
	query.Add("end", fmt.Sprintf("%d", time.Now().Unix()))
	query.Add("tags", fmt.Sprintf("service.name=%s", serviceName))
	tempoURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", tempoURL.String(), nil)
	if err != nil {
		return -1, err
	}

	tlog.Logf(t, "querying tempo: %s", tempoURL.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("tempo returned status %s", res.Status)
	}

	resp := &TempoResponse{}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		tlog.Logf(t, "failed to unmarshal response: %s", string(data))
		return -1, err
	}

	total := len(resp.Traces)
	tlog.Logf(t, "found %d traces from tempo for service %s", total, serviceName)
	return total, nil
}

// VerifyExpectedSpans checks that expected span names exist in the traces
// Note: Tempo's search API doesn't index child span names, so we need to fetch
// traces and inspect them directly
func VerifyExpectedSpans(t *testing.T, c client.Client, serviceName string, expectedSpanNames []string) (bool, error) {
	host, err := getTempoHost(t, c)
	if err != nil {
		return false, err
	}

	// First, get all traces for the service
	searchURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, "3100"),
		Path:   "/api/search",
	}
	query := searchURL.Query()
	query.Add("start", fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix()))
	query.Add("end", fmt.Sprintf("%d", time.Now().Unix()))
	query.Add("tags", fmt.Sprintf("service.name=%s", serviceName))
	searchURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", searchURL.String(), nil)
	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("tempo search returned status %s", res.Status)
	}

	searchResp := &TempoResponse{}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, &searchResp); err != nil {
		return false, err
	}

	if len(searchResp.Traces) == 0 {
		tlog.Logf(t, "no traces found for service %s", serviceName)
		return false, nil
	}

	// Now fetch each trace and check for the expected span names
	foundSpans := make(map[string]bool)
	for _, trace := range searchResp.Traces {
		traceID := trace["traceID"].(string)
		traceURL := url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(host, "3100"),
			Path:   fmt.Sprintf("/api/traces/%s", traceID),
		}

		req, err := http.NewRequest("GET", traceURL.String(), nil)
		if err != nil {
			continue
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}

		if res.StatusCode != http.StatusOK {
			res.Body.Close()
			continue
		}

		traceData, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			continue
		}

		// Parse the trace to find span names
		var traceResp map[string]interface{}
		if err := json.Unmarshal(traceData, &traceResp); err != nil {
			continue
		}

		// Extract span names from the trace
		if batches, ok := traceResp["batches"].([]interface{}); ok {
			for _, batch := range batches {
				if batchMap, ok := batch.(map[string]interface{}); ok {
					if scopeSpans, ok := batchMap["scopeSpans"].([]interface{}); ok {
						for _, scopeSpan := range scopeSpans {
							if scopeSpanMap, ok := scopeSpan.(map[string]interface{}); ok {
								if spans, ok := scopeSpanMap["spans"].([]interface{}); ok {
									for _, span := range spans {
										if spanMap, ok := span.(map[string]interface{}); ok {
											if name, ok := spanMap["name"].(string); ok {
												foundSpans[name] = true
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Check if all expected spans were found
	for _, expectedSpan := range expectedSpanNames {
		if !foundSpans[expectedSpan] {
			tlog.Logf(t, "span '%s' not found yet", expectedSpan)
			return false, nil
		}
		tlog.Logf(t, "found span '%s' in traces", expectedSpan)
	}

	return true, nil
}

// TempoResponse represents the response from Tempo's search API
type TempoResponse struct {
	Traces []map[string]interface{} `json:"traces,omitempty"`
}
