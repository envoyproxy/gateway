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
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	// RedisSocketTypeEnvVar is the redis socket type.
	RedisSocketTypeEnvVar = "REDIS_SOCKET_TYPE"
	// RedisURLEnvVar is the redis url.
	RedisURLEnvVar = "REDIS_URL"
	// RedisTLSEnvVar is the redis tls.
	RedisTLSEnvVar = "REDIS_TLS"
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
	// GRPCTLSCertFilename is the GRPC tls cert file.
	GRPCTLSCertFilename = "/certs/tls.crt"
	// GRPCServerTLSKeyEnvVarEnvVar is the grpc server tls key.
	GRPCServerTLSKeyEnvVarEnvVar = "GRPC_SERVER_TLS_KEY"
	// GRPCTLSKeyFilename is the grpc server key file.
	GRPCTLSKeyFilename = "/certs/tls.key"
	// GRPCServerTLSCACertEnvVar is the grpc server tls ca cert.
	GRPCServerTLSCACertEnvVar = "GRPC_SERVER_TLS_CA_CERT"
	// GRPCTLSCACertFilename is the grpc server tls ca cert file.
	GRPCTLSCACertFilename = "/certs/ca.crt"
	// ConfigGRPCXDSServerUseTLSEnvVar is tls enable option for grpc xds server.
	ConfigGRPCXDSServerUseTLSEnvVar = "CONFIG_GRPC_XDS_SERVER_USE_TLS"
	// ConfigGRPCXDSClientTLSCertEnvVar is the grpc xds client tls cert.
	ConfigGRPCXDSClientTLSCertEnvVar = "CONFIG_GRPC_XDS_CLIENT_TLS_CERT"
	// ConfigGRPCXDSClientTLSKeyEnvVar is the grpc xds client tls key.
	ConfigGRPCXDSClientTLSKeyEnvVar = "CONFIG_GRPC_XDS_CLIENT_TLS_KEY"
	// ConfigGRPCXDSServerTLSCACertEnvVar is the grpc xds server tls ca cert.
	ConfigGRPCXDSServerTLSCACertEnvVar = "CONFIG_GRPC_XDS_SERVER_TLS_CACERT"
	// LogLevelEnvVar is the log level.
	LogLevelEnvVar = "LOG_LEVEL"
	// UseStatsdEnvVar is the use statsd.
	UseStatsdEnvVar = "USE_STATSD"
	// StatsdPortEnvVar is the use statsd port.
	StatsdPortEnvVar = "STATSD_PORT"
	// ForceStartWithoutInitialConfigEnvVar enables start the ratelimit server without initial config.
	ForceStartWithoutInitialConfigEnvVar = "FORCE_START_WITHOUT_INITIAL_CONFIG"
	// ConfigTypeEnvVar is the configuration loading method for ratelimit.
	ConfigTypeEnvVar = "CONFIG_TYPE"
	// ConfigGrpcXdsServerURLEnvVar is the url of ratelimit config xds server.
	ConfigGrpcXdsServerURLEnvVar = "CONFIG_GRPC_XDS_SERVER_URL"
	// ConfigGrpcXdsNodeIDEnvVar is the id of ratelimit node.
	ConfigGrpcXdsNodeIDEnvVar = "CONFIG_GRPC_XDS_NODE_ID"
	// TracingEnabledVar is enabled the tracing feature
	TracingEnabledVar = "TRACING_ENABLED"
	// TracingServiceNameVar is service name appears in tracing span
	TracingServiceNameVar = "TRACING_SERVICE_NAME"
	// TracingServiceNamespaceVar is service namespace appears in tracing span
	TracingServiceNamespaceVar = "TRACING_SERVICE_NAMESPACE"
	// TracingServiceInstanceIDVar is service instance id appears in tracing span
	TracingServiceInstanceIDVar = "TRACING_SERVICE_INSTANCE_ID"
	// TracingSamplingRateVar is trace sampling rate
	TracingSamplingRateVar = "TRACING_SAMPLING_RATE"
	// OTELExporterOTLPTraceEndpointVar is target url to which the trace exporter is going to send
	OTELExporterOTLPTraceEndpointVar = "OTEL_EXPORTER_OTLP_ENDPOINT"

	// InfraName is the base name for rate-limit resources. The per-instance
	// name is derived from the controller resource prefix via
	// config.InstanceResourceName, so that multiple instances can share a
	// namespace.
	InfraName = "envoy-ratelimit"
	// StatsdConfigMapName is the base name of the statsd exporter ConfigMap
	// created for the rate limit deployment.
	StatsdConfigMapName = "statsd-exporter-config"
	// InfraGRPCPort is the grpc port that the rate limit service listens on.
	InfraGRPCPort = 8081
	// XdsGrpcSotwConfigServerPort is the listening port of the ratelimit xDS config server.
	XdsGrpcSotwConfigServerPort = 18001
	// XdsGrpcSotwConfigServerHost is the base hostname of the ratelimit xDS
	// config server, fronted by the Envoy Gateway Service. The per-instance
	// hostname is derived from the controller resource prefix.
	XdsGrpcSotwConfigServerHost = "envoy-gateway"
	// ReadinessPath is readiness path for readiness probe.
	ReadinessPath = "/healthcheck"
	// ReadinessPort is readiness port for readiness probe.
	ReadinessPort  = 8080
	StatsdPort     = 9125
	PrometheusPort = 19001
)

