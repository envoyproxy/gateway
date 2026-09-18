// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"reflect"
	"testing"

	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func cidrRule() *ir.AuthorizationRule {
	return &ir.AuthorizationRule{
		Name:   "cidr",
		Action: egv1a1.AuthorizationActionDeny,
		Principal: ir.Principal{
			ClientCIDRs: []*ir.CIDRMatch{{CIDR: "10.0.0.0/8"}},
		},
	}
}

func geoRule() *ir.AuthorizationRule {
	country := "IR"
	return &ir.AuthorizationRule{
		Name:   "geo",
		Action: egv1a1.AuthorizationActionDeny,
		Principal: ir.Principal{
			ClientIPGeoLocations: []egv1a1.ClientIPGeoLocation{{Country: &country}},
		},
	}
}

func jwtRule() *ir.AuthorizationRule {
	return &ir.AuthorizationRule{
		Name:   "jwt",
		Action: egv1a1.AuthorizationActionAllow,
		Principal: ir.Principal{
			JWT: &egv1a1.JWTPrincipal{
				Provider: "example",
				Claims: []egv1a1.JWTClaim{{
					Name:   "role",
					Values: []string{"admin"},
				}},
			},
		},
	}
}

func headerRule() *ir.AuthorizationRule {
	return &ir.AuthorizationRule{
		Name:   "header",
		Action: egv1a1.AuthorizationActionDeny,
		Principal: ir.Principal{
			Headers: []egv1a1.AuthorizationHeaderMatch{{
				Name:   "x-test",
				Values: []string{"deny"},
			}},
		},
	}
}

func celRule() *ir.AuthorizationRule {
	expression := "request.headers['authorization'] != ''"
	return &ir.AuthorizationRule{
		Name:   "cel",
		Action: egv1a1.AuthorizationActionDeny,
		CEL:    &expression,
	}
}

func operationRule() *ir.AuthorizationRule {
	pathType := gwapiv1.PathMatchPathPrefix
	return &ir.AuthorizationRule{
		Name:   "operation",
		Action: egv1a1.AuthorizationActionDeny,
		Operation: &egv1a1.Operation{
			Methods: []gwapiv1.HTTPMethod{gwapiv1.HTTPMethodGet},
			Path:    &egv1a1.PathMatch{Type: &pathType, Value: "/admin"},
		},
	}
}

func ruleNames(rules []*ir.AuthorizationRule) []string {
	names := make([]string, 0, len(rules))
	for _, r := range rules {
		names = append(names, r.Name)
	}
	return names
}

func Test_authIndependentPrefix(t *testing.T) {
	operationAllow := operationRule()
	operationAllow.Action = egv1a1.AuthorizationActionAllow

	tests := []struct {
		name          string
		authorization *ir.Authorization
		want          []string
	}{
		{
			name:          "nil authorization",
			authorization: nil,
			want:          []string{},
		},
		{
			name: "all rules are authentication-independent",
			authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{geoRule(), cidrRule()},
			},
			want: []string{"geo", "cidr"},
		},
		{
			name: "prefix stops at the first jwt rule",
			authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{geoRule(), jwtRule(), cidrRule()},
			},
			want: []string{"geo"},
		},
		{
			name: "prefix stops at the first header rule",
			authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{cidrRule(), headerRule(), geoRule()},
			},
			want: []string{"cidr"},
		},
		{
			name: "prefix stops at the first CEL rule",
			authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{cidrRule(), celRule(), geoRule()},
			},
			want: []string{"cidr"},
		},
		{
			name: "prefix stops at the first operation rule",
			authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{cidrRule(), operationRule(), geoRule()},
			},
			want: []string{"cidr"},
		},
		{
			name: "prefix stops at an operation allow rule and does not pick a later CIDR deny",
			authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{operationAllow, cidrRule()},
			},
			want: []string{},
		},
		{
			name: "leading authentication-dependent rule yields an empty prefix",
			authorization: &ir.Authorization{
				Rules: []*ir.AuthorizationRule{jwtRule(), geoRule()},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := authIndependentPrefix(tt.authorization)
			assert.Equal(t, tt.want, ruleNames(got))
		})
	}
}

