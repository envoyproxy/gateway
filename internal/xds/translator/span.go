// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"context"

	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var tracer = otel.Tracer("envoy-gateway/xds/translator")

// startPhase starts a child span for one phase of the xDS translation and returns
// a function that ends it. The phases run sequentially, so the returned function
// is called at the end of the phase rather than deferred.
//
// The attrs are meant to record the size of the input the phase is about to
// process: without them a slow build cannot be told apart from a bigger one.
//
// Spans are per-phase, never per-resource. A span per route or per extension hook
// call would emit tens of thousands of spans for a single build on a large cluster,
// which bloats every trace and makes the exporter itself a cost inside the build.
// Per-resource latency belongs in a histogram metric instead.
func startPhase(ctx context.Context, name string, attrs ...attribute.KeyValue) func() {
	_, span := tracer.Start(ctx, name, trace.WithAttributes(attrs...))
	return func() { span.End() }
}

// countIRHTTPRoutes returns the total number of routes across all the HTTP listeners.
func countIRHTTPRoutes(listeners []*ir.HTTPListener) int {
	var routes int
	for _, l := range listeners {
		routes += len(l.Routes)
	}
	return routes
}

// xdsResourceCountAttrs reports how many resources of each type the table holds. The
// same set of keys is always emitted, so the counts of two builds are comparable.
func xdsResourceCountAttrs(tCtx *types.ResourceVersionTable) []attribute.KeyValue {
	countedTypes := []struct {
		key     string
		xdsType resourcev3.Type
	}{
		{"listeners", resourcev3.ListenerType},
		{"route-configurations", resourcev3.RouteType},
		{"clusters", resourcev3.ClusterType},
		{"endpoints", resourcev3.EndpointType},
		{"secrets", resourcev3.SecretType},
	}

	attrs := make([]attribute.KeyValue, 0, len(countedTypes)+1)
	for _, ct := range countedTypes {
		attrs = append(attrs, attribute.Int("xds-resources."+ct.key, len(tCtx.XdsResources[ct.xdsType])))
	}

	// Counted over the whole table so that resource types not listed above, for
	// example those injected by an extension server, are still reflected.
	var total int
	for _, resources := range tCtx.XdsResources {
		total += len(resources)
	}

	return append(attrs, attribute.Int("xds-resources.total", total))
}
