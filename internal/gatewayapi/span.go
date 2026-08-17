// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

var tracer = otel.Tracer("envoy-gateway/gateway-api/translator")

// startPhase starts a child span for one phase of the Gateway API translation and
// returns a function that ends it. The phases run sequentially, so the returned
// function is called at the end of the phase rather than deferred.
//
// The attrs are meant to record the size of the input the phase is about to
// process: without them a slow translation cannot be told apart from a bigger one.
//
// Spans are per-phase, never per-resource. A span per route or per policy would
// emit tens of thousands of spans for a single translation on a large cluster,
// which bloats every trace and makes the exporter itself a cost inside the
// translation. Per-resource latency belongs in a histogram metric instead.
func startPhase(ctx context.Context, name string, attrs ...attribute.KeyValue) func() {
	_, span := tracer.Start(ctx, name, trace.WithAttributes(attrs...))
	return func() { span.End() }
}

// inputSizeAttrs reports the size of the resource tree a translation is about to
// process, so that a slow translation can be attributed to a bigger input instead
// of being correlated against other signals by hand.
func inputSizeAttrs(resources *resource.Resources) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("gateways.count", len(resources.Gateways)),
		attribute.Int("listenersets.count", len(resources.ListenerSets)),
		attribute.Int("httproutes.count", len(resources.HTTPRoutes)),
		attribute.Int("grpcroutes.count", len(resources.GRPCRoutes)),
		attribute.Int("tlsroutes.count", len(resources.TLSRoutes)),
		attribute.Int("tcproutes.count", len(resources.TCPRoutes)),
		attribute.Int("udproutes.count", len(resources.UDPRoutes)),
		attribute.Int("backends.count", len(resources.Backends)),
		attribute.Int("services.count", len(resources.Services)),
		attribute.Int("endpointslices.count", len(resources.EndpointSlices)),
		attribute.Int("clienttrafficpolicies.count", len(resources.ClientTrafficPolicies)),
		attribute.Int("backendtrafficpolicies.count", len(resources.BackendTrafficPolicies)),
		attribute.Int("securitypolicies.count", len(resources.SecurityPolicies)),
		attribute.Int("backendtlspolicies.count", len(resources.BackendTLSPolicies)),
		attribute.Int("envoyextensionpolicies.count", len(resources.EnvoyExtensionPolicies)),
		attribute.Int("extensionserverpolicies.count", len(resources.ExtensionServerPolicies)),
		attribute.Int("envoypatchpolicies.count", len(resources.EnvoyPatchPolicies)),
	}
}
