// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
)

const (
	// RedisSocketTypeEnvVar is the redis socket type.
	RedisSocketTypeEnvVar = "REDIS_SOCKET_TYPE"
	// RedisURLEnvVar is the redis url.
	RedisURLEnvVar = "REDIS_URL"
	// RuntimeRootEnvVar is the runtime root.
	RuntimeRootEnvVar = "RUNTIME_ROOT"
	// RuntimeSubdirectoryEnvVar is the runtime subdirectory.
	RuntimeSubdirectoryEnvVar = "RUNTIME_SUBDIRECTORY"
	// RuntimeIgnoreDotfilesEnvVar is the runtime ignoredotfiles.
	RuntimeIgnoreDotfilesEnvVar = "RUNTIME_IGNOREDOTFILES"
	// RuntimeWatchRootEnvVar is the runtime watch root.
	RuntimeWatchRootEnvVar = "RUNTIME_WATCH_ROOT"
	// LogLevelEnvVar is the log level.
	LogLevelEnvVar = "LOG_LEVEL"
	// UseStatsdEnvVar is the use statsd.
	UseStatsdEnvVar = "USE_STATSD"
	// InfraName is the name for rate-limit resources.
	InfraName = "envoy-ratelimit"
	// InfraGRPCPort is the grpc port that the rate limit service listens on.
	InfraGRPCPort = 8081
)

// Deployment returns the expected rate limit Deployment based on the provided infra.
func (i *ResourceRender) Deployment() (*appsv1.Deployment, error) {
	containers := expectedRateLimitContainers(i.ratelimit, i.rateLimitDeployment)
	labels := rateLimitLabels()
	selector := resource.GetSelector(labels)

	var annos map[string]string
	if i.rateLimitDeployment.Pod.Annotations != nil {
		annos = i.rateLimitDeployment.Pod.Annotations
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      InfraName,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: i.rateLimitDeployment.Replicas,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      selector.MatchLabels,
					Annotations: annos,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            InfraName,
					AutomountServiceAccountToken:  pointer.Bool(false),
					TerminationGracePeriodSeconds: pointer.Int64(int64(300)),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					SecurityContext:               i.rateLimitDeployment.Pod.SecurityContext,
					Volumes: []corev1.Volume{
						{
							Name: InfraName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: InfraName,
									},
									DefaultMode: pointer.Int32(int32(420)),
									Optional:    pointer.Bool(false),
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}

func expectedRateLimitContainers(ratelimit *egcfgv1a1.RateLimit, rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec) []corev1.Container {
	ports := []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: InfraGRPCPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	containers := []corev1.Container{
		{
			Name:            InfraName,
			Image:           *rateLimitDeployment.Container.Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{
				"/bin/ratelimit",
			},
			Env: []corev1.EnvVar{
				{
					Name:  RedisSocketTypeEnvVar,
					Value: "tcp",
				},
				{
					Name:  RedisURLEnvVar,
					Value: ratelimit.Backend.Redis.URL,
				},
				{
					Name:  RuntimeRootEnvVar,
					Value: "/data",
				},
				{
					Name:  RuntimeSubdirectoryEnvVar,
					Value: "ratelimit",
				},
				{
					Name:  RuntimeIgnoreDotfilesEnvVar,
					Value: "true",
				},
				{
					Name:  RuntimeWatchRootEnvVar,
					Value: "false",
				},
				{
					Name:  LogLevelEnvVar,
					Value: "info",
				},
				{
					Name:  UseStatsdEnvVar,
					Value: "false",
				},
			},
			Ports:           ports,
			Resources:       *rateLimitDeployment.Container.Resources,
			SecurityContext: rateLimitDeployment.Container.SecurityContext,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      InfraName,
					MountPath: "/data/ratelimit/config",
					ReadOnly:  true,
				},
			},
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
		},
	}

	return containers
}
