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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
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
				cleanupCtx := context.Background()
				_ = suite.Client.Delete(cleanupCtx, &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
				_ = suite.Client.Delete(cleanupCtx, &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
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

			// Ensure deleting the Gateway does not garbage-collect the pre-existing ServiceAccount.
			require.NoError(t, suite.Client.Delete(ctx, gw))
			require.NoError(t, wait.PollUntilContextTimeout(ctx, time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
				func(ctx context.Context) (bool, error) {
					err := suite.Client.Get(ctx, gwNN, &gwapiv1.Gateway{})
					return apierrors.IsNotFound(err), nil
				}))
			require.NoError(t, suite.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: gwName}, &corev1.ServiceAccount{}))
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
				cleanupCtx := context.Background()
				_ = suite.Client.Delete(cleanupCtx, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
				_ = suite.Client.Delete(cleanupCtx, &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
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

			// Ensure deleting the Gateway does not garbage-collect the pre-existing ConfigMap.
			require.NoError(t, suite.Client.Delete(ctx, gw))
			require.NoError(t, wait.PollUntilContextTimeout(ctx, time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
				func(ctx context.Context) (bool, error) {
					err := suite.Client.Get(ctx, gwNN, &gwapiv1.Gateway{})
					return apierrors.IsNotFound(err), nil
				}))
			require.NoError(t, suite.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: gwName}, &corev1.ConfigMap{}))
		})

		t.Run("Deployment", func(t *testing.T) {
			gwName := "ownership-collision-deploy"
			ctx := t.Context()

			// A pre-existing, unmanaged Deployment that shares the Gateway's name — the
			// nginx scenario from #9132. It has a different selector and no envoy-gateway
			// ownership labels. replicas=0 keeps it inert: the ownership-collision
			// behavior under test needs no Pods to run, so we avoid scheduling/image-pull.
			preExistingDeploy := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      gwName,
					Labels:    map[string]string{"app": "pre-existing"},
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](0),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "pre-existing"}},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "pre-existing"}},
						Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "nginx", Image: "nginx"}}},
					},
				},
			}
			require.NoError(t, suite.Client.Create(ctx, preExistingDeploy))
			t.Cleanup(func() {
				cleanupCtx := context.Background()
				_ = suite.Client.Delete(cleanupCtx, &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
				_ = suite.Client.Delete(cleanupCtx, &gwapiv1.Gateway{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName}})
			})

			gw := newCollisionGateway(ns, gwName, suite.GatewayClassName)
			require.NoError(t, suite.Client.Create(ctx, gw))

			gwNN := types.NamespacedName{Name: gwName, Namespace: ns}
			waitForSettled(t, gwNN)

			fetchedDeploy := &appsv1.Deployment{}
			require.NoError(t, suite.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: gwName}, fetchedDeploy))
			for _, ref := range fetchedDeploy.OwnerReferences {
				assert.NotEqual(t, "Gateway", ref.Kind,
					"pre-existing Deployment must not have a Gateway ownerReference injected")
			}
			assert.Equal(t, map[string]string{"app": "pre-existing"}, fetchedDeploy.Labels,
				"pre-existing Deployment labels must not be modified")

			// Ensure deleting the Gateway does not garbage-collect the pre-existing Deployment.
			require.NoError(t, suite.Client.Delete(ctx, gw))
			require.NoError(t, wait.PollUntilContextTimeout(ctx, time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
				func(ctx context.Context) (bool, error) {
					err := suite.Client.Get(ctx, gwNN, &gwapiv1.Gateway{})
					return apierrors.IsNotFound(err), nil
				}))
			require.NoError(t, suite.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: gwName}, &appsv1.Deployment{}))
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
