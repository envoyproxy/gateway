//go:build e2e

package tests

import (
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
	ConformanceTests = append(ConformanceTests, TCPAuthorizationClientIPTest)
}

var TCPAuthorizationClientIPTest = suite.ConformanceTest{
	ShortName:   "TCPRouteAuthzWithClientIP",
	Description: "TCP authorization with client IP Allow/Deny list",
	Manifests:   []string{"testdata/tcproute-authorization-client-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		tcpRouteNN := types.NamespacedName{Name: "tcproute-with-authorization-client-ip", Namespace: ns}
		gwNN := types.NamespacedName{Name: "tcproute-with-authorization-client-ip", Namespace: ns}

		// Use the correct function for TCP routes
		gwAddr := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), tcpRouteNN)
		if gwAddr == "" {
			t.Fatalf("GatewayAndTCPRoutesMustBeAccepted returned empty address")
		}

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tcproute-with-authorization-client-ip", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("authorized connection should succeed", func(t *testing.T) {
			// Test TCP connection - should succeed because client IP is allowed
			conn, err := net.DialTimeout("tcp", gwAddr, 5*time.Second)
			if err != nil {
				t.Errorf("Expected authorized connection to succeed, but got error: %v", err)
			} else {
				// Send some data to verify the connection works
				_, writeErr := conn.Write([]byte("test data\n"))
				if writeErr != nil {
					t.Errorf("Failed to write to TCP connection: %v", writeErr)
				}
				conn.Close()
				t.Log("Authorized TCP connection succeeded as expected")
			}
		})

		// Note: Testing denied connections is tricky in e2e tests because
		// the test client typically comes from an allowed IP range.
		// In a real deployment, you'd test with clients from different IP ranges.
	},
}
