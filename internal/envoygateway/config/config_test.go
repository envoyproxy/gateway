// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/logging"
)

var (
	TLSSecretKind       = gwapiv1.Kind("Secret")
	TLSUnrecognizedKind = gwapiv1.Kind("Unrecognized")
)

func TestValidate(t *testing.T) {
	cfg, err := New(os.Stdout)
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
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Gateway:  egv1a1.DefaultGateway(),
						Provider: egv1a1.DefaultEnvoyGatewayProvider(),
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
				Logger:    logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
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
