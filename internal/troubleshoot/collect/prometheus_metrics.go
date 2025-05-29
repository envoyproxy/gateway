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
	"github.com/envoyproxy/gateway/internal/utils/str"
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

		scrape, reqPath, port, err := getPrometheusPathAndPort(&pod)
		if err != nil {
			logs = append(logs, fmt.Sprintf("pod %s/%s is skipped because of err: %v", pod.Namespace, pod.Name, err))
		}
		if !scrape {
			logs = append(logs, fmt.Sprintf("pod %s/%s is skipped because of annotation prometheus.io/scrape=false", pod.Namespace, pod.Name))
			continue
		}

		data, err := RequestWithPortForwarder(cliClient, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, port, reqPath)
		if err != nil {
			logs = append(logs, fmt.Sprintf("pod %s/%s:%v%s is skipped because of err: %v", pod.Namespace, pod.Name, port, reqPath, err))
			continue
		}

		_ = output.SaveResult(p.BundlePath, path.Join("prometheus-metrics", pod.Namespace, fmt.Sprintf("%s.prom", pod.Name)), bytes.NewBuffer(data))
	}
	if len(logs) > 0 {
		_ = output.SaveResult(p.BundlePath, path.Join("prometheus-metrics", "error.log"), bytes.NewBuffer([]byte(strings.Join(logs, "\n"))))
	}

	return output, nil
}

func getPrometheusPathAndPort(pod *corev1.Pod) (bool, string, int, error) {
	reqPath := "/metrics"
	port := 9090
	scrape := false
	annotations := pod.GetAnnotations()
	for k, v := range annotations {
		switch str.SanitizeLabelName(k) {
		case "prometheus_io_scrape":
			if v != "true" {
				return false, "", 0, fmt.Errorf("pod %s/%s is skipped because of missing annotation prometheus.io/scrape", pod.Namespace, pod.Name)
			}
			scrape = true
		case "prometheus_io_port":
			p, err := strconv.Atoi(v)
			if err != nil {
				return false, "", 0, fmt.Errorf("failed to parse port from annotation: %w", err)
			}
			port = p
		case "prometheus_io_path":
			reqPath = v
		}
	}

	return scrape, reqPath, port, nil
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
