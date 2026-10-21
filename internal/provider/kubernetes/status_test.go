// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/envoygateway"
)

func TestEnvoyEndpointNodeNamesForService(t *testing.T) {
	ns := "envoy-gateway-system"
	svcName := "envoy-svc"

	testCases := []struct {
		name           string
		endpointSlices []discoveryv1.EndpointSlice
		wantNodes      []string
	}{
		{
			name:      "no EndpointSlices returns empty",
			wantNodes: []string{},
		},
		{
			name: "ready endpoint included",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eps-1",
						Namespace: ns,
						Labels:    map[string]string{discoveryv1.LabelServiceName: svcName},
					},
					Endpoints: []discoveryv1.Endpoint{
						{NodeName: new("node1"), Conditions: discoveryv1.EndpointConditions{Ready: new(true)}},
					},
				},
			},
			wantNodes: []string{"node1"},
		},
		{
			name: "nil Ready treated as ready (k8s backward compat)",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eps-1",
						Namespace: ns,
						Labels:    map[string]string{discoveryv1.LabelServiceName: svcName},
					},
					Endpoints: []discoveryv1.Endpoint{
						{NodeName: new("node1"), Conditions: discoveryv1.EndpointConditions{Ready: nil}},
					},
				},
			},
			wantNodes: []string{"node1"},
		},
		{
			name: "not-ready endpoint excluded",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eps-1",
						Namespace: ns,
						Labels:    map[string]string{discoveryv1.LabelServiceName: svcName},
					},
					Endpoints: []discoveryv1.Endpoint{
						{NodeName: new("node1"), Conditions: discoveryv1.EndpointConditions{Ready: new(false)}},
					},
				},
			},
			wantNodes: []string{},
		},
		{
			name: "endpoint with empty NodeName excluded",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eps-1",
						Namespace: ns,
						Labels:    map[string]string{discoveryv1.LabelServiceName: svcName},
					},
					Endpoints: []discoveryv1.Endpoint{
						{NodeName: new(""), Conditions: discoveryv1.EndpointConditions{Ready: new(true)}},
						{NodeName: nil, Conditions: discoveryv1.EndpointConditions{Ready: new(true)}},
					},
				},
			},
			wantNodes: []string{},
		},
		{
			name: "duplicate nodes across multiple EndpointSlices deduplicated",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eps-1",
						Namespace: ns,
						Labels:    map[string]string{discoveryv1.LabelServiceName: svcName},
					},
					Endpoints: []discoveryv1.Endpoint{
						{NodeName: new("node1"), Conditions: discoveryv1.EndpointConditions{Ready: new(true)}},
						{NodeName: new("node2"), Conditions: discoveryv1.EndpointConditions{Ready: new(true)}},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eps-2",
						Namespace: ns,
						Labels:    map[string]string{discoveryv1.LabelServiceName: svcName},
					},
					Endpoints: []discoveryv1.Endpoint{
						{NodeName: new("node1"), Conditions: discoveryv1.EndpointConditions{Ready: new(true)}},
					},
				},
			},
			wantNodes: []string{"node1", "node2"},
		},
		{
			name: "EndpointSlice for different service ignored",
			endpointSlices: []discoveryv1.EndpointSlice{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "eps-other",
						Namespace: ns,
						Labels:    map[string]string{discoveryv1.LabelServiceName: "other-svc"},
					},
					Endpoints: []discoveryv1.Endpoint{
						{NodeName: new("node1"), Conditions: discoveryv1.EndpointConditions{Ready: new(true)}},
					},
				},
			},
			wantNodes: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme())
			for i := range tc.endpointSlices {
				builder = builder.WithObjects(&tc.endpointSlices[i])
			}
			fakeClient := builder.Build()

			r := &gatewayAPIReconciler{client: fakeClient}

			svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: svcName, Namespace: ns}}
			got, err := r.envoyEndpointNodeNamesForService(context.Background(), svc)
			require.NoError(t, err)

			sort.Strings(got)
			sort.Strings(tc.wantNodes)
			require.Equal(t, tc.wantNodes, got)
		})
	}
}

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
		{
			name: "another controller's status for the same parentRef is preserved",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue, Reason: "Accepted"},
						},
					},
					{
						ControllerName: "example.com/policy-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: "example.com/PolicyAffected", Status: metav1.ConditionTrue, Reason: "Accepted"},
						},
					},
				},
				new: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionFalse, Reason: "SomeReason"},
						},
					},
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
					Conditions: []metav1.Condition{
						{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionFalse, Reason: "SomeReason"},
					},
				},
				{
					ControllerName: "example.com/policy-controller",
					ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
					Conditions: []metav1.Condition{
						{Type: "example.com/PolicyAffected", Status: metav1.ConditionTrue, Reason: "Accepted"},
					},
				},
			},
		},
		{
			name: "duplicate entries of ours for the same parentRef are collapsed",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue, Reason: "Accepted"},
						},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue, Reason: "Accepted"},
						},
					},
					{
						ControllerName: "example.com/policy-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: "example.com/PolicyAffected", Status: metav1.ConditionTrue, Reason: "Accepted"},
						},
					},
				},
				new: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue, Reason: "Accepted"},
						},
					},
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
					Conditions: []metav1.Condition{
						{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue, Reason: "Accepted"},
					},
				},
				{
					ControllerName: "example.com/policy-controller",
					ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
					Conditions: []metav1.Condition{
						{Type: "example.com/PolicyAffected", Status: metav1.ConditionTrue, Reason: "Accepted"},
					},
				},
			},
		},
		{
			name: "only another controller's status exists for the parentRef - ours is appended",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "example.com/policy-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: "example.com/PolicyAffected", Status: metav1.ConditionTrue, Reason: "Accepted"},
						},
					},
				},
				new: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
						Conditions: []metav1.Condition{
							{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue, Reason: "Accepted"},
						},
					},
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "example.com/policy-controller",
					ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
					Conditions: []metav1.Condition{
						{Type: "example.com/PolicyAffected", Status: metav1.ConditionTrue, Reason: "Accepted"},
					},
				},
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
					Conditions: []metav1.Condition{
						{Type: string(gwapiv1.RouteConditionAccepted), Status: metav1.ConditionTrue, Reason: "Accepted"},
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

func Test_mergePolicyAncestorStatus(t *testing.T) {
	const (
		ours  = gwapiv1.GatewayController("gateway.envoyproxy.io/gatewayclass-controller")
		other = gwapiv1.GatewayController("example.com/other-controller")
	)
	ancestor := func(controller gwapiv1.GatewayController, name string, condType string, condStatus metav1.ConditionStatus) gwapiv1.PolicyAncestorStatus {
		return gwapiv1.PolicyAncestorStatus{
			ControllerName: controller,
			AncestorRef:    gwapiv1.ParentReference{Name: gwapiv1.ObjectName(name)},
			Conditions: []metav1.Condition{
				{Type: condType, Status: condStatus, Reason: "Reason"},
			},
		}
	}
	accepted := string(gwapiv1.PolicyConditionAccepted)
	foreign := "example.com/Audited"

	tests := []struct {
		name string
		old  []gwapiv1.PolicyAncestorStatus
		new  []gwapiv1.PolicyAncestorStatus
		want []gwapiv1.PolicyAncestorStatus
	}{
		{
			name: "no old status",
			new:  []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionTrue)},
			want: []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionTrue)},
		},
		{
			name: "our entry is updated",
			old:  []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionTrue)},
			new:  []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionFalse)},
			want: []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionFalse)},
		},
		{
			name: "stale entry of ours is removed",
			old: []gwapiv1.PolicyAncestorStatus{
				ancestor(ours, "gateway1", accepted, metav1.ConditionTrue),
				ancestor(ours, "gateway2", accepted, metav1.ConditionTrue),
			},
			new:  []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway2", accepted, metav1.ConditionTrue)},
			want: []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway2", accepted, metav1.ConditionTrue)},
		},
		{
			name: "all entries of ours are removed when the new status is empty",
			old: []gwapiv1.PolicyAncestorStatus{
				ancestor(ours, "gateway1", accepted, metav1.ConditionTrue),
				ancestor(other, "gateway1", foreign, metav1.ConditionTrue),
			},
			want: []gwapiv1.PolicyAncestorStatus{ancestor(other, "gateway1", foreign, metav1.ConditionTrue)},
		},
		{
			name: "another controller's entries are preserved in place",
			old: []gwapiv1.PolicyAncestorStatus{
				ancestor(other, "gateway0", foreign, metav1.ConditionTrue),
				ancestor(ours, "gateway1", accepted, metav1.ConditionTrue),
				ancestor(other, "gateway1", foreign, metav1.ConditionTrue),
			},
			new: []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionFalse)},
			want: []gwapiv1.PolicyAncestorStatus{
				ancestor(other, "gateway0", foreign, metav1.ConditionTrue),
				ancestor(ours, "gateway1", accepted, metav1.ConditionFalse),
				ancestor(other, "gateway1", foreign, metav1.ConditionTrue),
			},
		},
		{
			name: "only another controller's entry exists for the ancestorRef - ours is appended",
			old:  []gwapiv1.PolicyAncestorStatus{ancestor(other, "gateway1", foreign, metav1.ConditionTrue)},
			new:  []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionTrue)},
			want: []gwapiv1.PolicyAncestorStatus{
				ancestor(other, "gateway1", foreign, metav1.ConditionTrue),
				ancestor(ours, "gateway1", accepted, metav1.ConditionTrue),
			},
		},
		{
			name: "duplicate entries of ours are collapsed",
			old: []gwapiv1.PolicyAncestorStatus{
				ancestor(ours, "gateway1", accepted, metav1.ConditionTrue),
				ancestor(ours, "gateway1", accepted, metav1.ConditionTrue),
			},
			new:  []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionTrue)},
			want: []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionTrue)},
		},
		{
			name: "ancestorRef with explicit default namespace matches",
			old: []gwapiv1.PolicyAncestorStatus{
				{
					ControllerName: ours,
					AncestorRef:    gwapiv1.ParentReference{Name: "gateway1", Namespace: new(gwapiv1.Namespace("default"))},
					Conditions:     []metav1.Condition{{Type: accepted, Status: metav1.ConditionTrue, Reason: "Reason"}},
				},
				ancestor(other, "gateway2", foreign, metav1.ConditionTrue),
			},
			new: []gwapiv1.PolicyAncestorStatus{ancestor(ours, "gateway1", accepted, metav1.ConditionFalse)},
			want: []gwapiv1.PolicyAncestorStatus{
				ancestor(ours, "gateway1", accepted, metav1.ConditionFalse),
				ancestor(other, "gateway2", foreign, metav1.ConditionTrue),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergePolicyAncestorStatus("default", ours, tt.old, tt.new)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_unstructuredPolicyAncestors(t *testing.T) {
	policy := func(status any) *unstructured.Unstructured {
		obj := map[string]any{
			"apiVersion": "example.com/v1alpha1",
			"kind":       "ExampleExtPolicy",
			"metadata":   map[string]any{"name": "policy", "namespace": "default"},
		}
		if status != nil {
			obj["status"] = status
		}
		return &unstructured.Unstructured{Object: obj}
	}

	tests := []struct {
		name string
		obj  *unstructured.Unstructured
		want []gwapiv1.PolicyAncestorStatus
	}{
		{
			name: "no status",
			obj:  policy(nil),
		},
		{
			name: "status is not a map",
			obj:  policy("invalid"),
		},
		{
			name: "status is not a PolicyStatus",
			obj:  policy(map[string]any{"ancestors": "invalid"}),
		},
		{
			name: "ancestors are decoded",
			// As decoded from the API server's JSON response.
			obj: policy(map[string]any{
				"ancestors": []any{
					map[string]any{
						"ancestorRef":    map[string]any{"group": "gateway.networking.k8s.io", "kind": "Gateway", "name": "gateway1"},
						"controllerName": "example.com/other-controller",
						"conditions": []any{
							map[string]any{
								"type":               "example.com/Audited",
								"status":             "True",
								"reason":             "Checked",
								"message":            "added by another controller",
								"observedGeneration": int64(1),
								"lastTransitionTime": "2026-01-01T00:00:00Z",
							},
						},
					},
				},
			}),
			want: []gwapiv1.PolicyAncestorStatus{
				{
					AncestorRef: gwapiv1.ParentReference{
						Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
						Kind:  new(gwapiv1.Kind("Gateway")),
						Name:  "gateway1",
					},
					ControllerName: "example.com/other-controller",
					Conditions: []metav1.Condition{
						{
							Type:               "example.com/Audited",
							Status:             metav1.ConditionTrue,
							Reason:             "Checked",
							Message:            "added by another controller",
							ObservedGeneration: 1,
							LastTransitionTime: metav1.NewTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Local()),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, unstructuredPolicyAncestors(tt.obj))
		})
	}
}
