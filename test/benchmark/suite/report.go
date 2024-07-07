// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build benchmark
// +build benchmark

package suite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	prom "github.com/envoyproxy/gateway/test/utils/prometheus"
)

type BenchmarkReport struct {
	Name    string
	Result  []byte
	Metrics map[string]float64 // metricTableHeaderName:metricValue

	kubeClient kube.CLIClient
	promClient *prom.Client
}

func NewBenchmarkReport(name string, kubeClient kube.CLIClient, promClient *prom.Client) *BenchmarkReport {
	return &BenchmarkReport{
		Name:       name,
		Metrics:    make(map[string]float64),
		kubeClient: kubeClient,
		promClient: promClient,
	}
}

func (r *BenchmarkReport) Collect(ctx context.Context, job *types.NamespacedName) error {
	r.GetMetrics(ctx)

	if err := r.GetResult(ctx, job); err != nil {
		return err
	}

	return nil
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
	var pod *corev1.Pod
	for _, p := range pods.Items {
		if p.Status.Phase == corev1.PodSucceeded {
			pod = &p
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

func (r *BenchmarkReport) GetMetrics(ctx context.Context) {
	for _, h := range metricsTableHeader {
		if len(h.promQL) == 0 {
			continue
		}

		v, err := r.promClient.QuerySum(ctx, h.promQL)
		if err == nil {
			r.Metrics[h.name], _ = strconv.ParseFloat(fmt.Sprintf("%.2f", v), 64)
		}
	}
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
