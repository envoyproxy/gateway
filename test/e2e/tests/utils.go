// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"
	flog "fortio.org/log"
	"github.com/go-logfmt/logfmt"
	"github.com/gogo/protobuf/jsonpb" // nolint: depguard // tempopb use gogo/protobuf
	"github.com/google/go-cmp/cmp"
	"github.com/grafana/tempo/pkg/tempopb"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const defaultServiceStartupTimeout = 5 * time.Minute

// WaitForPods waits for the pods in the given namespace and with the given selector
// to be in the given phase and condition.
func WaitForPods(t *testing.T, cl client.Client, namespace string, selectors map[string]string, phase corev1.PodPhase, condition corev1.PodCondition) {
	tlog.Logf(t, "waiting for %s/[%s] to be %v...", namespace, selectors, phase)

	require.Eventually(t, func() bool {
		pods := &corev1.PodList{}

		err := cl.List(context.Background(), pods, &client.ListOptions{
			Namespace:     namespace,
			LabelSelector: labels.SelectorFromSet(selectors),
		})

		if err != nil || len(pods.Items) == 0 {
			return false
		}

	checkPods:
		for _, p := range pods.Items {
			if p.Status.Phase != phase {
				return false
			}

			if p.Status.Conditions == nil {
				return false
			}

			for _, c := range p.Status.Conditions {
				if c.Type == condition.Type && c.Status == condition.Status {
					continue checkPods // pod is ready, check next pod
				}
			}

			tlog.Logf(t, "pod %s/%s status: %v", p.Namespace, p.Name, p.Status)
			return false
		}

		return true
	}, defaultServiceStartupTimeout, 2*time.Second)
}

// SecurityPolicyMustBeAccepted waits for the specified SecurityPolicy to be accepted.
func SecurityPolicyMustBeAccepted(t *testing.T, client client.Client, policyName types.NamespacedName, controllerName string, ancestorRef gwapiv1a2.ParentReference) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.SecurityPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching SecurityPolicy: %w", err)
		}

		if policyAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef) {
			tlog.Logf(t, "SecurityPolicy has been accepted: %v", policy)
			return true, nil
		}

		tlog.Logf(t, "SecurityPolicy not yet accepted: %v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for SecurityPolicy to be accepted")
}

// SecurityPolicyMustFail waits for an SecurityPolicy to fail with the specified reason.
func SecurityPolicyMustFail(
	t *testing.T, client client.Client, policyName types.NamespacedName,
	controllerName string, ancestorRef gwapiv1a2.ParentReference, message string,
) {
	t.Helper()

	policy := &egv1a1.SecurityPolicy{}
	waitErr := wait.PollUntilContextTimeout(
		context.Background(), 1*time.Second, 60*time.Second,
		true, func(ctx context.Context) (bool, error) {
			err := client.Get(ctx, policyName, policy)
			if err != nil {
				return false, fmt.Errorf("error fetching SecurityPolicy: %w", err)
			}

			if policyFailAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef, message) {
				tlog.Logf(t, "SecurityPolicy has been failed: %v", policy)
				return true, nil
			}

			return false, nil
		})

	require.NoErrorf(t, waitErr, "error waiting for SecurityPolicy to fail with message: %s policy %v", message, policy)
}

// BackendTrafficPolicyMustBeAccepted waits for the specified BackendTrafficPolicy to be accepted.
func BackendTrafficPolicyMustBeAccepted(t *testing.T, client client.Client, policyName types.NamespacedName, controllerName string, ancestorRef gwapiv1a2.ParentReference) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.BackendTrafficPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching BackendTrafficPolicy: %w", err)
		}

		if policyAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef) {
			return true, nil
		}

		tlog.Logf(t, "BackendTrafficPolicy not yet accepted: %v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for BackendTrafficPolicy to be accepted")
}

