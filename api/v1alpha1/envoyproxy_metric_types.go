// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type MetricSinkType string

const (
	MetricSinkTypeOpenTelemetry MetricSinkType = "OpenTelemetry"
)

type ProxyMetrics struct {
	// Prometheus defines the configuration for Admin endpoint `/stats/prometheus`.
	Prometheus *ProxyPrometheusProvider `json:"prometheus,omitempty"`
	// Sinks defines the metric sinks where metrics are sent to.
	// +kubebuilder:validation:MaxItems=16
	Sinks []ProxyMetricSink `json:"sinks,omitempty"`
	// Matches defines configuration for selecting specific metrics instead of generating all metrics stats
	// that are enabled by default. This helps reduce CPU and memory overhead in Envoy, but eliminating some stats
	// may after critical functionality. Here are the stats that we strongly recommend not disabling:
	// `cluster_manager.warming_clusters`, `cluster.<cluster_name>.membership_total`,`cluster.<cluster_name>.membership_healthy`,
	// `cluster.<cluster_name>.membership_degraded`，reference  https://github.com/envoyproxy/envoy/issues/9856,
	// https://github.com/envoyproxy/envoy/issues/14610
	//
	Matches []StringMatch `json:"matches,omitempty"`

	// EnableVirtualHostStats enables envoy stat metrics for virtual hosts.
	//
	// +optional
	EnableVirtualHostStats *bool `json:"enableVirtualHostStats,omitempty"`

	// EnablePerEndpointStats enables per endpoint envoy stats metrics.
	// Please use with caution.
	//
	// +optional
	EnablePerEndpointStats *bool `json:"enablePerEndpointStats,omitempty"`

	// EnableRequestResponseSizesStats enables publishing of histograms tracking header and body sizes of requests and responses.
	//
	// +optional
	EnableRequestResponseSizesStats *bool `json:"enableRequestResponseSizesStats,omitempty"`

	// ClusterStatName defines the value of cluster alt_stat_name, determining how cluster stats are named.
	// For more details, see envoy docs: https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto.html
	// The supported operators for this pattern are:
	// %ROUTE_NAME%: name of Gateway API xRoute resource
	// %ROUTE_NAMESPACE%: namespace of Gateway API xRoute resource
	// %ROUTE_KIND%: kind of Gateway API xRoute resource
	// %ROUTE_RULE_NAME%: name of the Gateway API xRoute section
	// %ROUTE_RULE_NUMBER%: name of the Gateway API xRoute section
	// %BACKEND_REFS%: names of all backends referenced in <NAMESPACE>/<NAME>|<NAMESPACE>/<NAME>|... format
	// Only xDS Clusters created for HTTPRoute and GRPCRoute are currently supported.
	// Default: %ROUTE_KIND%/%ROUTE_NAMESPACE%/%ROUTE_NAME%/rule/%ROUTE_RULE_NUMBER%
	// Example: httproute/my-ns/my-route/rule/0
	//
	// +optional
	ClusterStatName *string `json:"clusterStatName,omitempty"`
}

// ProxyMetricSink defines the sink of metrics.
// Default metrics sink is OpenTelemetry.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'OpenTelemetry' ? has(self.openTelemetry) : !has(self.openTelemetry)",message="If MetricSink type is OpenTelemetry, openTelemetry field needs to be set."
type ProxyMetricSink struct {
	// Type defines the metric sink type.
	// EG currently only supports OpenTelemetry.
	// +kubebuilder:validation:Enum=OpenTelemetry
	// +kubebuilder:default=OpenTelemetry
	// +unionDiscriminator
	Type MetricSinkType `json:"type"`
	// OpenTelemetry defines the configuration for OpenTelemetry sink.
	// It's required if the sink type is OpenTelemetry.
	// +optional
	OpenTelemetry *ProxyOpenTelemetrySink `json:"openTelemetry,omitempty"`
}

// ProxyOpenTelemetrySink defines the configuration for OpenTelemetry sink.
//
// +kubebuilder:validation:XValidation:message="host or backendRefs needs to be set",rule="has(self.host) || self.backendRefs.size() > 0"
// +kubebuilder:validation:XValidation:message="BackendRefs must be used, backendRef is not supported.",rule="!has(self.backendRef)"
// +kubebuilder:validation:XValidation:message="BackendRefs only support Service and Backend kind.",rule="has(self.backendRefs) ? self.backendRefs.all(f, f.kind == 'Service' || f.kind == 'Backend') : true"
// +kubebuilder:validation:XValidation:message="BackendRefs only support Core and gateway.envoyproxy.io group.",rule="has(self.backendRefs) ? (self.backendRefs.all(f, f.group == \"\" || f.group == 'gateway.envoyproxy.io')) : true"
type ProxyOpenTelemetrySink struct {
	BackendCluster `json:",inline"`
	// Host define the service hostname.
	// Deprecated: Use BackendRefs instead.
	//
	// +optional
	Host *string `json:"host,omitempty"`
	// Port defines the port the service is exposed on.
	// Deprecated: Use BackendRefs instead.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=4317
	Port int32 `json:"port,omitempty"`

	// TODO: add support for customizing OpenTelemetry sink in https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/stat_sinks/open_telemetry/v3/open_telemetry.proto#envoy-v3-api-msg-extensions-stat-sinks-open-telemetry-v3-sinkconfig
}

type ProxyPrometheusProvider struct {
	// Disable the Prometheus endpoint.
	Disable bool `json:"disable,omitempty"`
	// Configure the compression on Prometheus endpoint. Compression is useful in situations when bandwidth is scarce and large payloads can be effectively compressed at the expense of higher CPU load.
	// +optional
	Compression *Compression `json:"compression,omitempty"`
}
