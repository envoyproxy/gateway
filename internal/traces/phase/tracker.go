// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package phase records one tracing span per phase of a long, sequential operation.
//
// It deliberately lives outside the parent traces package, which pulls in the OTLP
// exporters: the packages that instrument their phases should not gain a dependency
// on the exporters to do it. Only the OpenTelemetry API is used here.
package phase

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracker records one child span per phase of a long, sequential operation such as a
// translation, so that its total duration can be attributed to a phase instead of
// being reported as one opaque span.
//
// The phases run one after another, so a single span is in flight at any time:
//
//	phases := phase.NewTracker(ctx, tracer)
//	defer phases.EndInFlight()
//
//	phases.Start("Translator.processRoutes", attribute.Int("routes.count", len(routes)))
//	processRoutes(routes)
//	phases.End()
//
// Phase spans stay flat: Start does not hand back the phase's context, so a span
// started inside a phase parents to the enclosing span, not to the phase.
type Tracker struct {
	ctx      context.Context
	tracer   trace.Tracer
	inFlight trace.Span
}

// NewTracker returns a tracker that starts its phase spans as children of the span
// in ctx.
func NewTracker(ctx context.Context, tracer trace.Tracer) *Tracker {
	return &Tracker{ctx: ctx, tracer: tracer}
}

// Start begins a phase span. The attrs are meant to record the size of the input the
// phase is about to process: without them a slow run cannot be told apart from a run
// over a bigger input.
func (p *Tracker) Start(name string, attrs ...attribute.KeyValue) {
	_, p.inFlight = p.tracer.Start(p.ctx, name, trace.WithAttributes(attrs...))
}

// End ends the phase span started by the last call to Start.
func (p *Tracker) End() {
	if p.inFlight == nil {
		return
	}
	p.inFlight.End()
	p.inFlight = nil
}

// EndInFlight ends a phase span that was started but never ended, marking it as
// incomplete. Deferring it keeps the phase visible when the operation panics: a
// panicking phase is exactly the one an operator needs to see, and an unended span
// is never exported.
func (p *Tracker) EndInFlight() {
	if p.inFlight == nil {
		return
	}
	p.inFlight.SetStatus(codes.Error, "phase did not complete")
	p.End()
}
