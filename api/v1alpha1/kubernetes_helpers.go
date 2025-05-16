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
