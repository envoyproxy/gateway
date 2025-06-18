// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"context"
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
	policyv1 "k8s.io/api/policy/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	gwapiresource "github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/test"
)

const (
	// envoyHTTPPort is the container port number of Envoy's HTTP endpoint.
	envoyHTTPPort = int32(8080)
	// envoyHTTPSPort is the container port number of Envoy's HTTPS endpoint.
	envoyHTTPSPort = int32(8443)
	// gatewayClassName is the gateway class name used in tests.
	gatewayClassName = "envoy-gateway-class"
)

type fakeKubernetesInfraProvider struct {
	ControllerNamespace string
	DNSDomain           string
	EnvoyGateway        *egv1a1.EnvoyGateway
}

func newFakeKubernetesInfraProvider(cfg *config.Server) KubernetesInfraProvider {
	return &fakeKubernetesInfraProvider{
		ControllerNamespace: cfg.ControllerNamespace,
		DNSDomain:           cfg.DNSDomain,
		EnvoyGateway:        cfg.EnvoyGateway,
	}
}

func (f *fakeKubernetesInfraProvider) GetControllerNamespace() string {
	return f.ControllerNamespace
}

func (f *fakeKubernetesInfraProvider) GetDNSDomain() string {
	return f.DNSDomain
}

func (f *fakeKubernetesInfraProvider) GetEnvoyGateway() *egv1a1.EnvoyGateway {
	return f.EnvoyGateway
}

func (f *fakeKubernetesInfraProvider) GetOwnerReferenceUID(ctx context.Context, infra *ir.Infra) (map[string]types.UID, error) {
	if f.EnvoyGateway.GatewayNamespaceMode() {
		return map[string]types.UID{
			gwapiresource.KindGateway: "test-owner-reference-uid-for-gateway",
		}, nil
	}
	return map[string]types.UID{
		gwapiresource.KindGatewayClass: "test-owner-reference-uid-for-gatewayclass",
	}, nil
}

func (f *fakeKubernetesInfraProvider) GetResourceNamespace(infra *ir.Infra) string {
	if f.EnvoyGateway.GatewayNamespaceMode() {
		return infra.Proxy.Namespace
	}
	return f.ControllerNamespace
}

func newTestInfra() *ir.Infra {
	return newTestInfraWithAnnotations(nil)
}

func newTestInfraWithNamespacedName(gwNN types.NamespacedName) *ir.Infra {
	i := newTestInfraWithAnnotations(nil)
	i.Proxy.Name = gwNN.Name
	i.Proxy.Namespace = gwNN.Namespace
	i.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = gwNN.Namespace
	i.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = gwNN.Name
	i.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
		Kind: gwapiresource.KindGateway,
		Name: gwNN.Name,
	}

	return i
}

func newTestInfraWithIPFamily(family *egv1a1.IPFamily) *ir.Infra {
	i := newTestInfra()
	i.Proxy.Config = &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			IPFamily: family,
		},
	}
	return i
}

func newTestIPv6Infra() *ir.Infra {
	i := newTestInfra()
	i.Proxy.Config = &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			IPFamily: ptr.To(egv1a1.IPv6),
		},
	}
	return i
}

