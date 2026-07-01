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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayNamespaceOwnership)
}

var GatewayNamespaceOwnership = suite.ConformanceTest{
	ShortName:   "GatewayNamespaceOwnership",
	Description: "Pre-existing SA/ConfigMap must not be hijacked when a Gateway name collides in GatewayNamespace mode",
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		if !IsGatewayNamespaceMode() {
			t.Skip("GatewayNamespaceOwnership test only applies to gateway-namespace-mode")
		}

		const ns = "gateway-conformance-infra"

		// waitForSettled blocks until the Gateway's Programmed condition is False,
		// ObservedGeneration matches Generation, and the condition is stable.
		waitForSettled := func(t *testing.T, gwNN types.NamespacedName) {
			t.Helper()
			var settledTransitionTime metav1.Time
			var settledGeneration int64
			err := wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
				func(ctx context.Context) (bool, error) {
					fetched := &gwapiv1.Gateway{}
					if err := suite.Client.Get(ctx, gwNN, fetched); err != nil {
						return false, nil
					}
					var programmed *metav1.Condition
					for i := range fetched.Status.Conditions {
						if fetched.Status.Conditions[i].Type == string(gwapiv1.GatewayConditionProgrammed) {
							programmed = &fetched.Status.Conditions[i]
							break
						}
					}
					if programmed == nil || programmed.Status != metav1.ConditionFalse {
						return false, nil
					}
					if programmed.ObservedGeneration != fetched.Generation {
						return false, nil
					}
					if settledGeneration == fetched.Generation &&
						settledTransitionTime.Equal(&programmed.LastTransitionTime) {
						tlog.Logf(t, "Gateway Programmed=False settled: reason=%s message=%s",
							programmed.Reason, programmed.Message)
						return true, nil
					}
					settledGeneration = fetched.Generation
					settledTransitionTime = programmed.LastTransitionTime
					return false, nil
				})
			require.NoError(t, err, "expected Gateway Programmed=False to settle after ownership conflict")
		}

		t.Run("ServiceAccount", func(t *testing.T) {
			gwName := "ownership-collision-sa"
			ctx := t.Context()

			preExistingSA := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      gwName,
					Labels:    map[string]string{"app": "pre-existing"},
				},
			}
			require.NoError(t, suite.Client.Create(ctx, preExistingSA))
			t.Cleanup(func() {
				_ = suite.Client.Delete(ctx, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
				_ = suite.Client.Delete(ctx, &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
			})

			gw := newCollisionGateway(ns, gwName, suite.GatewayClassName)
			require.NoError(t, suite.Client.Create(ctx, gw))

			gwNN := types.NamespacedName{Name: gwName, Namespace: ns}
			waitForSettled(t, gwNN)

			fetchedSA := &corev1.ServiceAccount{}
			require.NoError(t, suite.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: gwName}, fetchedSA))
			for _, ref := range fetchedSA.OwnerReferences {
				assert.NotEqual(t, "Gateway", ref.Kind,
					"pre-existing ServiceAccount must not have a Gateway ownerReference injected")
			}
			assert.Equal(t, map[string]string{"app": "pre-existing"}, fetchedSA.Labels,
				"pre-existing ServiceAccount labels must not be modified")
		})

		t.Run("ConfigMap", func(t *testing.T) {
			gwName := "ownership-collision-cm"
			ctx := t.Context()

			preExistingCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      gwName,
					Labels:    map[string]string{"app": "pre-existing"},
				},
			}
			require.NoError(t, suite.Client.Create(ctx, preExistingCM))
			t.Cleanup(func() {
				_ = suite.Client.Delete(ctx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
				_ = suite.Client.Delete(ctx, &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
			})

			gw := newCollisionGateway(ns, gwName, suite.GatewayClassName)
			require.NoError(t, suite.Client.Create(ctx, gw))

			gwNN := types.NamespacedName{Name: gwName, Namespace: ns}
			waitForSettled(t, gwNN)

			fetchedCM := &corev1.ConfigMap{}
			require.NoError(t, suite.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: gwName}, fetchedCM))
			for _, ref := range fetchedCM.OwnerReferences {
				assert.NotEqual(t, "Gateway", ref.Kind,
					"pre-existing ConfigMap must not have a Gateway ownerReference injected")
			}
			assert.Equal(t, map[string]string{"app": "pre-existing"}, fetchedCM.Labels,
				"pre-existing ConfigMap labels must not be modified")
		})
	},
}

// newCollisionGateway returns a minimal Gateway whose auto-generated infra resource
// names equal gwName in GatewayNamespace mode.
func newCollisionGateway(ns, gwName, gwClassName string) *gwapiv1.Gateway {
	return &gwapiv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{Name: gwName, Namespace: ns},
		Spec: gwapiv1.GatewaySpec{
			GatewayClassName: gwapiv1.ObjectName(gwClassName),
			Listeners: []gwapiv1.Listener{
				{Name: "http", Port: 8000, Protocol: "HTTP"},
			},
		},
	}
}
