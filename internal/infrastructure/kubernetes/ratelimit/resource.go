// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
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

// GetServiceURL returns the URL for the rate limit service.
// TODO: support custom trust domain
func GetServiceURL(namespace string) string {
	return fmt.Sprintf("grpc://%s.%s.svc.cluster.local:%d", InfraName, namespace, InfraGRPCPort)
}

// rateLimitLabels returns the labels used for all envoy rate limit resources.
func rateLimitLabels() map[string]string {
	return map[string]string{
		"app.gateway.envoyproxy.io/name": InfraName,
	}
}

// expectedRateLimitContainers returns expected rateLimit containers.
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
			Env:             expectedRateLimitContainerEnv(ratelimit, rateLimitDeployment),
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

// expectedRateLimitContainerEnv returns expected ratelimit container envs.
func expectedRateLimitContainerEnv(ratelimit *egcfgv1a1.RateLimit, rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	defaultEnvMapping := map[string]string{
		RedisSocketTypeEnvVar:       "tcp",
		RuntimeRootEnvVar:           "/data",
		RuntimeSubdirectoryEnvVar:   "ratelimit",
		RuntimeIgnoreDotfilesEnvVar: "true",
		RuntimeWatchRootEnvVar:      "false",
		LogLevelEnvVar:              "info",
		UseStatsdEnvVar:             "false",
	}

	for _, envVar := range rateLimitDeployment.Container.Env {
		// extension env provides current env, remove from default env mapping
		if defaultEnvMapping[envVar.Name] != "" {
			delete(defaultEnvMapping, envVar.Name)
		}

		// override env except REDIS_URL
		if envVar.Name != RedisURLEnvVar {
			env = append(env, envVar)
		}
	}

	// append default env that extension env does not provide.
	for key, value := range defaultEnvMapping {
		env = append(env, corev1.EnvVar{Name: key, Value: value})
	}

	return append(env, corev1.EnvVar{
		Name:  RedisURLEnvVar,
		Value: ratelimit.Backend.Redis.URL,
	})
}
