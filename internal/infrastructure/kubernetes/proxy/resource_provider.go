// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/utils"
	"github.com/envoyproxy/gateway/internal/ir"
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

type ResourceRender struct {
	infra *ir.Infra

	// Namespace is the Namespace used for managed infra.
	Namespace string
}

func NewResourceRender(ns string, infra *ir.Infra) *ResourceRender {
	return &ResourceRender{
		Namespace: ns,
		infra:     infra,
	}
}

func (r *ResourceRender) Name() string {
	return utils.ExpectedResourceHashedName(r.infra.Proxy.Name)
}

// ServiceAccount returns the expected proxy serviceAccount.
func (r *ResourceRender) ServiceAccount() (*corev1.ServiceAccount, error) {
	// Set the labels based on the owning gateway name.
	labels := utils.EnvoyLabels(r.infra.GetProxyInfra().GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      utils.ExpectedResourceHashedName(r.infra.Proxy.Name),
			Labels:    labels,
		},
	}, nil
}

// Service returns the expected Service based on the provided infra.
func (r *ResourceRender) Service() (*corev1.Service, error) {
	var ports []corev1.ServicePort
	for _, listener := range r.infra.Proxy.Listeners {
		for _, port := range listener.Ports {
			target := intstr.IntOrString{IntVal: port.ContainerPort}
			protocol := corev1.ProtocolTCP
			if port.Protocol == ir.UDPProtocolType {
				protocol = corev1.ProtocolUDP
			}
			p := corev1.ServicePort{
				Name:       port.Name,
				Protocol:   protocol,
				Port:       port.ServicePort,
				TargetPort: target,
			}
			ports = append(ports, p)
		}
	}

	// Set the labels based on the owning gatewayclass name.
	labels := utils.EnvoyLabels(r.infra.GetProxyInfra().GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	// Get annotations
	var annotations map[string]string
	provider := r.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider()
	envoyServiceConfig := provider.GetEnvoyProxyKubeProvider().EnvoyService
	if envoyServiceConfig.Annotations != nil {
		annotations = envoyServiceConfig.Annotations
	}
	serviceSpec := utils.ExpectedServiceSpec(envoyServiceConfig.Type)
	serviceSpec.Ports = ports
	serviceSpec.Selector = utils.GetSelector(labels).MatchLabels

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
			Name:        utils.ExpectedResourceHashedName(r.infra.Proxy.Name),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: serviceSpec,
	}

	return svc, nil
}

// ConfigMap returns the expected ConfigMap based on the provided infra.
func (r *ResourceRender) ConfigMap() (*corev1.ConfigMap, error) {
	// Set the labels based on the owning gateway name.
	labels := utils.EnvoyLabels(r.infra.GetProxyInfra().GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      utils.ExpectedResourceHashedName(r.infra.Proxy.Name),
			Labels:    labels,
		},
		Data: map[string]string{
			utils.SdsCAFilename:   utils.SdsCAConfigMapData,
			utils.SdsCertFilename: utils.SdsCertConfigMapData,
		},
	}, nil
}

// ExpectedDeployment returns the expected Deployment based on the provided infra.
func (r *ResourceRender) Deployment() (*appsv1.Deployment, error) {
	// Get the EnvoyProxy config to configure the deployment.
	provider := r.infra.GetProxyInfra().GetProxyConfig().GetEnvoyProxyProvider()
	if provider.Type != egcfgv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}
	deploymentConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDeployment

	// Get expected bootstrap configurations rendered ProxyContainers
	containers, err := expectedProxyContainers(r.infra, deploymentConfig)
	if err != nil {
		return nil, err
	}

	// Set the labels based on the owning gateway name.
	labels := utils.EnvoyLabels(r.infra.GetProxyInfra().GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	selector := utils.GetSelector(labels)

	// Get annotations
	var annotations map[string]string
	if deploymentConfig.Pod.Annotations != nil {
		annotations = deploymentConfig.Pod.Annotations
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      utils.ExpectedResourceHashedName(r.infra.Proxy.Name),
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: deploymentConfig.Replicas,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      selector.MatchLabels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            utils.ExpectedResourceHashedName(r.infra.Proxy.Name),
					AutomountServiceAccountToken:  pointer.Bool(false),
					TerminationGracePeriodSeconds: pointer.Int64(int64(300)),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					SecurityContext:               deploymentConfig.Pod.SecurityContext,
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "envoy",
								},
							},
						},
						{
							Name: "sds",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: utils.ExpectedResourceHashedName(r.infra.Proxy.Name),
									},
									Items: []corev1.KeyToPath{
										{
											Key:  utils.SdsCAFilename,
											Path: utils.SdsCAFilename,
										},
										{
											Key:  utils.SdsCertFilename,
											Path: utils.SdsCertFilename,
										},
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

func expectedProxyContainers(infra *ir.Infra, deploymentConfig *egcfgv1a1.KubernetesDeploymentSpec) ([]corev1.Container, error) {
	// Define slice to hold container ports
	var ports []corev1.ContainerPort

	// Iterate over listeners and ports to get container ports
	for _, listener := range infra.Proxy.Listeners {
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

	var bootstrapConfigurations string
	// Get Bootstrap from EnvoyProxy API if set by the user
	// The config should have been validated already
	if infra.Proxy.Config != nil &&
		infra.Proxy.Config.Spec.Bootstrap != nil {
		bootstrapConfigurations = *infra.Proxy.Config.Spec.Bootstrap
	} else {
		var err error
		// Use the default Bootstrap
		bootstrapConfigurations, err = bootstrap.GetRenderedBootstrapConfig()
		if err != nil {
			return nil, err
		}
	}

	containers := []corev1.Container{
		{
			Name:            envoyContainerName,
			Image:           *deploymentConfig.Container.Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{
				"envoy",
			},
			Args: []string{
				fmt.Sprintf("--service-cluster %s", infra.Proxy.Name),
				fmt.Sprintf("--service-node $(%s)", envoyPodEnvVar),
				fmt.Sprintf("--config-yaml %s", bootstrapConfigurations),
				"--log-level info",
			},
			Env: []corev1.EnvVar{
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
			},
			Resources:       *deploymentConfig.Container.Resources,
			SecurityContext: deploymentConfig.Container.SecurityContext,
			Ports:           ports,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "certs",
					MountPath: "/certs",
					ReadOnly:  true,
				},
				{
					Name:      "sds",
					MountPath: "/sds",
				},
			},
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
		},
	}

	return containers, nil
}