func newTestDualStackInfra() *ir.Infra {
	i := newTestInfra()
	i.Proxy.Config = &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			IPFamily: ptr.To(egv1a1.DualStack),
		},
	}
	return i
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
	i.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
		Kind: gwapiresource.KindGatewayClass,
		Name: gatewayClassName,
	}
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
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)

	cases := []struct {
		caseName             string
		infra                *ir.Infra
		deploy               *egv1a1.KubernetesDeploymentSpec
		shutdown             *egv1a1.ShutdownConfig
		shutdownManager      *egv1a1.ShutdownManager
		proxyLogging         map[egv1a1.ProxyLogComponent]egv1a1.LogLevel
		bootstrap            string
		telemetry            *egv1a1.ProxyTelemetry
		concurrency          *int32
		extraArgs            []string
		gatewayNamespaceMode bool
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
					Value: apiextensionsv1.JSON{
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
					Value: apiextensionsv1.JSON{
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
											]
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
			shutdownManager: &egv1a1.ShutdownManager{
				Image: ptr.To("privaterepo/envoyproxy/gateway-dev:v1.2.3"),
			},
		},
		{
			caseName:  "bootstrap",
			infra:     newTestInfra(),
			deploy:    nil,
			bootstrap: `test bootstrap config`,
		},
		{
			caseName: "ipv6",
			infra:    newTestIPv6Infra(),
			deploy:   nil,
		},
		{
			caseName: "dual-stack",
			infra:    newTestDualStackInfra(),
			deploy:   nil,
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
				egv1a1.LogComponentDefault: egv1a1.LogLevelTrace,
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
		{
			caseName: "with-empty-memory-limits",
			infra:    newTestInfra(),
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Container: &egv1a1.KubernetesContainerSpec{
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("400m"),
						},
					},
				},
			},
		},
		{
			caseName: "with-name",
			infra:    newTestInfra(),
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Name: ptr.To("custom-deployment-name"),
			},
		},
		{
			caseName:             "gateway-namespace-mode",
			infra:                newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "ns1", Name: "gateway-1"}),
			deploy:               nil,
			gatewayNamespaceMode: true,
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
				bsValue := tc.bootstrap
				tc.infra.Proxy.Config.Spec.Bootstrap = &egv1a1.ProxyBootstrap{
					Type:  &replace,
					Value: &bsValue,
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

			if tc.shutdownManager != nil {
				cfg.EnvoyGateway.Provider.Kubernetes.ShutdownManager = tc.shutdownManager
			} else {
				cfg.EnvoyGateway.Provider.Kubernetes.ShutdownManager = nil
			}

			if len(tc.extraArgs) > 0 {
				tc.infra.Proxy.Config.Spec.ExtraArgs = tc.extraArgs
			}
			if tc.gatewayNamespaceMode {
				cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
						Deploy: &egv1a1.KubernetesDeployMode{
							Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
						},
					},
				}
			}

			r, err := NewResourceRender(context.Background(), newFakeKubernetesInfraProvider(cfg), tc.infra)
			require.NoError(t, err)
			dp, err := r.Deployment()
			require.NoError(t, err)

			if test.OverrideTestData() {
				deploymentYAML, err := yaml.Marshal(dp)
				require.NoError(t, err)
				// nolint: gosec
				err = os.WriteFile(fmt.Sprintf("testdata/deployments/%s.yaml", tc.caseName), deploymentYAML, 0o644)
				require.NoError(t, err)
				return
			}

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

