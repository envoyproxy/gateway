// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecation

import (
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestClientTrafficPolicy(t *testing.T) {
	require.Empty(t, ClientTrafficPolicy(&egv1a1.ClientTrafficPolicy{}))

	policy := &egv1a1.ClientTrafficPolicy{}
	policy.Spec.TLS = &egv1a1.ClientTLSSettings{ClientValidation: &egv1a1.ClientValidationContext{Optional: true}}
	require.Equal(t, map[string]DeprecatedField{
		"spec.tls.clientValidation.optional": {Replacement: "spec.tls.clientValidation.mode"},
	}, ClientTrafficPolicy(policy))
}

func TestEnvoyProxy(t *testing.T) {
	require.Empty(t, EnvoyProxy(&egv1a1.EnvoyProxy{}))

	ep := &egv1a1.EnvoyProxy{}
	ep.Spec.LuaValidation = new(egv1a1.LuaValidationDisabled)
	require.Equal(t, map[string]DeprecatedField{
		"spec.luaValidation": {Replacement: "spec.lua.validationType"},
	}, EnvoyProxy(ep))
}