// ClientTrafficPolicyMustBeAccepted waits for the specified ClientTrafficPolicy to be accepted.
func ClientTrafficPolicyMustBeAccepted(t *testing.T, client client.Client, policyName types.NamespacedName, controllerName string, ancestorRef gwapiv1a2.ParentReference) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.ClientTrafficPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching ClientTrafficPolicy: %w", err)
		}

		if policyAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef) {
			return true, nil
		}

		tlog.Logf(t, "ClientTrafficPolicy not yet accepted: %v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for ClientTrafficPolicy to be accepted")
}

// AlmostEquals We use a solution similar to istio:
// Given an offset, calculate whether the actual value is within the offset of the expected value
func AlmostEquals(actual, expect, offset int) bool {
	upper := actual + offset
	lower := actual - offset
	if expect < lower || expect > upper {
		return false
	}
	return true
}

// runs a load test with options described in opts
// the done channel is used to notify caller of execution result
// the execution may end due to an external abort or timeout
func runLoadAndWait(t *testing.T, timeoutConfig config.TimeoutConfig, done chan bool, aborter *periodic.Aborter, reqURL string) {
	flog.SetLogLevel(flog.Error)
	opts := fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			QPS: 5000,
			// allow some overhead time for setting up workers and tearing down after restart
			Duration:   timeoutConfig.CreateTimeout + timeoutConfig.CreateTimeout/2,
			NumThreads: 50,
			Stop:       aborter,
			Out:        io.Discard,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: reqURL,
		},
	}

	res, err := fhttp.RunHTTPTest(&opts)
	if err != nil {
		done <- false
		tlog.Logf(t, "failed to create load: %v", err)
	}

	// collect stats
	okReq := res.RetCodes[200]
	totalReq := res.DurationHistogram.Count
	failedReq := totalReq - okReq
	errorReq := res.ErrorsDurationHistogram.Count
	timedOut := res.ActualDuration == opts.Duration
	tlog.Logf(t, "Load completed after %s with %d requests, %d success, %d failures and %d errors", res.ActualDuration, totalReq, okReq, failedReq, errorReq)

	if okReq == totalReq && errorReq == 0 && !timedOut {
		done <- true
	}
	done <- false
}

func policyAcceptedByAncestor(ancestors []gwapiv1a2.PolicyAncestorStatus, controllerName string, ancestorRef gwapiv1a2.ParentReference) bool {
	for _, ancestor := range ancestors {
		if string(ancestor.ControllerName) == controllerName && cmp.Equal(ancestor.AncestorRef, ancestorRef) {
			for _, condition := range ancestor.Conditions {
				if condition.Type == string(gwapiv1a2.PolicyConditionAccepted) && condition.Status == metav1.ConditionTrue {
					return true
				}
			}
		}
	}
	return false
}

// EnvoyExtensionPolicyMustFail waits for an EnvoyExtensionPolicy to fail with the specified reason.
func EnvoyExtensionPolicyMustFail(
	t *testing.T, client client.Client, policyName types.NamespacedName,
	controllerName string, ancestorRef gwapiv1a2.ParentReference, message string,
) {
	t.Helper()

	policy := &egv1a1.EnvoyExtensionPolicy{}
	waitErr := wait.PollUntilContextTimeout(
		context.Background(), 1*time.Second, 60*time.Second,
		true, func(ctx context.Context) (bool, error) {
			err := client.Get(ctx, policyName, policy)
			if err != nil {
				return false, fmt.Errorf("error fetching EnvoyExtensionPolicy: %w", err)
			}

			if policyFailAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef, message) {
				tlog.Logf(t, "EnvoyExtensionPolicy has been failed: %v", policy)
				return true, nil
			}

			return false, nil
		})

	require.NoErrorf(t, waitErr, "error waiting for EnvoyExtensionPolicy to fail with message: %s policy %v", message, policy)
}

func policyFailAcceptedByAncestor(ancestors []gwapiv1a2.PolicyAncestorStatus, controllerName string, ancestorRef gwapiv1a2.ParentReference, message string) bool {
	for _, ancestor := range ancestors {
		if string(ancestor.ControllerName) == controllerName && cmp.Equal(ancestor.AncestorRef, ancestorRef) {
			for _, condition := range ancestor.Conditions {
				if condition.Type == string(gwapiv1a2.PolicyConditionAccepted) &&
					condition.Status == metav1.ConditionFalse &&
					strings.Contains(condition.Message, message) {
					return true
				}
			}
		}
	}
	return false
}

