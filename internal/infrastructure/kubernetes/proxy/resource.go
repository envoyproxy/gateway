// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/containers/image/v5/docker/reference"
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
	// envoyNsEnvVar is the name of the Envoy pod namespace environment variable.
	envoyNsEnvVar = "ENVOY_POD_NAMESPACE"
	// envoyPodEnvVar is the name of the Envoy pod name environment variable.
	envoyPodEnvVar = "ENVOY_POD_NAME"
	// envoyZoneEnvVar is the Envoy pod locality zone name
	envoyZoneEnvVar = "ENVOY_SERVICE_ZONE"
)

// ExpectedResourceHashedName returns expected resource hashed name including up to the 48 characters of the original name.
// WARNING: DO NOT USE THIS FUNCTION IN MOST OF THE CASES. Use ResourceRender.Name() instead.
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
	topologyInjectorDisabled bool,
	controllerNamespace, dnsDomain string, gatewayNamespaceMode bool,
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
		MaxHeapSizeBytes:         maxHeapSizeBytes,
		XdsServerHost:            ptr.To(fmt.Sprintf("%s.%s.svc.%s.", config.EnvoyGatewayServiceName, controllerNamespace, dnsDomain)),
		TopologyInjectorDisabled: topologyInjectorDisabled,
	}

	args, err := common.BuildProxyArgs(infra, shutdownConfig, bootstrapConfigOptions, fmt.Sprintf("$(%s)", envoyPodEnvVar), gatewayNamespaceMode)
	if err != nil {
		return nil, err
	}

	proxyImage, err := resolveProxyImage(containerSpec)
	if err != nil {
		return nil, err
	}

	containers := []corev1.Container{
		{
			Name:                     envoyContainerName,
			Image:                    proxyImage,
			ImagePullPolicy:          corev1.PullIfNotPresent,
			Command:                  []string{"envoy"},
			Args:                     args,
			Env:                      expectedContainerEnv(containerSpec),
			Resources:                *containerSpec.Resources,
			SecurityContext:          expectedEnvoySecurityContext(containerSpec),
			Ports:                    ports,
			VolumeMounts:             expectedContainerVolumeMounts(containerSpec, gatewayNamespaceMode),
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
			LivenessProbe: &corev1.Probe{
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
				FailureThreshold: 3,
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
		d, err := time.ParseDuration(string(*cfg.DrainTimeout))
		if err != nil {
			return nil
		}
		args = append(args, fmt.Sprintf("--ready-timeout=%.0fs", d.Seconds()+10))
	}
	return args
}

func expectedShutdownPreStopCommand(cfg *egv1a1.ShutdownConfig) []string {
	command := []string{"envoy-gateway", "envoy", "shutdown"}

	if cfg == nil {
		return command
	}

	if cfg.DrainTimeout != nil {
		d, err := time.ParseDuration(string(*cfg.DrainTimeout))
		if err != nil {
			return nil
		}
		command = append(command, fmt.Sprintf("--drain-timeout=%.0fs", d.Seconds()))
	}

	if cfg.MinDrainDuration != nil {
		d, err := time.ParseDuration(string(*cfg.MinDrainDuration))
		if err != nil {
			return nil
		}
		command = append(command, fmt.Sprintf("--min-drain-duration=%.0fs", d.Seconds()))
	}

	return command
}

// expectedContainerVolumeMounts returns expected proxy container volume mounts.
func expectedContainerVolumeMounts(containerSpec *egv1a1.KubernetesContainerSpec, gatewayNamespaceMode bool) []corev1.VolumeMount {
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
	if gatewayNamespaceMode {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "sa-token",
			MountPath: "/var/run/secrets/token",
			ReadOnly:  true,
		})
	}

	return resource.ExpectedContainerVolumeMounts(containerSpec, volumeMounts)
}

// expectedVolumes returns expected proxy deployment volumes.
func (r *ResourceRender) expectedVolumes(pod *egv1a1.KubernetesPodSpec) []corev1.Volume {
	var volumes []corev1.Volume
	certsVolume := corev1.Volume{
		Name: "certs",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  "envoy",
				DefaultMode: ptr.To[int32](420),
			},
		},
	}

	if r.GatewayNamespaceMode {
		certsVolume = corev1.Volume{
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.Name(),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  XdsTLSCaFileName,
							Path: XdsTLSCaFileName,
						},
					},
					DefaultMode: ptr.To[int32](420),
					Optional:    ptr.To(false),
				},
			},
		}
		saAudience := fmt.Sprintf("%s.%s.svc.%s", config.EnvoyGatewayServiceName, r.ControllerNamespace(), r.DNSDomain)
		saTokenProjectedVolume := corev1.Volume{
			Name: "sa-token",
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					Sources: []corev1.VolumeProjection{
						{
							ServiceAccountToken: &corev1.ServiceAccountTokenProjection{
								Path:              "sa-token",
								Audience:          saAudience,
								ExpirationSeconds: ptr.To[int64](3600),
							},
						},
					},
					DefaultMode: ptr.To[int32](420),
				},
			},
		}
		volumes = append(volumes, saTokenProjectedVolume)
	}

	volumes = append(volumes, certsVolume)

	sdsVolume := corev1.Volume{
		Name: "sds",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: r.Name(),
				},
				Items:       sdsConfigMapItems(r.GatewayNamespaceMode),
				DefaultMode: ptr.To[int32](420),
				Optional:    ptr.To(false),
			},
		},
	}

	volumes = append(volumes, sdsVolume)
	return resource.ExpectedVolumes(pod, volumes)
}

func sdsConfigMapItems(gatewayNamespaceMode bool) []corev1.KeyToPath {
	if gatewayNamespaceMode {
		return []corev1.KeyToPath{
			{
				Key:  common.SdsCAFilename,
				Path: common.SdsCAFilename,
			},
		}
	}

	return []corev1.KeyToPath{
		{
			Key:  common.SdsCAFilename,
			Path: common.SdsCAFilename,
		},
		{
			Key:  common.SdsCertFilename,
			Path: common.SdsCertFilename,
		},
	}
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
		{
			Name: envoyZoneEnvVar,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  fmt.Sprintf("metadata.annotations['%s']", corev1.LabelTopologyZone),
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

func resolveProxyImage(containerSpec *egv1a1.KubernetesContainerSpec) (string, error) {
	if containerSpec == nil {
		return "", fmt.Errorf("containerSpec is nil")
	}

	repo := ptr.Deref(containerSpec.ImageRepository, "")
	if repo != "" {
		tag, err := getImageTag(egv1a1.DefaultEnvoyProxyImage)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s:%s", repo, tag), nil
	}

	image := ptr.Deref(containerSpec.Image, "")
	if image != "" {
		return image, nil
	}

	return egv1a1.DefaultEnvoyProxyImage, nil
}

// getImageTag parses a Docker/OCI image reference and returns the tag if present.
// Returns an error if parsing fails or if no tag is found.
func getImageTag(image string) (string, error) {
	ref, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return "", fmt.Errorf("failed to parse image reference %q: %w", image, err)
	}

	tagged, ok := ref.(reference.Tagged)
	if !ok {
		return "", fmt.Errorf("no tag found in image reference %q", image)
	}

	return tagged.Tag(), nil
}
