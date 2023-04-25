// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

func TestExpectedServiceSpec(t *testing.T) {
	type args struct {
		serviceType *egcfgv1a1.ServiceType
	}
	tests := []struct {
		name string
		args args
		want corev1.ServiceSpec
	}{
		{
			name: "LoadBalancer",
			args: args{serviceType: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeLoadBalancer)},
			want: corev1.ServiceSpec{
				Type:                  corev1.ServiceTypeLoadBalancer,
				SessionAffinity:       corev1.ServiceAffinityNone,
				ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			},
		},
		{
			name: "ClusterIP",
			args: args{serviceType: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeClusterIP)},
			want: corev1.ServiceSpec{
				Type:            corev1.ServiceTypeClusterIP,
				SessionAffinity: corev1.ServiceAffinityNone,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, ExpectedServiceSpec(tt.args.serviceType), "expectedServiceSpec(%v)", tt.args.serviceType)
		})
	}
}

func TestGetSelector(t *testing.T) {
	cases := []struct {
		name     string
		in       map[string]string
		expected map[string]string
	}{
		{
			name: "default",
			in: map[string]string{
				"foo":                            "bar",
				"app.gateway.envoyproxy.io/name": "envoy",
			},
			expected: map[string]string{
				"foo":                            "bar",
				"app.gateway.envoyproxy.io/name": "envoy",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run("", func(t *testing.T) {
			got := GetSelector(tc.in)
			require.Equal(t, tc.expected, got.MatchLabels)
		})
	}
}
