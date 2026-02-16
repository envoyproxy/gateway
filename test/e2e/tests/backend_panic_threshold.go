// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendPanicThresholdHTTPTest)
}

var BackendPanicThresholdHTTPTest = suite.ConformanceTest{
	ShortName:   "BackendPanicThresholdHTTPTest",
	Description: "Resource with BackendPanicThreshold enabled",
	Manifests:   []string{"testdata/backend-panic-threshold.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("active http", func(t *testing.T) {
			ctx := context.Background()
			ns := "gateway-conformance-infra"
			passRouteNN := types.NamespacedName{Name: "http-with-panic-threshold-pass", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), passRouteNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "panic-threshold-pass-btp", Namespace: ns}, suite.ControllerName, ancestorRef)

			promClient, err := prometheus.NewClient(suite.Client,
				types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
			)
			require.NoError(t, err)

			passClusterName := fmt.Sprintf("httproute/%s/%s/rule/0", ns, passRouteNN.Name)
			gtwName := "same-namespace"

			// health check requests will be distributed to the cluster with configured path.
			// we can use envoy_cluster_health_check_failure stats to ensure HC requests have failed.
			hcFailPromQL := fmt.Sprintf(`envoy_cluster_health_check_failure{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, passClusterName, gtwName)

			http.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					// check hc failure stats from Prometheus
					v, err := promClient.QuerySum(ctx, hcFailPromQL)
					if err != nil {
						// wait until Prometheus sync stats
						return false
					}
					tlog.Logf(t, "cluster fail health check: failure stats query count: %v", v)

					if v == 0 {
						t.Error("failure is not same as expected")
					} else {
						t.Log("failure is same as expected")
					}

					return true
				},
			)

			t.Run("probes succeed with failed HC due to panic mode", func(t *testing.T) {
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path: "/ping",
					},
					Response: http.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})

			panicModePromQL := fmt.Sprintf(`envoy_cluster_lb_healthy_panic{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, passClusterName, gtwName)

			http.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					// check panic mode stats from Prometheus
					v, err := promClient.QuerySum(ctx, panicModePromQL)
					if err != nil {
						// wait until Prometheus sync stats
						return false
					}
					tlog.Logf(t, "cluster lb in panic mode: stats query count: %v", v)

					if v == 0 {
						t.Error("failure is not same as expected")
					} else {
						t.Log("failure is same as expected")
					}

					return true
				},
			)
		})
	},
}
