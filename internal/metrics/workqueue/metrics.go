// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package workqueue

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// This file is copied and adapted from k8s.io/component-base/metrics/prometheus/workqueue
// which registers metrics to the k8s legacy Registry. We require very
// similar functionality, but must register metrics to a different Registry.

// Metrics subsystem and all keys used by the workqueue.
const (
	WorkQueueSubsystem         = metrics.WorkQueueSubsystem
	DepthKey                   = metrics.DepthKey
	AddsKey                    = metrics.AddsKey
	QueueLatencyKey            = metrics.QueueLatencyKey
	WorkDurationKey            = metrics.WorkDurationKey
	UnfinishedWorkKey          = metrics.UnfinishedWorkKey
	LongestRunningProcessorKey = metrics.LongestRunningProcessorKey
	RetriesKey                 = metrics.RetriesKey
)

var (
	depth = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      DepthKey,
		Help:      "Current depth of workqueue by workqueue and priority",
	}, []string{"name", "controller", "priority"})

	adds = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      AddsKey,
		Help:      "Total number of adds handled by workqueue",
	}, []string{"name", "controller"})

	latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:                       WorkQueueSubsystem,
		Name:                            QueueLatencyKey,
		Help:                            "How long in seconds an item stays in workqueue before being requested",
		Buckets:                         prometheus.ExponentialBuckets(10e-9, 10, 12),
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: 1 * time.Hour,
	}, []string{"name", "controller"})

	workDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Subsystem:                       WorkQueueSubsystem,
		Name:                            WorkDurationKey,
		Help:                            "How long in seconds processing an item from workqueue takes.",
		Buckets:                         prometheus.ExponentialBuckets(10e-9, 10, 12),
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: 1 * time.Hour,
	}, []string{"name", "controller"})

	unfinished = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      UnfinishedWorkKey,
		Help: "How many seconds of work has been done that " +
			"is in progress and hasn't been observed by work_duration. Large " +
			"values indicate stuck threads. One can deduce the number of stuck " +
			"threads by observing the rate at which this increases.",
	}, []string{"name", "controller"})

	longestRunningProcessor = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      LongestRunningProcessorKey,
		Help: "How many seconds has the longest running " +
			"processor for workqueue been running.",
	}, []string{"name", "controller"})

	retries = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: WorkQueueSubsystem,
		Name:      RetriesKey,
		Help:      "Total number of retries handled by workqueue",
	}, []string{"name", "controller"})
)

func RegisterMetrics(registry prometheus.Registerer) {
	registry.MustRegister(depth, adds, latency, workDuration, unfinished, longestRunningProcessor, retries)
}

type WorkqueueMetricsProvider struct{}

func (WorkqueueMetricsProvider) NewDepthMetric(name string) workqueue.GaugeMetric {
	return depth.WithLabelValues(name, name, "") // no priority
}

func (WorkqueueMetricsProvider) NewAddsMetric(name string) workqueue.CounterMetric {
	return adds.WithLabelValues(name, name)
}

func (WorkqueueMetricsProvider) NewLatencyMetric(name string) workqueue.HistogramMetric {
	return latency.WithLabelValues(name, name)
}

func (WorkqueueMetricsProvider) NewWorkDurationMetric(name string) workqueue.HistogramMetric {
	return workDuration.WithLabelValues(name, name)
}

func (WorkqueueMetricsProvider) NewUnfinishedWorkSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return unfinished.WithLabelValues(name, name)
}

func (WorkqueueMetricsProvider) NewLongestRunningProcessorSecondsMetric(name string) workqueue.SettableGaugeMetric {
	return longestRunningProcessor.WithLabelValues(name, name)
}

func (WorkqueueMetricsProvider) NewRetriesMetric(name string) workqueue.CounterMetric {
	return retries.WithLabelValues(name, name)
}

type MetricsProviderWithPriority interface {
	workqueue.MetricsProvider

	NewDepthMetricWithPriority(name string) DepthMetricWithPriority
}

// DepthMetricWithPriority represents a depth metric with priority.
type DepthMetricWithPriority interface {
	Inc(priority int)
	Dec(priority int)
}

var _ MetricsProviderWithPriority = WorkqueueMetricsProvider{}

func (WorkqueueMetricsProvider) NewDepthMetricWithPriority(name string) DepthMetricWithPriority {
	return &depthWithPriorityMetric{lvs: []string{name, name}}
}

type depthWithPriorityMetric struct {
	lvs []string
}

func (g *depthWithPriorityMetric) Inc(priority int) {
	depth.WithLabelValues(append(g.lvs, strconv.Itoa(priority))...).Inc()
}

func (g *depthWithPriorityMetric) Dec(priority int) {
	depth.WithLabelValues(append(g.lvs, strconv.Itoa(priority))...).Dec()
}
