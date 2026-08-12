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
						Providers: []ir.JWTProvider{provider},
					},
				},
			},
			{
				Name: "httproute/default/route-b/rule/0/match/0/example.com",
				Security: &ir.SecurityFeatures{
					JWT: &ir.JWT{
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

func TestJWTAuthnNamesAreBounded(t *testing.T) {
	// Build a JWT config with many long provider names so the joined,
	// human-readable prefix would exceed the length guard if left unbounded.
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
	irJWT := &ir.JWT{Providers: providers, AllowMissing: true}
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
	}
	for requirementName := range jwtAuthn.GetRequirementMap() {
		require.LessOrEqual(t, len(requirementName), maxJWTNameLength)
		// The uniqueness-guaranteeing hash suffix must survive truncation.
		require.Contains(t, requirementName, "_")
	}
}