func TestDaemonSet(t *testing.T) {
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)

	cases := []struct {
		caseName             string
		infra                *ir.Infra
		daemonset            *egv1a1.KubernetesDaemonSetSpec
		shutdown             *egv1a1.ShutdownConfig
		proxyLogging         map[egv1a1.ProxyLogComponent]egv1a1.LogLevel
		bootstrap            string
		telemetry            *egv1a1.ProxyTelemetry
		concurrency          *int32
		extraArgs            []string
		gatewayNamespaceMode bool
	}{
		{
			caseName:  "default",
			infra:     newTestInfra(),
			daemonset: nil,
		},
		{
			caseName: "custom",
			infra:    newTestInfra(),
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
				Strategy: egv1a1.DefaultKubernetesDaemonSetStrategy(),
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
			caseName: "patch-daemonset",
			infra:    newTestInfra(),
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.StrategicMerge),
					Value: apiextensionsv1.JSON{
						Raw: []byte("{\"spec\":{\"template\":{\"spec\":{\"hostNetwork\":true,\"dnsPolicy\":\"ClusterFirstWithHostNet\"}}}}"),
					},
				},
			},
		},
		{
			caseName: "shutdown-manager",
			infra:    newTestInfra(),
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.StrategicMerge),
					Value: apiextensionsv1.JSON{
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
			caseName: "extension-env",
			infra:    newTestInfra(),
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
				Strategy: egv1a1.DefaultKubernetesDaemonSetStrategy(),
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
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
				Strategy: egv1a1.DefaultKubernetesDaemonSetStrategy(),
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
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
				Strategy: egv1a1.DefaultKubernetesDaemonSetStrategy(),
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
			caseName:  "component-level",
			infra:     newTestInfra(),
			daemonset: nil,
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
			daemonset:   nil,
			concurrency: ptr.To[int32](4),
			bootstrap:   `test bootstrap config`,
		},
		{
			caseName: "with-annotations",
			infra: newTestInfraWithAnnotations(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}),
			daemonset: nil,
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
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
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
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
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
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
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
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
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
		{
			caseName: "with-name",
			infra:    newTestInfra(),
			daemonset: &egv1a1.KubernetesDaemonSetSpec{
				Name: ptr.To("custom-daemonset-name"),
			},
		},
		{
			caseName:             "gateway-namespace-mode",
			infra:                newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "ns1", Name: "gateway-1"}),
			daemonset:            nil,
			gatewayNamespaceMode: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.gatewayNamespaceMode {
				cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
						Deploy: &egv1a1.KubernetesDeployMode{
							Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
						},
					},
				}
			}

			kube := tc.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider().GetEnvoyProxyKubeProvider()

			// fill deploument, use daemonset
			kube.EnvoyDeployment = nil
			kube.EnvoyDaemonSet = egv1a1.DefaultKubernetesDaemonSet(egv1a1.DefaultEnvoyProxyImage)

			if tc.daemonset != nil {
				kube.EnvoyDaemonSet = tc.daemonset
			}

			replace := egv1a1.BootstrapTypeReplace
			if tc.bootstrap != "" {
				bsValue := tc.bootstrap
				tc.infra.Proxy.Config.Spec.Bootstrap = &egv1a1.ProxyBootstrap{
					Type:  &replace,
					Value: &bsValue,
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

			r, err := NewResourceRender(context.Background(), newFakeKubernetesInfraProvider(cfg), tc.infra)
			require.NoError(t, err)
			ds, err := r.DaemonSet()
			require.NoError(t, err)

			expected, err := loadDaemonSet(tc.caseName)
			require.NoError(t, err)

			sortEnv := func(env []corev1.EnvVar) {
				sort.Slice(env, func(i, j int) bool {
					return env[i].Name > env[j].Name
				})
			}

			if test.OverrideTestData() {
				deploymentYAML, err := yaml.Marshal(ds)
				require.NoError(t, err)
				// nolint: gosec
				err = os.WriteFile(fmt.Sprintf("testdata/daemonsets/%s.yaml", tc.caseName), deploymentYAML, 0o644)
				require.NoError(t, err)
				return
			}

			sortEnv(ds.Spec.Template.Spec.Containers[0].Env)
			sortEnv(expected.Spec.Template.Spec.Containers[0].Env)
			assert.Equal(t, expected, ds)
		})
	}
}

func loadDaemonSet(caseName string) (*appsv1.DaemonSet, error) {
	daemonsetYAML, err := os.ReadFile(fmt.Sprintf("testdata/daemonsets/%s.yaml", caseName))
	if err != nil {
		return nil, err
	}
	daemonset := &appsv1.DaemonSet{}
	_ = yaml.Unmarshal(daemonsetYAML, daemonset)
	return daemonset, nil
}

