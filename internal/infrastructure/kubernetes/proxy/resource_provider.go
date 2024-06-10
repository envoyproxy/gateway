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
	v1 "k8s.io/api/policy/v1"
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

	ShutdownManager *egv1a1.ShutdownManager
}

func NewResourceRender(ns string, infra *ir.ProxyInfra, gateway *egv1a1.EnvoyGateway) *ResourceRender {
	return &ResourceRender{
		Namespace:       ns,
		infra:           infra,
		ShutdownManager: gateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().ShutdownManager,
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
				Name:       port.Name,
				Protocol:   protocol,
				Port:       port.ServicePort,
				TargetPort: target,
			}
			ports = append(ports, p)

			if port.Protocol == ir.HTTPSProtocolType {
				if listener.HTTP3 != nil {
					p := corev1.ServicePort{
						Name:       port.Name + "-h3",
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
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: serviceSpec,
	}

	// set name
	if envoyServiceConfig.Name != nil {
		svc.ObjectMeta.Name = *envoyServiceConfig.Name
	} else {
		svc.ObjectMeta.Name = r.Name()
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

	// If deployment config is nil, it's not Deployment installation.
	if deploymentConfig == nil {
		return nil, nil
	}

	// Get expected bootstrap configurations rendered ProxyContainers
	containers, err := expectedProxyContainers(r.infra, deploymentConfig.Container, proxyConfig.Spec.Shutdown, r.ShutdownManager)
	if err != nil {
		return nil, err
	}

	dpAnnotations := r.infra.GetProxyMetadata().Annotations
	podAnnotations := r.getPodAnnotations(dpAnnotations, deploymentConfig.Pod)

	// Set the labels based on the owning gateway name.
	dpLabels, err := r.getLabels()
	if err != nil {
		return nil, err
	}
	podLabels := r.getPodLabels(deploymentConfig.Pod)
	selector := resource.GetSelector(podLabels)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
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
					Volumes:                       expectedVolumes(r.infra.Name, deploymentConfig.Pod),
					ImagePullSecrets:              deploymentConfig.Pod.ImagePullSecrets,
					NodeSelector:                  deploymentConfig.Pod.NodeSelector,
					TopologySpreadConstraints:     deploymentConfig.Pod.TopologySpreadConstraints,
				},
			},
			RevisionHistoryLimit:    ptr.To[int32](10),
			ProgressDeadlineSeconds: ptr.To[int32](600),
		},
	}

	// set name
	if deploymentConfig.Name != nil {
		deployment.ObjectMeta.Name = *deploymentConfig.Name
	} else {
		deployment.ObjectMeta.Name = r.Name()
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

func (r *ResourceRender) DaemonSet() (*appsv1.DaemonSet, error) {
	proxyConfig := r.infra.GetProxyConfig()

	// Get the EnvoyProxy config to configure the daemonset.
	provider := proxyConfig.GetEnvoyProxyProvider()
	if provider.Type != egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}

	daemonSetConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDaemonSet

	// If daemonset config is nil, it's not DaemonSet installation.
	if daemonSetConfig == nil {
		return nil, nil
	}

	// Get expected bootstrap configurations rendered ProxyContainers
	containers, err := expectedProxyContainers(r.infra, daemonSetConfig.Container, proxyConfig.Spec.Shutdown, r.ShutdownManager)
	if err != nil {
		return nil, err
	}

	dsAnnotations := r.infra.GetProxyMetadata().Annotations
	podAnnotations := r.getPodAnnotations(dsAnnotations, daemonSetConfig.Pod)

	// Set the labels based on the owning gateway name.
	dsLabels, err := r.getLabels()
	if err != nil {
		return nil, err
	}
	podLabels := r.getPodLabels(daemonSetConfig.Pod)
	selector := resource.GetSelector(podLabels)

	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   r.Namespace,
			Labels:      dsLabels,
			Annotations: dsAnnotations,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector:       selector,
			UpdateStrategy: *daemonSetConfig.Strategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      selector.MatchLabels,
					Annotations: podAnnotations,
				},
				Spec: r.getPodSpec(containers, nil, daemonSetConfig.Pod, proxyConfig),
			},
		},
	}

	// set name
	if daemonSetConfig.Name != nil {
		daemonSet.ObjectMeta.Name = *daemonSetConfig.Name
	} else {
		daemonSet.ObjectMeta.Name = r.Name()
	}

	// apply merge patch to daemonset
	if daemonSet, err = daemonSetConfig.ApplyMergePatch(daemonSet); err != nil {
		return nil, err
	}

	return daemonSet, nil
}

