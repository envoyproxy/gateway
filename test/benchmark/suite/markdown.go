// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"strings"
	"text/template"

	"github.com/envoyproxy/gateway/test/benchmark/proto"
)

//go:embed benchmark_report.tpl.md
var mdTpl string

func ToMarkdown(report *BenchmarkSuiteReport) ([]byte, error) {
	t, err := template.New("benchmark").Parse(mdTpl)
	if err != nil {
		return nil, err
	}
	md := convertToMarkdownOutput(report)
	var buf bytes.Buffer
	if err := t.Execute(&buf, md); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func convertToMarkdownOutput(report *BenchmarkSuiteReport) *MarkdownOutput {
	out := &MarkdownOutput{
		RPS:        report.Settings["rps"],
		Connection: report.Settings["connection"],
		Duration:   report.Settings["duration"],
		CPU:        report.Settings["cpu"],
		Memory:     report.Settings["memory"],
	}

	for _, c := range report.Reports {
		for _, r := range c.Reports {
			md, metric := convertToMarkdownResult(r)
			out.Results = append(out.Results, md)
			out.Metrics = append(out.Metrics, metric)
			out.Heaps = append(out.Heaps, HeapResult{
				Title: r.Title,
				Name:  r.Title,
				URL:   r.HeapProfileImage,
			})
		}
	}

	return out
}

func convertToMarkdownResult(r *BenchmarkCaseReport) (MarkdownResult, MarkdownMetric) {
	md := MarkdownResult{}
	mm := MarkdownMetric{}

	md.Summary = r.Title
	md.Text = renderProtoResult(r.Result)

	mm.Name = r.Title
	metrics := getSamplesMinMaxMeans(r.Samples)
	mm.ControlPlaneContainerMemory = metrics[0]
	mm.ControlPlaneProcessMemory = metrics[1]
	mm.ControlPlaneCPU = metrics[2]
	mm.DataPlaneMemory = metrics[3]
	mm.DataPlaneCPU = metrics[4]

	return md, mm
}

func renderProtoResult(r *proto.Result) string {
	// https://github.com/envoyproxy/nighthawk/blob/main/source/client/output_formatter_impl.cc#L160-L161
	var sb strings.Builder
	for _, stat := range r.Statistics {
		if len(stat.GetPercentiles()) <= 1 {
			continue
		}
		sb.WriteString(fmt.Sprintf("%s (%d samples)\n", statIDFriendlyStatName(stat.Id), stat.Count))
		sb.WriteString(fmt.Sprintf("  min: %s | mean: %s | max: %s | pstdev: %s\n",
			stat.GetMin().AsDuration(), stat.GetMean().AsDuration(), stat.GetMax().AsDuration(), stat.GetPstdev().AsDuration()))
		sb.WriteString("\n")
		sb.WriteString("  Percentile\tCount\t\tValue\n")

		lastPercentile := float64(-1)
		for _, p := range []float64{0.0, 0.5, 0.75, 0.8, 0.9, 0.95, 0.99, 0.999, 1.0} {
			for _, percentile := range stat.Percentiles {
				current := percentile.GetPercentile()
				if current >= p && lastPercentile < current {
					lastPercentile = current
					if float64(current) > 0 && float64(current) < 1 {
						sb.WriteString(fmt.Sprintf("  %f\t\t%d\t\t%s\n",
							percentile.GetPercentile(), percentile.Count, percentile.GetDuration().AsDuration()))
					}
					break
				}
			}
		}

		sb.WriteString("\n")
	}
	return sb.String()
}

func statIDFriendlyStatName(statID string) string {
	switch statID {
	case "benchmark_http_client.queue_to_connect":
		return "Queueing and connection setup latency"
	case "benchmark_http_client.request_to_response":
		return "Request start to response end"
	case "sequencer.callback":
		return "Initiation to completion"
	case "sequencer.blocking":
		return "Blocking. Results are skewed when significant numbers are reported here."
	case "benchmark_http_client.response_body_size":
		return "Response body size in bytes"
	case "benchmark_http_client.response_header_size":
		return "Response header size in bytes"

	default:
		return statID
	}
}

type MarkdownOutput struct {
	RPS        string
	Connection string
	Duration   string
	CPU        string
	Memory     string

	Results []MarkdownResult
	Metrics []MarkdownMetric
	Heaps   []HeapResult
}

type HeapResult struct {
	Title string
	Name  string
	URL   string
}

type MarkdownResult struct {
	Summary string
	Text    string
}

type MarkdownMetric struct {
	Name                        string
	ControlPlaneContainerMemory string
	ControlPlaneProcessMemory   string
	ControlPlaneCPU             string

	DataPlaneMemory string
	DataPlaneCPU    string
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
	var minV, maxV, avg float64 = math.MaxFloat64, 0, 0
	for _, v := range metrics {
		minV = math.Min(v, minV)
		maxV = math.Max(v, maxV)
		avg += v
	}
	if minV == math.MaxFloat64 {
		minV = 0
	}
	if len(metrics) > 0 {
		avg /= float64(len(metrics))
	}

	return fmt.Sprintf("%.2f / %.2f / %.2f", minV, maxV, avg)
}
