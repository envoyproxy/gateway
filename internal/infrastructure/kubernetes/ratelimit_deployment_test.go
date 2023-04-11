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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

func testExpectedRateLimitDeployment(t *testing.T, envoyGateway *egcfgv1a1.EnvoyGateway, rateLimitInfra *ir.RateLimitInfra, expected *corev1.ResourceRequirements) {
	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)
	kube.EnvoyGateway = envoyGateway

	deployment := kube.expectedRateLimitDeployment(rateLimitInfra)

	// Check the deployment name is as expected.
	assert.Equal(t, deployment.Name, rateLimitInfraName)

	// Check container details, i.e. env vars, labels, etc. for the deployment are as expected.
	container := checkContainer(t, deployment, rateLimitInfraName, true)
	checkContainerImage(t, container, rateLimitInfraImage)
	checkContainerResources(t, container, expected)
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

	// Make sure no pod annotations are set by default
	checkPodAnnotations(t, deployment, nil)
}

func TestExpectedRateLimitDeployment(t *testing.T) {
	envoyGateway := &egcfgv1a1.EnvoyGateway{
		TypeMeta: metav1.TypeMeta{},
		EnvoyGatewaySpec: egcfgv1a1.EnvoyGatewaySpec{
			RateLimit: &egcfgv1a1.RateLimit{
				Backend: egcfgv1a1.RateLimitDatabaseBackend{
					Type:  egcfgv1a1.RedisBackendType,
					Redis: &egcfgv1a1.RateLimitRedisSettings{URL: ""},
				},
			},
		},
	}
	rateLimitInfra := &ir.RateLimitInfra{}
	testExpectedRateLimitDeployment(t, envoyGateway, rateLimitInfra, egcfgv1a1.DefaultResourceRequirements())
}

func TestExpectedRateLimitDeploymentForSpecifiedResources(t *testing.T) {
	requirements := corev1.ResourceRequirements{
		Limits: nil,
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("10m"),
			corev1.ResourceMemory: resource.MustParse("128Mi")},
		Claims: nil,
	}
	envoyGateway := &egcfgv1a1.EnvoyGateway{
		TypeMeta: metav1.TypeMeta{},
		EnvoyGatewaySpec: egcfgv1a1.EnvoyGatewaySpec{
			Provider: &egcfgv1a1.EnvoyGatewayProvider{
				Type: egcfgv1a1.ProviderTypeKubernetes,
				Kubernetes: &egcfgv1a1.EnvoyGatewayKubernetesProvider{RateLimitDeployment: &egcfgv1a1.KubernetesDeploymentSpec{
					Container: &egcfgv1a1.KubernetesContainerSpec{Resources: &requirements}},
				},
			},
			RateLimit: &egcfgv1a1.RateLimit{
				Backend: egcfgv1a1.RateLimitDatabaseBackend{
					Type:  egcfgv1a1.RedisBackendType,
					Redis: &egcfgv1a1.RateLimitRedisSettings{URL: ""},
				},
			},
		},
	}
	rateLimitInfra := &ir.RateLimitInfra{}
	testExpectedRateLimitDeployment(t, envoyGateway, rateLimitInfra, &requirements)
}

func TestExpectedRateLimitPodAnnotations(t *testing.T) {
	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)

	// Set service annotations into EnvoyProxy API and ensure the same
	// value is set in the generated service.
	annotations := map[string]string{
		"key1": "val1",
		"key2": "val2",
	}

	kube.EnvoyGateway = &egcfgv1a1.EnvoyGateway{
		TypeMeta: metav1.TypeMeta{},
		EnvoyGatewaySpec: egcfgv1a1.EnvoyGatewaySpec{
			Provider: &egcfgv1a1.EnvoyGatewayProvider{
				Type: egcfgv1a1.ProviderTypeKubernetes,
				Kubernetes: &egcfgv1a1.EnvoyGatewayKubernetesProvider{RateLimitDeployment: &egcfgv1a1.KubernetesDeploymentSpec{
					Pod: &egcfgv1a1.KubernetesPodSpec{Annotations: annotations}},
				},
			},
			RateLimit: &egcfgv1a1.RateLimit{
				Backend: egcfgv1a1.RateLimitDatabaseBackend{
					Type:  egcfgv1a1.RedisBackendType,
					Redis: &egcfgv1a1.RateLimitRedisSettings{URL: ""},
				},
			},
		},
	}

	rateLimitInfra := &ir.RateLimitInfra{}

	deploy := kube.expectedRateLimitDeployment(rateLimitInfra)
	checkPodAnnotations(t, deploy, annotations)
}

func TestCreateOrUpdateRateLimitDeployment(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	kube := NewInfra(nil, cfg)
	kube.EnvoyGateway = &egcfgv1a1.EnvoyGateway{
		TypeMeta: metav1.TypeMeta{},
		EnvoyGatewaySpec: egcfgv1a1.EnvoyGatewaySpec{
			Provider: &egcfgv1a1.EnvoyGatewayProvider{
				Type: egcfgv1a1.ProviderTypeKubernetes,
			},
			RateLimit: &egcfgv1a1.RateLimit{
				Backend: egcfgv1a1.RateLimitDatabaseBackend{
					Type:  egcfgv1a1.RedisBackendType,
					Redis: &egcfgv1a1.RateLimitRedisSettings{URL: ""},
				},
			},
		},
	}
	rateLimitInfra := &ir.RateLimitInfra{}
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
			in:      &ir.RateLimitInfra{},
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
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				Namespace: "test",
				EnvoyGateway: &egcfgv1a1.EnvoyGateway{
					TypeMeta: metav1.TypeMeta{},
					EnvoyGatewaySpec: egcfgv1a1.EnvoyGatewaySpec{
						Provider: &egcfgv1a1.EnvoyGatewayProvider{
							Type: egcfgv1a1.ProviderTypeKubernetes,
						},
						RateLimit: &egcfgv1a1.RateLimit{
							Backend: egcfgv1a1.RateLimitDatabaseBackend{
								Type:  egcfgv1a1.RedisBackendType,
								Redis: &egcfgv1a1.RateLimitRedisSettings{URL: ""},
							},
						},
					},
				},
			}
			rateLimitInfra := &ir.RateLimitInfra{}
			err := kube.createOrUpdateRateLimitDeployment(context.Background(), rateLimitInfra)
			require.NoError(t, err)

			err = kube.deleteRateLimitDeployment(context.Background(), rateLimitInfra)
			require.NoError(t, err)
		})
	}
}
