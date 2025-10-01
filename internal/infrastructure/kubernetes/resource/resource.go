// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// GetSelector returns a label selector used to select resources
// based on the provided labels.
func GetSelector(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: labels,
	}
}

// ExpectedServiceSpec returns service spec.
func ExpectedServiceSpec(service *egv1a1.KubernetesServiceSpec) corev1.ServiceSpec {
	serviceSpec := corev1.ServiceSpec{}
	serviceSpec.Type = corev1.ServiceType(*service.Type)
	serviceSpec.SessionAffinity = corev1.ServiceAffinityNone
	if service.ExternalTrafficPolicy == nil {
		service.ExternalTrafficPolicy = egv1a1.DefaultKubernetesServiceExternalTrafficPolicy()
	}
	if *service.Type == egv1a1.ServiceTypeLoadBalancer {
		if service.LoadBalancerClass != nil {
			serviceSpec.LoadBalancerClass = service.LoadBalancerClass
		}
		if service.AllocateLoadBalancerNodePorts != nil {
			serviceSpec.AllocateLoadBalancerNodePorts = service.AllocateLoadBalancerNodePorts
		}
		if len(service.LoadBalancerSourceRanges) > 0 {
			serviceSpec.LoadBalancerSourceRanges = service.LoadBalancerSourceRanges
		}
		if service.LoadBalancerIP != nil {
			serviceSpec.LoadBalancerIP = *service.LoadBalancerIP
		}
		serviceSpec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicy(*service.ExternalTrafficPolicy)
	}

	return serviceSpec
}

// ExpectedContainerEnv returns expected container envs.
func ExpectedContainerEnv(container *egv1a1.KubernetesContainerSpec, env []corev1.EnvVar) []corev1.EnvVar {
	amendFunc := func(envVar corev1.EnvVar) {
		for index, e := range env {
			if e.Name == envVar.Name {
				env[index] = envVar
				return
			}
		}
		env = append(env, envVar)
	}

	for _, envVar := range container.Env {
		amendFunc(envVar)
	}

	return env
}

// ExpectedVolumes returns expected deployment volumes.
func ExpectedVolumes(pod *egv1a1.KubernetesPodSpec, volumes []corev1.Volume) []corev1.Volume {
	amendFunc := func(volume corev1.Volume) {
		for index := range volumes {
			if volumes[index].Name == volume.Name {
				volumes[index] = volume
				return
			}
		}

		volumes = append(volumes, volume)
	}

	for i := range pod.Volumes {
		amendFunc(pod.Volumes[i])
	}

	return volumes
}

// ExpectedContainerVolumeMounts returns expected container volume mounts.
func ExpectedContainerVolumeMounts(container *egv1a1.KubernetesContainerSpec, volumeMounts []corev1.VolumeMount) []corev1.VolumeMount {
	amendFunc := func(volumeMount corev1.VolumeMount) {
		for index := range volumeMounts {
			if volumeMounts[index].Name == volumeMount.Name {
				volumeMounts[index] = volumeMount
				return
			}
		}

		volumeMounts = append(volumeMounts, volumeMount)
	}

	for _, volumeMount := range container.VolumeMounts {
		amendFunc(volumeMount)
	}

	return volumeMounts
}

// DefaultSecurityContext returns a default security context with minimal privileges.
func DefaultSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: ptr.To(false),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"ALL",
			},
		},
		Privileged:             ptr.To(false),
		ReadOnlyRootFilesystem: ptr.To(true),
		RunAsNonRoot:           ptr.To(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: "RuntimeDefault",
		},
	}
}
