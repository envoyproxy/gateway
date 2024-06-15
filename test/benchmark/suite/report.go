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
	"io"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type BenchmarkReport struct {
	RawResult []byte
}

func NewBenchmarkReport() *BenchmarkReport {
	return &BenchmarkReport{}
}

func (r *BenchmarkReport) Print(t *testing.T, name string) {
	t.Logf("The report of benchmark test: %s", name)

	t.Logf("=== Benchmark Result: \n %s \n", string(r.RawResult))
}

func (r *BenchmarkReport) GetResultFromJob(t *testing.T, ctx context.Context, job *types.NamespacedName) error {
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}

	pods, err := clientset.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{LabelSelector: "job-name=" + job.Name})
	if len(pods.Items) > 1 {
		t.Logf("Got %d pod(s) associated job %s, should be 1 pod, could be pod err and job backoff then restart, please check your pod(s) status",
			len(pods.Items), job.Name)
	}

	pod := &pods.Items[0]
	if err = r.getResultFromPodLogs(ctx, clientset,
		&types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}); err != nil {
		return err
	}

	return nil
}

// getResultFromPodLogs scrapes the logs directly from the pod (default container)
// and save it as the raw result in benchmark report.
func (r *BenchmarkReport) getResultFromPodLogs(ctx context.Context, clientset *kubernetes.Clientset, pod *types.NamespacedName) error {
	podLogOpts := corev1.PodLogOptions{}

	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
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