// GetServiceURL returns the URL for the rate limit service of the instance
// with the given controller resource prefix.
func GetServiceURL(resourcePrefix, namespace, dnsDomain string) string {
	return fmt.Sprintf("grpc://%s.%s.svc.%s:%d",
		config.InstanceResourceName(resourcePrefix, InfraName), namespace, dnsDomain, InfraGRPCPort)
}

// LabelSelector returns the string slice form of the labels used for the envoy
// rate limit resources of the instance with the given controller resource
// prefix.
func LabelSelector(resourcePrefix string) []string {
	rlLabelMap := rateLimitLabels(resourcePrefix)
	retLabels := make([]string, 0, len(rlLabelMap))

	for labelK, labelV := range rlLabelMap {
		ls := strings.Join([]string{labelK, labelV}, "=")
		retLabels = append(retLabels, ls)
	}

	return retLabels
}

// StableLabelSelector returns the label requirements shared by the envoy rate
// limit resources of every instance, regardless of the per-instance resource
// prefix. It is used by tooling (egctl) that needs to find the rate limit pods
// of a namespace without knowing the instance's prefix.
func StableLabelSelector() []string {
	return []string{
		"app.kubernetes.io/component=ratelimit",
		"app.kubernetes.io/managed-by=envoy-gateway",
	}
}

// rateLimitLabels returns the labels used for the envoy rate limit resources of
// the instance with the given controller resource prefix. The
// app.kubernetes.io/name label carries the per-instance resource name so that
// the Service of one instance never selects the pods of another instance in a
// shared namespace.
func rateLimitLabels(resourcePrefix string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       config.InstanceResourceName(resourcePrefix, InfraName),
		"app.kubernetes.io/component":  "ratelimit",
		"app.kubernetes.io/managed-by": "envoy-gateway",
	}
}

// expectedRateLimitContainers returns expected rateLimit containers.
func expectedRateLimitContainers(
	rateLimit *egv1a1.RateLimit,
	rateLimitDeployment *egv1a1.KubernetesDeploymentSpec,
	namespace string,
	resourcePrefix string,
) []corev1.Container {
	ports := []corev1.ContainerPort{
		{
			Name:          "grpc",
			ContainerPort: InfraGRPCPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	if enablePrometheus(rateLimit) {
		ports = append(ports, corev1.ContainerPort{
			Name:          "metrics",
			ContainerPort: PrometheusPort,
			Protocol:      corev1.ProtocolTCP,
		})
	}

	containers := []corev1.Container{
		{
			Name:            config.InstanceResourceName(resourcePrefix, InfraName),
			Image:           *rateLimitDeployment.Container.Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{
				"/bin/ratelimit",
			},
			Env:                      expectedRateLimitContainerEnv(rateLimit, rateLimitDeployment, namespace, resourcePrefix),
			Ports:                    ports,
			Resources:                *rateLimitDeployment.Container.Resources,
			SecurityContext:          expectedRateLimitContainerSecurityContext(rateLimitDeployment),
			VolumeMounts:             expectedContainerVolumeMounts(rateLimit, rateLimitDeployment),
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
			StartupProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   ReadinessPath,
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: ReadinessPort},
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				FailureThreshold: 30,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   ReadinessPath,
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: ReadinessPort},
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    5,
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   ReadinessPath,
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: ReadinessPort},
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				FailureThreshold: 3,
			},
		},
	}

	return containers
}

