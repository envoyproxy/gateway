// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// ResourceKind indicates the main resources of envoy-ratelimit,
// but also the key for the uid of their ownerReference.
const (
	ResourceKindDeployment = "Deployment"
)

func GetPodDisruptionBudget(pdb *egv1a1.KubernetesPodDisruptionBudgetSpec,
	selector *metav1.LabelSelector, nn *types.NamespacedName, ownerReferences []metav1.OwnerReference,
) (*policyv1.PodDisruptionBudget, error) {
	// If podDisruptionBudget config is nil, ignore PodDisruptionBudget.
	if pdb == nil {
		return nil, nil
	}

	pdbSpec := policyv1.PodDisruptionBudgetSpec{
		Selector: selector,
	}

	switch {
	case pdb.MinAvailable != nil:
		pdbSpec.MinAvailable = pdb.MinAvailable
	case pdb.MaxUnavailable != nil:
		pdbSpec.MaxUnavailable = pdb.MaxUnavailable
	default:
		pdbSpec.MinAvailable = &intstr.IntOrString{Type: intstr.Int, IntVal: 0}
	}

	podDisruptionBudget := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
			Labels:    selector.MatchLabels,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1",
			Kind:       "PodDisruptionBudget",
		},
		Spec: pdbSpec,
	}

	podDisruptionBudget.OwnerReferences = append(podDisruptionBudget.OwnerReferences, ownerReferences...)

	// apply merge patch to PodDisruptionBudget
	podDisruptionBudget, err := pdb.ApplyMergePatch(podDisruptionBudget)
	if err != nil {
		return nil, err
	}
	return podDisruptionBudget, nil
}

// MergeEnvVars merges two slices of environment variables.
// Variables in 'overrides' take precedence over those in 'defaults' by name.
func MergeEnvVars(defaults, overrides []corev1.EnvVar) []corev1.EnvVar {
	if len(defaults) == 0 {
		return overrides
	}
	if len(overrides) == 0 {
		return defaults
	}

	// Create a map of override env vars by name for quick lookup
	overrideMap := make(map[string]corev1.EnvVar)
	for _, env := range overrides {
		overrideMap[env.Name] = env
	}

	// Start with defaults, replacing any that are overridden.  Remove processed 
	// overrides from the overrideMap.
	merged := make([]corev1.EnvVar, 0, len(defaults)+len(overrides))
	for _, env := range defaults {
		if override, exists := overrideMap[env.Name]; exists {
			merged = append(merged, override)
			delete(overrideMap, env.Name) 
		} else {
			merged = append(merged, env)
		}
	}

	// Add any override env vars that weren't in defaults
	for _, env := range overrideMap {
		merged = append(merged, env)
	}

	return merged
}

// MergeVolumeMounts merges two slices of volume mounts.
// Mounts in 'overrides' take precedence over those in 'defaults' by name.
func MergeVolumeMounts(defaults, overrides []corev1.VolumeMount) []corev1.VolumeMount {
	if len(defaults) == 0 {
		return overrides
	}
	if len(overrides) == 0 {
		return defaults
	}

	overrideMap := make(map[string]corev1.VolumeMount)
	for _, vm := range overrides {
		overrideMap[vm.Name] = vm
	}

	merged := make([]corev1.VolumeMount, 0, len(defaults)+len(overrides))
	for _, vm := range defaults {
		if override, exists := overrideMap[vm.Name]; exists {
			merged = append(merged, override)
			delete(overrideMap, vm.Name) // Mark as processed
		} else {
			merged = append(merged, vm)
		}
	}

	for _, vm := range overrideMap {
		merged = append(merged, vm)
	}

	return merged
}

// MergeMaps merges two string maps.
// Keys in 'overrides' take precedence over those in 'defaults'.
func MergeMaps(defaults, overrides map[string]string) map[string]string {
	if len(defaults) == 0 {
		return overrides
	}
	if len(overrides) == 0 {
		return defaults
	}

	merged := make(map[string]string, len(defaults)+len(overrides))
	maps.Copy(merged, defaults)
	maps.Copy(merged, overrides)

	return merged
}

// MergeVolumes merges two slices of volumes.
// Volumes in 'overrides' take precedence over those in 'defaults' by name.
func MergeVolumes(defaults, overrides []corev1.Volume) []corev1.Volume {
	if len(defaults) == 0 {
		return overrides
	}
	if len(overrides) == 0 {
		return defaults
	}

	overrideMap := make(map[string]corev1.Volume)
	for _, vol := range overrides {
		overrideMap[vol.Name] = vol
	}

	merged := make([]corev1.Volume, 0, len(defaults)+len(overrides))
	for _, vol := range defaults {
		if override, exists := overrideMap[vol.Name]; exists {
			merged = append(merged, override)
			delete(overrideMap, vol.Name)
		} else {
			merged = append(merged, vol)
		}
	}

	for _, vol := range overrideMap {
		merged = append(merged, vol)
	}

	return merged
}

