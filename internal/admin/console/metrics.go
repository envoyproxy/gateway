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

// createCombinedMetricsHandler creates a handler that combines the controller-runtime and
// default Prometheus registries.
func createCombinedMetricsHandler() http.Handler {
	gatherer := prometheus.Gatherers{
		metricsserver.Registry,
		prometheus.DefaultGatherer,
	}
	return promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{})
}
