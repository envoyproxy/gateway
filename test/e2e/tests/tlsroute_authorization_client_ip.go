// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, TLSRouteAuthzWithClientIP)
}

var TLSRouteAuthzWithClientIP = suite.ConformanceTest{
	ShortName:   "TLSRouteAuthzWithClientIP",
	Description: "Authorization with client IP Allow/Deny list for TLS routes (passthrough)",
	Manifests:   []string{"testdata/tlsroute-authorization-client-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		tlsRouteNNBlocked := types.NamespacedName{Name: "tls-backend-authorization-blocked", Namespace: ns}
		tlsRouteNNAllowed := types.NamespacedName{Name: "tls-backend-authorization-allowed", Namespace: ns}
		gwNN := types.NamespacedName{Name: "tls-authorization-backend", Namespace: ns}
		gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls-passthrough"), tlsRouteNNBlocked, tlsRouteNNAllowed)

		// Create Prometheus client
		promClient, err := prometheus.NewClient(suite.Client,
			types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
		)
		require.NoError(t, err)

		// SecurityPolicy status.ancestors references the Gateway listener, not the route itself
		// This matches how TCPRoute authorization tests work
		tlsPassthroughSection := gwapiv1.SectionName("tls-passthrough")
		ancestorRef := gwapiv1a2.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:        gatewayapi.KindPtr(resource.KindGateway),
			Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:        gwapiv1.ObjectName(gwNN.Name),
			SectionName: &tlsPassthroughSection,
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tls-backend-authorization-blocked-security-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tls-backend-authorization-allowed-security-policy", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("blocked client IP cannot connect to blocked.example.com", func(t *testing.T) {
			testTLSRouteConnectionBlocked(t, gwAddr, "blocked.example.com")
			verifyRBACStats(t, promClient, false)
		})

		t.Run("allowed client IP can connect to allowed.example.com", func(t *testing.T) {
			testTLSRouteConnectionAllowed(t, gwAddr, "allowed.example.com")
			verifyRBACStats(t, promClient, true)
		})
	},
}

func testTLSRouteConnectionBlocked(t *testing.T, gwAddr, hostname string) {
	// Try to establish a TLS connection with SNI
	tlsConfig := &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: true, //nolint:gosec
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", gwAddr, tlsConfig)
	if err != nil {
		t.Logf("Connection blocked as expected: %v", err)
		return
	}
	defer conn.Close()

	// If connection was established, try sending data
	req := "GET / HTTP/1.1\r\nHost: " + hostname + "\r\nUser-Agent: test-client\r\nAccept: */*\r\n\r\n"
	_, err = conn.Write([]byte(req))
	if err != nil {
		t.Logf("Connection blocked during write as expected: %v", err)
		return
	}

	// Try to read response with a short timeout
	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Logf("Failed to set read deadline: %v", err)
		return
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)

	if errors.Is(err, io.EOF) || n == 0 {
		t.Log("Got empty reply from server as expected (connection blocked)")
		return
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		t.Log("Connection timed out as expected (connection blocked)")
		return
	}
	if err != nil {
		t.Logf("Connection blocked with error as expected: %v", err)
		return
	}

	// If we got here, we received some data, which means the connection was NOT blocked
	response := string(buf[:n])
	t.Fatalf("Expected connection to be blocked, but got response: %s", response)
}

func testTLSRouteConnectionAllowed(t *testing.T, gwAddr, hostname string) {
	// Establish a TLS connection with SNI
	tlsConfig := &tls.Config{
		ServerName:         hostname,
		InsecureSkipVerify: true, //nolint:gosec
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", gwAddr, tlsConfig)
	if err != nil {
		t.Fatalf("Failed to establish TLS connection (should be allowed): %v", err)
	}
	defer conn.Close()

	// Send a simple HTTP request
	req := "GET / HTTP/1.1\r\nHost: " + hostname + "\r\nUser-Agent: test-client\r\nAccept: */*\r\n\r\n"
	_, err = conn.Write([]byte(req))
	if err != nil {
		t.Fatalf("Failed to write data (should be allowed): %v", err)
	}

	// Try to read response
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("Failed to set read deadline: %v", err)
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)

	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Failed to read response (should be allowed): %v", err)
	}

	if n > 0 {
		t.Logf("Successfully received response: %s", string(buf[:n]))
	} else {
		t.Log("Connection allowed and completed successfully")
	}
}

func verifyRBACStats(t *testing.T, promClient *prometheus.Client, expectAllowed bool) {
	t.Helper()

	metric := "envoy_rbac_denied"
	if expectAllowed {
		metric = "envoy_rbac_allowed"
	}

	// Query RBAC metrics for the specific TLS listener (tls-passthrough-8443)
	// This ensures we're checking stats for THIS test's traffic, not other tests
	query := metric + `{namespace="envoy-gateway-system",envoy_rbac_prefix="tls-passthrough-8443"}`

	// Poll for RBAC stats with a short timeout
	err := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 10*time.Second, true,
		func(ctx context.Context) (bool, error) {
			val, err := promClient.QuerySum(ctx, query)
			if err != nil {
				t.Logf("Prometheus query failed (will retry): %v", err)
				return false, nil
			}

			if val > 0 {
				if expectAllowed {
					t.Logf("RBAC filter allowed %v connections (confirmed via Prometheus)", val)
				} else {
					t.Logf("RBAC filter denied %v connections (confirmed via Prometheus)", val)
				}
				return true, nil
			}

			return false, nil
		})

	require.NoError(t, err, "Failed to verify RBAC stats via Prometheus")
}
