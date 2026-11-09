// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestOrderAuthenticationGate(t *testing.T) {
	basic := egv1a1.EnvoyFilterBasicAuth.String() + "/policy"
	jwt := egv1a1.EnvoyFilterJWTAuthn.String()
	extAuthz := egv1a1.EnvoyFilterExtAuthz.String() + "/policy"
	rbac := egv1a1.EnvoyFilterRBAC.String()
	router := egv1a1.EnvoyFilterRouter.String()

	for _, test := range []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "default order",
			input: []string{extAuthz, basic, jwt, rbac, authenticationGateName, router},
			want:  []string{extAuthz, basic, jwt, authenticationGateName, rbac, router},
		},
		{
			name:  "basic reordered after jwt",
			input: []string{jwt, extAuthz, authenticationGateName, rbac, basic, router},
			want:  []string{jwt, extAuthz, rbac, basic, authenticationGateName, router},
		},
		{
			name:  "no gate",
			input: []string{basic, jwt, rbac, router},
			want:  []string{basic, jwt, rbac, router},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			filters := make([]*hcmv3.HttpFilter, 0, len(test.input))
			for _, name := range test.input {
				filters = append(filters, &hcmv3.HttpFilter{Name: name})
			}
			got := make([]string, 0, len(filters))
			for _, filter := range orderAuthenticationGate(filters) {
				got = append(got, filter.Name)
			}
			require.Equal(t, test.want, got)
		})
	}
}

func TestAuthenticationGateFailsClosedWithoutMethods(t *testing.T) {
	route := &routev3.Route{}
	irRoute := &ir.HTTPRoute{Security: &ir.SecurityFeatures{AuthenticationAny: true}}
	require.NoError(t, (&authenticationGate{}).patchRoute(route, irRoute, nil))

	perRoute := &rbacv3.RBACPerRoute{}
	require.NoError(t, route.TypedPerFilterConfig[authenticationGateName].UnmarshalTo(perRoute))
	require.Equal(t, "ALLOW", perRoute.GetRbac().GetRules().GetAction().String())
	require.Empty(t, perRoute.GetRbac().GetRules().GetPolicies())
}

func TestAuthenticationGateSkipsOtherRoutes(t *testing.T) {
	route := &routev3.Route{}
	irRoute := &ir.HTTPRoute{Security: &ir.SecurityFeatures{BasicAuth: &ir.BasicAuth{Name: "basic"}}}
	require.NoError(t, (&authenticationGate{}).patchRoute(route, irRoute, nil))
	require.Empty(t, route.TypedPerFilterConfig)
}
