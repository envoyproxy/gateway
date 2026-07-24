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

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteStaleStatusTest)
}

var HTTPRouteStaleStatusTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteStaleStatus",
	Description: "Stale Accepted=False/NoMatchingListenerHostname condition must be cleared after a matching listener is added to the Gateway",
	Manifests:   []string{"testdata/httproute-stale-status.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "stale-status-gw", Namespace: ns}
		routeNN := types.NamespacedName{Name: "stale-status-route", Namespace: ns}

		// Phase 1: Gateway listener hostname "other.example.com" does not match the
		// route's hostname "stale.example.com".
		// Expect Accepted=False / NoMatchingListenerHostname.
		t.Run("route rejected with NoMatchingListenerHostname when hostnames differ", func(t *testing.T) {
			tlog.Logf(t, "Waiting for HTTPRoute %s to be Accepted=False/NoMatchingListenerHostname", routeNN)
			require.NoError(t, wait.PollUntilContextTimeout(
				t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
				func(ctx context.Context) (bool, error) {
					route := &gwapiv1.HTTPRoute{}
					if err := suite.Client.Get(ctx, routeNN, route); err != nil {
						tlog.Logf(t, "failed to get HTTPRoute: %v", err)
						return false, nil
					}
					for _, parent := range route.Status.Parents {
						for _, cond := range parent.Conditions {
							if cond.Type == string(gwapiv1.RouteConditionAccepted) {
								if cond.Status == metav1.ConditionFalse &&
									cond.Reason == string(gwapiv1.RouteReasonNoMatchingListenerHostname) {
									tlog.Logf(t, "HTTPRoute correctly Accepted=False/NoMatchingListenerHostname: %s", cond.Message)
									return true, nil
								}
								tlog.Logf(t, "Accepted condition not yet expected (status=%s reason=%s)", cond.Status, cond.Reason)
							}
						}
					}
					return false, nil
				},
			))
		})

		// Phase 2: Add a second listener whose hostname matches the route's hostname.
		// The original non-matching listener is kept.
		t.Run("matching listener added to gateway", func(t *testing.T) {
			tlog.Logf(t, "Adding matching listener (stale.example.com) to gateway %s", gwNN)
			require.NoError(t, wait.PollUntilContextTimeout(
				t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
				func(ctx context.Context) (bool, error) {
					gw := &gwapiv1.Gateway{}
					if err := suite.Client.Get(ctx, gwNN, gw); err != nil {
						tlog.Logf(t, "failed to get Gateway: %v", err)
						return false, nil
					}
					gw.Spec.Listeners = []gwapiv1.Listener{
						{
							Name:     "http-first",
							Port:     8888,
							Protocol: gwapiv1.HTTPProtocolType,
							Hostname: ptrTo(gwapiv1.Hostname("other.example.com")),
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{From: ptrTo(gwapiv1.NamespacesFromSame)},
							},
						},
						{
							Name:     "http-match",
							Port:     8889,
							Protocol: gwapiv1.HTTPProtocolType,
							Hostname: ptrTo(gwapiv1.Hostname("stale.example.com")),
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{From: ptrTo(gwapiv1.NamespacesFromSame)},
							},
						},
					}
					if err := suite.Client.Update(ctx, gw, &client.UpdateOptions{}); err != nil {
						tlog.Logf(t, "failed to update Gateway: %v", err)
						return false, nil
					}
					tlog.Logf(t, "Gateway updated, new generation: %d", gw.Generation)
					return true, nil
				},
			))
		})

		// Phase 3: The NoMatchingListenerHostname condition must now be gone.
		// The route must be Accepted=True / Accepted.
		t.Run("NoMatchingListenerHostname cleared and route accepted after matching listener added", func(t *testing.T) {
			tlog.Logf(t, "Waiting for HTTPRoute %s to be Accepted=True/Accepted", routeNN)
			require.NoError(t, wait.PollUntilContextTimeout(
				t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
				func(ctx context.Context) (bool, error) {
					route := &gwapiv1.HTTPRoute{}
					if err := suite.Client.Get(ctx, routeNN, route); err != nil {
						tlog.Logf(t, "failed to get HTTPRoute: %v", err)
						return false, nil
					}
					for _, parent := range route.Status.Parents {
						for _, cond := range parent.Conditions {
							if cond.Type == string(gwapiv1.RouteConditionAccepted) {
								if cond.Status == metav1.ConditionTrue &&
									cond.Reason == string(gwapiv1.RouteReasonAccepted) {
									tlog.Logf(t, "NoMatchingListenerHostname cleared, route is Accepted=True/Accepted: %s", cond.Message)
									return true, nil
								}
								tlog.Logf(t, "Condition not yet updated (status=%s reason=%s msg=%s)", cond.Status, cond.Reason, cond.Message)
							}
						}
					}
					return false, nil
				},
			))
		})
	},
}

func ptrTo[T any](v T) *T {
	return &v
}
