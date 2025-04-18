// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
)

func TestEnvoyPodSelector(t *testing.T) {
	cases := []struct {
		name     string
		in       map[string]string
		expected map[string]string
	}{
		{
			name: "default",
			in:   map[string]string{"foo": "bar"},
			expected: map[string]string{
				"foo":                          "bar",
				"app.kubernetes.io/name":       "envoy",
				"app.kubernetes.io/component":  "proxy",
				"app.kubernetes.io/managed-by": "envoy-gateway",
			},
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := envoyLabels(tc.in)
			require.Equal(t, tc.expected, got)
		})
	}
}

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

func Test_expectedProxyInitContainers(t *testing.T) {
	type args struct {
		containerSpec   *egv1a1.KubernetesContainerSpec
		extraContainers []corev1.Container
	}
	tests := []struct {
		name string
		args args
		want []corev1.Container
	}{
		{
			name: "default",
			args: args{
				containerSpec:   &egv1a1.KubernetesContainerSpec{},
				extraContainers: []corev1.Container{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expectedProxyInitContainers(tt.args.extraContainers)
			require.Equal(t, tt.want, got)
		})
	}
}
