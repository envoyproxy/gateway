// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	// Register embed
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/net"
	"github.com/envoyproxy/gateway/internal/utils/regex"
)

const (
	// envoyCfgFileName is the name of the Envoy configuration file.
	envoyCfgFileName = "bootstrap.yaml"
	// envoyGatewayXdsServerHost is the DNS name of the Xds Server within Envoy Gateway.
	// It defaults to the Envoy Gateway Kubernetes service.
	envoyGatewayXdsServerHost = "envoy-gateway"
	// EnvoyAdminAddress is the listening address of the envoy admin interface.
	EnvoyAdminAddress = "127.0.0.1"
	// EnvoyAdminPort is the port used to expose admin interface.
	EnvoyAdminPort = 19000
	// envoyAdminAccessLogPath is the path used to expose admin access log.
	envoyAdminAccessLogPath = "/dev/null"

	// DefaultXdsServerPort is the default listening port of the xds-server.
	DefaultXdsServerPort = 18000

	envoyReadinessAddress = "0.0.0.0"
	EnvoyReadinessPort    = 19001
	EnvoyReadinessPath    = "/ready"
)

//go:embed bootstrap.yaml.tpl
var bootstrapTmplStr string

var bootstrapTmpl = template.Must(template.New(envoyCfgFileName).Parse(bootstrapTmplStr))

// bootstrapConfig defines the envoy Bootstrap configuration.
type bootstrapConfig struct {
	// parameters defines configurable bootstrap configuration parameters.
	parameters bootstrapParameters
	// rendered is the rendered bootstrap configuration.
	rendered string
}

// bootstrapParameters defines the envoy Bootstrap configuration.
type bootstrapParameters struct {
	// XdsServer defines the configuration of the XDS server.
	XdsServer xdsServerParameters
	// AdminServer defines the configuration of the Envoy admin interface.
	AdminServer adminServerParameters
	// ReadyServer defines the configuration for health check ready listener
	ReadyServer readyServerParameters
	// EnablePrometheus defines whether to enable metrics endpoint for prometheus.
	EnablePrometheus bool
	// EnablePrometheusCompression defines whether to enable HTTP compression on metrics endpoint for prometheus.
	EnablePrometheusCompression bool
	// PrometheusCompressionLibrary defines the HTTP compression library for metrics endpoint for prometheus.
	PrometheusCompressionLibrary string

	// OtelMetricSinks defines the configuration of the OpenTelemetry sinks.
	OtelMetricSinks []metricSink
	// EnableStatConfig defines whether to customize the Envoy proxy stats.
	EnableStatConfig bool
	// StatsMatcher is to control creation of custom Envoy stats with prefix,
	// suffix, and regex expressions match on the name of the stats.
	StatsMatcher *StatsMatcherParameters
	// OverloadManager defines the configuration of the Envoy overload manager.
	OverloadManager overloadManagerParameters
}

type xdsServerParameters struct {
	// Address is the address of the XDS Server that Envoy is managed by.
	Address string
	// Port is the port of the XDS Server that Envoy is managed by.
	Port int32
}

type metricSink struct {
	// Address is the address of the XDS Server that Envoy is managed by.
	Address string
	// Port is the port of the XDS Server that Envoy is managed by.
	Port uint32
}

type adminServerParameters struct {
	// Address is the address of the Envoy admin interface.
	Address string
	// Port is the port of the Envoy admin interface.
	Port int32
	// AccessLogPath is the path of the Envoy admin access log.
	AccessLogPath string
}

type readyServerParameters struct {
	// Address is the address of the Envoy readiness probe
	Address string
	// Port is the port of envoy readiness probe
	Port int32
	// ReadinessPath is the path for the envoy readiness probe
	ReadinessPath string
}

type StatsMatcherParameters struct {
	Exacts             []string
	Prefixes           []string
	Suffixes           []string
	RegularExpressions []string
}

type overloadManagerParameters struct {
	MaxHeapSizeBytes uint64
}

type RenderBootstrapConfigOptions struct {
	ProxyMetrics     *egv1a1.ProxyMetrics
	MaxHeapSizeBytes uint64
}

