// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
)

var (
	overrideTestData = flag.Bool("override-testdata", false, "if override the test output data.")
)

const (
	// envoyHTTPPort is the container port number of Envoy's HTTP endpoint.
	envoyHTTPPort = int32(8080)
	// envoyHTTPSPort is the container port number of Envoy's HTTPS endpoint.
	envoyHTTPSPort = int32(8443)
)

func newTestInfra() *ir.Infra {
	return newTestInfraWithAnnotations(nil)
}

func newTestInfraWithAnnotations(annotations map[string]string) *ir.Infra {
	return newTestInfraWithAnnotationsAndLabels(annotations, nil)
}

func newTestInfraWithAddresses(addresses []string) *ir.Infra {
	infra := newTestInfraWithAnnotationsAndLabels(nil, nil)
	infra.Proxy.Addresses = addresses

	return infra
}

func newTestInfraWithAnnotationsAndLabels(annotations, labels map[string]string) *ir.Infra {
	i := ir.NewInfra()

	i.Proxy.GetProxyMetadata().Annotations = annotations
	if len(labels) > 0 {
		i.Proxy.GetProxyMetadata().Labels = labels
	}
	i.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	i.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = i.Proxy.Name
	i.Proxy.Listeners = []*ir.ProxyListener{
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
		deploy       *egv1a1.KubernetesDeploymentSpec
		shutdown     *egv1a1.ShutdownConfig
		proxyLogging map[egv1a1.ProxyLogComponent]egv1a1.LogLevel
		bootstrap    string
		telemetry    *egv1a1.ProxyTelemetry
		concurrency  *int32
		extraArgs    []string
	}{
		{
			caseName: "default",
			infra:    newTestInfra(),
			deploy:   nil,
		},
		{
			caseName: "custom",
			infra:    newTestInfra(),
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](2),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					Labels: map[string]string{
						"foo.bar": "custom-label",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
					Image: ptr.To("envoyproxy/envoy:v1.2.3"),
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
			caseName: "patch-deployment",
			infra:    newTestInfra(),
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
			caseName: "shutdown-manager",
			infra:    newTestInfra(),
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.StrategicMerge),
					Value: v1.JSON{
						Raw: []byte(`{
							"spec":{
								"template":{
									"spec":{
										"containers":[{
											"name":"shutdown-manager",
											"resources":{
												"requests":{"cpu":"100m","memory":"64Mi"},
												"limits":{"cpu":"200m","memory":"96Mi"}
											},
											"securityContext":{"runAsUser":1234},
											"env":[
												{"name":"env_a","value":"env_a_value"},
												{"name":"env_b","value":"env_b_value"}
											],
											"image":"envoyproxy/gateway-dev:v1.2.3"
										}]
									}
								}
							}
						}`),
					},
				},
			},
			shutdown: &egv1a1.ShutdownConfig{
				DrainTimeout: &metav1.Duration{
					Duration: 30 * time.Second,
				},
				MinDrainDuration: &metav1.Duration{
					Duration: 15 * time.Second,
				},
			},
		},
		{
			caseName:  "bootstrap",
			infra:     newTestInfra(),
			deploy:    nil,
			bootstrap: `test bootstrap config`,
		},
		{
			caseName: "extension-env",
			infra:    newTestInfra(),
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
					Image: ptr.To("envoyproxy/envoy:v1.2.3"),
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
			caseName: "default-env",
			infra:    newTestInfra(),
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
					Image: ptr.To("envoyproxy/envoy:v1.2.3"),
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
			infra:    newTestInfra(),
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
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "custom-envoy-cert",
									DefaultMode: ptr.To[int32](420),
								},
							},
						},
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
					Image: ptr.To("envoyproxy/envoy:v1.2.3"),
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
			caseName: "component-level",
			infra:    newTestInfra(),
			deploy:   nil,
			proxyLogging: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
				egv1a1.LogComponentDefault: egv1a1.LogLevelError,
				egv1a1.LogComponentFilter:  egv1a1.LogLevelInfo,
			},
			bootstrap: `test bootstrap config`,
		},
		{
			caseName: "disable-prometheus",
			infra:    newTestInfra(),
			telemetry: &egv1a1.ProxyTelemetry{
				Metrics: &egv1a1.ProxyMetrics{
					Prometheus: &egv1a1.ProxyPrometheusProvider{
						Disable: true,
					},
				},
			},
		},
		{
			caseName:    "with-concurrency",
			infra:       newTestInfra(),
			deploy:      nil,
			concurrency: ptr.To[int32](4),
			bootstrap:   `test bootstrap config`,
		},
		{
			caseName: "custom_with_initcontainers",
			infra:    newTestInfra(),
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Replicas: ptr.To[int32](3),
				Strategy: egv1a1.DefaultKubernetesDeploymentStrategy(),
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"prometheus.io/scrape": "true",
					},
					Labels: map[string]string{
						"foo.bar": "custom-label",
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser: ptr.To[int64](1000),
					},
					Volumes: []corev1.Volume{
						{
							Name: "custom-libs",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
				Container: &egv1a1.KubernetesContainerSpec{
					Image: ptr.To("envoyproxy/envoy:v1.2.3"),
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
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "custom-libs",
							MountPath: "/lib/filter_foo.so",
						},
					},
				},
				InitContainers: []corev1.Container{
					{
						Name:    "install-filter-foo",
						Image:   "alpine:3.11.3",
						Command: []string{"/bin/sh", "-c"},
						Args:    []string{"echo \"Installing filter-foo\"; wget -q https://example.com/download/filter_foo_v1.0.0.tgz -O - | tar -xz --directory=/lib filter_foo.so; echo \"Done\";"},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "custom-libs",
								MountPath: "/lib",
							},
						},
					},
				},
			},
		},
		{
			caseName: "with-annotations",
			infra: newTestInfraWithAnnotations(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}),
			deploy: nil,
		},
		{
			caseName: "override-labels-and-annotations",
			infra: newTestInfraWithAnnotationsAndLabels(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}, map[string]string{
				"label1": "value1",
				"label2": "value2",
			}),
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Pod: &egv1a1.KubernetesPodSpec{
					Annotations: map[string]string{
						"anno1": "value1-override",
					},
					Labels: map[string]string{
						"label1": "value1-override",
					},
				},
			},
		},
		{
			caseName: "with-image-pull-secrets",
			infra:    newTestInfra(),
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Pod: &egv1a1.KubernetesPodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "aaa",
						},
						{
							Name: "bbb",
						},
					},
				},
			},
		},
		{
			caseName: "with-node-selector",
			infra:    newTestInfra(),
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
			caseName: "with-topology-spread-constraints",
			infra:    newTestInfra(),
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
		{
			caseName:  "with-extra-args",
			infra:     newTestInfra(),
			extraArgs: []string{"--key1 val1", "--key2 val2"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			kube := tc.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider().GetEnvoyProxyKubeProvider()
			if tc.deploy != nil {
				kube.EnvoyDeployment = tc.deploy
			}

			replace := egv1a1.BootstrapTypeReplace
			if tc.bootstrap != "" {
				tc.infra.Proxy.Config.Spec.Bootstrap = &egv1a1.ProxyBootstrap{
					Type:  &replace,
					Value: tc.bootstrap,
				}
			}

			if tc.telemetry != nil {
				tc.infra.Proxy.Config.Spec.Telemetry = tc.telemetry
			}

			if len(tc.proxyLogging) > 0 {
				tc.infra.Proxy.Config.Spec.Logging = egv1a1.ProxyLogging{
					Level: tc.proxyLogging,
				}
			}

			if tc.concurrency != nil {
				tc.infra.Proxy.Config.Spec.Concurrency = tc.concurrency
			}

			if tc.shutdown != nil {
				tc.infra.Proxy.Config.Spec.Shutdown = tc.shutdown
			}

			if len(tc.extraArgs) > 0 {
				tc.infra.Proxy.Config.Spec.ExtraArgs = tc.extraArgs
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

			if *overrideTestData {
				deploymentYAML, err := yaml.Marshal(dp)
				require.NoError(t, err)
				// nolint: gosec
				err = os.WriteFile(fmt.Sprintf("testdata/deployments/%s.yaml", tc.caseName), deploymentYAML, 0644)
				require.NoError(t, err)
				return
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

	svcType := egv1a1.ServiceTypeClusterIP
	cases := []struct {
		caseName string
		infra    *ir.Infra
		service  *egv1a1.KubernetesServiceSpec
	}{
		{
			caseName: "default",
			infra:    newTestInfra(),
			service:  nil,
		},
		{
			caseName: "custom",
			infra:    newTestInfra(),
			service: &egv1a1.KubernetesServiceSpec{
				Annotations: map[string]string{
					"key1": "value1",
				},
				Type: &svcType,
			},
		},
		{
			caseName: "with-annotations",
			infra: newTestInfraWithAnnotations(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}),
		},
		{
			caseName: "override-annotations",
			infra: newTestInfraWithAnnotationsAndLabels(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}, map[string]string{
				"label1": "value1",
				"label2": "value2",
			}),
			service: &egv1a1.KubernetesServiceSpec{
				Annotations: map[string]string{
					"anno1": "value1-override",
				},
			},
		},
		{
			caseName: "clusterIP-custom-addresses",
			infra: newTestInfraWithAddresses([]string{
				"10.102.168.100",
			}),
			service: &egv1a1.KubernetesServiceSpec{
				Type: &svcType,
			},
		},
		{
			caseName: "patch-service",
			infra:    newTestInfra(),
			service: &egv1a1.KubernetesServiceSpec{
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.StrategicMerge),
					Value: v1.JSON{
						Raw: []byte("{\"metadata\":{\"name\":\"foo\"}}"),
					},
				},
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
	cases := []struct {
		name  string
		infra *ir.Infra
	}{
		{
			name:  "default",
			infra: newTestInfra(),
		}, {
			name: "with-annotations",
			infra: newTestInfraWithAnnotations(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewResourceRender(cfg.Namespace, tc.infra.GetProxyInfra())
			cm, err := r.ConfigMap()
			require.NoError(t, err)

			expected, err := loadConfigmap(tc.name)
			require.NoError(t, err)

			assert.Equal(t, expected, cm)
		})
	}
}

func loadConfigmap(tc string) (*corev1.ConfigMap, error) {
	cmYAML, err := os.ReadFile(fmt.Sprintf("testdata/configmap/%s.yaml", tc))
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
	cases := []struct {
		name  string
		infra *ir.Infra
	}{
		{
			name:  "default",
			infra: newTestInfra(),
		}, {
			name: "with-annotations",
			infra: newTestInfraWithAnnotations(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewResourceRender(cfg.Namespace, tc.infra.GetProxyInfra())
			sa, err := r.ServiceAccount()
			require.NoError(t, err)

			expected, err := loadServiceAccount(tc.name)
			require.NoError(t, err)

			assert.Equal(t, expected, sa)
		})
	}
}

func loadServiceAccount(tc string) (*corev1.ServiceAccount, error) {
	saYAML, err := os.ReadFile(fmt.Sprintf("testdata/serviceaccount/%s.yaml", tc))
	if err != nil {
		return nil, err
	}
	sa := &corev1.ServiceAccount{}
	_ = yaml.Unmarshal(saYAML, sa)
	return sa, nil
}

func TestHorizontalPodAutoscaler(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	cases := []struct {
		caseName string
		infra    *ir.Infra
		hpa      *egv1a1.KubernetesHorizontalPodAutoscalerSpec
	}{
		{
			caseName: "default",
			infra:    newTestInfra(),
			hpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
				MaxReplicas: ptr.To[int32](1),
			},
		},
		{
			caseName: "custom",
			infra:    newTestInfra(),
			hpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
				MinReplicas: ptr.To[int32](5),
				MaxReplicas: ptr.To[int32](10),
				Metrics: []autoscalingv2.MetricSpec{
					{
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: corev1.ResourceCPU,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: ptr.To[int32](60),
							},
						},
						Type: autoscalingv2.ResourceMetricSourceType,
					},
					{
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: corev1.ResourceMemory,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: ptr.To[int32](70),
							},
						},
						Type: autoscalingv2.ResourceMetricSourceType,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			provider := tc.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider()
			provider.Kubernetes = egv1a1.DefaultEnvoyProxyKubeProvider()

			if tc.hpa != nil {
				provider.Kubernetes.EnvoyHpa = tc.hpa
			}

			provider.GetEnvoyProxyKubeProvider()

			r := NewResourceRender(cfg.Namespace, tc.infra.GetProxyInfra())
			hpa, err := r.HorizontalPodAutoscaler()
			require.NoError(t, err)

			want, err := loadHPA(tc.caseName)
			require.NoError(t, err)

			assert.Equal(t, want, hpa)
		})
	}
}