// expectedContainerVolumeMounts returns expected rateLimit container volume mounts.
func expectedContainerVolumeMounts(rateLimit *egv1a1.RateLimit, rateLimitDeployment *egv1a1.KubernetesDeploymentSpec) []corev1.VolumeMount {
	var volumeMounts []corev1.VolumeMount

	// mount the cert
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      "certs",
		MountPath: "/certs",
		ReadOnly:  true,
	})

	if enablePrometheus(rateLimit) {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "statsd-exporter-config",
			MountPath: "/etc/statsd-exporter",
			ReadOnly:  true,
		})
	}

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
func expectedDeploymentVolumes(rateLimit *egv1a1.RateLimit, rateLimitDeployment *egv1a1.KubernetesDeploymentSpec, resourcePrefix string) []corev1.Volume {
	var volumes []corev1.Volume

	if rateLimit.Backend.Redis != nil &&
		rateLimit.Backend.Redis.TLS != nil &&
		rateLimit.Backend.Redis.TLS.CertificateRef != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "redis-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  string(rateLimit.Backend.Redis.TLS.CertificateRef.Name),
					DefaultMode: new(int32(420)),
				},
			},
		})
	}

	volumes = append(volumes, corev1.Volume{
		Name: "certs",
		VolumeSource: corev1.VolumeSource{
			// #nosec G101 - This refers to a Kubernetes secret volume, not a credential
			Secret: &corev1.SecretVolumeSource{
				SecretName:  config.InstanceResourceName(resourcePrefix, "envoy-rate-limit"),
				DefaultMode: new(int32(420)),
			},
		},
	})

	if enablePrometheus(rateLimit) {
		volumes = append(volumes, corev1.Volume{
			Name: "statsd-exporter-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: config.InstanceResourceName(resourcePrefix, StatsdConfigMapName),
					},
					Optional:    new(true),
					DefaultMode: new(int32(420)),
				},
			},
		})
	}

	return resource.ExpectedVolumes(rateLimitDeployment.Pod, volumes)
}

// expectedRateLimitContainerEnv returns expected rateLimit container envs.
func expectedRateLimitContainerEnv(
	rateLimit *egv1a1.RateLimit,
	rateLimitDeployment *egv1a1.KubernetesDeploymentSpec,
	namespace string,
	resourcePrefix string,
) []corev1.EnvVar {
	env := []corev1.EnvVar{
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
			Name:  ConfigTypeEnvVar,
			Value: "GRPC_XDS_SOTW",
		},
		{
			Name: ConfigGrpcXdsServerURLEnvVar,
			Value: net.JoinHostPort(config.InstanceResourceName(resourcePrefix, XdsGrpcSotwConfigServerHost),
				strconv.Itoa(XdsGrpcSotwConfigServerPort)),
		},
		{
			Name:  ConfigGrpcXdsNodeIDEnvVar,
			Value: config.InstanceResourceName(resourcePrefix, InfraName),
		},
		{
			Name:  GRPCServerUseTLSEnvVar,
			Value: "true",
		},
		{
			Name:  GRPCServerTLSCertEnvVar,
			Value: GRPCTLSCertFilename,
		},
		{
			Name:  GRPCServerTLSKeyEnvVarEnvVar,
			Value: GRPCTLSKeyFilename,
		},
		{
			Name:  GRPCServerTLSCACertEnvVar,
			Value: GRPCTLSCACertFilename,
		},
		{
			Name:  ConfigGRPCXDSServerUseTLSEnvVar,
			Value: "true",
		},
		{
			Name:  ConfigGRPCXDSClientTLSCertEnvVar,
			Value: GRPCTLSCertFilename,
		},
		{
			Name:  ConfigGRPCXDSClientTLSKeyEnvVar,
			Value: GRPCTLSKeyFilename,
		},
		{
			Name:  ConfigGRPCXDSServerTLSCACertEnvVar,
			Value: GRPCTLSCACertFilename,
		},
		{
			Name:  ForceStartWithoutInitialConfigEnvVar,
			Value: "true",
		},
	}

	if rateLimit.Backend.Redis != nil {
		redisURLEnv := corev1.EnvVar{Name: RedisURLEnvVar}
		if rateLimit.Backend.Redis.URLRef != nil {
			redisURLEnv.ValueFrom = &corev1.EnvVarSource{
				SecretKeyRef: rateLimit.Backend.Redis.URLRef.SecretKeyRef,
			}
		} else if rateLimit.Backend.Redis.URL != nil {
			redisURLEnv.Value = *rateLimit.Backend.Redis.URL
		}
		env = append(env, []corev1.EnvVar{
			{
				Name:  RedisSocketTypeEnvVar,
				Value: "tcp",
			},
			redisURLEnv,
		}...)
	}

	if rateLimit.Backend.Redis != nil && rateLimit.Backend.Redis.TLS != nil {
		env = append(env, corev1.EnvVar{
			Name:  RedisTLSEnvVar,
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

	if enablePrometheus(rateLimit) {
		env = append(env, corev1.EnvVar{
			Name:  "USE_PROMETHEUS",
			Value: "true",
		}, corev1.EnvVar{
			Name:  "PROMETHEUS_ADDR",
			Value: ":19001",
		}, corev1.EnvVar{
			Name:  "PROMETHEUS_MAPPER_YAML",
			Value: "/etc/statsd-exporter/conf.yaml",
		})
	}

	if enableTracing(rateLimit) {
		sampleRate := 1.0
		if rateLimit.Telemetry.Tracing.SamplingRate != nil {
			sampleRate = float64(*rateLimit.Telemetry.Tracing.SamplingRate) / 100.0
		}

		traceEndpoint := checkTraceEndpointScheme(rateLimit.Telemetry.Tracing.Provider.URL)
		tracingEnvs := []corev1.EnvVar{
			{
				Name:  TracingEnabledVar,
				Value: "true",
			},
			{
				Name:  TracingServiceNameVar,
				Value: config.InstanceResourceName(resourcePrefix, InfraName),
			},
			{
				Name:  TracingServiceNamespaceVar,
				Value: namespace,
			},
			{
				// By default, this is a random instanceID,
				// we use the RateLimit pod name as the trace service instanceID.
				Name: TracingServiceInstanceIDVar,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "metadata.name",
					},
				},
			},
			{
				Name: TracingSamplingRateVar,
				// The api is configured with [0,100], but sampling can only be [0,1].
				// doc: https://github.com/envoyproxy/ratelimit?tab=readme-ov-file#tracing
				// You will lose precision during the conversion process, but don't worry,
				// this follows the rounding rule and won't make the expected sampling rate too different
				// from the actual sampling rate
				Value: strconv.FormatFloat(sampleRate, 'f', 1, 64),
			},
			{
				Name:  OTELExporterOTLPTraceEndpointVar,
				Value: traceEndpoint,
			},
		}
		env = append(env, tracingEnvs...)
	}

	return resource.ExpectedContainerEnv(rateLimitDeployment.Container, env)
}

