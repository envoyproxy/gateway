// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestDefaultShutdownManagerContainerResourceRequirements(t *testing.T) {
	got := DefaultShutdownManagerContainerResourceRequirements()

	assert.Equal(t, corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(DefaultShutdownManagerCPUResourceRequests),
			corev1.ResourceMemory: resource.MustParse(DefaultShutdownManagerMemoryResourceRequests),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse(DefaultShutdownManagerMemoryResourceLimits),
		},
	}, *got)

	// A CPU limit is intentionally not set to avoid CPU throttling.
	_, hasCPULimit := got.Limits[corev1.ResourceCPU]
	assert.False(t, hasCPULimit)
}
