// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"maps"
	"net"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteBackendPriorityTest)
}

var HTTPRouteBackendPriorityTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteBackendPriority",
	Description: "backendPriority HTTPRouteFilter fails over between the backendRefs of a rule by priority",
	Manifests:   []string{"testdata/httproute-with-backend-priority-filter.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		primaryRouteNN := types.NamespacedName{Name: "backend-priority-primary", Namespace: ns}
		multiTierRouteNN := types.NamespacedName{Name: "backend-priority-multi-tier", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, primaryRouteNN, multiTierRouteNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-priority-health-check", Namespace: ns}, suite.ControllerName, ancestorRef)

		// Each route has a healthy backend at a lower priority than the expected one, which would
		// receive traffic if the priorities weren't applied.
		t.Run("healthy priority 0 receives all traffic", func(t *testing.T) {
			runBackendPriorityTest(t, suite, gwNN, gwAddr, primaryRouteNN, "/backend-priority/primary", "infra-backend-v1",
				map[string]envoyHostState{
					"infra-backend-v1:8080": {priority: "0", healthy: true},
					"infra-backend-v2:8080": {priority: "1", healthy: true},
				})
		})
		t.Run("unhealthy priorities 0 and 1 fail over to priority 2", func(t *testing.T) {
			runBackendPriorityTest(t, suite, gwNN, gwAddr, multiTierRouteNN, "/backend-priority/multi-tier", "infra-backend-v3",
				map[string]envoyHostState{
					"infra-backend-v1:9999": {priority: "0", healthy: false},
					"infra-backend-v2:9999": {priority: "1", healthy: false},
					"infra-backend-v3:8080": {priority: "2", healthy: true},
					"infra-backend-v1:8080": {priority: "3", healthy: true},
				})
		})
	},
}

// envoyHostState is the priority and health of a host in an Envoy cluster.
type envoyHostState struct {
	priority string
	healthy  bool
}

// runBackendPriorityTest waits until the route's cluster has the wanted hosts, keyed by
// "<service name>:<port>", then checks that only the expected backend serves the path.
func runBackendPriorityTest(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, gwAddr string, routeNN types.NamespacedName,
	path, backend string, wantHosts map[string]envoyHostState,
) {
	t.Helper()

	clusterName := fmt.Sprintf("httproute/%s/%s/rule/0", routeNN.Namespace, routeNN.Name)
	waitForEnvoyClusterHosts(t, suite, gwNN, clusterName, wantHosts)

	expectedResponse := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
		},
		Response: http.Response{
			StatusCodes: []int{200},
		},
		Backend:   backend,
		Namespace: ConformanceInfraNamespace,
	}
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

	// Unlike weighted routing, priority failover is deterministic, so every request must hit the
	// expected backend.
	req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
	for range sendRequests {
		cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
		if err != nil {
			t.Fatalf("failed to get expected response: %v", err)
		}
		if err := http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse); err != nil {
			t.Fatalf("failed to compare request and response: %v", err)
		}
	}
}

// waitForEnvoyClusterHosts waits until the hosts of the cluster named clusterName in the Envoy proxy
// of the given gateway match want, keyed by "<service name>:<port>".
func waitForEnvoyClusterHosts(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, clusterName string, want map[string]envoyHostState) {
	t.Helper()

	if err := wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
		func(_ context.Context) (bool, error) {
			body, err := fetchEnvoyClustersOutput(t, suite,
				"app.kubernetes.io/name=envoy",
				"gateway.envoyproxy.io/owning-gateway-name="+gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace="+gwNN.Namespace,
			)
			if err != nil {
				tlog.Logf(t, "failed to fetch the clusters of the Envoy proxy of %s: %v", gwNN, err)
				return false, nil
			}

			got := envoyClusterHostStates(body, clusterName)
			tlog.Logf(t, "the cluster %q has hosts %v, want %v", clusterName, got, want)
			return maps.Equal(got, want), nil
		}); err != nil {
		t.Fatalf("the cluster %q never had hosts %v: %v", clusterName, want, err)
	}
}

// envoyClusterHostStates parses the hosts of the cluster named clusterName from the admin /clusters
// output, which has one stat per line, e.g.:
//
//	<cluster_name>::10.96.90.232:8080::hostname::infra-backend-v1.gateway-conformance-infra.svc.cluster.local
//	<cluster_name>::10.96.90.232:8080::health_flags::healthy
//	<cluster_name>::10.96.90.232:8080::priority::0
func envoyClusterHostStates(body, clusterName string) map[string]envoyHostState {
	type host struct {
		name     string
		priority string
		healthy  bool
	}
	hosts := map[string]*host{}
	for _, line := range strings.Split(body, "\n") {
		name, stat, ok := strings.Cut(line, "::")
		if !ok || name != clusterName {
			continue
		}
		addr := clusterHostRegexp.FindStringSubmatch(stat)
		if addr == nil {
			continue
		}
		h, ok := hosts[addr[1]]
		if !ok {
			h = &host{}
			hosts[addr[1]] = h
		}
		key, value, _ := strings.Cut(strings.TrimPrefix(stat, addr[0]), "::")
		switch key {
		case "hostname":
			h.name, _, _ = strings.Cut(value, ".")
		case "priority":
			h.priority = value
		case "health_flags":
			h.healthy = value == "healthy"
		}
	}

	states := make(map[string]envoyHostState, len(hosts))
	for addr, h := range hosts {
		_, port, _ := net.SplitHostPort(addr)
		states[h.name+":"+port] = envoyHostState{priority: h.priority, healthy: h.healthy}
	}
	return states
}