func (r *ResourceRender) PodDisruptionBudget() (*v1.PodDisruptionBudget, error) {
	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	if provider.Type != egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}

	podDisruptionBudget := provider.GetEnvoyProxyKubeProvider().EnvoyPDB
	if podDisruptionBudget == nil || podDisruptionBudget.MinAvailable == nil {
		return nil, nil
	}

	labels, err := r.getLabels()
	if err != nil {
		return nil, err
	}

	return &v1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name(),
			Namespace: r.Namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1",
			Kind:       "PodDisruptionBudget",
		},
		Spec: v1.PodDisruptionBudgetSpec{
			MinAvailable: podDisruptionBudget.MinAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
		},
		Status: v1.PodDisruptionBudgetStatus{},
	}, nil
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
			},
			MinReplicas: hpaConfig.MinReplicas,
			MaxReplicas: ptr.Deref(hpaConfig.MaxReplicas, 1),
			Metrics:     hpaConfig.Metrics,
			Behavior:    hpaConfig.Behavior,
		},
	}

	// set deployment target ref name
	deploymentConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDeployment
	if deploymentConfig.Name != nil {
		hpa.Spec.ScaleTargetRef.Name = *deploymentConfig.Name
	} else {
		hpa.Spec.ScaleTargetRef.Name = r.Name()
	}

	return hpa, nil
}

func expectedTerminationGracePeriodSeconds(cfg *egv1a1.ShutdownConfig) *int64 {
	s := 900 // default
	if cfg != nil && cfg.DrainTimeout != nil {
		s = int(cfg.DrainTimeout.Seconds() + 300) // 5 minutes longer than drain timeout
	}
	return ptr.To(int64(s))
}

func (r *ResourceRender) getPodSpec(
	containers, initContainers []corev1.Container,
	pod *egv1a1.KubernetesPodSpec,
	proxyConfig *egv1a1.EnvoyProxy,
) corev1.PodSpec {
	return corev1.PodSpec{
		Containers:                    containers,
		InitContainers:                initContainers,
		ServiceAccountName:            ExpectedResourceHashedName(r.infra.Name),
		AutomountServiceAccountToken:  ptr.To(false),
		TerminationGracePeriodSeconds: expectedTerminationGracePeriodSeconds(proxyConfig.Spec.Shutdown),
		DNSPolicy:                     corev1.DNSClusterFirst,
		RestartPolicy:                 corev1.RestartPolicyAlways,
		SchedulerName:                 "default-scheduler",
		SecurityContext:               pod.SecurityContext,
		Affinity:                      pod.Affinity,
		Tolerations:                   pod.Tolerations,
		Volumes:                       expectedVolumes(r.infra.Name, pod),
		ImagePullSecrets:              pod.ImagePullSecrets,
		NodeSelector:                  pod.NodeSelector,
		TopologySpreadConstraints:     pod.TopologySpreadConstraints,
	}
}

func (r *ResourceRender) getPodAnnotations(resourceAnnotation map[string]string, pod *egv1a1.KubernetesPodSpec) map[string]string {
	podAnnotations := map[string]string{}
	maps.Copy(podAnnotations, resourceAnnotation)
	maps.Copy(podAnnotations, pod.Annotations)

	if enablePrometheus(r.infra) {
		podAnnotations["prometheus.io/path"] = "/stats/prometheus" // TODO: make this configurable
		podAnnotations["prometheus.io/scrape"] = "true"
		podAnnotations["prometheus.io/port"] = strconv.Itoa(bootstrap.EnvoyReadinessPort)
	}

	if len(podAnnotations) == 0 {
		podAnnotations = nil
	}

	return podAnnotations
}

func (r *ResourceRender) getLabels() (map[string]string, error) {
	// Set the labels based on the owning gateway name.
	resourceLabels := envoyLabels(r.infra.GetProxyMetadata().Labels)
	if OwningGatewayLabelsAbsent(resourceLabels) {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return resourceLabels, nil
}

func (r *ResourceRender) getPodLabels(pod *egv1a1.KubernetesPodSpec) map[string]string {
	labels := r.infra.GetProxyMetadata().Labels
	maps.Copy(labels, pod.Labels)

	return envoyLabels(labels)
}

// OwningGatewayLabelsAbsent Check if labels are missing some OwningGatewayLabels
func OwningGatewayLabelsAbsent(labels map[string]string) bool {
	return (len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 ||
		len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0) &&
		len(labels[gatewayapi.OwningGatewayClassLabel]) == 0
}
