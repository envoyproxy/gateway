// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"
	"strconv"

	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
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
	serviceSpec := resource.ExpectedServiceSpec(envoyServiceConfig)
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

// Deployment returns the expected Deployment based on the provided infra.
func (r *ResourceRender) Deployment() (*appsv1.Deployment, error) {
	// Get the EnvoyProxy config to configure the deployment.
	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	if provider.Type != egcfgv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}
	deploymentConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDeployment

	enablePrometheus := false
	if r.infra.Config != nil &&
		r.infra.Config.Spec.Telemetry.Metrics != nil &&
		r.infra.Config.Spec.Telemetry.Metrics.Prometheus != nil {
		enablePrometheus = true
	}

	// Get expected bootstrap configurations rendered ProxyContainers
	containers, err := expectedProxyContainers(r.infra, deploymentConfig)
	if err != nil {
		return nil, err
	}

	// Set the labels based on the owning gateway name.
	labels := r.infra.GetProxyMetadata().Labels
	dpLabels := envoyLabels(labels)
	if len(dpLabels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(dpLabels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	maps.Copy(labels, deploymentConfig.Pod.Labels)
	podLabels := envoyLabels(labels)
	selector := resource.GetSelector(podLabels)

	// Get annotations
	var annotations map[string]string
	if deploymentConfig.Pod.Annotations != nil {
		annotations = deploymentConfig.Pod.Annotations
	}
	if enablePrometheus {
		if annotations == nil {
			annotations = make(map[string]string, 2)
		}
		annotations["prometheus.io/path"] = "/stats/prometheus" // TODO: make this configurable
		annotations["prometheus.io/scrape"] = "true"
		annotations["prometheus.io/port"] = strconv.Itoa(bootstrap.EnvoyReadinessPort)
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      ExpectedResourceHashedName(r.infra.Name),
			Labels:    dpLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: deploymentConfig.Replicas,
			Strategy: *deploymentConfig.Strategy,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      selector.MatchLabels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            ExpectedResourceHashedName(r.infra.Name),
					AutomountServiceAccountToken:  pointer.Bool(false),
					TerminationGracePeriodSeconds: pointer.Int64(int64(300)),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					SecurityContext:               deploymentConfig.Pod.SecurityContext,
					Affinity:                      deploymentConfig.Pod.Affinity,
					Tolerations:                   deploymentConfig.Pod.Tolerations,
					Volumes:                       expectedDeploymentVolumes(r.infra.Name, deploymentConfig),
				},
			},
			RevisionHistoryLimit:    pointer.Int32(10),
			ProgressDeadlineSeconds: pointer.Int32(600),
		},
	}

	return deployment, nil
}