// MergeContainerDefaults merges default container spec with specific container spec.
// Specific container spec fields take precedence over defaults.
// For the Image field, if the specific spec has the fallbackDefaultImage and defaults has an image,
// the defaults image is used (allowing templates to override the fallback default).
// The fallbackDefaultImage parameter allows this function to be used for different container types
// (e.g., proxy containers with DefaultEnvoyProxyImage, ratelimit with DefaultRateLimitImage).
func MergeContainerDefaults(defaults, containerSpec *egv1a1.KubernetesContainerSpec, fallbackDefaultImage string) *egv1a1.KubernetesContainerSpec {
	// If no defaults are configured, return containerSpec as-is
	if defaults == nil {
		return containerSpec
	}

	// If containerSpec is nil, use defaults directly
	if containerSpec == nil {
		return defaults
	}

	// Merge: containerSpec fields take precedence over defaults
	merged := &egv1a1.KubernetesContainerSpec{}

	// If the containerSpec has the fallback default image and template has an image, use template
	if containerSpec.Image != nil && *containerSpec.Image == fallbackDefaultImage && defaults.Image != nil {
		merged.Image = defaults.Image
	} else if containerSpec.Image != nil {
		merged.Image = containerSpec.Image
	} else if defaults.Image != nil {
		merged.Image = defaults.Image
	}

	if containerSpec.ImageRepository != nil {
		merged.ImageRepository = containerSpec.ImageRepository
	} else if defaults.ImageRepository != nil {
		merged.ImageRepository = defaults.ImageRepository
	}

	if containerSpec.ImagePullPolicy != nil {
		merged.ImagePullPolicy = containerSpec.ImagePullPolicy
	} else if defaults.ImagePullPolicy != nil {
		merged.ImagePullPolicy = defaults.ImagePullPolicy
	}

	// Env: merge env vars, specific values override defaults by name
	merged.Env = MergeEnvVars(defaults.Env, containerSpec.Env)

	if containerSpec.Resources != nil {
		merged.Resources = containerSpec.Resources
	} else if defaults.Resources != nil {
		merged.Resources = defaults.Resources
	}

	if containerSpec.SecurityContext != nil {
		merged.SecurityContext = containerSpec.SecurityContext
	} else if defaults.SecurityContext != nil {
		merged.SecurityContext = defaults.SecurityContext
	}

	// VolumeMounts: merge volume mounts, specific values override defaults by name
	merged.VolumeMounts = MergeVolumeMounts(defaults.VolumeMounts, containerSpec.VolumeMounts)

	return merged
}

// MergePodDefaults merges default pod spec with specific pod spec.
// Specific pod spec fields take precedence over defaults.
func MergePodDefaults(defaults, podSpec *egv1a1.KubernetesPodSpec) *egv1a1.KubernetesPodSpec {
	// If no defaults are configured, return podSpec as-is
	if defaults == nil {
		return podSpec
	}

	// If podSpec is nil, use defaults directly
	if podSpec == nil {
		return defaults
	}

	// Merge: podSpec fields take precedence over defaults
	merged := &egv1a1.KubernetesPodSpec{}

	merged.Annotations = MergeMaps(defaults.Annotations, podSpec.Annotations)
	merged.Labels = MergeMaps(defaults.Labels, podSpec.Labels)

	if podSpec.SecurityContext != nil {
		merged.SecurityContext = podSpec.SecurityContext
	} else if defaults.SecurityContext != nil {
		merged.SecurityContext = defaults.SecurityContext
	}

	if podSpec.Affinity != nil {
		merged.Affinity = podSpec.Affinity
	} else if defaults.Affinity != nil {
		merged.Affinity = defaults.Affinity
	}

	merged.Tolerations = append(merged.Tolerations, defaults.Tolerations...)
	merged.Tolerations = append(merged.Tolerations, podSpec.Tolerations...)

	merged.Volumes = MergeVolumes(defaults.Volumes, podSpec.Volumes)

	merged.ImagePullSecrets = append(merged.ImagePullSecrets, defaults.ImagePullSecrets...)
	merged.ImagePullSecrets = append(merged.ImagePullSecrets, podSpec.ImagePullSecrets...)

	merged.NodeSelector = MergeMaps(defaults.NodeSelector, podSpec.NodeSelector)

	merged.TopologySpreadConstraints = append(merged.TopologySpreadConstraints, defaults.TopologySpreadConstraints...)
	merged.TopologySpreadConstraints = append(merged.TopologySpreadConstraints, podSpec.TopologySpreadConstraints...)

	return merged
}
