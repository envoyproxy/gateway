// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
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
	ConformanceTests = append(ConformanceTests, WeightedBackendConsistentHashLoadBalancing)
}

// WeightedBackendConsistentHashLoadBalancing reproduces the single-cluster
// scenario this fix targets: two backendRefs with no per-backendRef filters
// collapse into one cluster expressed as weighted localities, combined with
// a ConsistentHash load balancer. It asserts both that a given hash key
// always pins to the same backend, and that across many distinct hash keys
// the configured 5/95 weight split is honored rather than being treated as
// equal.
var WeightedBackendConsistentHashLoadBalancing = suite.ConformanceTest{
	ShortName:   "WeightedBackendConsistentHashLoadBalancing",
	Description: "Test that backendRef weights are honored when backendRefs collapse into a single cluster under a ConsistentHash load balancer",
	Manifests: []string{
		"testdata/weighted-backend-consistent-hash.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const (
			sendRequests = 200
			// The configured split is 95:5. Leave a wide buffer since 200
			// distinct hash keys are not going to land exactly on it, and the
			// point of this assertion is to catch the weights being silently
			// ignored (i.e. an even ~50/50 split), not to pin the exact ratio.
			v2MinPct = 80
		)

		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "weighted-consistent-hash-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: SameNamespaceGateway.Name, Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "weighted-consistent-hash-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		expectedResponse := &http.ExpectedResponse{
			Request: http.Request{
				Path: "/weighted-consistent-hash",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		t.Run("same hash key always reaches the same backend", func(t *testing.T) {
			req := http.MakeRequest(t, expectedResponse, gwAddr, "HTTP", "http")
			req.Headers["Lb-Test-Header"] = []string{"pinned-user"}
			runConsistentHashLoadBalancingTest(t, suite, &req, expectedResponse)
		})

		t.Run("distinct hash keys are distributed according to the configured weights", func(t *testing.T) {
			if err := wait.PollUntilContextTimeout(t.Context(), time.Second, 30*time.Second, true, func(_ context.Context) (bool, error) {
				weightMap := make(map[string]int)
				for i := range sendRequests {
					req := http.MakeRequest(t, expectedResponse, gwAddr, "HTTP", "http")
					req.Headers["Lb-Test-Header"] = []string{fmt.Sprintf("user-%d", i)}

					cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
					if err != nil {
						tlog.Logf(t, "failed to get expected response: %v", err)
						continue
					}
					if err := http.CompareRoundTrip(t, &req, cReq, cResp, *expectedResponse); err != nil {
						tlog.Logf(t, "unexpected response: %v", err)
						continue
					}
					if len(cReq.Pod) == 0 {
						continue
					}
					weightMap[extractPodNamePrefix(cReq.Pod, "infra-backend")]++
				}

				total := weightMap["infra-backend-v1"] + weightMap["infra-backend-v2"]
				if total == 0 {
					return false, nil
				}
				v2Pct := (weightMap["infra-backend-v2"] * 100) / total
				tlog.Logf(t, "weighted consistent hash distribution: v1=%d, v2=%d (v2=%d%%)",
					weightMap["infra-backend-v1"], weightMap["infra-backend-v2"], v2Pct)
				return v2Pct >= v2MinPct, nil
			}); err != nil {
				tlog.Errorf(t, "backendRef weights were not honored under the ConsistentHash load balancer: %v", err)
			}
		})
	},
}
