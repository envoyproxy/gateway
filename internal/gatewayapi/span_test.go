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
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

// TestTranslatePhaseSpans pins the phase spans emitted by a single Translate call.
// The fan-out is fixed and small on purpose: it must not grow with the number of
// gateways, routes or policies being translated.
func TestTranslatePhaseSpans(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(sr))
	original := tracer
	tracer = tp.Tracer("test")
	t.Cleanup(func() { tracer = original })

	tr := &Translator{
		GatewayControllerName: egv1a1.GatewayControllerName,
		GatewayClassName:      "envoy-gateway-class",
		Logger:                logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo),
	}

	traceCtx, span := tp.Tracer("test").Start(t.Context(), "TranslateToIR")
	// The returned error is ignored: this bare resource set has no control plane TLS
	// secret, and translation is best-effort, so every phase still runs.
	_, _ = tr.Translate(traceCtx, &resource.Resources{
		HTTPRoutes: []*gwapiv1.HTTPRoute{{}, {}, {}},
	})
	span.End()

	names := make([]string, 0, len(sr.Ended()))
	for _, s := range sr.Ended() {
		names = append(names, s.Name())
	}
	require.ElementsMatch(t, []string{
		"GatewayApiTranslator.StatusDeepCopy",
		"GatewayApiTranslator.BuildTranslatorContext",
		"GatewayApiTranslator.BuildPolicyIndexes",
		"GatewayApiTranslator.ProcessListeners",
		"GatewayApiTranslator.ProcessBackends",
		"GatewayApiTranslator.ProcessHTTPRoutes",
		"GatewayApiTranslator.ProcessGRPCRoutes",
		"GatewayApiTranslator.ProcessL4Routes",
		"GatewayApiTranslator.ProcessClientTrafficPolicies",
		"GatewayApiTranslator.ProcessBackendTrafficPolicies",
		"GatewayApiTranslator.checkRouteOverlaps",
		"GatewayApiTranslator.ProcessSecurityPolicies",
		"GatewayApiTranslator.ProcessEnvoyExtensionPolicies",
		"GatewayApiTranslator.ProcessExtensionServerPolicies",
		"GatewayApiTranslator.ProcessGlobalResources",
		"TranslateToIR",
	}, names)

	// The input inventory lands on the enclosing stage span, so one lookup tells an
	// operator whether a slow translation processed a bigger input than the last one.
	for _, s := range sr.Ended() {
		if s.Name() != "TranslateToIR" {
			continue
		}
		require.Contains(t, s.Attributes(), attribute.Int("httproutes.count", 3))
	}
}
