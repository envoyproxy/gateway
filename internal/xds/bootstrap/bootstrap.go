// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	// Register embed
	_ "embed"
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/util/sets"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	netutils "github.com/envoyproxy/gateway/internal/utils/net"
	"github.com/envoyproxy/gateway/internal/utils/regex"
)

const (
	// envoyCfgFileName is the name of the Envoy configuration file.
	envoyCfgFileName = "bootstrap.yaml"
	// envoyGatewayXdsServerHost is the DNS name of the Xds Server within Envoy Gateway.
	// It defaults to the Envoy Gateway Kubernetes service.
	envoyGatewayXdsServerHost = "envoy-gateway"
	// EnvoyAdminAddress is the listening v4 address of the envoy admin interface.
	EnvoyAdminAddress   = "127.0.0.1"
	EnvoyAdminAddressV6 = "::1"
	// EnvoyAdminPort is the port used to expose admin interface.
	EnvoyAdminPort = 19000
	// envoyAdminAccessLogPath is the path used to expose admin access log.
	envoyAdminAccessLogPath = "/dev/null"

	// DefaultXdsServerPort is the default listening port of the xds-server.
	DefaultXdsServerPort = 18000

	wasmServerHost = envoyGatewayXdsServerHost
	// DefaultWasmServerPort is the default listening port of the wasm HTTP server.
	wasmServerPort = 18002

	EnvoyStatsPort = 19001

	EnvoyReadinessPort = 19003
	EnvoyReadinessPath = "/ready"

	defaultSdsTrustedCAPath   = "/sds/xds-trusted-ca.json"
	defaultSdsCertificatePath = "/sds/xds-certificate.json"
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
	XdsServer serverParameters
	// WasmServer defines the configuration of the Wasm HTTP server.
	WasmServer serverParameters
	// AdminServer defines the configuration of the Envoy admin interface.
	AdminServer adminServerParameters
	// StatsServer defines the configuration for stats listener
	StatsServer serverParameters
	// SdsCertificatePath defines the path to SDS certificate config.
	SdsCertificatePath string
	// SdsTrustedCAPath defines the path to SDS trusted CA config.
	SdsTrustedCAPath string

	// EnablePrometheus defines whether to enable metrics endpoint for prometheus.
	EnablePrometheus bool
	// EnablePrometheusCompression defines whether to enable HTTP compression on metrics endpoint for prometheus.
	EnablePrometheusCompression bool
	// PrometheusCompressionLibrary defines the HTTP compression library for metrics endpoint for prometheus.
	// TODO: remove this field because it is not used.
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

	// IPFamily of the Listener
	IPFamily             string
	GatewayNamespaceMode bool
}

type serverParameters struct {
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
	IPFamily             *egv1a1.IPFamily
	ProxyMetrics         *egv1a1.ProxyMetrics
	SdsConfig            SdsConfigPath
	XdsServerHost        *string
	XdsServerPort        *int32
	WasmServerPort       *int32
	AdminServerPort      *int32
	StatsServerPort      *int32
	MaxHeapSizeBytes     uint64
	GatewayNamespaceMode bool
}

type SdsConfigPath struct {
	Certificate string
	TrustedCA   string
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
		prometheusCompressionLibrary = "Gzip"
		metricSinks                  []metricSink
		StatsMatcher                 StatsMatcherParameters
	)

	if opts != nil && opts.ProxyMetrics != nil {
		proxyMetrics := opts.ProxyMetrics

		if proxyMetrics.Prometheus != nil {
			enablePrometheus = !proxyMetrics.Prometheus.Disable

			if proxyMetrics.Prometheus.Compression != nil {
				enablePrometheusCompression = true
				prometheusCompressionLibrary = string(proxyMetrics.Prometheus.Compression.Type)
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
				host, port = netutils.BackendHostAndPort(sink.OpenTelemetry.BackendRefs[0].BackendObjectReference, "")
			}
			addr := net.JoinHostPort(host, strconv.Itoa(int(port)))
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
			XdsServer: serverParameters{
				Address: envoyGatewayXdsServerHost,
				Port:    DefaultXdsServerPort,
			},
			WasmServer: serverParameters{
				Address: wasmServerHost,
				Port:    wasmServerPort,
			},
			AdminServer: adminServerParameters{
				Address:       EnvoyAdminAddress,
				Port:          EnvoyAdminPort,
				AccessLogPath: envoyAdminAccessLogPath,
			},
			StatsServer: serverParameters{
				Address: netutils.IPv4ListenerAddress,
				Port:    EnvoyStatsPort,
			},
			SdsCertificatePath:           defaultSdsCertificatePath,
			SdsTrustedCAPath:             defaultSdsTrustedCAPath,
			EnablePrometheus:             enablePrometheus,
			EnablePrometheusCompression:  enablePrometheusCompression,
			PrometheusCompressionLibrary: prometheusCompressionLibrary,
			OtelMetricSinks:              metricSinks,
		},
	}

	// Bootstrap config override
	if opts != nil {
		if opts.ProxyMetrics != nil && opts.ProxyMetrics.Matches != nil {
			cfg.parameters.StatsMatcher = &StatsMatcher
		}

		// Override Sds configs
		if len(opts.SdsConfig.Certificate) > 0 {
			cfg.parameters.SdsCertificatePath = opts.SdsConfig.Certificate
		}
		if len(opts.SdsConfig.TrustedCA) > 0 {
			cfg.parameters.SdsTrustedCAPath = opts.SdsConfig.TrustedCA
		}

		if opts.XdsServerHost != nil {
			cfg.parameters.XdsServer.Address = *opts.XdsServerHost
		}

		// Override the various server port
		if opts.XdsServerPort != nil {
			cfg.parameters.XdsServer.Port = *opts.XdsServerPort
		}
		if opts.AdminServerPort != nil {
			cfg.parameters.AdminServer.Port = *opts.AdminServerPort
		}
		if opts.StatsServerPort != nil {
			cfg.parameters.StatsServer.Port = *opts.StatsServerPort
		}
		if opts.WasmServerPort != nil {
			cfg.parameters.WasmServer.Port = *opts.WasmServerPort
		}

		if opts.IPFamily != nil {
			cfg.parameters.IPFamily = string(*opts.IPFamily)
			switch *opts.IPFamily {
			case egv1a1.IPv6:
				cfg.parameters.AdminServer.Address = EnvoyAdminAddressV6
				cfg.parameters.StatsServer.Address = netutils.IPv6ListenerAddress
			case egv1a1.DualStack:
				cfg.parameters.StatsServer.Address = netutils.IPv6ListenerAddress
			}
		}
		cfg.parameters.GatewayNamespaceMode = opts.GatewayNamespaceMode
		cfg.parameters.OverloadManager.MaxHeapSizeBytes = opts.MaxHeapSizeBytes
	}

	if err := cfg.render(); err != nil {
		return "", err
	}

	return cfg.rendered, nil
}
