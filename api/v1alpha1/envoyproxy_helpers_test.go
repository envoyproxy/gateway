// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestDefaultShutdownManagerContainerResourceRequirements(t *testing.T) {
	got := DefaultShutdownManagerContainerResourceRequirements()
	require.Equal(t, corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(DefaultShutdownManagerCPUResourceRequests),
		corev1.ResourceMemory: resource.MustParse(DefaultShutdownManagerMemoryResourceRequests),
	}, got.Requests)
	require.Equal(t, corev1.ResourceList{
		corev1.ResourceMemory: resource.MustParse(DefaultShutdownManagerMemoryResourceLimit),
	}, got.Limits)
	_, hasCPULimit := got.Limits[corev1.ResourceCPU]
	require.False(t, hasCPULimit)
}
