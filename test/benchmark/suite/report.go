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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	prom "github.com/envoyproxy/gateway/test/utils/prometheus"
)

type BenchmarkReport struct {
	Name              string
	Result            []byte
	Metrics           map[string]float64 // metricTableHeaderName:metricValue
	ProfilesPath      map[string]string  // profileKey:profileFilepath
	ProfilesOutputDir string

	kubeClient kube.CLIClient
	promClient *prom.Client
}

func NewBenchmarkReport(name, profilesOutputDir string, kubeClient kube.CLIClient, promClient *prom.Client) (*BenchmarkReport, error) {
	if _, err := createDirIfNotExist(profilesOutputDir); err != nil {
		return nil, err
	}

	return &BenchmarkReport{
		Name:              name,
		Metrics:           make(map[string]float64),
		ProfilesPath:      make(map[string]string),
		ProfilesOutputDir: profilesOutputDir,
		kubeClient:        kubeClient,
		promClient:        promClient,
	}, nil
}

func (r *BenchmarkReport) Collect(ctx context.Context, job *types.NamespacedName) error {
	if err := r.GetProfiles(ctx); err != nil {
		return err
	}

	if err := r.GetMetrics(ctx); err != nil {
		return err
	}

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
	var pod corev1.Pod
	for _, p := range pods.Items {
		if p.Status.Phase == corev1.PodSucceeded {
			pod = p
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

func (r *BenchmarkReport) GetMetrics(ctx context.Context) error {
	for _, h := range metricsTableHeader {
		if len(h.promQL) == 0 {
			continue
		}

		var (
			v   float64
			err error
		)
		switch h.queryType {
		case querySum:
			v, err = r.promClient.QuerySum(ctx, h.promQL)
		case queryAvg:
			v, err = r.promClient.QueryAvg(ctx, h.promQL)
		default:
			return fmt.Errorf("unsupported query type: %s", h.queryType)
		}

		if err == nil {
			r.Metrics[h.name], _ = strconv.ParseFloat(fmt.Sprintf("%.2f", v), 64)
		}
	}

	return nil
}

func (r *BenchmarkReport) GetProfiles(ctx context.Context) error {
	egPod, err := r.fetchEnvoyGatewayPod(ctx)
	if err != nil {
		return err
	}

	// Memory heap profiles. TODO: make the port of pod configurable if it's feasible.
	heapProf, err := r.getResponseFromPortForwarder(
		&types.NamespacedName{Name: egPod.Name, Namespace: egPod.Namespace}, 0, 19000, "/debug/pprof/heap",
	)
	if err != nil {
		return err
	}

	heapProfPath := path.Join(r.ProfilesOutputDir, fmt.Sprintf("heap.%s.pprof", r.Name))
	if err = os.WriteFile(heapProfPath, heapProf, 0644); err != nil {
		return fmt.Errorf("failed to write profiles %s: %v", heapProfPath, err)
	}

	// Remove parent output report dir.
	splits := strings.SplitN(heapProfPath, "/", 2)[0]
	heapProfPath = strings.TrimPrefix(heapProfPath, splits+"/")
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
func (r *BenchmarkReport) getResponseFromPortForwarder(pod *types.NamespacedName, localPort, podPort int, endpoint string) ([]byte, error) {
	fw, err := kube.NewLocalPortForwarder(r.kubeClient, *pod, localPort, podPort)
	if err != nil {
		return nil, fmt.Errorf("failed to build port forwarder for pod %s: %v", pod.String(), err)
	}

	if err = fw.Start(); err != nil {
		fw.Stop()
		return nil, fmt.Errorf("failed to start port forwarder for pod %s: %v", pod.String(), err)
	}

	var (
		out     []byte
		respErr error
	)
	// Retrieving response by requesting Pod with url endpoint.
	go func() {
		defer fw.Stop()

		var resp *http.Response
		url := fmt.Sprintf("http://%s%s", fw.Address(), endpoint)
		resp, respErr = http.Get(url)
		if respErr == nil {
			out, _ = io.ReadAll(resp.Body)
		}
	}()

	fw.WaitForStop()

	if respErr != nil {
		return nil, respErr
	}
	return out, nil
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
