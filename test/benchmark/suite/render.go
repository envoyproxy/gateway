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
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

const (
	omitEmptyValue     = "-"
	benchmarkEnvPrefix = "BENCHMARK_"

	querySum = "Sum"
	queryAvg = "Avg"
	queryMin = "Min"
	queryMax = "Max"
)

type tableHeader struct {
	name      string
	unit      string
	promQL    string // only valid for metrics table
	queryType string
}

var metricsTableHeader = []tableHeader{
	{
		name: "Test Name",
	},
	{
		name:      "Envoy Gateway Memory",
		unit:      "MiB",
		promQL:    `container_memory_working_set_bytes{namespace="envoy-gateway-system",container="envoy-gateway"}/1024/1024`,
		queryType: querySum,
	},
	{
		name:      "Envoy Gateway CPU",
		unit:      "s",
		promQL:    `container_cpu_usage_seconds_total{namespace="envoy-gateway-system",container="envoy-gateway"}`,
		queryType: querySum,
	},
}

// RenderReport renders a report out of given list of benchmark report in Markdown format.
func RenderReport(writer io.Writer, name, description string, titleLevel int, reports []*BenchmarkReport) error {
	writeSection(writer, "Test: "+name, titleLevel, description)

	writeSection(writer, "Results", titleLevel+1, "Expand to see the full results.")
	if err := renderResultsTable(writer, reports); err != nil {
		return err
	}

	writeSection(writer, "Metrics", titleLevel+1, "")
	renderMetricsTable(writer, reports)
	return nil
}

// newMarkdownStyleTableWriter returns a tabwriter that write table in Markdown style.
func newMarkdownStyleTableWriter(writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, 0, 0, ' ', tabwriter.Debug)
}

func renderEnvSettingsTable(writer io.Writer) {
	table := newMarkdownStyleTableWriter(writer)

	headers := []tableHeader{
		{name: "RPS"},
		{name: "Connections"},
		{name: "Duration", unit: "s"},
		{name: "CPU Limits", unit: "m"},
		{name: "Memory Limits", unit: "MiB"},
	}
	writeTableHeader(table, headers)

	writeTableRow(table, headers, func(_ int, h tableHeader) string {
		env := strings.ReplaceAll(strings.ToUpper(h.name), " ", "_")
		if v, ok := os.LookupEnv(benchmarkEnvPrefix + env); ok {
			return v
		}
		return omitEmptyValue
	})

	_ = table.Flush()
}

func renderResultsTable(writer io.Writer, reports []*BenchmarkReport) error {
	for _, report := range reports {
		tmpResults := bytes.Split(report.Result, []byte("\n"))
		outResults := make([][]byte, 0, len(tmpResults))
		for _, r := range tmpResults {
			if !bytes.HasPrefix(r, []byte("[")) && !bytes.HasPrefix(r, []byte("Nighthawk")) {
				outResults = append(outResults, r)
			}
		}

		writeCollapsibleSection(writer, report.Name, bytes.Join(outResults, []byte("\n")))
	}

	return nil
}

func renderMetricsTable(writer io.Writer, reports []*BenchmarkReport) {
	table := newMarkdownStyleTableWriter(writer)

	writeTableHeader(table, metricsTableHeader)

	for _, report := range reports {
		writeTableRow(table, metricsTableHeader, func(_ int, h tableHeader) string {
			if len(h.promQL) == 0 {
				return report.Name
			}

			if v, ok := report.Metrics[h.name]; ok {
				return strconv.FormatFloat(v, 'f', -1, 64)
			}

			return omitEmptyValue
		})
	}

	_ = table.Flush()
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

func writeTableHeader(table *tabwriter.Writer, headers []tableHeader) {
	writeTableRow(table, headers, func(_ int, h tableHeader) string {
		if len(h.unit) > 0 {
			return fmt.Sprintf("%s (%s)", h.name, h.unit)
		}
		return h.name
	})
	writeTableDelimiter(table, len(headers))
}

// writeTableRow writes one row in Markdown table style according to headers.
func writeTableRow(table *tabwriter.Writer, headers []tableHeader, on func(int, tableHeader) string) {
	row := "|"
	for i, v := range headers {
		row += on(i, v) + "\t"
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
