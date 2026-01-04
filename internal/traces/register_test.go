// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package traces

import (
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

func TestTracesRunner_New(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: &egv1a1.EnvoyGateway{
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
				Telemetry: &egv1a1.EnvoyGatewayTelemetry{
					Traces: &egv1a1.EnvoyGatewayTraces{},
				},
			},
		},
	}

	runner := New(cfg)
	require.NotNil(t, runner)
	require.Equal(t, cfg, runner.cfg)
	require.Nil(t, runner.tp)
}

func TestTracesRunner_Close(t *testing.T) {
	tests := []struct {
		name    string
		runner  *Runner
		wantErr bool
	}{
		{
			name: "close with nil tracer provider",
			runner: &Runner{
				cfg: &config.Server{},
				tp:  nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.runner.Close()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
