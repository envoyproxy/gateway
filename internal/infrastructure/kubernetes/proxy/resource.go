// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/cmd/envoy"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

const (
	// envoyContainerName is the name of the Envoy container.
	envoyContainerName = "envoy"
	// envoyNsEnvVar is the name of the Envoy Gateway namespace environment variable.
	envoyNsEnvVar = "ENVOY_GATEWAY_NAMESPACE"
	// envoyPodEnvVar is the name of the Envoy pod name environment variable.
	envoyPodEnvVar = "ENVOY_POD_NAME"
)

// ExpectedResourceHashedName returns expected resource hashed name including up to the 48 characters of the original name.
func ExpectedResourceHashedName(name string) string {
	hashedName := utils.GetHashedName(name, 48)
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, hashedName)
}

// EnvoyAppLabel returns the labels used for all Envoy resources.
func EnvoyAppLabel() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "envoy",
		"app.kubernetes.io/component":  "proxy",
		"app.kubernetes.io/managed-by": "envoy-gateway",
	}
}

// EnvoyAppLabelSelector returns the labels used for all Envoy resources.
func EnvoyAppLabelSelector() []string {
	return []string{
		"app.kubernetes.io/name=envoy",
		"app.kubernetes.io/component=proxy",
		"app.kubernetes.io/managed-by=envoy-gateway",
	}
}

// envoyLabels returns the labels, including extraLabels, used for Envoy resources.
func envoyLabels(extraLabels map[string]string) map[string]string {
	labels := EnvoyAppLabel()
	for k, v := range extraLabels {
		labels[k] = v
	}

	return labels
}

func enablePrometheus(infra *ir.ProxyInfra) bool {
	if infra.Config != nil &&
		infra.Config.Spec.Telemetry != nil &&
		infra.Config.Spec.Telemetry.Metrics != nil &&
		infra.Config.Spec.Telemetry.Metrics.Prometheus != nil &&
		infra.Config.Spec.Telemetry.Metrics.Prometheus.Disable {
		return false
	}

	return true
}

// expectedProxyContainers returns expected proxy containers.
func expectedProxyContainers(infra *ir.ProxyInfra,
	containerSpec *egv1a1.KubernetesContainerSpec,
	shutdownConfig *egv1a1.ShutdownConfig, shutdownManager *egv1a1.ShutdownManager,
	namespace string, dnsDomain string,
) ([]corev1.Container, error) {
	ports := make([]corev1.ContainerPort, 0, 2)
	if enablePrometheus(infra) {
		ports = append(ports, corev1.ContainerPort{
			Name:          "metrics",
			ContainerPort: bootstrap.EnvoyStatsPort, // TODO: make this configurable
			Protocol:      corev1.ProtocolTCP,
		})
	}

	ports = append(ports, corev1.ContainerPort{
		Name:          "readiness",
		ContainerPort: bootstrap.EnvoyReadinessPort, // TODO: make this configurable
		Protocol:      corev1.ProtocolTCP,
	})

	var proxyMetrics *egv1a1.ProxyMetrics
	if infra.Config != nil &&
		infra.Config.Spec.Telemetry != nil {
		proxyMetrics = infra.Config.Spec.Telemetry.Metrics
	}

	maxHeapSizeBytes := calculateMaxHeapSizeBytes(containerSpec.Resources)

	// Get the default Bootstrap
	bootstrapConfigOptions := &bootstrap.RenderBootstrapConfigOptions{
		ProxyMetrics: proxyMetrics,
		SdsConfig: bootstrap.SdsConfigPath{
			Certificate: filepath.Join("/sds", common.SdsCertFilename),
			TrustedCA:   filepath.Join("/sds", common.SdsCAFilename),
		},
		MaxHeapSizeBytes: maxHeapSizeBytes,
		XdsServerHost:    ptr.To(fmt.Sprintf("%s.%s.svc.%s", config.EnvoyGatewayServiceName, namespace, dnsDomain)),
	}

	args, err := common.BuildProxyArgs(infra, shutdownConfig, bootstrapConfigOptions, fmt.Sprintf("$(%s)", envoyPodEnvVar))
	if err != nil {
		return nil, err
	}

	containers := []corev1.Container{
		{
			Name:                     envoyContainerName,
			Image:                    *containerSpec.Image,
			ImagePullPolicy:          corev1.PullIfNotPresent,
			Command:                  []string{"envoy"},
			Args:                     args,
			Env:                      expectedContainerEnv(containerSpec),
			Resources:                *containerSpec.Resources,
			SecurityContext:          expectedEnvoySecurityContext(containerSpec),
			Ports:                    ports,
			VolumeMounts:             expectedContainerVolumeMounts(containerSpec),
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
			StartupProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   bootstrap.EnvoyReadinessPath,
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: bootstrap.EnvoyReadinessPort},
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
						Path:   bootstrap.EnvoyReadinessPath,
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: bootstrap.EnvoyReadinessPort},
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    5,
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   envoy.ShutdownManagerReadyPath,
						Port:   intstr.FromInt32(envoy.ShutdownManagerPort),
						Scheme: corev1.URISchemeHTTP,
					},
				},
			},
		},
		{
			Name:                     "shutdown-manager",
			Image:                    expectedShutdownManagerImage(shutdownManager),
			ImagePullPolicy:          corev1.PullIfNotPresent,
			Command:                  []string{"envoy-gateway"},
			Args:                     expectedShutdownManagerArgs(shutdownConfig),
			Env:                      expectedContainerEnv(nil),
			Resources:                *egv1a1.DefaultShutdownManagerContainerResourceRequirements(),
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
			StartupProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   envoy.ShutdownManagerHealthCheckPath,
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: envoy.ShutdownManagerPort},
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
						Path:   envoy.ShutdownManagerHealthCheckPath,
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: envoy.ShutdownManagerPort},
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				FailureThreshold: 3,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   envoy.ShutdownManagerHealthCheckPath,
						Port:   intstr.IntOrString{Type: intstr.Int, IntVal: envoy.ShutdownManagerPort},
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				FailureThreshold: 3,
			},
			Lifecycle: &corev1.Lifecycle{
				PreStop: &corev1.LifecycleHandler{
					Exec: &corev1.ExecAction{
						Command: expectedShutdownPreStopCommand(shutdownConfig),
					},
				},
			},
			SecurityContext: expectedShutdownManagerSecurityContext(containerSpec),
		},
	}

	return containers, nil
}

