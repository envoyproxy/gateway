// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/common/expfmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, MetricTest)
}

var MetricTest = suite.ConformanceTest{
	ShortName:   "Proxy Metrics",
	Description: "Make sure metric is working",
	Manifests:   []string{"testdata/metric.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("prometheus", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "metric-prometheus", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/prom",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// let's check the metric
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					if err := ScrapeMetrics(t, suite.Client, types.NamespacedName{
						Namespace: "envoy-gateway-system",
						Name:      "same-namespace-gw-metrics",
					}, 19001, "/stats/prometheus"); err != nil {
						t.Logf("failed to get metric: %v", err)
						return false, nil
					}
					return true, nil
				}); err != nil {
				t.Errorf("failed to scrape metrics: %v", err)
			}
		})

		t.Run("otel", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "metric-prometheus", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/prom",
				},
				Response: httputils.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			// make sure listener is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// let's check the metric
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					if err := ScrapeMetrics(t, suite.Client, types.NamespacedName{
						Namespace: "monitoring",
						Name:      "otel-collecot-prometheus",
					}, 19001, "/metrics"); err != nil {
						t.Logf("failed to get metric: %v", err)
						return false, nil
					}
					return true, nil
				}); err != nil {
				t.Errorf("failed to scrape metrics: %v", err)
			}
		})
	},
}

func ScrapeMetrics(t *testing.T, c client.Client, nn types.NamespacedName, port int32, path string) error {
	svc := corev1.Service{}
	if err := c.Get(context.Background(), nn, &svc); err != nil {
		return err
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

	url := fmt.Sprintf("http://%s:%d%s", host, port, path)
	t.Logf("try to request: %s", url)

	httpClient := http.Client{
		Timeout: 1 * time.Second,
	}
	res, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to scrape metrics: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to scrape metrics: %s", res.Status)
	}

	metrics, err := (&expfmt.TextParser{}).TextToMetricFamilies(res.Body)
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
