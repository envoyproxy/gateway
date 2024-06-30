// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"reflect"
	"strconv"
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

func TestGatewayReadyCondition(t *testing.T) {
	testCases := []struct {
		name string
		// serviceAddressNum indicates how many addresses are set in the Gateway status.
		serviceAddressNum int
		deploymentStatus  appsv1.DeploymentStatus
		expect            metav1.Condition
	}{
		{
			name:              "ready gateway",
			serviceAddressNum: 1,
			deploymentStatus:  appsv1.DeploymentStatus{AvailableReplicas: 1},
			expect: metav1.Condition{
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.GatewayConditionProgrammed),
			},
		},
		{
			name:              "not ready gateway without address",
			serviceAddressNum: 0,
			deploymentStatus:  appsv1.DeploymentStatus{AvailableReplicas: 1},
			expect: metav1.Condition{
				Status: metav1.ConditionFalse,
				Reason: string(gwapiv1.GatewayReasonAddressNotAssigned),
			},
		},
		{
			name:              "not ready gateway with too many addresses",
			serviceAddressNum: 17,
			deploymentStatus:  appsv1.DeploymentStatus{AvailableReplicas: 1},
			expect: metav1.Condition{
				Status: metav1.ConditionFalse,
				Reason: string(gwapiv1.GatewayReasonInvalid),
			},
		},
		{
			name:              "not ready gateway with address unavailable pods",
			serviceAddressNum: 1,
			deploymentStatus:  appsv1.DeploymentStatus{AvailableReplicas: 0},
			expect: metav1.Condition{
				Status: metav1.ConditionFalse,
				Reason: string(gwapiv1.GatewayReasonNoResources),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gtw := &gwapiv1.Gateway{}
			gtw.Status.Addresses = make([]gwapiv1.GatewayStatusAddress, tc.serviceAddressNum)
			for i := 0; i < tc.serviceAddressNum; i++ {
				gtw.Status.Addresses[i] = gwapiv1.GatewayStatusAddress{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: strconv.Itoa(i),
				}
			}

			deployment := &appsv1.Deployment{Status: tc.deploymentStatus}
			got := computeGatewayProgrammedCondition(gtw, deployment)

			assert.Equal(t, string(gwapiv1.GatewayConditionProgrammed), got.Type)
			assert.Equal(t, tc.expect.Status, got.Status)
			assert.Equal(t, tc.expect.Reason, got.Reason)
		})
	}
}

func TestComputeGatewayScheduledCondition(t *testing.T) {
	testCases := []struct {
		name   string
		sched  bool
		expect metav1.Condition
	}{
		{
			name:  "scheduled gateway",
			sched: true,
			expect: metav1.Condition{
				Type:   string(gwapiv1.GatewayReasonAccepted),
				Status: metav1.ConditionTrue,
			},
		},
		{
			name:  "not scheduled gateway",
			sched: false,
			expect: metav1.Condition{
				Type:   string(gwapiv1.GatewayReasonAccepted),
				Status: metav1.ConditionFalse,
			},
		},
	}

	for _, tc := range testCases {
		gw := &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "test",
			},
		}

		got := computeGatewayAcceptedCondition(gw, tc.sched)

		assert.Equal(t, tc.expect.Type, got.Type)
		assert.Equal(t, tc.expect.Status, got.Status)
	}
}
