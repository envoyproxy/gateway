// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"fmt"
	"net"
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
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, ConnectionLimitTest)
}

var ConnectionLimitTest = suite.ConformanceTest{
	ShortName:   "ConnectionLimit",
	Description: "Deny Requests over connection limit",
	Manifests:   []string{"testdata/connection-limit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ctx := context.Background()

		promClient, err := prometheus.NewClient(suite.Client, types.NamespacedName{Name: "prometheus", Namespace: "monitoring"})
		require.NoError(t, err)

		t.Run("Close connections over limit", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-connection-limit", Namespace: ns}
			gwNN := types.NamespacedName{Name: "connection-limit-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "connection-limit-ctp", Namespace: ns}, suite.ControllerName, ancestorRef)

			// we make the number of connections equal to the number of connectionLimit connections + 3
			// avoid partial connection errors or interruptions
			for i := 0; i < 6; i++ {
				conn, err := net.Dial("tcp", gwAddr)
				if err != nil {
					t.Errorf("failed to open connection: %v", err)
				} else {
					defer conn.Close()
				}
			}

			prefix := "http"
			gtwName := "connection-limit-gateway"
			promQL := fmt.Sprintf(`envoy_connection_limit_limited_connections{envoy_connection_limit_prefix="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`, prefix, gtwName)

			http.AwaitConvergence(
				t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(_ time.Duration) bool {
					// check connection_limit stats from Prometheus
					v, err := promClient.QuerySum(ctx, promQL)
					if err != nil {
						// wait until Prometheus sync stats
						return false
					}
					t.Logf("connection_limit stats query count: %v", v)

					// connection interruptions or other connection errors may occur
					// we just need to determine whether there is a connection limit stats
					if v == 0 {
						t.Error("connection is not limited as expected")
					} else {
						t.Log("connection is limited as expected")
					}

					return true
				},
			)
		})
	},
}
