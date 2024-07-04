// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package suite

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	prom "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	omitEmptyValue     = "-"
	benchmarkEnvPrefix = "BENCHMARK_"

	// Supported metric type.
	metricTypeGauge   = "gauge"
	metricTypeCounter = "counter"

	// Supported metric unit.
	metricUnitMiB      = "MiB"
	metricUnitSeconds  = "Seconds"
	metricUnitMilliCPU = "m"
)

type ReportTableHeader struct {
	Name   string
	Metric *MetricEntry

	// Underlying name of one envoy-proxy, used by data-plane metrics.
	ProxyName string
}

type MetricEntry struct {
	Name             string
	Type             string
	FromControlPlane bool
	DisplayUnit      string
	ConvertUnit      func(float64) float64
}

// RenderReport renders a report out of given list of benchmark report in Markdown format.
func RenderReport(writer io.Writer, name, description string, reports []*BenchmarkReport, titleLevel int) error {
	headerSettings := []ReportTableHeader{
		{
			Name: "Benchmark Name",
		},
		{
			Name: "Envoy Gateway Memory",
			Metric: &MetricEntry{
				Name:             "process_resident_memory_bytes",
				Type:             metricTypeGauge,
				DisplayUnit:      metricUnitMiB,
				FromControlPlane: true,
				ConvertUnit:      byteToMiB,
			},
		},
		{
			Name: "Envoy Gateway Total CPU",
			Metric: &MetricEntry{
				Name:             "process_cpu_seconds_total",
				Type:             metricTypeCounter,
				DisplayUnit:      metricUnitSeconds,
				FromControlPlane: true,
			},
		},
		{
			Name: "Envoy Proxy Memory",
			Metric: &MetricEntry{
				Name:             "envoy_server_memory_allocated",
				Type:             metricTypeGauge,
				DisplayUnit:      metricUnitMiB,
				FromControlPlane: false,
				ConvertUnit:      byteToMiB,
			},
		},
	}

	writeSection(writer, name, titleLevel, description)

	writeSection(writer, "Results", titleLevel+1, "Click to see the full results.")
	renderResultsTable(writer, reports)

	writeSection(writer, "Metrics", titleLevel+1, "")
	err := renderMetricsTable(writer, headerSettings, reports)
	if err != nil {
		return err
	}

	return nil
}

// newMarkdownStyleTableWriter returns a tabwriter that write table in Markdown style.
func newMarkdownStyleTableWriter(writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, 0, 0, ' ', tabwriter.Debug)
}

func renderEnvSettingsTable(writer io.Writer) {
	_, _ = fmt.Fprintln(writer, "Benchmark test settings:")

	table := newMarkdownStyleTableWriter(writer)

	headers := []ReportTableHeader{
		{
			Name: "RPS",
		},
		{
			Name: "Connections",
		},
		{
			Name: "Duration",
			Metric: &MetricEntry{
				DisplayUnit: metricUnitSeconds,
			},
		},
		{
			Name: "CPU Limits",
			Metric: &MetricEntry{
				DisplayUnit: metricUnitMilliCPU,
			},
		},
		{
			Name: "Memory Limits",
			Metric: &MetricEntry{
				DisplayUnit: metricUnitMiB,
			},
		},
	}

	renderMetricsTableHeader(table, headers)

	writeTableRow(table, headers, func(_ int, h ReportTableHeader) string {
		env := strings.ReplaceAll(strings.ToUpper(h.Name), " ", "_")
		if v, ok := os.LookupEnv(benchmarkEnvPrefix + env); ok {
			return v
		}
		return omitEmptyValue
	})

	_ = table.Flush()
}

func renderResultsTable(writer io.Writer, reports []*BenchmarkReport) {
	// TODO: better processing these benchmark results.
	for _, report := range reports {
		writeCollapsibleSection(writer, report.Name, report.RawResult)
	}
}