func expectedShutdownManagerImage(shutdownManager *egv1a1.ShutdownManager) string {
	if shutdownManager != nil && shutdownManager.Image != nil {
		return *shutdownManager.Image
	}
	return egv1a1.DefaultShutdownManagerImage
}

func expectedShutdownManagerArgs(cfg *egv1a1.ShutdownConfig) []string {
	args := []string{"envoy", "shutdown-manager"}
	if cfg != nil && cfg.DrainTimeout != nil {
		args = append(args, fmt.Sprintf("--ready-timeout=%.0fs", cfg.DrainTimeout.Seconds()+10))
	}
	return args
}

func expectedShutdownPreStopCommand(cfg *egv1a1.ShutdownConfig) []string {
	command := []string{"envoy-gateway", "envoy", "shutdown"}

	if cfg == nil {
		return command
	}

	if cfg.DrainTimeout != nil {
		command = append(command, fmt.Sprintf("--drain-timeout=%.0fs", cfg.DrainTimeout.Seconds()))
	}

	if cfg.MinDrainDuration != nil {
		command = append(command, fmt.Sprintf("--min-drain-duration=%.0fs", cfg.MinDrainDuration.Seconds()))
	}

	return command
}

// expectedContainerVolumeMounts returns expected proxy container volume mounts.
func expectedContainerVolumeMounts(containerSpec *egv1a1.KubernetesContainerSpec) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "certs",
			MountPath: "/certs",
			ReadOnly:  true,
		},
		{
			Name:      "sds",
			MountPath: "/sds",
		},
	}

	return resource.ExpectedContainerVolumeMounts(containerSpec, volumeMounts)
}

// expectedVolumes returns expected proxy deployment volumes.
func expectedVolumes(name string, pod *egv1a1.KubernetesPodSpec) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "envoy",
					DefaultMode: ptr.To[int32](420),
				},
			},
		},
		{
			Name: "sds",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ExpectedResourceHashedName(name),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  common.SdsCAFilename,
							Path: common.SdsCAFilename,
						},
						{
							Key:  common.SdsCertFilename,
							Path: common.SdsCertFilename,
						},
					},
					DefaultMode: ptr.To[int32](420),
					Optional:    ptr.To(false),
				},
			},
		},
	}

	return resource.ExpectedVolumes(pod, volumes)
}

// expectedContainerEnv returns expected proxy container envs.
func expectedContainerEnv(containerSpec *egv1a1.KubernetesContainerSpec) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name: envoyNsEnvVar,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: envoyPodEnvVar,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
	}

	if containerSpec != nil {
		return resource.ExpectedContainerEnv(containerSpec, env)
	} else {
		return env
	}
}

// calculateMaxHeapSizeBytes calculates the maximum heap size in bytes as 80% of Envoy container memory limits.
// In case no limits are defined '0' is returned, which means no heap size limit is set.
func calculateMaxHeapSizeBytes(envoyResourceRequirements *corev1.ResourceRequirements) uint64 {
	if envoyResourceRequirements == nil || envoyResourceRequirements.Limits == nil {
		return 0
	}

	if memLimit, ok := envoyResourceRequirements.Limits[corev1.ResourceMemory]; ok {
		memLimitBytes := memLimit.Value()
		return uint64(float64(memLimitBytes) * 0.8)
	}

	return 0
}

func expectedEnvoySecurityContext(containerSpec *egv1a1.KubernetesContainerSpec) *corev1.SecurityContext {
	if containerSpec != nil && containerSpec.SecurityContext != nil {
		return containerSpec.SecurityContext
	}

	sc := resource.DefaultSecurityContext()

	// run as non-root user
	sc.RunAsGroup = ptr.To(int64(65532))
	sc.RunAsUser = ptr.To(int64(65532))

	// Envoy container needs to write to the log file/UDS socket.
	sc.ReadOnlyRootFilesystem = nil
	return sc
}

func expectedShutdownManagerSecurityContext(containerSpec *egv1a1.KubernetesContainerSpec) *corev1.SecurityContext {
	if containerSpec != nil && containerSpec.SecurityContext != nil {
		return containerSpec.SecurityContext
	}

	sc := resource.DefaultSecurityContext()

	// run as non-root user
	sc.RunAsGroup = ptr.To(int64(65532))
	sc.RunAsUser = ptr.To(int64(65532))

	// ShutdownManger creates a file to indicate the connection drain process is completed,
	// so it needs file write permission.
	sc.ReadOnlyRootFilesystem = nil
	return sc
}
