// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
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
	}
	return x
}

// withNewPolicy adds a route on a new hostname whose own SecurityPolicy names
// a different OIDC client and a different authorization service, with the
// same JWT settings.
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
