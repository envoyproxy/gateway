// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package collect

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	troubleshootv1b2 "github.com/replicatedhq/troubleshoot/pkg/apis/troubleshoot/v1beta2"
	tbcollect "github.com/replicatedhq/troubleshoot/pkg/collect"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

var _ tbcollect.Collector = &PrometheusMetric{}

// PrometheusMetric defines a collector scraping Prometheus metrics from the selected pods.
type PrometheusMetric struct {
	BundlePath   string
	Namespace    string
	ClientConfig *rest.Config
}

func (p PrometheusMetric) Title() string {
	return "prometheus-metric"
}

func (p PrometheusMetric) IsExcluded() (bool, error) {
	return false, nil
}

func (p PrometheusMetric) GetRBACErrors() []error {
	return nil
}

func (p PrometheusMetric) HasRBACErrors() bool {
	return false
}

func (p PrometheusMetric) CheckRBAC(_ context.Context, _ tbcollect.Collector, _ *troubleshootv1b2.Collect, _ *rest.Config, _ string) error {
	return nil
}

func (p PrometheusMetric) Collect(_ chan<- interface{}) (tbcollect.CollectorResult, error) {
	client, err := kubernetes.NewForConfig(p.ClientConfig)
	if err != nil {
		return nil, err
	}

	pods, err := listPods(context.TODO(), client, p.Namespace, labels.Everything())
	if err != nil {
		return nil, err
	}

	output := tbcollect.NewResult()

	cliClient, err := kube.NewForRestConfig(p.ClientConfig)
	if err != nil {
		return output, err
	}

	logs := make([]string, 0)
	for _, pod := range pods {
		annos := pod.GetAnnotations()
		if v, ok := annos["prometheus.io/scrape"]; !ok || v != "true" {
			logs = append(logs, fmt.Sprintf("pod %s/%s is skipped because of missing annotation prometheus.io/scrape", pod.Namespace, pod.Name))
			continue
		}

		nn, port, reqPath := types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}, 19001, "/metrics"
		if v, ok := annos["prometheus.io/port"]; !ok {
			port, err = strconv.Atoi(v)
			if err != nil {
				logs = append(logs, fmt.Sprintf("pod %s/%s is skipped because of invalid prometheus.io/port", pod.Namespace, pod.Name))
				continue
			}
		}
		if v, ok := annos["prometheus.io/path"]; ok {
			reqPath = v
		}

		data, err := RequestWithPortForwarder(cliClient, nn, port, reqPath)
		if err != nil {
			logs = append(logs, fmt.Sprintf("pod %s/%s is skipped because of err: %v", pod.Namespace, pod.Name, err))
			continue
		}

		k := fmt.Sprintf("%s-%s.prom", pod.Namespace, pod.Name)
		_ = output.SaveResult(p.BundlePath, path.Join("prometheus-metrics", k), bytes.NewBuffer(data))
	}
	if len(logs) > 0 {
		_ = output.SaveResult(p.BundlePath, path.Join("prometheus-metrics", "error.log"), bytes.NewBuffer([]byte(strings.Join(logs, "\n"))))
	}

	return output, nil
}

func listPods(ctx context.Context, client kubernetes.Interface, namespace string, selector labels.Selector) ([]corev1.Pod, error) {
	pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func RequestWithPortForwarder(cli kube.CLIClient, nn types.NamespacedName, port int, reqPath string) ([]byte, error) {
	fw, err := kube.NewLocalPortForwarder(cli, nn, 0, port)
	if err != nil {
		return nil, err
	}

	if err := fw.Start(); err != nil {
		return nil, err
	}
	defer fw.Stop()

	if !strings.HasPrefix(reqPath, "/") {
		reqPath = "/" + reqPath
	}

	url := fmt.Sprintf("http://%s%s", fw.Address(), reqPath)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return io.ReadAll(resp.Body)
}
