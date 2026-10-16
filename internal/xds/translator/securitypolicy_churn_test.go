// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"testing"

	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/envoyproxy/gateway/internal/ir"
)

// securedIR returns a listener whose routes carry OIDC, ext_authz and JWT,
// assembled from the existing single-feature fixtures. extAuthRoute selects
// the ext-auth.yaml route to copy ext_authz from: 0 uses a gRPC service and 2
// an HTTP service.
func securedIR(t *testing.T, extAuthRoute int) *ir.Xds {
	t.Helper()
	x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/oidc.yaml")
	extAuth := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/ext-auth.yaml").HTTP[0].Routes[extAuthRoute].Security.ExtAuth
	jwt := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/jwt-multi-route-multi-provider.yaml").HTTP[0].Routes[0].Security.JWT
	for _, r := range x.HTTP[0].Routes {
		r.Security.ExtAuth = extAuth.DeepCopy()
		r.Security.JWT = jwt.DeepCopy()
		setJWKSBackend(r, "securitypolicy/default/policy-for-first-route", "10.0.0.5")
	}
	return x
}

// setJWKSBackend points every JWT provider of a route at a remote JWKS backend
// Service the way the Gateway API translator does: the destination is named
// after the owning policy, and the setting carries the Service's metadata and
// its current endpoint. Two policies with the same backend therefore get
// different destination names.
func setJWKSBackend(r *ir.HTTPRoute, policy, endpoint string) {
	for i := range r.Security.JWT.Providers {
		name := fmt.Sprintf("%s/jwt/%d", policy, i)
		r.Security.JWT.Providers[i].LocalJWKS = nil
		r.Security.JWT.Providers[i].RemoteJWKS = &ir.RemoteJWKS{
			URI: "https://id.example.com/keys",
			Destination: &ir.RouteDestination{
				Name:     name,
				Metadata: &ir.ResourceMetadata{Kind: "SecurityPolicy", Namespace: "default", Name: policy},
				Settings: []*ir.DestinationSetting{{
					Name:      name + "/backend/0",
					Weight:    new(uint32(1)),
					Protocol:  ir.HTTP,
					Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, endpoint, 5556, false, nil)},
					Metadata:  &ir.ResourceMetadata{Kind: "Service", Namespace: "default", Name: "idp", SectionName: "5556"},
				}},
			},
		}
	}
}

// withNewPolicy adds a route on a new hostname whose own SecurityPolicy names
// a different OIDC client and a different authorization service, with the
// same JWT settings and JWKS backend.
func withNewPolicy(x *ir.Xds) *ir.Xds {
	r := x.HTTP[0].Routes[0].DeepCopy()
	r.Name = "httproute/default/new-app/rule/0/match/0/new-app_example_com"
	r.Hostname = "new-app.example.com"
	r.Security.OIDC.Name = "securitypolicy/default/new-app"
	r.Security.OIDC.ClientID = "new-app-client"
	r.Security.OIDC.CookieSuffix = "newapp"
	r.Security.ExtAuth.Name = "securitypolicy/default/new-app"
	if grpc := r.Security.ExtAuth.GRPC; grpc != nil {
		grpc.Destination.Name = "securitypolicy/default/new-app/default/grpc-backend"
		grpc.Authority = "new-app-authz.default:9000"
	}
	if http := r.Security.ExtAuth.HTTP; http != nil {
		http.Destination.Name = "securitypolicy/default/new-app/default/http-backend"
		http.Authority = "new-app-authz.default:80"
	}
	setJWKSBackend(r, "securitypolicy/default/new-app", "10.0.0.5")
	x.HTTP[0].Routes = append(x.HTTP[0].Routes, r)
	return x
}

func translateListenersAndRoutes(t *testing.T, x *ir.Xds) (listeners, routes []proto.Message) {
	t.Helper()
	tCtx, err := (&Translator{ControllerNamespace: "envoy-gateway-system"}).Translate(t.Context(), x)
	require.NoError(t, err)
	for _, l := range tCtx.XdsResources[resourcev3.ListenerType] {
		listeners = append(listeners, l)
	}
	for _, r := range tCtx.XdsResources[resourcev3.RouteType] {
		routes = append(routes, r)
	}
	return listeners, routes
}

func equalMessages(a, b []proto.Message) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !proto.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}

// Creating or deleting a SecurityPolicy must reach Envoy as a route change
// only. Any change to a listener makes Envoy drain every connection on it
// (https://github.com/envoyproxy/gateway/issues/9130).
func TestSecurityPolicyChurnKeepsListener(t *testing.T) {
	for name, extAuthRoute := range map[string]int{"grpc ext_authz": 0, "http ext_authz": 2} {
		t.Run(name, func(t *testing.T) {
			// Translation mutates its input, so every translation gets a fresh IR.
			l1, r1 := translateListenersAndRoutes(t, securedIR(t, extAuthRoute))
			l2, r2 := translateListenersAndRoutes(t, withNewPolicy(securedIR(t, extAuthRoute)))
			require.True(t, equalMessages(l1, l2), "listener changed when a policy was created")
			require.False(t, equalMessages(r1, r2), "the new policy did not reach the routes")
		})
	}
}

// A JWKS backend's pods restarting changes its endpoints. That must update
// the JWKS cluster's endpoints, not rename the cluster, because the cluster
// name is part of the listener's jwt_authn filter.
func TestJWKSEndpointChangeKeepsListener(t *testing.T) {
	withEndpoint := func(endpoint string) *ir.Xds {
		x := securedIR(t, 0)
		for _, r := range x.HTTP[0].Routes {
			setJWKSBackend(r, "securitypolicy/default/policy-for-first-route", endpoint)
		}
		return x
	}
	l1, _ := translateListenersAndRoutes(t, withEndpoint("10.0.0.5"))
	l2, _ := translateListenersAndRoutes(t, withEndpoint("10.0.0.6"))
	require.True(t, equalMessages(l1, l2), "listener changed when a JWKS endpoint changed")
}

func TestJWKSDestinationName(t *testing.T) {
	jwks := func(policy, service, endpoint string, withMetadata bool) *ir.RemoteJWKS {
		r := &ir.HTTPRoute{Security: &ir.SecurityFeatures{JWT: &ir.JWT{Providers: []ir.JWTProvider{{}}}}}
		setJWKSBackend(r, policy, endpoint)
		j := r.Security.JWT.Providers[0].RemoteJWKS
		j.Destination.Settings[0].Metadata.Name = service
		if !withMetadata {
			j.Destination.Settings[0].Metadata = nil
		}
		return j
	}
	base := jwksDestination(jwks("policy-a", "idp", "10.0.0.5", true)).Name

	require.Equal(t, base, jwksDestination(jwks("policy-b", "idp", "10.0.0.5", true)).Name, "same backend from another policy")
	require.Equal(t, base, jwksDestination(jwks("policy-a", "idp", "10.0.0.6", true)).Name, "same backend with new endpoints")
	require.NotEqual(t, base, jwksDestination(jwks("policy-a", "other-idp", "10.0.0.5", true)).Name, "different backend")
	require.NotEqual(t,
		jwksDestination(jwks("policy-a", "idp", "10.0.0.5", false)).Name,
		jwksDestination(jwks("policy-a", "idp", "10.0.0.6", false)).Name,
		"without metadata, endpoints are the only identity")
}
