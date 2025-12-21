// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package traces

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

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
	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("envoy-gateway"),
		),
	)
	if err != nil {
		return err
	}

	tp := trace.NewTracerProvider(
		trace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	r.tp = tp

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