// render the stringified bootstrap config in yaml format.
func (b *bootstrapConfig) render() error {
	buf := new(strings.Builder)
	if err := bootstrapTmpl.Execute(buf, b.parameters); err != nil {
		return fmt.Errorf("failed to render bootstrap config: %w", err)
	}
	b.rendered = buf.String()

	return nil
}

// GetRenderedBootstrapConfig renders the bootstrap YAML string
func GetRenderedBootstrapConfig(opts *RenderBootstrapConfigOptions) (string, error) {
	var (
		enablePrometheus             = true
		enablePrometheusCompression  = false
		PrometheusCompressionLibrary = "gzip"
		metricSinks                  []metricSink
		StatsMatcher                 StatsMatcherParameters
	)

	if opts != nil && opts.ProxyMetrics != nil {
		proxyMetrics := opts.ProxyMetrics

		if proxyMetrics.Prometheus != nil {
			enablePrometheus = !proxyMetrics.Prometheus.Disable

			if proxyMetrics.Prometheus.Compression != nil {
				enablePrometheusCompression = true
				PrometheusCompressionLibrary = string(proxyMetrics.Prometheus.Compression.Type)
			}
		}

		addresses := sets.NewString()
		for _, sink := range proxyMetrics.Sinks {
			if sink.OpenTelemetry == nil {
				continue
			}

			// skip duplicate sinks
			var host string
			var port uint32
			if sink.OpenTelemetry.Host != nil {
				host, port = *sink.OpenTelemetry.Host, uint32(sink.OpenTelemetry.Port)
			}
			if len(sink.OpenTelemetry.BackendRefs) > 0 {
				host, port = net.BackendHostAndPort(sink.OpenTelemetry.BackendRefs[0].BackendObjectReference, "")
			}
			addr := fmt.Sprintf("%s:%d", host, port)
			if addresses.Has(addr) {
				continue
			}
			addresses.Insert(addr)

			metricSinks = append(metricSinks, metricSink{
				Address: host,
				Port:    port,
			})
		}

		if proxyMetrics.Matches != nil {
			// Add custom envoy proxy stats
			for _, match := range proxyMetrics.Matches {
				// matchType default to exact
				matchType := egv1a1.StringMatchExact
				if match.Type != nil {
					matchType = *match.Type
				}
				switch matchType {
				case egv1a1.StringMatchExact:
					StatsMatcher.Exacts = append(StatsMatcher.Exacts, match.Value)
				case egv1a1.StringMatchPrefix:
					StatsMatcher.Prefixes = append(StatsMatcher.Prefixes, match.Value)
				case egv1a1.StringMatchSuffix:
					StatsMatcher.Suffixes = append(StatsMatcher.Suffixes, match.Value)
				case egv1a1.StringMatchRegularExpression:
					if err := regex.Validate(match.Value); err != nil {
						return "", err
					}
					StatsMatcher.RegularExpressions = append(StatsMatcher.RegularExpressions, match.Value)
				}
			}
		}
	}

	cfg := &bootstrapConfig{
		parameters: bootstrapParameters{
			XdsServer: xdsServerParameters{
				Address: envoyGatewayXdsServerHost,
				Port:    DefaultXdsServerPort,
			},
			AdminServer: adminServerParameters{
				Address:       EnvoyAdminAddress,
				Port:          EnvoyAdminPort,
				AccessLogPath: envoyAdminAccessLogPath,
			},
			ReadyServer: readyServerParameters{
				Address:       envoyReadinessAddress,
				Port:          EnvoyReadinessPort,
				ReadinessPath: EnvoyReadinessPath,
			},
			EnablePrometheus:             enablePrometheus,
			EnablePrometheusCompression:  enablePrometheusCompression,
			PrometheusCompressionLibrary: PrometheusCompressionLibrary,
			OtelMetricSinks:              metricSinks,
		},
	}
	if opts != nil && opts.ProxyMetrics != nil && opts.ProxyMetrics.Matches != nil {
		cfg.parameters.StatsMatcher = &StatsMatcher
	}

	if opts != nil {
		cfg.parameters.OverloadManager.MaxHeapSizeBytes = opts.MaxHeapSizeBytes
	}

	if err := cfg.render(); err != nil {
		return "", err
	}

	return cfg.rendered, nil
}
