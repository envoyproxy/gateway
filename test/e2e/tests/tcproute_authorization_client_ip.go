// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, TCPRouteAuthzWithClientIP)
}

var TCPRouteAuthzWithClientIP = suite.ConformanceTest{
	ShortName:   "TCPRouteAuthzWithClientIP",
	Description: "Authorization with client IP Allow/Deny list for TCP routes",
	Manifests:   []string{"testdata/tcproute-authorization-client-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		tcpRouteNN := types.NamespacedName{Name: "tcp-backend-authorization-ip", Namespace: ns}
		tcpRouteFqdnNN := types.NamespacedName{Name: "tcp-backend-authorization-fqdn", Namespace: ns}
		gwNN := types.NamespacedName{Name: "tcp-authorization-backend", Namespace: ns}
		GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), tcpRouteNN, tcpRouteFqdnNN)

		// Test the blocked route (ip section)
		ipSection := gwapiv1.SectionName("ip")
		ancestorRefIP := gwapiv1a2.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:        gatewayapi.KindPtr(resource.KindGateway),
			Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:        gwapiv1.ObjectName(gwNN.Name),
			SectionName: &ipSection,
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tcp-backend-authorization-ip-security-policy", Namespace: ns}, suite.ControllerName, ancestorRefIP)

		// Test the allowed route (fqdn section)
		fqdnSection := gwapiv1.SectionName("fqdn")
		ancestorRefFqdn := gwapiv1a2.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:        gatewayapi.KindPtr(resource.KindGateway),
			Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:        gwapiv1.ObjectName(gwNN.Name),
			SectionName: &fqdnSection,
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tcp-backend-authorization-fqdn-security-policy", Namespace: ns}, suite.ControllerName, ancestorRefFqdn)

		t.Run("blocked client IP cannot connect", func(t *testing.T) {
			testTCPRouteWithBackendBlocked(t, suite, "tcp-authorization-backend", "tcp-backend-authorization-ip", "backend-fqdn")
		})

		t.Run("allowed client IP can connect", func(t *testing.T) {
			testTCPRouteWithBackend(t, suite, "tcp-authorization-backend", "tcp-backend-authorization-fqdn", "backend-fqdn")
		})
	},
}

func testTCPRouteWithBackendBlocked(t *testing.T, suite *suite.ConformanceTestSuite, gwName, routeName, backendName string) {
	ns := "gateway-conformance-infra"
	routeNN := types.NamespacedName{Name: routeName, Namespace: ns}
	gwNN := types.NamespacedName{Name: gwName, Namespace: ns}
	gwAddr := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), routeNN)
	BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: backendName, Namespace: ns})

	testTCPConnectionBlocked(t, gwAddr)
}

func testTCPConnectionBlocked(t *testing.T, gwAddr string) {
	// Try to establish a raw TCP connection
	conn, err := net.DialTimeout("tcp", gwAddr, 5*time.Second)
	if err != nil {
		t.Logf("Connection blocked as expected: %v", err)
		return
	}
	defer conn.Close()

	// If connection was established, try sending HTTP request
	req := "GET / HTTP/1.1\r\nHost: " + gwAddr + "\r\nUser-Agent: test-client\r\nAccept: */*\r\n\r\n"
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
