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
	"os"
	"path"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	localMetricsPort  = 0
	localProfilesPort = 0
	podMetricsPort    = 19001
	podProfilesPort   = 19000
)

type BenchmarkReport struct {
	Name         string
	RawResult    []byte
	RawCPMetrics []byte
	RawDPMetrics map[string][]byte

	ProfilesOutputDir string
	ProfilesPath      map[string]string

	kubeClient kube.CLIClient
}

func NewBenchmarkReport(name, profilesOutputDir string) (*BenchmarkReport, error) {
	kubeClient, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, err
	}

	return &BenchmarkReport{
		Name:              name,
		RawDPMetrics:      make(map[string][]byte),
		ProfilesPath:      make(map[string]string),
		ProfilesOutputDir: profilesOutputDir,
		kubeClient:        kubeClient,
	}, nil
}

// Print prints the raw report of one benchmark test.
func (r *BenchmarkReport) Print(t *testing.T, name string) {
	t.Logf("The raw report of benchmark test: %s", name)

	t.Logf("=== Benchmark Result: \n\n %s \n\n", r.RawResult)
	t.Logf("=== Control-Plane Metrics: \n\n %s \n\n", r.RawCPMetrics)

	for dpName, dpMetrics := range r.RawDPMetrics {
		t.Logf("=== Data-Plane Metrics for %s: \n\n %s \n\n", dpName, dpMetrics)
	}
}

func (r *BenchmarkReport) Collect(t *testing.T, ctx context.Context, job *types.NamespacedName) error {
	if err := r.GetResults(t, ctx, job); err != nil {
		return err
	}

	if err := r.GetControlPlaneMetrics(t, ctx); err != nil {
		return err
	}

	if err := r.GetDataPlaneMetrics(t, ctx); err != nil {
		return err
	}

	if err := r.GetProfiles(t, ctx); err != nil {
		return err
	}

	return nil
}

func (r *BenchmarkReport) GetResults(t *testing.T, ctx context.Context, job *types.NamespacedName) error {
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
	egPod, err := r.fetchEnvoyGatewayPod(t, ctx)
	if err != nil {
		return err
	}

	metrics, err := r.getResponseFromPortForwarder(
		t, &types.NamespacedName{Name: egPod.Name, Namespace: egPod.Namespace}, localMetricsPort, podMetricsPort, "/metrics",
	)
	if err != nil {
		return err
	}

	r.RawCPMetrics = metrics

	return nil
}

func (r *BenchmarkReport) GetDataPlaneMetrics(t *testing.T, ctx context.Context) error {
	epPods, err := r.fetchEnvoyProxyPodList(t, ctx)
	if err != nil {
		return err
	}

	for _, epPod := range epPods.Items {
		podNN := &types.NamespacedName{Name: epPod.Name, Namespace: epPod.Namespace}
		metrics, err := r.getResponseFromPortForwarder(t, podNN, localMetricsPort, podMetricsPort, "/stats/prometheus")
		if err != nil {
			return err
		}

		r.RawDPMetrics[podNN.String()] = metrics
	}

	return nil
}

func (r *BenchmarkReport) GetProfiles(t *testing.T, ctx context.Context) error {
	egPod, err := r.fetchEnvoyGatewayPod(t, ctx)
	if err != nil {
		return err
	}

	// Memory heap profiles.
	heapProf, err := r.getResponseFromPortForwarder(
		t, &types.NamespacedName{Name: egPod.Name, Namespace: egPod.Namespace}, localProfilesPort, podProfilesPort, "/debug/pprof/heap",
	)
	if err != nil {
		return err
	}

	heapProfPath := path.Join(r.ProfilesOutputDir, fmt.Sprintf("heap.%s.pprof", r.Name))
	if err = os.WriteFile(heapProfPath, heapProf, 0644); err != nil {
		return fmt.Errorf("failed to write profiles %s: %v", heapProfPath, err)
	}
	r.ProfilesPath["heap"] = heapProfPath

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

// getResponseFromPortForwarder gets response by sending endpoint request to pod port-forwarder.
func (r *BenchmarkReport) getResponseFromPortForwarder(t *testing.T, pod *types.NamespacedName, localPort, podPort int, endpoint string) ([]byte, error) {
	fw, err := kube.NewLocalPortForwarder(r.kubeClient, *pod, localPort, podPort)
	if err != nil {
		return nil, fmt.Errorf("failed to build port forwarder for pod %s: %v", pod.String(), err)
	}

	if err = fw.Start(); err != nil {
		fw.Stop()

		return nil, fmt.Errorf("failed to start port forwarder for pod %s: %v", pod.String(), err)
	}

	var out []byte
	// Retrieving response by requesting Pod with url endpoint.
	go func() {
		defer fw.Stop()

		url := fmt.Sprintf("http://%s%s", fw.Address(), endpoint)
		resp, err := http.Get(url)
		if err != nil {
			t.Errorf("failed to request %s: %v", url, err)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("failed to read response: %v", err)
			return
		}

		out = body
	}()

	fw.WaitForStop()

	return out, nil
}

func (r *BenchmarkReport) fetchEnvoyGatewayPod(t *testing.T, ctx context.Context) (*corev1.Pod, error) {
	egPods, err := r.kubeClient.Kube().CoreV1().
		Pods("envoy-gateway-system").
		List(ctx, metav1.ListOptions{LabelSelector: "control-plane=envoy-gateway"})
	if err != nil {
		return nil, err
	}

	if len(egPods.Items) < 1 {
		return nil, fmt.Errorf("failed to get any pods for envoy-gateway")
	}

	if len(egPods.Items) > 1 {
		t.Logf("Got %d pod(s), using the first one as default envoy-gateway pod", len(egPods.Items))
	}

	return &egPods.Items[0], nil
}

func (r *BenchmarkReport) fetchEnvoyProxyPodList(t *testing.T, ctx context.Context) (*corev1.PodList, error) {
	epPods, err := r.kubeClient.Kube().CoreV1().
		Pods("envoy-gateway-system").
		List(ctx, metav1.ListOptions{LabelSelector: "gateway.envoyproxy.io/owning-gateway-namespace=benchmark-test,gateway.envoyproxy.io/owning-gateway-name=benchmark"})
	if err != nil {
		return nil, err
	}

	if len(epPods.Items) < 1 {
		return nil, fmt.Errorf("failed to get any pods for envoy-proxies")
	}

	t.Logf("Got %d pod(s) from data-plane", len(epPods.Items))

	return epPods, nil
}
