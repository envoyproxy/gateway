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
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

type ResourceRender struct {
	infra *ir.ProxyInfra

	// Namespace is the Namespace used for managed infra.
	Namespace string
}

func NewResourceRender(ns string, infra *ir.ProxyInfra) *ResourceRender {
	return &ResourceRender{
		Namespace: ns,
		infra:     infra,
	}
}

func (r *ResourceRender) Name() string {
	return ExpectedResourceHashedName(r.infra.Name)
}

// ServiceAccount returns the expected proxy serviceAccount.
func (r *ResourceRender) ServiceAccount() (*corev1.ServiceAccount, error) {
	// Set the labels based on the owning gateway name.
	labels := envoyLabels(r.infra.GetProxyMetadata().Labels)
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
			Name:      ExpectedResourceHashedName(r.infra.Name),
			Labels:    labels,
		},
	}, nil
}

// Service returns the expected Service based on the provided infra.
func (r *ResourceRender) Service() (*corev1.Service, error) {
	var ports []corev1.ServicePort
	for _, listener := range r.infra.Listeners {
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
	labels := envoyLabels(r.infra.GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	// Get annotations
	var annotations map[string]string
	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	envoyServiceConfig := provider.GetEnvoyProxyKubeProvider().EnvoyService
	if envoyServiceConfig.Annotations != nil {
		annotations = envoyServiceConfig.Annotations
	}

	// Set the spec of gateway service
	serviceSpec := resource.ExpectedServiceSpec(envoyServiceConfig.Type)
	serviceSpec.Ports = ports
	serviceSpec.Selector = resource.GetSelector(labels).MatchLabels
	serviceSpec.ExternalIPs = r.infra.Addresses

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
			Name:        ExpectedResourceHashedName(r.infra.Name),
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
	labels := envoyLabels(r.infra.GetProxyMetadata().Labels)
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
			Name:      ExpectedResourceHashedName(r.infra.Name),
			Labels:    labels,
		},
		Data: map[string]string{
			SdsCAFilename:   SdsCAConfigMapData,
			SdsCertFilename: SdsCertConfigMapData,
		},
	}, nil
}

// DaemonSet returns the expected DaemonSet based on the provided infra.
func (r *ResourceRender) DaemonSet() (*appsv1.DaemonSet, error) {
	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	if provider.Type != egcfgv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}
	daemonSetConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDaemonSet

	if daemonSetConfig == nil {
		return nil, nil
	}

	objectMeta, err := r.podSetObjectMeta()
	if err != nil {
		return nil, err
	}

	selector := resource.GetSelector(objectMeta.Labels)
	podTemplate, err := r.podSetPodTemplateSpec(selector.MatchLabels, daemonSetConfig.Container, daemonSetConfig.Pod)
	if err != nil {
		return nil, err
	}

	return &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: objectMeta,
		Spec: appsv1.DaemonSetSpec{
			Selector: selector,
			Template: podTemplate,
		},
	}, nil
}

// Deployment returns the expected Deployment based on the provided infra.
func (r *ResourceRender) Deployment() (*appsv1.Deployment, error) {
	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	if provider.Type != egcfgv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}
	deploymentConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDeployment

	if deploymentConfig == nil {
		return nil, nil
	}

	objectMeta, err := r.podSetObjectMeta()
	if err != nil {
		return nil, err
	}

	selector := resource.GetSelector(objectMeta.Labels)
	podTemplate, err := r.podSetPodTemplateSpec(selector.MatchLabels, deploymentConfig.Container, deploymentConfig.Pod)
	if err != nil {
		return nil, err
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: objectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: deploymentConfig.Replicas,
			Selector: selector,
			Template: podTemplate,
		},
	}, nil
}

func (r *ResourceRender) podSetObjectMeta() (metav1.ObjectMeta, error) {
	// Set the labels based on the owning gateway name.
	labels := envoyLabels(r.infra.GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return metav1.ObjectMeta{}, fmt.Errorf("missing owning gateway labels")
	}

	return metav1.ObjectMeta{
		Namespace: r.Namespace,
		Name:      ExpectedResourceHashedName(r.infra.Name),
		Labels:    labels,
	}, nil
}

func (r *ResourceRender) podSetPodTemplateSpec(labels map[string]string, containerSpec *egcfgv1a1.KubernetesContainerSpec, podSpec *egcfgv1a1.KubernetesPodSpec) (corev1.PodTemplateSpec, error) {
	// Get expected bootstrap configurations rendered ProxyContainers
	containers, err := expectedProxyContainers(r.infra, containerSpec)
	if err != nil {
		return corev1.PodTemplateSpec{}, err
	}

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: podSpec.Annotations,
		},
		Spec: corev1.PodSpec{
			Containers:                    containers,
			ServiceAccountName:            ExpectedResourceHashedName(r.infra.Name),
			AutomountServiceAccountToken:  pointer.Bool(false),
			TerminationGracePeriodSeconds: pointer.Int64(int64(300)),
			DNSPolicy:                     corev1.DNSClusterFirst,
			RestartPolicy:                 corev1.RestartPolicyAlways,
			SchedulerName:                 "default-scheduler",
			SecurityContext:               podSpec.SecurityContext,
			Affinity:                      podSpec.Affinity,
			NodeSelector:                  podSpec.NodeSelector,
			Tolerations:                   podSpec.Tolerations,
			Volumes:                       expectedPodSetVolumes(r.infra.Name, podSpec),
		},
	}, nil
}