// Validate validates the ratelimit redis url/tls secret references.
func Validate(ctx context.Context, client client.Client, gateway *egv1a1.EnvoyGateway, namespace string) error {
	redis := gateway.RateLimit.Backend.Redis
	if redis == nil {
		return nil
	}

	if redis.URLRef != nil && redis.URLRef.SecretKeyRef != nil {
		ref := redis.URLRef.SecretKeyRef
		secret := &corev1.Secret{}
		err := client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: ref.Name}, secret)
		switch {
		case apierrors.IsNotFound(err):
			// The Secret may be provisioned after the EnvoyGateway config (e.g. by an
			// external controller such as Crossplane). The ratelimit Deployment consumes
			// the URL via valueFrom.secretKeyRef, so the kubelet starts the container once
			// the Secret and key exist. Don't block infra creation on a not-yet-present Secret.
		case err != nil:
			return fmt.Errorf("failed to get Secret %s in namespace %s: %w", ref.Name, namespace, err)
		default:
			// Validate the value only when present; a missing key is resolved later by the kubelet.
			if value, ok := secret.Data[ref.Key]; ok {
				if verr := validation.ValidateRedisURL(string(value)); verr != nil {
					return fmt.Errorf("invalid Redis URL in Secret %s/%s key %q: %w", namespace, ref.Name, ref.Key, verr)
				}
			}
		}
	}

	if redis.TLS != nil && redis.TLS.CertificateRef != nil {
		_, _, err := kubernetes.ValidateSecretObjectReference(ctx, client, redis.TLS.CertificateRef, namespace)
		return err
	}

	return nil
}

func enableTracing(rl *egv1a1.RateLimit) bool {
	// Other fields can use the default values,
	// but we have to make sure the user has the Provider.URL
	if rl != nil && rl.Telemetry != nil &&
		rl.Telemetry.Tracing != nil &&
		rl.Telemetry.Tracing.Provider != nil &&
		len(rl.Telemetry.Tracing.Provider.URL) != 0 {
		return true
	}

	return false
}

// checkTraceEndpointScheme Check the scheme prefix in the trace url
func checkTraceEndpointScheme(url string) string {
	// Since the OTLP collector needs to configure the scheme prefix,
	// we need to check if the user has configured this
	// TODO: It is currently assumed to be a normal connection,
	//  	 and a TLS connection will be added later.
	httpScheme := "http://"
	exist := strings.HasPrefix(url, httpScheme)
	if exist {
		return url
	}

	return fmt.Sprintf("%s%s", httpScheme, url)
}

func expectedRateLimitContainerSecurityContext(rateLimitDeployment *egv1a1.KubernetesDeploymentSpec) *corev1.SecurityContext {
	if rateLimitDeployment.Container.SecurityContext != nil {
		return rateLimitDeployment.Container.SecurityContext
	}
	return defaultSecurityContext()
}

func defaultSecurityContext() *corev1.SecurityContext {
	defaultSC := resource.DefaultSecurityContext()
	// run as non-root user
	defaultSC.RunAsGroup = new(int64(65534))
	defaultSC.RunAsUser = new(int64(65534))
	return defaultSC
}