func TestService(t *testing.T) {
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)

	svcType := egv1a1.ServiceTypeClusterIP
	cases := []struct {
		caseName             string
		infra                *ir.Infra
		service              *egv1a1.KubernetesServiceSpec
		gatewayNamespaceMode bool
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
				Labels: map[string]string{
					"key1": "value1",
				},
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
			caseName: "with-svc-labels",
			infra:    newTestInfra(),
			service: &egv1a1.KubernetesServiceSpec{
				Labels: map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
			},
		},
		{
			caseName: "override-labels",
			infra: newTestInfraWithAnnotationsAndLabels(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}, map[string]string{
				"label1": "value1",
				"label2": "value2",
			}),
			service: &egv1a1.KubernetesServiceSpec{
				Labels: map[string]string{
					"label1": "value1-override",
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
					Value: apiextensionsv1.JSON{
						Raw: []byte("{\"metadata\":{\"name\":\"foo\"}}"),
					},
				},
			},
		},
		{
			caseName: "with-name",
			infra:    newTestInfra(),
			service: &egv1a1.KubernetesServiceSpec{
				Name: ptr.To("custom-service-name"),
			},
		},
		{
			caseName: "dualstack",
			infra:    newTestInfraWithIPFamily(ptr.To(egv1a1.DualStack)),
			service:  nil,
		},
		{
			caseName: "ipv4-singlestack",
			infra:    newTestInfraWithIPFamily(ptr.To(egv1a1.IPv4)),
			service:  nil,
		},
		{
			caseName: "ipv6-singlestack",
			infra:    newTestInfraWithIPFamily(ptr.To(egv1a1.IPv6)),
			service:  nil,
		},
		{
			caseName:             "gateway-namespace-mode",
			infra:                newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "ns1", Name: "gateway-1"}),
			service:              nil,
			gatewayNamespaceMode: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.gatewayNamespaceMode {
				cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
						Deploy: &egv1a1.KubernetesDeployMode{
							Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
						},
					},
				}
			}

			provider := tc.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider().GetEnvoyProxyKubeProvider()
			if tc.service != nil {
				provider.EnvoyService = tc.service
			}

			r, err := NewResourceRender(context.Background(), newFakeKubernetesInfraProvider(cfg), tc.infra)
			require.NoError(t, err)
			svc, err := r.Service()
			require.NoError(t, err)

			if test.OverrideTestData() {
				data, err := yaml.Marshal(svc)
				require.NoError(t, err)
				err = os.WriteFile(fmt.Sprintf("testdata/services/%s.yaml", tc.caseName), data, 0o600)
				require.NoError(t, err)
				return
			}

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
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	cases := []struct {
		name                 string
		infra                *ir.Infra
		gatewayNamespaceMode bool
	}{
		{
			name:  "default",
			infra: newTestInfra(),
		},
		{
			name: "with-annotations",
			infra: newTestInfraWithAnnotations(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}),
		},
		{
			name:                 "gateway-namespace-mode",
			infra:                newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "ns1", Name: "gateway-1"}),
			gatewayNamespaceMode: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.gatewayNamespaceMode {
				cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
						Deploy: &egv1a1.KubernetesDeployMode{
							Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
						},
					},
				}
			}
			r, err := NewResourceRender(context.Background(), newFakeKubernetesInfraProvider(cfg), tc.infra)
			require.NoError(t, err)
			cm, err := r.ConfigMap("")
			require.NoError(t, err)

			if test.OverrideTestData() {
				data, err := yaml.Marshal(cm)
				require.NoError(t, err)
				err = os.WriteFile(fmt.Sprintf("testdata/configmap/%s.yaml", tc.name), data, 0o600)
				require.NoError(t, err)
				return
			}

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
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)
	cases := []struct {
		name                 string
		infra                *ir.Infra
		gatewayNamespaceMode bool
	}{
		{
			name:  "default",
			infra: newTestInfra(),
		},
		{
			name: "with-annotations",
			infra: newTestInfraWithAnnotations(map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			}),
		},
		{
			name:                 "gateway-namespace-mode",
			infra:                newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "ns1", Name: "gateway-1"}),
			gatewayNamespaceMode: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.gatewayNamespaceMode {
				cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
						Deploy: &egv1a1.KubernetesDeployMode{
							Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
						},
					},
				}
			}
			r, err := NewResourceRender(context.Background(), newFakeKubernetesInfraProvider(cfg), tc.infra)
			require.NoError(t, err)
			sa, err := r.ServiceAccount()
			require.NoError(t, err)

			if test.OverrideTestData() {
				saYAML, err := yaml.Marshal(sa)
				require.NoError(t, err)
				err = os.WriteFile(fmt.Sprintf("testdata/serviceaccount/%s.yaml", tc.name), saYAML, 0o600)
				require.NoError(t, err)
				return
			}

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