func Test_isPreAuthRule(t *testing.T) {
	principalType := reflect.TypeOf(ir.Principal{})
	principalFields := make([]string, 0, principalType.NumField())
	for i := 0; i < principalType.NumField(); i++ {
		principalFields = append(principalFields, principalType.Field(i).Name)
	}
	assert.ElementsMatch(t, []string{"ClientCIDRs", "JWT", "Headers", "ClientIPGeoLocations"}, principalFields,
		"update isPreAuthRule and its tests when adding a field to ir.Principal")

	ruleType := reflect.TypeOf(ir.AuthorizationRule{})
	ruleFields := make([]string, 0, ruleType.NumField())
	for i := 0; i < ruleType.NumField(); i++ {
		ruleFields = append(ruleFields, ruleType.Field(i).Name)
	}
	assert.ElementsMatch(t, []string{"Name", "Action", "Operation", "Principal", "CEL"}, ruleFields,
		"update isPreAuthRule and its tests when adding a field to ir.AuthorizationRule")

	tests := []struct {
		name string
		rule *ir.AuthorizationRule
		want bool
	}{
		{name: "empty principal", rule: &ir.AuthorizationRule{}, want: true},
		{name: "client CIDRs", rule: cidrRule(), want: true},
		{name: "client IP geo locations", rule: geoRule(), want: true},
		{name: "JWT", rule: jwtRule(), want: false},
		{name: "headers", rule: headerRule(), want: false},
		{name: "CEL", rule: celRule(), want: false},
		{name: "operation", rule: operationRule(), want: false},
		{
			name: "path only",
			rule: &ir.AuthorizationRule{
				Principal: cidrRule().Principal,
				Operation: &egv1a1.Operation{Path: &egv1a1.PathMatch{Value: "/old"}},
			},
			want: false,
		},
		{
			name: "method only",
			rule: &ir.AuthorizationRule{
				Principal: cidrRule().Principal,
				Operation: &egv1a1.Operation{Methods: []gwapiv1.HTTPMethod{gwapiv1.HTTPMethodGet}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isPreAuthRule(tt.rule))
		})
	}
}

func Test_hasPreRBACAuthentication(t *testing.T) {
	tests := []struct {
		name string
		sf   *ir.SecurityFeatures
		want bool
	}{
		{name: "nil", sf: nil, want: false},
		{name: "authorization only", sf: &ir.SecurityFeatures{Authorization: &ir.Authorization{}}, want: false},
		{name: "oidc", sf: &ir.SecurityFeatures{OIDC: &ir.OIDC{}}, want: true},
		{name: "ext auth", sf: &ir.SecurityFeatures{ExtAuth: &ir.ExtAuth{}}, want: true},
		{name: "basic auth", sf: &ir.SecurityFeatures{BasicAuth: &ir.BasicAuth{}}, want: true},
		{name: "api key auth", sf: &ir.SecurityFeatures{APIKeyAuth: &ir.APIKeyAuth{}}, want: true},
		{name: "jwt", sf: &ir.SecurityFeatures{JWT: &ir.JWT{}}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasPreRBACAuthentication(tt.sf))
		})
	}
}

