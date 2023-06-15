// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	"context"
	"fmt"
	"net"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
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
	// RedisTLSClientCertFilename is the redis tls client cert file.
	RedisTLSClientCertFilename = "/redis-certs/tls.crt"
	// RedisTLSClientKeyEnvVar is the redis tls client key.
	RedisTLSClientKeyEnvVar = "REDIS_TLS_CLIENT_KEY"
	// RedisTLSClientKeyFilename is the redis client key file.
	RedisTLSClientKeyFilename = "/redis-certs/tls.key"
	// RuntimeRootEnvVar is the runtime root.
	RuntimeRootEnvVar = "RUNTIME_ROOT"
	// RuntimeSubdirectoryEnvVar is the runtime subdirectory.
	RuntimeSubdirectoryEnvVar = "RUNTIME_SUBDIRECTORY"
	// RuntimeIgnoreDotfilesEnvVar is the runtime ignoredotfiles.
	RuntimeIgnoreDotfilesEnvVar = "RUNTIME_IGNOREDOTFILES"
	// RuntimeWatchRootEnvVar is the runtime watch root.
	RuntimeWatchRootEnvVar = "RUNTIME_WATCH_ROOT"
	// GRPCServerUseTLSEnvVar is tls enable option for grpc server.
	GRPCServerUseTLSEnvVar = "GRPC_SERVER_USE_TLS"
	// GRPCServerTLSCertEnvVar is the grpc server tls cert.
	GRPCServerTLSCertEnvVar = "GRPC_SERVER_TLS_CERT"
	// GRPCServerTLSClientCertFilename is the GRPC tls cert file.
	GRPCServerTLSCertFilename = "/certs/tls.crt"
	// GRPCServerTLSKeyEnvVarEnvVar is the grpc server tls key.
	GRPCServerTLSKeyEnvVarEnvVar = "GRPC_SERVER_TLS_KEY"
	// GRPCServerTLSKeyFilename is the grpc server key file.
	GRPCServerTLSKeyFilename = "/certs/tls.key"
	// GRPCServerTLSCACertEnvVar is the grpc server tls ca cert.
	GRPCServerTLSCACertEnvVar = "GRPC_SERVER_TLS_CA_CERT"
	// GRPCServerTLSKeyFilename is the grpc server tls ca cert file.
	GRPCServerTLSCACertFilename = "/certs/ca.crt"
	// LogLevelEnvVar is the log level.
	LogLevelEnvVar = "LOG_LEVEL"
	// UseStatsdEnvVar is the use statsd.
	UseStatsdEnvVar = "USE_STATSD"
	// InfraName is the name for rate-limit resources.
	InfraName = "envoy-ratelimit"
	// InfraGRPCPort is the grpc port that the rate limit service listens on.
	InfraGRPCPort = 8081
	// ConfigType is the configuration loading method for ratelimit.
	ConfigType = "CONFIG_TYPE"
	// ConfigGrpcXdsServerURL is the url of ratelimit config xds server.
	ConfigGrpcXdsServerURL = "CONFIG_GRPC_XDS_SERVER_URL"
	// ConfigGrpcXdsNodeID is the id of ratelimit node.
	ConfigGrpcXdsNodeID = "CONFIG_GRPC_XDS_NODE_ID"

	// XdsGrpcSotwConfigServerPort is the listening port of the ratelimit xDS config server.
	XdsGrpcSotwConfigServerPort = 18001
	// XdsGrpcSotwConfigServerHost is the hostname of the ratelimit xDS config server.
	XdsGrpcSotwConfigServerHost = "envoy-gateway"
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
			VolumeMounts:             expectedContainerVolumeMounts(rateLimit, rateLimitDeployment),
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
		},
	}

	return containers
}

// expectedContainerVolumeMounts returns expected rateLimit container volume mounts.
func expectedContainerVolumeMounts(rateLimit *egcfgv1a1.RateLimit, rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount

	// mount the cert
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "certs",
		MountPath: "/certs",
		ReadOnly:  true,
	})

	if rateLimit.Backend.Redis.TLS != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "redis-certs",
			MountPath: "/redis-certs",
			ReadOnly:  true,
		})
	}

	return resource.ExpectedContainerVolumeMounts(rateLimitDeployment.Container, volumeMounts)
}

// expectedDeploymentVolumes returns expected rateLimit deployment volumes.
func expectedDeploymentVolumes(rateLimit *egcfgv1a1.RateLimit, rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec) []corev1.Volume {
	var volumes []corev1.Volume

	if rateLimit.Backend.Redis.TLS != nil && rateLimit.Backend.Redis.TLS.CertificateRef != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "redis-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: string(rateLimit.Backend.Redis.TLS.CertificateRef.Name),
				},
			},
		})
	}

	volumes = append(volumes, corev1.Volume{
		Name: "certs",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: "envoy-rate-limit",
			},
		},
	})

	return resource.ExpectedDeploymentVolumes(rateLimitDeployment.Pod, volumes)
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
		{
			Name:  ConfigType,
			Value: "GRPC_XDS_SOTW",
		},
		{
			Name:  ConfigGrpcXdsServerURL,
			Value: net.JoinHostPort(XdsGrpcSotwConfigServerHost, strconv.Itoa(XdsGrpcSotwConfigServerPort)),
		},
		{
			Name:  ConfigGrpcXdsNodeID,
			Value: InfraName,
		},
		{
			Name:  GRPCServerUseTLSEnvVar,
			Value: "true",
		},
		{
			Name:  GRPCServerTLSCertEnvVar,
			Value: GRPCServerTLSCertFilename,
		},
		{
			Name:  GRPCServerTLSKeyEnvVarEnvVar,
			Value: GRPCServerTLSKeyFilename,
		},
		{
			Name:  GRPCServerTLSCACertEnvVar,
			Value: GRPCServerTLSCACertFilename,
		},
	}

	if rateLimit.Backend.Redis.TLS != nil {
		env = append(env, corev1.EnvVar{
			Name:  RedisTLS,
			Value: "true",
		})

		if rateLimit.Backend.Redis.TLS.CertificateRef != nil {
			env = append(env, []corev1.EnvVar{
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
	}

	return resource.ExpectedProxyContainerEnv(rateLimitDeployment.Container, env)
}

// Validate the ratelimit tls secret validating.
func Validate(ctx context.Context, client client.Client, gateway *egcfgv1a1.EnvoyGateway, namespace string) error {
	if gateway.RateLimit.Backend.Redis.TLS != nil && gateway.RateLimit.Backend.Redis.TLS.CertificateRef != nil {
		certificateRef := gateway.RateLimit.Backend.Redis.TLS.CertificateRef
		_, _, err := kubernetes.ValidateSecretObjectReference(ctx, client, certificateRef, namespace)
		return err
	}

	return nil
}
