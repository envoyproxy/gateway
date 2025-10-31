// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package traces

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

type Runner struct {
	cfg *config.Server
	tp  *trace.TracerProvider
}

func New(cfg *config.Server) *Runner {
	return &Runner{
		cfg: cfg,
	}
}

func (r *Runner) Start(ctx context.Context) error {
	if r.cfg.EnvoyGateway.DisableTraces() {
		return nil
	}

	config := r.cfg.EnvoyGateway.GetEnvoyGatewayTelemetry().Traces.Sink
	configObj := config.OpenTelemetry

	endpoint := fmt.Sprintf("%s:%d", config.OpenTelemetry.Host, config.OpenTelemetry.Port)
	if configObj.Protocol == egv1a1.GRPCProtocol {
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return err
		}

		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceNameKey.String("envoy-gateway"),
			),
		)
		if err != nil {
			return err
		}

		tp := trace.NewTracerProvider(
			trace.WithBatcher(exporter),
			trace.WithResource(res),
			trace.WithSampler(trace.AlwaysSample()), // TODO: configurable?
		)

		otel.SetTracerProvider(tp)
		r.tp = tp

		return nil
	}

	if configObj.Protocol == egv1a1.HTTPProtocol {
		// Create OTLP HTTP exporter
		exporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(),
			// TODO: should we make path configurable?
			// otlptracehttp.WithURLPath("/v1/traces"),   // Optional: custom path
		)
		if err != nil {
			return err
		}

		res, err := resource.New(ctx,
			resource.WithAttributes(
				semconv.ServiceNameKey.String("envoy-gateway"),
			),
		)
		if err != nil {
			return err
		}

		tp := trace.NewTracerProvider(
			trace.WithBatcher(exporter),
			trace.WithResource(res),
			trace.WithSampler(trace.AlwaysSample()), // TODO: configurable?
		)

		otel.SetTracerProvider(tp)
		r.tp = tp

		return nil
	}

	return nil
}

func (r *Runner) Name() string {
	return "traces"
}

func (r *Runner) Close() error {
	if r.tp != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return r.tp.Shutdown(ctx)
	}
	return nil
}
