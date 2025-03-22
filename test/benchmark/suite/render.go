// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"sort"
	"strings"
	"text/tabwriter"
)

const (
	benchmarkEnvPrefix = "BENCHMARK_"
)

// RenderReport renders a report out of given list of benchmark report in Markdown format.
func RenderReport(writer io.Writer, name, description string, titleLevel int, reports []*BenchmarkReport) error {
	writeSection(writer, "Test: "+name, titleLevel, description)

	writeSection(writer, "Results", titleLevel+1, "Expand to see the full results.")
	if err := renderResultsTable(writer, reports); err != nil {
		return err
	}

	writeSection(writer, "Metrics", titleLevel+1, "")
	renderMetricsTable(writer, reports)

	writeSection(writer, "Profiles", titleLevel+1, renderProfilesNote())
	renderProfilesTable(writer, "Heap/Memory", "heap", titleLevel+2, reports)

	return nil
}

// newMarkdownStyleTableWriter returns a tabwriter that write table in Markdown style.
func newMarkdownStyleTableWriter(writer io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(writer, 0, 0, 0, ' ', tabwriter.Debug)
}

func renderEnvSettingsTable(writer io.Writer) {
	table := newMarkdownStyleTableWriter(writer)

	headers := []string{
		"RPS",
		"Connections",
		"Duration (Seconds)",
		"CPU Limits (m)",
		"Memory Limits (MiB)",
	}
	writeTableHeader(table, headers)
	writeTableRow(table, headers)

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

	// write headers
	headers := []string{
		"Test Name",
		"Envoy Gateway Memory (MiB)\nmin/max/means",
		"Envoy Gateway CPU (Seconds)\nmin/max/means",
		"Averaged Envoy Proxy Memory (MiB)\nmin/max/means",
		"Averaged Envoy Proxy CPU (Seconds)\nmin/max/means",
	}
	writeTableHeader(table, headers)

	for _, report := range reports {
		data := []string{report.Name}
		data = append(data, getSamplesMinMaxMeans(report.Samples)...)
		writeTableRow(table, data)
	}

	_ = table.Flush()
}

func renderProfilesNote() string {
	return fmt.Sprintf(`The profiles at different scales are stored under %s directory in report, with name %s.

You can visualize them in a web page by running:

%s

Currently, the supported profile types are:
- heap (memory)
`, "`/profiles`", "`{ProfileType}.{TestCase}.pprof`", "```shell\ngo tool pprof -http=: path/to/your.pprof\n```")
}

func renderProfilesTable(writer io.Writer, target, key string, titleLevel int, reports []*BenchmarkReport) {
	writeSection(writer, target, titleLevel,
		"The profiles were sampled when Envoy Gateway Memory is at its maximum.")

	for _, report := range reports {
		// Get the heap profile when control plane memory is at its maximum.
		sortedSamples := make([]BenchmarkMetricSample, len(report.Samples))
		copy(sortedSamples, report.Samples)
		sort.Slice(sortedSamples, func(i, j int) bool {
			return sortedSamples[i].ControlPlaneMem > sortedSamples[j].ControlPlaneMem
		})

		heapPprof := sortedSamples[0].HeapProfile
		heapPprofPath := path.Join(report.ProfilesOutputDir, fmt.Sprintf("heap.%s.pprof", report.Name))
		_ = os.WriteFile(heapPprofPath, heapPprof, 0o600)

		// The image is not be rendered yet, so it is a placeholder for the path.
		// The image will be rendered after the test has finished.
		rootDir := strings.SplitN(heapPprofPath, "/", 2)[0]
		heapPprofPath = strings.TrimPrefix(heapPprofPath, rootDir+"/")
		writeSection(writer, report.Name, titleLevel+1,
			fmt.Sprintf("![%s-%s](%s.png)", key, report.Name, strings.TrimSuffix(heapPprofPath, ".pprof")))
	}
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

func writeTableHeader(table *tabwriter.Writer, headers []string) {
	writeTableRow(table, headers)
	writeTableDelimiter(table, len(headers))
}

// writeTableRow writes one row in Markdown table style.
func writeTableRow(table *tabwriter.Writer, data []string) {
	row := "|"
	for _, v := range data {
		row += v + "\t"
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

func getSamplesMinMaxMeans(samples []BenchmarkMetricSample) []string {
	cpMem := make([]float64, 0, len(samples))
	cpCpu := make([]float64, 0, len(samples))
	dpMem := make([]float64, 0, len(samples))
	dpCpu := make([]float64, 0, len(samples))
	for _, sample := range samples {
		cpMem = append(cpMem, sample.ControlPlaneMem)
		cpCpu = append(cpCpu, sample.ControlPlaneCpu)
		dpMem = append(dpMem, sample.DataPlaneMem)
		dpCpu = append(dpCpu, sample.DataPlaneCpu)
	}

	return []string{
		getMetricsMinMaxMeans(cpMem),
		getMetricsMinMaxMeans(cpCpu),
		getMetricsMinMaxMeans(dpMem),
		getMetricsMinMaxMeans(dpCpu),
	}
}

func getMetricsMinMaxMeans(metrics []float64) string {
	var min, max, sum float64 = metrics[0], 0, 0
	for _, v := range metrics {
		min = math.Min(v, min)
		max = math.Max(v, max)
		sum += v
	}
	return fmt.Sprintf("%.2f / %.2f / %.2f", min, max, sum/float64(len(metrics)))
}
