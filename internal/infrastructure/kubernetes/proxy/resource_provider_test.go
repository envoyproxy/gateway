// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"
	"os"
	"sort"
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
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	// envoyHTTPPort is the container port number of Envoy's HTTP endpoint.
	envoyHTTPPort = int32(8080)
	// envoyHTTPSPort is the container port number of Envoy's HTTPS endpoint.
	envoyHTTPSPort = int32(8443)
)

func newTestInfra() *ir.Infra {
	i := ir.NewInfra()

	i.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	i.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = i.Proxy.Name
	i.Proxy.Listeners = []ir.ProxyListener{
		{
			Ports: []ir.ListenerPort{
				{
					Name:          "EnvoyHTTPPort",
					Protocol:      ir.TCPProtocolType,
					ContainerPort: envoyHTTPPort,
				},
				{
					Name:          "EnvoyHTTPSPort",
					Protocol:      ir.TCPProtocolType,
					ContainerPort: envoyHTTPSPort,
				},
			},
		},
	}

	return i
}

func TestDeployment(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	cases := []struct {
		caseName     string
		infra        *ir.Infra
		deploy       *egcfgv1a1.KubernetesDeploymentSpec
		proxyLogging map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel
		bootstrap    *string
	}{
		{
			caseName: "default",
			infra:    newTestInfra(),
			deploy:   nil,
		},
		{
			caseName: "custom",
			infra:    newTestInfra(),
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
					Image: pointer.String("envoyproxy/envoy:v1.2.3"),
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
			caseName:  "bootstrap",
			infra:     newTestInfra(),
			deploy:    nil,
			bootstrap: pointer.String(`test bootstrap config`),
		},
		{
			caseName: "extension-env",
			infra:    newTestInfra(),
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
					Image: pointer.String("envoyproxy/envoy:v1.2.3"),
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
			infra:    newTestInfra(),
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
					Image: pointer.String("envoyproxy/envoy:v1.2.3"),
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
			caseName: "volumes",
			infra:    newTestInfra(),
			deploy: &egcfgv1a1.KubernetesDeploymentSpec{
				Replicas: pointer.Int32(2),
				Pod: &egcfgv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: pointer.Int64(1000),
					},
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "custom-envoy-cert",
								},
							},
						},
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
					Image: pointer.String("envoyproxy/envoy:v1.2.3"),
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
			caseName: "component-level",
			infra:    newTestInfra(),
			deploy:   nil,
			proxyLogging: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
				egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelError,
				egcfgv1a1.LogComponentFilter:  egcfgv1a1.LogLevelInfo,
			},
			bootstrap: pointer.String(`test bootstrap config`),
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			kube := tc.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider().GetEnvoyProxyKubeProvider()
			if tc.deploy != nil {
				kube.EnvoyDeployment = tc.deploy
			}

			if tc.bootstrap != nil && *tc.bootstrap != "" {
				tc.infra.Proxy.Config.Spec.Bootstrap = tc.bootstrap
			}

			if len(tc.proxyLogging) > 0 {
				tc.infra.Proxy.Config.Spec.Logging = egcfgv1a1.ProxyLogging{
					Level: tc.proxyLogging,
				}
			}

			r := NewResourceRender(cfg.Namespace, tc.infra.GetProxyInfra())
			dp, err := r.Deployment()
			require.NoError(t, err)

			expected, err := loadDeployment(tc.caseName)
			require.NoError(t, err)

			sortEnv := func(env []corev1.EnvVar) {
				sort.Slice(env, func(i, j int) bool {
					return env[i].Name > env[j].Name
				})
			}

			sortEnv(dp.Spec.Template.Spec.Containers[0].Env)
			sortEnv(expected.Spec.Template.Spec.Containers[0].Env)
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

func TestService(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	svcType := egcfgv1a1.ServiceTypeClusterIP
	cases := []struct {
		caseName string
		infra    *ir.Infra
		service  *egcfgv1a1.KubernetesServiceSpec
	}{
		{
			caseName: "default",
			infra:    newTestInfra(),
			service:  nil,
		},
		{
			caseName: "custom",
			infra:    newTestInfra(),
			service: &egcfgv1a1.KubernetesServiceSpec{
				Annotations: map[string]string{
					"key1": "value1",
				},
				Type: &svcType,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			provider := tc.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider().GetEnvoyProxyKubeProvider()
			if tc.service != nil {
				provider.EnvoyService = tc.service
			}

			r := NewResourceRender(cfg.Namespace, tc.infra.GetProxyInfra())
			svc, err := r.Service()
			require.NoError(t, err)

			expected, err := loadService(tc.caseName)
			require.NoError(t, err)

			assert.Equal(t, expected, svc)
		})
	}
}

func loadService(caseName string) (*corev1.Service, error) {
	serviceYAML, err := os.ReadFile(fmt.Sprintf("testdata/services/%s.yaml", caseName))
	if err != nil {
		return nil, err
	}
	svc := &corev1.Service{}
	_ = yaml.Unmarshal(serviceYAML, svc)
	return svc, nil
}

func TestConfigMap(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	infra := newTestInfra()

	r := NewResourceRender(cfg.Namespace, infra.GetProxyInfra())
	cm, err := r.ConfigMap()
	require.NoError(t, err)

	expected, err := loadConfigmap()
	require.NoError(t, err)

	assert.Equal(t, expected, cm)
}

func loadConfigmap() (*corev1.ConfigMap, error) {
	cmYAML, err := os.ReadFile("testdata/configmap.yaml")
	if err != nil {
		return nil, err
	}
	cm := &corev1.ConfigMap{}
	_ = yaml.Unmarshal(cmYAML, cm)
	return cm, nil
}

func TestServiceAccount(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	infra := newTestInfra()

	r := NewResourceRender(cfg.Namespace, infra.GetProxyInfra())
	sa, err := r.ServiceAccount()
	require.NoError(t, err)

	expected, err := loadServiceAccount()
	require.NoError(t, err)

	assert.Equal(t, expected, sa)
}

func loadServiceAccount() (*corev1.ServiceAccount, error) {
	saYAML, err := os.ReadFile("testdata/serviceaccount.yaml")
	if err != nil {
		return nil, err
	}
	sa := &corev1.ServiceAccount{}
	_ = yaml.Unmarshal(saYAML, sa)
	return sa, nil
}
