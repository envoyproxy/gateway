// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0

package gatewayapi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestHTTPFiltersContextRuleName(t *testing.T) {
	first, second := gwapiv1.SectionName("first"), gwapiv1.SectionName("second")
	httpRoute := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{Spec: gwapiv1.HTTPRouteSpec{Rules: []gwapiv1.HTTPRouteRule{{Name: &first}, {Name: &second}, {}}}}}
	grpcRoute := &GRPCRouteContext{GRPCRoute: &gwapiv1.GRPCRoute{Spec: gwapiv1.GRPCRouteSpec{Rules: []gwapiv1.GRPCRouteRule{{Name: &first}, {Name: &second}, {}}}}}
	for _, route := range []RouteContext{httpRoute, grpcRoute} {
		for _, tc := range []struct {
			index int
			want  *gwapiv1.SectionName
		}{{-1, nil}, {0, &first}, {1, &second}, {2, nil}, {3, nil}} {
			t.Run(fmt.Sprintf("%s/%d", route.GetRouteType(), tc.index), func(t *testing.T) {
				c := &HTTPFiltersContext{Route: route, RuleIdx: tc.index}
				require.Equal(t, tc.want, c.ruleName())
			})
		}
	}
	t.Run("no-route", func(t *testing.T) { require.Nil(t, (&HTTPFiltersContext{}).ruleName()) })
}

func TestListenerRoutingDestinationNameWithoutContext(t *testing.T) {
	translator := &Translator{}
	for _, tc := range []struct {
		name     string
		gateway  *GatewayContext
		listener *ListenerContext
	}{
		{name: "no-gateway", listener: &ListenerContext{}},
		{name: "no-listener", gateway: &GatewayContext{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, "destination", translator.listenerRoutingDestinationName("destination", tc.gateway, nil, tc.listener, ptr.To(gwapiv1.SectionName("rule"))))
		})
	}
}
