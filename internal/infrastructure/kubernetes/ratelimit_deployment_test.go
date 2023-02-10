// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExpectedRateLimitDeployment(t *testing.T) {
	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)

	rateLimitInfra := new(ir.RateLimitInfra)
	rateLimitInfra.Backend = &ir.RateLimitDBBackend{Redis: &ir.RateLimitRedis{URL: ""}}

	deployment := kube.expectedRateLimitDeployment(rateLimitInfra)

	// Check the deployment name is as expected.
	assert.Equal(t, deployment.Name, rateLimitInfraName)

	// Check container details, i.e. env vars, labels, etc. for the deployment are as expected.
	container := checkContainer(t, deployment, rateLimitInfraName, true)
	checkContainerImage(t, container, rateLimitInfraImage)
	checkEnvVar(t, deployment, rateLimitInfraName, rateLimitRedisSocketTypeEnvVar)
	checkEnvVar(t, deployment, rateLimitInfraName, rateLimitRedisURLEnvVar)
	checkEnvVar(t, deployment, rateLimitInfraName, rateLimitRuntimeRootEnvVar)
	checkEnvVar(t, deployment, rateLimitInfraName, rateLimitRuntimeSubdirectoryEnvVar)
	checkEnvVar(t, deployment, rateLimitInfraName, rateLimitRuntimeIgnoreDotfilesEnvVar)
	checkEnvVar(t, deployment, rateLimitInfraName, rateLimitRuntimeWatchRootEnvVar)
	checkEnvVar(t, deployment, rateLimitInfraName, rateLimitLogLevelEnvVar)
	checkEnvVar(t, deployment, rateLimitInfraName, rateLimitUseStatsdEnvVar)
	checkLabels(t, deployment, deployment.Labels)

	// Check container ports for the deployment are as expected.
	ports := []int32{rateLimitInfraGRPCPort}
	for _, port := range ports {
		checkContainerHasPort(t, deployment, port)
	}

	// Set the deployment replicas.
	repl := int32(1)
	// Check the number of replicas is as expected.
	assert.Equal(t, repl, *deployment.Spec.Replicas)
}

func TestCreateOrUpdateRateLimitDeployment(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	kube := NewInfra(nil, cfg)
	rateLimitInfra := new(ir.RateLimitInfra)
	rateLimitInfra.Backend = &ir.RateLimitDBBackend{Redis: &ir.RateLimitRedis{URL: ""}}

	deployment := kube.expectedRateLimitDeployment(rateLimitInfra)

	testCases := []struct {
		name    string
		in      *ir.RateLimitInfra
		current *appsv1.Deployment
		want    *appsv1.Deployment
	}{
		{
			name: "create ratelimit deployment",
			in:   rateLimitInfra,
			want: deployment,
		},
		{
			name:    "ratelimit deployment exists",
			in:      rateLimitInfra,
			current: deployment,
			want:    deployment,
		},
		{
			name:    "update ratelimit deployment image",
			in:      &ir.RateLimitInfra{Backend: &ir.RateLimitDBBackend{Redis: &ir.RateLimitRedis{URL: ""}}},
			current: deployment,
			want:    deploymentWithImage(deployment, rateLimitInfraImage),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.current != nil {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			} else {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			}
			err := kube.createOrUpdateRateLimitDeployment(context.Background(), tc.in)
			require.NoError(t, err)

			actual := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      rateLimitInfraName,
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))
			require.Equal(t, tc.want.Spec, actual.Spec)
		})
	}
}

func TestDeleteRateLimitDeployment(t *testing.T) {
	testCases := []struct {
		name   string
		expect bool
	}{
		{
			name:   "delete ratelimit deployment",
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				Namespace: "test",
			}
			rateLimitInfra := new(ir.RateLimitInfra)
			rateLimitInfra.Backend = &ir.RateLimitDBBackend{Redis: &ir.RateLimitRedis{URL: ""}}

			err := kube.createOrUpdateRateLimitDeployment(context.Background(), rateLimitInfra)
			require.NoError(t, err)

			err = kube.deleteRateLimitDeployment(context.Background(), rateLimitInfra)
			require.NoError(t, err)
		})
	}
}
