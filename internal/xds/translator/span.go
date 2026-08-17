// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var tracer = otel.Tracer("envoy-gateway/xds/translator")

// countIRHTTPRoutes returns the total number of routes across all the HTTP listeners.
// Nil listeners are skipped: instrumentation must not be the first thing to panic on a
// malformed IR, otherwise it moves the failure ahead of the phase that would report it.
func countIRHTTPRoutes(listeners []*ir.HTTPListener) int {
	var routes int
	for _, l := range listeners {
		if l == nil {
			continue
		}
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
