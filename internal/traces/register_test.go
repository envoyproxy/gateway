// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package traces

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

func TestTracesRunner_New(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: &egv1a1.EnvoyGateway{
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
				Telemetry: &egv1a1.EnvoyGatewayTelemetry{
					Traces: &egv1a1.EnvoyGatewayTraces{
						Sink: egv1a1.EnvoyGatewayTraceSink{
							Type: egv1a1.TraceSinkTypeOpenTelemetry,
							OpenTelemetry: &egv1a1.EnvoyGatewayOpenTelemetrySink{
								Host:     "localhost",
								Port:     4317,
								Protocol: egv1a1.GRPCProtocol,
							},
						},
					},
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

func TestTracesRunner_Start_ValidConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		host     string
		port     int32
	}{
		{
			name:     "grpc protocol configuration",
			protocol: egv1a1.GRPCProtocol,
			host:     "localhost",
			port:     4317,
		},
		{
			name:     "http protocol configuration",
			protocol: egv1a1.HTTPProtocol,
			host:     "localhost",
			port:     4318,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Telemetry: &egv1a1.EnvoyGatewayTelemetry{
							Traces: &egv1a1.EnvoyGatewayTraces{
								Sink: egv1a1.EnvoyGatewayTraceSink{
									Type: egv1a1.TraceSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.EnvoyGatewayOpenTelemetrySink{
										Host:     tt.host,
										Port:     tt.port,
										Protocol: tt.protocol,
									},
								},
							},
						},
					},
				},
			}

			runner := New(cfg)
			require.NotNil(t, runner)

			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Start will create the exporter
			err := runner.Start(ctx)
			// We don't expect an error during initialization
			require.NoError(t, err)

			// Clean up
			_ = runner.Close()
		})
	}
}

func TestTracesRunner_Start_Configuration(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
	}{
		{
			name:     "grpc protocol configuration",
			protocol: egv1a1.GRPCProtocol,
		},
		{
			name:     "http protocol configuration",
			protocol: egv1a1.HTTPProtocol,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Telemetry: &egv1a1.EnvoyGatewayTelemetry{
							Traces: &egv1a1.EnvoyGatewayTraces{
								Sink: egv1a1.EnvoyGatewayTraceSink{
									Type: egv1a1.TraceSinkTypeOpenTelemetry,
									OpenTelemetry: &egv1a1.EnvoyGatewayOpenTelemetrySink{
										Host:     "localhost",
										Port:     4317,
										Protocol: tt.protocol,
									},
								},
							},
						},
					},
				},
			}

			runner := New(cfg)
			require.NotNil(t, runner)
			require.Equal(t, "traces", runner.Name())

			// Note: We don't call Start() here because it requires a real OTLP endpoint
			// This test just verifies the runner can be created with valid configuration
		})
	}
}
