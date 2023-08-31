// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

// GetSelector returns a label selector used to select resources
// based on the provided labels.
func GetSelector(labels map[string]string) *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: labels,
	}
}

// ExpectedServiceSpec returns service spec.
func ExpectedServiceSpec(service *egcfgv1a1.KubernetesServiceSpec) corev1.ServiceSpec {
	serviceSpec := corev1.ServiceSpec{}
	serviceSpec.Type = corev1.ServiceType(*service.Type)
	serviceSpec.SessionAffinity = corev1.ServiceAffinityNone
	if *service.Type == egcfgv1a1.ServiceTypeLoadBalancer {
		if service.LoadBalancerClass != nil {
			serviceSpec.LoadBalancerClass = service.LoadBalancerClass
		}
		// Preserve the client source IP and avoid a second hop for LoadBalancer.
		serviceSpec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeLocal
	}
	return serviceSpec
}

// CompareSvc compare entire Svc.Spec but ignored the ports[*].nodePort, ClusterIP and ClusterIPs in case user have modified for some scene.
func CompareSvc(currentSvc, originalSvc *corev1.Service) bool {
	return cmp.Equal(currentSvc.Spec, originalSvc.Spec,
		cmpopts.IgnoreFields(corev1.ServicePort{}, "NodePort"),
		cmpopts.IgnoreFields(corev1.ServiceSpec{}, "ClusterIP", "ClusterIPs"))
}

// ExpectedProxyContainerEnv returns expected container envs.
func ExpectedProxyContainerEnv(container *egcfgv1a1.KubernetesContainerSpec, env []corev1.EnvVar) []corev1.EnvVar {
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

// ExpectedDeploymentVolumes returns expected deployment volumes.
func ExpectedDeploymentVolumes(pod *egcfgv1a1.KubernetesPodSpec, volumes []corev1.Volume) []corev1.Volume {
	amendFunc := func(volume corev1.Volume) {
		for index, e := range volumes {
			if e.Name == volume.Name {
				volumes[index] = volume
				return
			}
		}

		volumes = append(volumes, volume)
	}

	for _, envVar := range pod.Volumes {
		amendFunc(envVar)
	}

	return volumes
}

// ExpectedContainerVolumeMounts returns expected container volume mounts.
func ExpectedContainerVolumeMounts(container *egcfgv1a1.KubernetesContainerSpec, volumeMounts []corev1.VolumeMount) []corev1.VolumeMount {
	amendFunc := func(volumeMount corev1.VolumeMount) {
		for index, e := range volumeMounts {
			if e.Name == volumeMount.Name {
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
