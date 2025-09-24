// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestExpectedServiceSpec(t *testing.T) {
	type args struct {
		service *egv1a1.KubernetesServiceSpec
	}
	loadbalancerClass := "foobar"
	tests := []struct {
		name string
		args args
		want corev1.ServiceSpec
	}{
		{
			name: "LoadBalancer",
			args: args{service: &egv1a1.KubernetesServiceSpec{
				Type: egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
			}},
			want: corev1.ServiceSpec{
				Type:                  corev1.ServiceTypeLoadBalancer,
				SessionAffinity:       corev1.ServiceAffinityNone,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
		},
		{
			name: "LoadBalancerWithExternalTrafficPolicyCluster",
			args: args{service: &egv1a1.KubernetesServiceSpec{
				Type:                  egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
				ExternalTrafficPolicy: egv1a1.GetKubernetesServiceExternalTrafficPolicy(egv1a1.ServiceExternalTrafficPolicyCluster),
			}},
			want: corev1.ServiceSpec{
				Type:                  corev1.ServiceTypeLoadBalancer,
				SessionAffinity:       corev1.ServiceAffinityNone,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeCluster,
			},
		},
		{
			name: "LoadBalancerWithClass",
			args: args{service: &egv1a1.KubernetesServiceSpec{
				Type:              egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
				LoadBalancerClass: &loadbalancerClass,
			}},
			want: corev1.ServiceSpec{
				Type:                  corev1.ServiceTypeLoadBalancer,
				LoadBalancerClass:     &loadbalancerClass,
				SessionAffinity:       corev1.ServiceAffinityNone,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
		},
		{
			name: "LoadBalancerWithAllocateLoadBalancerNodePorts",
			args: args{service: &egv1a1.KubernetesServiceSpec{
				Type:                          egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
				AllocateLoadBalancerNodePorts: ptr.To(true),
			}},
			want: corev1.ServiceSpec{
				Type:                          corev1.ServiceTypeLoadBalancer,
				AllocateLoadBalancerNodePorts: ptr.To(true),
				SessionAffinity:               corev1.ServiceAffinityNone,
				ExternalTrafficPolicy:         corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
		},
		{
			name: "LoadBalancerWithLoadBalancerSourceRanges",
			args: args{service: &egv1a1.KubernetesServiceSpec{
				Type:                     egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
				LoadBalancerSourceRanges: []string{"1.1.1.1/32"},
			}},
			want: corev1.ServiceSpec{
				Type:                     corev1.ServiceTypeLoadBalancer,
				LoadBalancerSourceRanges: []string{"1.1.1.1/32"},
				SessionAffinity:          corev1.ServiceAffinityNone,
				ExternalTrafficPolicy:    corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
		},
		{
			name: "LoadBalancerWithLoadBalancerIP",
			args: args{service: &egv1a1.KubernetesServiceSpec{
				Type:           egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer),
				LoadBalancerIP: ptr.To("10.11.12.13"),
			}},
			want: corev1.ServiceSpec{
				Type:                  corev1.ServiceTypeLoadBalancer,
				LoadBalancerIP:        "10.11.12.13",
				SessionAffinity:       corev1.ServiceAffinityNone,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
		},
		{
			name: "ClusterIP",
			args: args{service: &egv1a1.KubernetesServiceSpec{
				Type: egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeClusterIP),
			}},
			want: corev1.ServiceSpec{
				Type:            corev1.ServiceTypeClusterIP,
				SessionAffinity: corev1.ServiceAffinityNone,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ExpectedServiceSpec(tt.args.service), "expectedServiceSpec(%v)", tt.args.service)
		})
	}
}

func TestGetSelector(t *testing.T) {
	cases := []struct {
		name     string
		in       map[string]string
		expected map[string]string
	}{
		{
			name: "proxy",
			in: map[string]string{
				"app.kubernetes.io/name":       "envoy",
				"app.kubernetes.io/component":  "proxy",
				"app.kubernetes.io/managed-by": "envoy-gateway",
			},
			expected: map[string]string{
				"app.kubernetes.io/name":       "envoy",
				"app.kubernetes.io/component":  "proxy",
				"app.kubernetes.io/managed-by": "envoy-gateway",
			},
		},
		{
			name: "ratelimit",
			in: map[string]string{
				"app.kubernetes.io/name":       "envoy-ratelimit",
				"app.kubernetes.io/component":  "ratelimit",
				"app.kubernetes.io/managed-by": "envoy-gateway",
			},
			expected: map[string]string{
				"app.kubernetes.io/name":       "envoy-ratelimit",
				"app.kubernetes.io/component":  "ratelimit",
				"app.kubernetes.io/managed-by": "envoy-gateway",
			},
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := GetSelector(tc.in)
			require.Equal(t, tc.expected, got.MatchLabels)
		})
	}
}

