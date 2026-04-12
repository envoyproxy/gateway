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

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

const (
	customDNSDomainEnvName  = "KUBERNETES_CLUSTER_DOMAIN"
	customHostRewriteDomain = "example.internal"
)

type deploymentEnvState struct {
	Value string
	Found bool
}

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteRewriteHostHeaderCustomDNSDomain)
}

var HTTPRouteRewriteHostHeaderCustomDNSDomain = suite.ConformanceTest{
	ShortName:   "HTTPRouteRewriteHostHeaderCustomDNSDomain",
	Description: "An HTTPRoute with backend host rewrite uses the configured DNS domain",
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		originalState := setEnvoyGatewayClusterDomain(t, suite, customHostRewriteDomain)
		defer restoreEnvoyGatewayClusterDomain(t, suite, originalState)

		suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/httproute-rewrite-host-custom-dnsdomain.yaml", true)

		ns := ConformanceInfraNamespace
		routeNN := types.NamespacedName{Name: "rewrite-host-custom-dnsdomain", Namespace: ns}
		gwNN := SameNamespaceGateway
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Path: "/backend-service-custom-dnsdomain",
			},
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Path: "/backend-service-custom-dnsdomain",
					Host: fmt.Sprintf("infra-backend-v1.%s.svc.%s", ns, customHostRewriteDomain),
				},
			},
			Backend:   "infra-backend-v1",
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
	},
}

func setEnvoyGatewayClusterDomain(t *testing.T, suite *suite.ConformanceTestSuite, value string) deploymentEnvState {
	t.Helper()

	deploymentNN := types.NamespacedName{Name: "envoy-gateway", Namespace: "envoy-gateway-system"}
	var originalState deploymentEnvState

	for i := 0; i < 5; i++ {
		deployment := &appsv1.Deployment{}
		err := suite.Client.Get(context.Background(), deploymentNN, deployment)
		require.NoError(t, err)

		originalState = getDeploymentEnvState(deployment.Spec.Template.Spec.Containers[0].Env, customDNSDomainEnvName)
		if originalState.Found && originalState.Value == value {
			waitForEnvoyGatewayRollout(t, suite, deploymentNN, value)
			return originalState
		}

		upsertDeploymentEnv(&deployment.Spec.Template.Spec.Containers[0].Env, customDNSDomainEnvName, value)
		err = suite.Client.Update(context.Background(), deployment)
		if err == nil {
			waitForEnvoyGatewayRollout(t, suite, deploymentNN, value)
			return originalState
		}
		if !apierrors.IsConflict(err) {
			require.NoError(t, err)
		}
	}

	t.Fatalf("failed to update %s on envoy-gateway deployment after retries", customDNSDomainEnvName)
	return deploymentEnvState{}
}

func restoreEnvoyGatewayClusterDomain(t *testing.T, suite *suite.ConformanceTestSuite, state deploymentEnvState) {
	t.Helper()

	deploymentNN := types.NamespacedName{Name: "envoy-gateway", Namespace: "envoy-gateway-system"}
	for i := 0; i < 5; i++ {
		deployment := &appsv1.Deployment{}
		err := suite.Client.Get(context.Background(), deploymentNN, deployment)
		require.NoError(t, err)

		if state.Found {
			upsertDeploymentEnv(&deployment.Spec.Template.Spec.Containers[0].Env, customDNSDomainEnvName, state.Value)
		} else {
			removeDeploymentEnv(&deployment.Spec.Template.Spec.Containers[0].Env, customDNSDomainEnvName)
		}

		err = suite.Client.Update(context.Background(), deployment)
		if err == nil {
			expected := ""
			if state.Found {
				expected = state.Value
			}
			waitForEnvoyGatewayRollout(t, suite, deploymentNN, expected)
			return
		}
		if !apierrors.IsConflict(err) {
			require.NoError(t, err)
		}
	}

	t.Fatalf("failed to restore %s on envoy-gateway deployment after retries", customDNSDomainEnvName)
}

func waitForEnvoyGatewayRollout(t *testing.T, suite *suite.ConformanceTestSuite, deploymentNN types.NamespacedName, expectedDNSDomain string) {
	t.Helper()

	require.Eventually(t, func() bool {
		deployment := &appsv1.Deployment{}
		if err := suite.Client.Get(context.Background(), deploymentNN, deployment); err != nil {
			return false
		}

		envState := getDeploymentEnvState(deployment.Spec.Template.Spec.Containers[0].Env, customDNSDomainEnvName)
		if expectedDNSDomain == "" {
			if envState.Found {
				return false
			}
		} else if !envState.Found || envState.Value != expectedDNSDomain {
			return false
		}

		replicas := int32(1)
		if deployment.Spec.Replicas != nil {
			replicas = *deployment.Spec.Replicas
		}

		return deployment.Generation <= deployment.Status.ObservedGeneration &&
			deployment.Status.UpdatedReplicas == replicas &&
			deployment.Status.ReadyReplicas == replicas &&
			deployment.Status.AvailableReplicas == replicas
	}, 2*time.Minute, 2*time.Second)

	WaitForPods(t, suite.Client, deploymentNN.Namespace, map[string]string{"control-plane": "envoy-gateway"}, corev1.PodRunning, &PodReady)
}

func getDeploymentEnvState(envs []corev1.EnvVar, name string) deploymentEnvState {
	for _, env := range envs {
		if env.Name == name {
			return deploymentEnvState{Value: env.Value, Found: true}
		}
	}
	return deploymentEnvState{}
}

func upsertDeploymentEnv(envs *[]corev1.EnvVar, name, value string) {
	for i := range *envs {
		if (*envs)[i].Name == name {
			(*envs)[i].Value = value
			(*envs)[i].ValueFrom = nil
			return
		}
	}
	*envs = append(*envs, corev1.EnvVar{Name: name, Value: value})
}

func removeDeploymentEnv(envs *[]corev1.EnvVar, name string) {
	filtered := (*envs)[:0]
	for _, env := range *envs {
		if env.Name != name {
			filtered = append(filtered, env)
		}
	}
	*envs = filtered
}