func TestPDB(t *testing.T) {
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)

	cases := []struct {
		caseName             string
		infra                *ir.Infra
		pdb                  *egv1a1.KubernetesPodDisruptionBudgetSpec
		deploy               *egv1a1.KubernetesDeploymentSpec
		gatewayNamespaceMode bool
	}{
		{
			caseName: "default",
			infra:    newTestInfra(),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				MinAvailable: ptr.To(intstr.IntOrString{Type: intstr.Int, IntVal: 1}),
			},
		},
		{
			caseName: "patch-json-pdb",
			infra:    newTestInfra(),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				MinAvailable: ptr.To(intstr.IntOrString{Type: intstr.Int, IntVal: 1}),
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.JSONMerge),
					Value: apiextensionsv1.JSON{
						Raw: []byte("{\"metadata\":{\"name\":\"foo\"}, \"spec\": {\"selector\": {\"matchLabels\": {\"app\": \"bar\"}}}}"),
					},
				},
			},
		},
		{
			caseName: "patch-strategic-pdb",
			infra:    newTestInfra(),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				MinAvailable: ptr.To(intstr.IntOrString{Type: intstr.Int, IntVal: 1}),
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.StrategicMerge),
					Value: apiextensionsv1.JSON{
						Raw: []byte("{\"metadata\":{\"name\":\"foo\"}, \"spec\": {\"selector\": {\"matchLabels\": {\"app\": \"bar\"}}}}"),
					},
				},
			},
		},
		{
			caseName: "max-unavailable",
			infra:    newTestInfra(),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				MaxUnavailable: ptr.To(intstr.IntOrString{Type: intstr.Int, IntVal: 1}),
			},
		},
		{
			caseName: "max-unavailable-percent",
			infra:    newTestInfra(),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				MaxUnavailable: ptr.To(intstr.IntOrString{Type: intstr.String, StrVal: "20%"}),
			},
		},
		{
			caseName: "min-available-percent",
			infra:    newTestInfra(),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				MinAvailable: ptr.To(intstr.IntOrString{Type: intstr.String, StrVal: "20%"}),
			},
		},
		{
			caseName: "patch-pdb-no-minmax",
			infra:    newTestInfra(),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.StrategicMerge),
					Value: apiextensionsv1.JSON{
						Raw: []byte("{\"metadata\":{\"name\":\"foo\"}, \"spec\": {\"minAvailable\": 1, \"selector\": {\"matchLabels\": {\"app\": \"bar\"}}}}"),
					},
				},
			},
		},
		{
			caseName: "with-name",
			infra:    newTestInfra(),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				MinAvailable: ptr.To(intstr.IntOrString{Type: intstr.Int, IntVal: 1}),
				Name:         ptr.To("custom-pdb-name"),
			},
		},
		{
			caseName: "gateway-namespace-mode",
			infra:    newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "ns1", Name: "gateway-1"}),
			pdb: &egv1a1.KubernetesPodDisruptionBudgetSpec{
				MinAvailable: ptr.To(intstr.IntOrString{Type: intstr.Int, IntVal: 1}),
			},
			gatewayNamespaceMode: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.gatewayNamespaceMode {
				cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
						Deploy: &egv1a1.KubernetesDeployMode{
							Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
						},
					},
				}
			}

			provider := tc.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider()
			provider.Kubernetes = egv1a1.DefaultEnvoyProxyKubeProvider()

			if tc.deploy != nil {
				provider.Kubernetes.EnvoyDeployment = tc.deploy
			}

			if tc.pdb != nil {
				provider.Kubernetes.EnvoyPDB = tc.pdb
			}

			provider.GetEnvoyProxyKubeProvider()

			r, err := NewResourceRender(context.Background(), newFakeKubernetesInfraProvider(cfg), tc.infra)
			require.NoError(t, err)

			pdb, err := r.PodDisruptionBudget()
			require.NoError(t, err)

			if test.OverrideTestData() {
				data, err := yaml.Marshal(pdb)
				require.NoError(t, err)
				err = os.WriteFile(fmt.Sprintf("testdata/pdb/%s.yaml", tc.caseName), data, 0o600)
				require.NoError(t, err)
				return
			}

			podPDBExpected, err := loadPDB(tc.caseName)
			require.NoError(t, err)
			assert.Equal(t, podPDBExpected, pdb)
		})
	}
}

