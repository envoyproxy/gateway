// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package contour contains the translation business logic that listens
// to Kubernetes ResourceEventHandler events and translates those into
// additions/deletions in caches connected to the Envoy xDS gRPC API server.
package contour

import (
	"time"

	"github.com/projectcontour/contour/internal/dag"
	"github.com/projectcontour/contour/internal/k8s"
	"github.com/projectcontour/contour/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/tools/cache"
)

// EventRecorder records the count and kind of events forwarded
// to another ResourceEventHandler.
type EventRecorder struct {
	Next    cache.ResourceEventHandler
	Counter *prometheus.CounterVec
}

func (e *EventRecorder) OnAdd(obj interface{}) {
	e.recordOperation("add", obj)
	e.Next.OnAdd(obj)
}

func (e *EventRecorder) OnUpdate(oldObj, newObj interface{}) {
	e.recordOperation("update", newObj) // the api server guarantees that an object's kind cannot be updated
	e.Next.OnUpdate(oldObj, newObj)
}

func (e *EventRecorder) OnDelete(obj interface{}) {
	e.recordOperation("delete", obj)
	e.Next.OnDelete(obj)
}

func (e *EventRecorder) recordOperation(op string, obj interface{}) {
	kind := k8s.KindOf(obj)
	if kind == "" {
		kind = "unknown"
	}
	e.Counter.WithLabelValues(op, kind).Inc()
}

// RebuildMetricsObserver is a dag.Observer that emits metrics for DAG rebuilds.
type RebuildMetricsObserver struct {
	// Metrics to emit.
	metrics *metrics.Metrics

	// httpProxyMetricsEnabled will become ready to read when this EventHandler becomes
	// the leader. If httpProxyMetricsEnabled is not readable, or nil, status events will
	// be suppressed.
	httpProxyMetricsEnabled chan struct{}

	// NextObserver contains the stack of dag.Observers that act on DAG rebuilds.
	nextObserver dag.Observer
}

func NewRebuildMetricsObserver(metrics *metrics.Metrics, nextObserver dag.Observer) *RebuildMetricsObserver {
	return &RebuildMetricsObserver{
		metrics:                 metrics,
		nextObserver:            nextObserver,
		httpProxyMetricsEnabled: make(chan struct{}),
	}
}

func (m *RebuildMetricsObserver) OnElectedLeader() {
	close(m.httpProxyMetricsEnabled)
}

func (m *RebuildMetricsObserver) OnChange(d *dag.DAG) {
	m.metrics.SetDAGLastRebuilt(time.Now())
	m.metrics.SetDAGRebuiltTotal()

	timer := prometheus.NewTimer(m.metrics.CacheHandlerOnUpdateSummary)
	m.nextObserver.OnChange(d)
	timer.ObserveDuration()
}
