// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package apihelpers

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
	return new(image)
}

// DefaultKubernetesDeployment returns a new KubernetesDeploymentSpec with default settings.
func DefaultKubernetesDeployment(image string) *egv1a1.KubernetesDeploymentSpec {
	return &egv1a1.KubernetesDeploymentSpec{
		Strategy:  DefaultKubernetesDeploymentStrategy(),
		Pod:       DefaultKubernetesPod(),
		Container: DefaultKubernetesContainer(image),
	}
}

// DefaultKubernetesDaemonSet returns a new DefaultKubernetesDaemonSet with default settings.
func DefaultKubernetesDaemonSet(image string) *egv1a1.KubernetesDaemonSetSpec {
	return &egv1a1.KubernetesDaemonSetSpec{
		Strategy:  DefaultKubernetesDaemonSetStrategy(),
		Pod:       DefaultKubernetesPod(),
		Container: DefaultKubernetesContainer(image),
	}
}

// DefaultKubernetesPod returns a new KubernetesPodSpec with default settings.
func DefaultKubernetesPod() *egv1a1.KubernetesPodSpec {
	return &egv1a1.KubernetesPodSpec{}
}

// DefaultKubernetesContainer returns a new KubernetesContainerSpec with default settings.
func DefaultKubernetesContainer(image string) *egv1a1.KubernetesContainerSpec {
	return &egv1a1.KubernetesContainerSpec{
		Resources: DefaultResourceRequirements(),
		Image:     DefaultKubernetesContainerImage(image),
	}
}

// DefaultResourceRequirements returns a new ResourceRequirements with default settings.
func DefaultResourceRequirements() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(egv1a1.DefaultDeploymentCPUResourceRequests),
			corev1.ResourceMemory: resource.MustParse(egv1a1.DefaultDeploymentMemoryResourceRequests),
		},
	}
}

// DefaultKubernetesService returns a new KubernetesServiceSpec with default settings.
func DefaultKubernetesService() *egv1a1.KubernetesServiceSpec {
	return &egv1a1.KubernetesServiceSpec{
		Type:                  DefaultKubernetesServiceType(),
		ExternalTrafficPolicy: DefaultKubernetesServiceExternalTrafficPolicy(),
	}
}

// DefaultKubernetesServiceType returns a new KubernetesServiceType with default settings.
func DefaultKubernetesServiceType() *egv1a1.ServiceType {
	return GetKubernetesServiceType(egv1a1.ServiceTypeLoadBalancer)
}

// GetKubernetesServiceType returns the KubernetesServiceType pointer.
func GetKubernetesServiceType(serviceType egv1a1.ServiceType) *egv1a1.ServiceType {
	return &serviceType
}

// DefaultKubernetesServiceExternalTrafficPolicy returns the default external traffic policy.
func DefaultKubernetesServiceExternalTrafficPolicy() *egv1a1.ServiceExternalTrafficPolicy {
	return GetKubernetesServiceExternalTrafficPolicy(egv1a1.ServiceExternalTrafficPolicyLocal)
}

// GetKubernetesServiceExternalTrafficPolicy returns the ServiceExternalTrafficPolicy pointer.
func GetKubernetesServiceExternalTrafficPolicy(serviceExternalTrafficPolicy egv1a1.ServiceExternalTrafficPolicy) *egv1a1.ServiceExternalTrafficPolicy {
	return &serviceExternalTrafficPolicy
}

// defaultKubernetesDeploymentSpec fill a default KubernetesDeploymentSpec if unspecified.
func defaultKubernetesDeploymentSpec(deployment *egv1a1.KubernetesDeploymentSpec, image string) {
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
func defaultKubernetesDaemonSetSpec(daemonset *egv1a1.KubernetesDaemonSetSpec, image string) {
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

// setHPADefaults fill a default HorizontalPodAutoscalerSpec if unspecified
func setHPADefaults(hpa *egv1a1.KubernetesHorizontalPodAutoscalerSpec) {
	if len(hpa.Metrics) == 0 {
		hpa.Metrics = DefaultEnvoyProxyHpaMetrics()
	}
}

// ApplyPodDisruptionBudgetPatch applies the spec's merge patch to a PodDisruptionBudget.
// If disabled is true, the patch is ignored and old is returned unchanged.
func ApplyPodDisruptionBudgetPatch(pdb *egv1a1.KubernetesPodDisruptionBudgetSpec, old *policyv1.PodDisruptionBudget, disabled bool) (*policyv1.PodDisruptionBudget, error) {
	if pdb.Patch == nil || disabled {
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
	case pdb.Patch.Type == nil || *pdb.Patch.Type == egv1a1.StrategicMerge:
		patchedJSON, err = strategicpatch.StrategicMergePatch(originalJSON, pdb.Patch.Value.Raw, policyv1.PodDisruptionBudget{})
	case *pdb.Patch.Type == egv1a1.JSONMerge:
		patchedJSON, err = jsonpatch.MergePatch(originalJSON, pdb.Patch.Value.Raw)
	default:
		return nil, fmt.Errorf("unsupported merge type: %s", *pdb.Patch.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("error applying merge patch: %w", err)
	}

	// Deserialize the patched JSON into a new PodDisruptionBudget object
	var patchedPdb policyv1.PodDisruptionBudget
	if err := json.Unmarshal(patchedJSON, &patchedPdb); err != nil {
		return nil, fmt.Errorf("error unmarshaling patched PodDisruptionBudget: %w", err)
	}

	return &patchedPdb, nil
}
