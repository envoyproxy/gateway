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

		const (
			ns     = "gateway-conformance-infra"
			gwName = "ownership-collision-test"
		)

		ctx := context.Background()

		preExistingSA := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      gwName,
				Labels:    map[string]string{"app": "pre-existing"},
			},
		}
		require.NoError(t, suite.Client.Create(ctx, preExistingSA))

		t.Cleanup(func() {
			_ = suite.Client.Delete(ctx, &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName},
			})
			_ = suite.Client.Delete(ctx, &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: gwName},
			})
		})

		// create a Gateway whose auto-generated SA name equals gwName, the collision case.
		gw := &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gwName,
				Namespace: ns,
			},
			Spec: gwapiv1.GatewaySpec{
				GatewayClassName: gwapiv1.ObjectName(suite.GatewayClassName),
				Listeners: []gwapiv1.Listener{
					{
						Name:     "http",
						Port:     8000,
						Protocol: "HTTP",
					},
				},
			},
		}
		require.NoError(t, suite.Client.Create(ctx, gw))

		gwNN := types.NamespacedName{Name: gwName, Namespace: ns}

		// The infra reconciler must detect the collision and surface it as a condition rather
		// than silently patching the pre-existing SA.
		tlog.Logf(t, "waiting for Gateway to surface ownership conflict")
		err := wait.PollUntilContextTimeout(ctx, time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
			func(ctx context.Context) (bool, error) {
				fetched := &gwapiv1.Gateway{}
				if err := suite.Client.Get(ctx, gwNN, fetched); err != nil {
					return false, nil
				}
				for _, cond := range fetched.Status.Conditions {
					// controller noticed the problem with Accepted=False or Programmed=False.
					if cond.Status == metav1.ConditionFalse {
						tlog.Logf(t, "Gateway has non-ready condition %s=%s: %s", cond.Type, cond.Status, cond.Message)
						return true, nil
					}
				}
				return false, nil
			})
		require.NoError(t, err, "expected Gateway to reflect an error condition due to SA ownership conflict")

		fetchedSA := &corev1.ServiceAccount{}
		require.NoError(t, suite.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: gwName}, fetchedSA))

		for _, ref := range fetchedSA.OwnerReferences {
			assert.NotEqual(t, "Gateway", ref.Kind,
				"pre-existing ServiceAccount must not have a Gateway ownerReference injected by Envoy Gateway")
		}

		assert.Equal(t, map[string]string{"app": "pre-existing"}, fetchedSA.Labels,
			"pre-existing ServiceAccount labels must not be modified")
	},
}
