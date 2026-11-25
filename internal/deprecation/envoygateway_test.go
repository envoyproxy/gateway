// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestWarnEnvoyGateway(t *testing.T) {
	eg := egv1a1.DefaultEnvoyGateway()

	testCases := []struct {
		name     string
		eg       *egv1a1.EnvoyGateway
		expected []string
	}{
		{
			name:     "nil EnvoyGateway",
			eg:       nil,
			expected: nil,
		},
		{
			name:     "nil ExtensionAPIs",
			eg:       eg,
			expected: nil,
		},
		{
			name: "disableLua is set",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionAPIs: &egv1a1.ExtensionAPISettings{
						DisableLua: new(true),
					},
				},
			},
			expected: []string{"disableLua is deprecated, use enableLua instead"},
		},
		{
			name: "enableLua is set",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:  egv1a1.DefaultGateway(),
					Provider: egv1a1.DefaultEnvoyGatewayProvider(),
					ExtensionAPIs: &egv1a1.ExtensionAPISettings{
						EnableLua: true,
					},
				},
			},
			expected: nil,
		},
		{
			name: "default envoyProxy uses deprecated fields",
			eg: &egv1a1.EnvoyGateway{
				EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
					Gateway:    egv1a1.DefaultGateway(),
					Provider:   egv1a1.DefaultEnvoyGatewayProvider(),
					EnvoyProxy: &egv1a1.EnvoyProxySpec{LuaValidation: new(egv1a1.LuaValidationStrict)},
				},
			},
			expected: []string{"envoyProxy.luaValidation is deprecated, use envoyProxy.lua.validationType instead"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			warnings := WarnEnvoyGateway(tc.eg)
			assert.Equal(t, tc.expected, warnings)
		})
	}
}
