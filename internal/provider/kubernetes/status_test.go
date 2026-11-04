// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
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
		old            []gwapiv1.RouteParentStatus
		new            []gwapiv1.RouteParentStatus
		specParentRefs []gwapiv1.ParentReference
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
				specParentRefs: []gwapiv1.ParentReference{
					{
						Name:        "gateway1",
						Namespace:   new(gwapiv1.Namespace("default")),
						SectionName: new(gwapiv1.SectionName("listener1")),
						Port:        new(gwapiv1.PortNumber(80)),
					},
					{Name: "gateway2"},
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
				specParentRefs: []gwapiv1.ParentReference{
					{
						Name:        "gateway1",
						Namespace:   new(gwapiv1.Namespace("default")),
						SectionName: new(gwapiv1.SectionName("listener1")),
						Port:        new(gwapiv1.PortNumber(80)),
					},
					{Name: "gateway2"},
					{Name: "gateway3"},
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
				specParentRefs: []gwapiv1.ParentReference{
					{
						Name:        "gateway1",
						Namespace:   new(gwapiv1.Namespace("default")),
						SectionName: new(gwapiv1.SectionName("listener1")),
						Port:        new(gwapiv1.PortNumber(80)),
					},
					{Name: "gateway2"},
					{Name: "gateway3"},
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
		// Regression test for https://github.com/envoyproxy/gateway/issues/774: a parentRef
		// that has been removed from the route's spec must not leave a stale status entry
		// behind, even though other, still-referenced, parentRefs are preserved.
		{
			name: "old contains one parentRef of ours and one of another controller's, ours is removed from the route's spec.",
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
				// gateway2 was removed from the route's spec.parentRefs; gateway1 (managed by
				// another controller) is still referenced.
				specParentRefs: []gwapiv1.ParentReference{
					{
						Name:        "gateway1",
						Namespace:   new(gwapiv1.Namespace("default")),
						SectionName: new(gwapiv1.SectionName("listener1")),
						Port:        new(gwapiv1.PortNumber(80)),
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
			},
		},

		{
			// Status entries owned by another controller must never be dropped by us, even if
			// their parentRef is no longer in the route's spec - that controller manages them.
			name: "old contains another controller's parentRef that was removed from the route's spec - should be preserved.",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "istio.io/gateway-controller",
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
				},
				new: []gwapiv1.RouteParentStatus{},
				specParentRefs: []gwapiv1.ParentReference{
					{Name: "gateway2"},
				},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "istio.io/gateway-controller",
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
			},
		},
		{
			// Regression test: new was computed while gateway2 was still in spec.ParentRefs, but by the time
			// the merge runs the parentRef has already been removed from spec. The stale entry in new must be
			// dropped rather than leaking into the merged status.
			name: "new contains a parentRef of ours that has since been removed from the route's spec - should be dropped.",
			args: args{
				old: []gwapiv1.RouteParentStatus{},
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
								Status: metav1.ConditionTrue,
								Reason: "ResolvedRefs",
							},
						},
					},
				},
				specParentRefs: []gwapiv1.ParentReference{},
			},
			want: []gwapiv1.RouteParentStatus{},
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
				specParentRefs: []gwapiv1.ParentReference{
					{Name: "gateway2"},
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
				specParentRefs: []gwapiv1.ParentReference{
					{Name: "gateway2"},
					{Name: "gateway3"},
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
				specParentRefs: []gwapiv1.ParentReference{
					{Name: "gateway2"},
					{Name: "gateway3"},
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
				specParentRefs: []gwapiv1.ParentReference{
					{Name: "gateway2"},
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
				specParentRefs: []gwapiv1.ParentReference{
					{Name: "gateway1"},
					{Name: "gateway2"},
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
			name: "parentRef in both old and new but removed from spec - dropped",
			args: args{
				old: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
					},
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway2"},
					},
				},
				new: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
					},
				},
				specParentRefs: []gwapiv1.ParentReference{{Name: "gateway2"}},
			},
			want: []gwapiv1.RouteParentStatus{
				{
					ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
					ParentRef:      gwapiv1.ParentReference{Name: "gateway2"},
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
				specParentRefs: []gwapiv1.ParentReference{{Name: "gateway1"}},
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
				specParentRefs: []gwapiv1.ParentReference{{Name: "gateway1"}},
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
				specParentRefs: []gwapiv1.ParentReference{{Name: "gateway1"}},
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
			if got := mergeRouteParentStatus("default", egv1a1.GatewayControllerName, tt.args.old, tt.args.new, tt.args.specParentRefs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeRouteParentStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

// fakeUpdater is a test Updater that forwards every Send to a channel, so a test can
// inspect the Update (and, in particular, invoke its Mutator) synchronously.
type fakeUpdater struct {
	updates chan Update
}

func (f *fakeUpdater) Send(_ context.Context, u Update) {
	f.updates <- u
}

// Test_updateStatusFromSubscriptions_HTTPRoute is a regression test for
// https://github.com/envoyproxy/gateway/issues/774: it drives an HTTPRoute status update
// through the real subscription/mutator wiring (not just mergeRouteParentStatus in
// isolation), confirming that the route's live spec.ParentRefs is threaded through
// correctly so a status entry for a parentRef removed from spec.ParentRefs is dropped.
func Test_updateStatusFromSubscriptions_HTTPRoute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	key := types.NamespacedName{Namespace: "default", Name: "httproute-1"}

	// This reconciliation batch computed no parent status at all for the route (e.g.
	// gateway1 no longer matches now that it's gone from spec.ParentRefs, and gateway2's
	// status hasn't been computed in this particular batch).
	var httpRouteStatuses watchable.Map[types.NamespacedName, *gwapiv1.HTTPRouteStatus]
	httpRouteStatuses.Store(key, &gwapiv1.HTTPRouteStatus{
		RouteStatus: gwapiv1.RouteStatus{Parents: []gwapiv1.RouteParentStatus{}},
	})

	updater := &fakeUpdater{updates: make(chan Update, 1)}
	r := &gatewayAPIReconciler{
		log:             logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		classController: egv1a1.GatewayControllerName,
		statusUpdater:   updater,
		subscriptions: &subscriptions{
			httpRouteStatuses: httpRouteStatuses.Subscribe(ctx),
		},
	}

	go r.updateStatusFromSubscriptions(ctx, false)

	var update Update
	select {
	case update = <-updater.updates:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for status update")
	}

	if update.NamespacedName != key {
		t.Fatalf("update.NamespacedName = %v, want %v", update.NamespacedName, key)
	}

	// gateway1 has been removed from the route's spec.ParentRefs; gateway2 is the only
	// remaining live parentRef and isn't part of this reconciliation batch's new status.
	route := &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
		Spec: gwapiv1.HTTPRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{{Name: "gateway2"}},
			},
		},
		Status: gwapiv1.HTTPRouteStatus{
			RouteStatus: gwapiv1.RouteStatus{
				Parents: []gwapiv1.RouteParentStatus{
					{
						ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
						ParentRef:      gwapiv1.ParentReference{Name: "gateway1"},
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
		},
	}

	mutatedObj := update.Mutator.Mutate(route)
	mutated, ok := mutatedObj.(*gwapiv1.HTTPRoute)
	if !ok {
		t.Fatalf("Mutate() returned unexpected type %T", mutatedObj)
	}

	if len(mutated.Status.Parents) != 0 {
		t.Errorf("mutated.Status.Parents = %v, want empty (stale gateway1 entry should be dropped)", mutated.Status.Parents)
	}
}

func policyAncestor(controllerName gwapiv1.GatewayController, name gwapiv1.ObjectName, conditionType string) gwapiv1.PolicyAncestorStatus {
	return gwapiv1.PolicyAncestorStatus{
		AncestorRef:    gwapiv1.ParentReference{Name: name},
		ControllerName: controllerName,
		Conditions: []metav1.Condition{
			{Type: conditionType, Status: metav1.ConditionTrue, Reason: "Accepted"},
		},
	}
}

func Test_mergePolicyAncestorStatus(t *testing.T) {
	const otherController gwapiv1.GatewayController = "example.com/policy-controller"
	ours := func(name gwapiv1.ObjectName) gwapiv1.PolicyAncestorStatus {
		return policyAncestor(egv1a1.GatewayControllerName, name, string(gwapiv1.PolicyConditionAccepted))
	}
	other := func(name gwapiv1.ObjectName) gwapiv1.PolicyAncestorStatus {
		return policyAncestor(otherController, name, "example.com/PolicyAffected")
	}
	oursN := func(n int) []gwapiv1.PolicyAncestorStatus {
		ret := make([]gwapiv1.PolicyAncestorStatus, 0, n)
		for i := range n {
			ret = append(ret, ours(gwapiv1.ObjectName(fmt.Sprintf("gateway-%d", i))))
		}
		return ret
	}
	othersN := func(n int) []gwapiv1.PolicyAncestorStatus {
		ret := make([]gwapiv1.PolicyAncestorStatus, 0, n)
		for i := range n {
			ret = append(ret, other(gwapiv1.ObjectName(fmt.Sprintf("other-%d", i))))
		}
		return ret
	}

	tests := []struct {
		name string
		old  []gwapiv1.PolicyAncestorStatus
		new  []gwapiv1.PolicyAncestorStatus
		want []gwapiv1.PolicyAncestorStatus
	}{
		{
			name: "no old status",
			old:  nil,
			new:  []gwapiv1.PolicyAncestorStatus{ours("gateway1")},
			want: []gwapiv1.PolicyAncestorStatus{ours("gateway1")},
		},
		{
			name: "only our entries: new replaces old, stale entries are dropped",
			old:  []gwapiv1.PolicyAncestorStatus{ours("gateway1"), ours("gateway2")},
			new:  []gwapiv1.PolicyAncestorStatus{ours("gateway2")},
			want: []gwapiv1.PolicyAncestorStatus{ours("gateway2")},
		},
		{
			name: "other controller's entry for the same ancestor is kept",
			old:  []gwapiv1.PolicyAncestorStatus{ours("gateway1"), other("gateway1")},
			new:  []gwapiv1.PolicyAncestorStatus{ours("gateway1")},
			want: []gwapiv1.PolicyAncestorStatus{ours("gateway1"), other("gateway1")},
		},
		{
			name: "our entry is added next to an entry written only by another controller",
			old:  []gwapiv1.PolicyAncestorStatus{other("gateway1")},
			new:  []gwapiv1.PolicyAncestorStatus{ours("gateway1")},
			want: []gwapiv1.PolicyAncestorStatus{ours("gateway1"), other("gateway1")},
		},
		{
			name: "other controller's entries for other ancestors are kept",
			old:  []gwapiv1.PolicyAncestorStatus{other("gateway2"), ours("gateway1"), other("gateway3")},
			new:  []gwapiv1.PolicyAncestorStatus{ours("gateway1")},
			want: []gwapiv1.PolicyAncestorStatus{ours("gateway1"), other("gateway2"), other("gateway3")},
		},
		{
			name: "duplicated entries of ours collapse",
			old:  []gwapiv1.PolicyAncestorStatus{ours("gateway1"), ours("gateway1"), other("gateway1")},
			new:  []gwapiv1.PolicyAncestorStatus{ours("gateway1")},
			want: []gwapiv1.PolicyAncestorStatus{ours("gateway1"), other("gateway1")},
		},
		{
			name: "empty new status drops only our entries",
			old:  []gwapiv1.PolicyAncestorStatus{ours("gateway1"), other("gateway1")},
			new:  []gwapiv1.PolicyAncestorStatus{},
			want: []gwapiv1.PolicyAncestorStatus{other("gateway1")},
		},
		{
			name: "our entries are trimmed to stay within the CRD limit",
			old:  othersN(4),
			new:  oursN(16),
			want: append(oursN(12), othersN(4)...),
		},
		{
			name: "no room left for our entries",
			old:  othersN(16),
			new:  oursN(2),
			want: othersN(16),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergePolicyAncestorStatus(egv1a1.GatewayControllerName, tt.old, tt.new)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_mergeUnstructuredPolicyStatus(t *testing.T) {
	const otherController gwapiv1.GatewayController = "example.com/policy-controller"
	ours := policyAncestor(egv1a1.GatewayControllerName, "gateway1", string(gwapiv1.PolicyConditionAccepted))
	other := policyAncestor(otherController, "gateway1", "example.com/PolicyAffected")

	toMap := func(t *testing.T, s *gwapiv1.PolicyStatus) map[string]any {
		t.Helper()
		m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
		require.NoError(t, err)
		return m
	}

	t.Run("keeps other controller's ancestors and other status fields", func(t *testing.T) {
		oldStatus := toMap(t, &gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{other}})
		oldStatus["observedGeneration"] = int64(3)

		got := mergeUnstructuredPolicyStatus(egv1a1.GatewayControllerName, oldStatus,
			&gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{ours}})

		want := toMap(t, &gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{ours, other}})
		want["observedGeneration"] = int64(3)
		require.Equal(t, want, got)
	})

	t.Run("no old status", func(t *testing.T) {
		got := mergeUnstructuredPolicyStatus(egv1a1.GatewayControllerName, nil,
			&gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{ours}})
		require.Equal(t, toMap(t, &gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{ours}}), got)
	})

	t.Run("undecodable old ancestors are replaced", func(t *testing.T) {
		oldStatus := map[string]any{"ancestors": "not-a-list"}
		got := mergeUnstructuredPolicyStatus(egv1a1.GatewayControllerName, oldStatus,
			&gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{ours}})
		require.Equal(t, toMap(t, &gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{ours}}), got)
	})
}

