// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

var (
	TLSSecretKind       = v1.Kind("Secret")
	TLSUnrecognizedKind = v1.Kind("Unrecognized")
)

func setEnv(t *testing.T, key, value string) {
	t.Helper() // Marks the function as a test helper function.
	prevValue, isSet := os.LookupEnv(key)
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() {
		if isSet {
			require.NoError(t, os.Setenv(key, prevValue))
		} else {
			require.NoError(t, os.Unsetenv(key))
		}
	})
}

func TestConfig_EnvOverrides(t *testing.T) {
	setEnv(t, "ENVOY_GATEWAY_LEADER_ELECTION_ENABLED", "false")
	cfg, err := New()
	require.NoError(t, err)
	require.Nil(t, cfg.EnvoyGateway.LeaderElection, "leader election should be disabled by env var")
}

func TestValidate(t *testing.T) {
	cfg, err := New()
	require.NoError(t, err)

	testCases := []struct {
		name   string
		cfg    *Server
		expect bool
	}{
		{
			name:   "nil cfg",
			cfg:    nil,
			expect: false,
		},
		{
			name:   "default",
			cfg:    cfg,
			expect: true,
		},
		{
			name: "empty namespace",
			cfg: &Server{
				EnvoyGateway: &v1alpha1.EnvoyGateway{
					EnvoyGatewaySpec: v1alpha1.EnvoyGatewaySpec{
						Gateway:  v1alpha1.DefaultGateway(),
						Provider: v1alpha1.DefaultEnvoyGatewayProvider(),
					},
				},
				Namespace: "",
			},
			expect: false,
		},
		{
			name: "unspecified envoy gateway",
			cfg: &Server{
				Namespace: "test-ns",
				Logger:    logging.DefaultLogger(v1alpha1.LogLevelInfo),
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
