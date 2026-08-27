// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, UDPRouteAuthzWithClientIP)
}

// udpAuthzDomain is answered by the coredns backend the conformance base
// manifest installs.
const udpAuthzDomain = "foo.bar.com."

var UDPRouteAuthzWithClientIP = suite.ConformanceTest{
	ShortName:   "UDPRouteAuthzWithClientIP",
	Description: "Authorization with client IP Allow/Deny list for UDP routes",
	Manifests:   []string{"testdata/udproute-authorization-client-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "udp-authorization-gateway", Namespace: ns}
		allowedNN := types.NamespacedName{Name: "udp-authz-allowed", Namespace: ns}
		deniedNN := types.NamespacedName{Name: "udp-authz-denied", Namespace: ns}
		denyAllNN := types.NamespacedName{Name: "udp-authz-deny-all", Namespace: ns}

		GatewayAndUDPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName,
			NewGatewayRef(gwNN), allowedNN, deniedNN, denyAllNN)

		// coredns must be answering before any probe, otherwise the positive
		// control below would fail for a reason unrelated to authorization.
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "udp"}, corev1.PodRunning, &PodReady)

		for _, policy := range []struct {
			name     string
			listener gwapiv1.SectionName
		}{
			{"udp-authz-allowed-policy", "allowed"},
			{"udp-authz-denied-policy", "denied"},
			{"udp-authz-deny-all-policy", "deny-all"},
		} {
			// Route-attached policies report against the Gateway listener too,
			// so every ancestorRef carries a section name.
			SecurityPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: policy.name, Namespace: ns},
				suite.ControllerName, udpAuthzAncestorRef(gwNN, policy.listener))
		}

		// Each listener has to be addressed by its own port. The address returned
		// by GatewayAndUDPRoutesMustBeAccepted is always the first listener's,
		// because it calls the upstream helper without any listener name.
		allowedAddr := udpAuthzListenerAddr(t, suite, gwNN, "allowed")
		deniedAddr := udpAuthzListenerAddr(t, suite, gwNN, "denied")
		denyAllAddr := udpAuthzListenerAddr(t, suite, gwNN, "deny-all")

		// This subtest runs first by design: it is the only evidence that the
		// proxy is up and the generated config was accepted. Without it a
		// dropped datagram is indistinguishable from a listener that never
		// loaded, which would make both deny subtests vacuous.
		t.Run("allowed client IP gets a response", func(t *testing.T) {
			udpQueryMustSucceed(t, allowedAddr)
		})

		t.Run("client IP outside the allowed CIDR is dropped", func(t *testing.T) {
			udpQueryMustBeDropped(t, deniedAddr)
		})

		t.Run("deny by default with no rules drops everything", func(t *testing.T) {
			udpQueryMustBeDropped(t, denyAllAddr)
		})
	},
}

func udpAuthzAncestorRef(gwNN types.NamespacedName, listener gwapiv1.SectionName) gwapiv1.ParentReference {
	return gwapiv1.ParentReference{
		Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:        gatewayapi.KindPtr(resource.KindGateway),
		Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:        gwapiv1.ObjectName(gwNN.Name),
		SectionName: &listener,
	}
}

// udpAuthzListenerAddr resolves the address of one named listener on a Gateway
// that has several.
func udpAuthzListenerAddr(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, listener string) string {
	t.Helper()

	addr, err := WaitForGatewayAddress(t, suite.Client, &suite.TimeoutConfig, gwNN, listener)
	require.NoErrorf(t, err, "timed out waiting for an address for listener %s", listener)
	return addr
}

func udpQueryMustSucceed(t *testing.T, addr string) {
	t.Helper()

	msg := new(dns.Msg)
	msg.SetQuestion(udpAuthzDomain, dns.TypeA)

	if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
		func(_ context.Context) (done bool, err error) {
			tlog.Logf(t, "performing DNS query %s on %s", udpAuthzDomain, addr)
			r, err := dns.Exchange(msg, addr)
			if err != nil {
				tlog.Logf(t, "failed to perform a UDP query: %v", err)
				return false, nil
			}
			tlog.Logf(t, "got DNS response: %s", r.String())
			return true, nil
		}); err != nil {
		// Fatal rather than an error: the deny assertions mean nothing if the
		// data plane never came up.
		t.Fatalf("failed to perform DNS query: %v", err)
	}
}

// udpQueryMustBeDropped requires several consecutive queries to go unanswered.
//
// A denied datagram draws no reply of any kind, so this deliberately is not a
// poll: a poll that retried until it succeeded would report success after a
// full minute of correct denials.
func udpQueryMustBeDropped(t *testing.T, addr string) {
	t.Helper()

	msg := new(dns.Msg)
	msg.SetQuestion(udpAuthzDomain, dns.TypeA)
	client := &dns.Client{Timeout: 2 * time.Second}

	const attempts = 3
	for i := 1; i <= attempts; i++ {
		tlog.Logf(t, "performing DNS query %s on %s, expecting no reply (%d/%d)", udpAuthzDomain, addr, i, attempts)
		r, _, err := client.Exchange(msg, addr)
		if err == nil {
			t.Fatalf("expected the datagram to be dropped, but got a DNS response: %s", r.String())
		}
		tlog.Logf(t, "datagram dropped as expected: %v", err)
	}
}
