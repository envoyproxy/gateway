// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package metrics

import (
	"errors"
	"sync"

	"go.opentelemetry.io/otel"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	log "github.com/envoyproxy/gateway/internal/logging"
)

var (
	meter = func() api.Meter {
		return otel.GetMeterProvider().Meter("envoy-gateway")
	}

	metricsLogger = log.DefaultLogger(v1alpha1.LogLevelInfo).WithName("metrics")
)

func init() {
	otel.SetLogger(metricsLogger.Logger)
}

// MetricType is the type of a metric.
type MetricType string

// Metric type supports:
// * Counter: A Counter is a simple metric that only goes up (increments).
//
// * Gauge: A Gauge is a metric that represent
// a single numerical value that can arbitrarily go up and down.
//
// * Histogram: A Histogram samples observations and counts them in configurable buckets.
// It also provides a sum of all observed values.
// It's used to visualize the statistical distribution of these observations.

const (
	CounterType   MetricType = "Counter"
	GaugeType     MetricType = "Gauge"
	HistogramType MetricType = "Histogram"
)

// Metadata records a metric's metadata.
type Metadata struct {
	Name        string
	Type        MetricType
	Description string
	Bounds      []float64
}

// metrics stores stores metrics
type store struct {
	started bool
	mu      sync.Mutex
	stores  map[string]Metadata
}

// stores is a global that stores all registered metrics
var stores = store{
	stores: map[string]Metadata{},
}

// register records a newly defined metric. Only valid before an exporter is set.
func (d *store) register(store Metadata) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.started {
		metricsLogger.Error(errors.New("cannot initialize metric after metric has started"), "metric", store.Name)
	}
	d.stores[store.Name] = store
}

// preAddOptions runs pre-run steps before adding to meter provider.
func (d *store) preAddOptions() []metric.Option {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.started = true
	opts := []metric.Option{}
	for name, store := range d.stores {
		if store.Bounds == nil {
			continue
		}
		// for each histogram metric (i.e. those with bounds), set up a view explicitly defining those buckets.
		v := metric.WithView(metric.NewView(
			metric.Instrument{Name: name},
			metric.Stream{
				Aggregation: metric.AggregationExplicitBucketHistogram{
					Boundaries: store.Bounds,
				}},
		))
		opts = append(opts, v)
	}
	return opts
}