// Test_updateStatusFromSubscriptions_ExtensionServerPolicy is a regression test for
// https://github.com/envoyproxy/gateway/issues/8808: the status of an extension server
// policy used to be replaced as a whole, dropping ancestors written by other controllers.
func Test_updateStatusFromSubscriptions_ExtensionServerPolicy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const otherController gwapiv1.GatewayController = "example.com/policy-controller"
	ours := policyAncestor(egv1a1.GatewayControllerName, "gateway1", string(gwapiv1.PolicyConditionAccepted))
	other := policyAncestor(otherController, "gateway1", "example.com/PolicyAffected")

	key := message.NamespacedNameAndGVK{
		NamespacedName:   types.NamespacedName{Namespace: "default", Name: "ext-policy"},
		GroupVersionKind: schema.GroupVersionKind{Group: "example.com", Version: "v1alpha1", Kind: "ExampleExtPolicy"},
	}
	var extensionPolicyStatuses watchable.Map[message.NamespacedNameAndGVK, *gwapiv1.PolicyStatus]
	extensionPolicyStatuses.Store(key, &gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{ours}})

	updater := &fakeUpdater{updates: make(chan Update, 1)}
	r := &gatewayAPIReconciler{
		log:             logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		classController: egv1a1.GatewayControllerName,
		statusUpdater:   updater,
		subscriptions: &subscriptions{
			extensionPolicyStatuses: extensionPolicyStatuses.Subscribe(ctx),
		},
	}

	go r.updateStatusFromSubscriptions(ctx, true)

	var update Update
	select {
	case update = <-updater.updates:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for status update")
	}
	require.Equal(t, key.NamespacedName, update.NamespacedName)

	oldStatus, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&gwapiv1.PolicyStatus{
		Ancestors: []gwapiv1.PolicyAncestorStatus{other},
	})
	require.NoError(t, err)
	policy := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "example.com/v1alpha1",
		"kind":       "ExampleExtPolicy",
		"metadata":   map[string]any{"namespace": "default", "name": "ext-policy"},
		"spec":       map[string]any{},
		"status":     oldStatus,
	}}

	mutated, ok := update.Mutator.Mutate(policy).(*unstructured.Unstructured)
	require.True(t, ok)

	got := &gwapiv1.PolicyStatus{}
	statusMap, ok := mutated.Object["status"].(map[string]any)
	require.True(t, ok)
	require.NoError(t, runtime.DefaultUnstructuredConverter.FromUnstructured(statusMap, got))
	require.Len(t, got.Ancestors, 2)
	require.Equal(t, egv1a1.GatewayControllerName, string(got.Ancestors[0].ControllerName))
	require.Equal(t, otherController, got.Ancestors[1].ControllerName)
	require.Equal(t, other.Conditions[0].Type, got.Ancestors[1].Conditions[0].Type)
}