func TestHorizontalPodAutoscaler(t *testing.T) {
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)

	cases := []struct {
		caseName             string
		infra                *ir.Infra
		hpa                  *egv1a1.KubernetesHorizontalPodAutoscalerSpec
		deploy               *egv1a1.KubernetesDeploymentSpec
		gatewayNamespaceMode bool
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
		{
			caseName: "patch-json-hpa",
			infra:    newTestInfra(),
			hpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
				MaxReplicas: ptr.To[int32](1),
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.JSONMerge),
					Value: apiextensionsv1.JSON{
						Raw: []byte("{\"metadata\":{\"name\":\"foo\"}, \"spec\": {\"scaleTargetRef\": {\"name\": \"bar\"}}}"),
					},
				},
			},
		},
		{
			caseName: "patch-strategic-hpa",
			infra:    newTestInfra(),
			hpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
				MaxReplicas: ptr.To[int32](1),
				Patch: &egv1a1.KubernetesPatchSpec{
					Type: ptr.To(egv1a1.StrategicMerge),
					Value: apiextensionsv1.JSON{
						Raw: []byte("{\"metadata\":{\"name\":\"foo\"}, \"spec\": {\"metrics\": [{\"resource\": {\"name\": \"cpu\", \"target\": {\"averageUtilization\": 50, \"type\": \"Utilization\"}}, \"type\": \"Resource\"}]}}"),
					},
				},
			},
		},
		{
			caseName: "with-deployment-name",
			infra:    newTestInfra(),
			hpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
				MinReplicas: ptr.To[int32](5),
				MaxReplicas: ptr.To[int32](10),
			},
			deploy: &egv1a1.KubernetesDeploymentSpec{
				Name: ptr.To("custom-deployment-name"),
			},
		},
		{
			caseName: "with-name",
			infra:    newTestInfra(),
			hpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
				MaxReplicas: ptr.To[int32](1),
				Name:        ptr.To("custom-hpa-name"),
			},
		},
		{
			caseName: "gateway-namespace-mode",
			infra:    newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "ns1", Name: "gateway-1"}),
			hpa: &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
				MaxReplicas: ptr.To[int32](1),
			},
			gatewayNamespaceMode: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.caseName, func(t *testing.T) {
			if tc.gatewayNamespaceMode {
				cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
					Type: egv1a1.ProviderTypeKubernetes,
					Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
						Deploy: &egv1a1.KubernetesDeployMode{
							Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
						},
					},
				}
			}

			provider := tc.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider()
			provider.Kubernetes = egv1a1.DefaultEnvoyProxyKubeProvider()

			if tc.hpa != nil {
				provider.Kubernetes.EnvoyHpa = tc.hpa
			}
			if tc.deploy != nil {
				provider.Kubernetes.EnvoyDeployment = tc.deploy
			}
			provider.GetEnvoyProxyKubeProvider()

			r, err := NewResourceRender(context.Background(), newFakeKubernetesInfraProvider(cfg), tc.infra)
			require.NoError(t, err)
			hpa, err := r.HorizontalPodAutoscaler()
			require.NoError(t, err)

			if test.OverrideTestData() {
				data, err := yaml.Marshal(hpa)
				require.NoError(t, err)
				err = os.WriteFile(fmt.Sprintf("testdata/hpa/%s.yaml", tc.caseName), data, 0o600)
				require.NoError(t, err)
				return
			}

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

