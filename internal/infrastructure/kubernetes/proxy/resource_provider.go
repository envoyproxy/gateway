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
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
	if OwningGatewayLabelsAbsent(labels) {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
			Name:        r.Name(),
			Labels:      labels,
			Annotations: r.infra.GetProxyMetadata().Annotations,
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
				Name:       ExpectedResourceHashedName(port.Name),
				Protocol:   protocol,
				Port:       port.ServicePort,
				TargetPort: target,
			}
			ports = append(ports, p)

			if port.Protocol == ir.HTTPSProtocolType {
				if listener.HTTP3 != nil {
					p := corev1.ServicePort{
						Name:       ExpectedResourceHashedName(port.Name + "-h3"),
						Protocol:   corev1.ProtocolUDP,
						Port:       port.ServicePort,
						TargetPort: target,
					}
					ports = append(ports, p)
				}
			}
		}
	}

	// Set the labels based on the owning gatewayclass name.
	labels := envoyLabels(r.infra.GetProxyMetadata().Labels)
	if OwningGatewayLabelsAbsent(labels) {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	// Get annotations
	annotations := map[string]string{}
	maps.Copy(annotations, r.infra.GetProxyMetadata().Annotations)

	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	envoyServiceConfig := provider.GetEnvoyProxyKubeProvider().EnvoyService
	if envoyServiceConfig.Annotations != nil {
		maps.Copy(annotations, envoyServiceConfig.Annotations)
	}
	if len(annotations) == 0 {
		annotations = nil
	}

	// Set the spec of gateway service
	serviceSpec := resource.ExpectedServiceSpec(envoyServiceConfig)
	serviceSpec.Ports = ports
	serviceSpec.Selector = resource.GetSelector(labels).MatchLabels

	if (*envoyServiceConfig.Type) == egv1a1.ServiceTypeClusterIP {
		if len(r.infra.Addresses) > 0 {
			// Since K8s Service requires specify no more than one IP for each IP family
			// So we only use the first address
			// if address is not set, the automatically assigned clusterIP is used
			serviceSpec.ClusterIP = r.infra.Addresses[0]
			serviceSpec.ClusterIPs = r.infra.Addresses[0:1]
		}
	} else {
		serviceSpec.ExternalIPs = r.infra.Addresses
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
			Name:        r.Name(),
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: serviceSpec,
	}

	// apply merge patch to service
	var err error
	if svc, err = envoyServiceConfig.ApplyMergePatch(svc); err != nil {
		return nil, err
	}

	return svc, nil
}

// ConfigMap returns the expected ConfigMap based on the provided infra.
func (r *ResourceRender) ConfigMap() (*corev1.ConfigMap, error) {
	// Set the labels based on the owning gateway name.
	labels := envoyLabels(r.infra.GetProxyMetadata().Labels)
	if OwningGatewayLabelsAbsent(labels) {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
			Name:        r.Name(),
			Labels:      labels,
			Annotations: r.infra.GetProxyMetadata().Annotations,
		},
		Data: map[string]string{
			SdsCAFilename:   SdsCAConfigMapData,
			SdsCertFilename: SdsCertConfigMapData,
		},
	}, nil
}