func renderMetricsTable(writer io.Writer, headerSettings []ReportTableHeader, reports []*BenchmarkReport) error {
	table := newMarkdownStyleTableWriter(writer)

	// Preprocess the table header for metrics table.
	var headers []ReportTableHeader
	// 1. Collect all the possible proxy names.
	proxyNames := sets.NewString()
	for _, report := range reports {
		for name := range report.RawDPMetrics {
			proxyNames.Insert(name)
		}
	}
	// 2. Generate header names for data-plane proxies.
	for _, hs := range headerSettings {
		if hs.Metric != nil && !hs.Metric.FromControlPlane {
			for i, proxyName := range proxyNames.List() {
				names := strings.Split(proxyName, "-")
				headers = append(headers, ReportTableHeader{
					Name:      fmt.Sprintf("%s: %s<sup>[%d]</sup>", hs.Name, names[len(names)-1], i+1),
					Metric:    hs.Metric,
					ProxyName: proxyName,
				})
			}
		} else {
			// For control-plane metrics or plain header.
			headers = append(headers, hs)
		}
	}

	renderMetricsTableHeader(table, headers)

	for _, report := range reports {
		mfCP, err := parseMetrics(report.RawCPMetrics)
		if err != nil {
			return err
		}

		mfDPs := make(map[string]map[string]*prom.MetricFamily, len(report.RawDPMetrics))
		for dpName, dpMetrics := range report.RawDPMetrics {
			mfDP, err := parseMetrics(dpMetrics)
			if err != nil {
				return err
			}
			mfDPs[dpName] = mfDP
		}

		writeTableRow(table, headers, func(_ int, h ReportTableHeader) string {
			if h.Metric == nil {
				return report.Name
			}

			if h.Metric.FromControlPlane {
				return processMetricValue(mfCP, h)
			} else {
				if mfDP, ok := mfDPs[h.ProxyName]; ok {
					return processMetricValue(mfDP, h)
				}
			}

			return omitEmptyValue
		})
	}

	_ = table.Flush()

	// Generate footnotes for envoy-proxy headers.
	for i, proxyName := range proxyNames.List() {
		_, _ = fmt.Fprintln(writer, fmt.Sprintf("%d.", i+1), proxyName)
	}

	return nil
}

func renderMetricsTableHeader(table *tabwriter.Writer, headers []ReportTableHeader) {
	writeTableRow(table, headers, func(_ int, h ReportTableHeader) string {
		if h.Metric != nil && len(h.Metric.DisplayUnit) > 0 {
			return fmt.Sprintf("%s (%s)", h.Name, h.Metric.DisplayUnit)
		}
		return h.Name
	})

	writeTableDelimiter(table, len(headers))
}

func byteToMiB(x float64) float64 {
	return math.Round(x / (1024 * 1024))
}

// writeSection writes one section in Markdown style, content is optional.
func writeSection(writer io.Writer, title string, level int, content string) {
	md := fmt.Sprintf("\n%s %s\n", strings.Repeat("#", level), title)
	if len(content) > 0 {
		md += fmt.Sprintf("\n%s\n", content)
	}
	_, _ = fmt.Fprintln(writer, md)
}

// writeCollapsibleSection writes one collapsible section in Markdown style.
func writeCollapsibleSection(writer io.Writer, title string, content []byte) {
	summary := fmt.Sprintf("```plaintext\n%s\n```", content)
	_, _ = fmt.Fprintf(writer, `
<details>
<summary>%s</summary>

%s

</details>
`, title, summary)
}

// writeTableRow writes one row in Markdown table style.
func writeTableRow[T any](table *tabwriter.Writer, values []T, get func(int, T) string) {
	row := "|"
	for i, v := range values {
		row += get(i, v) + "\t"
	}

	_, _ = fmt.Fprintln(table, row)
}

// writeTableDelimiter writes table delimiter in Markdown table style.
func writeTableDelimiter(table *tabwriter.Writer, n int) {
	sep := "|"
	for i := 0; i < n; i++ {
		sep += "-\t"
	}

	_, _ = fmt.Fprintln(table, sep)
}

// parseMetrics parses input metrics that in Prometheus format.
func parseMetrics(metrics []byte) (map[string]*prom.MetricFamily, error) {
	var (
		reader = bytes.NewReader(metrics)
		parser expfmt.TextParser
	)

	mf, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	return mf, nil
}

// processMetricValue process one metric value according to the given header and metric families.
func processMetricValue(metricFamilies map[string]*prom.MetricFamily, header ReportTableHeader) string {
	if mf, ok := metricFamilies[header.Metric.Name]; ok {
		var value float64

		switch header.Metric.Type {
		case metricTypeGauge:
			value = *mf.Metric[0].Gauge.Value
		case metricTypeCounter:
			value = *mf.Metric[0].Counter.Value
		default:
			return omitEmptyValue
		}

		if header.Metric.ConvertUnit != nil {
			value = header.Metric.ConvertUnit(value)
		}

		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	return omitEmptyValue
}
