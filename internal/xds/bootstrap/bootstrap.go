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
)

const (
	// envoyCfgFileName is the name of the Envoy configuration file.
	envoyCfgFileName = "bootstrap.yaml"
	// envoyGatewayXdsServerHost is the DNS name of the Xds Server within Envoy Gateway.
	// It defaults to the Envoy Gateway Kubernetes service.
	envoyGatewayXdsServerHost = "envoy-gateway"
	// envoyAdminAddress is the listening address of the envoy admin interface.
	envoyAdminAddress = "127.0.0.1"
	// envoyAdminPort is the port used to expose admin interface.
	envoyAdminPort = 19000
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

// envoyBootstrap defines the envoy Bootstrap configuration.
type bootstrapConfig struct {
	// parameters defines configurable bootstrap configuration parameters.
	parameters bootstrapParameters
	// rendered is the rendered bootstrap configuration.
	rendered string
}

// envoyBootstrap defines the envoy Bootstrap configuration.
type bootstrapParameters struct {
	// XdsServer defines the configuration of the XDS server.
	XdsServer xdsServerParameters
	// AdminServer defines the configuration of the Envoy admin interface.
	AdminServer adminServerParameters
	// ReadyServer defines the configuration for health check ready listener
	ReadyServer readyServerParameters
	// EnablePrometheus defines whether to enable metrics endpoint for prometheus.
	EnablePrometheus bool
	// OtelMetricSinks defines the configuration of the OpenTelemetry sinks.
	OtelMetricSinks []metricSink
	// EnableStatConfig defines whether to to customize the Envoy proxy stats.
	EnableStatConfig bool
	// StatsMatcher is to control creation of custom Envoy stats with prefix,
	// suffix, and regex expressions match on the name of the stats.
	StatsMatcher *StatsMatcherParameters
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
	Port int32
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
	Prefixs            []string
	Suffixs            []string
	RegularExpressions []string
}

// render the stringified bootstrap config in yaml format.
func (b *bootstrapConfig) render() error {
	buf := new(strings.Builder)
	if err := bootstrapTmpl.Execute(buf, b.parameters); err != nil {
		return fmt.Errorf("failed to render bootstrap config: %v", err)
	}
	b.rendered = buf.String()

	return nil
}

// GetRenderedBootstrapConfig renders the bootstrap YAML string
func GetRenderedBootstrapConfig(proxyMetrics *egv1a1.ProxyMetrics) (string, error) {
	var (
		enablePrometheus bool
		metricSinks      []metricSink
		StatsMatcher     StatsMatcherParameters
	)

	if proxyMetrics != nil {
		if proxyMetrics.Prometheus != nil {
			enablePrometheus = true
		}

		addresses := sets.NewString()
		for _, sink := range proxyMetrics.Sinks {
			if sink.OpenTelemetry == nil {
				continue
			}

			// skip duplicate sinks
			addr := fmt.Sprintf("%s:%d", sink.OpenTelemetry.Host, sink.OpenTelemetry.Port)
			if addresses.Has(addr) {
				continue
			}
			addresses.Insert(addr)

			metricSinks = append(metricSinks, metricSink{
				Address: sink.OpenTelemetry.Host,
				Port:    sink.OpenTelemetry.Port,
			})
		}

		if proxyMetrics.Matches != nil {

			// Add custom envoy proxy stats
			for _, match := range proxyMetrics.Matches {
				switch match.Type {
				case egv1a1.Prefix:
					StatsMatcher.Prefixs = append(StatsMatcher.Prefixs, match.Value)
				case egv1a1.Suffix:
					StatsMatcher.Suffixs = append(StatsMatcher.Suffixs, match.Value)
				case egv1a1.RegularExpression:
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
				Address:       envoyAdminAddress,
				Port:          envoyAdminPort,
				AccessLogPath: envoyAdminAccessLogPath,
			},
			ReadyServer: readyServerParameters{
				Address:       envoyReadinessAddress,
				Port:          EnvoyReadinessPort,
				ReadinessPath: EnvoyReadinessPath,
			},
			EnablePrometheus: enablePrometheus,
			OtelMetricSinks:  metricSinks,
		},
	}
	if proxyMetrics != nil && proxyMetrics.Matches != nil {
		cfg.parameters.StatsMatcher = &StatsMatcher
	}

	if err := cfg.render(); err != nil {
		return "", err
	}

	return cfg.rendered, nil
}
