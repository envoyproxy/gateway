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
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	localMetricsPort        = 0
	controlPlaneMetricsPort = 19001
)

type BenchmarkReport struct {
	Name         string
	RawResult    []byte
	RawCPMetrics []byte

	kubeClient kube.CLIClient
}

func NewBenchmarkReport(name string) (*BenchmarkReport, error) {
	kubeClient, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, err
	}

	return &BenchmarkReport{
		Name:       name,
		kubeClient: kubeClient,
	}, nil
}

// Print prints the raw report of one benchmark test.
func (r *BenchmarkReport) Print(t *testing.T, name string) {
	t.Logf("The raw report of benchmark test: %s", name)

	t.Logf("=== Benchmark Result: \n\n %s \n\n", r.RawResult)
	t.Logf("=== Control-Plane Metrics: \n\n %s \n\n", r.RawCPMetrics)
}

func (r *BenchmarkReport) GetBenchmarkResult(t *testing.T, ctx context.Context, job *types.NamespacedName) error {
	pods, err := r.kubeClient.Kube().CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + job.Name})

	if len(pods.Items) < 1 {
		return fmt.Errorf("failed to get any pods for job %s", job.String())
	}

	if len(pods.Items) > 1 {
		t.Logf("Got %d pod(s) associated job %s, should be 1 pod, could be pod err and job backoff then restart, please check your pod(s) status",
			len(pods.Items), job.Name)
	}

	pod := &pods.Items[0]
	logs, err := r.getLogsFromPod(
		ctx, &types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace},
	)
	if err != nil {
		return err
	}

	r.RawResult = logs

	return nil
}

func (r *BenchmarkReport) GetControlPlaneMetrics(t *testing.T, ctx context.Context) error {
	egPods, err := r.kubeClient.Kube().CoreV1().Pods("envoy-gateway-system").
		List(ctx, metav1.ListOptions{LabelSelector: "control-plane=envoy-gateway"})
	if err != nil {
		return err
	}

	if len(egPods.Items) < 1 {
		return fmt.Errorf("failed to get any pods for envoy-gateway")
	}

	if len(egPods.Items) > 1 {
		t.Logf("Got %d pod(s), using the first one as default envoy-gateway pod", len(egPods.Items))
	}

	egPod := &egPods.Items[0]
	metrics, err := r.getMetricsFromPortForwarder(
		t, &types.NamespacedName{Name: egPod.Name, Namespace: egPod.Namespace},
	)
	if err != nil {
		return err
	}

	r.RawCPMetrics = metrics

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

// getMetricsFromPortForwarder retrieves metrics from pod by request url `/metrics`.
func (r *BenchmarkReport) getMetricsFromPortForwarder(t *testing.T, pod *types.NamespacedName) ([]byte, error) {
	fw, err := kube.NewLocalPortForwarder(r.kubeClient, *pod, localMetricsPort, controlPlaneMetricsPort)
	if err != nil {
		return nil, fmt.Errorf("failed to build port forwarder for pod %s: %v", pod.String(), err)
	}

	if err = fw.Start(); err != nil {
		fw.Stop()

		return nil, fmt.Errorf("failed to start port forwarder for pod %s: %v", pod.String(), err)
	}

	var out []byte
	// Retrieving metrics from Pod.
	go func() {
		defer fw.Stop()

		url := fmt.Sprintf("http://%s/metrics", fw.Address())
		resp, err := http.Get(url)
		if err != nil {
			t.Errorf("failed to request %s: %v", url, err)
			return
		}

		metrics, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("failed to read metrics: %v", err)
			return
		}

		out = metrics
	}()

	fw.WaitForStop()

	return out, nil
}
