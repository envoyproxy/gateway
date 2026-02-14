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
	corev1 "k8s.io/api/core/v1"
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
	ConformanceTests = append(ConformanceTests,
		BackendHealthCheckActiveHTTPTest,
		BackendHealthCheckWithOverrideTest,
	)
}

var BackendHealthCheckActiveHTTPTest = suite.ConformanceTest{
	ShortName:   "BackendHealthCheckActiveHTTP",
	Description: "Resource with BackendHealthCheckActiveHTTP enabled",
	Manifests:   []string{"testdata/backend-health-check-active-http.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		if XDSNameSchemeV2() {
			t.Skip("cluster name format changed")
		}
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

var BackendHealthCheckWithOverrideTest = suite.ConformanceTest{
	ShortName:   "BackendHealthCheckWithOverride",
	Description: "Test backend health check with override configuration",
	Manifests:   []string{"testdata/backend-health-check-with-override.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		if XDSNameSchemeV2() {
			t.Skip("cluster name format changed")
		}
		ns := "gateway-conformance-infra"
		withOverrideRouteNN := types.NamespacedName{Name: "httproute-with-health-check-override", Namespace: ns}
		withoutOverrideRouteNN := types.NamespacedName{Name: "httproute-without-health-check-override", Namespace: ns}
		gtwName := "same-namespace"
		gwNN := types.NamespacedName{Name: gtwName, Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), withOverrideRouteNN, withoutOverrideRouteNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "btp-with-health-check-override", Namespace: ns}, suite.ControllerName, ancestorRef)
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "btp-without-health-check-override", Namespace: ns}, suite.ControllerName, ancestorRef)

		// Wait for the service pods to be ready
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "multi-ports-backend"}, corev1.PodRunning, &PodReady)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "single-port-backend"}, corev1.PodRunning, &PodReady)

		t.Run("health checks work with and without health check overrides", func(t *testing.T) {
			ctx := context.Background()
			promClient, err := prometheus.NewClient(suite.Client,
				types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
			)
			require.NoError(t, err)

			withOverrideClusterName := fmt.Sprintf("httproute/%s/%s/rule/0", ns, withOverrideRouteNN.Name)
			withoutOverrideClusterName := fmt.Sprintf("httproute/%s/%s/rule/0", ns, withoutOverrideRouteNN.Name)

			// both routes should have successful health checks
			withOverridePromQL := fmt.Sprintf(`envoy_cluster_health_check_success{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, withOverrideClusterName, gtwName)
			withoutOverridePromQL := fmt.Sprintf(`envoy_cluster_health_check_success{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, withoutOverrideClusterName, gtwName)

			// verify health check with override
			http.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					v, err := promClient.QuerySum(ctx, withOverridePromQL)
					if err != nil {
						return false
					}
					tlog.Logf(t, "cluster with health check override: success stats query count: %v", v)

					if v == 0 {
						t.Error("health check with override success count is not the same as expected")
					} else {
						t.Log("health check with override success count is the same as expected")
					}

					return true
				},
			)

			// verify health check without override
			http.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					v, err := promClient.QuerySum(ctx, withoutOverridePromQL)
					if err != nil {
						return false
					}
					tlog.Logf(t, "cluster without health check override: success stats query count: %v", v)

					if v == 0 {
						t.Error("health check without override success count is not the same as expected")
					} else {
						t.Log("health check without override success count is the same as expected")
					}

					return true
				},
			)

			t.Run("traffic routing works with health check override", func(t *testing.T) {
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path: "/health-check-with-override",
					},
					Response: http.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})

			t.Run("traffic routing works without health check override", func(t *testing.T) {
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path: "/health-check-without-override",
					},
					Response: http.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				}

				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})
		})
	},
}
