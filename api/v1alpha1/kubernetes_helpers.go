// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/utils/ptr"
)

// DefaultKubernetesDeploymentStrategy returns the default deployment strategy settings.
func DefaultKubernetesDeploymentStrategy() *appsv1.DeploymentStrategy {
	return &appsv1.DeploymentStrategy{
		Type: appsv1.RollingUpdateDeploymentStrategyType,
	}
}

// DefaultKubernetesDaemonSetStrategy returns the default daemonset strategy settings.
func DefaultKubernetesDaemonSetStrategy() *appsv1.DaemonSetUpdateStrategy {
	return &appsv1.DaemonSetUpdateStrategy{
		Type: appsv1.RollingUpdateDaemonSetStrategyType,
	}
}

// DefaultKubernetesContainerImage returns the default envoyproxy image.
func DefaultKubernetesContainerImage(image string) *string {
	return ptr.To(image)
}

// DefaultKubernetesDeployment returns a new KubernetesDeploymentSpec with default settings.
func DefaultKubernetesDeployment(image string) *KubernetesDeploymentSpec {
	return &KubernetesDeploymentSpec{
		Strategy:  DefaultKubernetesDeploymentStrategy(),
		Pod:       DefaultKubernetesPod(),
		Container: DefaultKubernetesContainer(image),
	}
}

// DefaultKubernetesDaemonSet returns a new DefaultKubernetesDaemonSet with default settings.
func DefaultKubernetesDaemonSet(image string) *KubernetesDaemonSetSpec {
	return &KubernetesDaemonSetSpec{
		Strategy:  DefaultKubernetesDaemonSetStrategy(),
		Pod:       DefaultKubernetesPod(),
		Container: DefaultKubernetesContainer(image),
	}
}

// DefaultKubernetesPod returns a new KubernetesPodSpec with default settings.
func DefaultKubernetesPod() *KubernetesPodSpec {
	return &KubernetesPodSpec{}
}

// DefaultKubernetesContainer returns a new KubernetesContainerSpec with default settings.
func DefaultKubernetesContainer(image string) *KubernetesContainerSpec {
	return &KubernetesContainerSpec{
		Resources: DefaultResourceRequirements(),
		Image:     DefaultKubernetesContainerImage(image),
	}
}

// DefaultKubernetesContainer returns a new KubernetesContainerSpec with default settings.
func DefaultKubernetesInitContainer(image string) *KubernetesContainerSpec {
	return &KubernetesContainerSpec{
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(DefaultInitContainerCPUResourceRequests),
				corev1.ResourceMemory: resource.MustParse(DefaultInitContainerMemoryResourceRequests),
			},
		},
		Image: DefaultKubernetesContainerImage(image),
	}
}

// DefaultResourceRequirements returns a new ResourceRequirements with default settings.
func DefaultResourceRequirements() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(DefaultDeploymentCPUResourceRequests),
			corev1.ResourceMemory: resource.MustParse(DefaultDeploymentMemoryResourceRequests),
		},
	}
}

// DefaultKubernetesService returns a new KubernetesServiceSpec with default settings.
func DefaultKubernetesService() *KubernetesServiceSpec {
	return &KubernetesServiceSpec{
		Type:                  DefaultKubernetesServiceType(),
		ExternalTrafficPolicy: DefaultKubernetesServiceExternalTrafficPolicy(),
	}
}

// DefaultKubernetesServiceType returns a new KubernetesServiceType with default settings.
func DefaultKubernetesServiceType() *ServiceType {
	return GetKubernetesServiceType(ServiceTypeLoadBalancer)
}

// GetKubernetesServiceType returns the KubernetesServiceType pointer.
func GetKubernetesServiceType(serviceType ServiceType) *ServiceType {
	return &serviceType
}

func DefaultKubernetesServiceExternalTrafficPolicy() *ServiceExternalTrafficPolicy {
	return GetKubernetesServiceExternalTrafficPolicy(ServiceExternalTrafficPolicyLocal)
}

func GetKubernetesServiceExternalTrafficPolicy(serviceExternalTrafficPolicy ServiceExternalTrafficPolicy) *ServiceExternalTrafficPolicy {
	return &serviceExternalTrafficPolicy
}

