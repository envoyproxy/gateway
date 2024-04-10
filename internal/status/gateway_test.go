// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestUpdateGatewayStatusProgrammedCondition(t *testing.T) {
	type args struct {
		gw         *gwapiv1.Gateway
		svc        *corev1.Service
		deployment *appsv1.Deployment
		addresses  []gwapiv1.GatewayStatusAddress
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nil svc",
			args: args{
				gw:        &gwapiv1.Gateway{},
				svc:       nil,
				addresses: nil,
			},
		},
		{
			name: "LoadBalancer svc with ingress ip",
			args: args{
				gw: &gwapiv1.Gateway{},
				svc: &corev1.Service{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.ServiceSpec{
						ClusterIPs: []string{"127.0.0.1"},
						Type:       corev1.ServiceTypeLoadBalancer,
					},
					Status: corev1.ServiceStatus{
						LoadBalancer: corev1.LoadBalancerStatus{
							Ingress: []corev1.LoadBalancerIngress{
								{
									IP: "127.0.0.1",
								},
							},
						},
					},
				},
				addresses: []gwapiv1.GatewayStatusAddress{
					{
						Type:  ptr.To(gwapiv1.IPAddressType),
						Value: "127.0.0.1",
					},
				},
			},
		},
		{
			name: "LoadBalancer svc with ingress hostname",
			args: args{
				gw: &gwapiv1.Gateway{},
				svc: &corev1.Service{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.ServiceSpec{
						ClusterIPs: []string{"127.0.0.1"},
						Type:       corev1.ServiceTypeLoadBalancer,
					},
					Status: corev1.ServiceStatus{
						LoadBalancer: corev1.LoadBalancerStatus{
							Ingress: []corev1.LoadBalancerIngress{
								{
									Hostname: "localhost",
								},
							},
						},
					},
				},
				addresses: []gwapiv1.GatewayStatusAddress{
					{
						Type:  ptr.To(gwapiv1.IPAddressType),
						Value: "127.0.0.1",
					},
					{
						Type:  ptr.To(gwapiv1.HostnameAddressType),
						Value: "localhost",
					},
				},
			},
		},
		{
			name: "ClusterIP svc",
			args: args{
				gw: &gwapiv1.Gateway{},
				svc: &corev1.Service{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.ServiceSpec{
						ClusterIPs: []string{"127.0.0.1"},
						Type:       corev1.ServiceTypeClusterIP,
					},
				},
				addresses: []gwapiv1.GatewayStatusAddress{
					{
						Type:  ptr.To(gwapiv1.IPAddressType),
						Value: "127.0.0.1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateGatewayStatusProgrammedCondition(tt.args.gw, tt.args.svc, tt.args.deployment)
			assert.True(t, reflect.DeepEqual(tt.args.addresses, tt.args.gw.Status.Addresses))
		})
	}
}
