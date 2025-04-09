// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Test_mergeRouteParentStatus(t *testing.T) {
	type args struct {
		ns  string
		old []gwapiv1.RouteParentStatus
		new []gwapiv1.RouteParentStatus
	}
	tests := []struct {
		name string
		args args
		want []gwapiv1.RouteParentStatus
	}{
		{
			name: "merge old and new",
			args: args{
				ns: "default",
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:        "gateway1",
							Namespace:   ptr.To[gwapiv1.Namespace]("default"),
							SectionName: ptr.To[gwapiv1.SectionName]("listener1"),
							Port:        ptr.To[gwapiv1.PortNumber](80),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
								Status: metav1.ConditionTrue,
								Reason: "ResolvedRefs",
							},
						},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:      "gateway2",
							Namespace: ptr.To[gwapiv1.Namespace]("default"),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "OldReason",
							},
						},
					},
				},
				new: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:      "gateway2",
							Namespace: ptr.To[gwapiv1.Namespace]("default"),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionFalse,
								Reason: "SomeReason",
							},
						},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:        "gateway1",
							Namespace:   ptr.To[gwapiv1.Namespace]("default"),
							SectionName: ptr.To[gwapiv1.SectionName]("listener1"),
							Port:        ptr.To[gwapiv1.PortNumber](80),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
								Status: metav1.ConditionTrue,
								Reason: "ResolvedRefs",
							},
						},
					},
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:        "gateway1",
						Namespace:   ptr.To[gwapiv1.Namespace]("default"),
						SectionName: ptr.To[gwapiv1.SectionName]("listener1"),
						Port:        ptr.To[gwapiv1.PortNumber](80),
					},
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.RouteConditionAccepted),
							Status: metav1.ConditionTrue,
							Reason: "Accepted",
						},
						{
							Type:   string(gwapiv1.RouteConditionResolvedRefs),
							Status: metav1.ConditionTrue,
							Reason: "ResolvedRefs",
						},
					},
				},
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:      "gateway2",
						Namespace: ptr.To[gwapiv1.Namespace]("default"),
					},
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.RouteConditionAccepted),
							Status: metav1.ConditionFalse,
							Reason: "SomeReason",
						},
					},
				},
			},
		},

		{
			name: "override an existing parent",
			args: args{
				ns: "default",
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway1",
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
								Status: metav1.ConditionTrue,
								Reason: "ResolvedRefs",
							},
						},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:      "gateway2",
							Namespace: ptr.To[gwapiv1.Namespace]("default"),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
								Status: metav1.ConditionTrue,
								Reason: "ResolvedRefs",
							},
						},
					},
				},
				new: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:      "gateway2",
							Namespace: ptr.To[gwapiv1.Namespace]("default"),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionFalse,
								Reason: "SomeReason",
							},
						},
					},
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:      "gateway2",
						Namespace: ptr.To[gwapiv1.Namespace]("default"),
					},
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.RouteConditionAccepted),
							Status: metav1.ConditionFalse,
							Reason: "SomeReason",
						},
					},
				},
			},
		},

		{
			name: "nothing changed",
			args: args{
				ns: "default",
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway1",
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
								Status: metav1.ConditionTrue,
								Reason: "ResolvedRefs",
							},
						},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway2",
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionFalse,
								Reason: "SomeReason",
							},
						},
					},
				},
				new: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:      "gateway2",
							Namespace: ptr.To[gwapiv1.Namespace]("default"),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionFalse,
								Reason: "SomeReason",
							},
						},
					},
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:      "gateway2",
						Namespace: ptr.To[gwapiv1.Namespace]("default"),
					},
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.RouteConditionAccepted),
							Status: metav1.ConditionFalse,
							Reason: "SomeReason",
						},
					},
				},
			},
		},
		{
			name: "remove old entries that are no longer referenced",
			args: args{
				ns: "default",
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:      "gateway1",
							Namespace: ptr.To[gwapiv1.Namespace]("default"),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
						},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:      "gateway2",
							Namespace: ptr.To[gwapiv1.Namespace]("default"),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
						},
					},
				},
				new: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:      "gateway1",
							Namespace: ptr.To[gwapiv1.Namespace]("default"),
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
						},
					},
					// gateway2 is removed from spec.parentRefs, so it should not appear in the merged result
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:      "gateway1",
						Namespace: ptr.To[gwapiv1.Namespace]("default"),
					},
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.RouteConditionAccepted),
							Status: metav1.ConditionTrue,
							Reason: "Accepted",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeRouteParentStatus(tt.args.ns, tt.args.old, tt.args.new); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeRouteParentStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
