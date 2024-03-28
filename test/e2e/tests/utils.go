// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package tests

import (
	"context"
	"fmt"
	"io"

	"testing"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"
	flog "fortio.org/log"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/config"

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

// ClientTrafficPolicyMustBeAccepted waits for the specified ClientTrafficPolicy to be accepted.
func ClientTrafficPolicyMustBeAccepted(t *testing.T, client client.Client, policyName types.NamespacedName, controllerName string, ancestorRef gwv1a2.ParentReference) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.ClientTrafficPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching ClientTrafficPolicy: %w", err)
		}

		if policyAcceptedByAncestor(policy.Status.Ancestors, controllerName, ancestorRef) {
			return true, nil
		}

		t.Logf("ClientTrafficPolicy not yet accepted: %v", policy)
		return false, nil
	})

	require.NoErrorf(t, waitErr, "error waiting for ClientTrafficPolicy to be accepted")
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

// runs a load test with options described in opts
// the done channel is used to notify caller of execution result
// the execution may end due to an external abort or timeout
func runLoadAndWait(t *testing.T, timeoutConfig config.TimeoutConfig, done chan bool, aborter *periodic.Aborter, reqURL string) {
	flog.SetLogLevel(flog.Error)
	opts := fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			QPS: 5000,
			// allow some overhead time for setting up workers and tearing down after restart
			Duration:   timeoutConfig.CreateTimeout + timeoutConfig.CreateTimeout/2,
			NumThreads: 50,
			Stop:       aborter,
			Out:        io.Discard,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: reqURL,
		},
	}

	res, err := fhttp.RunHTTPTest(&opts)
	if err != nil {
		done <- false
		t.Logf("failed to create load: %v", err)
	}

	// collect stats
	okReq := res.RetCodes[200]
	totalReq := res.DurationHistogram.Count
	failedReq := totalReq - okReq
	errorReq := res.ErrorsDurationHistogram.Count
	timedOut := res.ActualDuration == opts.Duration
	t.Logf("Load completed after %s with %d requests, %d success, %d failures and %d errors", res.ActualDuration, totalReq, okReq, failedReq, errorReq)

	if okReq == totalReq && errorReq == 0 && !timedOut {
		done <- true
	}
	done <- false
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