func Test_routeNeedsPreAuthRBAC(t *testing.T) {
	tests := []struct {
		name  string
		route *ir.HTTPRoute
		want  bool
	}{
		{
			name:  "no security",
			route: &ir.HTTPRoute{},
			want:  false,
		},
		{
			name: "authorization without authentication",
			route: &ir.HTTPRoute{
				Security: &ir.SecurityFeatures{
					Authorization: &ir.Authorization{Rules: []*ir.AuthorizationRule{geoRule()}},
				},
			},
			want: false,
		},
		{
			name: "authentication without an authentication-independent prefix",
			route: &ir.HTTPRoute{
				Security: &ir.SecurityFeatures{
					OIDC:          &ir.OIDC{},
					Authorization: &ir.Authorization{Rules: []*ir.AuthorizationRule{jwtRule()}},
				},
			},
			want: false,
		},
		{
			name: "authentication with an allow-only prefix",
			route: &ir.HTTPRoute{
				Security: &ir.SecurityFeatures{
					OIDC: &ir.OIDC{},
					Authorization: &ir.Authorization{Rules: []*ir.AuthorizationRule{
						{
							Name:      "allow-cidr",
							Action:    egv1a1.AuthorizationActionAllow,
							Principal: cidrRule().Principal,
						},
						jwtRule(),
					}},
				},
			},
			want: false,
		},
		{
			name: "authentication with an allow followed by a deny in the prefix",
			route: &ir.HTTPRoute{
				Security: &ir.SecurityFeatures{
					OIDC: &ir.OIDC{},
					Authorization: &ir.Authorization{Rules: []*ir.AuthorizationRule{
						{
							Name:      "allow-cidr",
							Action:    egv1a1.AuthorizationActionAllow,
							Principal: cidrRule().Principal,
						},
						geoRule(),
						jwtRule(),
					}},
				},
			},
			want: true,
		},
		{
			name: "authentication with a path deny rule followed by a CIDR deny",
			route: &ir.HTTPRoute{
				Security: &ir.SecurityFeatures{
					OIDC:          &ir.OIDC{},
					Authorization: &ir.Authorization{Rules: []*ir.AuthorizationRule{operationRule(), cidrRule()}},
				},
			},
			want: false,
		},
		{
			name: "authentication with a geo deny prefix",
			route: &ir.HTTPRoute{
				Security: &ir.SecurityFeatures{
					OIDC:          &ir.OIDC{},
					Authorization: &ir.Authorization{Rules: []*ir.AuthorizationRule{geoRule(), jwtRule()}},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, routeNeedsPreAuthRBAC(tt.route))
		})
	}
}

func Test_rbacPatchRoute(t *testing.T) {
	irRoute := &ir.HTTPRoute{
		Security: &ir.SecurityFeatures{
			OIDC: &ir.OIDC{},
			Authorization: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules: []*ir.AuthorizationRule{
					{
						Name:   "allow-admin-get",
						Action: egv1a1.AuthorizationActionAllow,
						Operation: &egv1a1.Operation{
							Methods: []gwapiv1.HTTPMethod{gwapiv1.HTTPMethodGet},
							Path: &egv1a1.PathMatch{
								Type:  new(gwapiv1.PathMatchPathPrefix),
								Value: "/admin",
							},
						},
					},
					cidrRule(),
				},
			},
		},
	}

	route := &routev3.Route{}
	require.NoError(t, (&rbac{}).patchRoute(route, irRoute, nil))

	cfg := route.GetTypedPerFilterConfig()

	// A path-dependent policy must not be pre-auth enforced: the path and
	// methods can change between the pre-auth filter and the main RBAC filter.
	_, hasPreAuth := cfg[rbacPreAuthFilterName]
	assert.False(t, hasPreAuth, "pre-auth RBAC config must not be set for a path-dependent policy")

	// The main RBAC filter must keep the full policy, including the operation
	// rule and the real default action.
	mainAny, ok := cfg[egv1a1.EnvoyFilterRBAC.String()]
	require.True(t, ok, "main RBAC per-route config must be set")
	var mainPerRoute rbacv3.RBACPerRoute
	require.NoError(t, mainAny.UnmarshalTo(&mainPerRoute))

	matchers := mainPerRoute.Rbac.Matcher.GetMatcherList().GetMatchers()
	require.Len(t, matchers, 2)
	assert.Equal(t, "allow-admin-get", matchers[0].GetOnMatch().GetAction().GetName())
	assert.Equal(t, "cidr", matchers[1].GetOnMatch().GetAction().GetName())

	defaultAction := mainPerRoute.Rbac.Matcher.GetOnNoMatch().GetAction()
	assert.Equal(t, "default", defaultAction.GetName())
	var defaultCfg rbacconfigv3.Action
	require.NoError(t, defaultAction.GetTypedConfig().UnmarshalTo(&defaultCfg))
	assert.Equal(t, rbacconfigv3.RBAC_DENY, defaultCfg.GetAction())
}
