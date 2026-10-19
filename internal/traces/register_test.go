// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package traces

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/cmd/version"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

func TestTracesRunner_New(t *testing.T) {
	cfg := &config.Server{
		Logger: logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
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
				cfg: &config.Server{
					Logger: logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
				},
				tp: nil,
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

func TestNewResource(t *testing.T) {
	res, err := newResource(t.Context())
	require.NoError(t, err)

	attrs := map[attribute.Key]string{}
	for _, kv := range res.Attributes() {
		attrs[kv.Key] = kv.Value.Emit()
	}
	v := version.Get()
	require.Equal(t, "envoy-gateway", attrs[semconv.ServiceNameKey])
	require.Equal(t, v.EnvoyGatewayVersion, attrs[semconv.ServiceVersionKey])
	require.Contains(t, attrs, gitCommitIDKey)
	require.Equal(t, v.GitCommitID, attrs[gitCommitIDKey])
}
