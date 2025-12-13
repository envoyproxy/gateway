// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestMergeEnvVars(t *testing.T) {
	tests := []struct {
		name      string
		defaults  []corev1.EnvVar
		overrides []corev1.EnvVar
		expected  []corev1.EnvVar
	}{
		{
			name:      "both empty",
			defaults:  nil,
			overrides: nil,
			expected:  nil,
		},
		{
			name:      "only defaults",
			defaults:  []corev1.EnvVar{{Name: "DEFAULT", Value: "value"}},
			overrides: nil,
			expected:  []corev1.EnvVar{{Name: "DEFAULT", Value: "value"}},
		},
		{
			name:      "only overrides",
			defaults:  nil,
			overrides: []corev1.EnvVar{{Name: "OVERRIDE", Value: "value"}},
			expected:  []corev1.EnvVar{{Name: "OVERRIDE", Value: "value"}},
		},
		{
			name:      "merge different vars",
			defaults:  []corev1.EnvVar{{Name: "DEFAULT", Value: "default"}},
			overrides: []corev1.EnvVar{{Name: "OVERRIDE", Value: "override"}},
			expected: []corev1.EnvVar{
				{Name: "DEFAULT", Value: "default"},
				{Name: "OVERRIDE", Value: "override"},
			},
		},
		{
			name:      "override same var",
			defaults:  []corev1.EnvVar{{Name: "VAR", Value: "default"}},
			overrides: []corev1.EnvVar{{Name: "VAR", Value: "override"}},
			expected:  []corev1.EnvVar{{Name: "VAR", Value: "override"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeEnvVars(tc.defaults, tc.overrides)
			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestMergeVolumeMounts(t *testing.T) {
	tests := []struct {
		name      string
		defaults  []corev1.VolumeMount
		overrides []corev1.VolumeMount
		expected  []corev1.VolumeMount
	}{
		{
			name:      "both empty",
			defaults:  nil,
			overrides: nil,
			expected:  nil,
		},
		{
			name:      "only defaults",
			defaults:  []corev1.VolumeMount{{Name: "default-vol", MountPath: "/default"}},
			overrides: nil,
			expected:  []corev1.VolumeMount{{Name: "default-vol", MountPath: "/default"}},
		},
		{
			name:      "merge different",
			defaults:  []corev1.VolumeMount{{Name: "default-vol", MountPath: "/default"}},
			overrides: []corev1.VolumeMount{{Name: "override-vol", MountPath: "/override"}},
			expected: []corev1.VolumeMount{
				{Name: "default-vol", MountPath: "/default"},
				{Name: "override-vol", MountPath: "/override"}},
		},
		{
			name:      "override by name",
			defaults:  []corev1.VolumeMount{{Name: "vol", MountPath: "/default"}},
			overrides: []corev1.VolumeMount{{Name: "vol", MountPath: "/override"}},
			expected:  []corev1.VolumeMount{{Name: "vol", MountPath: "/override"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeVolumeMounts(tc.defaults, tc.overrides)
			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name      string
		defaults  map[string]string
		overrides map[string]string
		expected  map[string]string
	}{
		{
			name:      "both empty",
			defaults:  nil,
			overrides: nil,
			expected:  nil,
		},
		{
			name:      "only defaults",
			defaults:  map[string]string{"key": "default"},
			overrides: nil,
			expected:  map[string]string{"key": "default"},
		},
		{
			name:      "only overrides",
			defaults:  nil,
			overrides: map[string]string{"key": "override"},
			expected:  map[string]string{"key": "override"},
		},
		{
			name:      "merge different keys",
			defaults:  map[string]string{"default-key": "value1"},
			overrides: map[string]string{"override-key": "value2"},
			expected:  map[string]string{"default-key": "value1", "override-key": "value2"},
		},
		{
			name:      "override same key",
			defaults:  map[string]string{"key": "default"},
			overrides: map[string]string{"key": "override"},
			expected:  map[string]string{"key": "override"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeMaps(tc.defaults, tc.overrides)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestMergeVolumes(t *testing.T) {
	tests := []struct {
		name      string
		defaults  []corev1.Volume
		overrides []corev1.Volume
		expected  []corev1.Volume
	}{
		{
			name:      "both empty",
			defaults:  nil,
			overrides: nil,
			expected:  nil,
		},
		{
			name: "only defaults",
			defaults: []corev1.Volume{
				{Name: "default-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
			overrides: nil,
			expected: []corev1.Volume{
				{Name: "default-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
		},
		{
			name:      "merge different",
			defaults:  []corev1.Volume{{Name: "default-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			overrides: []corev1.Volume{{Name: "override-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			expected: []corev1.Volume{
				{Name: "default-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "override-vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
			},
		},
		{
			name:     "override by name",
			defaults: []corev1.Volume{{Name: "vol", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
			overrides: []corev1.Volume{{Name: "vol", VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "config"}},
			}}},
			expected: []corev1.Volume{{Name: "vol", VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: "config"}},
			}}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeVolumes(tc.defaults, tc.overrides)
			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestMergeContainerDefaults(t *testing.T) {
	tests := []struct {
		name          string
		defaults      *egv1a1.KubernetesContainerSpec
		containerSpec *egv1a1.KubernetesContainerSpec
		expected      *egv1a1.KubernetesContainerSpec
	}{
		{
			name:          "both nil",
			defaults:      nil,
			containerSpec: nil,
			expected:      nil,
		},
		{
			name: "nil containerSpec returns defaults",
			defaults: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("default/image:v1"),
			},
			containerSpec: nil,
			expected: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("default/image:v1"),
			},
		},
		{
			name:     "nil defaults returns containerSpec",
			defaults: nil,
			containerSpec: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("specific/image:v1"),
			},
			expected: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("specific/image:v1"),
			},
		},
		{
			name: "specific image overrides default",
			defaults: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("default/image:v1"),
			},
			containerSpec: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("specific/image:v2"),
			},
			expected: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("specific/image:v2"),
			},
		},
		{
			name: "default image overrides fallback default",
			defaults: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("template/image:v1"),
			},
			containerSpec: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To(egv1a1.DefaultEnvoyProxyImage),
			},
			expected: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("template/image:v1"),
			},
		},
		{
			name: "env vars are merged",
			defaults: &egv1a1.KubernetesContainerSpec{
				Env: []corev1.EnvVar{{Name: "DEFAULT", Value: "value"}},
			},
			containerSpec: &egv1a1.KubernetesContainerSpec{
				Env: []corev1.EnvVar{{Name: "SPECIFIC", Value: "value"}},
			},
			expected: &egv1a1.KubernetesContainerSpec{
				Env: []corev1.EnvVar{
					{Name: "DEFAULT", Value: "value"},
					{Name: "SPECIFIC", Value: "value"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeContainerDefaults(tc.defaults, tc.containerSpec, egv1a1.DefaultEnvoyProxyImage)
			if tc.expected == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tc.expected.Image, got.Image)
				assert.ElementsMatch(t, tc.expected.Env, got.Env)
			}
		})
	}
}

func TestMergePodDefaults(t *testing.T) {
	tests := []struct {
		name     string
		defaults *egv1a1.KubernetesPodSpec
		podSpec  *egv1a1.KubernetesPodSpec
		expected *egv1a1.KubernetesPodSpec
	}{
		{
			name:     "both nil",
			defaults: nil,
			podSpec:  nil,
			expected: nil,
		},
		{
			name: "nil podSpec returns defaults",
			defaults: &egv1a1.KubernetesPodSpec{
				Annotations: map[string]string{"key": "value"},
			},
			podSpec: nil,
			expected: &egv1a1.KubernetesPodSpec{
				Annotations: map[string]string{"key": "value"},
			},
		},
		{
			name:     "nil defaults returns podSpec",
			defaults: nil,
			podSpec: &egv1a1.KubernetesPodSpec{
				Labels: map[string]string{"key": "value"},
			},
			expected: &egv1a1.KubernetesPodSpec{
				Labels: map[string]string{"key": "value"},
			},
		},
		{
			name: "annotations are merged",
			defaults: &egv1a1.KubernetesPodSpec{
				Annotations: map[string]string{"default": "value1"},
			},
			podSpec: &egv1a1.KubernetesPodSpec{
				Annotations: map[string]string{"specific": "value2"},
			},
			expected: &egv1a1.KubernetesPodSpec{
				Annotations: map[string]string{"default": "value1", "specific": "value2"},
			},
		},
		{
			name: "labels are merged",
			defaults: &egv1a1.KubernetesPodSpec{
				Labels: map[string]string{"default": "value1"},
			},
			podSpec: &egv1a1.KubernetesPodSpec{
				Labels: map[string]string{"specific": "value2"},
			},
			expected: &egv1a1.KubernetesPodSpec{
				Labels: map[string]string{"default": "value1", "specific": "value2"},
			},
		},
		{
			name: "tolerations are appended",
			defaults: &egv1a1.KubernetesPodSpec{
				Tolerations: []corev1.Toleration{{Key: "default", Operator: corev1.TolerationOpEqual, Value: "value"}},
			},
			podSpec: &egv1a1.KubernetesPodSpec{
				Tolerations: []corev1.Toleration{{Key: "specific", Operator: corev1.TolerationOpEqual, Value: "value"}},
			},
			expected: &egv1a1.KubernetesPodSpec{
				Tolerations: []corev1.Toleration{
					{Key: "default", Operator: corev1.TolerationOpEqual, Value: "value"},
					{Key: "specific", Operator: corev1.TolerationOpEqual, Value: "value"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MergePodDefaults(tc.defaults, tc.podSpec)
			if tc.expected == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.Equal(t, tc.expected.Annotations, got.Annotations)
				assert.Equal(t, tc.expected.Labels, got.Labels)
				assert.ElementsMatch(t, tc.expected.Tolerations, got.Tolerations)
			}
		})
	}
}