// defaultKubernetesDeploymentSpec fill a default KubernetesDeploymentSpec if unspecified.
func (deployment *KubernetesDeploymentSpec) defaultKubernetesDeploymentSpec(image string) {
	if deployment.Strategy == nil {
		deployment.Strategy = DefaultKubernetesDeploymentStrategy()
	}

	if deployment.Pod == nil {
		deployment.Pod = DefaultKubernetesPod()
	}

	if deployment.Container == nil {
		deployment.Container = DefaultKubernetesContainer(image)
	}

	if deployment.Container.Resources == nil {
		deployment.Container.Resources = DefaultResourceRequirements()
	}

	if deployment.Container.Image == nil {
		deployment.Container.Image = DefaultKubernetesContainerImage(image)
	}
}

// defaultKubernetesDaemonSetSpec fill a default KubernetesDaemonSetSpec if unspecified.
func (daemonset *KubernetesDaemonSetSpec) defaultKubernetesDaemonSetSpec(image string) {
	if daemonset.Strategy == nil {
		daemonset.Strategy = DefaultKubernetesDaemonSetStrategy()
	}

	if daemonset.Pod == nil {
		daemonset.Pod = DefaultKubernetesPod()
	}

	if daemonset.Container == nil {
		daemonset.Container = DefaultKubernetesContainer(image)
	}

	if daemonset.Container.Resources == nil {
		daemonset.Container.Resources = DefaultResourceRequirements()
	}

	if daemonset.Container.Image == nil {
		daemonset.Container.Image = DefaultKubernetesContainerImage(image)
	}
}

// setDefault fill a default HorizontalPodAutoscalerSpec if unspecified
func (hpa *KubernetesHorizontalPodAutoscalerSpec) setDefault() {
	if len(hpa.Metrics) == 0 {
		hpa.Metrics = DefaultEnvoyProxyHpaMetrics()
	}
}

// ApplyMergePatch applies a merge patch to a deployment based on the merge type
func (deployment *KubernetesDeploymentSpec) ApplyMergePatch(old *appsv1.Deployment) (*appsv1.Deployment, error) {
	if deployment.Patch == nil {
		return old, nil
	}

	var patchedJSON []byte
	var err error

	// Serialize the current deployment to JSON
	originalJSON, err := json.Marshal(old)
	if err != nil {
		return nil, fmt.Errorf("error marshaling original deployment: %w", err)
	}

	switch {
	case deployment.Patch.Type == nil || *deployment.Patch.Type == StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, deployment.Patch.Value.Raw, appsv1.Deployment{})
	case *deployment.Patch.Type == JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, deployment.Patch.Value.Raw)
	default:
		return nil, fmt.Errorf("unsupported merge type: %s", *deployment.Patch.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("error applying merge patch: %w", err)
	}

	// Deserialize the patched JSON into a new deployment object
	var patchedDeployment appsv1.Deployment
	if err := json.Unmarshal(patchedJSON, &patchedDeployment); err != nil {
		return nil, fmt.Errorf("error unmarshaling patched deployment: %w", err)
	}

	return &patchedDeployment, nil
}

// ApplyMergePatch applies a merge patch to a daemonset based on the merge type
func (daemonset *KubernetesDaemonSetSpec) ApplyMergePatch(old *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	if daemonset.Patch == nil {
		return old, nil
	}

	var patchedJSON []byte
	var err error

	// Serialize the current daemonset to JSON
	originalJSON, err := json.Marshal(old)
	if err != nil {
		return nil, fmt.Errorf("error marshaling original daemonset: %w", err)
	}

	switch {
	case daemonset.Patch.Type == nil || *daemonset.Patch.Type == StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, daemonset.Patch.Value.Raw, appsv1.DaemonSet{})
	case *daemonset.Patch.Type == JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, daemonset.Patch.Value.Raw)
	default:
		return nil, fmt.Errorf("unsupported merge type: %s", *daemonset.Patch.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("error applying merge patch: %w", err)
	}

	// Deserialize the patched JSON into a new daemonset object
	var patchedDaemonSet appsv1.DaemonSet
	if err := json.Unmarshal(patchedJSON, &patchedDaemonSet); err != nil {
		return nil, fmt.Errorf("error unmarshaling patched daemonset: %w", err)
	}

	return &patchedDaemonSet, nil
}

