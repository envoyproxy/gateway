// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

// recordSpans installs a span recorder for the duration of the test and returns it.
func recordSpans(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()

	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	original := tracer
	tracer = tp.Tracer("test")
	t.Cleanup(func() {
		tracer = original
	})

	return sr
}

func newTracingTestTranslator() *Translator {
	return &Translator{
		GatewayControllerName: egv1a1.GatewayControllerName,
		GatewayClassName:      "envoy-gateway-class",
		Logger:                logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
	}
}

// TestTranslatePhaseSpans pins the phase spans emitted by a single Translate call.
// The fan-out is fixed and small on purpose: it must not grow with the number of
// gateways, routes or policies being translated.
func TestTranslatePhaseSpans(t *testing.T) {
	sr := recordSpans(t)

	traceCtx, stage := tracer.Start(t.Context(), "TranslateToIR")
	// The returned error is ignored: this bare resource set has no control plane TLS
	// secret, and translation is best-effort, so every phase still runs.
	_, _ = newTracingTestTranslator().Translate(traceCtx, &resource.Resources{
		HTTPRoutes: []*gwapiv1.HTTPRoute{{}, {}, {}},
	})
	stage.End()

	names := make([]string, 0, len(sr.Ended()))
	for _, s := range sr.Ended() {
		names = append(names, s.Name())
		if s.Name() == "TranslateToIR" {
			// The input inventory lands on the enclosing stage span, so one lookup tells
			// an operator whether a slow translation processed a bigger input.
			require.Contains(t, s.Attributes(), attribute.Int("http-routes.count", 3))
			continue
		}
		// Every phase hangs off the stage span; an orphan phase cannot be attributed to
		// the translation it belongs to.
		require.Equal(t, stage.SpanContext().SpanID(), s.Parent().SpanID(), s.Name())
	}
	require.ElementsMatch(t, []string{
		"GatewayApiTranslator.BuildPolicyIndexes",
		"GatewayApiTranslator.ProcessListeners",
		"GatewayApiTranslator.ProcessHTTPRoutes",
		"GatewayApiTranslator.ProcessGRPCRoutes",
		"GatewayApiTranslator.ProcessClientTrafficPolicies",
		"GatewayApiTranslator.ProcessBackendTrafficPolicies",
		"GatewayApiTranslator.ProcessSecurityPolicies",
		"GatewayApiTranslator.ProcessEnvoyExtensionPolicies",
		"TranslateToIR",
	}, names)
}

// TestTranslatePhaseSpanEndedOnPanic covers the path an operator most needs the span
// for: the translator panics on some inputs and the runner recovers from it, so the
// phase that panicked has to reach the exporter instead of being dropped. The exact
// status wording is owned by internal/traces.
func TestTranslatePhaseSpanEndedOnPanic(t *testing.T) {
	sr := recordSpans(t)
	traceCtx, stage := tracer.Start(t.Context(), "TranslateToIR")

	// A nil HTTPRoute panics inside ProcessHTTPRoutes.
	require.Panics(t, func() {
		_, _ = newTracingTestTranslator().Translate(traceCtx, &resource.Resources{
			HTTPRoutes: []*gwapiv1.HTTPRoute{nil},
		})
	})
	stage.End()

	var found bool
	for _, s := range sr.Ended() {
		if s.Name() == "GatewayApiTranslator.ProcessHTTPRoutes" {
			found = true
			require.Equal(t, codes.Error, s.Status().Code)
		}
	}
	require.True(t, found, "the panicking phase span should still be exported")
}
