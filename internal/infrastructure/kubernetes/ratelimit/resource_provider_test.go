// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/yaml"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

var (
	rateLimitListener = "ratelimit-listener"
	rateLimitConfig   = `
domain: first-listener
descriptors:
  - key: first-route-key-rule-0-match-0
    value: first-route-value-rule-0-match-0
    rate_limit:
      requests_per_unit: 5
      unit: second
      unlimited: false
      name: ""
      replaces: []
    descriptors: []
    shadow_mode: false
`
)

func TestRateLimitLabels(t *testing.T) {
	cases := []struct {
		name     string
		expected map[string]string
	}{
		{
			name: "ratelimit-labels",
			expected: map[string]string{
				"app.gateway.envoyproxy.io/name": InfraName,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := rateLimitLabels()
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestServiceAccount(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	rateLimitInfra := new(ir.RateLimitInfra)
	cfg.EnvoyGateway.RateLimit = &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}
	r := NewResourceRender(cfg.Namespace, rateLimitInfra, cfg.EnvoyGateway)

	sa, err := r.ServiceAccount()
	require.NoError(t, err)

	expected, err := loadServiceAccount()
	require.NoError(t, err)

	assert.Equal(t, expected, sa)
}

func loadServiceAccount() (*corev1.ServiceAccount, error) {
	saYAML, err := os.ReadFile("testdata/envoy-ratelimit-serviceaccount.yaml")
	if err != nil {
		return nil, err
	}
	sa := &corev1.ServiceAccount{}
	_ = yaml.Unmarshal(saYAML, sa)
	return sa, nil
}

func TestConfigMap(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	rateLimitInfra := &ir.RateLimitInfra{
		ServiceConfigs: []*ir.RateLimitServiceConfig{
			{
				Name:   rateLimitListener,
				Config: rateLimitConfig,
			},
		},
	}
	cfg.EnvoyGateway.RateLimit = &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}

	r := NewResourceRender(cfg.Namespace, rateLimitInfra, cfg.EnvoyGateway)
	cm, err := r.ConfigMap()
	require.NoError(t, err)

	expected, err := loadConfigmap()
	require.NoError(t, err)

	assert.Equal(t, expected, cm)
}

func loadConfigmap() (*corev1.ConfigMap, error) {
	cmYAML, err := os.ReadFile("testdata/envoy-ratelimit-configmap.yaml")
	if err != nil {
		return nil, err
	}
	cm := &corev1.ConfigMap{}
	_ = yaml.Unmarshal(cmYAML, cm)
	return cm, nil
}

func TestService(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	rateLimitInfra := &ir.RateLimitInfra{
		ServiceConfigs: []*ir.RateLimitServiceConfig{
			{
				Name:   rateLimitListener,
				Config: rateLimitConfig,
			},
		},
	}
	cfg.EnvoyGateway.RateLimit = &egcfgv1a1.RateLimit{
		Backend: egcfgv1a1.RateLimitDatabaseBackend{
			Type: egcfgv1a1.RedisBackendType,
			Redis: &egcfgv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}
	r := NewResourceRender(cfg.Namespace, rateLimitInfra, cfg.EnvoyGateway)
	svc, err := r.Service()
	require.NoError(t, err)

	expected, err := loadService()
	require.NoError(t, err)

	assert.Equal(t, expected, svc)
}

func loadService() (*corev1.Service, error) {
	serviceYAML, err := os.ReadFile("testdata/envoy-ratelimit-service.yaml")
	if err != nil {
		return nil, err
	}
	svc := &corev1.Service{}
	_ = yaml.Unmarshal(serviceYAML, svc)
	return svc, nil
}

func TestDeployment(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	cases := []struct {
		caseName string
		deploy   *egcfgv1a1.KubernetesDeploymentSpec
	}{
		{
			caseName: "default",
			deploy:   cfg.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment,
		},
		{
			caseName: "custom",
			deploy: &egcfgv1a1.KubernetesDeploymentSpec{
				Replicas: pointer.Int32(2),
				Pod: &egcfgv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: pointer.Int64(1000),
					},
				},
				Container: &egcfgv1a1.KubernetesContainerSpec{
					Image: pointer.String("custom-image"),
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: pointer.Bool(true),
					},
				},
			},
		},
		{
			caseName: "extension-env",
			deploy: &egcfgv1a1.KubernetesDeploymentSpec{
				Replicas: pointer.Int32(2),
				Pod: &egcfgv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: pointer.Int64(1000),
					},
				},
				Container: &egcfgv1a1.KubernetesContainerSpec{
					Env: []corev1.EnvVar{
						{
							Name:  "env_a",
							Value: "env_a_value",
						},
						{
							Name:  "env_b",
							Value: "env_b_value",
						},
					},
					Image: pointer.String("custom-image"),
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: pointer.Bool(true),
					},
				},
			},
		},
		{
			caseName: "default-env",
			deploy: &egcfgv1a1.KubernetesDeploymentSpec{
				Replicas: pointer.Int32(2),
				Pod: &egcfgv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: pointer.Int64(1000),
					},
				},
				Container: &egcfgv1a1.KubernetesContainerSpec{
					Env:   nil,
					Image: pointer.String("custom-image"),
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: pointer.Bool(true),
					},
				},
			},
		},
		{
			caseName: "override-env",
			deploy: &egcfgv1a1.KubernetesDeploymentSpec{
				Replicas: pointer.Int32(2),
				Pod: &egcfgv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: pointer.Int64(1000),
					},
				},
				Container: &egcfgv1a1.KubernetesContainerSpec{
					Env: []corev1.EnvVar{
						{
							Name:  UseStatsdEnvVar,
							Value: "true",
						},
					},
					Image: pointer.String("custom-image"),
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("400m"),
							corev1.ResourceMemory: resource.MustParse("2Gi"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("200m"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: pointer.Bool(true),
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			rateLimitInfra := &ir.RateLimitInfra{
				ServiceConfigs: []*ir.RateLimitServiceConfig{
					{
						Name:   rateLimitListener,
						Config: rateLimitConfig,
					},
				},
			}

			cfg.EnvoyGateway.RateLimit = &egcfgv1a1.RateLimit{
				Backend: egcfgv1a1.RateLimitDatabaseBackend{
					Type: egcfgv1a1.RedisBackendType,
					Redis: &egcfgv1a1.RateLimitRedisSettings{
						URL: "redis.redis.svc:6379",
					},
				},
			}
			cfg.EnvoyGateway.Provider = &egcfgv1a1.EnvoyGatewayProvider{
				Type: egcfgv1a1.ProviderTypeKubernetes,
				Kubernetes: &egcfgv1a1.EnvoyGatewayKubernetesProvider{
					RateLimitDeployment: tc.deploy,
				}}
			r := NewResourceRender(cfg.Namespace, rateLimitInfra, cfg.EnvoyGateway)
			dp, err := r.Deployment()
			require.NoError(t, err)

			expected, err := loadDeployment(tc.caseName)
			require.NoError(t, err)

			assert.Equal(t, expected, dp)
		})
	}
}

func loadDeployment(caseName string) (*appsv1.Deployment, error) {
	deploymentYAML, err := os.ReadFile(fmt.Sprintf("testdata/deployments/%s.yaml", caseName))
	if err != nil {
		return nil, err
	}
	deployment := &appsv1.Deployment{}
	_ = yaml.Unmarshal(deploymentYAML, deployment)
	return deployment, nil
}

func TestGetServiceURL(t *testing.T) {
	got := GetServiceURL("envoy-gateway-system", "example-cluster.local")
	assert.Equal(t, "grpc://envoy-ratelimit.envoy-gateway-system.svc.example-cluster.local:8081", got)
}
