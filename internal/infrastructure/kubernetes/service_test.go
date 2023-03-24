// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

func checkServiceSpec(t *testing.T, svc *corev1.Service, expected corev1.ServiceSpec) {
	t.Helper()

	expected.Ports = svc.Spec.Ports
	expected.Selector = svc.Spec.Selector
	if apiequality.Semantic.DeepEqual(svc.Spec, expected) {
		return
	}

	t.Errorf("service has unexpected %q spec", &svc.Spec)
}

func TestExpectedServiceSpec(t *testing.T) {
	type args struct {
		serviceType *v1alpha1.ServiceType
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
			assert.Equalf(t, tt.want, expectedServiceSpec(tt.args.serviceType), "expectedServiceSpec(%v)", tt.args.serviceType)
		})
	}
}
