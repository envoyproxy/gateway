// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/test/utils/prometheus"
)

func init() {
	ConformanceTests = append(ConformanceTests, TLSRouteTerminateAuthzWithClientIP)
}

var TLSRouteTerminateAuthzWithClientIP = suite.ConformanceTest{
	ShortName:   "TLSRouteTerminateAuthzWithClientIP",
	Description: "Authorization with client IP Allow/Deny list for TLS routes (terminate)",
	Manifests:   []string{"testdata/tlsroute-terminate-authorization-client-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		tlsRouteNNBlocked := types.NamespacedName{Name: "tls-terminate-authorization-blocked", Namespace: ns}
		tlsRouteNNAllowed := types.NamespacedName{Name: "tls-terminate-authorization-allowed", Namespace: ns}
		gwNN := types.NamespacedName{Name: "tls-terminate-authorization-backend", Namespace: ns}
		gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "tls-terminate"), tlsRouteNNBlocked, tlsRouteNNAllowed)

		// Create Prometheus client
		promClient, err := prometheus.NewClient(suite.Client,
			types.NamespacedName{Name: "prometheus", Namespace: "monitoring"},
		)
		require.NoError(t, err)

		// SecurityPolicy status.ancestors references the Gateway listener, not the route itself
		tlsTerminateSection := gwapiv1.SectionName("tls-terminate")
		ancestorRef := gwapiv1a2.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:        gatewayapi.KindPtr(resource.KindGateway),
			Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:        gwapiv1.ObjectName(gwNN.Name),
			SectionName: &tlsTerminateSection,
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tls-terminate-authorization-blocked-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tls-terminate-authorization-allowed-policy", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("blocked client IP cannot connect to blocked-terminate.example.com", func(t *testing.T) {
			testTLSRouteConnectionBlocked(t, gwAddr, "blocked-terminate.example.com")
			verifyRBACStats(t, promClient, false, "tls-terminate-8443")
		})

		t.Run("allowed client IP can connect to allowed-terminate.example.com", func(t *testing.T) {
			testTLSRouteConnectionAllowed(t, gwAddr, "allowed-terminate.example.com")
			verifyRBACStats(t, promClient, true, "tls-terminate-8443")
		})
	},
}
