// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	providerutils "github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

const (
	SdsCAFilename   = "xds-trusted-ca.json"
	SdsCertFilename = "xds-certificate.json"
	// XdsTLSCertFilename is the fully qualified path of the file containing Envoy's
	// xDS server TLS certificate.
	XdsTLSCertFilename = "/certs/tls.crt"
	// XdsTLSKeyFilename is the fully qualified path of the file containing Envoy's
	// xDS server TLS key.
	XdsTLSKeyFilename = "/certs/tls.key"
	// XdsTLSCaFilename is the fully qualified path of the file containing Envoy's
	// trusted CA certificate.
	XdsTLSCaFilename = "/certs/ca.crt"
	// envoyContainerName is the name of the Envoy container.
	envoyContainerName = "envoy"
	// envoyNsEnvVar is the name of the Envoy Gateway namespace environment variable.
	envoyNsEnvVar = "ENVOY_GATEWAY_NAMESPACE"
	// envoyPodEnvVar is the name of the Envoy pod name environment variable.
	envoyPodEnvVar = "ENVOY_POD_NAME"
)

var (
	// xDS certificate rotation is supported by using SDS path-based resource files.
	SdsCAConfigMapData = fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_trusted_ca","validation_context":{"trusted_ca":{"filename":"%s"},`+
		`"match_typed_subject_alt_names":[{"san_type":"DNS","matcher":{"exact":"envoy-gateway"}}]}}]}`, XdsTLSCaFilename)
	SdsCertConfigMapData = fmt.Sprintf(`{"resources":[{"@type":"type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret",`+
		`"name":"xds_certificate","tls_certificate":{"certificate_chain":{"filename":"%s"},`+
		`"private_key":{"filename":"%s"}}}]}`, XdsTLSCertFilename, XdsTLSKeyFilename)
)

// ExpectedResourceHashedName returns expected resource hashed name.
func ExpectedResourceHashedName(name string) string {
	hashedName := providerutils.GetHashedName(name)
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

// expectedProxyContainers returns expected proxy containers.
func expectedProxyContainers(infra *ir.ProxyInfra, deploymentConfig *egv1a1.KubernetesDeploymentSpec) ([]corev1.Container, error) {
	// Define slice to hold container ports
	var ports []corev1.ContainerPort

	// Iterate over listeners and ports to get container ports
	for _, listener := range infra.Listeners {
		for _, p := range listener.Ports {
			var protocol corev1.Protocol
			switch p.Protocol {
			case ir.HTTPProtocolType, ir.HTTPSProtocolType, ir.TLSProtocolType, ir.TCPProtocolType:
				protocol = corev1.ProtocolTCP
			case ir.UDPProtocolType:
				protocol = corev1.ProtocolUDP
			default:
				return nil, fmt.Errorf("invalid protocol %q", p.Protocol)
			}
			port := corev1.ContainerPort{
				Name:          p.Name,
				ContainerPort: p.ContainerPort,
				Protocol:      protocol,
			}
			ports = append(ports, port)
		}
	}

	var proxyMetrics *egv1a1.ProxyMetrics
	if infra.Config != nil {
		proxyMetrics = infra.Config.Spec.Telemetry.Metrics
	}

	if proxyMetrics != nil && proxyMetrics.Prometheus != nil {
		ports = append(ports, corev1.ContainerPort{
			Name:          "metrics",
			ContainerPort: bootstrap.EnvoyReadinessPort, // TODO: make this configurable
			Protocol:      corev1.ProtocolTCP,
		})
	}

	var bootstrapConfigurations string

	// Get the default Bootstrap
	bootstrapConfigurations, err := bootstrap.GetRenderedBootstrapConfig(proxyMetrics)
	if err != nil {
		return nil, err
	}

	// Apply Bootstrap from EnvoyProxy API if set by the user
	// The config should have been validated already
	if infra.Config != nil && infra.Config.Spec.Bootstrap != nil {
		bootstrapConfigurations, err = bootstrap.ApplyBootstrapConfig(infra.Config.Spec.Bootstrap, bootstrapConfigurations)
		if err != nil {
			return nil, err
		}
	}

	logging := infra.Config.Spec.Logging

	args := []string{
		fmt.Sprintf("--service-cluster %s", infra.Name),
		fmt.Sprintf("--service-node $(%s)", envoyPodEnvVar),
		fmt.Sprintf("--config-yaml %s", bootstrapConfigurations),
		fmt.Sprintf("--log-level %s", logging.DefaultEnvoyProxyLoggingLevel()),
		"--cpuset-threads",
	}

	if infra.Config != nil &&
		infra.Config.Spec.Concurrency != nil {
		args = append(args, fmt.Sprintf("--concurrency %d", *infra.Config.Spec.Concurrency))
	}

	if componentsLogLevel := logging.GetEnvoyProxyComponentLevel(); componentsLogLevel != "" {
		args = append(args, fmt.Sprintf("--component-log-level %s", componentsLogLevel))
	}

	containers := []corev1.Container{
		{
			Name:                     envoyContainerName,
			Image:                    *deploymentConfig.Container.Image,
			ImagePullPolicy:          corev1.PullIfNotPresent,
			Command:                  []string{"envoy"},
			Args:                     args,
			Env:                      expectedProxyContainerEnv(deploymentConfig),
			Resources:                *deploymentConfig.Container.Resources,
			SecurityContext:          deploymentConfig.Container.SecurityContext,
			Ports:                    ports,
			VolumeMounts:             expectedContainerVolumeMounts(deploymentConfig),
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
			ReadinessProbe: &corev1.Probe{
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
		},
	}

	return containers, nil
}

// expectedContainerVolumeMounts returns expected proxy container volume mounts.
func expectedContainerVolumeMounts(deploymentSpec *egv1a1.KubernetesDeploymentSpec) []corev1.VolumeMount {
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

	return resource.ExpectedContainerVolumeMounts(deploymentSpec.Container, volumeMounts)
}

// expectedDeploymentVolumes returns expected proxy deployment volumes.
func expectedDeploymentVolumes(name string, deploymentSpec *egv1a1.KubernetesDeploymentSpec) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "envoy",
					DefaultMode: pointer.Int32(420),
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
							Key:  SdsCAFilename,
							Path: SdsCAFilename,
						},
						{
							Key:  SdsCertFilename,
							Path: SdsCertFilename,
						},
					},
					DefaultMode: pointer.Int32(420),
					Optional:    pointer.Bool(false),
				},
			},
		},
	}

	return resource.ExpectedDeploymentVolumes(deploymentSpec.Pod, volumes)
}

// expectedProxyContainerEnv returns expected proxy container envs.
func expectedProxyContainerEnv(deploymentConfig *egv1a1.KubernetesDeploymentSpec) []corev1.EnvVar {
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

	return resource.ExpectedProxyContainerEnv(deploymentConfig.Container, env)
}