// Deployment returns the expected Deployment based on the provided infra.
func (r *ResourceRender) Deployment() (*appsv1.Deployment, error) {
	proxyConfig := r.infra.GetProxyConfig()

	// Get the EnvoyProxy config to configure the deployment.
	provider := proxyConfig.GetEnvoyProxyProvider()
	if provider.Type != egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}
	deploymentConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDeployment

	// Get expected bootstrap configurations rendered ProxyContainers
	containers, err := expectedProxyContainers(r.infra, deploymentConfig, proxyConfig.Spec.Shutdown)
	if err != nil {
		return nil, err
	}

	// Set the labels based on the owning gateway name.
	dpAnnotations := r.infra.GetProxyMetadata().Annotations
	labels := r.infra.GetProxyMetadata().Labels
	dpLabels := envoyLabels(labels)
	if OwningGatewayLabelsAbsent(dpLabels) {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	maps.Copy(labels, deploymentConfig.Pod.Labels)
	podLabels := envoyLabels(labels)
	selector := resource.GetSelector(podLabels)

	// Get annotations
	podAnnotations := map[string]string{}
	maps.Copy(podAnnotations, dpAnnotations)
	maps.Copy(podAnnotations, deploymentConfig.Pod.Annotations)
	if enablePrometheus(r.infra) {
		podAnnotations["prometheus.io/path"] = "/stats/prometheus" // TODO: make this configurable
		podAnnotations["prometheus.io/scrape"] = "true"
		podAnnotations["prometheus.io/port"] = strconv.Itoa(bootstrap.EnvoyReadinessPort)
	}
	if len(podAnnotations) == 0 {
		podAnnotations = nil
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
			Name:        r.Name(),
			Labels:      dpLabels,
			Annotations: dpAnnotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: deploymentConfig.Replicas,
			Strategy: *deploymentConfig.Strategy,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      selector.MatchLabels,
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					InitContainers:                deploymentConfig.InitContainers,
					ServiceAccountName:            ExpectedResourceHashedName(r.infra.Name),
					AutomountServiceAccountToken:  ptr.To(false),
					TerminationGracePeriodSeconds: expectedTerminationGracePeriodSeconds(proxyConfig.Spec.Shutdown),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					SecurityContext:               deploymentConfig.Pod.SecurityContext,
					Affinity:                      deploymentConfig.Pod.Affinity,
					Tolerations:                   deploymentConfig.Pod.Tolerations,
					Volumes:                       expectedDeploymentVolumes(r.infra.Name, deploymentConfig),
					ImagePullSecrets:              deploymentConfig.Pod.ImagePullSecrets,
					NodeSelector:                  deploymentConfig.Pod.NodeSelector,
					TopologySpreadConstraints:     deploymentConfig.Pod.TopologySpreadConstraints,
				},
			},
			RevisionHistoryLimit:    ptr.To[int32](10),
			ProgressDeadlineSeconds: ptr.To[int32](600),
		},
	}

	// omit the deployment replicas if HPA is being set
	if provider.GetEnvoyProxyKubeProvider().EnvoyHpa != nil {
		deployment.Spec.Replicas = nil
	}

	// apply merge patch to deployment
	if deployment, err = deploymentConfig.ApplyMergePatch(deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

func expectedTerminationGracePeriodSeconds(cfg *egv1a1.ShutdownConfig) *int64 {
	s := 900 // default
	if cfg != nil && cfg.DrainTimeout != nil {
		s = int(cfg.DrainTimeout.Seconds() + 300) // 5 minutes longer than drain timeout
	}
	return ptr.To(int64(s))
}

func (r *ResourceRender) HorizontalPodAutoscaler() (*autoscalingv2.HorizontalPodAutoscaler, error) {
	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	if provider.Type != egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}

	hpaConfig := provider.GetEnvoyProxyKubeProvider().EnvoyHpa
	if hpaConfig == nil {
		return nil, nil
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
			Name:        r.Name(),
			Annotations: r.infra.GetProxyMetadata().Annotations,
			Labels:      r.infra.GetProxyMetadata().Labels,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       r.Name(),
			},
			MinReplicas: hpaConfig.MinReplicas,
			MaxReplicas: ptr.Deref(hpaConfig.MaxReplicas, 1),
			Metrics:     hpaConfig.Metrics,
			Behavior:    hpaConfig.Behavior,
		},
	}

	return hpa, nil
}

// OwningGatewayLabelsAbsent Check if labels are missing some OwningGatewayLabels
func OwningGatewayLabelsAbsent(labels map[string]string) bool {
	return (len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 ||
		len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0) &&
		len(labels[gatewayapi.OwningGatewayClassLabel]) == 0
}