func loadPDB(caseName string) (*policyv1.PodDisruptionBudget, error) {
	pdbYAML, err := os.ReadFile(fmt.Sprintf("testdata/pdb/%s.yaml", caseName))
	if err != nil {
		return nil, err
	}

	pdb := &policyv1.PodDisruptionBudget{}
	_ = yaml.Unmarshal(pdbYAML, pdb)
	return pdb, nil
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

func TestIPFamilyPresentInSpec(t *testing.T) {
	cases := []struct {
		name             string
		requestedFamily  *egv1a1.IPFamily
		expectedFamilies []corev1.IPFamily
		expectedPolicy   *corev1.IPFamilyPolicy
	}{
		{
			"no family specified",
			nil,
			nil,
			nil,
		},
		{
			"ipv4 specified",
			ptr.To(egv1a1.IPv4),
			nil,
			nil,
		},
		{
			"ipv6 specified",
			ptr.To(egv1a1.IPv6),
			[]corev1.IPFamily{corev1.IPv6Protocol},
			ptr.To(corev1.IPFamilyPolicySingleStack),
		},
		{
			"dual stack",
			ptr.To(egv1a1.DualStack),
			[]corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol},
			ptr.To(corev1.IPFamilyPolicyRequireDualStack),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := ResourceRender{infra: newTestInfraWithIPFamily(tc.requestedFamily).Proxy}
			svc, err := r.Service()
			require.NoError(t, err, "service render func")

			assert.ElementsMatch(t, tc.expectedFamilies, svc.Spec.IPFamilies, "families slice")
			assert.Equal(t, tc.expectedPolicy, svc.Spec.IPFamilyPolicy, "policy")
		})
	}
}

