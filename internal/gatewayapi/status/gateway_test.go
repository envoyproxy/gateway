// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// TestUpdateGatewayStatusProgrammedCondition tests whether UpdateGatewayStatusProgrammedCondition correctly updates the addresses in the Gateway status.
func TestUpdateGatewayStatusProgrammedCondition(t *testing.T) {
	type args struct {
		gw            *gwapiv1.Gateway
		svc           *corev1.Service
		deployment    *appsv1.Deployment
		nodeAddresses []string
	}
	tests := []struct {
		name          string
		args          args
		wantAddresses []gwapiv1.GatewayStatusAddress
	}{
		{
			name: "nil svc",
			args: args{
				gw:  &gwapiv1.Gateway{},
				svc: nil,
			},
			wantAddresses: nil,
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
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{
				{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: "127.0.0.1",
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
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{
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
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{
				{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: "127.0.0.1",
				},
			},
		},
		{
			name: "Nodeport svc",
			args: args{
				gw:            &gwapiv1.Gateway{},
				nodeAddresses: []string{"1", "2"},
				svc: &corev1.Service{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeNodePort,
					},
				},
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{
				{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: "1",
				},
				{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: "2",
				},
			},
		},
		{
			name: "Nodeport svc with too many node addresses",
			args: args{
				gw: &gwapiv1.Gateway{},
				// 20 node addresses
				nodeAddresses: func() (addr []string) {
					for i := 0; i < 20; i++ {
						addr = append(addr, strconv.Itoa(i))
					}
					return
				}(),
				svc: &corev1.Service{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeNodePort,
					},
				},
			},
			// Only the first 16 addresses should be set.
			wantAddresses: func() (addr []gwapiv1.GatewayStatusAddress) {
				for i := 0; i < 16; i++ {
					addr = append(addr, gwapiv1.GatewayStatusAddress{
						Type:  ptr.To(gwapiv1.IPAddressType),
						Value: strconv.Itoa(i),
					})
				}
				return
			}(),
		},
		{
			name: "LoadBalancer svc with IPv6 ingress ip",
			args: args{
				gw: &gwapiv1.Gateway{},
				svc: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeLoadBalancer,
					},
					Status: corev1.ServiceStatus{
						LoadBalancer: corev1.LoadBalancerStatus{
							Ingress: []corev1.LoadBalancerIngress{
								{IP: "2001:db8::1"},
							},
						},
					},
				},
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{
				{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: "2001:db8::1",
				},
			},
		},
		{
			name: "ClusterIP svc with IPv6",
			args: args{
				gw: &gwapiv1.Gateway{},
				svc: &corev1.Service{
					Spec: corev1.ServiceSpec{
						ClusterIPs: []string{"2001:db8::2"},
						Type:       corev1.ServiceTypeClusterIP,
					},
				},
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{
				{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: "2001:db8::2",
				},
			},
		},
		{
			name: "Nodeport svc with IPv6 node addresses",
			args: args{
				gw:            &gwapiv1.Gateway{},
				nodeAddresses: []string{"2001:db8::3", "2001:db8::4"},
				svc: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Type: corev1.ServiceTypeNodePort,
					},
				},
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{
				{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: "2001:db8::3",
				},
				{
					Type:  ptr.To(gwapiv1.IPAddressType),
					Value: "2001:db8::4",
				},
			},
		},
		{
			name: "Headless ClusterIP svc with None",
			args: args{
				gw: &gwapiv1.Gateway{},
				svc: &corev1.Service{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.ServiceSpec{
						ClusterIPs: []string{"None"},
						Type:       corev1.ServiceTypeClusterIP,
					},
				},
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{},
		},
		{
			name: "Headless ClusterIP svc with None and explicit Gateway addresses",
			args: args{
				gw: &gwapiv1.Gateway{
					Spec: gwapiv1.GatewaySpec{
						Addresses: []gwapiv1.GatewaySpecAddress{
							{
								Type:  ptr.To(gwapiv1.IPAddressType),
								Value: "10.0.0.1",
							},
						},
					},
				},
				svc: &corev1.Service{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.ServiceSpec{
						ClusterIPs: []string{"None"},
						Type:       corev1.ServiceTypeClusterIP,
					},
				},
			},
			wantAddresses: []gwapiv1.GatewayStatusAddress{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateGatewayStatusProgrammedCondition(tt.args.gw, tt.args.svc, tt.args.deployment, tt.args.nodeAddresses...)
			assert.True(t, reflect.DeepEqual(tt.wantAddresses, tt.args.gw.Status.Addresses))
		})
	}
}

func TestUpdateGatewayProgrammedCondition(t *testing.T) {
	testCases := []struct {
		name string
		// serviceAddressNum indicates how many addresses are set in the Gateway status.
		serviceAddressNum int
		deploymentStatus  appsv1.DeploymentStatus
		expectCondition   []metav1.Condition
	}{
		{
			name:              "ready gateway",
			serviceAddressNum: 1,
			deploymentStatus:  appsv1.DeploymentStatus{AvailableReplicas: 1, Replicas: 1},
			expectCondition: []metav1.Condition{
				{
					Type:    string(gwapiv1.GatewayConditionProgrammed),
					Status:  metav1.ConditionTrue,
					Reason:  string(gwapiv1.GatewayConditionProgrammed),
					Message: fmt.Sprintf(messageFmtProgrammed, 1, 1),
				},
			},
		},
		{
			name:              "not ready gateway without address",
			serviceAddressNum: 0,
			deploymentStatus:  appsv1.DeploymentStatus{AvailableReplicas: 1},
			expectCondition: []metav1.Condition{
				{
					Type:    string(gwapiv1.GatewayConditionProgrammed),
					Status:  metav1.ConditionFalse,
					Reason:  string(gwapiv1.GatewayReasonAddressNotAssigned),
					Message: messageAddressNotAssigned,
				},
			},
		},
		{
			name:              "not ready gateway with too many addresses",
			serviceAddressNum: 17,
			deploymentStatus:  appsv1.DeploymentStatus{AvailableReplicas: 1},
			expectCondition: []metav1.Condition{
				{
					Type:    string(gwapiv1.GatewayConditionProgrammed),
					Status:  metav1.ConditionFalse,
					Reason:  string(gwapiv1.GatewayReasonInvalid),
					Message: fmt.Sprintf(messageFmtTooManyAddresses, 17),
				},
			},
		},
		{
			name:              "not ready gateway with address unavailable pods",
			serviceAddressNum: 1,
			deploymentStatus:  appsv1.DeploymentStatus{AvailableReplicas: 0},
			expectCondition: []metav1.Condition{
				{
					Type:    string(gwapiv1.GatewayConditionProgrammed),
					Status:  metav1.ConditionFalse,
					Reason:  string(gwapiv1.GatewayReasonNoResources),
					Message: messageNoResources,
				},
			},
		},
	}

	for _, tc := range testCases {
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
			updateGatewayProgrammedCondition(gtw, deployment)

			if d := cmp.Diff(tc.expectCondition, gtw.Status.Conditions, cmpopts.IgnoreFields(metav1.Condition{}, "LastTransitionTime")); d != "" {
				t.Errorf("unexpected condition diff: %s", d)
			}
		})
	}
}
