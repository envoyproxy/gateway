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
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayInfraRetryTest)
}

// GatewayInfraRetryTest reproduces https://github.com/envoyproxy/gateway/issues/10000.
// A fail-closed admission webhook that cannot be called makes the API server reject
// the Envoy Service with an InternalError. Once the webhook is gone, Envoy Gateway
// must create the Service and program the Gateway without a restart, although
// nothing about the Gateway changes.
var GatewayInfraRetryTest = suite.ConformanceTest{
	ShortName:   "GatewayInfraRetry",
	Description: "Envoy Gateway retries creating proxy infra that failed on a transient admission webhook error",
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		if IsRemoteInfraMode() {
			t.Skip("the remote infra provider, not the Kubernetes API, creates the proxy infra in remote infra mode")
		}

		ctx := t.Context()
		gwNN := types.NamespacedName{Name: "infra-retry", Namespace: ConformanceInfraNamespace}
		ownerLabels := gatewayapi.GatewayOwnerLabels(gwNN.Namespace, gwNN.Name)
		restarts := envoyGatewayRestarts(t, suite.Client)

		// The webhook's service does not exist, so every call fails, and with
		// failurePolicy Fail the API server rejects matching Services.
		webhook := &admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{Name: "infra-retry.e2e.gateway.envoyproxy.io"},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{{
				Name: "infra-retry.e2e.gateway.envoyproxy.io",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Namespace: gwNN.Namespace,
						Name:      "infra-retry-missing-webhook",
					},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create, admissionregistrationv1.Update},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"services"},
					},
				}},
				ObjectSelector:          &metav1.LabelSelector{MatchLabels: ownerLabels},
				FailurePolicy:           new(admissionregistrationv1.Fail),
				SideEffects:             new(admissionregistrationv1.SideEffectClassNone),
				AdmissionReviewVersions: []string{"v1"},
			}},
		}
		require.NoError(t, suite.Client.Create(ctx, webhook))
		t.Cleanup(func() {
			_ = suite.Client.Delete(context.Background(), webhook)
		})

		// The webhook takes effect asynchronously. Wait until a dry run of a
		// matching Service fails the way the Envoy Service will.
		probe := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "infra-retry-probe", Namespace: gwNN.Namespace, Labels: ownerLabels},
			Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "http", Port: 80}}},
		}
		require.NoError(t, wait.PollUntilContextTimeout(ctx, time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true,
			func(ctx context.Context) (bool, error) {
				return apierrors.IsInternalError(suite.Client.Create(ctx, probe.DeepCopy(), client.DryRunAll)), nil
			}), "the webhook was not enforced")

		gw := &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{Name: gwNN.Name, Namespace: gwNN.Namespace},
			Spec: gwapiv1.GatewaySpec{
				GatewayClassName: gwapiv1.ObjectName(suite.GatewayClassName),
				Listeners:        []gwapiv1.Listener{{Name: "http", Port: 8000, Protocol: gwapiv1.HTTPProtocolType}},
			},
		}
		require.NoError(t, suite.Client.Create(ctx, gw))
		t.Cleanup(func() {
			_ = suite.Client.Delete(context.Background(), gw)
		})
		// Cleanups run in reverse order, so this logs the state before the
		// Gateway and the webhook are deleted, unlike the suite's failure dump.
		t.Cleanup(func() {
			if t.Failed() {
				logInfraRetryState(t, suite.Client, gwNN, ownerLabels)
			}
		})

		// Envoy Gateway applies the Service right after the Deployment, in the
		// same call. Make sure that call has failed before removing the webhook,
		// so that only a retry can create the Service.
		tlog.Logf(t, "waiting for the Envoy Deployment of Gateway %s", gwNN)
		require.NoError(t, wait.PollUntilContextTimeout(ctx, time.Second, suite.TimeoutConfig.CreateTimeout, true,
			func(context.Context) (bool, error) {
				_, err := getDeploymentForGateway(gwNN.Namespace, gwNN.Name, suite.Client)
				return err == nil, nil
			}), "the Envoy Deployment was not created")
		require.Never(t, func() bool {
			services, err := envoyServices(ctx, suite.Client, ownerLabels)
			return err == nil && len(services) > 0
		}, 5*time.Second, time.Second, "the webhook did not block the Envoy Service")
		kubernetes.GatewayMustHaveLatestConditions(t, suite.Client, suite.TimeoutConfig, gwNN)
		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gwapiv1.GatewayConditionProgrammed),
			Status: metav1.ConditionFalse,
		})

		require.NoError(t, suite.Client.Delete(ctx, webhook))
		tlog.Logf(t, "removed the webhook, waiting for Gateway %s to be programmed", gwNN)
		kubernetes.GatewayMustHaveLatestConditions(t, suite.Client, suite.TimeoutConfig, gwNN)
		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gwapiv1.GatewayConditionProgrammed),
			Status: metav1.ConditionTrue,
			Reason: string(gwapiv1.GatewayReasonProgrammed),
		})
		require.Equal(t, restarts, envoyGatewayRestarts(t, suite.Client), "Envoy Gateway restarted")
	},
}

// envoyGatewayRestarts returns the number of container restarts of each Envoy
// Gateway pod.
func envoyGatewayRestarts(t *testing.T, c client.Client) map[string]int32 {
	t.Helper()
	pods := &corev1.PodList{}
	require.NoError(t, c.List(t.Context(), pods, client.InNamespace("envoy-gateway-system"),
		client.MatchingLabels{"control-plane": "envoy-gateway"}))
	require.NotEmpty(t, pods.Items)
	restarts := make(map[string]int32, len(pods.Items))
	for i := range pods.Items {
		statuses := pods.Items[i].Status.ContainerStatuses
		for j := range statuses {
			restarts[pods.Items[i].Name] += statuses[j].RestartCount
		}
	}
	return restarts
}

// envoyServices lists the Envoy Services that carry ownerLabels.
func envoyServices(ctx context.Context, c client.Client, ownerLabels map[string]string) ([]corev1.Service, error) {
	services := &corev1.ServiceList{}
	err := c.List(ctx, services, client.InNamespace(GetGatewayResourceNamespace()), client.MatchingLabels(ownerLabels))
	return services.Items, err
}

// logInfraRetryState logs the Gateway's conditions and its Envoy Deployment
// and Services.
func logInfraRetryState(t *testing.T, c client.Client, gwNN types.NamespacedName, ownerLabels map[string]string) {
	t.Helper()
	gw := &gwapiv1.Gateway{}
	if err := c.Get(context.Background(), gwNN, gw); err != nil {
		tlog.Logf(t, "failed to get Gateway %s: %v", gwNN, err)
	} else {
		tlog.Logf(t, "Gateway %s generation %d conditions: %+v", gwNN, gw.Generation, gw.Status.Conditions)
	}
	deployments := &appsv1.DeploymentList{}
	if err := c.List(context.Background(), deployments, client.InNamespace(GetGatewayResourceNamespace()),
		client.MatchingLabels(ownerLabels)); err != nil {
		tlog.Logf(t, "failed to list the Envoy Deployments: %v", err)
	}
	for i := range deployments.Items {
		tlog.Logf(t, "Envoy Deployment %s: %d/%d replicas available", deployments.Items[i].Name,
			deployments.Items[i].Status.AvailableReplicas, deployments.Items[i].Status.Replicas)
	}
	services, err := envoyServices(context.Background(), c, ownerLabels)
	if err != nil {
		tlog.Logf(t, "failed to list the Envoy Services: %v", err)
	}
	tlog.Logf(t, "%d Envoy Service(s) for Gateway %s", len(services), gwNN)
}
