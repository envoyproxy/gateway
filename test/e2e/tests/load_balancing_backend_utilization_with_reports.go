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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

type backendUtilizationPodRole string

const (
	backendUtilizationWithReportsAppLabel                                           = "backend-utilization-with-reports"
	backendUtilizationWithReportsRoleLabel                                          = "utilization"
	backendUtilizationWithReportsLowRole                  backendUtilizationPodRole = "low"
	backendUtilizationWithReportsHighRole                 backendUtilizationPodRole = "high"
	backendUtilizationWithReportsBlackoutPeriod                                     = 1 * time.Second
	backendUtilizationWithReportsWeightUpdatePeriod                                 = 100 * time.Millisecond
	backendUtilizationWithReportsMinimumSeedAge                                     = backendUtilizationWithReportsBlackoutPeriod + 2*backendUtilizationWithReportsWeightUpdatePeriod
	backendUtilizationWithReportsMaximumSeedDuration                                = 5 * time.Second
	backendUtilizationWithReportsSeedRequestSpacing                                 = 50 * time.Millisecond
	backendUtilizationWithReportsMeasurementRequests                                = 60
	backendUtilizationWithReportsLowShareThresholdPercent                           = 70
)

var BackendUtilizationLoadBalancingWithReports = suite.ConformanceTest{
	ShortName:   "BackendUtilizationLoadBalancingWithReports",
	Description: "Test that BackendUtilization skews traffic toward lower-utilization backends when backends report load",
	Manifests: []string{
		"testdata/load_balancing_backend_utilization_with_reports.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "backend-utilization-with-reports-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-utilization-with-reports-policy", Namespace: ns}, suite.ControllerName, ancestorRef)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": backendUtilizationWithReportsAppLabel}, corev1.PodRunning, &PodReady)

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

		t.Run("traffic should skew toward lower-utilization backends", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/backend-utilization-with-reports",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			if err := wait.PollUntilContextTimeout(t.Context(), 250*time.Millisecond, 20*time.Second, true, func(_ context.Context) (bool, error) {
				podRoles, err := discoverBackendUtilizationPods(t.Context(), suite.Client, ns)
				if err != nil {
					tlog.Logf(t, "failed to discover backend utilization pods: %v", err)
					return false, nil
				}

				seedTrafficMap, err := seedBackendUtilizationTraffic(t, suite, &req, &expectedResponse, podRoles)
				if err != nil {
					tlog.Logf(t, "failed to seed backend utilization traffic: %v", err)
					return false, nil
				}
				tlog.Logf(t, "seed backend utilization traffic map: %v", seedTrafficMap)

				return runTrafficTest(t, suite, &req, &expectedResponse, backendUtilizationWithReportsMeasurementRequests,
					func(trafficMap map[string]int) bool {
						lowTraffic, highTraffic, unknownTraffic := backendUtilizationTrafficByRole(trafficMap, podRoles)
						if unknownTraffic > 0 {
							tlog.Logf(t, "observed traffic from unknown backend-utilization pods: traffic=%v podRoles=%v",
								trafficMap, podRoles)
							return false
						}

						total := lowTraffic + highTraffic
						if total == 0 {
							return false
						}

						lowPct := lowTraffic * 100 / total
						highPct := highTraffic * 100 / total
						tlog.Logf(t, "backend utilization traffic distribution: low=%d(%d%%), high=%d(%d%%)",
							lowTraffic, lowPct, highTraffic, highPct)

						return lowPct >= backendUtilizationWithReportsLowShareThresholdPercent
					}), nil
			}); err != nil {
				consistentHashDump(t, suite.RestConfig)
				t.Fatalf("failed to run backend utilization load balancing with reports test: %v", err)
			}
		})
	},
}

func discoverBackendUtilizationPods(ctx context.Context, cl client.Client, namespace string) (map[string]backendUtilizationPodRole, error) {
	podList := &corev1.PodList{}
	if err := cl.List(ctx, podList, client.InNamespace(namespace), client.MatchingLabels{"app": backendUtilizationWithReportsAppLabel}); err != nil {
		return nil, fmt.Errorf("list backend utilization pods: %w", err)
	}

	podRoles := make(map[string]backendUtilizationPodRole, len(podList.Items))
	var hasLow, hasHigh bool
	for i := range podList.Items {
		pod := &podList.Items[i]
		role := backendUtilizationPodRole(pod.Labels[backendUtilizationWithReportsRoleLabel])
		switch role {
		case backendUtilizationWithReportsLowRole:
			hasLow = true
		case backendUtilizationWithReportsHighRole:
			hasHigh = true
		default:
			return nil, fmt.Errorf("pod %s has unexpected %s label %q", pod.Name, backendUtilizationWithReportsRoleLabel, pod.Labels[backendUtilizationWithReportsRoleLabel])
		}

		podRoles[pod.Name] = role
	}

	if !hasLow || !hasHigh {
		return nil, fmt.Errorf("expected at least one low and one high utilization pod, got %v", podRoles)
	}

	return podRoles, nil
}

func seedBackendUtilizationTraffic(t *testing.T, suite *suite.ConformanceTestSuite,
	req *roundtripper.Request, expectedResponse *http.ExpectedResponse, podRoles map[string]backendUtilizationPodRole,
) (map[string]int, error) {
	t.Helper()

	trafficMap := make(map[string]int)
	firstSeen := make(map[string]time.Time, len(podRoles))
	deadline := time.Now().Add(backendUtilizationWithReportsMaximumSeedDuration)

	for {
		sample := captureTrafficMap(t, suite, req, expectedResponse, 1)

		now := time.Now()
		for pod, count := range sample {
			trafficMap[pod] += count
			if _, ok := podRoles[pod]; ok {
				if _, seen := firstSeen[pod]; !seen {
					firstSeen[pod] = now
				}
			}
		}

		if backendUtilizationPodsReady(podRoles, firstSeen, now) {
			return trafficMap, nil
		}
		if now.After(deadline) {
			return nil, fmt.Errorf("timed out after %s waiting for backend utilization seed convergence: firstSeen=%v traffic=%v",
				backendUtilizationWithReportsMaximumSeedDuration, firstSeen, trafficMap)
		}

		time.Sleep(backendUtilizationWithReportsSeedRequestSpacing)
	}
}

func backendUtilizationPodsReady(podRoles map[string]backendUtilizationPodRole, firstSeen map[string]time.Time, now time.Time) bool {
	var hasLow, hasHigh bool
	for pod, role := range podRoles {
		seenAt, ok := firstSeen[pod]
		if !ok || now.Sub(seenAt) < backendUtilizationWithReportsMinimumSeedAge {
			return false
		}

		switch role {
		case backendUtilizationWithReportsLowRole:
			hasLow = true
		case backendUtilizationWithReportsHighRole:
			hasHigh = true
		}
	}

	return hasLow && hasHigh
}

func backendUtilizationTrafficByRole(trafficMap map[string]int, podRoles map[string]backendUtilizationPodRole) (low, high, unknown int) {
	for pod, count := range trafficMap {
		switch podRoles[pod] {
		case backendUtilizationWithReportsLowRole:
			low += count
		case backendUtilizationWithReportsHighRole:
			high += count
		default:
			unknown += count
		}
	}

	return low, high, unknown
}
