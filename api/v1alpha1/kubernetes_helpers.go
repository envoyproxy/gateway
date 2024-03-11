// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/utils/ptr"
)

// DefaultKubernetesDeploymentStrategy returns the default deployment strategy settings.
func DefaultKubernetesDeploymentStrategy() *appv1.DeploymentStrategy {
	return &appv1.DeploymentStrategy{
		Type: appv1.RollingUpdateDeploymentStrategyType,
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

// setDefault fill a default HorizontalPodAutoscalerSpec if unspecified
func (hpa *KubernetesHorizontalPodAutoscalerSpec) setDefault() {
	if len(hpa.Metrics) == 0 {
		hpa.Metrics = DefaultEnvoyProxyHpaMetrics()
	}
}

// ApplyMergePatch applies a merge patch to a deployment based on the merge type
func (deployment *KubernetesDeploymentSpec) ApplyMergePatch(old *appv1.Deployment) (*appv1.Deployment, error) {
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
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, deployment.Patch.Value.Raw, appv1.Deployment{})
	case *deployment.Patch.Type == JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, deployment.Patch.Value.Raw)
	default:
		return nil, fmt.Errorf("unsupported merge type: %s", *deployment.Patch.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("error applying merge patch: %w", err)
	}

	// Deserialize the patched JSON into a new deployment object
	var patchedDeployment appv1.Deployment
	if err := json.Unmarshal(patchedJSON, &patchedDeployment); err != nil {
		return nil, fmt.Errorf("error unmarshaling patched deployment: %w", err)
	}

	return &patchedDeployment, nil
}

// ApplyMergePatch applies a merge patch to a service based on the merge type
func (service *KubernetesServiceSpec) ApplyMergePatch(old *corev1.Service) (*corev1.Service, error) {
	if service.Patch == nil {
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