func TestGatewayNamespaceModeMultipleResources(t *testing.T) {
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)

	// Configure gateway namespace mode
	cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
		Type: egv1a1.ProviderTypeKubernetes,
		Kubernetes: &egv1a1.EnvoyGatewayKubernetesProvider{
			Deploy: &egv1a1.KubernetesDeployMode{
				Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
			},
		},
	}

	// Create test infra with multiple namespaces
	var infraList []*ir.Infra
	infra1 := newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "namespace-1", Name: "gateway-1"})
	// Add HPA config to first infra
	if infra1.Proxy.Config == nil {
		infra1.Proxy.Config = &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{}}
	}
	if infra1.Proxy.Config.Spec.Provider == nil {
		infra1.Proxy.Config.Spec.Provider = &egv1a1.EnvoyProxyProvider{}
	}
	infra1.Proxy.Config.Spec.Provider.Type = egv1a1.ProviderTypeKubernetes
	if infra1.Proxy.Config.Spec.Provider.Kubernetes == nil {
		infra1.Proxy.Config.Spec.Provider.Kubernetes = &egv1a1.EnvoyProxyKubernetesProvider{}
	}
	infra1.Proxy.Config.Spec.Provider.Kubernetes.EnvoyHpa = &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
		MinReplicas: ptr.To[int32](1),
		MaxReplicas: ptr.To[int32](3),
	}

	infra2 := newTestInfraWithNamespacedName(types.NamespacedName{Namespace: "namespace-2", Name: "gateway-2"})
	// Add HPA config to second infra
	if infra2.Proxy.Config == nil {
		infra2.Proxy.Config = &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{}}
	}
	if infra2.Proxy.Config.Spec.Provider == nil {
		infra2.Proxy.Config.Spec.Provider = &egv1a1.EnvoyProxyProvider{}
	}
	infra2.Proxy.Config.Spec.Provider.Type = egv1a1.ProviderTypeKubernetes
	if infra2.Proxy.Config.Spec.Provider.Kubernetes == nil {
		infra2.Proxy.Config.Spec.Provider.Kubernetes = &egv1a1.EnvoyProxyKubernetesProvider{}
	}
	infra2.Proxy.Config.Spec.Provider.Kubernetes.EnvoyHpa = &egv1a1.KubernetesHorizontalPodAutoscalerSpec{
		MinReplicas: ptr.To[int32](1),
		MaxReplicas: ptr.To[int32](3),
	}

	infraList = append(infraList, infra1, infra2)

	deployments := make([]*appsv1.Deployment, 0, len(infraList))
	services := make([]*corev1.Service, 0, len(infraList))
	serviceAccounts := make([]*corev1.ServiceAccount, 0, len(infraList))
	hpas := make([]*autoscalingv2.HorizontalPodAutoscaler, 0, len(infraList))

	for _, infra := range infraList {
		r, err := NewResourceRender(context.Background(), newFakeKubernetesInfraProvider(cfg), infra)
		require.NoError(t, err)

		dp, err := r.Deployment()
		require.NoError(t, err)
		deployments = append(deployments, dp)

		svc, err := r.Service()
		require.NoError(t, err)
		services = append(services, svc)

		sa, err := r.ServiceAccount()
		require.NoError(t, err)
		serviceAccounts = append(serviceAccounts, sa)

		hpa, err := r.HorizontalPodAutoscaler()
		require.NoError(t, err)
		hpas = append(hpas, hpa)

	}

	// Verify correct number of resources
	require.Len(t, deployments, len(infraList))
	require.Len(t, services, len(infraList))
	require.Len(t, serviceAccounts, len(infraList))
	require.Len(t, hpas, len(infraList))

	if test.OverrideTestData() {
		deploymentInterfaces := make([]any, len(deployments))
		for i, dp := range deployments {
			deploymentInterfaces[i] = dp
		}

		err := writeTestDataToFile("testdata/gateway-namespace-mode/deployment.yaml", deploymentInterfaces)
		require.NoError(t, err)

		serviceInterfaces := make([]any, len(services))
		for i, svc := range services {
			serviceInterfaces[i] = svc
		}
		err = writeTestDataToFile("testdata/gateway-namespace-mode/service.yaml", serviceInterfaces)
		require.NoError(t, err)

		saInterfaces := make([]any, len(serviceAccounts))
		for i, sa := range serviceAccounts {
			saInterfaces[i] = sa
		}
		err = writeTestDataToFile("testdata/gateway-namespace-mode/serviceaccount.yaml", saInterfaces)
		require.NoError(t, err)

		hpaInterfaces := make([]any, len(hpas))
		for i, hpa := range hpas {
			hpaInterfaces[i] = hpa
		}
		err = writeTestDataToFile("testdata/gateway-namespace-mode/hpa.yaml", hpaInterfaces)
		require.NoError(t, err)

		return
	}

	for i, infra := range infraList {
		expectedNamespace := infra.GetProxyInfra().Namespace
		expectedName := infra.GetProxyInfra().Name

		require.Equal(t, expectedNamespace, deployments[i].Namespace)
		require.Equal(t, expectedName, deployments[i].Name)

		require.Equal(t, expectedNamespace, services[i].Namespace)
		require.Equal(t, expectedName, services[i].Name)

		require.Equal(t, expectedNamespace, serviceAccounts[i].Namespace)
		require.Equal(t, expectedName, serviceAccounts[i].Name)

		if i < len(hpas) {
			require.Equal(t, expectedNamespace, hpas[i].Namespace)
			require.Equal(t, expectedName, hpas[i].Name)
			require.Equal(t, expectedName, hpas[i].Spec.ScaleTargetRef.Name)
		}
	}
}

func writeTestDataToFile(filename string, resources []any) error {
	var combinedYAML []byte
	for i, resource := range resources {
		resourceYAML, err := yaml.Marshal(resource)
		if err != nil {
			return err
		}
		if i > 0 {
			combinedYAML = append(combinedYAML, []byte("---\n")...)
		}
		combinedYAML = append(combinedYAML, resourceYAML...)
	}

	return os.WriteFile(filename, combinedYAML, 0o600)
}
