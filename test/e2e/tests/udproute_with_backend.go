// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from upstream gateway-api, it will be moved to upstream.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/miekg/dns"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, UDPRouteBackendFQDNTest, UDPRouteBackendIPTest)
}

var UDPRouteBackendFQDNTest = suite.ConformanceTest{
	ShortName:   "UDPRouteBackendFQDNTest",
	Description: "UDPRoutes with a backend ref to a FQDN Backend",
	Manifests: []string{
		"testdata/udproute-to-backend-fqdn.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("UDPRoute with a FQDN type Backend", func(t *testing.T) {
			testUDPRouteWithBackend(t, suite, "backend-fqdn")
		})
	},
}

var UDPRouteBackendIPTest = suite.ConformanceTest{
	ShortName:   "UDPRouteBackendIP",
	Description: "UDPRoutes with a backend ref to an IP Backend",
	Manifests: []string{
		"testdata/udproute-to-backend-ip.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("UDPRoute with a IP type Backend", func(t *testing.T) {
			svcNN := types.NamespacedName{
				Name:      "coredns",
				Namespace: "gateway-conformance-infra",
			}
			svc, err := GetService(suite.Client, svcNN)
			if err != nil {
				t.Fatalf("failed to get service %s: %v", svcNN, err)
			}

			backendIPName := "backend-ip"
			ns := "gateway-conformance-infra"
			err = CreateBackend(suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns}, svc.Spec.ClusterIP, 53)
			if err != nil {
				t.Fatalf("failed to create backend %s: %v", backendIPName, err)
			}
			t.Cleanup(func() {
				if err := DeleteBackend(suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns}); err != nil {
					t.Fatalf("failed to delete backend %s: %v", backendIPName, err)
				}
			})
			testUDPRouteWithBackend(t, suite, backendIPName)
		})
	},
}

func testUDPRouteWithBackend(t *testing.T, suite *suite.ConformanceTestSuite, backend string) {
	namespace := "gateway-conformance-infra"
	domain := "foo.bar.com."
	routeNN := types.NamespacedName{Name: "udp-coredns", Namespace: namespace}
	gwNN := types.NamespacedName{Name: "udp-gateway", Namespace: namespace}
	gwAddr := GatewayAndUDPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), routeNN)

	BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: backend, Namespace: namespace})

	msg := new(dns.Msg)
	msg.SetQuestion(domain, dns.TypeA)

	if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
		func(_ context.Context) (done bool, err error) {
			tlog.Logf(t, "performing DNS query %s on %s", domain, gwAddr)
			r, err := dns.Exchange(msg, gwAddr)
			if err != nil {
				tlog.Logf(t, "failed to perform a UDP query: %v", err)
				return false, nil
			}
			tlog.Logf(t, "got DNS response: %s", r.String())
			return true, nil
		}); err != nil {
		t.Errorf("failed to perform DNS query: %v", err)
	}
}
