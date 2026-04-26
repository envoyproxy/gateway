// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

// routesOverlap is a test helper that returns true if two routes produce the
// same canonical overlap key (i.e., they match the same set of requests).
func routesOverlap(a, b *ir.HTTPRoute) bool {
	return buildOverlapKey(a) == buildOverlapKey(b)
}

func TestBuildOverlapKey(t *testing.T) {
	tests := []struct {
		name    string
		a       *ir.HTTPRoute
		b       *ir.HTTPRoute
		overlap bool
	}{
		{
			name: "identical exact path and hostname",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
			},
			overlap: true,
		},
		{
			name: "different exact paths",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/bar")},
			},
			overlap: false,
		},
		{
			name: "different hostnames",
			a: &ir.HTTPRoute{
				Hostname:  "a.example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "b.example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
			},
			overlap: false,
		},
		{
			name: "identical prefix paths are detected as overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: new("/api")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: new("/api")},
			},
			overlap: true,
		},
		{
			name: "nil path matches are detected as overlap",
			a: &ir.HTTPRoute{
				Hostname: "example.com",
			},
			b: &ir.HTTPRoute{
				Hostname: "example.com",
			},
			overlap: true,
		},
		{
			name: "nil path and root prefix path are detected as overlap",
			a: &ir.HTTPRoute{
				Hostname: "example.com",
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: new("/")},
			},
			overlap: true,
		},
		{
			name: "path prefixes with equivalent trailing slash are detected as overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: new("/api")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: new("/api/")},
			},
			overlap: true,
		},
		{
			name: "exact vs prefix same value not overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: new("/foo")},
			},
			overlap: false,
		},
		{
			name: "identical exact path with identical header matches",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: new("val1")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: new("val1")},
				},
			},
			overlap: true,
		},
		{
			name: "identical exact path with different header matches",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: new("val1")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: new("val2")},
				},
			},
			overlap: false,
		},
		{
			name: "identical exact path one has headers other does not",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: new("val1")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
			},
			overlap: false,
		},
		{
			name: "header names are compared case-insensitively",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: new("val1")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "x-custom", Exact: new("val1")},
				},
			},
			overlap: true,
		},
		{
			name: "header matches in different order still overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-A", Exact: new("1")},
					{Name: "X-B", Exact: new("2")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-B", Exact: new("2")},
					{Name: "X-A", Exact: new("1")},
				},
			},
			overlap: true,
		},
		{
			name: "identical exact path with identical query param matches",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				QueryParamMatches: []*ir.StringMatch{
					{Name: "key", Exact: new("value")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: new("/foo")},
				QueryParamMatches: []*ir.StringMatch{
					{Name: "key", Exact: new("value")},
				},
			},
			overlap: true,
		},
		{
			name: "query param matches in different order still overlap",
			a: &ir.HTTPRoute{
				Hostname: "example.com",
				QueryParamMatches: []*ir.StringMatch{
					{Name: "a", Exact: new("1")},
					{Name: "b", Exact: new("2")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname: "example.com",
				QueryParamMatches: []*ir.StringMatch{
					{Name: "b", Exact: new("2")},
					{Name: "a", Exact: new("1")},
				},
			},
			overlap: true,
		},
		{
			name: "cookie matches in different order still overlap",
			a: &ir.HTTPRoute{
				Hostname: "example.com",
				CookieMatches: []*ir.StringMatch{
					{Name: "a", Exact: new("1")},
					{Name: "b", Exact: new("2")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname: "example.com",
				CookieMatches: []*ir.StringMatch{
					{Name: "b", Exact: new("2")},
					{Name: "a", Exact: new("1")},
				},
			},
			overlap: true,
		},
		{
			name: "identical regex path matches are detected as overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{SafeRegex: new("/foo.*")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{SafeRegex: new("/foo.*")},
			},
			overlap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.overlap, routesOverlap(tt.a, tt.b))
		})
	}
}

func TestBuildResourceMetadataWrappedRouteKinds(t *testing.T) {
	tests := []struct {
		name string
		obj  client.Object
		want string
	}{
		{
			name: "http route context",
			obj: &HTTPRouteContext{
				HTTPRoute: &gwapiv1.HTTPRoute{},
			},
			want: resource.KindHTTPRoute,
		},
		{
			name: "grpc route context",
			obj: &GRPCRouteContext{
				GRPCRoute: &gwapiv1.GRPCRoute{},
			},
			want: resource.KindGRPCRoute,
		},
		{
			name: "tls route context",
			obj: &TLSRouteContext{
				TLSRoute: &gwapiv1.TLSRoute{},
			},
			want: resource.KindTLSRoute,
		},
		{
			name: "tcp route context",
			obj: &TCPRouteContext{
				TCPRoute: &gwapiv1a2.TCPRoute{},
			},
			want: resource.KindTCPRoute,
		},
		{
			name: "udp route context",
			obj: &UDPRouteContext{
				UDPRoute: &gwapiv1a2.UDPRoute{},
			},
			want: resource.KindUDPRoute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := buildResourceMetadata(tt.obj, nil)
			require.Equal(t, tt.want, metadata.Kind)
		})
	}
}

func TestCheckRouteOverlapsRemovesStaleCondition(t *testing.T) {
	parentRef := gwapiv1.ParentReference{Name: "gateway-1"}
	httpRoute := &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "route-1",
			Namespace: "default",
		},
		Status: gwapiv1.HTTPRouteStatus{
			RouteStatus: gwapiv1.RouteStatus{
				Parents: []gwapiv1.RouteParentStatus{
					{
						ParentRef: parentRef,
						Conditions: []metav1.Condition{
							{
								Type:   string(gwapiv1.RouteConditionAccepted),
								Status: metav1.ConditionTrue,
								Reason: string(gwapiv1.RouteReasonAccepted),
							},
							{
								Type:   string(status.RouteConditionRouteRulesOverlap),
								Status: metav1.ConditionTrue,
								Reason: string(status.RouteReasonRouteRulesOverlap),
							},
						},
					},
				},
			},
		},
	}
	route := &HTTPRouteContext{
		HTTPRoute: httpRoute,
	}

	gateway := &GatewayContext{
		Gateway: &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gateway-1",
				Namespace: "default",
			},
		},
	}
	listener := &ListenerContext{
		Listener: &gwapiv1.Listener{Name: "http"},
		gateway:  gateway,
	}
	route.ParentRefs = map[gwapiv1.ParentReference]*RouteParentContext{
		parentRef: {
			ParentReference:      &parentRef,
			HTTPRoute:            httpRoute,
			routeParentStatusIdx: 0,
			listeners:            []*ListenerContext{listener},
		},
	}

	xdsIR := resource.XdsIRMap{
		"default/gateway-1": {
			HTTP: []*ir.HTTPListener{
				{
					CoreListenerDetails: ir.CoreListenerDetails{
						Name: irListenerName(listener),
					},
					Routes: []*ir.HTTPRoute{
						{
							Hostname: "example.com",
							Metadata: &ir.ResourceMetadata{
								Kind:      resource.KindHTTPRoute,
								Namespace: "default",
								Name:      "route-1",
							},
						},
					},
				},
			},
		},
	}

	translator := &Translator{}
	translator.checkRouteOverlaps([]*HTTPRouteContext{route}, nil, xdsIR)

	conditions := route.Status.RouteStatus.Parents[0].Conditions
	require.NotNil(t, meta.FindStatusCondition(conditions, string(gwapiv1.RouteConditionAccepted)))
	assert.Nil(t, meta.FindStatusCondition(conditions, string(status.RouteConditionRouteRulesOverlap)))
}
