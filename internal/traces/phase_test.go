// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package traces

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestPhaseTracker(t *testing.T) {
	t.Run("records one span per phase", func(t *testing.T) {
		sr := tracetest.NewSpanRecorder()
		tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)).Tracer("test")

		ctx, parent := tracer.Start(t.Context(), "stage")
		phases := NewPhaseTracker(ctx, tracer)
		defer phases.EndInFlight()

		phases.Start("first", attribute.Int("routes.count", 3))
		phases.End()
		phases.Start("second")
		phases.End()
		parent.End()

		spans := sr.Ended()
		require.Len(t, spans, 3)
		require.Equal(t, "first", spans[0].Name())
		require.Contains(t, spans[0].Attributes(), attribute.Int("routes.count", 3))
		require.Equal(t, "second", spans[1].Name())
		require.Equal(t, codes.Unset, spans[1].Status().Code)

		// The phases have to hang off the enclosing stage span. Without this a phase
		// span becomes an orphan root and the stage it belongs to cannot be told.
		require.Equal(t, parent.SpanContext().SpanID(), spans[0].Parent().SpanID())
		require.Equal(t, parent.SpanContext().SpanID(), spans[1].Parent().SpanID())
	})

	t.Run("ends the in-flight phase when one panics", func(t *testing.T) {
		sr := tracetest.NewSpanRecorder()
		tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)).Tracer("test")

		// A panicking phase is the one an operator most needs to see, and an unended
		// span is never exported, so the deferred EndInFlight has to close it.
		require.Panics(t, func() {
			phases := NewPhaseTracker(t.Context(), tracer)
			defer phases.EndInFlight()

			phases.Start("completed")
			phases.End()
			phases.Start("panicked")
			panic("boom")
		})

		spans := sr.Ended()
		require.Len(t, spans, 2)
		require.Equal(t, "completed", spans[0].Name())
		require.Equal(t, codes.Unset, spans[0].Status().Code)
		require.Equal(t, "panicked", spans[1].Name())
		require.Equal(t, codes.Error, spans[1].Status().Code)
		require.Equal(t, "phase did not complete", spans[1].Status().Description)
	})

	t.Run("End and EndInFlight are safe with no phase in flight", func(t *testing.T) {
		sr := tracetest.NewSpanRecorder()
		tracer := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr)).Tracer("test")

		phases := NewPhaseTracker(t.Context(), tracer)
		phases.End()
		phases.EndInFlight()
		require.Empty(t, sr.Ended())
	})
}
