// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package traces

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/fraction"
)

type Runner struct {
	cfg *config.Server
	tp  *trace.TracerProvider
	log logging.Logger
}

func New(cfg *config.Server) *Runner {
	return &Runner{
		cfg: cfg,
		log: cfg.Logger.WithName("traces-runner"),
	}
}

func (r *Runner) Start(ctx context.Context) error {
	if r.cfg.EnvoyGateway.Telemetry == nil ||
		r.cfg.EnvoyGateway.Telemetry.Traces == nil {
		return nil
	}

	tracesConfig := r.cfg.EnvoyGateway.GetEnvoyGatewayTelemetry().Traces
	sinkConfig := tracesConfig.Sink
	configObj := sinkConfig.OpenTelemetry

	endpoint := net.JoinHostPort(sinkConfig.OpenTelemetry.Host, fmt.Sprint(sinkConfig.OpenTelemetry.Port))

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("envoy-gateway"),
		),
	)
	if err != nil {
		return err
	}

	// Get sampler configuration
	sampler := r.getSampler(tracesConfig)
	r.log.Info("start tracer",
		"endpoint", endpoint,
		"sampler", sampler.Description(),
	)
	switch configObj.Protocol {
	case egv1a1.HTTPProtocol:
		// Create OTLP HTTP exporter
		exporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			// TODO: support TLS configuration for OTLP exporter
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			return err
		}

		bsp := trace.NewBatchSpanProcessor(exporter)
		tp := trace.NewTracerProvider(
			trace.WithSpanProcessor(bsp),
			trace.WithResource(res),
			trace.WithSampler(sampler),
		)

		otel.SetTracerProvider(tp)
		r.tp = tp

		return nil
	default:
		// use GRPC protocol by default
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			// TODO: support TLS configuration for OTLP exporter
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return err
		}

		bsp := trace.NewBatchSpanProcessor(exporter)
		tp := trace.NewTracerProvider(
			trace.WithSpanProcessor(bsp),
			trace.WithResource(res),
			trace.WithSampler(sampler),
		)

		otel.SetTracerProvider(tp)
		r.tp = tp

		return nil
	}
}

// getSampler returns the configured sampler or a default sampler
func (r *Runner) getSampler(tracesConfig *egv1a1.EnvoyGatewayTraces) trace.Sampler {
	if tracesConfig.SamplingRate != nil {
		rate := fraction.Deref(tracesConfig.SamplingRate, 1.0)
		return trace.TraceIDRatioBased(rate)
	}
	// Default to always sample (100%)
	return trace.AlwaysSample()
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
