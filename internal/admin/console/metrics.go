// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package console

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// createCombinedMetricsHandler creates a handler that uses the controller-runtime registry
func createCombinedMetricsHandler() http.Handler {
	// Check if the controller-runtime registry has any metrics
	// In test environments, it might be empty
	if hasMetrics(metricsserver.Registry) {
		// Use the controller-runtime registry which includes all Envoy Gateway metrics
		return promhttp.HandlerFor(metricsserver.Registry, promhttp.HandlerOpts{})
	}

	// Fallback to default gatherer for test environments
	return promhttp.Handler()
}

// hasMetrics checks if a gatherer has any metrics
func hasMetrics(gatherer prometheus.Gatherer) bool {
	families, err := gatherer.Gather()
	return err == nil && len(families) > 0
}
