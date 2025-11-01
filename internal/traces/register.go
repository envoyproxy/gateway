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

	tracesConfig := r.cfg.EnvoyGateway.GetEnvoyGatewayTelemetry().Traces
	sinkConfig := tracesConfig.Sink
	configObj := sinkConfig.OpenTelemetry

	endpoint := fmt.Sprintf("%s:%d", sinkConfig.OpenTelemetry.Host, sinkConfig.OpenTelemetry.Port)

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

	// Get batch span processor options
	batchOptions := r.getBatchSpanProcessorOptions(tracesConfig)

	if configObj.Protocol == egv1a1.GRPCProtocol {
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return err
		}

		bsp := trace.NewBatchSpanProcessor(exporter, batchOptions...)
		tp := trace.NewTracerProvider(
			trace.WithSpanProcessor(bsp),
			trace.WithResource(res),
			trace.WithSampler(sampler),
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
		)
		if err != nil {
			return err
		}

		bsp := trace.NewBatchSpanProcessor(exporter, batchOptions...)
		tp := trace.NewTracerProvider(
			trace.WithSpanProcessor(bsp),
			trace.WithResource(res),
			trace.WithSampler(sampler),
		)

		otel.SetTracerProvider(tp)
		r.tp = tp

		return nil
	}

	return nil
}

// getSampler returns the configured sampler or a default sampler
func (r *Runner) getSampler(tracesConfig *egv1a1.EnvoyGatewayTraces) trace.Sampler {
	if tracesConfig.SamplingRate != nil {
		return trace.TraceIDRatioBased(*tracesConfig.SamplingRate)
	}
	// Default to always sample (100%)
	return trace.AlwaysSample()
}

// getBatchSpanProcessorOptions returns the configured batch span processor options
func (r *Runner) getBatchSpanProcessorOptions(tracesConfig *egv1a1.EnvoyGatewayTraces) []trace.BatchSpanProcessorOption {
	var options []trace.BatchSpanProcessorOption

	if tracesConfig.BatchSpanProcessorConfig != nil {
		cfg := tracesConfig.BatchSpanProcessorConfig

		if cfg.BatchTimeout != nil {
			timeout, err := time.ParseDuration(string(*cfg.BatchTimeout))
			if err == nil && timeout > 0 {
				options = append(options, trace.WithBatchTimeout(timeout))
			}
		}

		if cfg.MaxExportBatchSize != nil && *cfg.MaxExportBatchSize > 0 {
			options = append(options, trace.WithMaxExportBatchSize(*cfg.MaxExportBatchSize))
		}

		if cfg.MaxQueueSize != nil && *cfg.MaxQueueSize > 0 {
			options = append(options, trace.WithMaxQueueSize(*cfg.MaxQueueSize))
		}
	}

	// If no options were configured, use defaults
	// Default BatchTimeout is 5s, MaxExportBatchSize is 512, MaxQueueSize is 2048
	// These are the OpenTelemetry SDK defaults

	return options
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
