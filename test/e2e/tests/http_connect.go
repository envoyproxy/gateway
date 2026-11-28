// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	nethttp "net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPConnectTunnelTest)
}

var HTTPConnectTunnelTest = suite.ConformanceTest{
	ShortName:   "HTTPConnectTunnel",
	Description: "Tunnel traffic through the Gateway with HTTP CONNECT, terminated at the Gateway or proxied to the backend",
	Manifests:   []string{"testdata/http-connect.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		// Each case sends a plain GET through the tunnel. The echo backend can only answer it
		// if the expected hop terminated the CONNECT request:
		// - terminate: the Gateway, so the tunnel reaches the echo backend directly.
		// - proxy: the connect-proxy-backend Envoy, which the Gateway forwards CONNECT to.
		//   Had the Gateway terminated it instead, the GET would reach that Envoy, which only
		//   routes CONNECT requests.
		cases := []struct {
			name      string
			route     string
			authority string
			// backendApp is the app label of a backend deployed by this test, if any.
			backendApp string
		}{
			{name: "terminate", route: "http-connect-terminate", authority: "connect-terminate.example.com:443"},
			{name: "proxy", route: "http-connect-proxy", authority: "connect-proxy.example.com:443", backendApp: "connect-proxy-backend"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				// Until the backend is ready the Gateway rejects a proxied CONNECT with 403, and
				// pulling its image can take longer than the convergence timeout below.
				if tc.backendApp != "" {
					WaitForPodsReady(t, suite.Client, ns, map[string]string{"app": tc.backendApp})
				}

				routeNN := types.NamespacedName{Name: tc.route, Namespace: ns}
				gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
					suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

				ancestorRef := gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
					Name:      gwapiv1.ObjectName(gwNN.Name),
				}
				BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: tc.route, Namespace: ns},
					suite.ControllerName, ancestorRef)

				path := "/" + tc.route
				http.AwaitConvergence(t, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency,
					func(_ time.Duration) bool {
						if err := getThroughConnectTunnel(gwAddr, tc.authority, path); err != nil {
							tlog.Logf(t, "request through CONNECT tunnel to %s failed: %v", tc.authority, err)
							return false
						}
						return true
					})
			})
		}
	},
}

// getThroughConnectTunnel opens a CONNECT tunnel to authority through the Gateway, sends a GET
// for path through it and checks that the echo backend answered that request.
func getThroughConnectTunnel(gwAddr, authority, path string) error {
	conn, err := net.DialTimeout("tcp", gwAddr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	if _, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", authority, authority); err != nil {
		return err
	}
	br := bufio.NewReader(conn)
	// Only the headers are read: after a 2xx, the rest of the stream belongs to the tunnel.
	connectResp, err := nethttp.ReadResponse(br, &nethttp.Request{Method: nethttp.MethodConnect})
	if err != nil {
		return fmt.Errorf("reading CONNECT response: %w", err)
	}
	if connectResp.StatusCode != nethttp.StatusOK {
		return fmt.Errorf("unexpected CONNECT status %d", connectResp.StatusCode)
	}

	if _, err = fmt.Fprintf(conn, "GET %s HTTP/1.1\r\nHost: tunnel.example.com\r\nConnection: close\r\n\r\n", path); err != nil {
		return err
	}
	resp, err := nethttp.ReadResponse(br, &nethttp.Request{Method: nethttp.MethodGet})
	if err != nil {
		return fmt.Errorf("reading response through the tunnel: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusOK {
		return fmt.Errorf("unexpected status %d through the tunnel", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var echoed struct {
		Path string `json:"path"`
	}
	if err = json.Unmarshal(body, &echoed); err != nil {
		return fmt.Errorf("unexpected body %q: %w", body, err)
	}
	if echoed.Path != path {
		return fmt.Errorf("echo backend saw path %q, expected %q", echoed.Path, path)
	}
	return nil
}
