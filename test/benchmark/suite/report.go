// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark

package suite

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/internal/troubleshoot/collect"
	prom "github.com/envoyproxy/gateway/test/utils/prometheus"
)

const (
	controlPlaneContainerMemQL = `process_resident_memory_bytes{namespace="envoy-gateway-system", control_plane="envoy-gateway"}/1024/1024`
	controlPlaneProcessMemQL   = `go_memstats_heap_inuse_bytes{namespace="envoy-gateway-system", control_plane="envoy-gateway"}/1024/1024`
	controlPlaneCPUQL          = `rate(process_cpu_seconds_total{namespace="envoy-gateway-system", control_plane="envoy-gateway"}[%DURATIONs])*100`
	dataPlaneMemQL             = `container_memory_working_set_bytes{namespace="envoy-gateway-system", container="envoy"}/1024/1024`
	dataPlaneCPUQLFormat       = `rate(container_cpu_usage_seconds_total{namespace="envoy-gateway-system", container="envoy"}[%DURATIONs])*100`
	DurationFormatter          = "%DURATION"

	benchmarkCPURateWindow = 30 * time.Second
)

// BenchmarkMetricSample contains sampled metrics and profiles data.
type BenchmarkMetricSample struct {
	ControlPlaneContainerMem float64
	ControlPlaneProcessMem   float64
	ControlPlaneCPU          float64
	DataPlaneMem             float64
	DataPlaneCPU             float64

	HeapProfile []byte
}

func (s *BenchmarkMetricSample) String() string {
	return fmt.Sprintf("ControlPlaneContainerMem: %.2f MiB, ControlPlaneProcessMem: %.2f MiB, ControlPlaneCPU: %.2f%%, DataPlaneMem: %.2f MiB, DataPlaneCPU: %.2f%%",
		s.ControlPlaneContainerMem,
		s.ControlPlaneProcessMem,
		s.ControlPlaneCPU,
		s.DataPlaneMem,
		s.DataPlaneCPU)
}

type BenchmarkReport struct {
	Name              string
	Routes            int
	RoutesPerHost     int
	Phase             string
	ProfilesOutputDir string
	RouteConvergence  *PerfDuration
	// Nighthawk benchmark result
	Result []byte
	// Prometheus metrics and pprof profiles sampled data
	Samples []BenchmarkMetricSample

	kubeClient kube.CLIClient
	promClient *prom.Client
}

type PerfDuration struct {
	P50 time.Duration `json:"p50"`
	P90 time.Duration `json:"p90"`
	P99 time.Duration `json:"p99"`
}

func NewBenchmarkReport(name, profilesOutputDir string, kubeClient kube.CLIClient, promClient *prom.Client) *BenchmarkReport {
	return &BenchmarkReport{
		Name:              name,
		ProfilesOutputDir: profilesOutputDir,
		kubeClient:        kubeClient,
		promClient:        promClient,
	}
}

func (r *BenchmarkReport) Sample(t *testing.T, ctx context.Context, startTime time.Time) (err error) {
	sample := BenchmarkMetricSample{}

	if mErr := r.sampleMetrics(ctx, &sample, startTime); mErr != nil {
		err = errors.Join(mErr)
	}

	if pErr := r.sampleProfiles(ctx, &sample); pErr != nil {
		err = errors.Join(err, pErr)
	}

	// If we add sample when error occurs during sampling, the report would be incorrect.
	// For example, cpu/mem would be zero if query fails, and it would affect the calculation of
	// average/max values later.
	if err == nil {
		tlog.Logf(t, "Sampled metrics: %s", sample.String())
		r.Samples = append(r.Samples, sample)
	}
	return err
}

func (r *BenchmarkReport) GetResult(ctx context.Context, job *types.NamespacedName) error {
	pods, err := r.kubeClient.Kube().CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + job.Name})
	if err != nil {
		return err
	}

	if len(pods.Items) < 1 {
		return fmt.Errorf("failed to get any pods for job %s", job.String())
	}

	// Find the pod that complete successfully.
	var pod corev1.Pod
	for i := range pods.Items {
		p := &pods.Items[i]
		if p.Status.Phase == corev1.PodSucceeded {
			pod = *p
			break
		}
	}

	logs, err := r.getLogsFromPod(ctx, &types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace})
	if err != nil {
		return err
	}

	r.Result = logs

	return nil
}

