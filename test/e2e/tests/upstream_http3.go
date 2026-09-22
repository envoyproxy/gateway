// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"io"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, UpstreamHTTP3Test)
}

var UpstreamHTTP3Test = suite.ConformanceTest{
	ShortName:   "UpstreamHTTP3",
	Description: "BackendTrafficPolicy http3 makes Envoy talk HTTP/3 to the backend",
	Manifests:   []string{"testdata/upstream-http3.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ctx := context.Background()
		ns := "gateway-conformance-infra"
		gtwName := "same-namespace"
		alwaysRouteNN := types.NamespacedName{Name: "upstream-http3", Namespace: ns}
		autoRouteNN := types.NamespacedName{Name: "upstream-http3-auto", Namespace: ns}
		gwNN := types.NamespacedName{Name: gtwName, Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
			suite.ControllerName, kubernetes.NewGatewayRef(gwNN), alwaysRouteNN, autoRouteNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, alwaysRouteNN, suite.ControllerName, ancestorRef)
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, autoRouteNN, suite.ControllerName, ancestorRef)

		promClient, err := prometheus.NewClient(suite.Client,
			types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
		)
		require.NoError(t, err)

		// The backend is Caddy rather than the conformance echo server, because the echo
		// server does not speak HTTP/3. It answers with plain text, so the conformance
		// response comparison (which expects a JSON CapturedRequest) does not apply and the
		// status code is checked directly.
		getStatus := func(path string) (int, error) {
			req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodGet, fmt.Sprintf("http://%s%s", gwAddr, path), nil)
			if err != nil {
				return 0, err
			}
			resp, err := nethttp.DefaultClient.Do(req)
			if err != nil {
				return 0, err
			}
			defer resp.Body.Close()
			_, _ = io.Copy(io.Discard, resp.Body)
			return resp.StatusCode, nil
		}

		// Both routes share one Backend and differ only in the mode their
		// BackendTrafficPolicy sets, so the counter is scoped per cluster to tell the two
		// apart. A 200 on its own would not prove the protocol: a cluster that fell back to
		// TCP answers just the same. This counter is only incremented by Envoy's HTTP/3
		// connection pool.
		testMode := func(t *testing.T, path string, routeNN types.NamespacedName) {
			clusterName := fmt.Sprintf("httproute/%s/%s/rule/0", ns, routeNN.Name)
			promQL := fmt.Sprintf(
				`envoy_cluster_upstream_cx_http3_total{envoy_cluster_name="%s",gateway_envoyproxy_io_owning_gateway_name="%s"}`,
				clusterName, gtwName)

			sawOK := false
			err := wait.PollUntilContextTimeout(ctx, time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
				// Auto only switches to QUIC once a response has advertised alt-svc, so keep
				// sending requests rather than expecting the first one to be HTTP/3.
				code, err := getStatus(path)
				if err != nil {
					tlog.Logf(t, "request to %s failed: %v", path, err)
					return false, nil
				}
				if code != nethttp.StatusOK {
					tlog.Logf(t, "request to %s returned %d, want 200", path, code)
					return false, nil
				}
				sawOK = true

				v, err := promClient.QuerySum(ctx, promQL)
				if err != nil {
					tlog.Logf(t, "failed to query prometheus: %v", err)
					return false, nil
				}
				if v > 0 {
					tlog.Logf(t, "cluster %s has %v upstream HTTP/3 connections", clusterName, v)
					return true, nil
				}
				tlog.Logf(t, "cluster %s has no upstream HTTP/3 connections yet", clusterName)
				return false, nil
			})
			require.True(t, sawOK, "never got a 200 from %s", path)
			require.NoError(t, err)
		}

		// Always goes straight to QUIC, with no TCP fallback that could mask a failure.
		t.Run("mode Always uses HTTP/3 upstream", func(t *testing.T) {
			testMode(t, "/upstream-http3", alwaysRouteNN)
		})

		// Auto has to discover HTTP/3 through alt-svc first, which exercises the alternate
		// protocols cache and the upstream filter that populates it.
		t.Run("mode Auto upgrades to HTTP/3 after alt-svc", func(t *testing.T) {
			testMode(t, "/upstream-http3-auto", autoRouteNN)
		})
	},
}