// Test_updateStatusFromSubscriptions_BackendTLSPolicy checks that ancestors written by
// other controllers survive a status update of a typed policy.
func Test_updateStatusFromSubscriptions_BackendTLSPolicy(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const otherController gwapiv1.GatewayController = "example.com/gateway-controller"
	ours := policyAncestor(egv1a1.GatewayControllerName, "gateway1", string(gwapiv1.PolicyConditionAccepted))
	other := policyAncestor(otherController, "gateway1", string(gwapiv1.PolicyConditionAccepted))

	key := types.NamespacedName{Namespace: "default", Name: "backend-tls"}
	var backendTLSPolicyStatuses watchable.Map[types.NamespacedName, *gwapiv1.PolicyStatus]
	backendTLSPolicyStatuses.Store(key, &gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{ours}})

	updater := &fakeUpdater{updates: make(chan Update, 1)}
	r := &gatewayAPIReconciler{
		log:             logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		classController: egv1a1.GatewayControllerName,
		statusUpdater:   updater,
		subscriptions: &subscriptions{
			backendTLSPolicyStatuses: backendTLSPolicyStatuses.Subscribe(ctx),
		},
	}

	go r.updateStatusFromSubscriptions(ctx, false)

	var update Update
	select {
	case update = <-updater.updates:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for status update")
	}
	require.Equal(t, key, update.NamespacedName)

	policy := &gwapiv1.BackendTLSPolicy{
		ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
		Status:     gwapiv1.PolicyStatus{Ancestors: []gwapiv1.PolicyAncestorStatus{other}},
	}
	mutated, ok := update.Mutator.Mutate(policy).(*gwapiv1.BackendTLSPolicy)
	require.True(t, ok)
	require.Len(t, mutated.Status.Ancestors, 2)
	require.Equal(t, egv1a1.GatewayControllerName, string(mutated.Status.Ancestors[0].ControllerName))
	require.Equal(t, other, mutated.Status.Ancestors[1])
}
