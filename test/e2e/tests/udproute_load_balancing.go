// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, UDPRouteSourceIPConsistentHashTest)
}

var UDPRouteSourceIPConsistentHashTest = suite.ConformanceTest{
	ShortName:   "UDPRouteSourceIPConsistentHash",
	Description: "Test for source IP based consistent hash load balancing on a UDPRoute",
	Manifests:   []string{"testdata/udproute-consistent-hash-source-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const (
			// The number of the coredns backends behind the udp-lb-coredns service.
			backendReplicas = 2
			// The number of the DNS queries sent in a single sample.
			sampleQueries = 20
			// The minimum number of the queries of a sample that must be answered
			// for the sample to be conclusive.
			minAnsweredQueries = sampleQueries / 2
		)

		ns := "gateway-conformance-infra"
		domain := "lb.example.com."
		routeNN := types.NamespacedName{Name: "udp-lb-coredns", Namespace: ns}
		gwNN := types.NamespacedName{Name: "udp-lb-gateway", Namespace: ns}

		podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "udp-lb", "version": "v1"}, corev1.PodRunning, &podReady)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "udp-lb", "version": "v2"}, corev1.PodRunning, &podReady)

		ancestorRef := gwapiv1.ParentReference{
			Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:        gatewayapi.KindPtr(resource.KindGateway),
			Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:        gwapiv1.ObjectName(gwNN.Name),
			SectionName: gatewayapi.SectionNamePtr("coredns"),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "udp-lb-policy", Namespace: ns}, suite.ControllerName, ancestorRef)

		gwAddr := GatewayAndUDPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN, "coredns"), routeNN)

		// Envoy itself must know both backends before the traffic is sampled.
		// A hash table holding a single host pins the traffic on it whatever
		// the hash policy is, which would make the test pass vacuously. The
		// endpoints of the service are pushed to Envoy asynchronously, so its
		// view can still lag behind the ready pods checked above.
		WaitForEnvoyClusterHosts(t, suite, gwNN, routeNN.Name, backendReplicas)

		t.Run("all UDP traffic from the same source IP reaches the same backend", func(t *testing.T) {
			// The sample is retried because the data plane may not be serving
			// yet when the Gateway is reported as accepted: the first queries
			// sent to a fresh Gateway are expected to time out.
			if err := wait.PollUntilContextTimeout(t.Context(), time.Second, 2*time.Minute, true,
				func(_ context.Context) (bool, error) {
					// Every query opens a new UDP session, which is where the
					// load balancer picks a host. Without a source IP hash
					// policy on the UDP proxy, the hash key is missing and the
					// sessions are spread randomly across the backends.
					backends, answered := sampleUDPBackends(t, gwAddr, domain, sampleQueries)
					if answered < minAnsweredQueries {
						tlog.Logf(t, "only %d out of %d DNS queries were answered, retrying", answered, sampleQueries)
						return false, nil
					}
					if len(backends) != 1 {
						tlog.Logf(t, "UDP traffic reached more than one backend (A records %v), retrying", sets.List(backends))
						return false, nil
					}

					tlog.Logf(t, "all the %d answered UDP queries reached a single backend (A record %v)", answered, sets.List(backends))
					return true, nil
				}); err != nil {
				t.Fatalf("failed to route all the UDP traffic to a single backend: %v", err)
			}
		})
	},
}

// sampleUDPBackends sends count DNS queries to addr and returns the set of the
// backends that answered them, along with the number of the answered queries.
// Failed queries are only logged: a single dropped datagram must not fail the
// whole sample.
func sampleUDPBackends(t *testing.T, addr, domain string, count int) (sets.Set[string], int) {
	t.Helper()

	backends := sets.New[string]()
	answered := 0
	for range count {
		backend, err := udpDNSQueryBackend(addr, domain)
		if err != nil {
			tlog.Logf(t, "failed to perform a UDP DNS query: %v", err)
			continue
		}
		answered++
		backends.Insert(backend)
	}

	return backends, answered
}

// udpDNSQueryBackend sends a DNS query for domain to addr and returns the A
// record of the response, which identifies the backend that served the query.
// Each call uses a new source port, and therefore a new UDP proxy session.
func udpDNSQueryBackend(addr, domain string) (string, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(domain, dns.TypeA)

	res, err := dns.Exchange(msg, addr)
	if err != nil {
		return "", err
	}

	for _, answer := range res.Answer {
		if a, ok := answer.(*dns.A); ok {
			return a.A.String(), nil
		}
	}

	return "", fmt.Errorf("no A record in the DNS response: %s", res.String())
}

// WaitForEnvoyClusterHosts waits for the Envoy proxy of the given gateway to
// know at least the given number of hosts for the cluster whose name contains
// clusterName.
func WaitForEnvoyClusterHosts(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, clusterName string, least int) {
	t.Helper()
	tlog.Logf(t, "waiting for the Envoy proxy of %s to know at least %d hosts of the cluster %q...", gwNN, least, clusterName)

	if err := wait.PollUntilContextTimeout(t.Context(), 2*time.Second, defaultServiceStartupTimeout, true,
		func(_ context.Context) (bool, error) {
			hosts, err := envoyClusterHosts(t, suite, gwNN, clusterName)
			if err != nil {
				tlog.Logf(t, "failed to fetch the clusters of the Envoy proxy of %s: %v", gwNN, err)
				return false, nil
			}

			tlog.Logf(t, "the Envoy proxy of %s knows the hosts %v of the cluster %q, want at least %d", gwNN, hosts, clusterName, least)
			return len(hosts) >= least, nil
		}); err != nil {
		t.Fatalf("the Envoy proxy of %s never knew %d hosts of the cluster %q: %v", gwNN, least, clusterName, err)
	}
}

// envoyClusterHosts returns the distinct hosts that the Envoy proxy of the
// given gateway holds for the cluster whose name contains clusterName.
func envoyClusterHosts(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, clusterName string) ([]string, error) {
	t.Helper()

	body, err := fetchEnvoyClustersOutput(t, suite,
		"app.kubernetes.io/name=envoy",
		"gateway.envoyproxy.io/owning-gateway-name="+gwNN.Name,
		"gateway.envoyproxy.io/owning-gateway-namespace="+gwNN.Namespace,
	)
	if err != nil {
		return nil, err
	}

	// /clusters returns one stat per line as "<cluster_name>::<stat_path>::<value>".
	// The stats of a single host are prefixed by its address, e.g.:
	//   udproute/ns/udp-lb-coredns/rule/-1::10.244.0.168:53::health_flags::healthy
	//   udproute/ns/udp-lb-coredns/rule/-1::added_via_api::true
	// Only the former carries a port, which is what tells the two apart.
	hosts := sets.New[string]()
	for _, line := range strings.Split(body, "\n") {
		name, stat, ok := strings.Cut(line, "::")
		if !ok || !strings.Contains(name, clusterName) {
			continue
		}
		host, _, ok := strings.Cut(stat, "::")
		if !ok || !strings.Contains(host, ":") {
			continue
		}
		hosts.Insert(host)
	}

	return sets.List(hosts), nil
}
