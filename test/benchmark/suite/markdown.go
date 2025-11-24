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
	mm.ControlPlaneProcessMemory = metrics[2]
	mm.DataPlaneMemory = metrics[3]
	mm.DataPlaneCPU = metrics[4]

	return md, mm
}

func renderProtoResult(r *proto.Result) string {
	return "TODO: render result"
}

type MarkdownOutput struct {
	RPS        string
	Connection string
	Duration   string
	CPU        string
	Memory     string

	Results []MarkdownResult
	Metrics []MarkdownMetric
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
