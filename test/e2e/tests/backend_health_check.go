// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/test/e2e/utils/prometheus"
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
			ns := "gateway-conformance-infra"
			passRouteNN := types.NamespacedName{Name: "http-with-health-check-active-http-pass", Namespace: ns}
			failRouteNN := types.NamespacedName{Name: "http-with-health-check-active-http-fail", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), passRouteNN, failRouteNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "health-check-active-http-pass-btp", Namespace: ns}, suite.ControllerName, ancestorRef)
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "health-check-active-http-fail-btp", Namespace: ns}, suite.ControllerName, ancestorRef)

			promAddr, err := prometheus.Address(suite.Client,
				types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
			)
			require.NoError(t, err)

			passClusterName := fmt.Sprintf("httproute/%s/%s/rule/0", ns, passRouteNN.Name)
			failClusterName := fmt.Sprintf("httproute/%s/%s/rule/0", ns, failRouteNN.Name)
			gtwName := "same-namespace"

			// health check requests will be distributed to the cluster with configured path.
			// we can use membership_healthy stats to check whether health check works as expected.
			passPromQL := fmt.Sprintf(`envoy_cluster_membership_healthy{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, passClusterName, gtwName)
			failPromQL := fmt.Sprintf(`envoy_cluster_membership_healthy{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, failClusterName, gtwName)

			http.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					// check membership_healthy stats from Prometheus
					v, err := prometheus.QuerySum(promAddr, passPromQL)
					if err != nil {
						// wait until Prometheus sync stats
						return false
					}
					t.Logf("cluster pass health check: membership_healthy stats query count: %v", v)

					if v == 0 {
						t.Error("healthy membership is not the same as expected")
					} else {
						t.Log("healthy membership is the same as expected")
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
					v, err := prometheus.QuerySum(promAddr, failPromQL)
					if err != nil {
						// wait until Prometheus sync stats
						return false
					}
					t.Logf("cluster fail health check: membership_healthy stats query count: %v", v)

					if v == 0 {
						t.Log("healthy membership is same as expected")
					} else {
						t.Error("healthy membership is not same as expected")
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
						StatusCode: 200,
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
						StatusCode: 503,
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})
		})
	},
}
