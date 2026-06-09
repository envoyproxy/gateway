// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"

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

func ruleNames(rules []*ir.AuthorizationRule) []string {
	names := make([]string, 0, len(rules))
	for _, r := range rules {
		names = append(names, r.Name)
	}
	return names
}

func Test_authIndependentPrefix(t *testing.T) {
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
