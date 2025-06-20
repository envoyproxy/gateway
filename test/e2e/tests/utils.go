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
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"
	flog "fortio.org/log"
	"github.com/google/go-cmp/cmp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	k8sutils "sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/kubernetes"
	tb "github.com/envoyproxy/gateway/internal/troubleshoot"
)

var (
	IPFamily      = os.Getenv("IP_FAMILY")
	DeployProfile = os.Getenv("KUBE_DEPLOY_PROFILE")

	SameNamespaceGateway    = types.NamespacedName{Name: "same-namespace", Namespace: ConformanceInfraNamespace}
	SameNamespaceGatewayRef = k8sutils.NewGatewayRef(SameNamespaceGateway)

	PodReady = corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}
)

const (
	ConformanceInfraNamespace = "gateway-conformance-infra"

	defaultServiceStartupTimeout = 5 * time.Minute
)

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

// BackendTrafficPolicyMustFail waits for an BackendTrafficPolicy to fail with the specified reason.
func BackendTrafficPolicyMustFail(
	t *testing.T, client client.Client, policyName types.NamespacedName,
	controllerName string, ancestorRef gwapiv1a2.ParentReference, message string,
) {
	t.Helper()

	policy := &egv1a1.BackendTrafficPolicy{}
	waitErr := wait.PollUntilContextTimeout(
		context.Background(), 1*time.Second, 60*time.Second,
		true, func(ctx context.Context) (bool, error) {
			err := client.Get(ctx, policyName, policy)
			if err != nil {
				return false, fmt.Errorf("error fetching BackendTrafficPolicy: %w", err)
			}

			if policyFailAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef, message) {
				t.Logf("BackendTrafficPolicy has been failed: %v", policy)
				return true, nil
			}

			return false, nil
		})

	require.NoErrorf(t, waitErr, "error waiting for BackendTrafficPolicy to fail with message: %s policy %v", message, policy)
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
			tlog.Logf(t, "EnvoyExtensionPolicy has been accepted: %+v", policy)
			return true, nil
		}

		tlog.Logf(t, "EnvoyExtensionPolicy not yet accepted: %+v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for EnvoyExtensionPolicy to be accepted")
}

// BackendMustBeAccepted waits for the specified Backend to be accepted.
func BackendMustBeAccepted(t *testing.T, client client.Client, backendName types.NamespacedName) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		backend := &egv1a1.Backend{}
		err := client.Get(ctx, backendName, backend)
		if err != nil {
			return false, fmt.Errorf("error fetching Backend: %w", err)
		}

		for _, condition := range backend.Status.Conditions {
			if condition.Type == string(egv1a1.BackendConditionAccepted) && condition.Status == metav1.ConditionTrue {
				return true, nil
			}
		}

		tlog.Logf(t, "Backend not yet accepted: %v", backend)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for Backend to be accepted")
}

// ScrapeMetrics
// TODO: use QueryPrometheus from test/e2e/tests/promql.go instead
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
	host, err := ServiceHost(c, nn, port)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s%s", host, path), nil
}

