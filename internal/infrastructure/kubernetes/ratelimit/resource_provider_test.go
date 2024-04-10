// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

var (
	overrideTestData = flag.Bool("override-testdata", false, "if override the test output data.")
)

const (
	// RedisAuthEnvVar is the redis auth.
	RedisAuthEnvVar = "REDIS_AUTH"
)

var ownerReferenceUID = map[string]types.UID{
	ResourceKindService:        "test-owner-reference-uid-for-service",
	ResourceKindDeployment:     "test-owner-reference-uid-for-deployment",
	ResourceKindServiceAccount: "test-owner-reference-uid-for-service-account",
}

func TestRateLimitLabelSelector(t *testing.T) {

	cases := []struct {
		name     string
		expected []string
	}{
		{
			name: "rateLimit-labelSelector",
			expected: []string{
				"app.kubernetes.io/name=envoy-ratelimit",
				"app.kubernetes.io/component=ratelimit",
				"app.kubernetes.io/managed-by=envoy-gateway",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := LabelSelector()
			require.ElementsMatch(t, tc.expected, got)
		})
	}

}

func TestRateLimitLabels(t *testing.T) {
	cases := []struct {
		name     string
		expected map[string]string
	}{
		{
			name: "rateLimit-labels",
			expected: map[string]string{
				"app.kubernetes.io/name":       InfraName,
				"app.kubernetes.io/component":  "ratelimit",
				"app.kubernetes.io/managed-by": "envoy-gateway",
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

	cfg.EnvoyGateway.RateLimit = &egv1a1.RateLimit{
		Backend: egv1a1.RateLimitDatabaseBackend{
			Type: egv1a1.RedisBackendType,
			Redis: &egv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}
	r := NewResourceRender(cfg.Namespace, cfg.EnvoyGateway, ownerReferenceUID)

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

func TestService(t *testing.T) {

	cfg, err := config.New()
	require.NoError(t, err)

	cases := []struct {
		caseName  string
		rateLimit *egv1a1.RateLimit
	}{
		{
			caseName: "default",
			rateLimit: &egv1a1.RateLimit{
				Backend: egv1a1.RateLimitDatabaseBackend{
					Type: egv1a1.RedisBackendType,
					Redis: &egv1a1.RateLimitRedisSettings{
						URL: "redis.redis.svc:6379",
					},
				},
			},
		},
		{
			caseName: "no-prometheus",
			rateLimit: &egv1a1.RateLimit{
				Backend: egv1a1.RateLimitDatabaseBackend{
					Type: egv1a1.RedisBackendType,
					Redis: &egv1a1.RateLimitRedisSettings{
						URL: "redis.redis.svc:6379",
					},
				},
				Telemetry: &egv1a1.RateLimitTelemetry{
					Metrics: &egv1a1.RateLimitMetrics{
						Prometheus: &egv1a1.RateLimitMetricsPrometheusProvider{
							Disable: true,
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		cfg.EnvoyGateway.RateLimit = tc.rateLimit
		r := NewResourceRender(cfg.Namespace, cfg.EnvoyGateway, ownerReferenceUID)

		svc, err := r.Service()
		require.NoError(t, err)

		expected, err := loadService(tc.caseName)
		require.NoError(t, err)

		assert.Equal(t, expected, svc)
	}
}

func loadService(caseName string) (*corev1.Service, error) {
	serviceYAML, err := os.ReadFile(fmt.Sprintf("testdata/services/envoy-ratelimit-service-%s.yaml", caseName))
	if err != nil {
		return nil, err
	}
	svc := &corev1.Service{}
	_ = yaml.Unmarshal(serviceYAML, svc)
	return svc, nil
}

func TestConfigmap(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	cfg.EnvoyGateway.RateLimit = &egv1a1.RateLimit{
		Backend: egv1a1.RateLimitDatabaseBackend{
			Type: egv1a1.RedisBackendType,
			Redis: &egv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}
	r := NewResourceRender(cfg.Namespace, cfg.EnvoyGateway, ownerReferenceUID)
	cm, err := r.ConfigMap()
	require.NoError(t, err)

	if *overrideTestData {
		cmYAML, err := yaml.Marshal(cm)
		require.NoError(t, err)
		// nolint:gosec
		err = os.WriteFile("testdata/envoy-ratelimit-configmap.yaml", cmYAML, 0644)
		require.NoError(t, err)
		return
	}

	expected, err := loadConfigmap()
	require.NoError(t, err)

	assert.Equal(t, expected, cm)
}

func loadConfigmap() (*corev1.ConfigMap, error) {
	configmapYAML, err := os.ReadFile("testdata/envoy-ratelimit-configmap.yaml")
	if err != nil {
		return nil, err
	}
	cm := &corev1.ConfigMap{}
	_ = yaml.Unmarshal(configmapYAML, cm)
	return cm, nil
}

func TestDeployment(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	rateLimit := &egv1a1.RateLimit{
		Backend: egv1a1.RateLimitDatabaseBackend{
			Type: egv1a1.RedisBackendType,
			Redis: &egv1a1.RateLimitRedisSettings{
				URL: "redis.redis.svc:6379",
			},
		},
	}
	cases := []struct {
		caseName  string
		rateLimit *egv1a1.RateLimit
		deploy    *egv1a1.KubernetesDeploymentSpec
	}{
		{
			caseName:  "default",
			rateLimit: rateLimit,
			deploy:    cfg.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment,
		},
		{
			caseName: "disable-prometheus",
			rateLimit: &egv1a1.RateLimit{
				Backend: egv1a1.RateLimitDatabaseBackend{
					Type: egv1a1.RedisBackendType,
					Redis: &egv1a1.RateLimitRedisSettings{
						URL: "redis.redis.svc:6379",
					},
				},
				Telemetry: &egv1a1.RateLimitTelemetry{
					Metrics: &egv1a1.RateLimitMetrics{
						Prometheus: &egv1a1.RateLimitMetricsPrometheusProvider{
							Disable: true,
						},
					},
				},
			},
			deploy: cfg.EnvoyGateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment,
		},
		{
			caseName:  "patch-deployment",
			rateLimit: rateLimit,
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.StrategicMerge),
					Value: v1.JSON{
						Raw: []byte("{\"spec\":{\"template\":{\"spec\":{\"hostNetwork\":true,\"dnsPolicy\":\"ClusterFirstWithHostNet\"}}}}"),
					},
				},
			},
		},
		{
			caseName:  "custom",
			rateLimit: rateLimit,
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](2),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
					Image: ptr.To("custom-image"),
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
						Privileged: ptr.To(true),
					},
				},
			},
		},
		{
			caseName:  "extension-env",
			rateLimit: rateLimit,
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](2),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
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
					Image: ptr.To("custom-image"),
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
						Privileged: ptr.To(true),
					},
				},
			},
		},
		{
			caseName:  "default-env",
			rateLimit: rateLimit,
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](2),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
					Env:   nil,
					Image: ptr.To("custom-image"),
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
						Privileged: ptr.To(true),
					},
				},
			},
		},
		{
			caseName:  "override-env",
			rateLimit: rateLimit,
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](2),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
					Env: []corev1.EnvVar{
						{
							Name:  UseStatsdEnvVar,
							Value: "true",
						},
					},
					Image: ptr.To("custom-image"),
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
						Privileged: ptr.To(true),
					},
				},
			},
		},
		{
			caseName: "redis-tls-settings",
			rateLimit: &egv1a1.RateLimit{
				Backend: egv1a1.RateLimitDatabaseBackend{
					Type: egv1a1.RedisBackendType,
					Redis: &egv1a1.RateLimitRedisSettings{
						URL: "redis.redis.svc:6379",
						TLS: &egv1a1.RedisTLSSettings{
							CertificateRef: &gwapiv1.SecretObjectReference{
								Name: "ratelimit-cert",
							},
						},
					},
				},
			},
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](2),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
					Env: []corev1.EnvVar{
						{
							Name:  RedisAuthEnvVar,
							Value: "redis_auth_password",
						},
						{
							Name:  UseStatsdEnvVar,
							Value: "true",
						},
					},
					Image: ptr.To("custom-image"),
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
						Privileged: ptr.To(true),
					},
				},
			},
		},
		{
			caseName: "tolerations",
			rateLimit: &egv1a1.RateLimit{
				Backend: egv1a1.RateLimitDatabaseBackend{
					Type: egv1a1.RedisBackendType,
					Redis: &egv1a1.RateLimitRedisSettings{
						URL: "redis.redis.svc:6379",
						TLS: &egv1a1.RedisTLSSettings{
							CertificateRef: &gwapiv1.SecretObjectReference{
								Name: "ratelimit-cert",
							},
						},
					},
				},
			},
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](2),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-type",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
							Value:    "router",
						},
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
					Env: []corev1.EnvVar{
						{
							Name:  RedisAuthEnvVar,
							Value: "redis_auth_password",
						},
						{
							Name:  UseStatsdEnvVar,
							Value: "true",
						},
					},
					Image: ptr.To("custom-image"),
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
						Privileged: ptr.To(true),
					},
				},
			},
		},
		{
			caseName: "volumes",
			rateLimit: &egv1a1.RateLimit{
				Backend: egv1a1.RateLimitDatabaseBackend{
					Type: egv1a1.RedisBackendType,
					Redis: &egv1a1.RateLimitRedisSettings{
						URL: "redis.redis.svc:6379",
						TLS: &egv1a1.RedisTLSSettings{
							CertificateRef: &gwapiv1.SecretObjectReference{
								Name: "ratelimit-cert-origin",
							},
						},
					},
				},
			},
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](2),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-type",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
							Value:    "router",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "custom-cert",
									DefaultMode: ptr.To[int32](420),
								},
							},
						},
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
					Env: []corev1.EnvVar{
						{
							Name:  RedisAuthEnvVar,
							Value: "redis_auth_password",
						},
						{
							Name:  UseStatsdEnvVar,
							Value: "true",
						},
					},
					Image: ptr.To("custom-image"),
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
						Privileged: ptr.To(true),
					},
					VolumeMounts: []corev1.VolumeMount{},
				},
			},
		},
		{
			caseName:  "with-node-selector",
			rateLimit: rateLimit,
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Pod: &egv1a1.KubernetesPodSpec{
					NodeSelector: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
		},
		{
			caseName:  "with-topology-spread-constraints",
			rateLimit: rateLimit,
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Pod: &egv1a1.KubernetesPodSpec{
					TopologySpreadConstraints: []corev1.TopologySpreadConstraint{
						{
							MaxSkew:           1,
							TopologyKey:       "kubernetes.io/hostname",
							WhenUnsatisfiable: corev1.DoNotSchedule,
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"app": "foo"},
							},
							MatchLabelKeys: []string{"pod-template-hash"},
						},
					},
				},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			cfg.EnvoyGateway.RateLimit = tc.rateLimit

			cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
				Type: egv1a1.ProviderTypeKubernetes,
				Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
					RateLimitDeployment: tc.deploy,
				}}
			r := NewResourceRender(cfg.Namespace, cfg.EnvoyGateway, ownerReferenceUID)
			dp, err := r.Deployment()
			require.NoError(t, err)

			if *overrideTestData {
				deploymentYAML, err := yaml.Marshal(dp)
				require.NoError(t, err)
				// nolint:gosec
				err = os.WriteFile(fmt.Sprintf("testdata/deployments/%s.yaml", tc.caseName), deploymentYAML, 0644)
				require.NoError(t, err)
				return
			}

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
