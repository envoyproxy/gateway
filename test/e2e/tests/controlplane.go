// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, ControlPlaneMetricTest)
}

var ControlPlaneMetricTest = suite.ConformanceTest{
	ShortName:   "ControlPlane",
	Description: "Make sure control plane prometheus endpoint is working",
	Manifests:   []string{"testdata/prometheus.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Prometheus", func(t *testing.T) {
			nn := types.NamespacedName{Name: "envoy-gateway-metrics-lb", Namespace: "envoy-gateway-system"}
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					svc := corev1.Service{}
					if err := suite.Client.Get(context.Background(), nn, &svc); err != nil {
						return false, nil
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
						// do nothing
					}

					if host == "" {
						return false, nil
					}

					return true, nil
				}); err != nil {
				t.Errorf("failed to get service %s : %v", nn.String(), err)
			}

			// too much flakes in the test if timeout is 1 minute
			// this should not take so long, but we give it a long timeout to be safe, and poll every second
			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, 2*time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					if err := ScrapeMetrics(t, suite.Client, nn, 19001, "/metrics"); err != nil {
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
