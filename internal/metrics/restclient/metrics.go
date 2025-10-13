// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package restclient

import (
	"context"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	clientmetrics "k8s.io/client-go/tools/metrics"
)

var (
	// requestLatency is a Prometheus Histogram metric type partitioned by
	// "verb", and "host" labels. It is used for the rest client latency metrics.
	requestLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rest_client_request_duration_seconds",
			Help:    "Request latency in seconds. Broken down by verb, and host.",
			Buckets: []float64{0.005, 0.025, 0.1, 0.25, 0.5, 1.0, 2.0, 4.0, 8.0, 15.0, 30.0, 60.0},
		},
		[]string{"verb", "host"},
	)

	requestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "rest_client_request_size_bytes",
			Help: "Request size in bytes. Broken down by verb and host.",
			// 64 bytes to 16MB
			Buckets: []float64{64, 256, 512, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216},
		},
		[]string{"verb", "host"},
	)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "rest_client_response_size_bytes",
			Help: "Response size in bytes. Broken down by verb and host.",
			// 64 bytes to 16MB
			Buckets: []float64{64, 256, 512, 1024, 4096, 16384, 65536, 262144, 1048576, 4194304, 16777216},
		},
		[]string{"verb", "host"},
	)

	// RateLimiterLatency is the client side rate limiter latency metric.
	rateLimiterLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rest_client_rate_limiter_duration_seconds",
			Help:    "Rate limiter latency in seconds. Broken down by verb, and host.",
			Buckets: []float64{0.005, 0.025, 0.1, 0.25, 0.5, 1.0, 2.0, 4.0, 8.0, 15.0, 30.0, 60.0},
		},
		[]string{"verb", "host"},
	)

	// RequestRetry is the retry metric that tracks the number of
	// retries sent to the server.
	requestRetry = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rest_client_requests_retry_total",
			Help: "Number of HTTP requests retry, partitioned by status code, method, and host.",
		},
		[]string{"code", "method", "host"},
	)
)

// RegisterClientMetrics for controller-runtime sets up the client latency metrics from client-go.
func RegisterClientMetricsWithoutRequestTotal(registry prometheus.Registerer) {
	// register the metrics with our registry
	registry.MustRegister(requestLatency)
	registry.MustRegister(requestSize)
	registry.MustRegister(responseSize)
	registry.MustRegister(rateLimiterLatency)
	registry.MustRegister(requestRetry)

	// register the metrics with client-go
	clientmetrics.RequestLatency = &latencyAdapter{metric: requestLatency}
	clientmetrics.RequestSize = &sizeAdapter{metric: requestSize}
	clientmetrics.ResponseSize = &sizeAdapter{metric: responseSize}
	clientmetrics.RateLimiterLatency = &latencyAdapter{metric: rateLimiterLatency}
	clientmetrics.RequestRetry = &retryAdapter{metric: requestRetry}
}

// this section contains adapters, implementations, and other sundry organic, artisanally
// hand-crafted syntax trees required to convince client-go that it actually wants to let
// someone use its metrics.

// Client metrics adapters (method #1 for client-go metrics),
// copied (more-or-less directly) from k8s.io/kubernetes setup code
// (which isn't anywhere in an easily-importable place).

// latencyAdapter implements LatencyMetric.
type latencyAdapter struct {
	metric *prometheus.HistogramVec
}

// Observe increments the request latency metric for the given verb/URL.
//
//nolint:gocritic // client-go's LatencyMetric interface requires url.URL by value.
func (l *latencyAdapter) Observe(_ context.Context, verb string, u url.URL, latency time.Duration) {
	l.metric.WithLabelValues(verb, u.String()).Observe(latency.Seconds())
}

type sizeAdapter struct {
	metric *prometheus.HistogramVec
}

func (s *sizeAdapter) Observe(_ context.Context, verb, host string, size float64) {
	s.metric.WithLabelValues(verb, host).Observe(size)
}

type ResultAdapter struct {
	metric *prometheus.CounterVec
}

func (r *ResultAdapter) Increment(_ context.Context, code, method, host string) {
	r.metric.WithLabelValues(code, method, host).Inc()
}

type retryAdapter struct {
	metric *prometheus.CounterVec
}

func (r *retryAdapter) IncrementRetry(_ context.Context, code, method, host string) {
	r.metric.WithLabelValues(code, method, host).Inc()
}