func ServiceHost(c client.Client, nn types.NamespacedName, port int32) (string, error) {
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

	return net.JoinHostPort(host, strconv.Itoa(int(port))), nil
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

func RetrieveMetric(url, name string, timeout time.Duration) (*dto.MetricFamily, error) {
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

func OverLimitCount(suite *suite.ConformanceTestSuite) (int, error) {
	cli, err := kubernetes.NewForRestConfig(suite.RestConfig)
	if err != nil {
		return -1, err
	}

	pods, err := cli.PodsForSelector("envoy-gateway-system", "app.kubernetes.io/name=envoy-ratelimit")
	if err != nil {
		return -1, err
	}

	if len(pods.Items) == 0 {
		return -1, fmt.Errorf("no envoy-ratelimit pod found")
	}

	fwd, err := kubernetes.NewLocalPortForwarder(cli, types.NamespacedName{
		Namespace: "envoy-gateway-system",
		Name:      pods.Items[0].Name,
	}, 0, 19001)
	if err != nil {
		return -1, err
	}
	if err := fwd.Start(); err != nil {
		return -1, err
	}
	defer fwd.Stop()

	countMetric, err := RetrieveMetric(fmt.Sprintf("http://%s/metrics", fwd.Address()), "ratelimit_service_rate_limit_over_limit", time.Second)
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
	lokiQueryURL := fmt.Sprintf("http://%s/loki/api/v1/query_range?%s", net.JoinHostPort(lokiHost, "3100"), params.Encode())
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

// CollectAndDump collects and dumps the cluster data for troubleshooting and log.
// This function should be call within t.Cleanup.
func CollectAndDump(t *testing.T, rest *rest.Config) {
	if os.Getenv("ACTIONS_STEP_DEBUG") != "true" {
		tlog.Logf(t, "Skipping collecting and dumping cluster data, set ACTIONS_STEP_DEBUG=true to enable it")
		return
	}

	dumpedNamespaces := []string{"envoy-gateway-system"}
	if IsGatewayNamespaceMode() {
		dumpedNamespaces = append(dumpedNamespaces, ConformanceInfraNamespace)
	}

	result := tb.CollectResult(context.TODO(), rest, tb.CollectOptions{
		BundlePath:          "",
		CollectedNamespaces: dumpedNamespaces,
	})
	for r, data := range result {
		tlog.Logf(t, "\nfilename: %s", r)
		tlog.Logf(t, "\ndata: \n%s", data)
	}
}

func GetService(c client.Client, nn types.NamespacedName) (*corev1.Service, error) {
	svc := &corev1.Service{}
	if err := c.Get(context.Background(), nn, svc); err != nil {
		return nil, err
	}
	return svc, nil
}

func CreateBackend(c client.Client, nn types.NamespacedName, clusterIP string, port int32) error {
	backend := &egv1a1.Backend{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: nn.Namespace,
			Name:      nn.Name,
		},
		Spec: egv1a1.BackendSpec{
			Endpoints: []egv1a1.BackendEndpoint{
				{
					IP: &egv1a1.IPEndpoint{
						Address: clusterIP,
						Port:    port,
					},
				},
			},
		},
	}
	return c.Create(context.TODO(), backend)
}

func DeleteBackend(c client.Client, nn types.NamespacedName) error {
	backend := &egv1a1.Backend{}
	if err := c.Get(context.Background(), nn, backend); err != nil {
		return err
	}
	return c.Delete(context.Background(), backend)
}

func ContentEncoding(compressorType egv1a1.CompressorType) string {
	var encoding string
	switch compressorType {
	case egv1a1.BrotliCompressorType:
		encoding = "br"
	case egv1a1.GzipCompressorType:
		encoding = "gzip"
	}

	return encoding
}

func ExpectEnvoyProxyDeploymentCount(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, expectedNs string, expectedCount int) {
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, suite.TimeoutConfig.DeleteTimeout, true, func(ctx context.Context) (bool, error) {
		deploys := &appsv1.DeploymentList{}
		err := suite.Client.List(ctx, deploys, &client.ListOptions{
			Namespace: expectedNs,
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"app.kubernetes.io/managed-by":                   "envoy-gateway",
				"app.kubernetes.io/name":                         "envoy",
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}),
		})
		if err != nil {
			return false, err
		}

		return len(deploys.Items) == expectedCount, err
	})
	if err != nil {
		t.Fatalf("Failed to check Deployment count(%d) for the Gateway: %v", expectedCount, err)
	}
}

func ExpectEnvoyProxyHPACount(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, expectedNs string, expectedCount int) {
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, suite.TimeoutConfig.DeleteTimeout, true, func(ctx context.Context) (bool, error) {
		hpa := &autoscalingv2.HorizontalPodAutoscalerList{}
		err := suite.Client.List(ctx, hpa, &client.ListOptions{
			Namespace: expectedNs,
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}),
		})
		if err != nil {
			return false, err
		}

		return len(hpa.Items) == 1, err
	})
	if err != nil {
		t.Fatalf("Failed to check HPA count(%d) for the Gateway: %v", expectedCount, err)
	}
}

func IsGatewayNamespaceMode() bool {
	return DeployProfile == "gateway-namespace-mode"
}

func GetGatewayResourceNamespace() string {
	if IsGatewayNamespaceMode() {
		return "gateway-conformance-infra"
	}
	return "envoy-gateway-system"
}

func ExpectRequestTimeout(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr, path, query string, exceptedStatusCode int) {
	// Use raw http request to avoid chunked
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{Scheme: "http", Host: gwAddr, Path: path, RawQuery: query},
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	httputils.AwaitConvergence(t, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency,
		func(elapsed time.Duration) bool {
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			// return 504 instead of 400 when request timeout.
			// https://github.com/envoyproxy/envoy/blob/56021dbfb10b53c6d08ed6fc811e1ff4c9ac41fd/source/common/http/utility.cc#L1409
			if exceptedStatusCode == resp.StatusCode {
				return true
			} else {
				tlog.Logf(t, "%s%s response status code: %d after %v", gwAddr, path, resp.StatusCode, elapsed)
				return false
			}
		})
}
