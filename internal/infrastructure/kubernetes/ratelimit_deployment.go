// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	// Register embed
	_ "embed"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	// rateLimitRedisSocketTypeEnvVar is the redis socket type.
	rateLimitRedisSocketTypeEnvVar = "REDIS_SOCKET_TYPE"
	// rateLimitRedisURLEnvVar is the redis url.
	rateLimitRedisURLEnvVar = "REDIS_URL"
	// rateLimitRuntimeRootEnvVar is the runtime root.
	rateLimitRuntimeRootEnvVar = "RUNTIME_ROOT"
	// rateLimitRuntimeSubdirectoryEnvVar is the runtime subdirectory.
	rateLimitRuntimeSubdirectoryEnvVar = "RUNTIME_SUBDIRECTORY"
	// rateLimitRuntimeIgnoreDotfilesEnvVar is the runtime ignoredotfiles.
	rateLimitRuntimeIgnoreDotfilesEnvVar = "RUNTIME_IGNOREDOTFILES"
	// rateLimitRuntimeWatchRootEnvVar is the runtime watch root.
	rateLimitRuntimeWatchRootEnvVar = "RUNTIME_WATCH_ROOT"
	// rateLimitLogLevelEnvVar is the log level.
	rateLimitLogLevelEnvVar = "LOG_LEVEL"
	// rateLimitUseStatsdEnvVar is the use statsd.
	rateLimitUseStatsdEnvVar = "USE_STATSD"
	// rateLimitInfraName is the name for rate-limit resources.
	rateLimitInfraName = "envoy-ratelimit"
	// rateLimitInfraGRPCPort is the grpc port that the rate limit service listens on.
	rateLimitInfraGRPCPort = 8081
	// rateLimitInfraImage is the container image for the rate limit service.
	rateLimitInfraImage = "envoyproxy/ratelimit:f28024e3"
)

// expectedRateLimitDeployment returns the expected rate limit Deployment based on the provided infra.
func (i *Infra) expectedRateLimitDeployment(infra *ir.RateLimitInfra) *appsv1.Deployment {
	containers := expectedRateLimitContainers(infra)
	labels := rateLimitLabels()
	selector := getSelector(labels)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(int32(1)),
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            rateLimitInfraName,
					AutomountServiceAccountToken:  pointer.Bool(false),
					TerminationGracePeriodSeconds: pointer.Int64(int64(300)),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					Volumes: []corev1.Volume{
						{
							Name: rateLimitInfraName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: rateLimitInfraName,
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

	return deployment
}

func expectedRateLimitContainers(infra *ir.RateLimitInfra) []corev1.Container {
	ports := []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: rateLimitInfraGRPCPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	containers := []corev1.Container{
		{
			Name:            rateLimitInfraName,
			Image:           rateLimitInfraImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{
				"/bin/ratelimit",
			},
			Env: []corev1.EnvVar{
				{
					Name:  rateLimitRedisSocketTypeEnvVar,
					Value: "tcp",
				},
				{
					Name:  rateLimitRedisURLEnvVar,
					Value: infra.Backend.Redis.URL,
				},
				{
					Name:  rateLimitRuntimeRootEnvVar,
					Value: "/data",
				},
				{
					Name:  rateLimitRuntimeSubdirectoryEnvVar,
					Value: "ratelimit",
				},
				{
					Name:  rateLimitRuntimeIgnoreDotfilesEnvVar,
					Value: "true",
				},
				{
					Name:  rateLimitRuntimeWatchRootEnvVar,
					Value: "false",
				},
				{
					Name:  rateLimitLogLevelEnvVar,
					Value: "info",
				},
				{
					Name:  rateLimitUseStatsdEnvVar,
					Value: "false",
				},
			},
			Ports: ports,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      rateLimitInfraName,
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

// createOrUpdateRateLimitDeployment creates a Deployment in the kube api server based on the provided
// infra, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateRateLimitDeployment(ctx context.Context, infra *ir.RateLimitInfra) error {
	deployment := i.expectedRateLimitDeployment(infra)
	return i.createOrUpdateDeployment(ctx, deployment)
}

// deleteRateLimitDeployment deletes the Envoy RateLimit Deployment in the kube api server, if it exists.
func (i *Infra) deleteRateLimitDeployment(ctx context.Context, _ *ir.RateLimitInfra) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
		},
	}

	return i.deleteDeployment(ctx, deployment)
}