// ApplyMergePatch applies a merge patch to a service based on the merge type
func (service *KubernetesServiceSpec) ApplyMergePatch(old *corev1.Service) (*corev1.Service, error) {
	if service.Patch == nil {
		return old, nil
	}

	var patchedJSON []byte
	var err error

	// Serialize the current service to JSON
	originalJSON, err := json.Marshal(old)
	if err != nil {
		return nil, fmt.Errorf("error marshaling original service: %w", err)
	}

	switch {
	case service.Patch.Type == nil || *service.Patch.Type == StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, service.Patch.Value.Raw, corev1.Service{})
	case *service.Patch.Type == JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, service.Patch.Value.Raw)
	default:
		return nil, fmt.Errorf("unsupported merge type: %s", *service.Patch.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("error applying merge patch: %w", err)
	}

	// Deserialize the patched JSON into a new service object
	var patchedService corev1.Service
	if err := json.Unmarshal(patchedJSON, &patchedService); err != nil {
		return nil, fmt.Errorf("error unmarshaling patched service: %w", err)
	}

	return &patchedService, nil
}

// ApplyMergePatch applies a merge patch to a HorizontalPodAutoscaler based on the merge type
func (hpa *KubernetesHorizontalPodAutoscalerSpec) ApplyMergePatch(old *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	if hpa.Patch == nil {
		return old, nil
	}

	var patchedJSON []byte
	var err error

	// Serialize the current HPA to JSON
	originalJSON, err := json.Marshal(old)
	if err != nil {
		return nil, fmt.Errorf("error marshaling original HorizontalPodAutoscaler: %w", err)
	}

	switch {
	case hpa.Patch.Type == nil || *hpa.Patch.Type == StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, hpa.Patch.Value.Raw, autoscalingv2.HorizontalPodAutoscaler{})
	case *hpa.Patch.Type == JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, hpa.Patch.Value.Raw)
	default:
		return nil, fmt.Errorf("unsupported merge type: %s", *hpa.Patch.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("error applying merge patch: %w", err)
	}

	// Deserialize the patched JSON into a new HorizontalPodAutoscaler object
	var patchedHpa autoscalingv2.HorizontalPodAutoscaler
	if err := json.Unmarshal(patchedJSON, &patchedHpa); err != nil {
		return nil, fmt.Errorf("error unmarshaling patched HorizontalPodAutoscaler: %w", err)
	}

	return &patchedHpa, nil
}

// ApplyMergePatch applies a merge patch to a PodDisruptionBudget based on the merge type
func (pdb *KubernetesPodDisruptionBudgetSpec) ApplyMergePatch(old *policyv1.PodDisruptionBudget) (*policyv1.PodDisruptionBudget, error) {
	if pdb.Patch == nil {
		return old, nil
	}

	var patchedJSON []byte
	var err error

	// Serialize the PDB deployment to JSON
	originalJSON, err := json.Marshal(old)
	if err != nil {
		return nil, fmt.Errorf("error marshaling original PodDisruptionBudget: %w", err)
	}

	switch {
	case pdb.Patch.Type == nil || *pdb.Patch.Type == StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, pdb.Patch.Value.Raw, policyv1.PodDisruptionBudget{})
	case *pdb.Patch.Type == JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, pdb.Patch.Value.Raw)
	default:
		return nil, fmt.Errorf("unsupported merge type: %s", *pdb.Patch.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("error applying merge patch: %w", err)
	}

	// Deserialize the patched JSON into a new HorizontalPodAutoscaler object
	var patchedPdb policyv1.PodDisruptionBudget
	if err := json.Unmarshal(patchedJSON, &patchedPdb); err != nil {
		return nil, fmt.Errorf("error unmarshaling patched PodDisruptionBudget: %w", err)
	}

	return &patchedPdb, nil
}
