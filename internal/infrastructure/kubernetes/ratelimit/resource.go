// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	// RedisSocketTypeEnvVar is the redis socket type.
	RedisSocketTypeEnvVar = "REDIS_SOCKET_TYPE"
	// RedisURLEnvVar is the redis url.
	RedisURLEnvVar = "REDIS_URL"
	// RedisTLS is the redis tls.
	RedisTLS = "REDIS_TLS"
	// RedisTLSClientCertEnvVar is the redis tls client cert.
	RedisTLSClientCertEnvVar = "REDIS_TLS_CLIENT_CERT"
	// RedisTLSClientCertFilename is the reds tls client cert file.
	RedisTLSClientCertFilename = "/certs/tls.crt"
	// RedisTLSClientKeyEnvVar is the redis tls client key.
	RedisTLSClientKeyEnvVar = "REDIS_TLS_CLIENT_KEY"
	// RedisTLSClientKeyFilename is the redis client key file.
	RedisTLSClientKeyFilename = "/certs/tls.key"
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

// GetServiceURL returns the URL for the rate limit service.
func GetServiceURL(namespace string, dnsDomain string) string {
	return fmt.Sprintf("grpc://%s.%s.svc.%s:%d", InfraName, namespace, dnsDomain, InfraGRPCPort)
}

// rateLimitLabels returns the labels used for all envoy rate limit resources.
func rateLimitLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       InfraName,
		"app.kubernetes.io/component":  "ratelimit",
		"app.kubernetes.io/managed-by": "envoy-gateway",
	}
}

// expectedRateLimitContainers returns expected rateLimit containers.
func expectedRateLimitContainers(rateLimit *egcfgv1a1.RateLimit, rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec) []corev1.Container {
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
			Env:                      expectedRateLimitContainerEnv(rateLimit, rateLimitDeployment),
			Ports:                    ports,
			Resources:                *rateLimitDeployment.Container.Resources,
			SecurityContext:          rateLimitDeployment.Container.SecurityContext,
			VolumeMounts:             expectedContainerVolumeMounts(rateLimit),
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
		},
	}

	return containers
}

// expectedContainerVolumeMounts returns expected rateLimit container volume mounts.
func expectedContainerVolumeMounts(rateLimit *egcfgv1a1.RateLimit) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      InfraName,
			MountPath: "/data/ratelimit/config",
			ReadOnly:  true,
		},
	}

	// mount the cert
	if rateLimit.Backend.Redis.TLS != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "certs",
			MountPath: "/certs",
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

// expectedDeploymentVolumes returns expected rateLimit deployment volumes.
func expectedDeploymentVolumes(rateLimit *egcfgv1a1.RateLimit) []corev1.Volume {
	volumes := []corev1.Volume{
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
	}

	if rateLimit.Backend.Redis.TLS != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: string(rateLimit.Backend.Redis.TLS.CertificateRef.Name),
				},
			},
		})
	}
	return volumes
}

// expectedRateLimitContainerEnv returns expected rateLimit container envs.
func expectedRateLimitContainerEnv(rateLimit *egcfgv1a1.RateLimit, rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name:  RedisSocketTypeEnvVar,
			Value: "tcp",
		},
		{
			Name:  RedisURLEnvVar,
			Value: rateLimit.Backend.Redis.URL,
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
	}

	if rateLimit.Backend.Redis.TLS != nil {
		env = append(env, []corev1.EnvVar{
			{
				Name:  RedisTLS,
				Value: "true",
			},
			{
				Name:  RedisTLSClientCertEnvVar,
				Value: RedisTLSClientCertFilename,
			},
			{
				Name:  RedisTLSClientKeyEnvVar,
				Value: RedisTLSClientKeyFilename,
			},
		}...)
	}

	envAmendFunc := func(envVar corev1.EnvVar) {
		for index, e := range env {
			if e.Name == envVar.Name {
				env[index] = envVar
				return
			}
		}
		env = append(env, envVar)
	}

	for _, envVar := range rateLimitDeployment.Container.Env {
		envAmendFunc(envVar)
	}

	return env
}

// Validate the ratelimit tls secret validating.
func Validate(ctx context.Context, client client.Client, gateway *egcfgv1a1.EnvoyGateway, namespace string) error {
	if gateway.RateLimit.Backend.Redis.TLS == nil {
		return nil
	}

	certificateRef := gateway.RateLimit.Backend.Redis.TLS.CertificateRef
	_, _, err := kubernetes.ValidateSecretObjectReference(ctx, client, certificateRef, namespace)

	return err
}