func TestExpectedProxyContainerEnv(t *testing.T) {
	type args struct {
		container *egv1a1.KubernetesContainerSpec
		env       []corev1.EnvVar
	}
	tests := []struct {
		name string
		args args
		want []corev1.EnvVar
	}{
		{
			name: "TestExpectedProxyContainerEnv",
			args: args{
				container: &egv1a1.KubernetesContainerSpec{
					Env: []corev1.EnvVar{
						{
							Name:  "env_a",
							Value: "override_env_a_value",
						},
						{
							Name:  "env_b",
							Value: "override_env_a_value",
						},
						{
							Name:  "env_c",
							Value: "new_env_c_value",
						},
					},
				},
				env: []corev1.EnvVar{
					{
						Name:  "env_a",
						Value: "default_env_a_value",
					},
					{
						Name:  "env_b",
						Value: "default_env_a_value",
					},
					{
						Name:  "default_env",
						Value: "default_env_value",
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name:  "env_a",
					Value: "override_env_a_value",
				},
				{
					Name:  "env_b",
					Value: "override_env_a_value",
				},
				{
					Name:  "default_env",
					Value: "default_env_value",
				},
				{
					Name:  "env_c",
					Value: "new_env_c_value",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ExpectedContainerEnv(tt.args.container, tt.args.env), "ExpectedProxyContainerEnv(%v, %v)", tt.args.container, tt.args.env)
		})
	}
}

func TestExpectedDeploymentVolumes(t *testing.T) {
	type args struct {
		pod     *egv1a1.KubernetesPodSpec
		volumes []corev1.Volume
	}
	tests := []struct {
		name string
		args args
		want []corev1.Volume
	}{
		{
			name: "TestExpectedDeploymentVolumes",
			args: args{
				pod: &egv1a1.KubernetesPodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "override-cert",
								},
							},
						},
					},
				},
				volumes: []corev1.Volume{
					{
						Name: "certs",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "cert",
							},
						},
					},
					{
						Name: "default-certs",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "default-cert",
							},
						},
					},
				},
			},
			want: []corev1.Volume{
				{
					Name: "certs",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "override-cert",
						},
					},
				},
				{
					Name: "default-certs",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "default-cert",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ExpectedVolumes(tt.args.pod, tt.args.volumes), "ExpectedVolumes(%v, %v)", tt.args.pod, tt.args.volumes)
		})
	}
}

func TestExpectedContainerVolumeMounts(t *testing.T) {
	type args struct {
		container    *egv1a1.KubernetesContainerSpec
		volumeMounts []corev1.VolumeMount
	}
	tests := []struct {
		name string
		args args
		want []corev1.VolumeMount
	}{
		{
			name: "TestExpectedContainerVolumeMounts",
			args: args{
				container: &egv1a1.KubernetesContainerSpec{
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "certs",
							MountPath: "/override-certs",
							ReadOnly:  true,
						},
					},
				},
				volumeMounts: []corev1.VolumeMount{
					{
						Name:      "certs",
						MountPath: "/certs",
						ReadOnly:  true,
					},
					{
						Name:      "default-certs",
						MountPath: "/default-certs",
						ReadOnly:  true,
					},
				},
			},
			want: []corev1.VolumeMount{
				{
					Name:      "certs",
					MountPath: "/override-certs",
					ReadOnly:  true,
				},
				{
					Name:      "default-certs",
					MountPath: "/default-certs",
					ReadOnly:  true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ExpectedContainerVolumeMounts(tt.args.container, tt.args.volumeMounts), "ExpectedContainerVolumeMounts(%v, %v)", tt.args.container, tt.args.volumeMounts)
		})
	}
}