// EnvoyExtensionPolicyMustBeAccepted waits for the specified EnvoyExtensionPolicy to be accepted.
func EnvoyExtensionPolicyMustBeAccepted(t *testing.T, client client.Client, policyName types.NamespacedName, controllerName string, ancestorRef gwapiv1a2.ParentReference) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.EnvoyExtensionPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching EnvoyExtensionPolicy: %w", err)
		}

		if policyAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef) {
			tlog.Logf(t, "EnvoyExtensionPolicy has been accepted: %v", policy)
			return true, nil
		}

		tlog.Logf(t, "EnvoyExtensionPolicy not yet accepted: %v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for EnvoyExtensionPolicy to be accepted")
}

func ScrapeMetrics(t *testing.T, c client.Client, nn types.NamespacedName, port int32, path string) error {
	url, err := RetrieveURL(c, nn, port, path)
	if err != nil {
		return err
	}

	tlog.Logf(t, "scraping metrics from %s", url)

	metrics, err := RetrieveMetrics(url, time.Second)
	if err != nil {
		return err
	}

	// TODO: support metric matching
	// for now, just check metric exists
	if len(metrics) > 0 {
		return nil
	}

	return errors.New("no metrics found")
}

func RetrieveURL(c client.Client, nn types.NamespacedName, port int32, path string) (string, error) {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), nn, &svc); err != nil {
		return "", err
	}
	host := ""
	switch svc.Spec.Type {
	case corev1.ServiceTypeLoadBalancer:
		for _, ing := range svc.Status.LoadBalancer.Ingress {
			if ing.IP != "" {
				host = ing.IP
				break
			}
		}
	default:
		host = fmt.Sprintf("%s.%s.svc", nn.Name, nn.Namespace)
	}
	return fmt.Sprintf("http://%s:%d%s", host, port, path), nil
}

var metricParser = &expfmt.TextParser{}

func RetrieveMetrics(url string, timeout time.Duration) (map[string]*dto.MetricFamily, error) {
	httpClient := http.Client{
		Timeout: timeout,
	}
	res, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape metrics: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to scrape metrics: %s", res.Status)
	}

	return metricParser.TextToMetricFamilies(res.Body)
}

func RetrieveMetric(url string, name string, timeout time.Duration) (*dto.MetricFamily, error) {
	metrics, err := RetrieveMetrics(url, timeout)
	if err != nil {
		return nil, err
	}

	if mf, ok := metrics[name]; ok {
		return mf, nil
	}

	return nil, nil
}

func WaitForLoadBalancerAddress(t *testing.T, client client.Client, timeout time.Duration, nn types.NamespacedName) (string, error) {
	t.Helper()

	var ipAddr string
	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		s := &corev1.Service{}
		err := client.Get(ctx, nn, s)
		if err != nil {
			tlog.Logf(t, "error fetching Service: %v", err)
			return false, fmt.Errorf("error fetching Service: %w", err)
		}

		if len(s.Status.LoadBalancer.Ingress) > 0 {
			ipAddr = s.Status.LoadBalancer.Ingress[0].IP
			return true, nil
		}
		return false, nil
	})
	require.NoErrorf(t, waitErr, "error waiting for Service to have at least one load balancer IP address in status")
	return ipAddr, nil
}

func ALSLogCount(suite *suite.ConformanceTestSuite) (int, error) {
	metricPath, err := RetrieveURL(suite.Client, types.NamespacedName{
		Namespace: "monitoring",
		Name:      "envoy-als",
	}, 19001, "/metrics")
	if err != nil {
		return -1, err
	}

	countMetric, err := RetrieveMetric(metricPath, "log_count", time.Second)
	if err != nil {
		return -1, err
	}

	// metric not found or empty
	if countMetric == nil {
		return 0, nil
	}

	total := 0
	for _, m := range countMetric.Metric {
		if m.Counter != nil && m.Counter.Value != nil {
			total += int(*m.Counter.Value)
		}
	}

	return total, nil
}

