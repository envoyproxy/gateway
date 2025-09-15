// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
)

func TestExpectedShutdownManagerSecurityContext(t *testing.T) {
	defaultSecurityContext := func() *corev1.SecurityContext {
		sc := resource.DefaultSecurityContext()

		// run as non-root user
		sc.RunAsGroup = ptr.To(int64(65532))
		sc.RunAsUser = ptr.To(int64(65532))

		// ShutdownManger creates a file to indicate the connection drain process is completed,
		// so it needs file write permission.
		sc.ReadOnlyRootFilesystem = nil
		return sc
	}

	customSc := &corev1.SecurityContext{
		Privileged: ptr.To(true),
		RunAsUser:  ptr.To(int64(21)),
		RunAsGroup: ptr.To(int64(2100)),
	}

	tests := []struct {
		name     string
		in       *egv1a1.KubernetesContainerSpec
		expected *corev1.SecurityContext
	}{
		{
			name:     "default",
			in:       nil,
			expected: defaultSecurityContext(),
		},
		{
			name: "default",
			in: &egv1a1.KubernetesContainerSpec{
				SecurityContext: customSc,
			},
			expected: customSc,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := expectedShutdownManagerSecurityContext(tc.in)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestResolveProxyImage(t *testing.T) {
	defaultImage := egv1a1.DefaultEnvoyProxyImage
	defaultTag := "distroless-v1.35.3"

	tests := []struct {
		name        string
		container   *egv1a1.KubernetesContainerSpec
		expected    string
		expectError bool
	}{
		{
			name:        "nil containerSpec",
			container:   nil,
			expectError: true,
		},
		{
			name: "imageRepository set",
			container: &egv1a1.KubernetesContainerSpec{
				ImageRepository: ptr.To("envoyproxy/envoy"),
			},
			expected: fmt.Sprintf("envoyproxy/envoy:%s", defaultTag),
		},
		{
			name: "image set",
			container: &egv1a1.KubernetesContainerSpec{
				Image: ptr.To("envoyproxy/envoy:v1.2.3"),
			},
			expected: "envoyproxy/envoy:v1.2.3",
		},
		{
			name:      "neither set",
			container: &egv1a1.KubernetesContainerSpec{},
			expected:  defaultImage,
		},
		{
			name: "both image and imageRepository set (invalid per CRD, but still testable)",
			container: &egv1a1.KubernetesContainerSpec{
				Image:           ptr.To("envoyproxy/envoy:v1.2.3"),
				ImageRepository: ptr.To("envoyproxy/envoy"),
			},
			expected: fmt.Sprintf("envoyproxy/envoy:%s", defaultTag),
		},
		{
			name: "imageRepository with port",
			container: &egv1a1.KubernetesContainerSpec{
				ImageRepository: ptr.To("docker.io:443/envoyproxy/envoy"),
			},
			expected: fmt.Sprintf("docker.io:443/envoyproxy/envoy:%s", defaultTag),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			image, err := resolveProxyImage(tc.container)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, image)
			}
		})
	}
}

func TestGetImageTag(t *testing.T) {
	tests := []struct {
		name        string
		image       string
		expectedTag string
		expectErr   bool
	}{
		{
			name:        "valid image with tag",
			image:       "docker.io/envoyproxy/envoy:distroless-v1.34.1",
			expectedTag: "distroless-v1.34.1",
			expectErr:   false,
		},
		{
			name:      "image without tag",
			image:     "docker.io/envoyproxy/envoy",
			expectErr: true,
		},
		{
			name:      "image with digest but no tag",
			image:     "docker.io/envoyproxy/envoy@sha256:abcdef123456",
			expectErr: true,
		},
		{
			name:        "localhost with port and tag",
			image:       "localhost:5000/myimage:v2.0",
			expectedTag: "v2.0",
			expectErr:   false,
		},
		{
			name:      "invalid image format",
			image:     "!!!not-a-valid-image###",
			expectErr: true,
		},
		{
			name:        "ghcr image with tag",
			image:       "ghcr.io/org/image:latest",
			expectedTag: "latest",
			expectErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := getImageTag(tt.image)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedTag, tag)
			}
		})
	}
}
