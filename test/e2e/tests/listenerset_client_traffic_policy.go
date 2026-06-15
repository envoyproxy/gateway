// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

// This is an EG-specific e2e test, not an upstream conformance test.

package tests

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	gatewayapi "github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetClientTrafficPolicyTest)
}

// ListenerSetClientTrafficPolicyTest verifies that a ClientTrafficPolicy targeting a ListenerSet
// enforces TLS settings on the ListenerSet's listeners. This is an EG-specific feature
// and is not part of the upstream Gateway API conformance suite.
var ListenerSetClientTrafficPolicyTest = suite.ConformanceTest{
	ShortName:   "ListenerSetClientTrafficPolicy",
	Description: "ClientTrafficPolicy targeting a ListenerSet enforces TLS minVersion on ListenerSet HTTPS listeners",
	Manifests: []string{
		"testdata/listenerset-base.yaml",
		"testdata/listenerset-client-traffic-policy.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "listener-gateway", Namespace: ns}
		policyNN := types.NamespacedName{Name: "listenerset-ctp", Namespace: ns}
		lsNN := types.NamespacedName{Name: "listener-set-http", Namespace: ns}

		routeNN := types.NamespacedName{Name: "listenerset-ctp-httpsroute", Namespace: ns}

		gwAddrWithPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "core"))
		if err != nil {
			t.Fatalf("failed to get gateway address: %v", err)
		}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindListenerSet),
			Namespace: gatewayapi.NamespacePtr(lsNN.Namespace),
			Name:      gwapiv1.ObjectName(lsNN.Name),
		}
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, policyNN, suite.ControllerName, ancestorRef)

		// Wait for the route to be accepted by the ListenerSet listener before attempting TLS connections.
		routeParents := []gwapiv1.RouteParentStatus{
			createListenerSetParent(suite.ControllerName, lsNN.Name, "extra-https"),
		}
		kubernetes.RouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, routeParents, false, &gwapiv1.HTTPRoute{})

		listenerAddr := getListenerAddr(gwAddrWithPort, "18443")

		certNN := types.NamespacedName{Name: "listener-https-certificate", Namespace: ns}
		serverCertificate, _, _, err := GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(serverCertificate) {
			t.Fatal("failed to add server certificate to pool")
		}

		//nolint: gosec
		baseTLSConfig := &tls.Config{
			ServerName: "www.example.com",
			RootCAs:    certPool,
		}

		t.Run("tls 1.3 succeeds", func(t *testing.T) {
			dialWithTLSVersion(t, listenerAddr, baseTLSConfig, tls.VersionTLS13, false)
		})

		t.Run("tls 1.2 rejected by min version policy", func(t *testing.T) {
			dialWithTLSVersion(t, listenerAddr, baseTLSConfig, tls.VersionTLS12, true)
		})
	},
}
