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
	ConformanceTests = append(ConformanceTests, BackendHealthCheckActiveHTTPTest)
}

var BackendHealthCheckActiveHTTPTest = suite.ConformanceTest{
	ShortName:   "BackendHealthCheckActiveHTTP",
	Description: "Resource with BackendHealthCheckActiveHTTP enabled",
	Manifests:   []string{"testdata/backend-health-check-active-http.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("active http", func(t *testing.T) {
			ctx := context.Background()
			ns := "gateway-conformance-infra"
			passRouteNN := types.NamespacedName{Name: "http-with-health-check-active-http-pass", Namespace: ns}
			failRouteNN := types.NamespacedName{Name: "http-with-health-check-active-http-fail", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), passRouteNN, failRouteNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "health-check-active-http-pass-btp", Namespace: ns}, suite.ControllerName, ancestorRef)
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "health-check-active-http-fail-btp", Namespace: ns}, suite.ControllerName, ancestorRef)

			promClient, err := prometheus.NewClient(suite.Client,
				types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
			)
			require.NoError(t, err)

			passClusterName := fmt.Sprintf("httproute/%s/%s/rule/0", ns, passRouteNN.Name)
			failClusterName := fmt.Sprintf("httproute/%s/%s/rule/0", ns, failRouteNN.Name)
			gtwName := "same-namespace"

			// health check requests will be distributed to the cluster with configured path.
			// we can use membership_healthy stats to check whether health check works as expected.
			passPromQL := fmt.Sprintf(`envoy_cluster_health_check_success{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, passClusterName, gtwName)
			failPromQL := fmt.Sprintf(`envoy_cluster_health_check_failure{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, failClusterName, gtwName)

			http.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					// check membership_healthy stats from Prometheus
					v, err := promClient.QuerySum(ctx, passPromQL)
					if err != nil {
						// wait until Prometheus sync stats
						return false
					}
					tlog.Logf(t, "cluster pass health check: success stats query count: %v", v)

					if v == 0 {
						t.Error("success is not the same as expected")
					} else {
						t.Log("success is the same as expected")
					}

					return true
				},
			)

			http.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					// check membership_healthy stats from Prometheus
					v, err := promClient.QuerySum(ctx, failPromQL)
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

			t.Run("health check pass", func(t *testing.T) {
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path: "/health-check-active-http-pass",
					},
					Response: http.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})

			t.Run("health check fail", func(t *testing.T) {
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path: "/health-check-active-http-fail",
					},
					Response: http.Response{
						StatusCodes: []int{503},
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})
		})
	},
}
