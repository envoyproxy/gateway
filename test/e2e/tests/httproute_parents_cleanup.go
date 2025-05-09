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

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteParentsCleanupTest)
}

// HTTPRouteParentsCleanupTest tests that when a parentRef is removed from HTTPRoute spec.parentRefs,
// the corresponding entry in status.parents is also removed.
var HTTPRouteParentsCleanupTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteParentsCleanup",
	Description: "Test HTTPRoute status.parents cleanup when parentRefs are removed",
	Manifests:   []string{"testdata/httproute-parents-cleanup.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("HTTPRoute status.parents should be cleaned up when parentRefs are removed", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "httproute-parents-cleanup", Namespace: ns}
			gw1NN := types.NamespacedName{Name: "gateway-1", Namespace: ns}
			gw2NN := types.NamespacedName{Name: "gateway-2", Namespace: ns}

			// Wait for the HTTPRoute to be accepted by both gateways
			kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gw1NN), routeNN)
			kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gw2NN), routeNN)

			// Verify that the HTTPRoute has two entries in status.parents
			route := &gwapiv1.HTTPRoute{}
			err := suite.Client.Get(context.Background(), routeNN, route)
			if err != nil {
				t.Fatalf("Failed to get HTTPRoute: %v", err)
			}

			if len(route.Status.Parents) != 2 {
				t.Fatalf("Expected HTTPRoute to have 2 parents in status, got %d", len(route.Status.Parents))
			}

			// Update the HTTPRoute to remove the second parentRef
			route.Spec.ParentRefs = route.Spec.ParentRefs[:1] // Keep only the first parentRef
			err = suite.Client.Update(context.Background(), route)
			if err != nil {
				t.Fatalf("Failed to update HTTPRoute: %v", err)
			}

			// Wait for the HTTPRoute status to be updated
			err = wait.PollUntilContextTimeout(context.Background(), 1*time.Second, suite.TimeoutConfig.RouteMustHaveParents, true, func(ctx context.Context) (bool, error) {
				updatedRoute := &gwapiv1.HTTPRoute{}
				err := suite.Client.Get(ctx, routeNN, updatedRoute)
				if err != nil {
					return false, err
				}

				// Check if status.parents has been cleaned up
				return len(updatedRoute.Status.Parents) == 1, nil
			})
			if err != nil {
				t.Fatalf("Timed out waiting for HTTPRoute status.parents to be cleaned up: %v", err)
			}

			// Get the updated HTTPRoute
			updatedRoute := &gwapiv1.HTTPRoute{}
			err = suite.Client.Get(context.Background(), routeNN, updatedRoute)
			if err != nil {
				t.Fatalf("Failed to get updated HTTPRoute: %v", err)
			}

			// Verify that the remaining parent is for gateway-1
			if len(updatedRoute.Status.Parents) != 1 {
				t.Fatalf("Expected HTTPRoute to have 1 parent in status, got %d", len(updatedRoute.Status.Parents))
			}

			parent := updatedRoute.Status.Parents[0]
			if string(parent.ParentRef.Name) != "gateway-1" {
				t.Fatalf("Expected remaining parent to be gateway-1, got %s", parent.ParentRef.Name)
			}
		})
	},
}
