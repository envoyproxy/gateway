---
date: 2023-10-10
title: "Control Plane Observability: Metrics"
author: Xunzhuo Liu
linkTitle: "Control Plane Observability: Metrics"
---

{{% alert title="State" color="warning" %}}

+ Author: [Xunzhuo Liu](https://github.com/Xunzhuo)
+ Affiliation: Tencent
+ Data: 2023-10-12
+ Status: Done
{{% /alert %}}

This document aims to cover all aspects of envoy gateway control plane metrics observability.

{{% alert title="Note" color="secondary" %}}
**Data plane** observability (while important) is outside of scope for this document.
{{% /alert %}}

## Current State

At present, the Envoy Gateway control plane provides logs and controller-runtime metrics, without traces. Logs are managed through our proprietary library (`internal/logging`, a shim to `zap`) and are written to `/dev/stdout`.

The absence of comprehensive and robust control plane metrics observability hinders the effective monitoring of Envoy Gateway in a production environment, a critical requirement before deploying Envoy Gateway into production.

## Goals

Our objectives include:

+ Supporting **PULL** mode for Prometheus metrics and exposing these metrics on the admin address.
+ Supporting **PUSH** mode for Prometheus metrics, thereby sending metrics to the Open Telemetry Stats sink.
+ Offering a **COMMON** metrics library so developers can effortlessly add new metrics/labels to each Envoy Gateway component.
+ Providing **BASIC** metrics produced by each Envoy Gateway component, including:
  + Provider
  + Resource Translator
  + Infra Manager
  + xDS Translator
  + Extension Manager

## Non-Goals

Our non-goals include:

+ Supporting other stats sinks.

## Use-Cases

The use-cases include:

+ Exposing Prometheus metrics in the Envoy Gateway Control Plane.
+ Pushing Envoy Gateway Control Plane metrics via the Open Telemetry Sink.

## Design

### Standards

Our metrics, and traces in the future, will be built upon the [OpenTelemetry](https://opentelemetry.io/) standards. All metrics will be configured via the [OpenTelemetry SDK](https://opentelemetry.io/docs/specs/otel/metrics/sdk/), which offers neutral libraries that can be connected to various backends.

This approach allows the Envoy Gateway code to concentrate on the crucial aspect - generating the metrics - and delegate all other tasks to systems designed for telemetry ingestion.

### Attributes

OpenTelemetry defines a set of [Semantic Conventions](https://opentelemetry.io/docs/concepts/semantic-conventions/), including [Kubernetes specific ones](https://opentelemetry.io/docs/specs/otel/resource/semantic_conventions/k8s/).

These attributes can be expressed in logs (as keys of structured logs), traces (as attributes), and metrics (as labels).

We aim to use attributes consistently where applicable. Where possible, these should adhere to codified Semantic Conventions; when not possible, they should maintain consistency across the project.

### Extensibility

Envoy Gateway supports both **PULL/PUSH** mode metrics, with Metrics exported via Prometheus by default.

Additionally, Envoy Gateway can export metrics using both the [OTEL gRPC metrics exporter](https://opentelemetry.io/docs/specs/otel/metrics/sdk_exporters/otlp/#general) and [OTEL HTTP metrics exporter](https://opentelemetry.io/docs/specs/otel/metrics/sdk_exporters/otlp/#general), which pushes metrics by grpc/http to a remote OTEL collector.

Users can extend these in two ways:

#### Downstream Collection

Based on the exported data, other tools can collect, process, and export telemetry as needed. Some examples include:

+ Metrics in **PULL** mode: The OTEL collector can scrape Prometheus and export to X.
+ Metrics in **PUSH** mode: The OTEL collector can receive OTEL gRPC/HTTP exporter metrics and export to X.

While the examples above involve OTEL collectors, there are numerous other systems available.

#### Vendor extensions

The OTEL libraries allow for the registration of Providers/Handlers. While we will offer the default ones (PULL via Prometheus, PUSH via OTEL HTTP metrics exporter) mentioned in Envoy Gateway's extensibility, we can easily allow custom builds of Envoy Gateway to plug in alternatives if the default options don't meet their needs.

For instance, users may prefer to write metrics over the OTLP gRPC metrics exporter instead of the HTTP metrics exporter. This is perfectly acceptable -- and almost impossible to prevent. The OTEL has ways to register their providers/exporters, and Envoy Gateway can ensure its usage is such that it's not overly difficult to swap out a different provider/exporter.

### Stability

Observability is, in essence, a user-facing API. Its primary purpose is to be consumed - by both humans and tooling. Therefore, having well-defined guarantees around their formats is crucial.

Please note that this refers only to the contents of the telemetry - what we emit, the names of things, semantics, etc. Other settings like Prometheus vs OTLP, JSON vs plaintext, logging levels, etc., are not considered.

I propose the following:

#### Metrics

Metrics offer the greatest potential for providing guarantees. They often directly influence alerts and dashboards, making changes highly impactful. This contrasts with traces and logs, which are often used for ad-hoc analysis, where minor changes to information can be easily understood by a human.

Moreover, there is precedent for this: [Kubernetes Metrics Lifecycle](https://kubernetes.io/docs/concepts/cluster-administration/system-metrics/#metric-lifecycle) has well-defined processes, and Envoy Gateway's dataplane (Envoy Proxy) metrics are de facto stable.

Currently, all Envoy Gateway metrics lack defined stability. I suggest we categorize all existing metrics as either:

+ ***Deprecated***: a metric that is intended to be phased out.
+ ***Experimental***: a metric that is off by default.
+ ***Alpha***: a metric that is on by default.

We should aim to promote a core set of metrics to **Stable** within a few releases.

## Library

Envoy Gateway should offer a metrics library abstraction wrapper, effectively hiding OTEL from the rest of the codebase.

### Deep Dive

Although OpenTelemetry has a fairly user-friendly API, I believe we still benefit from a wrapper, which provides the following advantages:

+ Workarounds for various library limitations (histograms, gauges, counters).
+ Codification of best practices; while we don't do much today, this may become more important as we start to define [Stability](#stability).
+ Provision of optimizations that would otherwise be tedious.

For now, I believe having a wrapper benefits us. However, it seems plausible that the need for this wrapper could diminish over time.

#### Metric Abstraction

##### Metric Types

```go
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
```

##### Metric

```go
// A Metric collects numerical observations.
type Metric interface {
	// Name returns the name value of a Metric.
	Name() string

	// Record makes an observation of the provided value for the given measure.
	Record(value float64)

	// RecordInt makes an observation of the provided value for the measure.
	RecordInt(value int64)

	// Increment records a value of 1 for the current measure.
	// For Counters, this is equivalent to adding 1 to the current value.
	// For Gauges, this is equivalent to setting the value to 1.
	// For Histograms, this is equivalent to making an observation of value 1.
	Increment()

	// Decrement records a value of -1 for the current measure.
	// For Counters, this is equivalent to subtracting -1 to the current value.
	// For Gauges, this is equivalent to setting the value to -1.
	// For Histograms, this is equivalent to making an observation of value -1.
	Decrement()

	// With creates a new Metric, with the LabelValues provided.
	// This allows creating a set of pre-dimensioned data for recording purposes.
	// This is primarily used for documentation and convenience.
	// Metrics created with this method do not need to be registered (they share the registration of their parent Metric).
	With(labelValues ...LabelValue) Metric
}
```

##### Metric Label

```go
// Label holds a metric dimension which can be operated on using the interface
// methods.
type Label interface {
	// Value will set the provided value for the Label.
	Value(value string) LabelValue
}

// LabelValue holds an action to take on a metric dimension's value.
type LabelValue interface {
	// Key will get the key of the Label.
	Key() Label
	// Value will get the value of the Label.
	Value() string
}

```

##### Metric Metadata

```go
// Metadata records a metric's metadata.
type Metadata struct {
	Name        string
	Type        MetricType
	Description string
	Bounds      []float64
}

// metrics stores stores metrics
type metricstore struct {
	started bool
	mu      sync.Mutex
	stores  map[string]Metadata
}

// stores is a global that stores all registered metrics
var stores = metricstore{
	stores: map[string]Metadata{},
}
```

#### Metric Implementation

##### Exported Methods

```go

// NewCounter creates a new Counter Metric (the values will be cumulative).
// That means that data collected by the new Metric will be summed before export.
func NewCounter(name, description string, opts ...MetricOption) Metric {
	stores.register(Metadata{
		Name:        name,
		Type:        CounterType,
		Description: description,
	})
	o, disabled := metricOptions(name, description, opts...)
	if disabled != nil {
		return disabled
	}
	return newCounter(o)
}

// NewGauge creates a new Gauge Metric. That means that data collected by the new
// Metric will export only the last recorded value.
func NewGauge(name, description string, opts ...MetricOption) Metric {
	stores.register(Metadata{
		Name:        name,
		Type:        GaugeType,
		Description: description,
	})
	o, disabled := metricOptions(name, description, opts...)
	if disabled != nil {
		return disabled
	}

	return newGauge(o)
}

// NewHistogram creates a new Metric with an aggregation type of Histogram.
// This means that the data collected by the Metric will be collected and exported as a histogram, with the specified bounds.
func NewHistogram(name, description string, bounds []float64, opts ...MetricOption) Metric {
	stores.register(Metadata{
		Name:        name,
		Type:        HistogramType,
		Description: description,
		Bounds:      bounds,
	})
	o, disabled := metricOptions(name, description, opts...)
	if disabled != nil {
		return disabled
	}
	return newHistogram(o)
}

func newCounter(o MetricOptions) *otelCounter {
	c, err := meter().Float64Counter(o.Name,
		api.WithDescription(o.Description),
		api.WithUnit(string(o.Unit)))
	if err != nil {
		metricsLogger.Error(err, "failed to create otel Counter")
	}
	m := &otelCounter{c: c}
	m.embed = embed{
		name: o.Name,
		m:    m,
	}
	return m
}

func newGauge(o MetricOptions) *otelGauge {
	r := &otelGauge{
		mutex: &sync.RWMutex{},
	}
	r.stores = map[attribute.Set]*otelGaugeValues{}
	g, err := meter().Float64ObservableGauge(o.Name,
		api.WithFloat64Callback(func(ctx context.Context, observer api.Float64Observer) error {
			r.mutex.Lock()
			defer r.mutex.Unlock()
			for _, gv := range r.stores {
				observer.Observe(gv.val, gv.opt...)
			}
			return nil
		}),
		api.WithDescription(o.Description),
		api.WithUnit(string(o.Unit)))
	if err != nil {
		metricsLogger.Error(err, "failed to create otel Gauge")
	}
	r.g = g
	r.embed = embed{
		name: o.Name,
		m:    r,
	}

	return r
}

func newHistogram(o MetricOptions) *otelHistogram {
	d, err := meter().Float64Histogram(o.Name,
		api.WithDescription(o.Description),
		api.WithUnit(string(o.Unit)))
	if err != nil {
		metricsLogger.Error(err, "failed to create otel Histogram")
	}
	m := &otelHistogram{d: d}
	m.embed = embed{
		name: o.Name,
		m:    m,
	}
	return m
}

```

##### Embed Metric

```go
// embed metric implementation.
type embed struct {
	name  string
	attrs []attribute.KeyValue
	m     Metric
}

func (f embed) Name() string {
	return f.name
}

func (f embed) Increment() {
	f.m.Record(1)
}

func (f embed) Decrement() {
	f.m.Record(-1)
}

func (f embed) RecordInt(value int64) {
	f.m.Record(float64(value))
}
```

#### Disabled Metric

```go
// disabled metric implementation.
type disabled struct {
	name string
}

// Decrement implements Metric
func (dm *disabled) Decrement() {}

// Increment implements Metric
func (dm *disabled) Increment() {}

// Name implements Metric
func (dm *disabled) Name() string {
	return dm.name
}

// Record implements Metric
func (dm *disabled) Record(value float64) {}

// RecordInt implements Metric
func (dm *disabled) RecordInt(value int64) {}

// With implements Metric
func (dm *disabled) With(labelValues ...LabelValue) Metric {
	return dm
}
```

##### Otel Counter Metric

```go
type otelCounter struct {
	embed

	c                api.Float64Counter
	preRecordOptions []api.AddOption
}

func (f *otelCounter) Record(value float64) {
	if f.preRecordOptions != nil {
		f.c.Add(context.Background(), value, f.preRecordOptions...)
	} else {
		f.c.Add(context.Background(), value)
	}
}

func (f *otelCounter) With(labelValues ...LabelValue) Metric {
	attrs, set := mergeLabelValues(f.embed, labelValues)
	m := &otelCounter{
		c:                f.c,
		preRecordOptions: []api.AddOption{api.WithAttributeSet(set)},
	}
	m.embed = embed{
		name:  f.name,
		attrs: attrs,
		m:     m,
	}
	return m
}

```

##### Otel Gauge Metric

```go
type otelGauge struct {
	embed

	g       api.Float64ObservableGauge
	mutex   *sync.RWMutex
	stores  map[attribute.Set]*otelGaugeValues
	current *otelGaugeValues
}

type otelGaugeValues struct {
	val float64
	opt []api.ObserveOption
}

func (f *otelGauge) Record(value float64) {
	f.mutex.Lock()
	if f.current == nil {
		f.current = &otelGaugeValues{}
		f.stores[attribute.NewSet()] = f.current
	}
	f.current.val = value
	f.mutex.Unlock()
}

func (f *otelGauge) With(labelValues ...LabelValue) Metric {
	attrs, set := mergeLabelValues(f.embed, labelValues)
	m := &otelGauge{
		g:      f.g,
		mutex:  f.mutex,
		stores: f.stores,
	}
	if _, f := m.stores[set]; !f {
		m.stores[set] = &otelGaugeValues{
			opt: []api.ObserveOption{api.WithAttributeSet(set)},
		}
	}
	m.current = m.stores[set]
	m.embed = embed{
		name:  f.name,
		attrs: attrs,
		m:     m,
	}
	return m
}
```

##### Otel Histogram Metric

```go
type otelHistogram struct {
	embed

	d                api.Float64Histogram
	preRecordOptions []api.RecordOption
}

func (f *otelHistogram) Record(value float64) {
	if f.preRecordOptions != nil {
		f.d.Record(context.Background(), value, f.preRecordOptions...)
	} else {
		f.d.Record(context.Background(), value)
	}
}

func (f *otelHistogram) With(labelValues ...LabelValue) Metric {
	attrs, set := mergeLabelValues(f.embed, labelValues)
	m := &otelHistogram{
		d:                f.d,
		preRecordOptions: []api.RecordOption{api.WithAttributeSet(set)},
	}
	m.embed = embed{
		name:  f.name,
		attrs: attrs,
		m:     m,
	}
	return m
}
```

##### Otel Label

```go
// NewLabel will attempt to create a new Label.
func NewLabel(key string) Label {
	return otelLabel{attribute.Key(key)}
}

// A otelLabel provides a named dimension for a Metric.
type otelLabel struct {
	key attribute.Key
}

// Value creates a new LabelValue for the Label.
func (l otelLabel) Value(value string) LabelValue {
	return otelLabelValue{l.key.String(value)}
}

// A LabelValue represents a Label with a specific value. It is used to record
// values for a Metric.
type otelLabelValue struct {
	keyValue attribute.KeyValue
}

func (l otelLabelValue) Key() Label {
	return otelLabel{l.keyValue.Key}
}

func (l otelLabelValue) Value() string {
	return l.keyValue.Value.AsString()
}
```

#### Extensibility

##### Register

```go
// registerForHandler sets the global metrics registry to the provided Prometheus registerer.
// if enables prometheus, it will return a prom http handler.
func registerForHandler(opts registerOptions) (http.Handler, error) {
	otelOpts := []metric.Option{}

	if err := registerOTELPromExporter(&otelOpts, opts); err != nil {
		return nil, err
	}
	if err := registerOTELHTTPexporter(&otelOpts, opts); err != nil {
		return nil, err
	}
	if err := registerOTELgRPCexporter(&otelOpts, opts); err != nil {
		return nil, err
	}
	otelOpts = append(otelOpts, stores.preAddOptions()...)

	mp := metric.NewMeterProvider(otelOpts...)
	otel.SetMeterProvider(mp)

	if opts.pullOptions.enable {
		return promhttp.HandlerFor(opts.pullOptions.gatherer, promhttp.HandlerOpts{}), nil
	}
	return nil, nil
}
```

##### Exporter

```go
// registerOTELPromExporter registers OTEL prometheus exporter (PULL mode).
func registerOTELPromExporter(otelOpts *[]metric.Option, opts metricsOptions) error {
	if opts.pullOptions.enable {
		promOpts := []otelprom.Option{
			otelprom.WithoutScopeInfo(),
			otelprom.WithoutTargetInfo(),
			otelprom.WithoutUnits(),
			otelprom.WithRegisterer(opts.pullOptions.reg),
			otelprom.WithoutCounterSuffixes(),
		}
		promreader, err := otelprom.New(promOpts...)
		if err != nil {
			return err
		}

		*otelOpts = append(*otelOpts, metric.WithReader(promreader))
		metricsLogger.Info("initialized metrics pull endpoint", "address", opts.pullOptions.address)
	}

	return nil
}

// registerOTELHTTPexporter registers OTEL HTTP metrics exporter (PUSH mode).
func registerOTELHTTPexporter(otelOpts *[]metric.Option, opts metricsOptions) error {
	for _, sink := range opts.pushOptions.sinks {
		address := fmt.Sprintf("%s:%d", sink.host, sink.port)
		httpexporter, err := otlpmetrichttp.New(
			context.Background(),
			otlpmetrichttp.WithEndpoint(address),
			otlpmetrichttp.WithInsecure(),
		)
		if err != nil {
			return err
		}

		otelreader := metric.NewPeriodicReader(httpexporter)
		*otelOpts = append(*otelOpts, metric.WithReader(otelreader))
		metricsLogger.Info("initialized metrics push endpoint", "address", address)
	}

	return nil
}

// registerOTELgRPCexporter registers OTEL gRPC metrics exporter (PUSH mode).
func registerOTELgRPCexporter(otelOpts *[]metric.Option, opts metricsOptions) error {
	for _, sink := range opts.pushOptions.sinks {
		if sink.protocol == "grpc" {
			address := fmt.Sprintf("%s:%d", sink.host, sink.port)
			httpexporter, err := otlpmetricgrpc.New(
				context.Background(),
				otlpmetricgrpc.WithEndpoint(address),
				otlpmetricgrpc.WithInsecure(),
			)
			if err != nil {
				return err
			}

			otelreader := metric.NewPeriodicReader(httpexporter)
			*otelOpts = append(*otelOpts, metric.WithReader(otelreader))
			metricsLogger.Info("initialized otel grpc metrics push endpoint", "address", address)
		}
	}

	return nil
}
```

#### How to Use?

> Let me take metrics instrumentation in watchable message queue as an example

+ create metrics in target pkg:

```go
	// metrics definitions
	watchableHandleUpdates = metrics.NewCounter(
		"watchable_queue_handle_updates_total",
		"Total number of updates handled by watchable queue.",
	)

	watchableHandleUpdateErrors = metrics.NewCounter(
		"watchable_queue_handle_updates_errors_total",
		"Total number of update errors handled by watchable queue.",
	)

	watchableDepth = metrics.NewGauge(
		"watchable_queue_depth",
		"Current depth of watchable message queue.",
	)

	watchableHandleUpdateTimeSeconds = metrics.NewHistogram(
		"watchable_queue_handle_update_time_seconds",
		"How long in seconds a update handled by watchable queue.",
		[]float64{0.001, 0.01, 0.1, 1, 5, 10},
	)

```

+ create metric labels if needed in target pkg:

```go
	// metrics label definitions
	// component is which component the update belong to.
	componentNameLabel = metrics.NewLabel("component_name")
	// resource is which resource the update belong to.
	resourceTypeLabel = metrics.NewLabel("resource_type")
```

+ record the metric value with label in target pkg:

```go
type Update[K comparable, V any] watchable.Update[K, V]

var logger = logging.DefaultLogger(v1alpha1.LogLevelInfo).WithName("watchable")

type UpdateMetadata struct {
	Component string
	Resource  string
}

func (m UpdateMetadata) LabelValues() []metrics.LabelValue {
	labels := []metrics.LabelValue{}
	if m.Component != "" {
		labels = append(labels, componentNameLabel.Value(m.Component))
	}
	if m.Resource != "" {
		labels = append(labels, resourceTypeLabel.Value(m.Resource))
	}

	return labels
}

// HandleSubscription takes a channel returned by
// watchable.Map.Subscribe() (or .SubscribeSubset()), and calls the
// given function for each initial value in the map, and for any
// updates.
//
// This is better than simply iterating over snapshot.Updates because
// it handles the case where the watchable.Map already contains
// entries before .Subscribe is called.
func HandleSubscription[K comparable, V any](
	meta UpdateMetadata,
	subscription <-chan watchable.Snapshot[K, V],
	handle func(updateFunc Update[K, V], errChans chan error),
) {
	errChans := make(chan error, 10)
	go func() {
		for err := range errChans {
			logger.WithValues("component", meta.Component).Error(err, "observed an error")
			watchableHandleUpdateErrors.With(meta.LabelValues()...).Increment()
		}
	}()

	if snapshot, ok := <-subscription; ok {
		for k, v := range snapshot.State {
			startHandleTime := time.Now()
			handle(Update[K, V]{Key: k, Value: v}, errChans)
			watchableHandleUpdates.With(meta.LabelValues()...).Increment()
			watchableHandleUpdateTimeSeconds.With(meta.LabelValues()...).Record(time.Since(startHandleTime).Seconds())
		}
	}
	for snapshot := range subscription {
		watchableDepth.With(meta.LabelValues()...).RecordInt(int64(len(subscription)))
		for _, update := range snapshot.Updates {
			startHandleTime := time.Now()
			handle(Update[K, V](update), errChans)
			watchableHandleUpdates.With(meta.LabelValues()...).Increment()
			watchableHandleUpdateTimeSeconds.With(meta.LabelValues()...).Record(time.Since(startHandleTime).Seconds())
		}
	}
}

```

+ Build and Test

Visit the `{adminPort}:/metrics` and you can see the new added metrics:

![metrics](/img/metrics-demo-1.png)

## Basic Metrics Instrumentation

### Provider

#### Label

##### Label Name: component_name

Scope:

1. watchable_queue_handle_updates_total
2. watchable_queue_handle_updates_errors_total
3. watchable_queue_depth
4. watchable_queue_handle_update_time_seconds

Supported values:

+ provider

##### Label Name: resource_type

Scope:

1. watchable_queue_handle_updates_total
2. watchable_queue_handle_updates_errors_total
3. watchable_queue_depth
4. watchable_queue_handle_update_time_seconds

Supported values:

+ httproute-status
+ tcproute-status
+ udproute-status
+ tlsroute-status
+ envoypatchpolicy-status
+ ...

#### Counter

+ watchable_queue_handle_updates_total: Total number of updates handled by watchable queue.
+ watchable_queue_handle_updates_errors_total: Total number of update errors handled by watchable queue.

---

> controller-runtime metrics

+ certwatcher_read_certificate_total: Total number of certificate reads
+ certwatcher_read_certificate_errors_total: Total number of certificate read errors
+ controller_runtime_reconcile_errors_total: Total number of reconciliation errors per controller
+ controller_runtime_reconcile_total: Total number of reconciliations per controller
+ rest_client_requests_total: Number of HTTP requests, partitioned by status code, method, and host.
+ workqueue_adds_total: Total number of adds handled by workqueue
+ workqueue_retries_total: Total number of retries handled by workqueue

#### Gauge

+ watchable_queue_depth: Current depth of watchable message queue.

---

> controller-runtime metrics

+ controller_runtime_active_workers: Number of currently used workers per controller
+ controller_runtime_max_concurrent_reconciles: Maximum number of concurrent reconciles per controller
+ workqueue_depth: Current depth of workqueue

#### Histogram

+ watchable_queue_handle_update_time_seconds: How long in seconds a update handled by watchable queue.

---

> controller-runtime metrics

+ controller_runtime_reconcile_time_seconds: Length of time per reconciliation per controller
+ workqueue_longest_running_processor_seconds: How many seconds has the longest running processor for workqueue been running.
+ workqueue_queue_duration_seconds: How long in seconds an item stays in workqueue before being requested
+ workqueue_unfinished_work_seconds: How many seconds of work has been done that is in progress and hasn't been observed by work_duration. Large values indicate stuck threads. One can deduce the number of stuck threads by observing the rate at which this increases.
+ workqueue_work_duration_seconds: How long in seconds processing an item from workqueue takes.

### Resource Translator

#### Label

##### Label Name: component_name

Scope:

1. watchable_queue_handle_updates_total
2. watchable_queue_handle_updates_errors_total
3. watchable_queue_depth
4. watchable_queue_handle_update_time_seconds

Supported values:

+ gateway-api

#### Counter

+ watchable_queue_handle_updates_total: Total number of updates handled by watchable queue.
+ watchable_queue_handle_updates_errors_total: Total number of update errors handled by watchable queue.

#### Gauge

+ watchable_queue_depth: Current depth of watchable message queue.

#### Histogram

+ watchable_queue_handle_update_time_seconds: How long in seconds a update handled by watchable queue.

### Infra Manager

#### Label

##### Label Name: component_name

Scope:

1. watchable_queue_handle_updates_total
2. watchable_queue_handle_updates_errors_total
3. watchable_queue_depth
4. watchable_queue_handle_update_time_seconds

Supported values:

+ infrastructure

##### Label Name: operation

Scope:

1. infra_manager_resources_errors_total

Supported values:

+ created
+ updated
+ deleted

##### Label Name: k8s_resource_type

Scope:

1. infra_manager_resources_created_total
2. infra_manager_resources_updated_total
3. infra_manager_resources_deleted_total
4. infra_manager_resources_errors_total

Supported values:

+ Deployment
+ Service
+ ServiceAccount
+ ConfigMap

##### Label Name: k8s_resource_name

Scope:

1. infra_manager_resources_created_total
2. infra_manager_resources_updated_total
3. infra_manager_resources_deleted_total
4. infra_manager_resources_errors_total

Supported values:

resource name

##### Label Name: k8s_resource_namespace

Scope:

1. infra_manager_resources_created_total
2. infra_manager_resources_updated_total
3. infra_manager_resources_deleted_total
4. infra_manager_resources_errors_total

Supported values:

resource namespace

#### Counter

+ watchable_queue_handle_updates_total: Total number of updates handled by watchable queue.
+ watchable_queue_handle_updates_errors_total: Total number of update errors handled by watchable queue.

+ infra_manager_resources_created_total: Total number of the resources created by infra manager.
+ infra_manager_resources_updated_total: Total number of the resources updated by infra manager.
+ infra_manager_resources_deleted_total: Total number of the resources deleted by infra manager.

+ infra_manager_resources_errors_total: Total number of the resources errors encountered by infra manager.

#### Gauge

+ watchable_queue_depth: Current depth of watchable message queue.

#### Histogram

+ watchable_queue_handle_update_time_seconds: How long in seconds a update handled by watchable queue.

### xDS Translator

#### Label

##### Label Name: component_name

Scope:

1. watchable_queue_handle_updates_total
2. watchable_queue_handle_updates_errors_total
3. watchable_queue_depth
4. watchable_queue_handle_update_time_seconds

Supported values:

+ xds-translator

#### Counter

+ watchable_queue_handle_updates_total: Total number of updates handled by watchable queue.
+ watchable_queue_handle_updates_errors_total: Total number of update errors handled by watchable queue.

#### Gauge

+ watchable_queue_depth: Current depth of watchable message queue.

#### Histogram

+ watchable_queue_handle_update_time_seconds: How long in seconds a update handled by watchable queue.

### Extension Manager

#### Label

##### Label Name: target

Scope:

1. extension_manager_post_hook_calls_total
2. extension_manager_post_hook_call_errors_total
3. extension_manager_post_hook_time_seconds

Supported values:

+ route
+ virtualHost
+ listener
+ cluster

#### Counter

+ extension_manager_post_hook_calls_total: Total number of the post hook calls in extension manager.
+ extension_manager_post_hook_call_errors_total: Total number of the post hook call errors in extension manager.

#### Histogram

+ extension_manager_post_hook_time_seconds: How long in seconds a post hook called in extension manager.

## Envoy Gateway API Types

New APIs will be added to Envoy Gateway config, which are used to manage Control Plane Telemetry bootstrap configs.

### EnvoyGatewayTelemetry

```go
// EnvoyGatewayTelemetry defines telemetry configurations for envoy gateway control plane.
// Control plane will focus on metrics observability telemetry and tracing telemetry later.
type EnvoyGatewayTelemetry struct {
	// Metrics defines metrics configuration for envoy gateway.
	Metrics *EnvoyGatewayMetrics `json:"metrics,omitempty"`
}
```

### EnvoyGatewayMetrics

```go
// EnvoyGatewayMetrics defines control plane push/pull metrics configurations.
type EnvoyGatewayMetrics struct {
	// Address defines the address of Envoy Gateway Metrics Server.
	Address *EnvoyGatewayMetricsAddress
	// Sinks defines the metric sinks where metrics are sent to.
	Sinks []EnvoyGatewayMetricSink `json:"sinks,omitempty"`
	// Prometheus defines the configuration for prometheus endpoint.
	Prometheus *EnvoyGatewayPrometheusProvider `json:"prometheus,omitempty"`
}

// EnvoyGatewayMetricSink defines control plane
// metric sinks where metrics are sent to.
type EnvoyGatewayMetricSink struct {
	// Type defines the metric sink type.
	// EG control plane currently supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	Type MetricSinkType `json:"type"`
	// Protocol define the sink service protocol. 
	// Currently supported: grpc, http.
	Protocol string `json:"protocol"`
	// Host define the sink service hostname.
	Host string `json:"host"`
	// Port defines the port the sink service is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=4318
	Port int32 `json:"port,omitempty"`
}

// EnvoyGatewayPrometheusProvider will expose prometheus endpoint
// `/stats/prometheus` and reuse Envoy Gateway admin port.
type EnvoyGatewayPrometheusProvider struct {
	// Enable defines if enables the prometheus metrics in pull mode. Default is true.
	//
	// +optional
	// +kubebuilder:default=true
	Enable bool `json:"enable,omitempty"`
}

// EnvoyGatewayMetricsAddress defines the Envoy Gateway Metrics Address configuration.
type EnvoyGatewayMetricsAddress struct {
	// Port defines the port the metrics server is exposed on.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=19010
	Port int `json:"port,omitempty"`
	// Host defines the metrics server hostname.
	//
	// +optional
	// +kubebuilder:default="0.0.0.0"
	Host string `json:"host,omitempty"`
}
```

#### Example

+ The following is an example to enable prometheus metric.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
	controllerName: gateway.envoyproxy.io/gatewayclass-controller
logging:
	level:
	default: info
provider:
	type: Kubernetes
telemetry:
	metrics:
	  address:
	    host: 0.0.0.0
		port: 19010
	  prometheus:
		enable: true
```

+ The following is an example to send metric via Open Telemetry sink to OTEL gRPC Collector.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
	controllerName: gateway.envoyproxy.io/gatewayclass-controller
logging:
	level:
	default: info
provider:
	type: Kubernetes
telemetry:
	metrics:
	  sinks:
	  - type: OpenTelemetry
	    host: otel-collector.monitoring.svc.cluster.local
	    port: 4317
		protocol: grpc
```

+ The following is an example to enable prometheus metric and send metric via Open Telemetry sink to OTEL HTTP Collector at the same time.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyGateway
gateway:
	controllerName: gateway.envoyproxy.io/gatewayclass-controller
logging:
	level:
	default: info
provider:
	type: Kubernetes
telemetry:
	metrics:
	  prometheus:
		enable: true
	  sinks:
	  - type: OpenTelemetry
	    host: otel-collector.monitoring.svc.cluster.local
	    port: 4318
		protocol: http
```