// QueryLogCountFromLoki queries log count from loki
func QueryLogCountFromLoki(t *testing.T, c client.Client, keyValues map[string]string, match string) (int, error) {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "monitoring",
		Name:      "loki",
	}, &svc); err != nil {
		return -1, err
	}
	lokiHost := ""
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			lokiHost = ing.IP
			break
		}
	}

	qParams := make([]string, 0, len(keyValues))
	for k, v := range keyValues {
		qParams = append(qParams, fmt.Sprintf("%s=\"%s\"", k, v))
	}

	q := "{" + strings.Join(qParams, ",") + "}"
	if match != "" {
		q = q + "|~\"" + match + "\""
	}
	params := url.Values{}
	params.Add("query", q)
	params.Add("start", fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())) // query logs from last 10 minutes
	lokiQueryURL := fmt.Sprintf("http://%s:3100/loki/api/v1/query_range?%s", lokiHost, params.Encode())
	res, err := http.DefaultClient.Get(lokiQueryURL)
	if err != nil {
		return -1, err
	}
	tlog.Logf(t, "get response from loki, query=%s, status=%s", q, res.Status)

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return -1, err
	}

	lokiResponse := &LokiQueryResponse{}
	if err := json.Unmarshal(b, lokiResponse); err != nil {
		return -1, err
	}

	if len(lokiResponse.Data.Result) == 0 {
		return 0, nil
	}

	total := 0
	for _, res := range lokiResponse.Data.Result {
		total += len(res.Values)
	}
	tlog.Logf(t, "get response from loki, query=%s, total=%d", q, total)
	return total, nil
}

type LokiQueryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric interface{}
			Values []interface{} `json:"values"`
		}
	}
}

// QueryTraceFromTempo queries span count from tempo
func QueryTraceFromTempo(t *testing.T, c client.Client, tags map[string]string) (int, error) {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), types.NamespacedName{
		Namespace: "monitoring",
		Name:      "tempo",
	}, &svc); err != nil {
		return -1, err
	}
	host := ""
	for _, ing := range svc.Status.LoadBalancer.Ingress {
		if ing.IP != "" {
			host = ing.IP
			break
		}
	}

	tagsQueryParam, err := createTagsQueryParam(tags)
	if err != nil {
		return -1, err
	}

	tempoURL := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, "3100"),
		Path:   "/api/search",
	}
	query := tempoURL.Query()
	query.Add("start", fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())) // query traces from last 10 minutes
	query.Add("end", fmt.Sprintf("%d", time.Now().Unix()))
	query.Add("tags", tagsQueryParam)
	tempoURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", tempoURL.String(), nil)
	if err != nil {
		return -1, err
	}

	tlog.Logf(t, "send request to %s", tempoURL.String())
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}

	if res.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("failed to query tempo, url=%s, status=%s", tempoURL.String(), res.Status)
	}

	tempoResponse := &tempopb.SearchResponse{}
	if err := jsonpb.Unmarshal(res.Body, tempoResponse); err != nil {
		return -1, err
	}

	total := len(tempoResponse.Traces)
	tlog.Logf(t, "get response from tempo, url=%s, response=%v, total=%d", tempoURL.String(), tempoResponse, total)
	return total, nil
}

// copy from https://github.com/grafana/tempo/blob/c0127c78c368319433c7c67ca8967adbfed2259e/cmd/tempo-query/tempo/plugin.go#L361
func createTagsQueryParam(tags map[string]string) (string, error) {
	tagsBuilder := &strings.Builder{}
	tagsEncoder := logfmt.NewEncoder(tagsBuilder)
	for k, v := range tags {
		err := tagsEncoder.EncodeKeyval(k, v)
		if err != nil {
			return "", err
		}
	}
	return tagsBuilder.String(), nil
}
