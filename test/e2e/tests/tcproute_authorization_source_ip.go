// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, TCPRouteAuthzWithSourceIP)
}

var TCPRouteAuthzWithSourceIP = suite.ConformanceTest{
	ShortName:   "TCPRouteAuthzWithSourceIP",
	Description: "Authorization with direct source IP Allow/Deny list for TCP routes",
	Manifests:   []string{"testdata/tcproute-authorization-source-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		blockedRouteNN := types.NamespacedName{Name: "tcp-backend-authorization-sourceip-blocked", Namespace: ns}
		allowedRouteNN := types.NamespacedName{Name: "tcp-backend-authorization-sourceip-allowed", Namespace: ns}
		gwNN := types.NamespacedName{Name: "tcp-sourceip-authorization-backend", Namespace: ns}
		GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), blockedRouteNN, allowedRouteNN)

		blockedSection := gwapiv1.SectionName("blocked")
		ancestorRefBlocked := gwapiv1a2.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:        gatewayapi.KindPtr(resource.KindGateway),
			Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:        gwapiv1.ObjectName(gwNN.Name),
			SectionName: &blockedSection,
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tcp-backend-authorization-sourceip-blocked-security-policy", Namespace: ns}, suite.ControllerName, ancestorRefBlocked)

		allowedSection := gwapiv1.SectionName("allowed")
		ancestorRefAllowed := gwapiv1a2.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:        gatewayapi.KindPtr(resource.KindGateway),
			Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:        gwapiv1.ObjectName(gwNN.Name),
			SectionName: &allowedSection,
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "tcp-backend-authorization-sourceip-allowed-security-policy", Namespace: ns}, suite.ControllerName, ancestorRefAllowed)

		t.Run("blocked source IP cannot connect", func(t *testing.T) {
			testTCPRouteWithBackendBlocked(t, suite, "tcp-sourceip-authorization-backend", "tcp-backend-authorization-sourceip-blocked", "backend-fqdn")
		})

		t.Run("allowed source IP can connect", func(t *testing.T) {
			testTCPRouteWithBackend(t, suite, "tcp-sourceip-authorization-backend", "tcp-backend-authorization-sourceip-allowed", "backend-fqdn")
		})
	},
}