func loadHPA(caseName string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	hpaYAML, err := os.ReadFile(fmt.Sprintf("testdata/hpa/%s.yaml", caseName))
	if err != nil {
		return nil, err
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{}
	_ = yaml.Unmarshal(hpaYAML, hpa)
	return hpa, nil
}

func TestOwningGatewayLabelsAbsent(t *testing.T) {

	cases := []struct {
		caseName string
		labels   map[string]string
		expect   bool
	}{
		{
			caseName: "OwningGatewayClassLabel exist, but lack OwningGatewayNameLabel or OwningGatewayNamespaceLabel",
			labels: map[string]string{
				"gateway.envoyproxy.io/owning-gatewayclass": "eg-class",
			},
			expect: false,
		},
		{
			caseName: "OwningGatewayNameLabel and OwningGatewayNamespaceLabel exist, but lack OwningGatewayClassLabel",
			labels: map[string]string{
				"gateway.envoyproxy.io/owning-gateway-name":      "eg",
				"gateway.envoyproxy.io/owning-gateway-namespace": "default",
			},
			expect: false,
		},
		{
			caseName: "OwningGatewayNameLabel exist, but lack OwningGatewayClassLabel and OwningGatewayNamespaceLabel",
			labels: map[string]string{
				"gateway.envoyproxy.io/owning-gateway-name": "eg",
			},
			expect: true,
		},
		{
			caseName: "OwningGatewayNamespaceLabel exist, but lack OwningGatewayClassLabel and OwningGatewayNameLabel",
			labels: map[string]string{
				"gateway.envoyproxy.io/owning-gateway-namespace": "default",
			},
			expect: true,
		},
		{
			caseName: "lack all labels",
			labels:   map[string]string{},
			expect:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			actual := OwningGatewayLabelsAbsent(tc.labels)
			require.Equal(t, tc.expect, actual)
		})
	}

}
