//go:build e2e

package tests

import (
	"net"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	// ConformanceTests = append(ConformanceTests, TCPAuthorizationClientIPTest)
}

var TCPAuthorizationClientIPTest = suite.ConformanceTest{
	ShortName:   "TCPAuthzWithClientIP",
	Description: "TCP authorization with client IP Allow/Deny list",
	Manifests:   []string{"testdata/tcproute-authorization-client-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		tcpRouteNN := types.NamespacedName{Name: "tcproute-with-authorization-client-ip", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), tcpRouteNN)

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tcp-authorization-client-ip", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("connection test", func(t *testing.T) {
			// Test basic TCP connection
			conn, err := net.DialTimeout("tcp", gwAddr, 5*time.Second)
			if err != nil {
				t.Errorf("Expected connection to succeed, but got error: %v", err)
			} else {
				conn.Close()
			}
		})
	},
}
