// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

const defaultServiceStartupTimeout = 5 * time.Minute

// WaitForPods waits for the pods in the given namespace and with the given selector
// to be in the given phase and condition.
func WaitForPods(t *testing.T, cl client.Client, namespace string, selectors map[string]string, phase corev1.PodPhase, condition corev1.PodCondition) {
	t.Logf("waiting for %s/[%s] to be %v...", namespace, selectors, phase)

	require.Eventually(t, func() bool {
		pods := &corev1.PodList{}

		err := cl.List(context.Background(), pods, &client.ListOptions{
			Namespace:     namespace,
			LabelSelector: labels.SelectorFromSet(selectors),
		})

		if err != nil || len(pods.Items) == 0 {
			return false
		}

	checkPods:
		for _, p := range pods.Items {
			if p.Status.Phase != phase {
				return false
			}

			if p.Status.Conditions == nil {
				return false
			}

			for _, c := range p.Status.Conditions {
				if c.Type == condition.Type && c.Status == condition.Status {
					continue checkPods // pod is ready, check next pod
				}
			}

			return false
		}

		return true
	}, defaultServiceStartupTimeout, 2*time.Second)
}

// SecurityPolicyMustBeAccepted waits for the specified SecurityPolicy to be accepted.
func SecurityPolicyMustBeAccepted(t *testing.T, client client.Client, policyName types.NamespacedName, controllerName string, ancestorRef gwv1a2.ParentReference) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.SecurityPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching SecurityPolicy: %w", err)
		}

		if policyAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef) {
			return true, nil
		}

		t.Logf("SecurityPolicy not yet accepted: %v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for SecurityPolicy to be accepted")
}

// BackendTrafficPolicyMustBeAccepted waits for the specified BackendTrafficPolicy to be accepted.
func BackendTrafficPolicyMustBeAccepted(t *testing.T, client client.Client, policyName types.NamespacedName, controllerName string, ancestorRef gwv1a2.ParentReference) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.BackendTrafficPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching BackendTrafficPolicy: %w", err)
		}

		if policyAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef) {
			return true, nil
		}

		t.Logf("BackendTrafficPolicy not yet accepted: %v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for BackendTrafficPolicy to be accepted")
}

// AlmostEquals We use a solution similar to istio:
// Given an offset, calculate whether the actual value is within the offset of the expected value
func AlmostEquals(actual, expect, offset int) bool {
	upper := actual + offset
	lower := actual - offset
	if expect < lower || expect > upper {
		return false
	}
	return true
}

func policyAcceptedByAncestor(ancestors []gwv1a2.PolicyAncestorStatus, controllerName string, ancestorRef gwv1a2.ParentReference) bool {
	for _, ancestor := range ancestors {
		if string(ancestor.ControllerName) == controllerName && cmp.Equal(ancestor.AncestorRef, ancestorRef) {
			for _, condition := range ancestor.Conditions {
				if condition.Type == string(gwv1a2.PolicyConditionAccepted) && condition.Status == metav1.ConditionTrue {
					return true
				}
			}
		}
	}
	return false
}
