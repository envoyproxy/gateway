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
	ConformanceTests = append(ConformanceTests, BackendHealthCheckHTTPVersionTest)
}

// BackendHealthCheckHTTPVersionTest verifies the HTTP version of active health checks:
// derived from the backend protocol by default, negotiated through ALPN on a TLS backend
// that negotiates its own protocol, and fixed by an explicit version.
var BackendHealthCheckHTTPVersionTest = suite.ConformanceTest{
	ShortName:   "BackendHealthCheckHTTPVersion",
	Description: "Active HTTP health checks are sent over the HTTP version of the backend",
	Manifests:   []string{"testdata/backend-health-check-http-version.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gtwName := "same-namespace"
		gwNN := types.NamespacedName{Name: gtwName, Namespace: ns}

		cases := []struct {
			name string
			// healthy is whether the health check is expected to pass.
			healthy bool
		}{
			// The h2c port of the backend only answers HTTP/2, which Auto derives from its appProtocol.
			{name: "h2c-auto", healthy: true},
			{name: "h2c-http1", healthy: false},
			// The TLS backend negotiates HTTP/2 through ALPN, which the health check matches in the body.
			{name: "tls-auto", healthy: true},
			{name: "tls-http1", healthy: false},
		}

		routeNNs := make([]types.NamespacedName, 0, len(cases))
		for _, tc := range cases {
			routeNNs = append(routeNNs, types.NamespacedName{Name: "http-with-health-check-" + tc.name, Namespace: ns})
		}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeNNs...)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		for _, tc := range cases {
			BackendTrafficPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: fmt.Sprintf("health-check-%s-btp", tc.name), Namespace: ns},
				suite.ControllerName, ancestorRef)
		}

		promClient, err := prometheus.NewClient(suite.Client,
			types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
		)
		require.NoError(t, err)

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				ctx := context.Background()
				clusterName := fmt.Sprintf("httproute/%s/http-with-health-check-%s/rule/0", ns, tc.name)

				// A health check that passes over the expected protocol counts as a success, and one
				// sent over the wrong protocol counts as a failure.
				metric := "envoy_cluster_health_check_success"
				if !tc.healthy {
					metric = "envoy_cluster_health_check_failure"
				}
				promQL := fmt.Sprintf(`%s{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`,
					metric, clusterName, gtwName)
				http.AwaitConvergence(
					t,
					suite.TimeoutConfig.RequiredConsecutiveSuccesses,
					suite.TimeoutConfig.MaxTimeToConsistency,
					func(_ time.Duration) bool {
						v, err := promClient.QuerySum(ctx, promQL)
						if err != nil {
							// wait until Prometheus sync stats
							return false
						}
						tlog.Logf(t, "cluster %s: %s query count: %v", clusterName, metric, v)
						return v > 0
					},
				)

				// With panic mode disabled, a cluster whose health checks fail has no host to route to.
				expectedStatus := 200
				if !tc.healthy {
					expectedStatus = 503
				}
				expectedResponse := http.ExpectedResponse{
					Request: http.Request{
						Path: "/health-check-" + tc.name,
					},
					Response: http.Response{
						StatusCodes: []int{expectedStatus},
					},
					Namespace: ns,
				}
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
			})
		}
	},
}
