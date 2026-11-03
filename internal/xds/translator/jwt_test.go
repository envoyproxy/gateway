// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"strings"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	jwtauthnv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/jwt_authn/v3"
	"github.com/stretchr/testify/require"
	gproto "google.golang.org/protobuf/proto"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestJWTAuthnDeduplicatesIdenticalRouteProviders(t *testing.T) {
	provider := ir.JWTProvider{
		Name:   "azure-jwt",
		Issuer: "https://sts.windows.net/example/",
		RemoteJWKS: &ir.RemoteJWKS{
			URI: "https://login.microsoftonline.com/common/discovery/keys",
		},
	}
	irListener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{
			{
				Name: "httproute/default/route-a/rule/0/match/0/example.com",
				Security: &ir.SecurityFeatures{
					JWT: &ir.JWT{
						Name:      "default/sp",
						Providers: []ir.JWTProvider{provider},
					},
				},
			},
			{
				Name: "httproute/default/route-b/rule/0/match/0/example.com",
				Security: &ir.SecurityFeatures{
					JWT: &ir.JWT{
						Name:      "default/sp",
						Providers: []ir.JWTProvider{provider},
					},
				},
			},
		},
	}

	var jwtAuthn jwtauthnv3.JwtAuthentication
	require.NoError(t, buildJWTAuthn(irListener, &jwtAuthn))
	require.Len(t, jwtAuthn.GetProviders(), 1)
	require.Len(t, jwtAuthn.GetRequirementMap(), 1)

	var requirementName string
	for name := range jwtAuthn.GetRequirementMap() {
		requirementName = name
	}
	for providerName := range jwtAuthn.GetProviders() {
		require.Equal(t, providerName, jwtAuthn.GetRequirementMap()[requirementName].GetProviderName())
	}

	jwtFilter := &jwt{}
	for _, irRoute := range irListener.Routes {
		xdsRoute := &routev3.Route{}
		require.NoError(t, jwtFilter.patchRoute(xdsRoute, irRoute, irListener))

		var perRouteConfig jwtauthnv3.PerRouteConfig
		require.NoError(t, xdsRoute.GetTypedPerFilterConfig()["envoy.filters.http.jwt_authn"].UnmarshalTo(&perRouteConfig))
		require.Equal(t, requirementName, perRouteConfig.GetRequirementName())
	}
}

func TestBuildJWTProviderClaimToHeaders(t *testing.T) {
	irProvider := &ir.JWTProvider{
		Name:   "test-jwt",
		Issuer: "https://issuer.example.com/",
		RemoteJWKS: &ir.RemoteJWKS{
			URI: "https://issuer.example.com/keys",
		},
		ClaimToHeaders: []egv1a1.ClaimToHeader{
			{
				Header: "X-Sub",
				Claim:  "sub",
			},
			{
				Header:    "X-Tenant-Name",
				ClaimPath: []string{"https://auth.sitecorecloud.io/claims/tenant_name"},
			},
			{
				Header:    "X-Nested-Claim",
				ClaimPath: []string{"a.b", "c.d"},
			},
		},
	}

	jwtProvider, err := buildJWTProvider(irProvider)
	require.NoError(t, err)
	require.Len(t, jwtProvider.GetClaimToHeaders(), 3)

	byHeader := make(map[string]*jwtauthnv3.JwtClaimToHeader, len(jwtProvider.GetClaimToHeaders()))
	for _, c := range jwtProvider.GetClaimToHeaders() {
		byHeader[c.GetHeaderName()] = c
	}

	sub := byHeader["X-Sub"]
	require.Equal(t, "sub", sub.GetClaimName())
	require.Empty(t, sub.GetClaimPath())

	tenant := byHeader["X-Tenant-Name"]
	require.Empty(t, tenant.GetClaimName())
	require.Len(t, tenant.GetClaimPath(), 1)
	require.Equal(t, "https://auth.sitecorecloud.io/claims/tenant_name", tenant.GetClaimPath()[0].GetKey())

	nested := byHeader["X-Nested-Claim"]
	require.Empty(t, nested.GetClaimName())
	require.Len(t, nested.GetClaimPath(), 2)
	require.Equal(t, "a.b", nested.GetClaimPath()[0].GetKey())
	require.Equal(t, "c.d", nested.GetClaimPath()[1].GetKey())
}

