// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
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

// TestTranslatePhaseSpans pins the phase spans emitted by a single Translate call.
// The fan-out is fixed and small on purpose: it must not grow with the number of
// listeners, routes or policies being translated.
func TestTranslatePhaseSpans(t *testing.T) {
	sr := recordSpans(t)

	tr := &Translator{Logger: logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo)}
	traceCtx, parent := tracer.Start(t.Context(), "Translator.Translate")
	// Two listeners carrying three IR routes between them, so that the recorded counts
	// are pinned to real inputs rather than to zero. The returned error is ignored:
	// these listeners are too bare to pass xDS validation, and translation is
	// best-effort, so every phase still runs.
	_, _ = tr.Translate(traceCtx, &ir.Xds{
		HTTP: []*ir.HTTPListener{
			{Routes: []*ir.HTTPRoute{{Name: "route-1"}, {Name: "route-2"}}},
			{Routes: []*ir.HTTPRoute{{Name: "route-3"}}},
		},
		EnvoyPatchPolicies: []*ir.EnvoyPatchPolicy{
			{EnvoyPatchPolicyStatus: ir.EnvoyPatchPolicyStatus{Name: "policy", Status: &gwapiv1.PolicyStatus{}}},
		},
	})
	parent.End()

	names := make([]string, 0, len(sr.Ended()))
	for _, span := range sr.Ended() {
		names = append(names, span.Name())
		// The input inventory lands on the enclosing stage span, so one lookup tells an
		// operator whether a slow build processed a bigger input than the last one.
		if span.Name() == "Translator.Translate" {
			require.Contains(t, span.Attributes(), attribute.Int("envoy-patch-policies.count", 1))
			require.Contains(t, span.Attributes(), attribute.Int("http-listeners.count", 2))
			require.Contains(t, span.Attributes(), attribute.Int("ir-http-routes.count", 3))
			continue
		}
		// Every phase hangs off the stage span; an orphan phase cannot be attributed
		// to the build it belongs to.
		require.Equal(t, parent.SpanContext().SpanID(), span.Parent().SpanID(), span.Name())
	}
	require.ElementsMatch(t, []string{
		"Translator.Translate",
		"XdsTranslator.processHTTPListenerXdsTranslation",
		"XdsTranslator.notifyExtensionServerAboutListeners",
		"XdsTranslator.processJSONPatches",
		"XdsTranslator.processExtensionPostTranslationHook",
		"XdsTranslator.validateAllXdsResources",
	}, names)
}

// TestCountIRHTTPRoutes checks that the instrumentation does not become the first
// thing to panic on a malformed IR, which would move the failure ahead of the phase
// span that is supposed to report it.
func TestCountIRHTTPRoutes(t *testing.T) {
	require.Equal(t, 0, countIRHTTPRoutes(nil))
	require.Equal(t, 3, countIRHTTPRoutes([]*ir.HTTPListener{
		{Routes: []*ir.HTTPRoute{{}, {}}},
		nil,
		{Routes: []*ir.HTTPRoute{{}}},
	}))
}
