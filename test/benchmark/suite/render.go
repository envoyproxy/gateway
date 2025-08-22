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

// RenderReport renders a report out of given list of benchmark report in Markdown format.
func RenderReport(writer io.Writer, name, description string, titleLevel int, reports []*BenchmarkReport) error {
	writeSection(writer, "Test: "+name, titleLevel, description)

	writeSection(writer, "Results", titleLevel+1, "Expand to see the full results.")
	if err := renderResultsTable(writer, reports); err != nil {
		return err
	}

	writeSection(writer, "Metrics", titleLevel+1,
		"The CPU usage statistics of both control-plane and data-plane are the CPU usage per second over the past 30 seconds.")
	renderMetricsTable(writer, reports)

	writeSection(writer, "Profiles", titleLevel+1, renderProfilesNote())
	renderProfilesTable(writer, "Heap", "heap", titleLevel+2, reports)

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

	data := []string{
		os.Getenv("BENCHMARK_RPS"),
		os.Getenv("BENCHMARK_CONNECTIONS"),
		os.Getenv("BENCHMARK_DURATION"),
		os.Getenv("BENCHMARK_CPU_LIMITS"),
		os.Getenv("BENCHMARK_MEMORY_LIMITS"),
	}
	writeTableRow(table, data)

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
		"Route Convergence Time <br> p50/p90/p99",
		"Envoy Gateway Memory Container (MiB) <br> min/max/means",
		"Envoy Gateway Memory Process (MiB) <br> min/max/means",
		"Envoy Gateway CPU (%) <br> min/max/means",
		"Averaged Envoy Proxy Memory (MiB) <br> min/max/means",
		"Averaged Envoy Proxy CPU (%) <br> min/max/means",
	}
	writeTableHeader(table, headers)

	for _, report := range reports {
		routeConvergenceDuration := "N/A"
		if report.RouteConvergence != nil {
			routeConvergenceDuration = fmt.Sprintf("%s/%s/%s",
				report.RouteConvergence.P50,
				report.RouteConvergence.P90,
				report.RouteConvergence.P99)
		}

		data := []string{report.Name, routeConvergenceDuration}
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
			return sortedSamples[i].ControlPlaneContainerMem > sortedSamples[j].ControlPlaneContainerMem
		})

		sort.Slice(sortedSamples, func(i, j int) bool {
			return sortedSamples[i].ControlPlaneProcessMem > sortedSamples[j].ControlPlaneProcessMem
		})

		heapPprof := sortedSamples[0].HeapProfile
		// report name contains spaces, replace them with dashes to make it URL-friendly.
		friendlyFilename := strings.ReplaceAll(report.Name, " ", "-")
		heapPprofPath := path.Join(report.ProfilesOutputDir, fmt.Sprintf("heap.%s.pprof", friendlyFilename))
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
	cpcMem := make([]float64, 0, len(samples))
	cppMem := make([]float64, 0, len(samples))
	cpCPU := make([]float64, 0, len(samples))
	dpMem := make([]float64, 0, len(samples))
	dpCPU := make([]float64, 0, len(samples))
	for _, sample := range samples {
		cpcMem = append(cpcMem, sample.ControlPlaneContainerMem)
		cppMem = append(cppMem, sample.ControlPlaneProcessMem)
		cpCPU = append(cpCPU, sample.ControlPlaneCPU)
		dpMem = append(dpMem, sample.DataPlaneMem)
		dpCPU = append(dpCPU, sample.DataPlaneCPU)
	}

	return []string{
		getMetricsMinMaxMeans(cpcMem),
		getMetricsMinMaxMeans(cppMem),
		getMetricsMinMaxMeans(cpCPU),
		getMetricsMinMaxMeans(dpMem),
		getMetricsMinMaxMeans(dpCPU),
	}
}

func getMetricsMinMaxMeans(metrics []float64) string {
	var min, max, avg float64 = math.MaxFloat64, 0, 0
	for _, v := range metrics {
		min = math.Min(v, min)
		max = math.Max(v, max)
		avg += v
	}
	if min == math.MaxFloat64 {
		min = 0
	}
	if len(metrics) > 0 {
		avg /= float64(len(metrics))
	}

	return fmt.Sprintf("%.2f / %.2f / %.2f", min, max, avg)
}