// Build a JWT config with long provider names so the human-readable prefix would
// exceed the length guard if left unbounded.
func TestJWTAuthnProviderNamesAreBounded(t *testing.T) {
	providers := make([]ir.JWTProvider, 0, 50)
	for i := 0; i < 50; i++ {
		providers = append(providers, ir.JWTProvider{
			Name:   strings.Repeat("very-long-jwt-provider-name", 3),
			Issuer: "https://issuer.example.com/",
			RemoteJWKS: &ir.RemoteJWKS{
				URI: "https://issuer.example.com/keys",
			},
		})
	}
	irJWT := &ir.JWT{Name: "default/sp", Providers: providers, AllowMissing: true}
	irListener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{
			{
				Name:     "httproute/default/route-a/rule/0/match/0/example.com",
				Security: &ir.SecurityFeatures{JWT: irJWT},
			},
		},
	}

	var jwtAuthn jwtauthnv3.JwtAuthentication
	require.NoError(t, buildJWTAuthn(irListener, &jwtAuthn))

	for providerName := range jwtAuthn.GetProviders() {
		require.LessOrEqual(t, len(providerName), maxJWTNameLength)
		// The uniqueness-guaranteeing hash suffix must survive truncation.
		require.Contains(t, providerName, "_")
	}
}

// The gateway API translator names every JWT, so an unnamed one is refused rather
// than named after its content, which would rename the requirement on every edit.
func TestJWTAuthnRequiresName(t *testing.T) {
	irListener := &ir.HTTPListener{
		Routes: []*ir.HTTPRoute{
			{
				Name: "httproute/default/route-a/rule/0/match/0/example.com",
				Security: &ir.SecurityFeatures{JWT: &ir.JWT{Providers: []ir.JWTProvider{{
					Name:   "example",
					Issuer: "https://issuer.example.com/",
					RemoteJWKS: &ir.RemoteJWKS{
						URI: "https://issuer.example.com/keys",
					},
				}}}},
			},
		},
	}

	var jwtAuthn jwtauthnv3.JwtAuthentication
	require.ErrorContains(t, buildJWTAuthn(irListener, &jwtAuthn), "jwt name is empty")
	require.ErrorContains(t, (&jwt{}).patchRoute(&routev3.Route{}, irListener.Routes[0], irListener), "jwt name is empty")
}

// policyJWTListener builds one listener whose routes all carry irJWT, and returns
// its jwt_authn config together with the requirement name each route refers to.
func policyJWTListener(t *testing.T, irJWT *ir.JWT, routeNames ...string) (*jwtauthnv3.JwtAuthentication, []string) {
	t.Helper()
	irListener := &ir.HTTPListener{}
	for _, name := range routeNames {
		irListener.Routes = append(irListener.Routes, &ir.HTTPRoute{
			Name:     name,
			Security: &ir.SecurityFeatures{JWT: irJWT},
		})
	}

	jwtAuthn := &jwtauthnv3.JwtAuthentication{}
	require.NoError(t, buildJWTAuthn(irListener, jwtAuthn))

	jwtFilter := &jwt{}
	refs := make([]string, 0, len(irListener.Routes))
	for _, irRoute := range irListener.Routes {
		xdsRoute := &routev3.Route{}
		require.NoError(t, jwtFilter.patchRoute(xdsRoute, irRoute, irListener))
		var perRouteConfig jwtauthnv3.PerRouteConfig
		require.NoError(t, xdsRoute.GetTypedPerFilterConfig()["envoy.filters.http.jwt_authn"].UnmarshalTo(&perRouteConfig))
		refs = append(refs, perRouteConfig.GetRequirementName())
	}
	return jwtAuthn, refs
}

func TestJWTAuthnNamedRequirementSurvivesEdit(t *testing.T) {
	provider := ir.JWTProvider{
		Name:      "example",
		Issuer:    "https://issuer.example.com/",
		Audiences: []string{"aud-a"},
		RemoteJWKS: &ir.RemoteJWKS{
			URI: "https://issuer.example.com/keys",
		},
	}
	edited := provider
	edited.Audiences = []string{"aud-a", "aud-b"}

	routes := []string{
		"httproute/default/route-a/rule/0/match/0/example.com",
		"httproute/default/route-b/rule/0/match/0/example.com",
	}
	before, beforeRefs := policyJWTListener(t, &ir.JWT{
		Name:      "default/sp",
		Providers: []ir.JWTProvider{provider},
	}, routes...)
	after, afterRefs := policyJWTListener(t, &ir.JWT{
		AllowMissing: true,
		Name:         "default/sp",
		Providers:    []ir.JWTProvider{edited},
	}, routes...)

	// Routes that share a name share one requirement, so adding a route leaves the
	// listener unchanged.
	require.Len(t, before.GetRequirementMap(), 1)
	require.Contains(t, before.GetRequirementMap(), "default/sp")
	require.Equal(t, []string{"default/sp", "default/sp"}, beforeRefs)

	// The edit, of the providers and the mode alike, changes the provider but not
	// the requirement the routes refer to: the route configuration stays put and
	// only the requirement_map entry changes.
	require.Equal(t, beforeRefs, afterRefs)
	require.Contains(t, after.GetRequirementMap(), "default/sp")
	require.False(t, gproto.Equal(
		before.GetRequirementMap()["default/sp"],
		after.GetRequirementMap()["default/sp"]))
	for providerName := range after.GetProviders() {
		require.NotContains(t, before.GetProviders(), providerName)
	}
}
