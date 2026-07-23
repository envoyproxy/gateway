// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/logging"
)

func Test_mergeRouteParentStatus(t *testing.T) {
	type args struct {
		old []gwapiv1.RouteParentStatus
		new []gwapiv1.RouteParentStatus
	}
	tests := []struct {
		name string
		args args
		want []gwapiv1.RouteParentStatus
	}{
		{
			name: "old contains one parentRef of ours and one of another controller's, status of ours changed in new.",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "istio.io/gateway-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:        "gateway1",
							Namespace:   new(gwapiv1.Namespace("default")),
							SectionName: new(gwapiv1.SectionName("listener1")),
							Port:        new(gwapiv1.PortNumber(80)),
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
							Name: "gateway2",
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
								Status: metav1.ConditionFalse,
								Reason: "SomeReason",
							},
						},
					},
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "istio.io/gateway-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:        "gateway1",
						Namespace:   new(gwapiv1.Namespace("default")),
						SectionName: new(gwapiv1.SectionName("listener1")),
						Port:        new(gwapiv1.PortNumber(80)),
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
							Status: metav1.ConditionTrue,
							Reason: "Accepted",
						},
						{
							Type:   string(gwapiv1.RouteConditionResolvedRefs),
							Status: metav1.ConditionFalse,
							Reason: "SomeReason",
						},
					},
				},
			},
		},
		{
			name: "old contains one parentRef of ours and one of another controller's, status of ours changed in new with an additional parentRef of ours",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "istio.io/gateway-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:        "gateway1",
							Namespace:   new(gwapiv1.Namespace("default")),
							SectionName: new(gwapiv1.SectionName("listener1")),
							Port:        new(gwapiv1.PortNumber(80)),
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
							Name: "gateway2",
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
								Status: metav1.ConditionFalse,
								Reason: "SomeReason",
							},
						},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway3",
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
					ControllerName: "istio.io/gateway-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:        "gateway1",
						Namespace:   new(gwapiv1.Namespace("default")),
						SectionName: new(gwapiv1.SectionName("listener1")),
						Port:        new(gwapiv1.PortNumber(80)),
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
							Status: metav1.ConditionTrue,
							Reason: "Accepted",
						},
						{
							Type:   string(gwapiv1.RouteConditionResolvedRefs),
							Status: metav1.ConditionFalse,
							Reason: "SomeReason",
						},
					},
				},
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef: gwapiv1.ParentReference{
						Name: "gateway3",
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
		{
			name: "old contains one parentRef of ours and one of another controller's, ours gets dropped in new and a different parentRef of ours is added",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "istio.io/gateway-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:        "gateway1",
							Namespace:   new(gwapiv1.Namespace("default")),
							SectionName: new(gwapiv1.SectionName("listener1")),
							Port:        new(gwapiv1.PortNumber(80)),
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
							Name: "gateway3",
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
					ControllerName: "istio.io/gateway-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:        "gateway1",
						Namespace:   new(gwapiv1.Namespace("default")),
						SectionName: new(gwapiv1.SectionName("listener1")),
						Port:        new(gwapiv1.PortNumber(80)),
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
						Name: "gateway3",
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
		// Practically this will never occur, since having no parentRefs in the new
		// status means the route doesn't attach (in the spec) to any of our gateways.
		//
		// But then we'd consider it irrelevant before ever computing such status for it, i.e, the
		// route will forever have a dangling status parentRef referencing us that will not be removed.
		//
		// TODO: maybe this needs to be fixed.
		{
			name: "old contains one parentRef of ours and one of another controller's, ours gets dropped in new.",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "istio.io/gateway-controller",
						ParentRef: gwapiv1.ParentReference{
							Name:        "gateway1",
							Namespace:   new(gwapiv1.Namespace("default")),
							SectionName: new(gwapiv1.SectionName("listener1")),
							Port:        new(gwapiv1.PortNumber(80)),
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
				new: []gwapiv1.RouteParentStatus{},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "istio.io/gateway-controller",
					ParentRef: gwapiv1.ParentReference{
						Name:        "gateway1",
						Namespace:   new(gwapiv1.Namespace("default")),
						SectionName: new(gwapiv1.SectionName("listener1")),
						Port:        new(gwapiv1.PortNumber(80)),
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

		{
			name: "old contains one parentRef of ours, status of ours changed in new.",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway2",
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
							Name: "gateway2",
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
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
						Name: "gateway2",
					},
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.RouteConditionAccepted),
							Status: metav1.ConditionTrue,
							Reason: "Accepted",
						},
						{
							Type:   string(gwapiv1.RouteConditionResolvedRefs),
							Status: metav1.ConditionFalse,
							Reason: "SomeReason",
						},
					},
				},
			},
		},
		{
			name: "old contains one parentRef of ours, status of ours changed in new with an additional parentRef of ours",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway2",
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
							Name: "gateway2",
						},
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: "Accepted",
							},
							{
								Type:   string(gwapiv1.RouteConditionResolvedRefs),
								Status: metav1.ConditionFalse,
								Reason: "SomeReason",
							},
						},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway3",
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
						Name: "gateway2",
					},
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.RouteConditionAccepted),
							Status: metav1.ConditionTrue,
							Reason: "Accepted",
						},
						{
							Type:   string(gwapiv1.RouteConditionResolvedRefs),
							Status: metav1.ConditionFalse,
							Reason: "SomeReason",
						},
					},
				},
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef: gwapiv1.ParentReference{
						Name: "gateway3",
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
		{
			name: "old contains one parentRef of ours, ours gets dropped in new and a different parentRef of ours is added",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway2",
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
							Name: "gateway3",
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
						Name: "gateway2",
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
						Name: "gateway3",
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
		// Test that parent refs managed by our controller are preserved even when not in new update.
		// This is important for routes with multiple parent references.
		{
			name: "old contains one parentRef of ours, and it's not in new - should be preserved.",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef: gwapiv1.ParentReference{
							Name: "gateway2",
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
				new: []gwapiv1.RouteParentStatus{},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef: gwapiv1.ParentReference{
						Name: "gateway2",
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
		// Test multi-parent scenario where only one parent is updated at a time.
		{
			name: "multiple parents from same controller - update one, preserve others",
			args: args{
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
				},
			},
			want: []gwapiv1.RouteParentStatus{
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
			if got := mergeRouteParentStatus("default", tt.args.old, tt.args.new); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeRouteParentStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removePolicyStatusForController(t *testing.T) {
	const ours = gwapiv1.GatewayController("gateway.envoyproxy.io/gatewayclass-controller")
	const theirs = gwapiv1.GatewayController("istio.io/gateway-controller")

	ancestor := func(controller gwapiv1.GatewayController, name string) gwapiv1.PolicyAncestorStatus {
		return gwapiv1.PolicyAncestorStatus{
			ControllerName: controller,
			AncestorRef:    gwapiv1.ParentReference{Name: gwapiv1.ObjectName(name)},
			Conditions: []metav1.Condition{
				{
					Type:               string(gwapiv1.PolicyConditionAccepted),
					Status:             metav1.ConditionTrue,
					Reason:             string(gwapiv1.PolicyReasonAccepted),
					ObservedGeneration: 1,
				},
			},
		}
	}

	tests := []struct {
		name       string
		ancestors  []gwapiv1.PolicyAncestorStatus
		controller gwapiv1.GatewayController
		want       []gwapiv1.PolicyAncestorStatus
	}{
		{
			// #8926 / #8927: policy attached to nothing this pass, its only ancestor
			// is ours and stale -> it gets cleared so the stale Accepted=True is gone.
			name:       "single stale ancestor of ours is cleared",
			ancestors:  []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1")},
			controller: ours,
			want:       nil,
		},
		{
			// Multi-controller: only our ancestor is removed, the other controller's
			// status is preserved untouched.
			name:       "only our ancestor removed, other controller preserved",
			ancestors:  []gwapiv1.PolicyAncestorStatus{ancestor(theirs, "gateway1"), ancestor(ours, "gateway2")},
			controller: ours,
			want:       []gwapiv1.PolicyAncestorStatus{ancestor(theirs, "gateway1")},
		},
		{
			name:       "no ancestors of ours leaves status untouched",
			ancestors:  []gwapiv1.PolicyAncestorStatus{ancestor(theirs, "gateway1")},
			controller: ours,
			want:       []gwapiv1.PolicyAncestorStatus{ancestor(theirs, "gateway1")},
		},
		{
			name:       "empty input returns empty",
			ancestors:  nil,
			controller: ours,
			want:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removePolicyStatusForController(tt.ancestors, tt.controller); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removePolicyStatusForController() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestBackendTrafficPolicyStatusCleanupOnDelete exercises the delete-path cleanup
// end-to-end through the real UpdateHandler.apply against a fake client. It covers
// both filed issues (a BackendTrafficPolicy whose target went missing, #8926, and one
// whose selectors stopped matching, #8927) since both surface identically: the object
// carries a stale Accepted=True condition with an out-of-date observedGeneration that
// must be removed once the translator stops producing status for it.
func TestBackendTrafficPolicyStatusCleanupOnDelete(t *testing.T) {
	const ours = gwapiv1.GatewayController("gateway.envoyproxy.io/gatewayclass-controller")
	const theirs = gwapiv1.GatewayController("istio.io/gateway-controller")

	staleAncestor := func(controller gwapiv1.GatewayController, name string) gwapiv1.PolicyAncestorStatus {
		return gwapiv1.PolicyAncestorStatus{
			ControllerName: controller,
			AncestorRef:    gwapiv1.ParentReference{Name: gwapiv1.ObjectName(name)},
			Conditions: []metav1.Condition{{
				Type:               string(gwapiv1.PolicyConditionAccepted),
				Status:             metav1.ConditionTrue,
				Reason:             string(gwapiv1.PolicyReasonAccepted),
				ObservedGeneration: 1, // stale: object generation below is 2
			}},
		}
	}

	tests := []struct {
		name          string
		existing      *egv1a1.BackendTrafficPolicy
		key           types.NamespacedName
		wantAncestors []gwapiv1.PolicyAncestorStatus
	}{
		{
			name: "stale status owned solely by our controller is cleared",
			existing: &egv1a1.BackendTrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "btp-1", Generation: 2},
				Status:     gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{staleAncestor(ours, "gateway-1")}},
			},
			key:           types.NamespacedName{Namespace: "default", Name: "btp-1"},
			wantAncestors: nil,
		},
		{
			name: "only our ancestor is cleared, other controller's status preserved",
			existing: &egv1a1.BackendTrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "btp-2", Generation: 2},
				Status: gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{
					staleAncestor(theirs, "gateway-1"),
					staleAncestor(ours, "gateway-2"),
				}},
			},
			key:           types.NamespacedName{Namespace: "default", Name: "btp-2"},
			wantAncestors: []gwapiv1.PolicyAncestorStatus{staleAncestor(theirs, "gateway-1")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithObjects(tt.existing).
				WithStatusSubresource(tt.existing).
				Build()

			logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
			u := NewUpdateHandler(logger.Logger, cli)
			r := &gatewayAPIReconciler{
				log:             logger,
				classController: ours,
			}

			u.apply(r.backendTrafficPolicyStatusCleanupUpdate(tt.key, nil))

			got := &egv1a1.BackendTrafficPolicy{}
			require.NoError(t, cli.Get(context.Background(), tt.key, got))
			require.Equal(t, tt.wantAncestors, got.Status.Ancestors)
		})
	}

	t.Run("delete for an already-removed object is a no-op", func(t *testing.T) {
		cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
		logger := logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo)
		u := NewUpdateHandler(logger.Logger, cli)
		r := &gatewayAPIReconciler{
			log:             logger,
			classController: ours,
		}

		// Must not panic or error when the object no longer exists.
		u.apply(r.backendTrafficPolicyStatusCleanupUpdate(types.NamespacedName{Namespace: "default", Name: "gone"}, nil))

		got := &egv1a1.BackendTrafficPolicy{}
		err := cli.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "gone"}, got)
		require.Error(t, err) // still not found
	})
}
