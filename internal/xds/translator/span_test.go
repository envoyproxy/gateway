// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
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

	tr := &Translator{}
	traceCtx, parent := tracer.Start(t.Context(), "Translator.Translate")
	_, err := tr.Translate(traceCtx, &ir.Xds{
		EnvoyPatchPolicies: []*ir.EnvoyPatchPolicy{
			{EnvoyPatchPolicyStatus: ir.EnvoyPatchPolicyStatus{Name: "policy", Status: &gwapiv1.PolicyStatus{}}},
		},
	})
	parent.End()
	require.NoError(t, err)

	names := make([]string, 0, len(sr.Ended()))
	for _, span := range sr.Ended() {
		names = append(names, span.Name())
		// The input inventory lands on the enclosing stage span, so one lookup tells an
		// operator whether a slow build processed a bigger input than the last one.
		if span.Name() == "Translator.Translate" {
			require.Contains(t, span.Attributes(), attribute.Int("envoy-patch-policies.count", 1))
			require.Contains(t, span.Attributes(), attribute.Int("http-routes.count", 0))
		}
	}
	require.ElementsMatch(t, []string{
		"Translator.Translate",
		"XdsTranslator.processHTTPListenerXdsTranslation",
		"XdsTranslator.notifyExtensionServerAboutListeners",
		"XdsTranslator.processMergedBackendClusters",
		"XdsTranslator.processJSONPatches",
		"XdsTranslator.processExtensionPostTranslationHook",
		"ResourceVersionTable.ValidateAll",
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
