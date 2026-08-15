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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, NamespaceSelectorRelabelTest)
}

const (
	nsSelectorRelabelLabelKey   = "gateway-e2e-access"
	nsSelectorRelabelLabelValue = "granted"

	nsSelectorRelabelGatewayName = "namespace-selector-relabel"
	nsSelectorRelabelNamespace   = "namespace-selector-relabel-app"
	nsSelectorRelabelRouteName   = "namespace-selector-relabel-route"
)

// NamespaceSelectorRelabelTest verifies that labeling a namespace so that it newly matches a
// Gateway listener's allowedRoutes namespace selector causes an HTTPRoute in that namespace to
// become Accepted, without requiring a controller restart.
//
// See https://github.com/envoyproxy/gateway/issues/9625
var NamespaceSelectorRelabelTest = suite.ConformanceTest{
	ShortName:   "NamespaceSelectorRelabel",
	Description: "A namespace label change matching a Gateway listener's allowedRoutes selector must cause a pending HTTPRoute to become Accepted without a controller restart",
	Manifests:   []string{"testdata/namespace-selector-label.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ctx := t.Context()

		gwNN := types.NamespacedName{Name: nsSelectorRelabelGatewayName, Namespace: ConformanceInfraNamespace}
		routeNN := types.NamespacedName{Name: nsSelectorRelabelRouteName, Namespace: nsSelectorRelabelNamespace}

		// The app namespace was created without the selector label, so the HTTPRoute is
		// expected to start out rejected.
		tlog.Logf(t, "waiting for HTTPRoute %s to be rejected as NotAllowedByListeners before namespace %s carries the selector label",
			routeNN, nsSelectorRelabelNamespace)
		// Before labeling the namespace, we expect the HTTPRoute to be rejected because the namespace does not match the listener's allowedRoutes selector.
		// and the Gateway must be programmed before the HTTPRoute can be evaluated.
		//
		// Wait for the Gateway's conditions to reflect its current generation first:
		// GatewayMustHaveCondition treats a stale observedGeneration as a fatal error rather than
		// a retryable one, so calling it right after creating the Gateway can race with the
		// controller's first reconcile and fail before the configured timeout is ever used.
		kubernetes.GatewayMustHaveLatestConditions(t, suite.Client, suite.TimeoutConfig, gwNN)
		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gwapiv1.GatewayConditionProgrammed),
			Status: metav1.ConditionTrue,
			Reason: string(gwapiv1.GatewayReasonProgrammed),
		})
		kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN, metav1.Condition{
			Type:   string(gwapiv1.RouteConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(gwapiv1.RouteReasonNotAllowedByListeners),
		})

		tlog.Logf(t, "labeling namespace %s to match the listener's allowedRoutes selector", nsSelectorRelabelNamespace)
		require.NoError(t, wait.PollUntilContextTimeout(ctx, time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
			latest := &corev1.Namespace{}
			if err := suite.Client.Get(ctx, types.NamespacedName{Name: nsSelectorRelabelNamespace}, latest); err != nil {
				return false, err
			}
			if latest.Labels == nil {
				latest.Labels = map[string]string{}
			}
			latest.Labels[nsSelectorRelabelLabelKey] = nsSelectorRelabelLabelValue
			if err := suite.Client.Update(ctx, latest); err != nil {
				tlog.Logf(t, "failed to label namespace %s, retrying: %v", nsSelectorRelabelNamespace, err)
				return false, nil
			}
			return true, nil
		}))

		// This is the regression check for https://github.com/envoyproxy/gateway/issues/9625:
		// the controller must react to the namespace label change on its own. Prior to the fix,
		// this only passed after restarting the envoy-gateway controller.
		tlog.Logf(t, "waiting for HTTPRoute %s to become Accepted without a controller restart", routeNN)
		kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN, metav1.Condition{
			Type:   string(gwapiv1.RouteConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(gwapiv1.RouteReasonAccepted),
		})
	},
}