func (r *BenchmarkReport) sampleMetrics(ctx context.Context, sample *BenchmarkMetricSample, startTime time.Time) (err error) {
	// Sample memory
	cpcMem, qErr := r.promClient.QuerySum(ctx, controlPlaneContainerMemQL)
	if qErr != nil {
		err = errors.Join(fmt.Errorf("failed to query control plane container memory: %w", qErr))
	}
	cppMem, qErr := r.promClient.QuerySum(ctx, controlPlaneProcessMemQL)
	if qErr != nil {
		err = errors.Join(fmt.Errorf("failed to query control plane process memory: %w", qErr))
	}
	dpMem, qErr := r.promClient.QueryAvg(ctx, dataPlaneMemQL)
	if qErr != nil {
		err = errors.Join(err, fmt.Errorf("failed to query data plane memory: %w", qErr))
	}
	// Sample cpu

	// CPU usages is calculated based on the Kubernetes container_cpu_usage_seconds_total counter metric.
	// We use a fixed window size of 30s for rate calculation. However, to ensure that we only capture
	// metrics during the benchmark run period (and not before), if the benchmark run duration is
	// less than the fixed window size,
	durationSeconds := int(benchmarkCPURateWindow.Seconds())
	elapsed := time.Since(startTime)
	if elapsed < benchmarkCPURateWindow {
		durationSeconds = int(elapsed.Seconds())
		if durationSeconds < 1 {
			durationSeconds = 1
		}
	}
	durationStr := fmt.Sprintf("%d", durationSeconds)
	cpCPUQL := strings.ReplaceAll(controlPlaneCPUQL, DurationFormatter, durationStr)

	cpCPU, qErr := r.promClient.QuerySum(ctx, cpCPUQL)
	if qErr != nil {
		err = errors.Join(err, fmt.Errorf("failed to query control plane cpu: %w", qErr))
	}

	dpCPUQL := strings.ReplaceAll(dataPlaneCPUQLFormat, DurationFormatter, durationStr)
	dpCPU, qErr := r.promClient.QueryAvg(ctx, dpCPUQL)
	if qErr != nil {
		err = errors.Join(err, fmt.Errorf("failed to query data plane cpu: %w", qErr))
	}

	sample.ControlPlaneContainerMem = cpcMem
	sample.ControlPlaneProcessMem = cppMem
	sample.ControlPlaneCPU = cpCPU
	sample.DataPlaneMem = dpMem
	sample.DataPlaneCPU = dpCPU
	return err
}

func (r *BenchmarkReport) sampleProfiles(ctx context.Context, sample *BenchmarkMetricSample) error {
	egPod, err := r.fetchEnvoyGatewayPod(ctx)
	if err != nil {
		return err
	}

	// Memory heap profiles.
	heapProf, err := collect.RequestWithPortForwarder(
		r.kubeClient, types.NamespacedName{Name: egPod.Name, Namespace: egPod.Namespace}, 19000, "/debug/pprof/heap",
	)
	if err != nil {
		return err
	}

	sample.HeapProfile = heapProf
	return nil
}

// getLogsFromPod scrapes the logs directly from the pod (default container).
func (r *BenchmarkReport) getLogsFromPod(ctx context.Context, pod *types.NamespacedName) ([]byte, error) {
	podLogOpts := corev1.PodLogOptions{}

	req := r.kubeClient.Kube().CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *BenchmarkReport) fetchEnvoyGatewayPod(ctx context.Context) (*corev1.Pod, error) {
	egPods, err := r.kubeClient.Kube().CoreV1().
		Pods("envoy-gateway-system").
		List(ctx, metav1.ListOptions{LabelSelector: "control-plane=envoy-gateway"})
	if err != nil {
		return nil, err
	}

	if len(egPods.Items) < 1 {
		return nil, fmt.Errorf("failed to get any pods for envoy-gateway")
	}

	// Using the first one pod as default envoy-gateway pod
	return &egPods.Items[0], nil
}
