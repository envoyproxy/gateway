// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
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
