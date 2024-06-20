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
)

const (
	omitEmptyValue     = "-"
	benchmarkEnvPrefix = "BENCHMARK_"

	// Supported metric type.
	metricTypeGauge   = "gauge"
	metricTypeCounter = "counter"

	// Supported metric unit. TODO: associate them with func map
	metricUnitMiB     = "MiB"
	metricUnitSeconds = "Seconds"
)

type ReportTableHeader struct {
	Name   string       `json:"name"`
	Metric *MetricEntry `json:"metric,omitempty"`
	Result *ResultEntry `json:"result,omitempty"`
}

type MetricEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Unit string `json:"unit"`
	Help string `json:"help,omitempty"`
}

type ResultEntry struct {
	//
}

var (
	controlPlaneMetricHeaders = []*ReportTableHeader{ // TODO: convert and save this config with json
		{
			Name: "Benchmark Name",
		},
		{
			Name: "Envoy Gateway Memory",
			Metric: &MetricEntry{
				Name: "process_resident_memory_bytes",
				Type: metricTypeGauge,
				Unit: metricUnitMiB,
			},
		},
		{
			Name: "Envoy Gateway Total CPU",
			Metric: &MetricEntry{
				Name: "process_cpu_seconds_total",
				Type: metricTypeCounter,
				Unit: metricUnitSeconds,
			},
		},
	}
)

// RenderReport renders a report out of given list of benchmark report in Markdown format.
func RenderReport(writer io.Writer, name, description string, reports []*BenchmarkReport, titleLevel int) error {
	writeSection(writer, name, titleLevel, description)

	writeSection(writer, "Results", titleLevel+1, "Click to see the full results.")
	renderResultsTable(writer, reports)

	writeSection(writer, "Metrics", titleLevel+1, "")
	err := renderMetricsTable(writer, controlPlaneMetricHeaders, reports)
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
	table := newMarkdownStyleTableWriter(writer)

	headers := []string{"RPS", "Connections", "Duration", "CPU Limits", "Memory Limits"}
	writeTableRow(table, headers, func(h string) string {
		return h
	})

	writeTableDelimiter(table, len(headers))

	writeTableRow(table, headers, func(h string) string {
		if v, ok := os.LookupEnv(benchmarkEnvPrefix + strings.ToUpper(h)); ok {
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

func renderMetricsTable(writer io.Writer, headers []*ReportTableHeader, reports []*BenchmarkReport) error {
	table := newMarkdownStyleTableWriter(writer)

	// Write metrics table header.
	writeTableRow(table, headers, func(h *ReportTableHeader) string { // TODO: add footnote support
		if h.Metric != nil && len(h.Metric.Unit) > 0 {
			return fmt.Sprintf("%s (%s)", h.Name, h.Metric.Unit)
		}
		return h.Name
	})

	// Write metrics table delimiter.
	writeTableDelimiter(table, len(headers))

	// Write metrics table body.
	for _, report := range reports {
		mf, err := parseMetrics(report.RawCPMetrics) // TODO: move metrics outside, and add envoyproxy metrics support
		if err != nil {
			return err
		}

		writeTableRow(table, headers, func(h *ReportTableHeader) string {
			if h.Metric != nil {
				if mv, ok := mf[h.Metric.Name]; ok {
					// Store the help of metric for later usage.
					h.Metric.Help = *mv.Help

					switch *mv.Type {
					case prom.MetricType_GAUGE:
						return strconv.FormatFloat(byteToMiB(*mv.Metric[0].Gauge.Value), 'f', -1, 64)
					case prom.MetricType_COUNTER:
						return strconv.FormatFloat(*mv.Metric[0].Counter.Value, 'f', -1, 64)
					}
				}
				return omitEmptyValue
			}

			// Despite metrics, we still got benchmark test name.
			return report.Name
		})
	}

	_ = table.Flush()

	return nil
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
	_, _ = fmt.Fprintln(writer, fmt.Sprintf(`
<details>
<summary>%s</summary>

%s

</details>`, title, fmt.Sprintf("```plaintext\n%s\n```", content)))
}

// writeTableRow writes one row in Markdown table style.
func writeTableRow[T any](table *tabwriter.Writer, values []T, get func(T) string) {
	row := "|"
	for _, v := range values {
		row += get(v) + "\t"
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
