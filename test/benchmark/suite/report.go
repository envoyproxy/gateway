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
	RawResult  []byte
	RawMetrics []byte

	kubeClient kube.CLIClient
}

func NewBenchmarkReport() (*BenchmarkReport, error) {
	kubeClient, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, err
	}

	return &BenchmarkReport{
		kubeClient: kubeClient,
	}, nil
}

func (r *BenchmarkReport) Print(t *testing.T, name string) {
	t.Logf("The report of benchmark test: %s", name)

	t.Logf("=== Benchmark Result: \n\n %s \n\n", r.RawResult)
	t.Logf("=== Control-Plane Metrics: \n\n %s \n\n", r.RawMetrics)
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
	if err = r.getBenchmarkResultFromPodLogs(
		ctx, &types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace},
	); err != nil {
		return err
	}

	return nil
}

func (r *BenchmarkReport) GetControlPlaneMetrics(t *testing.T, ctx context.Context) error {
	egPods, err := r.kubeClient.Kube().CoreV1().Pods("envoy-gateway-system").List(ctx, metav1.ListOptions{LabelSelector: "control-plane=envoy-gateway"})
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
	if err = r.getMetricsFromPodPortForwarder(
		t, &types.NamespacedName{Name: egPod.Name, Namespace: egPod.Namespace},
	); err != nil {
		return err
	}

	return nil
}

// getBenchmarkResultFromPodLogs scrapes the logs directly from the pod (default container)
// and save it as the raw result in benchmark report.
func (r *BenchmarkReport) getBenchmarkResultFromPodLogs(ctx context.Context, pod *types.NamespacedName) error {
	podLogOpts := corev1.PodLogOptions{}

	req := r.kubeClient.Kube().CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return err
	}

	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return err
	}

	r.RawResult = buf.Bytes()
	return nil
}

func (r *BenchmarkReport) getMetricsFromPodPortForwarder(t *testing.T, pod *types.NamespacedName) error {
	fw, err := kube.NewLocalPortForwarder(r.kubeClient, *pod, localMetricsPort, controlPlaneMetricsPort)
	if err != nil {
		return fmt.Errorf("failed to build port forwarder for pod %s: %v", pod.String(), err)
	}

	if err = fw.Start(); err != nil {
		fw.Stop()
		return fmt.Errorf("failed to start port forwarder for pod %s: %v", pod.String(), err)
	}

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

		r.RawMetrics = metrics
	}()

	fw.WaitForStop()

	return nil
}
